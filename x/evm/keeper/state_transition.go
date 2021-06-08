package keeper

import (
	"errors"
	"fmt"
	"math/big"
	"os"
	"time"

	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	"github.com/cosmos/ethermint/x/evm/types"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/params"
)

// NewEVM generates an ethereum VM from the provided Message fields and the ChainConfig.
func (k *Keeper) NewEVM(msg core.Message, config *params.ChainConfig) *vm.EVM {
	blockCtx := vm.BlockContext{
		CanTransfer: core.CanTransfer,
		Transfer:    core.Transfer,
		GetHash:     k.GetHashFn(),
		Coinbase:    common.Address{}, // there's no beneficiary since we're not mining
		GasLimit:    k.ctx.BlockGasMeter().Limit(),
		BlockNumber: big.NewInt(k.ctx.BlockHeight()),
		Time:        big.NewInt(k.ctx.BlockHeader().Time.Unix()),
		Difficulty:  big.NewInt(0), // unused. Only required in PoW context
	}

	txCtx := core.NewEVMTxContext(msg)
	vmConfig := k.VMConfig()

	return vm.NewEVM(blockCtx, txCtx, k, config, vmConfig)
}

// VMConfig creates an EVM configuration from the module parameters and the debug setting.
// The config generated uses the default JumpTable from the EVM.
func (k Keeper) VMConfig() vm.Config {
	params := k.GetParams(k.ctx)

	return vm.Config{
		Debug:       k.debug,
		Tracer:      vm.NewJSONLogger(&vm.LogConfig{Debug: k.debug}, os.Stderr), // TODO: consider using the Struct Logger too
		NoRecursion: false,                                                      // TODO: consider disabling recursion though params
		ExtraEips:   params.EIPs(),
	}
}

// GetHashFn implements vm.GetHashFunc for Ethermint. It handles 3 cases:
//  1. The requested height matches the current height from context (and thus same epoch number)
//  2. The requested height is from an previous height from the same chain epoch
//  3. The requested height is from a height greater than the latest one
func (k Keeper) GetHashFn() vm.GetHashFunc {
	return func(height uint64) common.Hash {
		h := int64(height)
		switch {
		case k.ctx.BlockHeight() == h:
			// Case 1: The requested height matches the one from the context so we can retrieve the header
			// hash directly from the context.
			// TODO: deprecate field from the keeper on next SDK release
			return k.headerHash

		case k.ctx.BlockHeight() > h:
			// Case 2: if the chain is not the current height we need to retrieve the hash from the store for the
			// current chain epoch. This only applies if the current height is greater than the requested height.
			histInfo, found := k.stakingKeeper.GetHistoricalInfo(k.ctx, h)
			if !found {
				k.Logger(k.ctx).Debug("historical info not found", "height", h)
				return common.Hash{}
			}

			header, err := tmtypes.HeaderFromProto(&histInfo.Header)
			if err != nil {
				k.Logger(k.ctx).Error("failed to cast tendermint header from proto", "error", err)
				return common.Hash{}
			}

			return common.BytesToHash(header.Hash())
		default:
			// Case 3: heights greater than the current one returns an empty hash.
			return common.Hash{}
		}
	}
}

