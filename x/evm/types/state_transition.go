package types

import (
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
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
}

// TransitionCSDB performs an evm state transition from a transaction
func (st StateTransition) TransitionCSDB(ctx sdk.Context) sdk.Result {
	contractCreation := st.Recipient == nil

	// Create context for evm
	context := vm.Context{
		CanTransfer: core.CanTransfer,
		Transfer:    core.Transfer,
		Origin:      st.Sender,
		Coinbase:    common.Address{},
		BlockNumber: big.NewInt(ctx.BlockHeight()),
		Time:        big.NewInt(time.Now().Unix()),
		Difficulty:  big.NewInt(0x30000), // unused
		GasLimit:    ctx.GasMeter().Limit(),
		GasPrice:    ctx.MinGasPrices().AmountOf(emint.DenomDefault).Int,
	}

	vmenv := vm.NewEVM(context, st.Csdb.WithContext(ctx), GenerateChainConfig(st.ChainID), vm.Config{})

	var (
		leftOverGas uint64
		addr        common.Address
		vmerr       error
		senderRef   = vm.AccountRef(st.Sender)
	)

	if contractCreation {
		_, addr, leftOverGas, vmerr = vmenv.Create(senderRef, st.Payload, st.GasLimit, st.Amount)
	} else {
		// Increment the nonce for the next transaction
		st.Csdb.SetNonce(st.Sender, st.Csdb.GetNonce(st.Sender)+1)
		_, leftOverGas, vmerr = vmenv.Call(senderRef, *st.Recipient, st.Payload, st.GasLimit, st.Amount)
	}

	// handle errors
	if vmerr != nil {
		return emint.ErrVMExecution(vmerr.Error()).Result()
	}

	// Refund remaining gas from tx (Check these values and ensure gas is being consumed correctly)
	refundGas(st.Csdb, &leftOverGas, st.GasLimit, context.GasPrice, st.Sender)

	// add balance for the processor of the tx (determine who rewards are being processed to)
	// TODO: Double check nothing needs to be done here

	st.Csdb.Finalise(true) // Change to depend on config

	// TODO: Consume gas from sender

	return sdk.Result{Data: addr.Bytes(), GasUsed: st.GasLimit - leftOverGas}
}

func refundGas(
	st vm.StateDB, gasRemaining *uint64, initialGas uint64, gasPrice *big.Int,
	from common.Address,
) {
	// Apply refund counter, capped to half of the used gas.
	refund := (initialGas - *gasRemaining) / 2
	if refund > st.GetRefund() {
		refund = st.GetRefund()
	}
	*gasRemaining += refund

	// // Return ETH for remaining gas, exchanged at the original rate.
	// remaining := new(big.Int).Mul(new(big.Int).SetUint64(*gasRemaining), gasPrice)
	// st.AddBalance(from, remaining)

	// // Also return remaining gas to the block gas counter so it is
	// // available for the next transaction.
	// TODO: Return gas to block gas meter?
	// st.gp.AddGas(st.gas)
}
