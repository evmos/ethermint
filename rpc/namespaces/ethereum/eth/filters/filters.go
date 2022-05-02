package filters

import (
	"context"
	"encoding/binary"
	"math/big"

	"github.com/tharsis/ethermint/rpc/types"

	"github.com/pkg/errors"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/eth/filters"
)

// BloomIV represents the bit indexes and value inside the bloom filter that belong
// to some key.
type BloomIV struct {
	I [3]uint
	V [3]byte
}

// Filter can be used to retrieve and filter logs.
type Filter struct {
	logger   log.Logger
	backend  Backend
	criteria filters.FilterCriteria

	bloomFilters [][]BloomIV // Filter the system is matching for
}

// NewBlockFilter creates a new filter which directly inspects the contents of
// a block to figure out whether it is interesting or not.
func NewBlockFilter(logger log.Logger, backend Backend, criteria filters.FilterCriteria) *Filter {
	// Create a generic filter and convert it into a block filter
	return newFilter(logger, backend, criteria, nil)
}

// NewRangeFilter creates a new filter which uses a bloom filter on blocks to
// figure out whether a particular block is interesting or not.
func NewRangeFilter(logger log.Logger, backend Backend, begin, end int64, addresses []common.Address, topics [][]common.Hash) *Filter {
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

	// Create a generic filter and convert it into a range filter
	criteria := filters.FilterCriteria{
		FromBlock: big.NewInt(begin),
		ToBlock:   big.NewInt(end),
		Addresses: addresses,
		Topics:    topics,
	}

	return newFilter(logger, backend, criteria, createBloomFilters(filtersBz, logger))
}

// newFilter returns a new Filter
func newFilter(logger log.Logger, backend Backend, criteria filters.FilterCriteria, bloomFilters [][]BloomIV) *Filter {
	return &Filter{
		logger:       logger,
		backend:      backend,
		criteria:     criteria,
		bloomFilters: bloomFilters,
	}
}

const (
	maxToOverhang = 600
)

// Logs searches the blockchain for matching log entries, returning all from the
// first block that contains matches, updating the start of the filter accordingly.
func (f *Filter) Logs(_ context.Context, logLimit int, blockLimit int64) ([]*ethtypes.Log, error) {
	logs := []*ethtypes.Log{}
	var err error

	// If we're doing singleton block filtering, execute and return
	if f.criteria.BlockHash != nil && *f.criteria.BlockHash != (common.Hash{}) {
		header, err := f.backend.HeaderByHash(*f.criteria.BlockHash)
		if err != nil {
			return nil, errors.Wrap(err, "failed to fetch header by hash")
		}

		if header == nil {
			return nil, errors.Errorf("unknown block header %s", f.criteria.BlockHash.String())
		}

		return f.blockLogs(header.Number.Int64(), header.Bloom)
	}

	// Figure out the limits of the filter range
	header, err := f.backend.HeaderByNumber(types.EthLatestBlockNumber)
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch header by number (latest)")
	}

	if header == nil || header.Number == nil {
		f.logger.Debug("header not found or has no number")
		return nil, nil
	}

	head := header.Number.Int64()
	if f.criteria.FromBlock.Int64() < 0 {
		f.criteria.FromBlock = big.NewInt(head)
	} else if f.criteria.FromBlock.Int64() == 0 {
		f.criteria.FromBlock = big.NewInt(1)
	}
	if f.criteria.ToBlock.Int64() < 0 {
		f.criteria.ToBlock = big.NewInt(head)
	} else if f.criteria.ToBlock.Int64() == 0 {
		f.criteria.ToBlock = big.NewInt(1)
	}

	if f.criteria.ToBlock.Int64()-f.criteria.FromBlock.Int64() > blockLimit {
		return nil, errors.Errorf("maximum [from, to] blocks distance: %d", blockLimit)
	}

	// check bounds
	if f.criteria.FromBlock.Int64() > head {
		return []*ethtypes.Log{}, nil
	} else if f.criteria.ToBlock.Int64() > head+maxToOverhang {
		f.criteria.ToBlock = big.NewInt(head + maxToOverhang)
	}

	from := f.criteria.FromBlock.Int64()
	to := f.criteria.ToBlock.Int64()

	for height := from; height <= to; height++ {
		bloom, err := f.backend.BlockBloom(&height)
		if err != nil {
			return nil, err
		}

		filtered, err := f.blockLogs(height, bloom)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to fetch block by number %d", height)
		}

		// check logs limit
		if len(logs)+len(filtered) > logLimit {
			return nil, errors.Errorf("query returned more than %d results", logLimit)
		}
		logs = append(logs, filtered...)
	}
	return logs, nil
}

