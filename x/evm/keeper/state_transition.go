package keeper

import (
	"math/big"
	"os"
	"time"

	"github.com/palantir/stacktrace"
	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	ethermint "github.com/tharsis/ethermint/types"
	"github.com/tharsis/ethermint/x/evm/types"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/params"
)

// NewEVM generates an ethereum VM from the provided Message fields and the chain parameters
// (config). It sets the validator operator address as the coinbase address to make it available for
// the COINBASE opcode, even though there is no beneficiary (since we're not mining).
func (k *Keeper) NewEVM(msg core.Message, config *params.ChainConfig, params types.Params, coinbase common.Address) *vm.EVM {
	blockCtx := vm.BlockContext{
		CanTransfer: core.CanTransfer,
		Transfer:    core.Transfer,
		GetHash:     k.GetHashFn(),
		Coinbase:    coinbase,
		GasLimit:    ethermint.BlockGasLimit(k.ctx),
		BlockNumber: big.NewInt(k.ctx.BlockHeight()),
		Time:        big.NewInt(k.ctx.BlockHeader().Time.Unix()),
		Difficulty:  big.NewInt(0), // unused. Only required in PoW context
	}

	txCtx := core.NewEVMTxContext(msg)
	vmConfig := k.VMConfig(params)

	return vm.NewEVM(blockCtx, txCtx, k, config, vmConfig)
}

// VMConfig creates an EVM configuration from the module parameters and the debug setting.
// The config generated uses the default JumpTable from the EVM.
func (k Keeper) VMConfig(params types.Params) vm.Config {
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
			// Note: The headerHash is only set at begin block, it will be nil in case of a query context
			headerHash := k.ctx.HeaderHash()
			if len(headerHash) != 0 {
				return common.BytesToHash(headerHash)
			}

			// only recompute the hash if not set
			contextBlockHeader := k.ctx.BlockHeader()
			header, err := tmtypes.HeaderFromProto(&contextBlockHeader)
			if err != nil {
				k.Logger(k.ctx).Error("failed to cast tendermint header from proto", "error", err)
				return common.Hash{}
			}
			headerHash = header.Hash()
			return common.BytesToHash(headerHash)

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

// ApplyTransaction runs and attempts to perform a state transition with the given transaction (i.e Message), that will
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
func (k *Keeper) ApplyTransaction(tx *ethtypes.Transaction) (*types.MsgEthereumTxResponse, error) {
	defer telemetry.ModuleMeasureSince(types.ModuleName, time.Now(), types.MetricKeyTransitionDB)

	params := k.GetParams(k.ctx)
	ethCfg := params.ChainConfig.EthereumConfig(k.eip155ChainID)

	// get the latest signer according to the chain rules from the config
	signer := ethtypes.MakeSigner(ethCfg, big.NewInt(k.ctx.BlockHeight()))

	msg, err := tx.AsMessage(signer)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to return ethereum transaction as core message")
	}

	// create an ethereum StateTransition instance and run TransitionDb
	// we use a ctx context to avoid modifying to state in case EVM msg is reverted
	originalCtx := k.ctx
	cacheCtx, commit := k.ctx.CacheContext()
	k.ctx = cacheCtx

	// get the coinbase address from the block proposer
	coinbase, err := k.GetCoinbaseAddress()
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to obtain coinbase address")
	}

	evm := k.NewEVM(msg, ethCfg, params, coinbase)

	k.SetTxHashTransient(tx.Hash())
	k.IncreaseTxIndexTransient()
	// pass false to execute in real mode, which do actual gas refunding
	res, err := k.ApplyMessage(evm, msg, ethCfg, false)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to apply ethereum core message")
	}

	txHash := tx.Hash()
	res.Hash = txHash.Hex()
	logs := k.GetTxLogs(txHash)

	// Commit and switch to original context
	if !res.Failed() {
		commit()
	}
	k.ctx = originalCtx

	// Logs needs to be ignored when tx is reverted
	// Set the log and bloom filter only when the tx is NOT REVERTED
	if !res.Failed() {
		res.Logs = types.NewLogsFromEth(logs)
		// Update block bloom filter in the original context because blockbloom is set in EndBlock
		bloom := k.GetBlockBloomTransient()
		bloom.Or(bloom, big.NewInt(0).SetBytes(ethtypes.LogsBloom(logs)))
		k.SetBlockBloomTransient(bloom)
	}

	// update the gas used after refund
	k.resetGasMeterAndConsumeGas(res.GasUsed)
	return res, nil
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

