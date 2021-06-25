package filters

import (
	"context"
	"fmt"
	"math/big"

	"github.com/tharsis/ethermint/ethereum/rpc/types"

	"github.com/pkg/errors"
	log "github.com/xlab/suplog"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/bloombits"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/eth/filters"
)

// Filter can be used to retrieve and filter logs.
type Filter struct {
	backend  Backend
	criteria filters.FilterCriteria
	matcher  *bloombits.Matcher
}

// NewBlockFilter creates a new filter which directly inspects the contents of
// a block to figure out whether it is interesting or not.
func NewBlockFilter(backend Backend, criteria filters.FilterCriteria) *Filter {
	// Create a generic filter and convert it into a block filter
	return newFilter(backend, criteria, nil)
}

// NewRangeFilter creates a new filter which uses a bloom filter on blocks to
// figure out whether a particular block is interesting or not.
func NewRangeFilter(backend Backend, begin, end int64, addresses []common.Address, topics [][]common.Hash) *Filter {
	// Flatten the address and topic filter clauses into a single bloombits filter
	// system. Since the bloombits are not positional, nil topics are permitted,
	// which get flattened into a nil byte slice.
	var filtersBz [][][]byte // nolint: prealloc
	if len(addresses) > 0 {
		filter := make([][]byte, len(addresses))
		for i, address := range addresses {
			filter[i] = address.Bytes()
		}
		filtersBz = append(filtersBz, filter)
	}

	for _, topicList := range topics {
		filter := make([][]byte, len(topicList))
		for i, topic := range topicList {
			filter[i] = topic.Bytes()
		}
		filtersBz = append(filtersBz, filter)
	}

	size, _ := backend.BloomStatus()

	// Create a generic filter and convert it into a range filter
	criteria := filters.FilterCriteria{
		FromBlock: big.NewInt(begin),
		ToBlock:   big.NewInt(end),
		Addresses: addresses,
		Topics:    topics,
	}

	return newFilter(backend, criteria, bloombits.NewMatcher(size, filtersBz))
}

// newFilter returns a new Filter
func newFilter(backend Backend, criteria filters.FilterCriteria, matcher *bloombits.Matcher) *Filter {
	return &Filter{
		backend:  backend,
		criteria: criteria,
		matcher:  matcher,
	}
}

const (
	maxFilterBlocks = 100000
	maxToOverhang   = 600
)

// Logs searches the blockchain for matching log entries, returning all from the
// first block that contains matches, updating the start of the filter accordingly.
func (f *Filter) Logs(_ context.Context) ([]*ethtypes.Log, error) {
	logs := []*ethtypes.Log{}
	var err error

	// If we're doing singleton block filtering, execute and return
	if f.criteria.BlockHash != nil && *f.criteria.BlockHash != (common.Hash{}) {
		header, err := f.backend.HeaderByHash(*f.criteria.BlockHash)
		if err != nil {
			err = errors.Wrap(err, "failed to fetch header by hash")
			return nil, err
		}

		if header == nil {
			err := errors.Errorf("unknown block header %s", f.criteria.BlockHash.String())
			return nil, err
		}

		return f.blockLogs(header)
	}

	// Figure out the limits of the filter range
	header, err := f.backend.HeaderByNumber(types.EthLatestBlockNumber)
	if err != nil {
		err = errors.Wrap(err, "failed to fetch header by number (latest)")
		return nil, err
	}

	if header == nil || header.Number == nil {
		log.Warningln("header not found or has no number")
		return nil, nil
	}

	head := header.Number.Int64()
	if f.criteria.FromBlock.Int64() == -1 {
		f.criteria.FromBlock = big.NewInt(head)
	}
	if f.criteria.ToBlock.Int64() == -1 {
		f.criteria.ToBlock = big.NewInt(head)
	}

	if f.criteria.ToBlock.Int64()-f.criteria.FromBlock.Int64() > maxFilterBlocks {
		err := errors.Errorf("maximum [from, to] blocks distance: %d", maxFilterBlocks)
		return nil, err
	}

	// check bounds
	if f.criteria.FromBlock.Int64() > head {
		return []*ethtypes.Log{}, nil
	} else if f.criteria.ToBlock.Int64() > head+maxToOverhang {
		f.criteria.ToBlock = big.NewInt(head + maxToOverhang)
	}

	for i := f.criteria.FromBlock.Int64(); i <= f.criteria.ToBlock.Int64(); i++ {
		block, err := f.backend.GetBlockByNumber(types.BlockNumber(i), false)
		if err != nil {
			err = errors.Wrapf(err, "failed to fetch block by number %d", i)
			return logs, err
		} else if block["transactions"] == nil {
			continue
		}

		var txHashes []common.Hash

		txs, ok := block["transactions"].([]interface{})
		if !ok {
			txHashes, ok = block["transactions"].([]common.Hash)
			if !ok {
				log.WithField(
					"transactions",
					fmt.Sprintf("%T", block["transactions"]),
				).Errorln("reading transactions from block data: bad field type")
				continue
			}
		} else if len(txs) == 0 && len(txHashes) == 0 {
			continue
		}

		for _, tx := range txs {
			txHash, ok := tx.(common.Hash)
			if !ok {
				log.WithField(
					"tx",
					fmt.Sprintf("%T", tx),
				).Errorln("transactions list contains non-hash element")
			} else {
				txHashes = append(txHashes, txHash)
			}
		}

		logsMatched := f.checkMatches(txHashes)
		logs = append(logs, logsMatched...)
	}

	return logs, nil
}

// blockLogs returns the logs matching the filter criteria within a single block.
func (f *Filter) blockLogs(header *ethtypes.Header) ([]*ethtypes.Log, error) {
	if !bloomFilter(header.Bloom, f.criteria.Addresses, f.criteria.Topics) {
		return []*ethtypes.Log{}, nil
	}

	// DANGER: do not call GetLogs(header.Hash())
	// eth header's hash doesn't match tm block hash
	logsList, err := f.backend.GetLogsByNumber(types.BlockNumber(header.Number.Int64()))
	if err != nil {
		err = errors.Wrapf(err, "failed to fetch logs block number %d", header.Number.Int64())

		return []*ethtypes.Log{}, err
	}

	var unfiltered []*ethtypes.Log // nolint: prealloc
	for _, logs := range logsList {
		unfiltered = append(unfiltered, logs...)
	}

	logs := FilterLogs(unfiltered, nil, nil, f.criteria.Addresses, f.criteria.Topics)
	if len(logs) == 0 {
		return []*ethtypes.Log{}, nil
	}

	return logs, nil
}

// checkMatches checks if the logs from the a list of transactions transaction
// contain any log events that  match the filter criteria. This function is
// called when the bloom filter signals a potential match.
func (f *Filter) checkMatches(transactions []common.Hash) []*ethtypes.Log {
	unfiltered := []*ethtypes.Log{}
	for _, tx := range transactions {
		logs, err := f.backend.GetTransactionLogs(tx)
		if err != nil {
			// ignore error if transaction didn't set any logs (eg: when tx type is not
			// MsgEthereumTx or MsgEthermint)
			continue
		}

		unfiltered = append(unfiltered, logs...)
	}

	return FilterLogs(unfiltered, f.criteria.FromBlock, f.criteria.ToBlock, f.criteria.Addresses, f.criteria.Topics)
}
