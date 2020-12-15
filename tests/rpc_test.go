// This is a test utility for Ethermint's Web3 JSON-RPC services.
//
// To run these tests please first ensure you have the ethermintd running
// and have started the RPC service with `ethermintcli rest-server`.
//
// You can configure the desired HOST and MODE as well
package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	ethcmn "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	ethtypes "github.com/ethereum/go-ethereum/core/types"

	rpctypes "github.com/cosmos/ethermint/rpc/types"
	ethermint "github.com/cosmos/ethermint/types"
)

const (
	addrA         = "0xc94770007dda54cF92009BFF0dE90c06F603a09f"
	addrAStoreKey = 0
)

var (
	MODE       = os.Getenv("MODE")
	from       = []byte{}
	zeroString = "0x0"
)

func TestMain(m *testing.M) {
	if MODE != "rpc" {
		_, _ = fmt.Fprintln(os.Stdout, "Skipping RPC test")
		return
	}

	var err error
	from, err = GetAddress()
	if err != nil {
		fmt.Printf("failed to get account: %s\n", err)
		os.Exit(1)
	}

	// Start all tests
	code := m.Run()
	os.Exit(code)
}

func TestBlockBloom(t *testing.T) {
	hash := DeployTestContractWithFunction(t, from)
	receipt := WaitForReceipt(t, hash)

	number := receipt["blockNumber"].(string)
	param := []interface{}{number, false}
	rpcRes := Call(t, "eth_getBlockByNumber", param)

	block := make(map[string]interface{})
	err := json.Unmarshal(rpcRes.Result, &block)
	require.NoError(t, err)

	lb := HexToBigInt(t, block["logsBloom"].(string))
	require.NotEqual(t, big.NewInt(0), lb)
	require.Equal(t, hash.String(), block["transactions"].([]interface{})[0])
}

func TestEth_GetLogs_NoLogs(t *testing.T) {
	param := make([]map[string][]string, 1)
	param[0] = make(map[string][]string)
	param[0]["topics"] = []string{}
	rpcRes := Call(t, "eth_getLogs", param)
	require.NotNil(t, rpcRes)
	require.Nil(t, rpcRes.Error)

	var logs []*ethtypes.Log
	err := json.Unmarshal(rpcRes.Result, &logs)
	require.NoError(t, err)
	require.NotEmpty(t, logs)
}

func TestEth_GetLogs_Topics_AB(t *testing.T) {
	// TODO: this test passes on when run on its own, but fails when run with the other tests
	if testing.Short() {
		t.Skip("skipping TestEth_GetLogs_Topics_AB")
	}

	rpcRes := Call(t, "eth_blockNumber", []string{})

	var res hexutil.Uint64
	err := res.UnmarshalJSON(rpcRes.Result)
	require.NoError(t, err)

	param := make([]map[string]interface{}, 1)
	param[0] = make(map[string]interface{})
	param[0]["topics"] = []string{helloTopic, worldTopic}
	param[0]["fromBlock"] = res.String()

	hash := DeployTestContractWithFunction(t, from)
	WaitForReceipt(t, hash)

	rpcRes = Call(t, "eth_getLogs", param)

	var logs []*ethtypes.Log
	err = json.Unmarshal(rpcRes.Result, &logs)
	require.NoError(t, err)

	require.Equal(t, 1, len(logs))
}

func TestEth_GetTransactionCount(t *testing.T) {
	// TODO: this test passes on when run on its own, but fails when run with the other tests
	if testing.Short() {
		t.Skip("skipping TestEth_GetTransactionCount")
	}

	prev := GetNonce(t, "latest")
	SendTestTransaction(t, from)
	post := GetNonce(t, "latest")
	require.Equal(t, prev, post-1)
}

func TestEth_GetTransactionLogs(t *testing.T) {
	// TODO: this test passes on when run on its own, but fails when run with the other tests
	if testing.Short() {
		t.Skip("skipping TestEth_GetTransactionLogs")
	}

	hash, _ := DeployTestContract(t, from)

	param := []string{hash.String()}
	rpcRes := Call(t, "eth_getTransactionLogs", param)

	logs := new([]*ethtypes.Log)
	err := json.Unmarshal(rpcRes.Result, logs)
	require.NoError(t, err)
	require.Equal(t, 1, len(*logs))
}

