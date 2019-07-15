package rpc

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/rpc"
	"github.com/stretchr/testify/require"
)

type TestService struct{}

func (s *TestService) Foo(arg string) string {
	return arg
}

func TestStartHTTPEndpointStartStop(t *testing.T) {
	config := &Config{
		RPCAddr: "127.0.0.1",
		RPCPort: randomPort(),
	}

	ctx, cancel := context.WithCancel(context.Background())

	_, err := StartHTTPEndpoint(
		ctx, config, []rpc.API{
			{
				Namespace: "test",
				Version:   "1.0",
				Service:   &TestService{},
				Public:    true,
			},
		},
		rpc.HTTPTimeouts{
			ReadTimeout:  5 * time.Second,
			WriteTimeout: 5 * time.Second,
			IdleTimeout:  5 * time.Second,
		},
	)
	require.Nil(t, err, "unexpected error")

	res, err := rpcCall(config.RPCPort, "test_foo", []string{"baz"})
	require.Nil(t, err, "unexpected error")

	resStr := res.(string)
	require.Equal(t, "baz", resStr)

	cancel()

	_, err = rpcCall(config.RPCPort, "test_foo", []string{"baz"})
	require.NotNil(t, err)
}

func rpcCall(port int, method string, params []string) (interface{}, error) {
	parsedParams, err := json.Marshal(params)
	if err != nil {
		return nil, err
	}

	fullBody := fmt.Sprintf(
		`{ "id": 1, "jsonrpc": "2.0", "method": "%s", "params": %s }`,
		method, string(parsedParams),
	)

	res, err := http.Post(fmt.Sprintf("http://127.0.0.1:%d", port), "application/json", strings.NewReader(fullBody))
	if err != nil {
		return nil, err
	}

	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	var out map[string]interface{}
	err = json.Unmarshal(data, &out)
	if err != nil {
		return nil, err
	}

	result := out["result"].(interface{})
	return result, nil
}

func randomPort() int {
	return rand.Intn(65535-1025) + 1025
}
