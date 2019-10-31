package rpc

import (
	"encoding/json"
	"fmt"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/ethermint/x/evm/types"

	"math/big"

	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/eth/filters"
	"github.com/ethereum/go-ethereum/rpc"
)

// PublicFilterAPI is the eth_ prefixed set of APIs in the Web3 JSON-RPC spec.
type PublicFilterAPI struct {
	cliCtx context.CLIContext
}

// NewPublicEthAPI creates an instance of the public ETH Web3 API.
func NewPublicFilterAPI(cliCtx context.CLIContext) *PublicFilterAPI {
	return &PublicFilterAPI{
		cliCtx: cliCtx,
	}
}

// GetLogs returns logs matching the given argument that are stored within the state.
//
// https://github.com/ethereum/wiki/wiki/JSON-RPC#eth_getlogs
func (e *PublicFilterAPI) GetLogs(criteria filters.FilterCriteria) ([]*ethtypes.Log, error) {
	var filter *Filter
	if criteria.BlockHash != nil {
		/*
			Still need to add blockhash in prepare function for log entry
		*/
		filter = NewBlockFilter(*criteria.BlockHash, criteria.Addresses, criteria.Topics)
		results := e.getLogs()
		logs := filterLogs(results, nil, nil, filter.addresses, filter.topics)
		return logs, nil
	} else {
		// Convert the RPC block numbers into internal representations
		begin := rpc.LatestBlockNumber.Int64()
		if criteria.FromBlock != nil {
			begin = criteria.FromBlock.Int64()
		}
		from := big.NewInt(begin)
		end := rpc.LatestBlockNumber.Int64()
		if criteria.ToBlock != nil {
			end = criteria.ToBlock.Int64()
		}
		to := big.NewInt(end)
		results := e.getLogs()
		logs := filterLogs(results, from, to, criteria.Addresses, criteria.Topics)

		return returnLogs(logs), nil
	}
}

func (e *PublicFilterAPI) getLogs() (results []*ethtypes.Log) {
	l, _, err := e.cliCtx.QueryWithData(fmt.Sprintf("custom/%s/logs", types.ModuleName), nil)
	if err != nil {
		fmt.Printf("error from querier %e ", err)
	}

	if err := json.Unmarshal(l, &results); err != nil {
		panic(err)
	}
	return results
}

// returnLogs is a helper that will return an empty log array in case the given logs array is nil,
// otherwise the given logs array is returned.
func returnLogs(logs []*ethtypes.Log) []*ethtypes.Log {
	if logs == nil {
		return []*ethtypes.Log{}
	}
	return logs
}