func TestEth_protocolVersion(t *testing.T) {
	expectedRes := hexutil.Uint(ethermint.ProtocolVersion)

	rpcRes := Call(t, "eth_protocolVersion", []string{})

	var res hexutil.Uint
	err := res.UnmarshalJSON(rpcRes.Result)
	require.NoError(t, err)

	t.Logf("Got protocol version: %s\n", res.String())
	require.Equal(t, expectedRes, res, "expected: %s got: %s\n", expectedRes.String(), rpcRes.Result)
}

func TestEth_chainId(t *testing.T) {
	rpcRes := Call(t, "eth_chainId", []string{})

	var res hexutil.Uint
	err := res.UnmarshalJSON(rpcRes.Result)
	require.NoError(t, err)
	require.NotEqual(t, "0x0", res.String())
}

func TestEth_blockNumber(t *testing.T) {
	rpcRes := Call(t, "eth_blockNumber", []string{})

	var res hexutil.Uint64
	err := res.UnmarshalJSON(rpcRes.Result)
	require.NoError(t, err)

	t.Logf("Got block number: %s\n", res.String())
}

func TestEth_coinbase(t *testing.T) {
	zeroAddress := hexutil.Bytes(ethcmn.Address{}.Bytes())
	rpcRes := Call(t, "eth_coinbase", []string{})

	var res hexutil.Bytes
	err := res.UnmarshalJSON(rpcRes.Result)
	require.NoError(t, err)

	t.Logf("Got coinbase block proposer: %s\n", res.String())
	require.NotEqual(t, zeroAddress.String(), res.String(), "expected: not %s got: %s\n", zeroAddress.String(), res.String())
}

func TestEth_GetBalance(t *testing.T) {
	rpcRes := Call(t, "eth_getBalance", []string{addrA, zeroString})

	var res hexutil.Big
	err := res.UnmarshalJSON(rpcRes.Result)
	require.NoError(t, err)

	t.Logf("Got balance %s for %s\n", res.String(), addrA)

	// 0 if x == y; where x is res, y is 0
	if res.ToInt().Cmp(big.NewInt(0)) != 0 {
		t.Errorf("expected balance: %d, got: %s", 0, res.String())
	}
}

