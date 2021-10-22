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

// EVMConfig creates the EVMConfig based on current state
func (k *Keeper) EVMConfig(ctx sdk.Context) (*types.EVMConfig, error) {
	params := k.GetParams(ctx)
	ethCfg := params.ChainConfig.EthereumConfig(k.eip155ChainID)

	// get the coinbase address from the block proposer
	coinbase, err := k.GetCoinbaseAddress(ctx)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to obtain coinbase address")
	}

	var baseFee *big.Int
	if types.IsLondon(ethCfg, ctx.BlockHeight()) {
		baseFee = k.feeMarketKeeper.GetBaseFee(ctx)
	}

	return &types.EVMConfig{
		Params:      params,
		ChainConfig: ethCfg,
		CoinBase:    coinbase,
		BaseFee:     baseFee,
	}, nil
}

// NewEVM generates a go-ethereum VM from the provided Message fields and the chain parameters
// (ChainConfig and module Params). It additionally sets the validator operator address as the
// coinbase address to make it available for the COINBASE opcode, even though there is no
// beneficiary of the coinbase transaction (since we're not mining).
func (k *Keeper) NewEVM(
	msg core.Message,
	cfg *types.EVMConfig,
	tracer vm.Tracer,
) *vm.EVM {
	blockCtx := vm.BlockContext{
		CanTransfer: core.CanTransfer,
		Transfer:    core.Transfer,
		GetHash:     k.GetHashFn(),
		Coinbase:    cfg.CoinBase,
		GasLimit:    ethermint.BlockGasLimit(k.Ctx()),
		BlockNumber: big.NewInt(k.Ctx().BlockHeight()),
		Time:        big.NewInt(k.Ctx().BlockHeader().Time.Unix()),
		Difficulty:  big.NewInt(0), // unused. Only required in PoW context
		BaseFee:     cfg.BaseFee,
	}

	txCtx := core.NewEVMTxContext(msg)
	if tracer == nil {
		tracer = k.Tracer(msg, cfg.ChainConfig)
	}
	vmConfig := k.VMConfig(msg, cfg.Params, tracer)
	return vm.NewEVM(blockCtx, txCtx, k, cfg.ChainConfig, vmConfig)
}

// VMConfig creates an EVM configuration from the debug setting and the extra EIPs enabled on the
// module parameters. The config generated uses the default JumpTable from the EVM.
func (k Keeper) VMConfig(msg core.Message, params types.Params, tracer vm.Tracer) vm.Config {
	fmParams := k.feeMarketKeeper.GetParams(k.Ctx())

	return vm.Config{
		Debug:       k.debug,
		Tracer:      tracer,
		NoRecursion: false, // TODO: consider disabling recursion though params
		NoBaseFee:   fmParams.NoBaseFee,
		ExtraEips:   params.EIPs(),
	}
}

