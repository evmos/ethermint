// This is a test utility for Ethermint's Web3 JSON-RPC services.
//
// To run these tests please first ensure you have the ethermintd running
// and have started the RPC service with `ethermintd rest-server`.
//
// You can configure the desired HOST and MODE as well
package rpc

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	rpctypes "github.com/tharsis/ethermint/rpc/ethereum/types"
	ethermint "github.com/tharsis/ethermint/types"
	evmtypes "github.com/tharsis/ethermint/x/evm/types"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
)

const (
	addrA         = "0xc94770007dda54cF92009BFF0dE90c06F603a09f"
	addrAStoreKey = 0
)

var (
	MODE       = os.Getenv("MODE")
	zeroString = "0x0"
	from       = []byte{}
)

func TestMain(m *testing.M) {
	if MODE != "rpc" {
		_, _ = fmt.Fprintln(os.Stdout, "Skipping RPC test")
		return
	}

	if HOST == "" {
		HOST = "http://localhost:8545"
	}

	var err error
	from, err = getAddress()
	if err != nil {
		fmt.Printf("failed to get account: %s\n", err)
		os.Exit(1)
	}

	// Start all tests
	code := m.Run()
	os.Exit(code)
}

func getAddress() ([]byte, error) {
	rpcRes, err := callWithError("eth_accounts", []string{})
	if err != nil {
		return nil, err
	}

	var res []hexutil.Bytes
	err = json.Unmarshal(rpcRes.Result, &res)
	if err != nil {
		return nil, err
	}

	return res[0], nil
}

func createRequest(method string, params interface{}) Request {
	return Request{
		Version: "2.0",
		Method:  method,
		Params:  params,
		ID:      1,
	}
}

func call(t *testing.T, method string, params interface{}) *Response {
	req, err := json.Marshal(createRequest(method, params))
	require.NoError(t, err)

	time.Sleep(1 * time.Second)
	/* #nosec */
	res, err := http.Post(HOST, "application/json", bytes.NewBuffer(req))
	require.NoError(t, err)

	decoder := json.NewDecoder(res.Body)
	rpcRes := new(Response)
	err = decoder.Decode(&rpcRes)
	require.NoError(t, err)

	err = res.Body.Close()
	require.NoError(t, err)
	require.Nil(t, rpcRes.Error)

	return rpcRes
}

func callWithError(method string, params interface{}) (*Response, error) {
	req, err := json.Marshal(createRequest(method, params))
	if err != nil {
		return nil, err
	}

	time.Sleep(1 * time.Second)
	/* #nosec */
	res, err := http.Post(HOST, "application/json", bytes.NewBuffer(req))
	if err != nil {
		return nil, err
	}

	decoder := json.NewDecoder(res.Body)
	rpcRes := new(Response)
	err = decoder.Decode(&rpcRes)
	if err != nil {
		return nil, err
	}

	err = res.Body.Close()
	if err != nil {
		return nil, err
	}

	if rpcRes.Error != nil {
		return nil, fmt.Errorf(rpcRes.Error.Message)
	}

	return rpcRes, nil
}

// turns a 0x prefixed hex string to a big.Int
func hexToBigInt(t *testing.T, in string) *big.Int {
	s := in[2:]
	b, err := hex.DecodeString(s)
	require.NoError(t, err)
	return big.NewInt(0).SetBytes(b)
}

func TestBlockBloom(t *testing.T) {
	hash := deployTestContractWithFunction(t)
	receipt := waitForReceipt(t, hash)

	number := receipt["blockNumber"].(string)
	param := []interface{}{number, false}
	rpcRes := call(t, "eth_getBlockByNumber", param)

	block := make(map[string]interface{})
	err := json.Unmarshal(rpcRes.Result, &block)
	require.NoError(t, err)

	lb := hexToBigInt(t, block["logsBloom"].(string))
	require.NotEqual(t, big.NewInt(0), lb)
	require.Equal(t, hash.String(), block["transactions"].([]interface{})[0])
}

func TestEth_GetLogs_NoLogs(t *testing.T) {
	param := make([]map[string][]string, 1)
	param[0] = make(map[string][]string)
	param[0]["topics"] = []string{}
	call(t, "eth_getLogs", param)
}

func TestEth_GetLogs_Topics_AB(t *testing.T) {
	// TODO: this test passes on when run on its own, but fails when run with the other tests
	if testing.Short() {
		t.Skip("skipping TestEth_GetLogs_Topics_AB")
	}

	rpcRes := call(t, "eth_blockNumber", []string{})

	var res hexutil.Uint64
	err := res.UnmarshalJSON(rpcRes.Result)
	require.NoError(t, err)

	param := make([]map[string]interface{}, 1)
	param[0] = make(map[string]interface{})
	param[0]["topics"] = []string{helloTopic, worldTopic}
	param[0]["fromBlock"] = res.String()

	hash := deployTestContractWithFunction(t)
	waitForReceipt(t, hash)

	rpcRes = call(t, "eth_getLogs", param)

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

	prev := getNonce(t)
	sendTestTransaction(t)
	post := getNonce(t)
	require.Equal(t, prev, post-1)
}