// ApplyMessage computes the new state by applying the given message against the existing state.
// If the message fails, the VM execution error with the reason will be returned to the client
// and the transaction won't be committed to the store.
//
// Reverted state
//
// The transaction is never "reverted" since there is no snapshot + rollback performed on the StateDB.
// Only successful transactions are written to the store during DeliverTx mode.
//
// Prechecks and Preprocessing
//
// All relevant state transition prechecks for the MsgEthereumTx are performed on the AnteHandler,
// prior to running the transaction against the state. The prechecks run are the following:
//
// 1. the nonce of the message caller is correct
// 2. caller has enough balance to cover transaction fee(gaslimit * gasprice)
// 3. the amount of gas required is available in the block
// 4. the purchased gas is enough to cover intrinsic usage
// 5. there is no overflow when calculating intrinsic gas
// 6. caller has enough balance to cover asset transfer for **topmost** call
//
// The preprocessing steps performed by the AnteHandler are:
//
// 1. set up the initial access list (iff fork > Berlin)
//
// Query mode
//
// The grpc query endpoint EthCall calls this in query mode, and since the query handler don't call AnteHandler,
// so we don't do real gas refund in that case.
func (k *Keeper) ApplyMessage(evm *vm.EVM, msg core.Message, cfg *params.ChainConfig, query bool) (*types.MsgEthereumTxResponse, error) {
	var (
		ret   []byte // return bytes from evm execution
		vmErr error  // vm errors do not effect consensus and are therefore not assigned to err
	)

	sender := vm.AccountRef(msg.From())
	contractCreation := msg.To() == nil

	intrinsicGas, err := k.GetEthIntrinsicGas(msg, cfg, contractCreation)
	if err != nil {
		// should have already been checked on Ante Handler
		return nil, stacktrace.Propagate(err, "intrinsic gas failed")
	}
	// Should check again even if it is checked on Ante Handler, because eth_call don't go through Ante Handler.
	if msg.Gas() < intrinsicGas {
		// eth_estimateGas will check for this exact error
		return nil, stacktrace.Propagate(core.ErrIntrinsicGas, "intrinsic gas too low")
	}
	leftoverGas := msg.Gas() - intrinsicGas

	if contractCreation {
		ret, _, leftoverGas, vmErr = evm.Create(sender, msg.Data(), leftoverGas, msg.Value())
	} else {
		ret, leftoverGas, vmErr = evm.Call(sender, *msg.To(), msg.Data(), leftoverGas, msg.Value())
	}

	if query {
		// query handlers don't call ante handler to deduct gas fee, so don't do actual refund here, because the
		// module account balance might not be enough
		leftoverGas += k.GasToRefund(msg.Gas() - leftoverGas)
	} else {
		// refund gas prior to handling the vm error in order to set the updated gas meter
		leftoverGas, err = k.RefundGas(msg, leftoverGas)
		if err != nil {
			return nil, stacktrace.Propagate(err, "failed to refund gas leftover gas to sender %s", msg.From())
		}
	}

	var vmError string
	if vmErr != nil {
		vmError = vmErr.Error()
	}

	gasUsed := msg.Gas() - leftoverGas
	return &types.MsgEthereumTxResponse{
		GasUsed: gasUsed,
		VmError: vmError,
		Ret:     ret,
	}, nil
}

// GetEthIntrinsicGas get the transaction intrinsic gas cost
func (k *Keeper) GetEthIntrinsicGas(msg core.Message, cfg *params.ChainConfig, isContractCreation bool) (uint64, error) {
	height := big.NewInt(k.ctx.BlockHeight())
	homestead := cfg.IsHomestead(height)
	istanbul := cfg.IsIstanbul(height)

	return core.IntrinsicGas(msg.Data(), msg.AccessList(), isContractCreation, homestead, istanbul)
}

