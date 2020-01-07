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
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"testing"

	"github.com/cosmos/ethermint/version"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

const (
	host          = "localhost"
	port          = 8545
	addrA         = "0xc94770007dda54cF92009BFF0dE90c06F603a09f"
	addrAStoreKey = 0
)

var addr = fmt.Sprintf("http://%s:%d", host, port)

type Request struct {
	Version string   `json:"jsonrpc"`
	Method  string   `json:"method"`
	Params  []string `json:"params"`
	ID      int      `json:"id"`
}

type RPCError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

type Response struct {
	Error  *RPCError       `json:"error"`
	ID     int             `json:"id"`
	Result json.RawMessage `json:"result,omitempty"`
}

func createRequest(method string, params []string) Request {
	return Request{
		Version: "2.0",
		Method:  method,
		Params:  params,
		ID:      1,
	}
}

func call(method string, params []string) (*Response, error) {
	req, err := json.Marshal(createRequest(method, params))
	if err != nil {
		return nil, err
	}

	/* #nosec */
	res, err := http.Post(addr, "application/json", bytes.NewBuffer(req))
	if err != nil {
		return nil, err
	}

	decoder := json.NewDecoder(res.Body)
	var rpcRes *Response
	err = decoder.Decode(&rpcRes)
	if err != nil {
		return nil, err
	}

	if rpcRes.Error != nil {
		return nil, errors.New(rpcRes.Error.Message)
	}

	err = res.Body.Close()
	if err != nil {
		return nil, err
	}

	return rpcRes, nil
}

func TestEth_protocolVersion(t *testing.T) {
	expectedRes := hexutil.Uint(version.ProtocolVersion)

	rpcRes, err := call("eth_protocolVersion", []string{})
	if err != nil {
		t.Fatal(err)
	}

	var res hexutil.Uint
	err = res.UnmarshalJSON(rpcRes.Result)

	if err != nil {
		t.Fatal(err)
	}

	t.Logf("Got protocol version: %s\n", res.String())

	if res != expectedRes {
		t.Fatalf("expected: %s got: %s\n", expectedRes.String(), rpcRes.Result)
	}
}

func TestEth_blockNumber(t *testing.T) {
	rpcRes, err := call("eth_blockNumber", []string{})
	if err != nil {
		t.Fatal(err)
	}
	var res hexutil.Uint64
	err = res.UnmarshalJSON(rpcRes.Result)

	if err != nil {
		t.Fatal(err)
	}

	t.Logf("Got block number: %s\n", res.String())
}

func TestEth_GetBalance(t *testing.T) {
	rpcRes, err := call("eth_getBalance", []string{addrA, "0x0"})
	if err != nil {
		t.Fatal(err)
		return
	}

	var res hexutil.Big
	err = res.UnmarshalJSON(rpcRes.Result)
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("Got balance %s for %s\n", res.String(), addrA)

	// 0 if x == y; where x is res, y is 0
	if res.ToInt().Cmp(big.NewInt(0)) != 0 {
		t.Errorf("expected balance: %d, got: %s", 0, res.String())
	}
}

func TestEth_GetStorageAt(t *testing.T) {
	expectedRes := hexutil.Bytes{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	rpcRes, err := call("eth_getStorageAt", []string{addrA, string(addrAStoreKey), "0x0"})
	if err != nil {
		t.Fatal(err)
	}

	var storage hexutil.Bytes
	err = storage.UnmarshalJSON(rpcRes.Result)

	if err != nil {
		t.Fatal(err)
	}

	t.Logf("Got value [%X] for %s with key %X\n", storage, addrA, addrAStoreKey)

	if !bytes.Equal(storage, expectedRes) {
		t.Errorf("expected: %d (%d bytes) got: %d (%d bytes)", expectedRes, len(expectedRes), storage, len(storage))
	}
}

func TestEth_GetCode(t *testing.T) {
	expectedRes := hexutil.Bytes{}
	rpcRes, err := call("eth_getCode", []string{addrA, "0x0"})
	if err != nil {
		t.Error(err)
	}

	var code hexutil.Bytes
	err = code.UnmarshalJSON(rpcRes.Result)

	if err != nil {
		t.Fatal(err)
	}

	t.Logf("Got code [%X] for %s\n", code, addrA)
	if !bytes.Equal(expectedRes, code) {
		t.Errorf("expected: %X got: %X", expectedRes, code)
	}
}