func TestETH_GetBlockTransactionCountByHash(t *testing.T) {
	txHash := sendTestTransaction(t)

	receipt := waitForReceipt(t, txHash)
	require.NotNil(t, receipt, "transaction failed")
	require.Equal(t, "0x1", receipt["status"].(string))

	blockHash := receipt["blockHash"].(string)
	param := []string{blockHash}
	rpcRes := call(t, "eth_getBlockTransactionCountByHash", param)

	var res hexutil.Uint
	err := res.UnmarshalJSON(rpcRes.Result)
	require.NoError(t, err)
	require.Equal(t, "0x1", res.String())
}

func TestETH_GetBlockTransactionCountByHash_BlockHashNotFound(t *testing.T) {
	anyBlockHash := "0xb3b20624f8f0f86eb50dd04688409e5cea4bd02d700bf6e79e9384d47d6a5a35"
	param := []string{anyBlockHash}
	rpcRes := call(t, "eth_getBlockTransactionCountByHash", param)

	var result interface{}
	err := json.Unmarshal(rpcRes.Result, &result)
	require.NoError(t, err)
	require.Nil(t, result)
}

func TestETH_GetTransactionByBlockHashAndIndex(t *testing.T) {
	txHash := sendTestTransaction(t)

	receipt := waitForReceipt(t, txHash)
	require.NotNil(t, receipt, "transaction failed")
	require.Equal(t, "0x1", receipt["status"].(string))
	blockHash := receipt["blockHash"].(string)

	param := []string{blockHash, "0x0"}
	rpcRes := call(t, "eth_getTransactionByBlockHashAndIndex", param)

	tx := make(map[string]interface{})
	err := json.Unmarshal(rpcRes.Result, &tx)
	require.NoError(t, err)
	require.NotNil(t, tx)
	require.Equal(t, blockHash, tx["blockHash"].(string))
	require.Equal(t, "0x0", tx["transactionIndex"].(string))
}

func TestETH_GetTransactionByBlockHashAndIndex_BlockHashNotFound(t *testing.T) {
	anyBlockHash := "0xb3b20624f8f0f86eb50dd04688409e5cea4bd02d700bf6e79e9384d47d6a5a35"

	param := []string{anyBlockHash, "0x0"}
	rpcRes := call(t, "eth_getTransactionByBlockHashAndIndex", param)

	var result interface{}
	err := json.Unmarshal(rpcRes.Result, &result)
	require.NoError(t, err)
	require.Nil(t, result)
}

func TestEth_GetTransactionLogs(t *testing.T) {
	// TODO: this test passes on when run on its own, but fails when run with the other tests
	if testing.Short() {
		t.Skip("skipping TestEth_GetTransactionLogs")
	}

	hash, _ := deployTestContract(t)

	param := []string{hash.String()}
	rpcRes := call(t, "eth_getTransactionLogs", param)

	logs := new([]*ethtypes.Log)
	err := json.Unmarshal(rpcRes.Result, logs)
	require.NoError(t, err)
	require.Equal(t, 1, len(*logs))
}

func TestEth_protocolVersion(t *testing.T) {
	expectedRes := hexutil.Uint(ethermint.ProtocolVersion)

	rpcRes := call(t, "eth_protocolVersion", []string{})

	var res hexutil.Uint
	err := res.UnmarshalJSON(rpcRes.Result)
	require.NoError(t, err)

	t.Logf("Got protocol version: %s\n", res.String())
	require.Equal(t, expectedRes, res, "expected: %s got: %s\n", expectedRes.String(), rpcRes.Result)
}

func TestEth_chainId(t *testing.T) {
	rpcRes := call(t, "eth_chainId", []string{})

	var res hexutil.Uint
	err := res.UnmarshalJSON(rpcRes.Result)
	require.NoError(t, err)
	require.NotEqual(t, "0x0", res.String())
}

func TestEth_blockNumber(t *testing.T) {
	rpcRes := call(t, "eth_blockNumber", []string{})

	var res hexutil.Uint64
	err := res.UnmarshalJSON(rpcRes.Result)
	require.NoError(t, err)

	t.Logf("Got block number: %s\n", res.String())
}

func TestEth_coinbase(t *testing.T) {
	zeroAddress := hexutil.Bytes(common.Address{}.Bytes())
	rpcRes := call(t, "eth_coinbase", []string{})

	var res hexutil.Bytes
	err := res.UnmarshalJSON(rpcRes.Result)
	require.NoError(t, err)

	t.Logf("Got coinbase block proposer: %s\n", res.String())
	require.NotEqual(t, zeroAddress.String(), res.String(), "expected: not %s got: %s\n", zeroAddress.String(), res.String())
}

