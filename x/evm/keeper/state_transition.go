package keeper

import (
	"math/big"

	"github.com/palantir/stacktrace"
	tmtypes "github.com/tendermint/tendermint/types"

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

// NewEVM generates a go-ethereum VM from the provided Message fields and the chain parameters
// (ChainConfig and module Params). It additionally sets the validator operator address as the
// coinbase address to make it available for the COINBASE opcode, even though there is no
// beneficiary of the coinbase transaction (since we're not mining).
func (k *Keeper) NewEVM(
	msg core.Message,
	config *params.ChainConfig,
	params types.Params,
	coinbase common.Address,
	tracer vm.Tracer,
) *vm.EVM {
	blockCtx := vm.BlockContext{
		CanTransfer: core.CanTransfer,
		Transfer:    core.Transfer,
		GetHash:     k.GetHashFn(),
		Coinbase:    coinbase,
		GasLimit:    ethermint.BlockGasLimit(k.Ctx()),
		BlockNumber: big.NewInt(k.Ctx().BlockHeight()),
		Time:        big.NewInt(k.Ctx().BlockHeader().Time.Unix()),
		Difficulty:  big.NewInt(0), // unused. Only required in PoW context
	}

	txCtx := core.NewEVMTxContext(msg)
	vmConfig := k.VMConfig(msg, params, tracer)

	return vm.NewEVM(blockCtx, txCtx, k, config, vmConfig)
}

// VMConfig creates an EVM configuration from the debug setting and the extra EIPs enabled on the
// module parameters. The config generated uses the default JumpTable from the EVM.
func (k Keeper) VMConfig(msg core.Message, params types.Params, tracer vm.Tracer) vm.Config {
	return vm.Config{
		Debug:       k.debug,
		Tracer:      tracer,
		NoRecursion: false, // TODO: consider disabling recursion though params
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
		ctx := k.Ctx()
		switch {
		case ctx.BlockHeight() == h:
			// Case 1: The requested height matches the one from the context so we can retrieve the header
			// hash directly from the context.
			// Note: The headerHash is only set at begin block, it will be nil in case of a query context
			headerHash := ctx.HeaderHash()
			if len(headerHash) != 0 {
				return common.BytesToHash(headerHash)
			}

			// only recompute the hash if not set (eg: checkTxState)
			contextBlockHeader := ctx.BlockHeader()
			header, err := tmtypes.HeaderFromProto(&contextBlockHeader)
			if err != nil {
				k.Logger(ctx).Error("failed to cast tendermint header from proto", "error", err)
				return common.Hash{}
			}

			headerHash = header.Hash()
			return common.BytesToHash(headerHash)

		case ctx.BlockHeight() > h:
			// Case 2: if the chain is not the current height we need to retrieve the hash from the store for the
			// current chain epoch. This only applies if the current height is greater than the requested height.
			histInfo, found := k.stakingKeeper.GetHistoricalInfo(ctx, h)
			if !found {
				k.Logger(ctx).Debug("historical info not found", "height", h)
				return common.Hash{}
			}

			header, err := tmtypes.HeaderFromProto(&histInfo.Header)
			if err != nil {
				k.Logger(ctx).Error("failed to cast tendermint header from proto", "error", err)
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
// only be persisted (committed) to the underlying KVStore if the transaction does not fail.
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
	ctx := k.Ctx()
	params := k.GetParams(ctx)

	// return error if contract creation or call are disabled through governance
	if !params.EnableCreate && tx.To() == nil {
		return nil, stacktrace.Propagate(types.ErrCreateDisabled, "failed to create new contract")
	} else if !params.EnableCall && tx.To() != nil {
		return nil, stacktrace.Propagate(types.ErrCallDisabled, "failed to call contract")
	}

	ethCfg := params.ChainConfig.EthereumConfig(k.eip155ChainID)

	// get the latest signer according to the chain rules from the config
	signer := ethtypes.MakeSigner(ethCfg, big.NewInt(ctx.BlockHeight()))

	msg, err := tx.AsMessage(signer)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to return ethereum transaction as core message")
	}

	// get the coinbase address from the block proposer
	coinbase, err := k.GetCoinbaseAddress(ctx)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to obtain coinbase address")
	}

	// create an ethereum EVM instance and run the message
	tracer := types.NewTracer(k.tracer, msg, ethCfg, ctx.BlockHeight(), k.debug)
	evm := k.NewEVM(msg, ethCfg, params, coinbase, tracer)

	txHash := tx.Hash()

	// set the transaction hash and index to the impermanent (transient) block state so that it's also
	// available on the StateDB functions (eg: AddLog)
	k.SetTxHashTransient(txHash)
	k.IncreaseTxIndexTransient()

	if !k.ctxStack.IsEmpty() {
		panic("context stack shouldn't be dirty before apply message")
	}

	var revision int
	if k.hooks != nil {
		// snapshot to contain the tx processing and post processing in same scope
		revision = k.Snapshot()
	}

	// pass false to execute in real mode, which do actual gas refunding
	res, err := k.ApplyMessage(evm, msg, ethCfg, false)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to apply ethereum core message")
	}

	res.Hash = txHash.Hex()
	logs := k.GetTxLogsTransient(txHash)

	if !res.Failed() {
		// Only call hooks if tx executed successfully.
		if err = k.PostTxProcessing(txHash, logs); err != nil {
			// If hooks return error, revert the whole tx.
			k.RevertToSnapshot(revision)
			res.VmError = types.ErrPostTxProcessing.Error()
			k.Logger(ctx).Error("tx post processing failed", "error", err)
		}
	}

	if len(logs) > 0 {
		res.Logs = types.NewLogsFromEth(logs)
		// Update transient block bloom filter
		bloom := k.GetBlockBloomTransient()
		bloom.Or(bloom, big.NewInt(0).SetBytes(ethtypes.LogsBloom(logs)))
		k.SetBlockBloomTransient(bloom)
	}

	// Since we've implemented `RevertToSnapshot` api, so for the vm error cases,
	// the state is reverted, so it's ok to call the commit here anyway.
	k.CommitCachedContexts()

	// update the gas used after refund
	k.resetGasMeterAndConsumeGas(res.GasUsed)
	return res, nil
}

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
// The gRPC query endpoint from 'eth_call' calls this method in query mode, and since the query handler don't call AnteHandler,
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
		return nil, stacktrace.Propagate(core.ErrIntrinsicGas, "apply message")
	}
	leftoverGas := msg.Gas() - intrinsicGas

	// access list preparaion is moved from ante handler to here, because it's needed when `ApplyMessage` is called
	// under contexts where ante handlers are not run, for example `eth_call` and `eth_estimateGas`.
	if rules := cfg.Rules(big.NewInt(k.Ctx().BlockHeight())); rules.IsBerlin {
		k.PrepareAccessList(msg.From(), msg.To(), vm.ActivePrecompiles(rules), msg.AccessList())
	}

	if contractCreation {
		ret, _, leftoverGas, vmErr = evm.Create(sender, msg.Data(), leftoverGas, msg.Value())
	} else {
		ret, leftoverGas, vmErr = evm.Call(sender, *msg.To(), msg.Data(), leftoverGas, msg.Value())
	}

	refundQuotient := uint64(2)

	if query {
		// gRPC query handlers don't go through the AnteHandler to deduct the gas fee from the sender or have access historical state.
		// We don't refund gas to the sender.
		// For more info, see: https://github.com/tharsis/ethermint/issues/229 and https://github.com/cosmos/cosmos-sdk/issues/9636
		gasConsumed := msg.Gas() - leftoverGas
		leftoverGas += k.GasToRefund(gasConsumed, refundQuotient)
	} else {
		// refund gas prior to handling the vm error in order to match the Ethereum gas consumption instead of the default SDK one.
		leftoverGas, err = k.RefundGas(msg, leftoverGas, refundQuotient)
		if err != nil {
			return nil, stacktrace.Propagate(err, "failed to refund gas leftover gas to sender %s", msg.From())
		}
	}

	// EVM execution error needs to be available for the JSON-RPC client
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

