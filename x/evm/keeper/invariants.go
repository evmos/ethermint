package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	ethermint "github.com/cosmos/ethermint/types"
	"github.com/cosmos/ethermint/x/evm/types"
)

const (
	balanceInvariant = "balance"
	nonceInvariant   = "nonce"
)

// RegisterInvariants registers the evm module invariants
func RegisterInvariants(ir sdk.InvariantRegistry, k Keeper) {
	ir.RegisterRoute(types.ModuleName, balanceInvariant, k.BalanceInvariant())
	ir.RegisterRoute(types.ModuleName, nonceInvariant, k.NonceInvariant())
}

// BalanceInvariant checks that all auth module's EthAccounts in the application have the same balance
// as the EVM one.
func (k Keeper) BalanceInvariant() sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		var (
			msg   string
			count int
		)

		k.CommitStateDB.WithContext(ctx)

		k.accountKeeper.IterateAccounts(ctx, func(account authtypes.AccountI) bool {
			ethAccount, ok := account.(*ethermint.EthAccount)
			if !ok {
				// ignore non EthAccounts
				return false
			}

			evmDenom := k.GetParams(ctx).EvmDenom

			accountBalance := k.bankKeeper.GetBalance(ctx, ethAccount.GetAddress(), evmDenom)
			evmBalance := k.CommitStateDB.GetBalance(ethAccount.EthAddress())

			if evmBalance.Cmp(accountBalance.Amount.BigInt()) != 0 {
				count++
				msg += fmt.Sprintf(
					"\tbalance mismatch for address %s: account balance %s, evm balance %s\n",
					account.GetAddress(), accountBalance.String(), evmBalance.String(),
				)
			}

			return false
		})

		broken := count != 0

		return sdk.FormatInvariant(
			types.ModuleName, balanceInvariant,
			fmt.Sprintf("account balances mismatches found %d\n%s", count, msg),
		), broken
	}
}

// NonceInvariant checks that all auth module's EthAccounts in the application have the same nonce
// sequence as the EVM.
func (k Keeper) NonceInvariant() sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		var (
			msg   string
			count int
		)

		k.CommitStateDB.WithContext(ctx)

		k.accountKeeper.IterateAccounts(ctx, func(account authtypes.AccountI) bool {
			ethAccount, ok := account.(*ethermint.EthAccount)
			if !ok {
				// ignore non EthAccounts
				return false
			}

			evmNonce := k.CommitStateDB.GetNonce(ethAccount.EthAddress())

			if evmNonce != ethAccount.Sequence {
				count++
				msg += fmt.Sprintf(
					"\nonce mismatch for address %s: account nonce %d, evm nonce %d\n",
					account.GetAddress(), ethAccount.Sequence, evmNonce,
				)
			}

			return false
		})

		broken := count != 0

		return sdk.FormatInvariant(
			types.ModuleName, nonceInvariant,
			fmt.Sprintf("account nonces mismatches found %d\n%s", count, msg),
		), broken
	}
}