func TestEth_GetBalance(t *testing.T) {
	rpcRes := call(t, "eth_getBalance", []string{addrA, zeroString})

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
	rpcRes := call(t, "eth_getStorageAt", []string{addrA, fmt.Sprint(addrAStoreKey), zeroString})

	var storage hexutil.Bytes
	err := storage.UnmarshalJSON(rpcRes.Result)
	require.NoError(t, err)

	t.Logf("Got value [%X] for %s with key %X\n", storage, addrA, addrAStoreKey)

	require.True(t, bytes.Equal(storage, expectedRes), "expected: %d (%d bytes) got: %d (%d bytes)", expectedRes, len(expectedRes), storage, len(storage))
}

func TestEth_GetProof(t *testing.T) {
	rpcRes := call(t, "eth_sendTransaction", makeEthTxParam())

	var hash hexutil.Bytes
	err := json.Unmarshal(rpcRes.Result, &hash)
	require.NoError(t, err)

	receipt := waitForReceipt(t, hash)
	require.NotNil(t, receipt)
	require.Equal(t, "0x1", receipt["status"].(string))

	params := make([]interface{}, 3)
	params[0] = addrA
	params[1] = []string{fmt.Sprint(addrAStoreKey)}
	params[2] = "latest"
	rpcRes = call(t, "eth_getProof", params)
	require.NotNil(t, rpcRes)

	var accRes rpctypes.AccountResult
	err = json.Unmarshal(rpcRes.Result, &accRes)
	require.NoError(t, err)
	require.NotEmpty(t, accRes.AccountProof)
	require.NotEmpty(t, accRes.StorageProof)

	t.Logf("Got AccountResult %s", rpcRes.Result)
}

func TestEth_GetCode(t *testing.T) {
	expectedRes := hexutil.Bytes{}
	rpcRes := call(t, "eth_getCode", []string{addrA, zeroString})

	var code hexutil.Bytes
	err := code.UnmarshalJSON(rpcRes.Result)

	require.NoError(t, err)

	t.Logf("Got code [%X] for %s\n", code, addrA)
	require.True(t, bytes.Equal(expectedRes, code), "expected: %X got: %X", expectedRes, code)
}

func TestEth_SendTransaction_Transfer(t *testing.T) {
	rpcRes := call(t, "eth_sendTransaction", makeEthTxParam())

	var hash hexutil.Bytes
	err := json.Unmarshal(rpcRes.Result, &hash)
	require.NoError(t, err)

	receipt := waitForReceipt(t, hash)
	require.NotNil(t, receipt)
	require.Equal(t, "0x1", receipt["status"].(string))
}

func TestEth_SendTransaction_ContractDeploy(t *testing.T) {
	param := makeTestContractDeployParam(t, true)
	rpcRes, err := callWithError("eth_sendTransaction", param)
	require.NoError(t, err)

	var hash hexutil.Bytes
	err = json.Unmarshal(rpcRes.Result, &hash)
	require.NoError(t, err)
}

func TestEth_SendTransaction_ContractDeploy_no_gas_param(t *testing.T) {
	param := makeTestContractDeployParam(t, false)
	_, err := callWithError("eth_sendTransaction", param)
	// server returns internal error.
	require.Error(t, err)
}

func TestEth_NewFilter(t *testing.T) {
	param := make([]map[string][]string, 1)
	param[0] = make(map[string][]string)
	param[0]["topics"] = []string{"0x0000000000000000000000000000000000000000000000000000000012341234"}
	rpcRes := call(t, "eth_newFilter", param)

	var ID string
	err := json.Unmarshal(rpcRes.Result, &ID)
	require.NoError(t, err)
}

func TestEth_NewBlockFilter(t *testing.T) {
	rpcRes := call(t, "eth_newBlockFilter", []string{})

	var ID string
	err := json.Unmarshal(rpcRes.Result, &ID)
	require.NoError(t, err)
}

func TestEth_GetFilterChanges_BlockFilter(t *testing.T) {
	rpcRes := call(t, "eth_newBlockFilter", []string{})

	var ID string
	err := json.Unmarshal(rpcRes.Result, &ID)
	require.NoError(t, err)

	txHash := sendTestTransaction(t)
	receipt := waitForReceipt(t, txHash)
	require.NotNil(t, receipt, "transaction failed")
	require.Equal(t, "0x1", receipt["status"].(string))

	changesRes := call(t, "eth_getFilterChanges", []string{ID})
	var hashes []common.Hash
	err = json.Unmarshal(changesRes.Result, &hashes)
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(hashes), 1)
}