func TestEth_GetStorageAt(t *testing.T) {
	expectedRes := hexutil.Bytes{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	rpcRes := Call(t, "eth_getStorageAt", []string{addrA, fmt.Sprint(addrAStoreKey), zeroString})

	var storage hexutil.Bytes
	err := storage.UnmarshalJSON(rpcRes.Result)
	require.NoError(t, err)

	t.Logf("Got value [%X] for %s with key %X\n", storage, addrA, addrAStoreKey)

	require.True(t, bytes.Equal(storage, expectedRes), "expected: %d (%d bytes) got: %d (%d bytes)", expectedRes, len(expectedRes), storage, len(storage))
}

func TestEth_GetProof(t *testing.T) {
	params := make([]interface{}, 3)
	params[0] = addrA
	params[1] = []string{fmt.Sprint(addrAStoreKey)}
	params[2] = "latest"
	rpcRes := Call(t, "eth_getProof", params)
	require.NotNil(t, rpcRes)

	var accRes rpctypes.AccountResult
	err := json.Unmarshal(rpcRes.Result, &accRes)
	require.NoError(t, err)
	require.NotEmpty(t, accRes.AccountProof)
	require.NotEmpty(t, accRes.StorageProof)

	t.Logf("Got AccountResult %s", rpcRes.Result)
}

func TestEth_GetCode(t *testing.T) {
	expectedRes := hexutil.Bytes{}
	rpcRes := Call(t, "eth_getCode", []string{addrA, zeroString})

	var code hexutil.Bytes
	err := code.UnmarshalJSON(rpcRes.Result)

	require.NoError(t, err)

	t.Logf("Got code [%X] for %s\n", code, addrA)
	require.True(t, bytes.Equal(expectedRes, code), "expected: %X got: %X", expectedRes, code)
}

func TestEth_SendTransaction_Transfer(t *testing.T) {
	param := make([]map[string]string, 1)
	param[0] = make(map[string]string)
	param[0]["from"] = "0x" + fmt.Sprintf("%x", from)
	param[0]["to"] = "0x0000000000000000000000000000000012341234"
	param[0]["value"] = "0x16345785d8a0000"
	param[0]["gasLimit"] = "0x5208"
	param[0]["gasPrice"] = "0x55ae82600"

	rpcRes := Call(t, "eth_sendTransaction", param)

	var hash hexutil.Bytes
	err := json.Unmarshal(rpcRes.Result, &hash)
	require.NoError(t, err)

	receipt := WaitForReceipt(t, hash)
	require.NotNil(t, receipt)
	require.Equal(t, "0x1", receipt["status"].(string))
}

func TestEth_SendTransaction_ContractDeploy(t *testing.T) {
	param := make([]map[string]string, 1)
	param[0] = make(map[string]string)
	param[0]["from"] = "0x" + fmt.Sprintf("%x", from)
	param[0]["data"] = "0x6080604052348015600f57600080fd5b5060117f775a94827b8fd9b519d36cd827093c664f93347070a554f65e4a6f56cd73889860405160405180910390a2603580604b6000396000f3fe6080604052600080fdfea165627a7a723058206cab665f0f557620554bb45adf266708d2bd349b8a4314bdff205ee8440e3c240029"

	rpcRes := Call(t, "eth_sendTransaction", param)

	var hash hexutil.Bytes
	err := json.Unmarshal(rpcRes.Result, &hash)
	require.NoError(t, err)
}

func TestEth_NewFilter(t *testing.T) {
	param := make([]map[string][]string, 1)
	param[0] = make(map[string][]string)
	param[0]["topics"] = []string{"0x0000000000000000000000000000000000000000000000000000000012341234"}
	rpcRes := Call(t, "eth_newFilter", param)

	var ID string
	err := json.Unmarshal(rpcRes.Result, &ID)
	require.NoError(t, err)
}

func TestEth_NewBlockFilter(t *testing.T) {
	rpcRes := Call(t, "eth_newBlockFilter", []string{})

	var ID string
	err := json.Unmarshal(rpcRes.Result, &ID)
	require.NoError(t, err)
}

func TestEth_GetFilterChanges_BlockFilter(t *testing.T) {
	rpcRes := Call(t, "eth_newBlockFilter", []string{})

	var ID string
	err := json.Unmarshal(rpcRes.Result, &ID)
	require.NoError(t, err)

	time.Sleep(5 * time.Second)

	changesRes := Call(t, "eth_getFilterChanges", []string{ID})
	var hashes []ethcmn.Hash
	err = json.Unmarshal(changesRes.Result, &hashes)
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(hashes), 1)
}

func TestEth_GetFilterChanges_NoLogs(t *testing.T) {
	param := make([]map[string][]string, 1)
	param[0] = make(map[string][]string)
	param[0]["topics"] = []string{}
	rpcRes := Call(t, "eth_newFilter", param)

	var ID string
	err := json.Unmarshal(rpcRes.Result, &ID)
	require.NoError(t, err)

	changesRes := Call(t, "eth_getFilterChanges", []string{ID})

	var logs []*ethtypes.Log
	err = json.Unmarshal(changesRes.Result, &logs)
	require.NoError(t, err)
}

func TestEth_GetFilterChanges_WrongID(t *testing.T) {
	req, err := json.Marshal(CreateRequest("eth_getFilterChanges", []string{"0x1122334400000077"}))
	require.NoError(t, err)

	var rpcRes *Response
	time.Sleep(1 * time.Second)
	/* #nosec */
	res, err := http.Post(HOST, "application/json", bytes.NewBuffer(req))
	require.NoError(t, err)

	decoder := json.NewDecoder(res.Body)
	rpcRes = new(Response)
	err = decoder.Decode(&rpcRes)
	require.NoError(t, err)

	err = res.Body.Close()
	require.NoError(t, err)
	require.NotNil(t, "invalid filter ID", rpcRes.Error.Message)
}