// GasToRefund calculate the amount of gas should refund to sender
func (k *Keeper) GasToRefund(gasConsumed uint64) uint64 {
	// Apply refund counter, capped to half of the used gas.
	refund := gasConsumed / 2
	availableRefund := k.GetRefund()
	if refund > availableRefund {
		return availableRefund
	}
	return refund
}

// RefundGas transfers the leftover gas to the sender of the message, caped to half of the total gas
// consumed in the transaction. Additionally, the function sets the total gas consumed to the value
// returned by the EVM execution, thus ignoring the previous intrinsic gas consumed during in the
// AnteHandler.
func (k *Keeper) RefundGas(msg core.Message, leftoverGas uint64) (uint64, error) {
	if leftoverGas > msg.Gas() {
		return leftoverGas, stacktrace.Propagate(
			sdkerrors.Wrapf(types.ErrInconsistentGas, "leftover gas cannot be greater than gas limit (%d > %d)", leftoverGas, msg.Gas()),
			"failed to update gas consumed after refund of leftover gas",
		)
	}

	gasConsumed := msg.Gas() - leftoverGas
	refund := k.GasToRefund(gasConsumed)

	leftoverGas += refund

	if leftoverGas > msg.Gas() {
		return leftoverGas, stacktrace.Propagate(
			sdkerrors.Wrapf(types.ErrInconsistentGas, "leftover gas cannot be greater than gas limit (%d > %d)", leftoverGas, msg.Gas()),
			"failed to update gas consumed after refund of %d gas", refund,
		)
	}

	// Return EVM tokens for remaining gas, exchanged at the original rate.
	remaining := new(big.Int).Mul(new(big.Int).SetUint64(leftoverGas), msg.GasPrice())

	switch remaining.Sign() {
	case -1:
		// negative refund errors
		return leftoverGas, sdkerrors.Wrapf(types.ErrInvalidRefund, "refunded amount value cannot be negative %d", remaining.Int64())
	case 1:
		// positive amount refund
		params := k.GetParams(k.ctx)
		refundedCoins := sdk.Coins{sdk.NewCoin(params.EvmDenom, sdk.NewIntFromBigInt(remaining))}

		// refund to sender from the fee collector module account, which is the escrow account in charge of collecting tx fees

		err := k.bankKeeper.SendCoinsFromModuleToAccount(k.ctx, authtypes.FeeCollectorName, msg.From().Bytes(), refundedCoins)
		if err != nil {
			err = sdkerrors.Wrapf(sdkerrors.ErrInsufficientFunds, "fee collector account failed to refund fees: %s", err.Error())
			return leftoverGas, stacktrace.Propagate(err, "failed to refund %d leftover gas (%s)", leftoverGas, refundedCoins.String())
		}
	default:
		// no refund, consume gas and update the tx gas meter
	}

	return leftoverGas, nil
}

// resetGasMeterAndConsumeGas reset first the gas meter consumed value to zero and set it back to the new value
// 'gasUsed'
func (k *Keeper) resetGasMeterAndConsumeGas(gasUsed uint64) {
	// reset the gas count
	k.ctx.GasMeter().RefundGas(k.ctx.GasMeter().GasConsumed(), "reset the gas count")
	k.ctx.GasMeter().ConsumeGas(gasUsed, "apply evm transaction")
}

// GetCoinbaseAddress returns the block proposer's validator operator address.
func (k Keeper) GetCoinbaseAddress() (common.Address, error) {
	consAddr := sdk.ConsAddress(k.ctx.BlockHeader().ProposerAddress)
	validator, found := k.stakingKeeper.GetValidatorByConsAddr(k.ctx, consAddr)
	if !found {
		return common.Address{}, stacktrace.Propagate(
			sdkerrors.Wrap(stakingtypes.ErrNoValidatorFound, consAddr.String()),
			"failed to retrieve validator from block proposer address",
		)
	}

	coinbase := common.BytesToAddress(validator.GetOperator())
	return coinbase, nil
}
