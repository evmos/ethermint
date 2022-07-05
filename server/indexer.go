package server

import (
	"errors"
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authante "github.com/cosmos/cosmos-sdk/x/auth/ante"
	"github.com/ethereum/go-ethereum/common"
	"github.com/evmos/ethermint/rpc/backend"
	rpctypes "github.com/evmos/ethermint/rpc/types"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
	tmtypes "github.com/tendermint/tendermint/types"
	dbm "github.com/tendermint/tm-db"

	ethermint "github.com/evmos/ethermint/types"
	evmtypes "github.com/evmos/ethermint/x/evm/types"
)

var (
	_ ethermint.EVMTxIndexer = &KVIndexer{}

	TxIndexKeyLength = len(TxIndexKey(0, 0))
)

const (
	KeyPrefixTxHash  = 1
	KeyPrefixTxIndex = 2
)

func TxHashKey(hash common.Hash) []byte {
	return append([]byte{KeyPrefixTxHash}, hash.Bytes()...)
}

func TxIndexKey(blockNumber int64, txIndex int32) []byte {
	bz1 := sdk.Uint64ToBigEndian(uint64(blockNumber))
	bz2 := sdk.Uint64ToBigEndian(uint64(txIndex))
	return append(append([]byte{KeyPrefixTxIndex}, bz1...), bz2...)
}

func parseBlockNumberFromKey(key []byte) (int64, error) {
	if len(key) != TxIndexKeyLength {
		return 0, fmt.Errorf("wrong tx index key length, expect: %d, got: %d", TxIndexKeyLength, len(key))
	}

	return int64(sdk.BigEndianToUint64(key[1:9])), nil
}

// LoadLastBlock returns -1 if db is empty
func LoadLastBlock(db dbm.DB) (int64, error) {
	it, err := db.ReverseIterator([]byte{KeyPrefixTxIndex}, []byte{KeyPrefixTxIndex + 1})
	if err != nil {
		return 0, err
	}
	defer it.Close()
	if !it.Valid() {
		return -1, nil
	}
	return parseBlockNumberFromKey(it.Key())
}

// LoadFirstBlock returns -1 if db is empty
func LoadFirstBlock(db dbm.DB) (int64, error) {
	it, err := db.Iterator([]byte{KeyPrefixTxIndex}, []byte{KeyPrefixTxIndex + 1})
	if err != nil {
		return 0, err
	}
	defer it.Close()
	if !it.Valid() {
		return -1, nil
	}
	return parseBlockNumberFromKey(it.Key())
}

type KVIndexer struct {
	db        dbm.DB
	logger    log.Logger
	clientCtx client.Context
}

func NewKVIndexer(db dbm.DB, logger log.Logger, clientCtx client.Context) *KVIndexer {
	return &KVIndexer{db, logger, clientCtx}
}

func (kv *KVIndexer) LastIndexedBlock() (int64, error) {
	return LoadLastBlock(kv.db)
}

func (kv *KVIndexer) FirstIndexedBlock() (int64, error) {
	return LoadFirstBlock(kv.db)
}

func (kv *KVIndexer) IndexBlock(blk *tmtypes.Block, txResults []*abci.ResponseDeliverTx) error {
	height := blk.Header.Height

	batch := kv.db.NewBatch()
	defer batch.Close()

	var ethTxIndex int32
	for txIndex, tx := range blk.Txs {
		result := txResults[txIndex]
		if !backend.TxSuccessOrExceedsBlockGasLimit(result) {
			continue
		}
		tx, err := kv.clientCtx.TxConfig.TxDecoder()(tx)
		if err != nil {
			kv.logger.Error("Fail to decode tx", "err", err)
			continue
		}
		extTx, ok := tx.(authante.HasExtensionOptionsTx)
		if !ok {
			// not eth tx
			continue
		}
		opts := extTx.GetExtensionOptions()
		if len(opts) != 1 || opts[0].GetTypeUrl() != "/ethermint.evm.v1.ExtensionOptionsEthereumTx" {
			// not eth tx
			continue
		}

		txs, err := rpctypes.ParseTxResult(result, tx)
		if err != nil {
			kv.logger.Error("Fail to parse event", "err", err)
			continue
		}

		var cumulativeGasUsed uint64
		for msgIndex, msg := range tx.GetMsgs() {
			ethMsg := msg.(*evmtypes.MsgEthereumTx)
			txHash := common.HexToHash(ethMsg.Hash)

			var txResult *ethermint.TxResult
			if result.Code != abci.CodeTypeOK {
				// exceeds block gas limit scenario, some old versions don't emit any events, workaround directly.
				cumulativeGasUsed += ethMsg.GetGas()
				txResult = &ethermint.TxResult{
					Height:            height,
					TxIndex:           uint32(txIndex),
					MsgIndex:          uint32(msgIndex),
					EthTxIndex:        ethTxIndex,
					GasUsed:           ethMsg.GetGas(),
					CumulativeGasUsed: cumulativeGasUsed,
					Failed:            true,
				}
			} else {
				parsedTx := txs.GetTxByMsgIndex(msgIndex)
				if parsedTx == nil {
					kv.logger.Error("msg index not found in events: %d", msgIndex)
					continue
				}
				if parsedTx.EthTxIndex >= 0 && parsedTx.EthTxIndex != ethTxIndex {
					kv.logger.Error("eth tx index don't match %d != %d\n", parsedTx.EthTxIndex, ethTxIndex)
				}
				cumulativeGasUsed += parsedTx.GasUsed
				txResult = &ethermint.TxResult{
					Height:            height,
					TxIndex:           uint32(txIndex),
					MsgIndex:          uint32(msgIndex),
					EthTxIndex:        ethTxIndex,
					GasUsed:           parsedTx.GasUsed,
					CumulativeGasUsed: cumulativeGasUsed,
					Failed:            parsedTx.Failed,
				}
			}
			bz := kv.clientCtx.Codec.MustMarshal(txResult)
			if err := batch.Set(TxHashKey(txHash), bz); err != nil {
				return err
			}
			if err := batch.Set(TxIndexKey(txResult.Height, txResult.EthTxIndex), txHash.Bytes()); err != nil {
				return err
			}
			ethTxIndex++
		}
	}
	return batch.Write()
}

func (kv *KVIndexer) GetByTxHash(hash common.Hash) (*ethermint.TxResult, error) {
	bz, err := kv.db.Get(TxHashKey(hash))
	if err != nil {
		return nil, err
	}
	if len(bz) == 0 {
		return nil, errors.New("tx not found")
	}
	var txKey ethermint.TxResult
	if err := kv.clientCtx.Codec.Unmarshal(bz, &txKey); err != nil {
		return nil, err
	}
	return &txKey, nil
}

func (kv *KVIndexer) GetByBlockAndIndex(blockNumber int64, txIndex int32) (*ethermint.TxResult, error) {
	bz, err := kv.db.Get(TxIndexKey(blockNumber, txIndex))
	if err != nil {
		return nil, err
	}
	if len(bz) == 0 {
		return nil, errors.New("tx not found")
	}
	return kv.GetByTxHash(common.BytesToHash(bz))
}