func TestEth_GetTransactionReceipt(t *testing.T) {
	hash := SendTestTransaction(t, from)

	time.Sleep(time.Second * 5)

	param := []string{hash.String()}
	rpcRes := Call(t, "eth_getTransactionReceipt", param)
	require.Nil(t, rpcRes.Error)

	receipt := make(map[string]interface{})
	err := json.Unmarshal(rpcRes.Result, &receipt)
	require.NoError(t, err)
	require.NotEmpty(t, receipt)
	require.Equal(t, "0x1", receipt["status"].(string))
	require.Equal(t, []interface{}{}, receipt["logs"].([]interface{}))
}

func TestEth_GetTransactionReceipt_ContractDeployment(t *testing.T) {
	hash, _ := DeployTestContract(t, from)

	time.Sleep(time.Second * 5)

	param := []string{hash.String()}
	rpcRes := Call(t, "eth_getTransactionReceipt", param)

	receipt := make(map[string]interface{})
	err := json.Unmarshal(rpcRes.Result, &receipt)
	require.NoError(t, err)
	require.Equal(t, "0x1", receipt["status"].(string))

	require.NotEqual(t, ethcmn.Address{}.String(), receipt["contractAddress"].(string))
	require.NotNil(t, receipt["logs"])

}

func TestEth_GetFilterChanges_NoTopics(t *testing.T) {
	rpcRes := Call(t, "eth_blockNumber", []string{})

	var res hexutil.Uint64
	err := res.UnmarshalJSON(rpcRes.Result)
	require.NoError(t, err)

	param := make([]map[string]interface{}, 1)
	param[0] = make(map[string]interface{})
	param[0]["topics"] = []string{}
	param[0]["fromBlock"] = res.String()

	// instantiate new filter
	rpcRes = Call(t, "eth_newFilter", param)
	require.Nil(t, rpcRes.Error)
	var ID string
	err = json.Unmarshal(rpcRes.Result, &ID)
	require.NoError(t, err)

	// deploy contract, emitting some event
	DeployTestContract(t, from)

	// get filter changes
	changesRes := Call(t, "eth_getFilterChanges", []string{ID})

	var logs []*ethtypes.Log
	err = json.Unmarshal(changesRes.Result, &logs)
	require.NoError(t, err)
	require.Equal(t, 1, len(logs))
}

func TestEth_GetFilterChanges_Addresses(t *testing.T) {
	t.Skip()
	// TODO: need transaction receipts to determine contract deployment address
}

func TestEth_GetFilterChanges_BlockHash(t *testing.T) {
	t.Skip()
	// TODO: need transaction receipts to determine tx block
}

// hash of Hello event
var helloTopic = "0x775a94827b8fd9b519d36cd827093c664f93347070a554f65e4a6f56cd738898"

// world parameter in Hello event
var worldTopic = "0x0000000000000000000000000000000000000000000000000000000000000011"

// Tests topics case where there are topics in first two positions
func TestEth_GetFilterChanges_Topics_AB(t *testing.T) {
	time.Sleep(time.Second)

	rpcRes := Call(t, "eth_blockNumber", []string{})

	var res hexutil.Uint64
	err := res.UnmarshalJSON(rpcRes.Result)
	require.NoError(t, err)

	param := make([]map[string]interface{}, 1)
	param[0] = make(map[string]interface{})
	param[0]["topics"] = []string{helloTopic, worldTopic}
	param[0]["fromBlock"] = res.String()

	// instantiate new filter
	rpcRes = Call(t, "eth_newFilter", param)
	var ID string
	err = json.Unmarshal(rpcRes.Result, &ID)
	require.NoError(t, err, string(rpcRes.Result))

	DeployTestContractWithFunction(t, from)

	// get filter changes
	changesRes := Call(t, "eth_getFilterChanges", []string{ID})

	var logs []*ethtypes.Log
	err = json.Unmarshal(changesRes.Result, &logs)
	require.NoError(t, err)

	require.Equal(t, 1, len(logs))
}