// blockLogs returns the logs matching the filter criteria within a single block.
func (f *Filter) blockLogs(height int64, bloom ethtypes.Bloom) ([]*ethtypes.Log, error) {
	if !bloomFilter(bloom, f.criteria.Addresses, f.criteria.Topics) {
		return []*ethtypes.Log{}, nil
	}

	// DANGER: do not call GetLogs(header.Hash())
	// eth header's hash doesn't match tm block hash
	logsList, err := f.backend.GetLogsByNumber(types.BlockNumber(height))
	if err != nil {
		return []*ethtypes.Log{}, errors.Wrapf(err, "failed to fetch logs block number %d", height)
	}

	unfiltered := make([]*ethtypes.Log, 0)
	for _, logs := range logsList {
		unfiltered = append(unfiltered, logs...)
	}

	logs := FilterLogs(unfiltered, nil, nil, f.criteria.Addresses, f.criteria.Topics)
	if len(logs) == 0 {
		return []*ethtypes.Log{}, nil
	}

	return logs, nil
}

func createBloomFilters(filters [][][]byte, logger log.Logger) [][]BloomIV {
	bloomFilters := make([][]BloomIV, 0)
	for _, filter := range filters {
		// Gather the bit indexes of the filter rule, special casing the nil filter
		if len(filter) == 0 {
			continue
		}
		bloomIVs := make([]BloomIV, len(filter))

		// Transform the filter rules (the addresses and topics) to the bloom index and value arrays
		// So it can be used to compare with the bloom of the block header. If the rule has any nil
		// clauses. The rule will be ignored.
		for i, clause := range filter {
			if clause == nil {
				bloomIVs = nil
				break
			}

			iv, err := calcBloomIVs(clause)
			if err != nil {
				bloomIVs = nil
				logger.Error("calcBloomIVs error: %v", err)
				break
			}

			bloomIVs[i] = iv
		}
		// Accumulate the filter rules if no nil rule was within
		if bloomIVs != nil {
			bloomFilters = append(bloomFilters, bloomIVs)
		}
	}
	return bloomFilters
}

// calcBloomIVs returns BloomIV for the given data,
// revised from https://github.com/ethereum/go-ethereum/blob/401354976bb44f0ad4455ca1e0b5c0dc31d9a5f5/core/types/bloom9.go#L139
func calcBloomIVs(data []byte) (BloomIV, error) {
	hashbuf := make([]byte, 6)
	biv := BloomIV{}

	sha := crypto.NewKeccakState()
	sha.Reset()
	if _, err := sha.Write(data); err != nil {
		return BloomIV{}, err
	}
	if _, err := sha.Read(hashbuf); err != nil {
		return BloomIV{}, err
	}

	// The actual bits to flip
	biv.V[0] = byte(1 << (hashbuf[1] & 0x7))
	biv.V[1] = byte(1 << (hashbuf[3] & 0x7))
	biv.V[2] = byte(1 << (hashbuf[5] & 0x7))
	// The indices for the bytes to OR in
	biv.I[0] = ethtypes.BloomByteLength - uint((binary.BigEndian.Uint16(hashbuf)&0x7ff)>>3) - 1
	biv.I[1] = ethtypes.BloomByteLength - uint((binary.BigEndian.Uint16(hashbuf[2:])&0x7ff)>>3) - 1
	biv.I[2] = ethtypes.BloomByteLength - uint((binary.BigEndian.Uint16(hashbuf[4:])&0x7ff)>>3) - 1

	return biv, nil
}
