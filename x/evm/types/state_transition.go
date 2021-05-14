package types

import (
	"math/big"
	"os"
	"time"

	"github.com/pkg/errors"
	log "github.com/xlab/suplog"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// StateTransition defines data to transitionDB in evm
type StateTransition struct {
	// TxData fields
	Message core.Message

	ChainID  *big.Int
	Csdb     *CommitStateDB
	TxHash   *common.Hash
	Simulate bool // i.e CheckTx execution
	Debug    bool // enable EVM debugging
}

// GasInfo returns the gas limit, gas consumed and gas refunded from the EVM transition
// execution
type GasInfo struct {
	GasLimit    uint64
	GasConsumed uint64
	GasRefunded uint64
}

// ExecutionResult represents what's returned from a transition
type ExecutionResult struct {
	Logs     []*ethtypes.Log
	Bloom    *big.Int
	Response *MsgEthereumTxResponse
	GasInfo  GasInfo
}

// GetHashFn implements vm.GetHashFunc for Ethermint. It handles 3 cases:
//  1. The requested height matches the current height from context (and thus same epoch number)
//  2. The requested height is from an previous height from the same chain epoch
//  3. The requested height is from a height greater than the latest one
func GetHashFn(ctx sdk.Context, csdb *CommitStateDB) vm.GetHashFunc {
	return func(height uint64) common.Hash {
		switch {
		case ctx.BlockHeight() == int64(height):
			// Case 1: The requested height matches the one from the context so we can retrieve the header
			// hash directly from the context.
			return csdb.bhash

		case ctx.BlockHeight() > int64(height):
			// Case 2: if the chain is not the current height we need to retrieve the hash from the store for the
			// current chain epoch. This only applies if the current height is greater than the requested height.
			return csdb.WithContext(ctx).GetHeightHash(height)

		default:
			// Case 3: heights greater than the current one returns an empty hash.
			return common.Hash{}
		}
	}
}

func (st *StateTransition) newEVM(
	ctx sdk.Context,
	csdb *CommitStateDB,
	gasLimit uint64,
	config ChainConfig,
	extraEIPs []int64,
) *vm.EVM {
	// Create context for evm
	blockCtx := vm.BlockContext{
		CanTransfer: core.CanTransfer,
		Transfer:    core.Transfer,
		GetHash:     GetHashFn(ctx, csdb),
		Coinbase:    common.Address{}, // there's no beneficiary since we're not mining
		BlockNumber: big.NewInt(ctx.BlockHeight()),
		Time:        big.NewInt(ctx.BlockHeader().Time.Unix()),
		Difficulty:  big.NewInt(0), // unused. Only required in PoW context
		GasLimit:    gasLimit,
	}

	txCtx := core.NewEVMTxContext(st.Message)

	eips := make([]int, len(extraEIPs))
	for i, eip := range extraEIPs {
		eips[i] = int(eip)
	}

	vmConfig := vm.Config{
		ExtraEips: eips,
	}

	if st.Debug {
		vmConfig.Tracer = vm.NewJSONLogger(&vm.LogConfig{
			Debug: true,
		}, os.Stderr)

		vmConfig.Debug = true
	}

	return vm.NewEVM(blockCtx, txCtx, csdb, config.EthereumConfig(st.ChainID), vmConfig)
}

// TransitionDb will transition the state by applying the current transaction and
// returning the evm execution result.
// NOTE: State transition checks are run during AnteHandler execution.
func (st *StateTransition) TransitionDb(ctx sdk.Context, config ChainConfig) (resp *ExecutionResult, err error) {
	defer telemetry.ModuleMeasureSince(ModuleName, time.Now(), MetricKeyTransitionDB)

	contractCreation := st.Message.To() == nil

	cost, err := core.IntrinsicGas(st.Message.Data(), st.Message.AccessList(), true, false, true)
	if err != nil {
		err = sdkerrors.Wrap(err, "invalid intrinsic gas for transaction")
		return nil, err
	}

	// This gas limit the the transaction gas limit with intrinsic gas subtracted
	gasLimit := st.Message.Gas() - ctx.GasMeter().GasConsumed()

	csdb := st.Csdb.WithContext(ctx)
	if st.Simulate {
		// gasLimit is set here because stdTxs incur gaskv charges in the ante handler, but for eth_call
		// the cost needs to be the same as an Ethereum transaction sent through the web3 API
		consumedGas := ctx.GasMeter().GasConsumed()
		gasLimit = st.Message.Gas() - cost
		if consumedGas < cost {
			// If Cosmos standard tx ante handler cost is less than EVM intrinsic cost
			// gas must be consumed to match to accurately simulate an Ethereum transaction
			ctx.GasMeter().ConsumeGas(cost-consumedGas, "Intrinsic gas match")
		}

		csdb = st.Csdb.Copy()
	}

	// This gas meter is set up to consume gas from gaskv during evm execution and be ignored
	currentGasMeter := ctx.GasMeter()
	evmGasMeter := sdk.NewInfiniteGasMeter()
	csdb.WithContext(ctx.WithGasMeter(evmGasMeter))

	// Clear cache of accounts to handle changes outside of the EVM
	csdb.UpdateAccounts()

	params := csdb.GetParams()

	gasPrice := ctx.MinGasPrices().AmountOf(params.EvmDenom)
	//gasPrice := sdk.ZeroDec()
	if gasPrice.IsNil() {
		return nil, errors.New("min gas price cannot be nil")
	}

	evm := st.newEVM(ctx, csdb, gasLimit, config, params.ExtraEIPs)

	var (
		ret             []byte
		leftOverGas     uint64
		contractAddress common.Address
		senderRef       = vm.AccountRef(st.Message.From())
	)

	// Get nonce of account outside of the EVM
	currentNonce := csdb.GetNonce(st.Message.From())
	// Set nonce of sender account before evm state transition for usage in generating Create address
	csdb.SetNonce(st.Message.From(), st.Message.Nonce())

	// create contract or execute call
	switch contractCreation {
	case true:
		if !params.EnableCreate {
			return nil, ErrCreateDisabled
		}

		ret, contractAddress, leftOverGas, err = evm.Create(senderRef, st.Message.Data(), gasLimit, st.Message.Value())

		if err != nil {
			log.WithField("simulate", st.Simulate).
				WithField("nonce", st.Message.Nonce()).
				WithField("contract", contractAddress.String()).
				WithError(err).Warningln("evm contract creation failed")
		}

		gasConsumed := gasLimit - leftOverGas
		resp = &ExecutionResult{
			Response: &MsgEthereumTxResponse{
				Ret: ret,
			},
			GasInfo: GasInfo{
				GasConsumed: gasConsumed,
				GasLimit:    gasLimit,
				GasRefunded: leftOverGas,
			},
		}
	default:
		if !params.EnableCall {
			return nil, ErrCallDisabled
		}

		// Increment the nonce for the next transaction	(just for evm state transition)
		csdb.SetNonce(st.Message.From(), csdb.GetNonce(st.Message.From())+1)

		ret, leftOverGas, err = evm.Call(senderRef, *st.Message.To(), st.Message.Data(), gasLimit, st.Message.Value())

		// fmt.Println("EVM CALL!!!", senderRef.Address().Hex(), (*st.Message.To()).Hex(), gasLimit)
		// fmt.Println("EVM CALL RESULT", common.ToHex(ret), leftOverGas, err)

		if err != nil {
			log.WithField("recipient", st.Message.To().String()).
				WithError(err).Debugln("evm call failed")
		}

		gasConsumed := gasLimit - leftOverGas
		resp = &ExecutionResult{
			Response: &MsgEthereumTxResponse{
				Ret: ret,
			},
			GasInfo: GasInfo{
				GasConsumed: gasConsumed,
				GasLimit:    gasLimit,
				GasRefunded: leftOverGas,
			},
		}
	}

	if err != nil {
		// Consume gas before returning
		ctx.GasMeter().ConsumeGas(resp.GasInfo.GasConsumed, "evm execution consumption")
		return resp, err
	}

	// Resets nonce to value pre state transition
	csdb.SetNonce(st.Message.From(), currentNonce)

	// Generate bloom filter to be saved in tx receipt data
	bloomInt := big.NewInt(0)

	var (
		bloomFilter ethtypes.Bloom
		logs        []*ethtypes.Log
	)

	if st.TxHash != nil && !st.Simulate {
		logs, err = csdb.GetLogs(*st.TxHash)
		if err != nil {
			err = errors.Wrap(err, "failed to get logs")
			return nil, err
		}

		bloomInt = big.NewInt(0).SetBytes(ethtypes.LogsBloom(logs))
		bloomFilter = ethtypes.BytesToBloom(bloomInt.Bytes())
	}

	if !st.Simulate {
		// Finalise state if not a simulated transaction
		// TODO: change to depend on config
		if err := csdb.Finalise(true); err != nil {
			return nil, err
		}
	}

	resp.Logs = logs
	resp.Bloom = bloomInt
	resp.Response = &MsgEthereumTxResponse{
		Bloom:  bloomFilter.Bytes(),
		TxLogs: NewTransactionLogsFromEth(*st.TxHash, logs),
		Ret:    ret,
	}

	if contractCreation {
		resp.Response.ContractAddress = contractAddress.String()
	}

	// TODO: Refund unused gas here, if intended in future

	// Consume gas from evm execution
	// Out of gas check does not need to be done here since it is done within the EVM execution

	// TODO: @albert, @maxim, decide if can take this out, since InternalEthereumTx may want to continue execution afterwards
	// which will use gas.
	_ = currentGasMeter
	//ctx.WithGasMeter(currentGasMeter).GasMeter().ConsumeGas(resp.GasInfo.GasConsumed, "EVM execution consumption")

	return resp, nil
}

// StaticCall executes the contract associated with the addr with the given input
// as parameters while disallowing any modifications to the state during the call.
// Opcodes that attempt to perform such modifications will result in exceptions
// instead of performing the modifications.
func (st *StateTransition) StaticCall(ctx sdk.Context, config ChainConfig) ([]byte, error) {
	defer telemetry.ModuleMeasureSince(ModuleName, time.Now(), MetricKeyStaticCall)

	// This gas limit the the transaction gas limit with intrinsic gas subtracted
	gasLimit := st.Message.Gas() - ctx.GasMeter().GasConsumed()
	csdb := st.Csdb.WithContext(ctx)

	// This gas meter is set up to consume gas from gaskv during evm execution and be ignored
	evmGasMeter := sdk.NewInfiniteGasMeter()
	csdb.WithContext(ctx.WithGasMeter(evmGasMeter))

	// Clear cache of accounts to handle changes outside of the EVM
	csdb.UpdateAccounts()

	params := csdb.GetParams()

	gasPrice := ctx.MinGasPrices().AmountOf(params.EvmDenom)
	if gasPrice.IsNil() {
		return []byte{}, errors.New("min gas price cannot be nil")
	}

	evm := st.newEVM(ctx, csdb, gasLimit, config, params.ExtraEIPs)
	senderRef := vm.AccountRef(st.Message.From())

	ret, _, err := evm.StaticCall(senderRef, *st.Message.To(), st.Message.Data(), gasLimit)

	// fmt.Println("EVM STATIC CALL!!!", senderRef.Address().Hex(), (*st.Message.To()).Hex(), st.Message.Data(), gasLimit)
	// fmt.Println("EVM STATIC CALL RESULT", common.ToHex(ret), leftOverGas, err)

	return ret, err
}

// HashFromContext returns the Ethereum Header hash from the context's Tendermint
// block header.
func HashFromContext(ctx sdk.Context) common.Hash {
	// cast the ABCI header to tendermint Header type
	protoHeader := ctx.BlockHeader()
	tmHeader, err := tmtypes.HeaderFromProto(&protoHeader)
	if err != nil {
		return common.Hash{}
	}

	// get the Tendermint block hash from the current header
	tmBlockHash := tmHeader.Hash()

	// NOTE: if the validator set hash is missing the hash will be returned as nil,
	// so we need to check for this case to prevent a panic when calling Bytes()
	if tmBlockHash == nil {
		return common.Hash{}
	}

	return common.BytesToHash(tmBlockHash.Bytes())
}
