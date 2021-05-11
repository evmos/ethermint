package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/ethereum/go-ethereum/common/hexutil"
)

var (
	HOST = os.Getenv("HOST")
	HOME = os.Getenv("PWD")
)

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

func getTransactionReceipt(hash hexutil.Bytes) (map[string]interface{}, error) {
	param := []string{hash.String()}

	rpcRes, err := call("eth_getTransactionReceipt", param)
	if err != nil {
		return nil, err
	}

	receipt := make(map[string]interface{})

	err = json.Unmarshal(rpcRes.Result, &receipt)
	if err != nil {
		return nil, err
	}

	return receipt, nil
}

func waitForReceipt(hash hexutil.Bytes) (map[string]interface{}, error) {
	for i := 0; i < 10; i++ {
		receipt, err := getTransactionReceipt(hash)
		if receipt != nil {
			return receipt, err
		} else if err != nil {
			return nil, err
		}

		time.Sleep(time.Second)
	}

	return nil, errors.New("cound not find transaction on chain")
}

func call(method string, params interface{}) (*Response, error) {
	var rpcRes *Response

	if HOST == "" {
		HOST = "http://localhost:8545"
	}

	req, err := json.Marshal(createRequest(method, params))
	if err != nil {
		return nil, err
	}

	time.Sleep(1000000 * time.Nanosecond)
	/* #nosec */
	res, err := http.Post(HOST, "application/json", bytes.NewBuffer(req))
	if err != nil {
		return nil, err
	}

	decoder := json.NewDecoder(res.Body)
	rpcRes = new(Response)

	if err := decoder.Decode(&rpcRes); err != nil {
		return nil, err
	}

	if err := res.Body.Close(); err != nil {
		return nil, err
	}

	return rpcRes, nil
}

func main() {
	var hash hexutil.Bytes

	dat, err := ioutil.ReadFile(HOME + "/counter/counter_sol.bin")
	if err != nil {
		log.Fatal(err)
	}

	param := make([]map[string]string, 1)
	param[0] = make(map[string]string)
	param[0]["from"] = os.Args[1]
	param[0]["data"] = "0x" + string(dat)

	txRPCRes, err := call("eth_sendTransaction", param)
	if err != nil {
		log.Fatal(err)
	}

	if err := json.Unmarshal(txRPCRes.Result, &hash); err != nil {
		log.Fatal(err)
	}

	fmt.Println("Contract TX hash: ", hash)

	receipt, err := waitForReceipt(hash)
	if err != nil {
		log.Fatal(err)
	}

	/*
		//test for bad hash
		testhash, err := hexutil.Decode("0xe146d95c74a48e730bf825c2a3dcbce8122b8a463bc15bcbb38b9c195402f0a5")
		if err != nil {
			log.Fatal(err)
		}
		receipt, err := waitForReceipt(testhash)
		if err != nil {
			log.Fatal(err)
		}
	*/

	fmt.Println("receipt: ", receipt)
}
