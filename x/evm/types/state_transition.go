package types

import (
	"errors"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	emint "github.com/cosmos/ethermint/types"
)

// StateTransition defines data to transitionDB in evm
type StateTransition struct {
	Payload      []byte
	Recipient    *common.Address
	AccountNonce uint64
	GasLimit     uint64
	Price        *big.Int
	Amount       *big.Int
	ChainID      *big.Int
	Csdb         *CommitStateDB
	THash        *common.Hash
	Sender       common.Address
	Simulate     bool
}

// ReturnData represents what's returned from a transition
type ReturnData struct {
	Logs   []*ethtypes.Log
	Bloom  *big.Int
	Result *sdk.Result
}

// TODO: move to keeper
// TransitionCSDB performs an evm state transition from a transaction
// TODO: update godoc, it doesn't explain what it does in depth.
func (st StateTransition) TransitionCSDB(ctx sdk.Context) (*ReturnData, error) {
	contractCreation := st.Recipient == nil

	cost, err := core.IntrinsicGas(st.Payload, contractCreation, true)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "invalid intrinsic gas for transaction")
	}

	// This gas limit the the transaction gas limit with intrinsic gas subtracted
	gasLimit := st.GasLimit - ctx.GasMeter().GasConsumed()

	csdb := st.Csdb.WithContext(ctx)
	if st.Simulate {
		// gasLimit is set here because stdTxs incur gaskv charges in the ante handler, but for eth_call
		// the cost needs to be the same as an Ethereum transaction sent through the web3 API
		consumedGas := ctx.GasMeter().GasConsumed()
		gasLimit = st.GasLimit - cost
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

	gasPrice := ctx.MinGasPrices().AmountOf(emint.DenomDefault)
	if gasPrice.IsNil() {
		return nil, errors.New("gas price cannot be nil")
	}

	// Create context for evm
	context := vm.Context{
		CanTransfer: core.CanTransfer,
		Transfer:    core.Transfer,
		Origin:      st.Sender,
		Coinbase:    common.Address{}, // TODO: explain why this is empty
		BlockNumber: big.NewInt(ctx.BlockHeight()),
		Time:        big.NewInt(ctx.BlockHeader().Time.Unix()),
		Difficulty:  big.NewInt(0), // unused. Only required in PoW context
		GasLimit:    gasLimit,
		GasPrice:    gasPrice.Int,
	}

	evm := vm.NewEVM(context, csdb, GenerateChainConfig(st.ChainID), vm.Config{})

	var (
		ret         []byte
		leftOverGas uint64
		addr        common.Address
		senderRef   = vm.AccountRef(st.Sender)
	)

	// Get nonce of account outside of the EVM
	currentNonce := st.Csdb.GetNonce(st.Sender)
	// Set nonce of sender account before evm state transition for usage in generating Create address
	st.Csdb.SetNonce(st.Sender, st.AccountNonce)

	switch contractCreation {
	case true:
		ret, addr, leftOverGas, err = evm.Create(senderRef, st.Payload, gasLimit, st.Amount)
	default:
		// Increment the nonce for the next transaction	(just for evm state transition)
		csdb.SetNonce(st.Sender, csdb.GetNonce(st.Sender)+1)
		ret, leftOverGas, err = evm.Call(senderRef, *st.Recipient, st.Payload, gasLimit, st.Amount)
	}

	if err != nil {
		return nil, err
	}

	gasConsumed := gasLimit - leftOverGas

	// Resets nonce to value pre state transition
	st.Csdb.SetNonce(st.Sender, currentNonce)

	// Generate bloom filter to be saved in tx receipt data
	bloomInt := big.NewInt(0)
	var bloomFilter ethtypes.Bloom
	var logs []*ethtypes.Log

	if st.THash != nil && !st.Simulate {
		logs, err = csdb.GetLogs(*st.THash)
		if err != nil {
			return nil, err
		}

		bloomInt = ethtypes.LogsBloom(logs)
		bloomFilter = ethtypes.BytesToBloom(bloomInt.Bytes())
	}

	// Encode all necessary data into slice of bytes to return in sdk result
	res := &ResultData{
		Address: addr,
		Bloom:   bloomFilter,
		Logs:    logs,
		Ret:     ret,
		TxHash:  *st.THash,
	}

	resultData, err := EncodeResultData(res)
	if err != nil {
		return nil, err
	}

	// handle errors
	if err != nil {
		if err == vm.ErrOutOfGas || err == vm.ErrCodeStoreOutOfGas {
			return nil, sdkerrors.Wrap(err, "evm execution went out of gas")
		}

		// Consume gas before returning
		ctx.GasMeter().ConsumeGas(gasConsumed, "EVM execution consumption")
		return nil, err
	}

	// TODO: Refund unused gas here, if intended in future

	if !st.Simulate {
		// Finalise state if not a simulated transaction
		// TODO: change to depend on config
		if err := st.Csdb.Finalise(true); err != nil {
			return nil, err
		}
	}

	// Consume gas from evm execution
	// Out of gas check does not need to be done here since it is done within the EVM execution
	ctx.WithGasMeter(currentGasMeter).GasMeter().ConsumeGas(gasConsumed, "EVM execution consumption")

	err = st.Csdb.SetLogs(*st.THash, logs)
	if err != nil {
		return nil, err
	}

	returnData := &ReturnData{
		Logs:   logs,
		Bloom:  bloomInt,
		Result: &sdk.Result{Data: resultData},
	}

	return returnData, nil
}