// GetEthIntrinsicGas returns the intrinsic gas cost for the transaction
func (k *Keeper) GetEthIntrinsicGas(msg core.Message, cfg *params.ChainConfig, isContractCreation bool) (uint64, error) {
	height := big.NewInt(k.Ctx().BlockHeight())
	homestead := cfg.IsHomestead(height)
	istanbul := cfg.IsIstanbul(height)

	return core.IntrinsicGas(msg.Data(), msg.AccessList(), isContractCreation, homestead, istanbul)
}

// GasToRefund calculates the amount of gas the state machine should refund to the sender. It is
// capped by the refund quotient value.
func (k *Keeper) GasToRefund(gasConsumed, refundQuotient uint64) uint64 {
	// Apply refund counter
	refund := gasConsumed / refundQuotient
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
func (k *Keeper) RefundGas(msg core.Message, leftoverGas, refundQuotient uint64) (uint64, error) {
	// safety check: leftover gas after execution should never exceed the gas limit defined on the message
	if leftoverGas > msg.Gas() {
		return leftoverGas, stacktrace.Propagate(
			sdkerrors.Wrapf(types.ErrInconsistentGas, "leftover gas cannot be greater than gas limit (%d > %d)", leftoverGas, msg.Gas()),
			"failed to update gas consumed after refund of leftover gas",
		)
	}

	gasConsumed := msg.Gas() - leftoverGas

	// calculate available gas to refund and add it to the leftover gas amount
	refund := k.GasToRefund(gasConsumed, refundQuotient)
	leftoverGas += refund

	// safety check: leftover gas after refund should never exceed the gas limit defined on the message
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
		params := k.GetParams(k.Ctx())
		refundedCoins := sdk.Coins{sdk.NewCoin(params.EvmDenom, sdk.NewIntFromBigInt(remaining))}

		// refund to sender from the fee collector module account, which is the escrow account in charge of collecting tx fees

		err := k.bankKeeper.SendCoinsFromModuleToAccount(k.Ctx(), authtypes.FeeCollectorName, msg.From().Bytes(), refundedCoins)
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
	ctx := k.Ctx()
	ctx.GasMeter().RefundGas(ctx.GasMeter().GasConsumed(), "reset the gas count")
	ctx.GasMeter().ConsumeGas(gasUsed, "apply evm transaction")
}

// GetCoinbaseAddress returns the block proposer's validator operator address.
func (k Keeper) GetCoinbaseAddress(ctx sdk.Context) (common.Address, error) {
	consAddr := sdk.ConsAddress(ctx.BlockHeader().ProposerAddress)
	validator, found := k.stakingKeeper.GetValidatorByConsAddr(ctx, consAddr)
	if !found {
		return common.Address{}, stacktrace.Propagate(
			sdkerrors.Wrap(stakingtypes.ErrNoValidatorFound, consAddr.String()),
			"failed to retrieve validator from block proposer address",
		)
	}

	coinbase := common.BytesToAddress(validator.GetOperator())
	return coinbase, nil
}
