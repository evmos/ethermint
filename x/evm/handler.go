package evm

import (
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/vm"

	sdk "github.com/cosmos/cosmos-sdk/types"
	emint "github.com/cosmos/ethermint/types"
	"github.com/cosmos/ethermint/x/evm/types"
)

// NewHandler returns a handler for Ethermint type messages.
func NewHandler(keeper Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		switch msg := msg.(type) {
		case types.EthereumTxMsg:
			return handleETHTxMsg(ctx, keeper, msg)
		default:
			errMsg := fmt.Sprintf("Unrecognized ethermint Msg type: %v", msg.Type())
			return sdk.ErrUnknownRequest(errMsg).Result()
		}
	}
}

// Handle an Ethereum specific tx
func handleETHTxMsg(ctx sdk.Context, keeper Keeper, msg types.EthereumTxMsg) sdk.Result {
	if err := msg.ValidateBasic(); err != nil {
		return err.Result()
	}

	// parse the chainID from a string to a base-10 integer
	intChainID, ok := new(big.Int).SetString(ctx.ChainID(), 10)
	if !ok {
		return emint.ErrInvalidChainID(fmt.Sprintf("invalid chainID: %s", ctx.ChainID())).Result()
	}

	// Verify signature and retrieve sender address
	sender, err := msg.VerifySig(intChainID)
	if err != nil {
		return emint.ErrInvalidSender(err.Error()).Result()
	}
	contractCreation := msg.To() == nil

	// Pay intrinsic gas
	// TODO: Check config for homestead enabled
	cost, err := core.IntrinsicGas(msg.Data.Payload, contractCreation, true)
	if err != nil {
		return emint.ErrInvalidIntrinsicGas(err.Error()).Result()
	}

	usableGas := msg.Data.GasLimit - cost

	// Create context for evm
	context := vm.Context{
		CanTransfer: core.CanTransfer,
		Transfer:    core.Transfer,
		Origin:      sender,
		Coinbase:    common.Address{},
		BlockNumber: big.NewInt(ctx.BlockHeight()),
		Time:        big.NewInt(time.Now().Unix()),
		Difficulty:  big.NewInt(0x30000), // unused
		GasLimit:    ctx.GasMeter().Limit(),
		GasPrice:    ctx.MinGasPrices().AmountOf(emint.DenomDefault).Int,
	}

	vmenv := vm.NewEVM(context, keeper.csdb.WithContext(ctx), types.GenerateChainConfig(intChainID), vm.Config{})

	var (
		leftOverGas uint64
		addr        common.Address
		vmerr       error
		senderRef   = vm.AccountRef(sender)
	)

	if contractCreation {
		_, addr, leftOverGas, vmerr = vmenv.Create(senderRef, msg.Data.Payload, usableGas, msg.Data.Amount)
	} else {
		// Increment the nonce for the next transaction
		keeper.csdb.SetNonce(sender, keeper.csdb.GetNonce(sender)+1)
		_, leftOverGas, vmerr = vmenv.Call(senderRef, *msg.To(), msg.Data.Payload, usableGas, msg.Data.Amount)
	}

	// handle errors
	if vmerr != nil {
		return emint.ErrVMExecution(vmerr.Error()).Result()
	}

	// Refund remaining gas from tx (Check these values and ensure gas is being consumed correctly)
	refundGas(keeper.csdb, &leftOverGas, msg.Data.GasLimit, context.GasPrice, sender)

	// add balance for the processor of the tx (determine who rewards are being processed to)
	// TODO: Double check nothing needs to be done here

	keeper.csdb.Finalise(true) // Change to depend on config

	// TODO: Remove commit from tx handler (should be done at end of block)
	_, err = keeper.csdb.Commit(true)
	if err != nil {
		return sdk.ErrUnknownRequest("Failed to write data to kv store").Result()
	}

	// TODO: Consume gas from sender

	return sdk.Result{Data: addr.Bytes(), GasUsed: msg.Data.GasLimit - leftOverGas}
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

	// Return ETH for remaining gas, exchanged at the original rate.
	remaining := new(big.Int).Mul(new(big.Int).SetUint64(*gasRemaining), gasPrice)
	st.AddBalance(from, remaining)

	// // Also return remaining gas to the block gas counter so it is
	// // available for the next transaction.
	// TODO: Return gas to block gas meter?
	// st.gp.AddGas(st.gas)
}
