package precompiles

import (
	"bytes"
	"errors"
	"math/big"
	"sort"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/evmos/ethermint/x/evm/statedb"
	"github.com/evmos/ethermint/x/evm/types"
)

const EVMDenomPrefix = "evm/"

var (
	MintMethod      abi.Method
	BalanceOfMethod abi.Method

	_ statedb.StatefulPrecompiledContract = (*BankContract)(nil)
	_ statedb.JournalEntry                = bankMintChange{}
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
}

func EVMDenom(token common.Address) string {
	return EVMDenomPrefix + token.Hex()
}

type Balance struct {
	OriginAmount *big.Int
	DirtyAmount  *big.Int
}

func (b Balance) Changed() *big.Int {
	return new(big.Int).Sub(b.DirtyAmount, b.OriginAmount)
}

type BankContract struct {
	ctx        sdk.Context
	bankKeeper types.BankKeeper
	balances   map[common.Address]map[common.Address]*Balance
}

// NewBankContractCreator creates the precompiled contract to manage native tokens
func NewBankContractCreator(bankKeeper types.BankKeeper) statedb.PrecompiledContractCreator {
	return func(ctx sdk.Context) statedb.StatefulPrecompiledContract {
		return &BankContract{
			ctx:        ctx,
			bankKeeper: bankKeeper,
			balances:   make(map[common.Address]map[common.Address]*Balance),
		}
	}
}

// RequiredGas calculates the contract gas use
func (bc *BankContract) RequiredGas(input []byte) uint64 {
	// TODO estimate required gas
	return 0
}

func (bc *BankContract) Run(evm *vm.EVM, input []byte, caller common.Address, value *big.Int, readonly bool) ([]byte, error) {
	stateDB, ok := evm.StateDB.(ExtStateDB)
	if !ok {
		return nil, errors.New("not run in ethermint")
	}

	// parse input
	methodID := input[:4]
	if bytes.Equal(methodID, MintMethod.ID) {
		if readonly {
			return nil, errors.New("the method is not readonly")
		}
		args, err := MintMethod.Inputs.Unpack(input[4:])
		if err != nil {
			return nil, errors.New("fail to unpack input arguments")
		}
		recipient := args[0].(common.Address)
		amount := args[1].(*big.Int)
		if amount.Sign() <= 0 {
			return nil, errors.New("invalid amount")
		}

		if _, ok := bc.balances[caller]; !ok {
			bc.balances[caller] = make(map[common.Address]*Balance)
		}
		balances := bc.balances[caller]
		if balance, ok := balances[recipient]; ok {
			balance.DirtyAmount = new(big.Int).Add(balance.DirtyAmount, amount)
		} else {
			// query original amount
			addr := sdk.AccAddress(recipient.Bytes())
			originAmount := bc.bankKeeper.GetBalance(bc.ctx, addr, EVMDenom(caller)).Amount.BigInt()
			dirtyAmount := new(big.Int).Add(originAmount, amount)
			balances[recipient] = &Balance{
				OriginAmount: originAmount,
				DirtyAmount:  dirtyAmount,
			}
		}
		stateDB.AppendJournalEntry(bankMintChange{bc: bc, caller: caller, recipient: recipient, amount: amount})
	} else if bytes.Equal(methodID, BalanceOfMethod.ID) {
		args, err := BalanceOfMethod.Inputs.Unpack(input[4:])
		if err != nil {
			return nil, errors.New("fail to unpack input arguments")
		}
		token := args[0].(common.Address)
		addr := args[1].(common.Address)
		if balances, ok := bc.balances[token]; ok {
			if balance, ok := balances[addr]; ok {
				return BalanceOfMethod.Outputs.Pack(balance.DirtyAmount)
			}
		}
		// query from storage
		amount := bc.bankKeeper.GetBalance(bc.ctx, sdk.AccAddress(addr.Bytes()), EVMDenom(token)).Amount.BigInt()
		return BalanceOfMethod.Outputs.Pack(amount)
	} else {
		return nil, errors.New("unknown method")
	}
	return nil, nil
}

func (bc *BankContract) Commit(ctx sdk.Context) error {
	// sorted iteration
	sortedContracts := make([]common.Address, len(bc.balances))
	i := 0
	for contract := range bc.balances {
		sortedContracts[i] = contract
		i++
	}
	sort.Slice(sortedContracts, func(i, j int) bool {
		return bytes.Compare(sortedContracts[i].Bytes(), sortedContracts[j].Bytes()) < 0
	})
	for _, contract := range sortedContracts {
		denom := EVMDenom(contract)
		balances := bc.balances[contract]
		sortedRecipients := make([]common.Address, len(balances))
		i = 0
		for recipient := range balances {
			sortedRecipients[i] = recipient
			i++
		}
		sort.Slice(sortedRecipients, func(i, j int) bool {
			return bytes.Compare(sortedRecipients[i].Bytes(), sortedRecipients[j].Bytes()) < 0
		})
		for _, recipient := range sortedRecipients {
			cosmosAddr := sdk.AccAddress(recipient.Bytes())
			changed := balances[recipient].Changed()
			switch changed.Sign() {
			case 1:
				amt := sdk.NewCoins(sdk.NewCoin(denom, sdk.NewIntFromBigInt(changed)))
				if err := bc.bankKeeper.MintCoins(ctx, types.ModuleName, amt); err != nil {
					return sdkerrors.Wrap(err, "fail to mint coins in precompiled contract")
				}
				if err := bc.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, cosmosAddr, amt); err != nil {
					return sdkerrors.Wrap(err, "fail to send mint coins to account")
				}
			case -1:
				amt := sdk.NewCoins(sdk.NewCoin(denom, sdk.NewIntFromBigInt(new(big.Int).Neg(changed))))
				if err := bc.bankKeeper.SendCoinsFromAccountToModule(ctx, cosmosAddr, types.ModuleName, amt); err != nil {
					return sdkerrors.Wrap(err, "fail to send burn coins to module")
				}
				if err := bc.bankKeeper.BurnCoins(ctx, types.ModuleName, amt); err != nil {
					return sdkerrors.Wrap(err, "fail to burn coins in precompiled contract")
				}

			}
		}
	}
	return nil
}

type bankMintChange struct {
	bc        *BankContract
	caller    common.Address
	recipient common.Address
	amount    *big.Int
}

func (ch bankMintChange) Revert(*statedb.StateDB) {
	balance := ch.bc.balances[ch.caller][ch.recipient]
	balance.DirtyAmount = new(big.Int).Sub(balance.DirtyAmount, ch.amount)
}

func (ch bankMintChange) Dirtied() *common.Address {
	return nil
}