func TestEth_GetFilterChanges_NoLogs(t *testing.T) {
	param := make([]map[string][]string, 1)
	param[0] = make(map[string][]string)
	param[0]["topics"] = []string{}
	rpcRes := call(t, "eth_newFilter", param)

	var ID string
	err := json.Unmarshal(rpcRes.Result, &ID)
	require.NoError(t, err)

	changesRes := call(t, "eth_getFilterChanges", []string{ID})

	var logs []*ethtypes.Log
	err = json.Unmarshal(changesRes.Result, &logs)
	require.NoError(t, err)
}

func TestEth_GetFilterChanges_WrongID(t *testing.T) {
	req, err := json.Marshal(createRequest("eth_getFilterChanges", []string{"0x1122334400000077"}))
	require.NoError(t, err)

	/* #nosec */
	res, err := http.Post(HOST, "application/json", bytes.NewBuffer(req))
	require.NoError(t, err)

	decoder := json.NewDecoder(res.Body)
	rpcRes := new(Response)
	err = decoder.Decode(&rpcRes)
	require.NoError(t, err)

	err = res.Body.Close()
	require.NoError(t, err)
	require.NotNil(t, "invalid filter ID", rpcRes.Error.Message)
}

// sendTestTransaction sends a dummy transaction
func sendTestTransaction(t *testing.T) hexutil.Bytes {
	gasPrice := GetGasPrice(t)

	param := make([]map[string]string, 1)
	param[0] = make(map[string]string)
	param[0]["from"] = "0x" + fmt.Sprintf("%x", from)
	param[0]["to"] = "0x1122334455667788990011223344556677889900"
	param[0]["value"] = "0x1"
	param[0]["gasPrice"] = gasPrice
	rpcRes := call(t, "eth_sendTransaction", param)

	var hash hexutil.Bytes
	err := json.Unmarshal(rpcRes.Result, &hash)
	require.NoError(t, err)
	return hash
}

func TestEth_GetTransactionReceipt(t *testing.T) {
	hash := sendTestTransaction(t)

	receipt := waitForReceipt(t, hash)

	require.NotNil(t, receipt, "transaction failed")
	require.Equal(t, "0x1", receipt["status"].(string))
	require.Equal(t, []interface{}{}, receipt["logs"].([]interface{}))
}

// deployTestERC20Contract deploys a contract that emits an event in the constructor
func deployTestERC20Contract(t *testing.T) common.Address {
	gasPrice := GetGasPrice(t)

	param := make([]map[string]string, 1)
	param[0] = make(map[string]string)
	param[0]["from"] = "0x" + fmt.Sprintf("%x", from)

	ctorArgs, err := evmtypes.ERC20Contract.ABI.Pack("", common.BytesToAddress(from), big.NewInt(100000000))
	require.NoError(t, err)
	data := append(evmtypes.ERC20Contract.Bin, ctorArgs...)
	param[0]["data"] = hexutil.Encode(data)

	param[0]["gas"] = "0x200000"
	param[0]["gasPrice"] = gasPrice

	rpcRes := call(t, "eth_sendTransaction", param)

	var hash hexutil.Bytes
	err = json.Unmarshal(rpcRes.Result, &hash)
	require.NoError(t, err)

	receipt := expectSuccessReceipt(t, hash)
	contractAddress := common.HexToAddress(receipt["contractAddress"].(string))
	require.NotEqual(t, common.Address{}, contractAddress)

	require.NotNil(t, receipt["logs"])

	return contractAddress
}

// sendTestERC20Transaction sends a typical erc20 transfer transaction
func sendTestERC20Transaction(t *testing.T, contract common.Address, amount *big.Int) hexutil.Bytes {
	// transfer
	gasPrice := GetGasPrice(t)
	param := make([]map[string]string, 1)
	param[0] = make(map[string]string)
	param[0]["from"] = "0x" + fmt.Sprintf("%x", from)
	param[0]["to"] = contract.Hex()
	data, err := evmtypes.ERC20Contract.ABI.Pack("transfer", common.BigToAddress(big.NewInt(1)), amount)
	require.NoError(t, err)
	param[0]["data"] = hexutil.Encode(data)
	param[0]["gas"] = "0x50000"
	param[0]["gasPrice"] = gasPrice

	rpcRes := call(t, "eth_sendTransaction", param)

	var hash hexutil.Bytes
	err = json.Unmarshal(rpcRes.Result, &hash)
	require.NoError(t, err)
	return hash
}

