package evm

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	ethcmn "github.com/ethereum/go-ethereum/common"

	ethermint "github.com/cosmos/ethermint/types"
	"github.com/cosmos/ethermint/x/evm/types"

	abci "github.com/tendermint/tendermint/abci/types"
)

// InitGenesis initializes genesis state based on exported genesis
func InitGenesis(ctx sdk.Context, k Keeper, accountKeeper types.AccountKeeper, data GenesisState) []abci.ValidatorUpdate { // nolint: interfacer
	evmDenom := data.Params.EvmDenom

	for _, account := range data.Accounts {
		address := ethcmn.HexToAddress(account.Address)
		accAddress := sdk.AccAddress(address.Bytes())

		// check that the EVM balance the matches the account balance
		acc := accountKeeper.GetAccount(ctx, accAddress)
		if acc == nil {
			panic(fmt.Errorf("account not found for address %s", account.Address))
		}

		_, ok := acc.(*ethermint.EthAccount)
		if !ok {
			panic(
				fmt.Errorf("account %s must be an %T type, got %T",
					account.Address, &ethermint.EthAccount{}, acc,
				),
			)
		}

		evmBalance := acc.GetCoins().AmountOf(evmDenom)
		if !evmBalance.Equal(account.Balance) {
			panic(
				fmt.Errorf(
					"balance mismatch for account %s, expected %s%s, got %s%s",
					account.Address, evmBalance, evmDenom, account.Balance, evmDenom,
				),
			)
		}

		k.SetBalance(ctx, address, account.Balance.BigInt())
		k.SetCode(ctx, address, account.Code)
		for _, storage := range account.Storage {
			k.SetState(ctx, address, storage.Key, storage.Value)
		}
	}

	var err error
	for _, txLog := range data.TxsLogs {
		if err = k.SetLogs(ctx, txLog.Hash, txLog.Logs); err != nil {
			panic(err)
		}
	}

	k.SetChainConfig(ctx, data.ChainConfig)
	k.SetParams(ctx, data.Params)

	// set state objects and code to store
	_, err = k.Commit(ctx, false)
	if err != nil {
		panic(err)
	}

	// set storage to store
	// NOTE: don't delete empty object to prevent import-export simulation failure
	err = k.Finalise(ctx, false)
	if err != nil {
		panic(err)
	}

	return []abci.ValidatorUpdate{}
}

// ExportGenesis exports genesis state of the EVM module
func ExportGenesis(ctx sdk.Context, k Keeper, ak types.AccountKeeper) GenesisState {
	// nolint: prealloc
	var ethGenAccounts []types.GenesisAccount
	accounts := ak.GetAllAccounts(ctx)

	for _, account := range accounts {
		ethAccount, ok := account.(*ethermint.EthAccount)
		if !ok {
			continue
		}

		addr := ethAccount.EthAddress()

		storage, err := k.GetAccountStorage(ctx, addr)
		if err != nil {
			panic(err)
		}

		balanceInt := k.GetBalance(ctx, addr)
		balance := sdk.NewIntFromBigInt(balanceInt)

		genAccount := types.GenesisAccount{
			Address: addr.String(),
			Balance: balance,
			Code:    k.GetCode(ctx, addr),
			Storage: storage,
		}

		ethGenAccounts = append(ethGenAccounts, genAccount)
	}

	config, _ := k.GetChainConfig(ctx)

	return GenesisState{
		Accounts:    ethGenAccounts,
		TxsLogs:     k.GetAllTxLogs(ctx),
		ChainConfig: config,
		Params:      k.GetParams(ctx),
	}
}
