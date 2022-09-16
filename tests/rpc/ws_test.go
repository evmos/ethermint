package rpc

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/require"
	"log"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"
)

var (
	HOST_WS = os.Getenv("HOST_WS")
	wsUrl   string
)

func init() {
	if HOST_WS == "" {
		HOST_WS = "localhost:8542"
	}

	u := url.URL{Scheme: "ws", Host: HOST_WS, Path: ""}
	wsUrl = u.String()
}

func TestWsSingleRequest(t *testing.T) {
	log.Printf("test simple rpc request net_version with websocket")

	wc, _, err := websocket.DefaultDialer.Dial(wsUrl, nil)
	require.NoError(t, err)
	defer wc.Close()

	err = wc.WriteMessage(websocket.TextMessage, []byte(`{"jsonrpc":"2.0","method":"net_version","params":[],"id":1}`))
	require.NoError(t, err)

	time.Sleep(1 * time.Second)
	_, mb, err := wc.ReadMessage()
	require.NoError(t, err)
	log.Printf("recv: %s", mb)

	var msg map[string]interface{}
	err = json.Unmarshal(mb, &msg)
	require.NoError(t, err)
	result, ok := msg["result"].(string)
	require.True(t, ok)
	require.Equal(t, "9000", result)
}

//func TestBatchRequest(t *testing.T) {
//	log.Printf("test batch request with websocket")
//	err := wc.WriteMessage(websocket.TextMessage, []byte(`[{"jsonrpc":"2.0","method":"net_version","params":[],"id":1},{"jsonrpc":"2.0","method":"eth_protocolVersion","params":[],"id":2}]`))
//	require.NoError(t, err)
//
//	time.Sleep(1 * time.Second)
//	_, mb, err := wc.ReadMessage()
//	require.NoError(t, err)
//
//	log.Printf("recv: %s", mb)
//}

func TestWsEth_subscribe_newHeads(t *testing.T) {
	log.Printf("test eth_subscribe newHeads with websocket")

	wc, _, err := websocket.DefaultDialer.Dial(wsUrl, nil)
	require.NoError(t, err)
	defer wc.Close()

	err = wc.WriteMessage(websocket.TextMessage, []byte(`{"id":1,"method":"eth_subscribe","params":["newHeads",{}]}`))
	require.NoError(t, err)

	time.Sleep(1 * time.Second)
	_, mb, err := wc.ReadMessage()
	require.NoError(t, err)
	log.Printf("recv: %s", mb)

	var msg map[string]interface{}
	err = json.Unmarshal(mb, &msg)
	require.NoError(t, err)
	subscribeId, ok := msg["result"].(string)
	require.True(t, ok)
	require.True(t, strings.HasPrefix(subscribeId, "0x"))

	time.Sleep(3 * time.Second)
	_, mb, err = wc.ReadMessage()
	require.NoError(t, err)
	log.Printf("recv: %s", mb)
	err = json.Unmarshal(mb, &msg)
	require.NoError(t, err)
	method, ok := msg["method"].(string)
	require.Equal(t, "eth_subscription", method)

	// id should not exist with eth_subscription event
	_, ok = msg["id"].(uint32)
	require.False(t, ok)

	wsUnsubscribe(t, wc, subscribeId) // eth_unsubscribe
}

func TestWsEth_subscribe_log(t *testing.T) {
	log.Printf("test eth_subscribe log with websocket")

	wc, _, err := websocket.DefaultDialer.Dial(wsUrl, nil)
	require.NoError(t, err)
	defer wc.Close()

	strJson := fmt.Sprintf(`{"jsonrpc":"2.0","id":0,"method":"eth_subscribe","params":["logs",{"topics":["%s", "%s"]}]}`, helloTopic, worldTopic)
	err = wc.WriteMessage(websocket.TextMessage, []byte(strJson))
	require.NoError(t, err)

	time.Sleep(1 * time.Second)
	_, mb, err := wc.ReadMessage()
	require.NoError(t, err)
	log.Printf("recv: %s", mb)

	var msg map[string]interface{}
	err = json.Unmarshal(mb, &msg)
	require.NoError(t, err)
	subscribeId, ok := msg["result"].(string)
	require.True(t, ok)
	require.True(t, strings.HasPrefix(subscribeId, "0x"))

	// do something here to receive subscription messages
	deployTestContractWithFunction(t)
	//

	time.Sleep(3 * time.Second)
	_, mb, err = wc.ReadMessage()
	require.NoError(t, err)
	log.Printf("recv: %s", mb)
	err = json.Unmarshal(mb, &msg)
	require.NoError(t, err)
	method, ok := msg["method"].(string)
	require.Equal(t, "eth_subscription", method)

	// id should not exist with eth_subscription event
	_, ok = msg["id"].(uint32)
	require.False(t, ok)

	wsUnsubscribe(t, wc, subscribeId) // eth_unsubscribe
}

func wsUnsubscribe(t *testing.T, wc *websocket.Conn, subscribeId string) {
	log.Printf("eth_unsubscribe %s", subscribeId)
	err := wc.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf(`{"id":1,"method":"eth_unsubscribe","params":["%s"]}`, subscribeId)))
	require.NoError(t, err)

	time.Sleep(1 * time.Second)
	_, mb, err := wc.ReadMessage()
	require.NoError(t, err)
	log.Printf("recv: %s", mb)
	var msg map[string]interface{}
	err = json.Unmarshal(mb, &msg)
	require.NoError(t, err)
	result, ok := msg["result"].(bool)
	require.True(t, ok)
	require.True(t, result)
}
