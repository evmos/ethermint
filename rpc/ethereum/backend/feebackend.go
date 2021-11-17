package backend

import (
	"fmt"
	"math/big"
	"sort"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/rpc"
	tmrpctypes "github.com/tendermint/tendermint/rpc/core/types"
	rpctypes "github.com/tharsis/ethermint/rpc/ethereum/types"
	evmtypes "github.com/tharsis/ethermint/x/evm/types"
)

type (
	txGasAndReward struct {
		gasUsed uint64
		reward  *big.Int
	}
	sortGasAndReward []txGasAndReward
)

func (s sortGasAndReward) Len() int { return len(s) }
func (s sortGasAndReward) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s sortGasAndReward) Less(i, j int) bool {
	return s[i].reward.Cmp(s[j].reward) < 0
}

// output: targetOneFeeHistory
func (e *EVMBackend) processBlock(
	tendermintBlock *tmrpctypes.ResultBlock,
	ethBlock *map[string]interface{},
	rewardPercentiles []float64,
	tendermintBlockResult *tmrpctypes.ResultBlockResults,
	targetOneFeeHistory *rpctypes.OneFeeHistory) error {
	blockHeight := tendermintBlock.Block.Height
	blockBaseFee, err := e.BaseFee(blockHeight)
	if err != nil {
		return err
	}

	// set basefee
	targetOneFeeHistory.BaseFee = blockBaseFee

	// set gasused ratio
	gasLimitUint64 := (*ethBlock)["gasLimit"].(hexutil.Uint64)
	gasUsedBig := (*ethBlock)["gasUsed"].(*hexutil.Big)
	gasusedfloat, _ := new(big.Float).SetInt(gasUsedBig.ToInt()).Float64()

	if gasLimitUint64 <= 0 {
		return fmt.Errorf("gasLimit of block height %d should be bigger than 0 , current gaslimit %d", blockHeight, gasLimitUint64)
	}

	gasUsedRatio := gasusedfloat / float64(gasLimitUint64)
	blockGasUsed := gasusedfloat
	targetOneFeeHistory.GasUsedRatio = gasUsedRatio

	rewardCount := len(rewardPercentiles)
	targetOneFeeHistory.Reward = make([]*big.Int, rewardCount)
	for i := 0; i < rewardCount; i++ {
		targetOneFeeHistory.Reward[i] = big.NewInt(2000)
	}

	// check tendermintTxs
	tendermintTxs := tendermintBlock.Block.Txs
	tendermintTxResults := tendermintBlockResult.TxsResults
	tendermintTxCount := len(tendermintTxs)
	sorter := make(sortGasAndReward, tendermintTxCount)

	for i := 0; i < tendermintTxCount; i++ {
		eachTendermintTx := tendermintTxs[i]
		eachTendermintTxResult := tendermintTxResults[i]

		tx, err := e.clientCtx.TxConfig.TxDecoder()(eachTendermintTx)
		if err != nil {
			e.logger.Debug("failed to decode transaction in block", "height", blockHeight, "error", err.Error())
			continue
		}
		txGasUsed := uint64(eachTendermintTxResult.GasUsed)
		for _, msg := range tx.GetMsgs() {
			ethMsg, ok := msg.(*evmtypes.MsgEthereumTx)
			if !ok {
				continue
			}
			tx := ethMsg.AsTransaction()
			reward := tx.EffectiveGasTipValue(blockBaseFee)
			sorter[i] = txGasAndReward{gasUsed: txGasUsed, reward: reward}
			break
		}
	}
	sort.Sort(sorter)

	var txIndex int
	sumGasUsed := uint64(0)
	if len(sorter) > 0 {
		sumGasUsed = sorter[0].gasUsed
	}
	for i, p := range rewardPercentiles {
		thresholdGasUsed := uint64(blockGasUsed * p / 100)
		for sumGasUsed < thresholdGasUsed && txIndex < tendermintTxCount-1 {
			txIndex++
			sumGasUsed += sorter[txIndex].gasUsed
		}

		chosenReward := big.NewInt(0)
		if 0 <= txIndex && txIndex < len(sorter) {
			chosenReward = sorter[txIndex].reward
		}
		targetOneFeeHistory.Reward[i] = chosenReward
	}

	return nil
}

func (e *EVMBackend) FeeHistory(
	userBlockCount rpc.DecimalOrHex, // number blocks to fetch, maximum is 100
	lastBlock rpc.BlockNumber, // the block to start search , to oldest
	rewardPercentiles []float64, // percentiles to fetch reward
) (*rpctypes.FeeHistoryResult, error) {
	blockEnd := int64(lastBlock)

	if blockEnd <= 0 {
		blockNumber, err := e.BlockNumber()
		if err != nil {
			return nil, err
		}
		blockEnd = int64(blockNumber)
	}
	userBlockCountInt := int64(userBlockCount)
	maxBlockCount := int64(e.cfg.JSONRPC.FeeHistoryCap)
	if userBlockCountInt > maxBlockCount {
		return nil, fmt.Errorf("FeeHistory user block count %d higher than %d", userBlockCountInt, maxBlockCount)
	}
	blockStart := blockEnd - userBlockCountInt
	if blockStart < 0 {
		blockStart = 0
	}

	blockCount := blockEnd - blockStart

	oldestBlock := (*hexutil.Big)(big.NewInt(blockStart))

	// prepare space
	reward := make([][]*hexutil.Big, blockCount)
	rewardcount := len(rewardPercentiles)
	for i := 0; i < int(blockCount); i++ {
		reward[i] = make([]*hexutil.Big, rewardcount)
	}
	thisBaseFee := make([]*hexutil.Big, blockCount)
	thisGasUsedRatio := make([]float64, blockCount)

	// fetch block
	for blockID := blockStart; blockID < blockEnd; blockID++ {
		index := int32(blockID - blockStart)
		// eth block
		ethBlock, err := e.GetBlockByNumber(rpctypes.BlockNumber(blockID), true)
		if ethBlock == nil {
			return nil, err
		}

		// tendermint block
		tendermintblock, err := e.GetTendermintBlockByNumber(rpctypes.BlockNumber(blockID))
		if tendermintblock == nil {
			return nil, err
		}

		// tendermint block result
		tendermintBlockResult, err := e.clientCtx.Client.BlockResults(e.ctx, &tendermintblock.Block.Height)
		if tendermintBlockResult == nil {
			e.logger.Debug("block result not found", "height", tendermintblock.Block.Height, "error", err.Error())
			return nil, err
		}

		onefeehistory := rpctypes.OneFeeHistory{}
		err = e.processBlock(tendermintblock, &ethBlock, rewardPercentiles, tendermintBlockResult, &onefeehistory)
		if err != nil {
			return nil, err
		}

		// copy
		thisBaseFee[index] = (*hexutil.Big)(onefeehistory.BaseFee)
		thisGasUsedRatio[index] = onefeehistory.GasUsedRatio
		for j := 0; j < rewardcount; j++ {
			reward[index][j] = (*hexutil.Big)(onefeehistory.Reward[j])
		}
	}

	feeHistory := rpctypes.FeeHistoryResult{
		OldestBlock:  oldestBlock,
		Reward:       reward,
		BaseFee:      thisBaseFee,
		GasUsedRatio: thisGasUsedRatio,
	}
	return &feeHistory, nil
}