// GetHashFn implements vm.GetHashFunc for Ethermint. It handles 3 cases:
//  1. The requested height matches the current height from context (and thus same epoch number)
//  2. The requested height is from an previous height from the same chain epoch
//  3. The requested height is from a height greater than the latest one
func (k Keeper) GetHashFn() vm.GetHashFunc {
	return func(height uint64) common.Hash {
		ctx := k.Ctx()

		h, err := ethermint.SafeInt64(height)
		if err != nil {
			k.Logger(ctx).Error("failed to cast height to int64", "error", err)
			return common.Hash{}
		}

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

	// ensure keeper state error is cleared
	defer k.ClearStateError()

	cfg, err := k.EVMConfig(ctx)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to load evm config")
	}

	// get the latest signer according to the chain rules from the config
	signer := ethtypes.MakeSigner(cfg.ChainConfig, big.NewInt(ctx.BlockHeight()))

	var baseFee *big.Int
	if types.IsLondon(cfg.ChainConfig, ctx.BlockHeight()) {
		baseFee = k.feeMarketKeeper.GetBaseFee(ctx)
	}

	msg, err := tx.AsMessage(signer, baseFee)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to return ethereum transaction as core message")
	}

	txHash := tx.Hash()

	// set the transaction hash and index to the impermanent (transient) block state so that it's also
	// available on the StateDB functions (eg: AddLog)
	k.SetTxHashTransient(txHash)

	// snapshot to contain the tx processing and post processing in same scope
	var commit func()
	if k.hooks != nil {
		// Create a cache context to revert state when tx hooks fails,
		// the cache context is only committed when both tx and hooks executed successfully.
		// Didn't use `Snapshot` because the context stack has exponential complexity on certain operations,
		// thus restricted to be used only inside `ApplyMessage`.
		var cacheCtx sdk.Context
		cacheCtx, commit = ctx.CacheContext()
		k.WithContext(cacheCtx)
		defer (func() {
			k.WithContext(ctx)
		})()
	}

	res, err := k.ApplyMessageWithConfig(msg, nil, true, cfg)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to apply ethereum core message")
	}

	// refund gas prior to handling the vm error in order to match the Ethereum gas consumption instead of the default SDK one.
	err = k.RefundGas(msg, msg.Gas()-res.GasUsed, cfg.Params.EvmDenom)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to refund gas leftover gas to sender %s", msg.From())
	}

	res.Hash = txHash.Hex()

	logs := k.GetTxLogsTransient(txHash)

	if !res.Failed() {
		// Only call hooks if tx executed successfully.
		if err = k.PostTxProcessing(txHash, logs); err != nil {
			// If hooks return error, revert the whole tx.
			res.VmError = types.ErrPostTxProcessing.Error()
			k.Logger(k.Ctx()).Error("tx post processing failed", "error", err)
		} else if commit != nil {
			// PostTxProcessing is successful, commit the cache context
			commit()
			ctx.EventManager().EmitEvents(k.Ctx().EventManager().Events())
		}
	}

	if len(logs) > 0 {
		res.Logs = types.NewLogsFromEth(logs)
		// Update transient block bloom filter
		bloom := k.GetBlockBloomTransient()
		bloom.Or(bloom, big.NewInt(0).SetBytes(ethtypes.LogsBloom(logs)))
		k.SetBlockBloomTransient(bloom)
	}

	k.IncreaseTxIndexTransient()

	// update the gas used after refund
	k.ResetGasMeterAndConsumeGas(res.GasUsed)
	return res, nil
}

