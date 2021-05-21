package keeper

import (
	"math/big"
	"os"
	"time"

	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/ethermint/x/evm/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/params"
)

func (k *Keeper) NewEVM(msg core.Message, config *params.ChainConfig, gasLimit uint64) *vm.EVM {
	blockCtx := vm.BlockContext{
		CanTransfer: core.CanTransfer,
		Transfer:    core.Transfer,
		GetHash:     k.GetHashFn(),
		Coinbase:    common.Address{}, // there's no beneficiary since we're not mining
		GasLimit:    gasLimit,
		BlockNumber: big.NewInt(k.ctx.BlockHeight()),
		Time:        big.NewInt(k.ctx.BlockHeader().Time.Unix()),
		Difficulty:  big.NewInt(0), // unused. Only required in PoW context
	}

	txCtx := core.NewEVMTxContext(msg)
	vmConfig := k.VMConfig()

	return vm.NewEVM(blockCtx, txCtx, k, config, vmConfig)
}

func (k Keeper) VMConfig() vm.Config {
	params := k.GetParams(k.ctx)

	eips := make([]int, len(params.ExtraEIPs))
	for i, eip := range params.ExtraEIPs {
		eips[i] = int(eip)
	}

	// TODO: define on keeper fields
	debug := false

	return vm.Config{
		ExtraEips: eips,
		Tracer:    vm.NewJSONLogger(&vm.LogConfig{Debug: debug}, os.Stderr),
		Debug:     debug,
	}
}

// GetHashFn implements vm.GetHashFunc for Ethermint. It handles 3 cases:
//  1. The requested height matches the current height from context (and thus same epoch number)
//  2. The requested height is from an previous height from the same chain epoch
//  3. The requested height is from a height greater than the latest one
func (k Keeper) GetHashFn() vm.GetHashFunc {
	return func(height uint64) common.Hash {
		switch {
		case k.ctx.BlockHeight() == int64(height):
			// Case 1: The requested height matches the one from the context so we can retrieve the header
			// hash directly from the context.
			return k.cache.blockHash

		case k.ctx.BlockHeight() > int64(height):
			// Case 2: if the chain is not the current height we need to retrieve the hash from the store for the
			// current chain epoch. This only applies if the current height is greater than the requested height.
			return k.GetHeightHash(k.ctx, height)

		default:
			// Case 3: heights greater than the current one returns an empty hash.
			return common.Hash{}
		}
	}
}

func (k *Keeper) TransitionDb(ctx sdk.Context, msg core.Message) (*types.ExecutionResult, error) {
	defer telemetry.ModuleMeasureSince(types.ModuleName, time.Now(), types.MetricKeyTransitionDB)

	// transaction gas meter (tracks limit and usage)
	initialGasMeter := ctx.GasMeter()

	// NOTE: Since CRUD operations on the SDK store consume gasm we need to set up an infinite gas meter so that we only consume
	// the gas used by the Ethereum message execution.
	// Not setting the infinite gas meter here would mean that we are incurring in additional gas costs
	k.ctx = ctx.WithGasMeter(sdk.NewInfiniteGasMeter())

	cfg, found := k.GetChainConfig(ctx)
	if !found {
		// error
	}

	gasLimit := uint64(0) // TODO: define
	evm := k.NewEVM(msg, cfg.EthereumConfig(k.eip155ChainID), gasLimit)
	gasPool := core.GasPool(ctx.BlockGasMeter().Limit()) // available gas left in the block for the tx execution

	// create an ethereum StateTransition instance and run TransitionDb
	result, err := core.ApplyMessage(evm, msg, &gasPool)
	// return precheck errors (nonce, signature, balance and gas)
	// NOTE: these should be checked previously on the AnteHandler
	if err != nil {
		// log error
		return nil, err
	}

	// The gas used on the state transition will
	// be returned in the execution result so we need to deduct it from the transaction (?) GasMeter // TODO: double-check

	initialGasMeter.ConsumeGas(result.UsedGas-k.cache.refund, "evm state transition")

	// set the gas meter to current_gas_consumed = initial_gas + used_gas - refund
	k.ctx = k.ctx.WithGasMeter(initialGasMeter)

	// return the VM Execution error (see go-ethereum/core/vm/errors.go)

	revertReason := result.Revert()
	reverted := len(revertReason) > 0

	if reverted {
		// log error
		reason, errUnpack := abi.UnpackRevert(revertReason)
		var keyvals []string
		if errUnpack == nil {
			keyvals = []string{"reason", reason}
		} else {
			keyvals = []string{"hex-value", hexutil.Encode(revertReason)}
		}

		k.Logger(ctx).Error(vm.ErrExecutionReverted.Error(), keyvals)
	}

	executionRes := &types.ExecutionResult{
		Response: &types.MsgEthereumTxResponse{
			Ret:      result.ReturnData,
			Reverted: reverted,
		},
		GasInfo: types.GasInfo{
			GasConsumed: result.UsedGas,
			GasLimit:    uint64(gasPool),
		},
	}

	return executionRes, nil
}
