// This is a test utility for Ethermint's Web3 JSON-RPC services.
//
// To run these tests please first ensure you have the emintd running
// and have started the RPC service with `emintcl rest-server`.
//
// You can configure the desired port (or host) below.

package tester

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/cosmos/ethermint/version"
	"github.com/cosmos/ethermint/x/evm/types"
	"io/ioutil"
	"math/big"
	"net/http"
	"testing"
)

const (
	host          = "127.0.0.1"
	port          = 1317
	addrA         = "0xc94770007dda54cF92009BFF0dE90c06F603a09f"
	addrAStoreKey = 0
)

var addr = fmt.Sprintf("http://%s:%d/rpc", host, port)

type Request struct {
	Version string   `json:"jsonrpc"`
	Method  string   `json:"method"`
	Params  []string `json:"params"`
	Id      int      `json:"id"`
}

func createRequest(method string, params []string) Request {
	return Request{
		Version: "2.0",
		Method:  method,
		Params:  params,
		Id:      1,
	}
}

func call(t *testing.T, method string, params []string, resp interface{}) {
	req, err := json.Marshal(createRequest(method, params))
	if err != nil {
		t.Error(err)
	}

	res, err := http.Post(addr, "application/json", bytes.NewBuffer(req))
	if err != nil {
		t.Error(err)
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		t.Error(err)
	}

	err = json.Unmarshal(body, resp)
	if err != nil {
		t.Error(err)
	}
}

func TestEth_protocolVersion(t *testing.T) {
	expectedRes := version.ProtocolVersion

	res := &types.QueryResProtocolVersion{}
	call(t, "eth_protocolVersion", []string{}, res)

	t.Logf("Got protocol version: %s\n", res.Version)

	if res.Version != expectedRes {
		t.Errorf("expected: %s got: %s\n", expectedRes, res)
	}
}

func TestEth_blockNumber(t *testing.T) {
	res := &types.QueryResBlockNumber{}
	call(t, "eth_blockNumber", []string{}, res)

	t.Logf("Got block number: %s\n", res.Number.String())

	// -1 if x <  y, 0 if x == y; where x is res, y is 0
	if res.Number.Cmp(big.NewInt(0)) < 1 {
		t.Errorf("Invalid block number got: %v", res)
	}
}

func TestEth_GetBalance(t *testing.T) {
	//expectedRes := types.QueryResBalance{Balance:}
	res := &types.QueryResBalance{}
	call(t, "eth_getBalance", []string{addrA, "latest"}, res)

	t.Logf("Got balance %s for %s\n", res.Balance.String(), addrA)

	// 0 if x == y; where x is res, y is 0
	if res.Balance.ToInt().Cmp(big.NewInt(0)) != 0 {
		t.Errorf("expected balance: %d, got: %s", 0, res.Balance.String())
	}
}

func TestEth_GetStorageAt(t *testing.T) {
	expectedRes := types.QueryResStorage{Value: []byte{}}
	res := &types.QueryResStorage{}
	call(t, "eth_getStorageAt", []string{addrA, string(addrAStoreKey), "latest"}, res)

	t.Logf("Got value [%X] for %s with key %X\n", res.Value, addrA, addrAStoreKey)

	if !bytes.Equal(res.Value, expectedRes.Value) {
		t.Errorf("expected: %X got: %X", expectedRes.Value, res.Value)
	}
}

func TestEth_GetCode(t *testing.T) {
	expectedRes := types.QueryResCode{Code: []byte{}}
	res := &types.QueryResCode{}
	call(t, "eth_getCode", []string{addrA, "latest"}, res)

	t.Logf("Got code [%X] for %s\n", res.Code, addrA)
	if !bytes.Equal(expectedRes.Code, res.Code) {
		t.Errorf("expected: %X got: %X", expectedRes.Code, res.Code)
	}
}