// ApplyMessageWithConfig computes the new state by applying the given message against the existing state.
// If the message fails, the VM execution error with the reason will be returned to the client
// and the transaction won't be committed to the store.
//
// Reverted state
//
// The snapshot and rollback are supported by the `ContextStack`, which should be only used inside `ApplyMessage`,
// because some operations has exponential computational complexity with deep stack.
//
// Different Callers
//
// It's called in three scenarios:
// 1. `ApplyTransaction`, in the transaction processing flow.
// 2. `EthCall/EthEstimateGas` grpc query handler.
// 3. Called by other native modules directly.
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
// Tracer parameter
//
// It should be a `vm.Tracer` object or nil, if pass `nil`, it'll create a default one based on keeper options.
//
// Commit parameter
//
// If commit is true, the cache context stack will be committed, otherwise discarded.
func (k *Keeper) ApplyMessageWithConfig(msg core.Message, tracer vm.Tracer, commit bool, cfg *types.EVMConfig) (*types.MsgEthereumTxResponse, error) {
	var (
		ret   []byte // return bytes from evm execution
		vmErr error  // vm errors do not effect consensus and are therefore not assigned to err
	)

	if !k.ctxStack.IsEmpty() {
		panic("context stack shouldn't be dirty before apply message")
	}

	evm := k.NewEVM(msg, cfg, tracer)

	// ensure keeper state error is cleared
	defer k.ClearStateError()

	// return error if contract creation or call are disabled through governance
	if !cfg.Params.EnableCreate && msg.To() == nil {
		return nil, stacktrace.Propagate(types.ErrCreateDisabled, "failed to create new contract")
	} else if !cfg.Params.EnableCall && msg.To() != nil {
		return nil, stacktrace.Propagate(types.ErrCallDisabled, "failed to call contract")
	}

	sender := vm.AccountRef(msg.From())
	contractCreation := msg.To() == nil
	isLondon := cfg.ChainConfig.IsLondon(evm.Context.BlockNumber)

	intrinsicGas, err := k.GetEthIntrinsicGas(msg, cfg.ChainConfig, contractCreation)
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
	if rules := cfg.ChainConfig.Rules(big.NewInt(k.Ctx().BlockHeight())); rules.IsBerlin {
		k.PrepareAccessList(msg.From(), msg.To(), vm.ActivePrecompiles(rules), msg.AccessList())
	}

	if contractCreation {
		ret, _, leftoverGas, vmErr = evm.Create(sender, msg.Data(), leftoverGas, msg.Value())
	} else {
		ret, leftoverGas, vmErr = evm.Call(sender, *msg.To(), msg.Data(), leftoverGas, msg.Value())
	}

	refundQuotient := params.RefundQuotient

	// After EIP-3529: refunds are capped to gasUsed / 5
	if isLondon {
		refundQuotient = params.RefundQuotientEIP3529
	}

	// calculate gas refund
	if msg.Gas() < leftoverGas {
		return nil, stacktrace.Propagate(types.ErrGasOverflow, "apply message")
	}
	gasUsed := msg.Gas() - leftoverGas
	refund := k.GasToRefund(gasUsed, refundQuotient)
	if refund > gasUsed {
		return nil, stacktrace.Propagate(types.ErrGasOverflow, "apply message")
	}
	gasUsed -= refund

	// EVM execution error needs to be available for the JSON-RPC client
	var vmError string
	if vmErr != nil {
		vmError = vmErr.Error()
	}

	// The context stack is designed specifically for `StateDB` interface, it should only be used in `ApplyMessage`,
	// after return, the stack should be clean, the cached states are either committed or discarded.
	if commit {
		k.CommitCachedContexts()
	} else {
		k.ctxStack.RevertAll()
	}

	return &types.MsgEthereumTxResponse{
		GasUsed: gasUsed,
		VmError: vmError,
		Ret:     ret,
	}, nil
}

// ApplyMessage calls ApplyMessageWithConfig with default EVMConfig
func (k *Keeper) ApplyMessage(msg core.Message, tracer vm.Tracer, commit bool) (*types.MsgEthereumTxResponse, error) {
	cfg, err := k.EVMConfig(k.Ctx())
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to load evm config")
	}
	return k.ApplyMessageWithConfig(msg, tracer, commit, cfg)
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
// Note: do not pass 0 to refundQuotient
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
func (k *Keeper) RefundGas(msg core.Message, leftoverGas uint64, denom string) error {
	// Return EVM tokens for remaining gas, exchanged at the original rate.
	remaining := new(big.Int).Mul(new(big.Int).SetUint64(leftoverGas), msg.GasPrice())

	switch remaining.Sign() {
	case -1:
		// negative refund errors
		return sdkerrors.Wrapf(types.ErrInvalidRefund, "refunded amount value cannot be negative %d", remaining.Int64())
	case 1:
		// positive amount refund
		refundedCoins := sdk.Coins{sdk.NewCoin(denom, sdk.NewIntFromBigInt(remaining))}

		// refund to sender from the fee collector module account, which is the escrow account in charge of collecting tx fees

		err := k.bankKeeper.SendCoinsFromModuleToAccount(k.Ctx(), authtypes.FeeCollectorName, msg.From().Bytes(), refundedCoins)
		if err != nil {
			err = sdkerrors.Wrapf(sdkerrors.ErrInsufficientFunds, "fee collector account failed to refund fees: %s", err.Error())
			return stacktrace.Propagate(err, "failed to refund %d leftover gas (%s)", leftoverGas, refundedCoins.String())
		}
	default:
		// no refund, consume gas and update the tx gas meter
	}

	return nil
}

// ResetGasMeterAndConsumeGas reset first the gas meter consumed value to zero and set it back to the new value
// 'gasUsed'
func (k *Keeper) ResetGasMeterAndConsumeGas(gasUsed uint64) {
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