func TestEth_GetTransactionReceipt_ERC20Transfer(t *testing.T) {
	// deploy erc20 contract
	contract := deployTestERC20Contract(t)
	amount := big.NewInt(10)
	hash := sendTestERC20Transaction(t, contract, amount)
	receipt := expectSuccessReceipt(t, hash)

	require.Equal(t, 1, len(receipt["logs"].([]interface{})))
	log := receipt["logs"].([]interface{})[0].(map[string]interface{})

	require.Equal(t, contract, common.HexToAddress(log["address"].(string)))

	valueBz, err := hexutil.Decode(log["data"].(string))
	require.NoError(t, err)
	require.Equal(t, amount, big.NewInt(0).SetBytes(valueBz))

	require.Equal(t, false, log["removed"].(bool))
	require.Equal(t, "0x0", log["logIndex"].(string))
	require.Equal(t, "0x0", log["transactionIndex"].(string))

	expectedTopics := []interface{}{
		"0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef",
		"0x000000000000000000000000" + fmt.Sprintf("%x", from),
		"0x0000000000000000000000000000000000000000000000000000000000000001",
	}
	require.Equal(t, expectedTopics, log["topics"].([]interface{}))
}

// deployTestContract deploys a contract that emits an event in the constructor
func deployTestContract(t *testing.T) (hexutil.Bytes, map[string]interface{}) {
	gasPrice := GetGasPrice(t)
	param := make([]map[string]string, 1)
	param[0] = make(map[string]string)
	param[0]["from"] = "0x" + fmt.Sprintf("%x", from)
	param[0]["data"] = "0x6080604052348015600f57600080fd5b5060117f775a94827b8fd9b519d36cd827093c664f93347070a554f65e4a6f56cd73889860405160405180910390a2603580604b6000396000f3fe6080604052600080fdfea165627a7a723058206cab665f0f557620554bb45adf266708d2bd349b8a4314bdff205ee8440e3c240029"
	param[0]["gas"] = "0x200000"
	param[0]["gasPrice"] = gasPrice

	rpcRes := call(t, "eth_sendTransaction", param)

	var hash hexutil.Bytes
	err := json.Unmarshal(rpcRes.Result, &hash)
	require.NoError(t, err)

	receipt := waitForReceipt(t, hash)
	require.NotNil(t, receipt, "transaction failed")
	require.Equal(t, "0x1", receipt["status"].(string))

	return hash, receipt
}

func TestEth_GetTransactionReceipt_ContractDeployment(t *testing.T) {
	hash, _ := deployTestContract(t)

	param := []string{hash.String()}
	rpcRes := call(t, "eth_getTransactionReceipt", param)

	receipt := make(map[string]interface{})
	err := json.Unmarshal(rpcRes.Result, &receipt)
	require.NoError(t, err)
	require.Equal(t, "0x1", receipt["status"].(string))

	require.NotEqual(t, common.Address{}.String(), receipt["contractAddress"].(string))
	require.NotNil(t, receipt["logs"])
}

func getTransactionReceipt(t *testing.T, hash hexutil.Bytes) map[string]interface{} {
	param := []string{hash.String()}
	rpcRes := call(t, "eth_getTransactionReceipt", param)

	receipt := make(map[string]interface{})
	err := json.Unmarshal(rpcRes.Result, &receipt)
	require.NoError(t, err)

	return receipt
}

func waitForReceipt(t *testing.T, hash hexutil.Bytes) map[string]interface{} {
	timeout := time.After(12 * time.Second)
	ticker := time.Tick(500 * time.Millisecond)

	for {
		select {
		case <-timeout:
			return nil
		case <-ticker:
			receipt := getTransactionReceipt(t, hash)
			if receipt != nil {
				return receipt
			}
		}
	}
}

func expectSuccessReceipt(t *testing.T, hash hexutil.Bytes) map[string]interface{} {
	receipt := waitForReceipt(t, hash)
	require.NotNil(t, receipt, "transaction failed")
	require.Equal(t, "0x1", receipt["status"].(string))
	return receipt
}

