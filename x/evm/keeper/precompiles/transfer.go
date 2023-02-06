package precompiles

import (
	"errors"
	"math/big"

	sdkmath "cosmossdk.io/math"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/evmos/ethermint/x/evm/types"
	evm "github.com/evmos/ethermint/x/evm/vm"
)

var (
	NativeTransferMethod   abi.Method
	TransferTransferMethod abi.Method

	_ evm.StatefulPrecompiledContract = (*TransferContract)(nil)
)

func init() {
	addressType, _ := abi.NewType("address", "", nil)
	uint256Type, _ := abi.NewType("uint256", "", nil)
	NativeTransferMethod = abi.NewMethod(
		"nativeTransfer", "nativeTransfer", abi.Function, "", false, false, abi.Arguments{abi.Argument{
			Name: "receiver",
			Type: addressType,
		}, abi.Argument{
			Name: "amount",
			Type: uint256Type,
		}},
		nil,
	)
}

type TransferContract struct {
	ctx        sdk.Context
	bankKeeper types.BankKeeper
	stateDB    evm.ExtStateDB
}

// NewTransferContractCreator creates the precompiled contract to manage native tokens
func NewTransferContractCreator(bankKeeper types.BankKeeper) evm.PrecompiledContractCreator {
	return func(
		ctx sdk.Context,
		stateDB evm.ExtStateDB,
	) evm.StatefulPrecompiledContract {
		return &TransferContract{
			ctx:        ctx,
			bankKeeper: bankKeeper,
			stateDB:    stateDB,
		}
	}
}

func (tc *TransferContract) Address() common.Address {
	return common.BytesToAddress([]byte{102})
}

// RequiredGas calculates the contract gas use
func (tc *TransferContract) RequiredGas(input []byte) uint64 {
	// TODO estimate required gas
	return 0
}

func (tc *TransferContract) IsStateful() bool {
	return true
}

func (tc *TransferContract) Run(evm *vm.EVM, contract *vm.Contract, readonly bool) ([]byte, error) {
	// parse input
	methodID := contract.Input[:4]
	switch string(methodID) {
	case string(NativeTransferMethod.ID):
		if readonly {
			return nil, errors.New("the method is not readonly")
		}
		args, err := NativeTransferMethod.Inputs.Unpack(contract.Input[4:])
		if err != nil {
			return nil, errors.New("fail to unpack input arguments")
		}
		recipient := args[0].(common.Address)
		amount := args[1].(*big.Int)
		if amount.Sign() <= 0 {
			return nil, errors.New("invalid amount")
		}
		err = tc.stateDB.ExecuteNativeAction(func(ctx sdk.Context) error {
			from := sdk.AccAddress(contract.CallerAddress.Bytes())
			to := sdk.AccAddress(recipient.Bytes())
			amt := sdk.NewCoins(sdk.NewCoin(types.DefaultEVMDenom, sdkmath.NewIntFromBigInt(amount)))
			if err := tc.bankKeeper.SendCoins(ctx, from, to, amt); err != nil {
				return sdkerrors.Wrap(err, "fail to send coins in precompiled contract")
			}
			return nil
		})
		return nil, err
	default:
		return nil, errors.New("unknown method")
	}
}
