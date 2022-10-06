package codec

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/log"
	"github.com/evmos/ethermint/rpc/types"
	"github.com/gorilla/websocket"
)

type websocketCodec struct {
	*JSONCodec
	conn *websocket.Conn
	info PeerInfo

	wg        sync.WaitGroup
	pingReset chan struct{}
}

func NewWebsocketCodec(conn *websocket.Conn, host string, req http.Header) ServerCodec {
	conn.SetReadLimit(types.WsMessageSizeLimit)
	conn.SetPongHandler(func(appData string) error {
		err := conn.SetReadDeadline(time.Time{})
		if err != nil {
			log.Trace("set read deadline for websocket codec failed", "err", err)
		}
		return nil
	})
	wc := &websocketCodec{
		JSONCodec: NewFuncCodec(conn, conn.WriteJSON, conn.ReadJSON).(*JSONCodec),
		conn:      conn,
		pingReset: make(chan struct{}, 1),
		info: PeerInfo{
			Transport:  "ws",
			RemoteAddr: conn.RemoteAddr().String(),
		},
	}
	// Fill in connection details.
	wc.info.HTTP.Host = host
	wc.info.HTTP.Origin = req.Get("Origin")
	wc.info.HTTP.UserAgent = req.Get("User-Agent")
	// Start pinger.
	wc.wg.Add(1)
	go wc.PingLoop()
	return wc
}

func (wc *websocketCodec) Close() {
	wc.JSONCodec.Close()
	wc.wg.Wait()
}

func (wc *websocketCodec) PeerInfo() PeerInfo {
	return wc.info
}

func (wc *websocketCodec) WriteJSON(ctx context.Context, v interface{}) error {
	err := wc.JSONCodec.WriteJSON(ctx, v)
	if err == nil {
		// Notify pingLoop to delay the next idle ping.
		select {
		case wc.pingReset <- struct{}{}:
		default:
		}
	}
	return err
}

// pingLoop sends periodic ping frames when the connection is idle.
func (wc *websocketCodec) PingLoop() {
	timer := time.NewTimer(types.WsPingInterval)
	defer wc.wg.Done()
	defer timer.Stop()

	for {
		select {
		case <-wc.Closed():
			return
		case <-wc.pingReset:
			if !timer.Stop() {
				<-timer.C
			}
			timer.Reset(types.WsPingInterval)
		case <-timer.C:
			wc.JSONCodec.encMu.Lock()
			err := wc.conn.SetWriteDeadline(time.Now().Add(types.WsPingWriteTimeout))
			if err != nil {
				log.Trace("set write deadline for ping failed", "err", err)
			}
			err = wc.conn.WriteMessage(websocket.PingMessage, nil)
			if err != nil {
				log.Trace("write message for ping failed", "err", err)
			}
			err = wc.conn.SetReadDeadline(time.Now().Add(types.WsPongTimeout))
			if err != nil {
				log.Trace("set read deadline for ping failed", "err", err)
			}
			wc.JSONCodec.encMu.Unlock()
			timer.Reset(types.WsPingInterval)
		}
	}
}
