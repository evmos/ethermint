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
	"time"

	"github.com/cosmos/ethermint/version"
	"github.com/ethereum/go-ethereum/common/hexutil"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/require"
)

const (
	host          = "localhost"
	port          = 8545
	addrA         = "0xc94770007dda54cF92009BFF0dE90c06F603a09f"
	addrAStoreKey = 0
)

var addr = fmt.Sprintf("http://%s:%d", host, port)
var zeroString = "0x0"

type Request struct {
	Version string      `json:"jsonrpc"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params"`
	ID      int         `json:"id"`
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

func createRequest(method string, params interface{}) Request {
	return Request{
		Version: "2.0",
		Method:  method,
		Params:  params,
		ID:      1,
	}
}

func call(t *testing.T, method string, params interface{}) (*Response, error) {
	req, err := json.Marshal(createRequest(method, params))
	if err != nil {
		return nil, err
	}

	/* #nosec */
	res, err := http.Post(addr, "application/json", bytes.NewBuffer(req))
	if err != nil {
		t.Fatal(err)
	}

	decoder := json.NewDecoder(res.Body)
	var rpcRes *Response
	err = decoder.Decode(&rpcRes)
	if err != nil {
		t.Fatal(err)
	}

	if rpcRes.Error != nil {
		return nil, errors.New(rpcRes.Error.Message)
	}

	err = res.Body.Close()
	if err != nil {
		t.Fatal(err)
	}

	return rpcRes, nil
}

func TestEth_protocolVersion(t *testing.T) {
	expectedRes := hexutil.Uint(version.ProtocolVersion)

	rpcRes, err := call(t, "eth_protocolVersion", []string{})
	require.NoError(t, err)

	var res hexutil.Uint
	err = res.UnmarshalJSON(rpcRes.Result)
	require.NoError(t, err)

	t.Logf("Got protocol version: %s\n", res.String())
	require.Equal(t, expectedRes, res, "expected: %s got: %s\n", expectedRes.String(), rpcRes.Result)
}

func TestEth_blockNumber(t *testing.T) {
	rpcRes, err := call(t, "eth_blockNumber", []string{})
	require.NoError(t, err)

	var res hexutil.Uint64
	err = res.UnmarshalJSON(rpcRes.Result)
	require.NoError(t, err)

	t.Logf("Got block number: %s\n", res.String())
}

func TestEth_GetBalance(t *testing.T) {
	rpcRes, err := call(t, "eth_getBalance", []string{addrA, zeroString})
	require.NoError(t, err)

	var res hexutil.Big
	err = res.UnmarshalJSON(rpcRes.Result)
	require.NoError(t, err)

	t.Logf("Got balance %s for %s\n", res.String(), addrA)

	// 0 if x == y; where x is res, y is 0
	if res.ToInt().Cmp(big.NewInt(0)) != 0 {
		t.Errorf("expected balance: %d, got: %s", 0, res.String())
	}
}

func TestEth_GetStorageAt(t *testing.T) {
	expectedRes := hexutil.Bytes{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	rpcRes, err := call(t, "eth_getStorageAt", []string{addrA, string(addrAStoreKey), zeroString})
	require.NoError(t, err)

	var storage hexutil.Bytes
	err = storage.UnmarshalJSON(rpcRes.Result)
	require.NoError(t, err)

	t.Logf("Got value [%X] for %s with key %X\n", storage, addrA, addrAStoreKey)

	require.True(t, bytes.Equal(storage, expectedRes), "expected: %d (%d bytes) got: %d (%d bytes)", expectedRes, len(expectedRes), storage, len(storage))
}

func TestEth_GetCode(t *testing.T) {
	expectedRes := hexutil.Bytes{}
	rpcRes, err := call(t, "eth_getCode", []string{addrA, zeroString})
	require.NoError(t, err)

	var code hexutil.Bytes
	err = code.UnmarshalJSON(rpcRes.Result)

	require.NoError(t, err)

	t.Logf("Got code [%X] for %s\n", code, addrA)
	require.True(t, bytes.Equal(expectedRes, code), "expected: %X got: %X", expectedRes, code)
}

func getAddress(t *testing.T) []byte {
	rpcRes, err := call(t, "eth_accounts", []string{})
	require.NoError(t, err)

	var res []hexutil.Bytes
	err = json.Unmarshal(rpcRes.Result, &res)
	require.NoError(t, err)

	return res[0]
}

func TestEth_SendTransaction(t *testing.T) {
	from := getAddress(t)

	param := make([]map[string]string, 1)
	param[0] = make(map[string]string)
	param[0]["from"] = "0x" + fmt.Sprintf("%x", from)
	param[0]["data"] = "0x6080604052348015600f57600080fd5b5060117f775a94827b8fd9b519d36cd827093c664f93347070a554f65e4a6f56cd73889860405160405180910390a2603580604b6000396000f3fe6080604052600080fdfea165627a7a723058206cab665f0f557620554bb45adf266708d2bd349b8a4314bdff205ee8440e3c240029"

	rpcRes, err := call(t, "eth_sendTransaction", param)
	require.NoError(t, err)

	var hash hexutil.Bytes
	err = json.Unmarshal(rpcRes.Result, &hash)
	require.NoError(t, err)
}

func TestEth_NewFilter(t *testing.T) {
	param := make([]map[string][]string, 1)
	param[0] = make(map[string][]string)
	param[0]["topics"] = []string{"0x0000000000000000000000000000000000000000000000000000000012341234"}
	rpcRes, err := call(t, "eth_newFilter", param)
	require.NoError(t, err)

	var ID hexutil.Bytes
	err = json.Unmarshal(rpcRes.Result, &ID)
	require.NoError(t, err)
}

func TestEth_NewBlockFilter(t *testing.T) {
	rpcRes, err := call(t, "eth_newBlockFilter", []string{})
	require.NoError(t, err)

	var ID hexutil.Bytes
	err = json.Unmarshal(rpcRes.Result, &ID)
	require.NoError(t, err)
}

func TestEth_GetFilterChanges_NoLogs(t *testing.T) {
	param := make([]map[string][]string, 1)
	param[0] = make(map[string][]string)
	param[0]["topics"] = []string{}
	rpcRes, err := call(t, "eth_newFilter", param)
	require.NoError(t, err)

	var ID hexutil.Bytes
	err = json.Unmarshal(rpcRes.Result, &ID)
	require.NoError(t, err)

	changesRes, err := call(t, "eth_getFilterChanges", []string{ID.String()})
	require.NoError(t, err)

	var logs []*ethtypes.Log
	err = json.Unmarshal(changesRes.Result, &logs)
	require.NoError(t, err)
}

func TestEth_GetFilterChanges_WrongID(t *testing.T) {
	_, err := call(t, "eth_getFilterChanges", []string{"0x1122334400000077"})
	require.NotNil(t, err)
}

// deployTestContract deploys a contract that emits an event in the constructor
func deployTestContract(t *testing.T) hexutil.Bytes {
	from := getAddress(t)

	param := make([]map[string]string, 1)
	param[0] = make(map[string]string)
	param[0]["from"] = "0x" + fmt.Sprintf("%x", from)
	param[0]["data"] = "0x6080604052348015600f57600080fd5b5060117f775a94827b8fd9b519d36cd827093c664f93347070a554f65e4a6f56cd73889860405160405180910390a2603580604b6000396000f3fe6080604052600080fdfea165627a7a723058206cab665f0f557620554bb45adf266708d2bd349b8a4314bdff205ee8440e3c240029"

	rpcRes, err := call(t, "eth_sendTransaction", param)
	require.NoError(t, err)

	var hash hexutil.Bytes
	err = json.Unmarshal(rpcRes.Result, &hash)
	require.NoError(t, err)

	return hash
}

func TestEth_GetTransactionReceipt(t *testing.T) {
	hash := deployTestContract(t)

	time.Sleep(time.Second * 2)

	param := []string{hash.String()}
	rpcRes, err := call(t, "eth_getTransactionReceipt", param)
	require.NoError(t, err)

	t.Log(rpcRes.Result)
	// TODO: why does this not return a receipt?
}

func TestEth_GetTxLogs(t *testing.T) {
	// currently fails due to eth_sendTransaction returning the tendermint hash,
	// while the logs are stored in the db using the ethereum hash
	t.Skip()
	hash := deployTestContract(t)

	time.Sleep(time.Second * 5)

	param := []string{hash.String()}
	rpcRes, err := call(t, "eth_getTxLogs", param)
	require.NoError(t, err)

	logs := new([]*ethtypes.Log)
	err = json.Unmarshal(rpcRes.Result, logs)
	require.NoError(t, err)

	require.Equal(t, 1, len(*logs))
	t.Log((*logs)[0])
	time.Sleep(time.Second)
}

func TestEth_GetFilterChanges_NoTopics(t *testing.T) {
	rpcRes, err := call(t, "eth_blockNumber", []string{})
	require.NoError(t, err)

	var res hexutil.Uint64
	err = res.UnmarshalJSON(rpcRes.Result)
	require.NoError(t, err)

	param := make([]map[string]interface{}, 1)
	param[0] = make(map[string]interface{})
	param[0]["topics"] = []string{}
	param[0]["fromBlock"] = res.String()
	param[0]["toBlock"] = zeroString // latest

	// deploy contract, emitting some event
	deployTestContract(t)

	rpcRes, err = call(t, "eth_newFilter", param)
	require.NoError(t, err)

	var ID hexutil.Bytes
	err = json.Unmarshal(rpcRes.Result, &ID)
	require.NoError(t, err)

	time.Sleep(time.Second)

	// get filter changes
	changesRes, err := call(t, "eth_getFilterChanges", []string{ID.String()})
	require.NoError(t, err)

	var logs []*ethtypes.Log
	err = json.Unmarshal(changesRes.Result, &logs)
	require.NoError(t, err)

	require.Equal(t, 1, len(logs))
	time.Sleep(time.Second)

	//t.Log(logs[0])
	// TODO: why is the tx hash in the log not the same as the tx hash of the transaction?
	//require.Equal(t, logs[0].TxHash, common.BytesToHash(hash))
}

func TestEth_GetFilterChanges_Addresses(t *testing.T) {
	// TODO: need transaction receipts to determine contract deployment address
}

func TestEth_GetFilterChanges_BlockHash(t *testing.T) {
	// TODO: need transaction receipts to determine tx block
}

// hash of Hello event
var helloTopic = "0x775a94827b8fd9b519d36cd827093c664f93347070a554f65e4a6f56cd738898"

// world parameter in Hello event
var worldTopic = "0x0000000000000000000000000000000000000000000000000000000000000011"

func deployTestContractWithFunction(t *testing.T) hexutil.Bytes {
	// pragma solidity ^0.5.1;

	// contract Test {
	//     event Hello(uint256 indexed world);
	//     event Test(uint256 indexed a, uint256 indexed b);

	//     constructor() public {
	//         emit Hello(17);
	//     }

	//     function test(uint256 a, uint256 b) public {
	//         emit Test(a, b);
	//     }
	// }

	bytecode := "0x608060405234801561001057600080fd5b5060117f775a94827b8fd9b519d36cd827093c664f93347070a554f65e4a6f56cd73889860405160405180910390a260c98061004d6000396000f3fe6080604052348015600f57600080fd5b506004361060285760003560e01c8063eb8ac92114602d575b600080fd5b606060048036036040811015604157600080fd5b8101908080359060200190929190803590602001909291905050506062565b005b80827f91916a5e2c96453ddf6b585497262675140eb9f7a774095fb003d93e6dc6921660405160405180910390a3505056fea265627a7a72315820ef746422e676b3ed22147cd771a6f689e7c33ef17bf5cd91921793b5dd01e3e064736f6c63430005110032"

	from := getAddress(t)

	param := make([]map[string]string, 1)
	param[0] = make(map[string]string)
	param[0]["from"] = "0x" + fmt.Sprintf("%x", from)
	param[0]["data"] = bytecode

	rpcRes, err := call(t, "eth_sendTransaction", param)
	require.NoError(t, err)

	var hash hexutil.Bytes
	err = json.Unmarshal(rpcRes.Result, &hash)
	require.NoError(t, err)

	return hash
}

// Tests topics case where there are topics in first two positions
func TestEth_GetFilterChanges_Topics_AB(t *testing.T) {
	time.Sleep(time.Second)

	rpcRes, err := call(t, "eth_blockNumber", []string{})
	require.NoError(t, err)

	var res hexutil.Uint64
	err = res.UnmarshalJSON(rpcRes.Result)
	require.NoError(t, err)

	param := make([]map[string]interface{}, 1)
	param[0] = make(map[string]interface{})
	param[0]["topics"] = []string{helloTopic, worldTopic}
	param[0]["fromBlock"] = res.String()
	param[0]["toBlock"] = zeroString // latest

	deployTestContractWithFunction(t)

	rpcRes, err = call(t, "eth_newFilter", param)
	require.NoError(t, err)

	var ID hexutil.Bytes
	err = json.Unmarshal(rpcRes.Result, &ID)
	require.NoError(t, err)

	time.Sleep(time.Second * 2)

	// get filter changes
	changesRes, err := call(t, "eth_getFilterChanges", []string{ID.String()})
	require.NoError(t, err)

	var logs []*ethtypes.Log
	err = json.Unmarshal(changesRes.Result, &logs)
	require.NoError(t, err)

	require.Equal(t, 1, len(logs))
	time.Sleep(time.Second * 2)
}

func TestEth_GetFilterChanges_Topics_XB(t *testing.T) {
	rpcRes, err := call(t, "eth_blockNumber", []string{})
	require.NoError(t, err)

	var res hexutil.Uint64
	err = res.UnmarshalJSON(rpcRes.Result)
	require.NoError(t, err)

	param := make([]map[string]interface{}, 1)
	param[0] = make(map[string]interface{})
	param[0]["topics"] = []interface{}{nil, worldTopic}
	param[0]["fromBlock"] = res.String()
	param[0]["toBlock"] = "0x0" // latest

	deployTestContractWithFunction(t)

	rpcRes, err = call(t, "eth_newFilter", param)
	require.NoError(t, err)

	var ID hexutil.Bytes
	err = json.Unmarshal(rpcRes.Result, &ID)
	require.NoError(t, err)

	time.Sleep(time.Second * 2)

	// get filter changes
	changesRes, err := call(t, "eth_getFilterChanges", []string{ID.String()})
	require.NoError(t, err)

	var logs []*ethtypes.Log
	err = json.Unmarshal(changesRes.Result, &logs)
	require.NoError(t, err)

	require.Equal(t, 1, len(logs))
	time.Sleep(time.Second)
}

func TestEth_GetFilterChanges_Topics_XXC(t *testing.T) {
	// TODO: call test function, need tx receipts to determine contract address
}

func TestEth_GetLogs_NoLogs(t *testing.T) {
	param := make([]map[string][]string, 1)
	param[0] = make(map[string][]string)
	param[0]["topics"] = []string{}
	_, err := call(t, "eth_getLogs", param)
	require.NoError(t, err)
}

func TestEth_GetLogs_Topics_AB(t *testing.T) {
	rpcRes, err := call(t, "eth_blockNumber", []string{})
	require.NoError(t, err)

	var res hexutil.Uint64
	err = res.UnmarshalJSON(rpcRes.Result)
	require.NoError(t, err)

	param := make([]map[string]interface{}, 1)
	param[0] = make(map[string]interface{})
	param[0]["topics"] = []string{helloTopic, worldTopic}
	param[0]["fromBlock"] = res.String()
	param[0]["toBlock"] = zeroString // latest

	deployTestContractWithFunction(t)

	rpcRes, err = call(t, "eth_getLogs", param)
	require.NoError(t, err)

	var logs []*ethtypes.Log
	err = json.Unmarshal(rpcRes.Result, &logs)
	require.NoError(t, err)

	require.Equal(t, 1, len(logs))
}