// TransitionDb runs and attempts to perform a state transition with the given transaction (i.e Message), that will
// only be persisted to the underlying KVStore if the transaction does not error.
//
// Gas tracking
//
// Ethereum consumes gas according to the EVM opcodes instead of general reads and writes to store. Because of this, the
// state transition needs to ignore the SDK gas consumption mechanism defined by the GasKVStore and instead consume the
// amount of gas used by the VM execution. The amount of gas used is tracked by the EVM and returned in the execution
// result.
//
// Prior to the execution, the starting tx gas meter is saved and replaced with an infinite gas meter in a new context
// in order to ignore the SDK gas consumption config values (read, write, has, delete).
// After the execution, the gas used from the message execution will be added to the starting gas consumed, taking into
// consideration the amount of gas returned. Finally, the context is updated with the EVM gas consumed value prior to
// returning.
//
// For relevant discussion see: https://github.com/cosmos/cosmos-sdk/discussions/9072
func (k *Keeper) TransitionDb(msg core.Message) (*types.ExecutionResult, error) {
	defer telemetry.ModuleMeasureSince(types.ModuleName, time.Now(), types.MetricKeyTransitionDB)

	cfg, found := k.GetChainConfig(k.ctx)
	if !found {
		return nil, types.ErrChainConfigNotFound
	}

	evm := k.NewEVM(msg, cfg.EthereumConfig(k.eip155ChainID))

	// create an ethereum StateTransition instance and run TransitionDb
	result, err := k.ApplyMessage(evm, msg)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// Gas consumption notes (write doc from this)

// gas = remaining gas = limit - consumed

// Gas consumption in ethereum:
// 0. Buy gas -> deduct gasLimit * gasPrice from user account
// 		0.1 leftover gas = gas limit
// 1. consume intrinsic gas
//   1.1 leftover gas = leftover gas - intrinsic gas
// 2. Exec vm functions by passing the gas (i.e remaining gas)
//   2.1 final leftover gas returned after spending gas from the opcodes jump tables
// 3. Refund amount =  max(gasConsumed / 2, gas refund), where gas refund is a local variable

// TODO: (@fedekunze) currently we consume the entire gas limit in the ante handler, so if a transaction fails
// the amount spent will be grater than the gas spent in an Ethereum tx (i.e here the leftover gas won't be refunded).

func (k *Keeper) ApplyMessage(evm *vm.EVM, msg core.Message) (*types.ExecutionResult, error) {
	var (
		ret        []byte // return bytes from evm execution
		vmErr, err error  // vm errors do not effect consensus and are therefore not assigned to err
	)

	sender := vm.AccountRef(msg.From())
	contractCreation := msg.To() == nil

	// transaction gas meter (tracks limit and usage)
	gasConsumed := k.ctx.GasMeter().GasConsumed()
	leftoverGas := k.ctx.GasMeter().Limit() - k.ctx.GasMeter().GasConsumedToLimit()

	// NOTE: Since CRUD operations on the SDK store consume gasm we need to set up an infinite gas meter so that we only consume
	// the gas used by the Ethereum message execution.
	// Not setting the infinite gas meter here would mean that we are incurring in additional gas costs
	k.WithContext(k.ctx.WithGasMeter(sdk.NewInfiniteGasMeter()))

	// NOTE: gas limit is the GasLimit defied in the message minus the Intrinsic Gas that has already been
	// consumed on the AnteHandler.

	// ensure gas is consistent during CheckTx
	if k.ctx.IsCheckTx() {
		if err := k.checkGasConsumption(msg, gasConsumed, contractCreation); err != nil {
			return nil, err
		}
	}

	if contractCreation {
		ret, _, leftoverGas, vmErr = evm.Create(sender, msg.Data(), leftoverGas, msg.Value())
	} else {
		ret, leftoverGas, vmErr = evm.Call(sender, *msg.To(), msg.Data(), leftoverGas, msg.Value())
	}

	// refund gas prior to handling the vm error in order to set the updated gas meter
	gasConsumed, leftoverGas, err = k.refundGas(msg, leftoverGas)
	if err != nil {
		return nil, err
	}

	if vmErr != nil {
		if errors.Is(vmErr, vm.ErrExecutionReverted) {
			// unpack the return data bytes from the err if the execution has been reverted on the VM
			return nil, types.NewExecErrorWithReson(ret)
		}

		// wrap the VM error
		return nil, sdkerrors.Wrap(types.ErrVMExecution, vmErr.Error())
	}

	return &types.ExecutionResult{
		Response: &types.MsgEthereumTxResponse{
			Ret: ret,
		},
		GasInfo: types.GasInfo{
			GasLimit:    k.ctx.GasMeter().Limit(),
			GasConsumed: gasConsumed,
			GasRefunded: leftoverGas,
		},
	}, nil
}

// checkGasConsumption verifies that the amount of gas consumed so far matches the intrinsic gas value.
func (k *Keeper) checkGasConsumption(msg core.Message, gasConsumed uint64, isContractCreation bool) error {
	cfg, _ := k.GetChainConfig(k.ctx)
	ethCfg := cfg.EthereumConfig(k.eip155ChainID)

	height := big.NewInt(k.ctx.BlockHeight())
	homestead := ethCfg.IsHomestead(height)
	istanbul := ethCfg.IsIstanbul(height)

	intrinsicGas, err := core.IntrinsicGas(msg.Data(), msg.AccessList(), isContractCreation, homestead, istanbul)
	if err != nil {
		// should have already been checked on Ante Handler
		return err
	}

	if intrinsicGas != gasConsumed {
		return fmt.Errorf("inconsistent gas. Expected gas consumption to be %d (intrinsic gas only), got %d", intrinsicGas, gasConsumed)
	}

	return nil
}

// refundGas transfers the leftover gas to the sender of the message, caped to half of the total gas
// consumed in the transaction. Additionally, the function sets the total gas consumed to the value
// returned by the EVM execution, thus ignoring the previous intrinsic gas inconsumed during in the
// AnteHandler.
func (k *Keeper) refundGas(msg core.Message, leftoverGas uint64) (consumed, leftover uint64, err error) {
	gasConsumed := msg.Gas() - leftoverGas

	// Apply refund counter, capped to half of the used gas.
	refund := gasConsumed / 2
	if refund > k.GetRefund() {
		refund = k.GetRefund()
	}

	leftoverGas += refund
	gasConsumed = msg.Gas() - leftoverGas

	// Return EVM tokens for remaining gas, exchanged at the original rate.
	remaining := new(big.Int).Mul(new(big.Int).SetUint64(leftoverGas), msg.GasPrice())

	switch remaining.Sign() {
	case -1:
		// negative refund errors
		return 0, 0, fmt.Errorf("refunded amount value cannot be negative %d", remaining.Int64())
	case 1:
		// positive amount refund
		params := k.GetParams(k.ctx)
		refundedCoins := sdk.Coins{sdk.NewCoin(params.EvmDenom, sdk.NewIntFromBigInt(remaining))}

		// refund to sender from the fee collector module account, which is the escrow account in charge of collecting tx fees
		if err := k.bankKeeper.SendCoinsFromModuleToAccount(k.ctx, authtypes.FeeCollectorName, msg.From().Bytes(), refundedCoins); err != nil {
			return 0, 0, sdkerrors.Wrapf(sdkerrors.ErrInsufficientFunds, "fee collector account failed to refund fees: %s", err.Error())
		}
	default:
		// no refund, consume gas and update the tx gas meter
	}

	// set the gas consumed into the context with the new gas meter. This gas meter will have the
	// original gas limit defined in the msg and will consume the gas now that the amount has been
	// refunded
	gasMeter := sdk.NewGasMeter(msg.Gas())
	gasMeter.ConsumeGas(gasConsumed, "update gas consumption after refund")
	k.WithContext(k.ctx.WithGasMeter(gasMeter))

	return gasConsumed, leftoverGas, nil
}
