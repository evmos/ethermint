package rpc

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
)

/*
	- Filter functions derived from go-ethereum
	Used to set the criteria passed in from RPC params
*/

// Filter can be used to retrieve and filter logs.
type Filter struct {
	addresses []common.Address
	topics    [][]common.Hash

	block      common.Hash // Block hash if filtering a single block
}

// NewBlockFilter creates a new filter which directly inspects the contents of
// a block to figure out whether it is interesting or not.
func NewBlockFilter(block common.Hash, addresses []common.Address, topics [][]common.Hash) *Filter {
	// Create a generic filter and convert it into a block filter
	filter := newFilter(addresses, topics)
	filter.block = block
	return filter
}

// newFilter creates a generic filter that can either filter based on a block hash,
// or based on range queries. The search criteria needs to be explicitly set.
func newFilter(addresses []common.Address, topics [][]common.Hash) *Filter {
	return &Filter{
		addresses: addresses,
		topics:    topics,
	}
}

func includes(addresses []common.Address, a common.Address) bool {
	for _, addr := range addresses {
		if addr == a {
			return true
		}
	}

	return false
}

// filterLogs creates a slice of logs matching the given criteria.
func filterLogs(logs []*ethtypes.Log, fromBlock, toBlock *big.Int, addresses []common.Address, topics [][]common.Hash) []*ethtypes.Log {
	var ret []*ethtypes.Log
Logs:
	for _, log := range logs {
		if fromBlock != nil && fromBlock.Int64() >= 0 && fromBlock.Uint64() > log.BlockNumber {
			continue
		}
		if toBlock != nil && toBlock.Int64() >= 0 && toBlock.Uint64() < log.BlockNumber {
			continue
		}
		if len(addresses) > 0 && !includes(addresses, log.Address) {
			continue
		}
		// If the to filtered topics is greater than the amount of topics in logs, skip.
		if len(topics) > len(log.Topics) {
			continue Logs
		}
		for i, sub := range topics {
			match := len(sub) == 0 // empty rule set == wildcard
			for _, topic := range sub {
				if log.Topics[i] == topic {
					match = true
					break
				}
			}
			if !match {
				continue Logs
			}
		}
		ret = append(ret, log)
	}
	return ret
}