func TestEth_GetFilterChanges_NoTopics(t *testing.T) {
	rpcRes := call(t, "eth_blockNumber", []string{})

	var res hexutil.Uint64
	err := res.UnmarshalJSON(rpcRes.Result)
	require.NoError(t, err)

	param := make([]map[string]interface{}, 1)
	param[0] = make(map[string]interface{})
	param[0]["topics"] = []string{}
	param[0]["fromBlock"] = res.String()

	// instantiate new filter
	rpcRes = call(t, "eth_newFilter", param)
	require.Nil(t, rpcRes.Error)
	var ID string
	err = json.Unmarshal(rpcRes.Result, &ID)
	require.NoError(t, err)

	// deploy contract, emitting some event
	deployTestContract(t)

	// get filter changes
	changesRes := call(t, "eth_getFilterChanges", []string{ID})

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

func deployTestContractWithFunction(t *testing.T) hexutil.Bytes {
	// pragma solidity ^0.5.1;

	// contract Test {
	//     event Hello(uint256 indexed world);
	//     event TestEvent(uint256 indexed a, uint256 indexed b);

	//     uint256 myStorage;

	//     constructor() public {
	//         emit Hello(17);
	//     }

	//     function test(uint256 a, uint256 b) public {
	//         myStorage = a;
	//         emit TestEvent(a, b);
	//     }
	// }

	bytecode := "0x608060405234801561001057600080fd5b5060117f775a94827b8fd9b519d36cd827093c664f93347070a554f65e4a6f56cd73889860405160405180910390a260d08061004d6000396000f3fe6080604052348015600f57600080fd5b506004361060285760003560e01c8063eb8ac92114602d575b600080fd5b606060048036036040811015604157600080fd5b8101908080359060200190929190803590602001909291905050506062565b005b8160008190555080827ff3ca124a697ba07e8c5e80bebcfcc48991fc16a63170e8a9206e30508960d00360405160405180910390a3505056fea265627a7a723158201d94d2187aaf3a6790527b615fcc40970febf0385fa6d72a2344848ebd0df3e964736f6c63430005110032"
	gasPrice := GetGasPrice(t)

	param := make([]map[string]string, 1)
	param[0] = make(map[string]string)
	param[0]["from"] = "0x" + fmt.Sprintf("%x", from)
	param[0]["data"] = bytecode
	param[0]["gas"] = "0x200000"
	param[0]["gasPrice"] = gasPrice

	rpcRes := call(t, "eth_sendTransaction", param)

	var hash hexutil.Bytes
	err := json.Unmarshal(rpcRes.Result, &hash)
	require.NoError(t, err)

	receipt := waitForReceipt(t, hash)
	require.NotNil(t, receipt, "transaction failed")
	require.Equal(t, "0x1", receipt["status"].(string))

	return hash
}

// Tests topics case where there are topics in first two positions
func TestEth_GetFilterChanges_Topics_AB(t *testing.T) {
	rpcRes := call(t, "eth_blockNumber", []string{})

	var res hexutil.Uint64
	err := res.UnmarshalJSON(rpcRes.Result)
	require.NoError(t, err)

	param := make([]map[string]interface{}, 1)
	param[0] = make(map[string]interface{})
	param[0]["topics"] = []string{helloTopic, worldTopic}
	param[0]["fromBlock"] = res.String()

	// instantiate new filter
	rpcRes = call(t, "eth_newFilter", param)
	var ID string
	err = json.Unmarshal(rpcRes.Result, &ID)
	require.NoError(t, err, string(rpcRes.Result))

	deployTestContractWithFunction(t)

	// get filter changes
	changesRes := call(t, "eth_getFilterChanges", []string{ID})

	var logs []*ethtypes.Log
	err = json.Unmarshal(changesRes.Result, &logs)
	require.NoError(t, err)

	require.Equal(t, 1, len(logs))
}

func TestEth_GetFilterChanges_Topics_XB(t *testing.T) {
	rpcRes := call(t, "eth_blockNumber", []string{})

	var res hexutil.Uint64
	err := res.UnmarshalJSON(rpcRes.Result)
	require.NoError(t, err)

	param := make([]map[string]interface{}, 1)
	param[0] = make(map[string]interface{})
	param[0]["topics"] = []interface{}{nil, worldTopic}
	param[0]["fromBlock"] = res.String()

	// instantiate new filter
	rpcRes = call(t, "eth_newFilter", param)
	var ID string
	err = json.Unmarshal(rpcRes.Result, &ID)
	require.NoError(t, err)

	deployTestContractWithFunction(t)

	// get filter changes
	changesRes := call(t, "eth_getFilterChanges", []string{ID})

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
	rpcRes := call(t, "eth_newPendingTransactionFilter", []string{})

	var ID string
	err := json.Unmarshal(rpcRes.Result, &ID)
	require.NoError(t, err)

	for i := 0; i < 5; i++ {
		deployTestContractWithFunction(t)
	}

	// get filter changes
	changesRes := call(t, "eth_getFilterChanges", []string{ID})
	require.NotNil(t, changesRes)

	var txs []*hexutil.Bytes
	err = json.Unmarshal(changesRes.Result, &txs)
	require.NoError(t, err, string(changesRes.Result))

	require.True(t, len(txs) >= 2, "could not get any txs", "changesRes.Result", string(changesRes.Result))
}

func getNonce(t *testing.T) hexutil.Uint64 {
	param := []interface{}{hexutil.Bytes(from), "latest"}
	rpcRes := call(t, "eth_getTransactionCount", param)

	var nonce hexutil.Uint64
	err := json.Unmarshal(rpcRes.Result, &nonce)
	require.NoError(t, err)
	return nonce
}

func TestEth_EstimateGas(t *testing.T) {
	param := make([]map[string]string, 1)
	param[0] = make(map[string]string)
	param[0]["from"] = "0x" + fmt.Sprintf("%x", from)
	param[0]["to"] = "0x1122334455667788990011223344556677889900"
	param[0]["value"] = "0x1"
	param[0]["gas"] = "0x5209"
	rpcRes := call(t, "eth_estimateGas", param)
	require.NotNil(t, rpcRes)

	var gas string
	err := json.Unmarshal(rpcRes.Result, &gas)
	require.NoError(t, err, string(rpcRes.Result))
	require.Equal(t, "0x5208", gas)
}

func TestEth_EstimateGas_ContractDeployment(t *testing.T) {
	bytecode := "0x608060405234801561001057600080fd5b5060117f775a94827b8fd9b519d36cd827093c664f93347070a554f65e4a6f56cd73889860405160405180910390a260d08061004d6000396000f3fe6080604052348015600f57600080fd5b506004361060285760003560e01c8063eb8ac92114602d575b600080fd5b606060048036036040811015604157600080fd5b8101908080359060200190929190803590602001909291905050506062565b005b8160008190555080827ff3ca124a697ba07e8c5e80bebcfcc48991fc16a63170e8a9206e30508960d00360405160405180910390a3505056fea265627a7a723158201d94d2187aaf3a6790527b615fcc40970febf0385fa6d72a2344848ebd0df3e964736f6c63430005110032"

	param := make([]map[string]string, 1)
	param[0] = make(map[string]string)
	param[0]["from"] = "0x" + fmt.Sprintf("%x", from)
	param[0]["data"] = bytecode

	rpcRes := call(t, "eth_estimateGas", param)
	require.NotNil(t, rpcRes)
	require.NotEmpty(t, rpcRes.Result)

	var gas hexutil.Uint64
	err := json.Unmarshal(rpcRes.Result, &gas)
	require.NoError(t, err, string(rpcRes.Result))

	require.Equal(t, "0x1879c", gas.String())
}

func TestEth_ExportAccount_WithStorage(t *testing.T) {
	t.Skip("skipping TestEth_ExportAccount_WithStorage due to the server haven't implmented yet")

	hash := deployTestContractWithFunction(t)
	receipt := waitForReceipt(t, hash)
	addr := receipt["contractAddress"].(string)

	// call function to set storage
	calldata := "0xeb8ac92100000000000000000000000000000000000000000000000000000000000000630000000000000000000000000000000000000000000000000000000000000000"

	param := make([]map[string]string, 1)
	param[0] = make(map[string]string)
	param[0]["from"] = "0x" + fmt.Sprintf("%x", from)
	param[0]["to"] = addr
	param[0]["data"] = calldata
	param[0]["gas"] = "0x200000"
	param[0]["gasPrice"] = "0x1"

	rpcRes := call(t, "eth_sendTransaction", param)

	var txhash hexutil.Bytes
	err := json.Unmarshal(rpcRes.Result, &txhash)
	require.NoError(t, err)
	waitForReceipt(t, txhash)

	// get exported account
	eap := []string{}
	eap = append(eap, addr)
	eap = append(eap, "latest")
	rpcRes = call(t, "eth_exportAccount", eap)

	var res string
	err = json.Unmarshal(rpcRes.Result, &res)
	require.NoError(t, err)

	var account evmtypes.GenesisAccount
	err = json.Unmarshal([]byte(res), &account)
	require.NoError(t, err)

	// deployed bytecode
	bytecode := "0x6080604052348015600f57600080fd5b506004361060285760003560e01c8063eb8ac92114602d575b600080fd5b606060048036036040811015604157600080fd5b8101908080359060200190929190803590602001909291905050506062565b005b8160008190555080827ff3ca124a697ba07e8c5e80bebcfcc48991fc16a63170e8a9206e30508960d00360405160405180910390a3505056fea265627a7a723158201d94d2187aaf3a6790527b615fcc40970febf0385fa6d72a2344848ebd0df3e964736f6c63430005110032"
	require.Equal(t, addr, account.Address)
	require.Equal(t, bytecode, account.Code)
	require.NotEqual(t, evmtypes.Storage(nil), account.Storage)
}

func TestEth_GetBlockByHash(t *testing.T) {
	param := []interface{}{"0x1", false}
	rpcRes := call(t, "eth_getBlockByNumber", param)

	block := make(map[string]interface{})
	err := json.Unmarshal(rpcRes.Result, &block)
	require.NoError(t, err)
	blockHash := block["hash"].(string)

	param = []interface{}{blockHash, false}
	rpcRes = call(t, "eth_getBlockByHash", param)
	block = make(map[string]interface{})
	err = json.Unmarshal(rpcRes.Result, &block)
	require.NoError(t, err)
	require.Equal(t, "0x1", block["number"].(string))
}

func TestEth_GetBlockByHash_BlockHashNotFound(t *testing.T) {
	anyBlockHash := "0xb3b20624f8f0f86eb50dd04688409e5cea4bd02d700bf6e79e9384d47d6a5a35"
	param := []interface{}{anyBlockHash, false}
	rpcRes := call(t, "eth_getBlockByHash", param)

	var result interface{}
	err := json.Unmarshal(rpcRes.Result, &result)
	require.NoError(t, err)
	require.Nil(t, result)
}

func TestEth_GetBlockByNumber(t *testing.T) {
	param := []interface{}{"0x1", false}
	rpcRes := call(t, "eth_getBlockByNumber", param)

	block := make(map[string]interface{})
	err := json.Unmarshal(rpcRes.Result, &block)
	require.NoError(t, err)
	require.Equal(t, "0x", block["extraData"].(string))
	require.Equal(t, []interface{}{}, block["uncles"].([]interface{}))
}

func TestEth_GetLogs(t *testing.T) {
	rpcRes := call(t, "eth_blockNumber", []string{})

	var res hexutil.Uint64
	err := res.UnmarshalJSON(rpcRes.Result)
	require.NoError(t, err)

	param := make([]map[string]interface{}, 1)
	param[0] = make(map[string]interface{})
	param[0]["topics"] = []string{helloTopic, worldTopic}
	param[0]["fromBlock"] = res.String()

	deployTestContractWithFunction(t)

	// get filter changes
	logRes := call(t, "eth_getLogs", param)

	var logs []*ethtypes.Log
	err = json.Unmarshal(logRes.Result, &logs)
	require.NoError(t, err)

	require.Equal(t, 1, len(logs))

	// filter log with address
	param[0] = make(map[string]interface{})
	param[0]["address"] = "0x" + fmt.Sprintf("%x", from)
	param[0]["fromBlock"] = res.String()
	err = json.Unmarshal(logRes.Result, &logs)
	require.NoError(t, err)

	require.Equal(t, 1, len(logs))
}

func makeEthTxParam() []map[string]string {
	param := make([]map[string]string, 1)
	param[0] = make(map[string]string)
	param[0]["from"] = "0x" + fmt.Sprintf("%x", from)
	param[0]["to"] = "0x0000000000000000000000000000000012341234"
	param[0]["value"] = "0x16345785d8a0000"
	param[0]["gasLimit"] = "0x5208"
	param[0]["gasPrice"] = "0x55ae82600"

	return param
}

func makeTestContractDeployParam(t *testing.T, withGas bool) []map[string]string {
	param := make([]map[string]string, 1)
	param[0] = make(map[string]string)
	param[0]["from"] = "0x" + fmt.Sprintf("%x", from)
	param[0]["data"] = "0x6080604052348015600f57600080fd5b5060117f775a94827b8fd9b519d36cd827093c664f93347070a554f65e4a6f56cd73889860405160405180910390a2603580604b6000396000f3fe6080604052600080fdfea165627a7a723058206cab665f0f557620554bb45adf266708d2bd349b8a4314bdff205ee8440e3c240029"
	if withGas {
		gasPrice := GetGasPrice(t)
		param[0]["gas"] = "0x200000"
		param[0]["gasPrice"] = gasPrice
	}

	return param
}

func TestEth_EthResend(t *testing.T) {
	tx := make(map[string]string)
	tx["from"] = "0x" + fmt.Sprintf("%x", from)
	tx["to"] = "0x0000000000000000000000000000000012341234"
	tx["value"] = "0x16345785d8a0000"
	tx["nonce"] = "0x2"
	tx["gasLimit"] = "0x5208"
	tx["gasPrice"] = "0x55ae82600"
	param := []interface{}{tx, "0x1", "0x2"}
	_, rpcerror := callWithError("eth_resend", param)
	require.Equal(t, "transaction 0x3bf28b46ee1bb3925e50ec6003f899f95913db4b0f579c4e7e887efebf9ecd1b not found", fmt.Sprintf("%s", rpcerror))
}

func TestEth_FeeHistory(t *testing.T) {
	params := make([]interface{}, 0)
	params = append(params, 4)
	params = append(params, "0x1c")
	params = append(params, []int{25, 75})

	rpcRes := call(t, "eth_feeHistory", params)

	info := make(map[string]interface{})
	err := json.Unmarshal(rpcRes.Result, &info)
	require.NoError(t, err)
	reward := info["reward"].([]interface{})
	baseFeePerGas := info["baseFeePerGas"].([]interface{})
	gasUsedRatio := info["gasUsedRatio"].([]interface{})

	require.Equal(t, info["oldestBlock"].(string), "0x18")
	require.Equal(t, 4, len(gasUsedRatio))
	require.Equal(t, 4, len(baseFeePerGas))
	require.Equal(t, 4, len(reward))
}
