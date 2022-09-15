package ws

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/require"
	"log"
	"net/url"
	"os"
	"testing"
	"time"
)

var (
	HOST = os.Getenv("HOST")
	from = []byte{}

	wc *websocket.Conn
)

func TestMain(m *testing.M) {
	if HOST == "" {
		HOST = "localhost:8542"
	}

	u := url.URL{Scheme: "ws", Host: HOST, Path: ""}
	log.Printf("connecting to %s", u.String())

	var err error
	wc, _, err = websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		fmt.Printf("failed to dial websocket")
		os.Exit(1)
	}

	defer wc.Close()

	// Start all tests
	code := m.Run()
	os.Exit(code)
}

func TestSingleRequest(t *testing.T) {
	log.Printf("test simple rpc request net_version with websocket")
	err := wc.WriteMessage(websocket.TextMessage, []byte(`{"jsonrpc":"2.0","method":"net_version","params":[],"id":0}`))
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
