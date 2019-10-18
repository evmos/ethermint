package types

import (
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"

	sdk "github.com/cosmos/cosmos-sdk/types"
	emint "github.com/cosmos/ethermint/types"
)

// StateTransition defines data to transitionDB in evm
type StateTransition struct {
	Sender       common.Address
	AccountNonce uint64
	Price        *big.Int
	GasLimit     uint64
	Recipient    *common.Address
	Amount       *big.Int
	Payload      []byte
	Csdb         *CommitStateDB
	ChainID      *big.Int
	THash        *common.Hash
}

// TransitionCSDB performs an evm state transition from a transaction
func (st StateTransition) TransitionCSDB(ctx sdk.Context) (sdk.Result, *big.Int) {
	contractCreation := st.Recipient == nil

	// This gas limit the the transaction gas limit with intrinsic gas subtracted
	gasLimit := ctx.GasMeter().Limit()

	// Create context for evm
	context := vm.Context{
		CanTransfer: core.CanTransfer,
		Transfer:    core.Transfer,
		Origin:      st.Sender,
		Coinbase:    common.Address{},
		BlockNumber: big.NewInt(ctx.BlockHeight()),
		Time:        big.NewInt(time.Now().Unix()),
		Difficulty:  big.NewInt(0x30000), // unused
		GasLimit:    gasLimit,
		GasPrice:    ctx.MinGasPrices().AmountOf(emint.DenomDefault).Int,
	}

	// This gas meter is set up to consume gas from gaskv during evm execution and be ignored
	evmGasMeter := sdk.NewInfiniteGasMeter()

	vmenv := vm.NewEVM(
		context, st.Csdb.WithContext(ctx.WithGasMeter(evmGasMeter)),
		GenerateChainConfig(st.ChainID), vm.Config{},
	)

	var (
		leftOverGas uint64
		addr        common.Address
		vmerr       error
		senderRef   = vm.AccountRef(st.Sender)
	)

	if contractCreation {
		_, addr, leftOverGas, vmerr = vmenv.Create(senderRef, st.Payload, gasLimit, st.Amount)
	} else {
		// Increment the nonce for the next transaction
		st.Csdb.SetNonce(st.Sender, st.Csdb.GetNonce(st.Sender)+1)
		_, leftOverGas, vmerr = vmenv.Call(senderRef, *st.Recipient, st.Payload, gasLimit, st.Amount)
	}

	// handle errors
	if vmerr != nil {
		return emint.ErrVMExecution(vmerr.Error()).Result(), nil
	}

	// Refunds would happen here, if intended in future

	st.Csdb.Finalise(true) // Change to depend on config

	// Consume gas from evm execution
	// Out of gas check does not need to be done here since it is done within the EVM execution
	ctx.GasMeter().ConsumeGas(gasLimit-leftOverGas, "EVM execution consumption")

	// Generate bloom filter to be saved in tx receipt data
	bloomInt := big.NewInt(0)
	var bloomFilter ethtypes.Bloom
	if st.THash != nil {
		logs := st.Csdb.GetLogs(*st.THash)
		bloomInt = ethtypes.LogsBloom(logs)
		bloomFilter = ethtypes.BytesToBloom(bloomInt.Bytes())
	}

	// TODO: coniditionally add either/ both of these to return data
	returnData := append(addr.Bytes(), bloomFilter.Bytes()...)

	return sdk.Result{Data: returnData, GasUsed: st.GasLimit - leftOverGas}, bloomInt
}
