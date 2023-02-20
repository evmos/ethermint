package precompiles

import (
	"errors"
	"math/big"

	sdkmath "cosmossdk.io/math"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/evmos/ethermint/x/evm/types"
)

const EVMDenomPrefix = "evm/"

var (
	MintMethod      abi.Method
	BurnMethod      abi.Method
	BalanceOfMethod abi.Method
	TransferMethod  abi.Method
)

func init() {
	addressType, _ := abi.NewType("address", "", nil)
	uint256Type, _ := abi.NewType("uint256", "", nil)
	MintMethod = abi.NewMethod(
		"mint", "mint", abi.Function, "", false, false, abi.Arguments{abi.Argument{
			Name: "recipient",
			Type: addressType,
		}, abi.Argument{
			Name: "amount",
			Type: uint256Type,
		}},
		nil,
	)
	BurnMethod = abi.NewMethod(
		"burn", "burn", abi.Function, "", false, false, abi.Arguments{abi.Argument{
			Name: "recipient",
			Type: addressType,
		}, abi.Argument{
			Name: "amount",
			Type: uint256Type,
		}},
		nil,
	)
	BalanceOfMethod = abi.NewMethod(
		"balanceOf", "balanceOf", abi.Function, "", false, false, abi.Arguments{abi.Argument{
			Name: "token",
			Type: addressType,
		}, abi.Argument{
			Name: "address",
			Type: addressType,
		}},
		abi.Arguments{abi.Argument{
			Name: "amount",
			Type: uint256Type,
		}},
	)
	TransferMethod = abi.NewMethod(
		"transfer", "transfer", abi.Function, "", false, false, abi.Arguments{abi.Argument{
			Name: "sender",
			Type: addressType,
		}, abi.Argument{
			Name: "recipient",
			Type: addressType,
		}, abi.Argument{
			Name: "amount",
			Type: uint256Type,
		}},
		nil,
	)
}

func EVMDenom(token common.Address) string {
	return EVMDenomPrefix + token.Hex()
}

type BankContract struct {
	ctx        sdk.Context
	bankKeeper types.BankKeeper
	stateDB    ExtStateDB
}

// NewBankContract creates the precompiled contract to manage native tokens
func NewBankContract(ctx sdk.Context, bankKeeper types.BankKeeper, stateDB ExtStateDB) StatefulPrecompiledContract {
	return &BankContract{ctx, bankKeeper, stateDB}
}

func (bc *BankContract) Address() common.Address {
	return common.BytesToAddress([]byte{100})
}

// RequiredGas calculates the contract gas use
func (bc *BankContract) RequiredGas(input []byte) uint64 {
	// TODO estimate required gas
	return 0
}

func (bc *BankContract) IsStateful() bool {
	return true
}

func (bc *BankContract) Run(evm *vm.EVM, contract *vm.Contract, readonly bool) ([]byte, error) {
	// parse input
	methodID := contract.Input[:4]
	switch string(methodID) {
	case string(MintMethod.ID), string(BurnMethod.ID):
		var method abi.Method
		var mintDenom, burnDenom string
		if string(methodID) == string(MintMethod.ID) {
			method = MintMethod
			mintDenom = EVMDenom(contract.CallerAddress)
			burnDenom = types.DefaultEVMDenom
		} else {
			method = BurnMethod
			burnDenom = EVMDenom(contract.CallerAddress)
			mintDenom = types.DefaultEVMDenom
		}
		if readonly {
			return nil, errors.New("the method is not readonly")
		}
		args, err := method.Inputs.Unpack(contract.Input[4:])
		if err != nil {
			return nil, errors.New("fail to unpack input arguments")
		}
		recipient := args[0].(common.Address)
		amount := args[1].(*big.Int)
		if amount.Sign() <= 0 {
			return nil, errors.New("invalid amount")
		}
		err = bc.stateDB.ExecuteNativeAction(func(ctx sdk.Context) error {
			addr := sdk.AccAddress(recipient.Bytes())
			nativeAmt := sdk.NewCoins(sdk.NewCoin(burnDenom, sdkmath.NewIntFromBigInt(amount)))
			if err := bc.bankKeeper.SendCoinsFromAccountToModule(ctx, addr, types.ModuleName, nativeAmt); err != nil {
				return errorsmod.Wrap(err, "fail to send burn coins to module")
			}
			if err := bc.bankKeeper.BurnCoins(ctx, types.ModuleName, nativeAmt); err != nil {
				return errorsmod.Wrap(err, "fail to burn coins in precompiled contract")
			}
			amt := sdk.NewCoins(sdk.NewCoin(mintDenom, sdkmath.NewIntFromBigInt(amount)))
			if err := bc.bankKeeper.MintCoins(ctx, types.ModuleName, amt); err != nil {
				return errorsmod.Wrap(err, "fail to mint coins in precompiled contract")
			}
			if err := bc.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, addr, amt); err != nil {
				return errorsmod.Wrap(err, "fail to send mint coins to account")
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
	case string(BalanceOfMethod.ID):
		args, err := BalanceOfMethod.Inputs.Unpack(contract.Input[4:])
		if err != nil {
			return nil, errors.New("fail to unpack input arguments")
		}
		token := args[0].(common.Address)
		addr := args[1].(common.Address)
		// query from storage
		balance := bc.bankKeeper.GetBalance(bc.ctx, sdk.AccAddress(addr.Bytes()), EVMDenom(token)).Amount.BigInt()
		return BalanceOfMethod.Outputs.Pack(balance)
	case string(TransferMethod.ID):
		if readonly {
			return nil, errors.New("the method is not readonly")
		}
		args, err := TransferMethod.Inputs.Unpack(contract.Input[4:])
		if err != nil {
			return nil, errors.New("fail to unpack input arguments")
		}
		sender := args[0].(common.Address)
		recipient := args[1].(common.Address)
		amount := args[2].(*big.Int)
		if amount.Sign() <= 0 {
			return nil, errors.New("invalid amount")
		}
		err = bc.stateDB.ExecuteNativeAction(func(ctx sdk.Context) error {
			from := sdk.AccAddress(sender.Bytes())
			to := sdk.AccAddress(recipient.Bytes())
			nativeAmt := sdk.NewCoins(sdk.NewCoin(types.DefaultEVMDenom, sdkmath.NewIntFromBigInt(amount)))
			if err := bc.bankKeeper.SendCoins(ctx, from, to, nativeAmt); err != nil {
				return errorsmod.Wrap(err, "fail to send coins from sender to recipient")
			}
			denom := EVMDenom(contract.CallerAddress)
			amt := sdk.NewCoins(sdk.NewCoin(denom, sdkmath.NewIntFromBigInt(amount)))
			if err := bc.bankKeeper.SendCoins(ctx, from, to, amt); err != nil {
				return errorsmod.Wrap(err, "fail to send coins in precompiled contract")
			}
			return nil
		})
		return nil, err
	default:
		return nil, errors.New("unknown method")
	}
	return nil, nil
}