func TestEth_GetFilterChanges_Topics_XB(t *testing.T) {
	rpcRes := Call(t, "eth_blockNumber", []string{})

	var res hexutil.Uint64
	err := res.UnmarshalJSON(rpcRes.Result)
	require.NoError(t, err)

	param := make([]map[string]interface{}, 1)
	param[0] = make(map[string]interface{})
	param[0]["topics"] = []interface{}{nil, worldTopic}
	param[0]["fromBlock"] = res.String()

	// instantiate new filter
	rpcRes = Call(t, "eth_newFilter", param)
	var ID string
	err = json.Unmarshal(rpcRes.Result, &ID)
	require.NoError(t, err)

	DeployTestContractWithFunction(t, from)

	// get filter changes
	changesRes := Call(t, "eth_getFilterChanges", []string{ID})

	var logs []*ethtypes.Log
	err = json.Unmarshal(changesRes.Result, &logs)
	require.NoError(t, err)

	require.Equal(t, 1, len(logs))
}

func TestEth_GetFilterChanges_Topics_XXC(t *testing.T) {
	t.Skip()
	// TODO: call test function, need tx receipts to determine contract address
}

func TestEth_PendingTransactionFilter(t *testing.T) {
	rpcRes := Call(t, "eth_newPendingTransactionFilter", []string{})

	var ID string
	err := json.Unmarshal(rpcRes.Result, &ID)
	require.NoError(t, err)

	for i := 0; i < 5; i++ {
		DeployTestContractWithFunction(t, from)
	}

	time.Sleep(10 * time.Second)

	// get filter changes
	changesRes := Call(t, "eth_getFilterChanges", []string{ID})
	require.NotNil(t, changesRes)

	var txs []*hexutil.Bytes
	err = json.Unmarshal(changesRes.Result, &txs)
	require.NoError(t, err, string(changesRes.Result))

	require.True(t, len(txs) >= 2, "could not get any txs", "changesRes.Result", string(changesRes.Result))
}

func TestEth_EstimateGas(t *testing.T) {
	param := make([]map[string]string, 1)
	param[0] = make(map[string]string)
	param[0]["from"] = "0x" + fmt.Sprintf("%x", from)
	param[0]["to"] = "0x1122334455667788990011223344556677889900"
	param[0]["value"] = "0x1"
	rpcRes := Call(t, "eth_estimateGas", param)
	require.NotNil(t, rpcRes)
	require.NotEmpty(t, rpcRes.Result)

	var gas string
	err := json.Unmarshal(rpcRes.Result, &gas)
	require.NoError(t, err, string(rpcRes.Result))

	require.Equal(t, "0xf54c", gas)
}

func TestEth_EstimateGas_ContractDeployment(t *testing.T) {
	bytecode := "0x608060405234801561001057600080fd5b5060117f775a94827b8fd9b519d36cd827093c664f93347070a554f65e4a6f56cd73889860405160405180910390a260d08061004d6000396000f3fe6080604052348015600f57600080fd5b506004361060285760003560e01c8063eb8ac92114602d575b600080fd5b606060048036036040811015604157600080fd5b8101908080359060200190929190803590602001909291905050506062565b005b8160008190555080827ff3ca124a697ba07e8c5e80bebcfcc48991fc16a63170e8a9206e30508960d00360405160405180910390a3505056fea265627a7a723158201d94d2187aaf3a6790527b615fcc40970febf0385fa6d72a2344848ebd0df3e964736f6c63430005110032"

	param := make([]map[string]string, 1)
	param[0] = make(map[string]string)
	param[0]["from"] = "0x" + fmt.Sprintf("%x", from)
	param[0]["data"] = bytecode

	rpcRes := Call(t, "eth_estimateGas", param)
	require.NotNil(t, rpcRes)
	require.NotEmpty(t, rpcRes.Result)

	var gas hexutil.Uint64
	err := json.Unmarshal(rpcRes.Result, &gas)
	require.NoError(t, err, string(rpcRes.Result))

	require.Equal(t, "0x1c2c4", gas.String())
}

func TestEth_GetBlockByNumber(t *testing.T) {
	param := []interface{}{"0x1", false}
	rpcRes := Call(t, "eth_getBlockByNumber", param)

	block := make(map[string]interface{})
	err := json.Unmarshal(rpcRes.Result, &block)
	require.NoError(t, err)
	require.Equal(t, "0x0", block["extraData"].(string))
	require.Equal(t, []interface{}{}, block["uncles"].([]interface{}))
}
