package evm

import (
	"github.com/ethereum/go-ethereum/common"

	sdk "github.com/cosmos/cosmos-sdk/types"

	emint "github.com/cosmos/ethermint/types"
	"github.com/cosmos/ethermint/x/evm/types"
	abci "github.com/tendermint/tendermint/abci/types"
)

// InitGenesis initializes genesis state based on exported genesis
func InitGenesis(ctx sdk.Context, k Keeper, data GenesisState) []abci.ValidatorUpdate {
	for _, account := range data.Accounts {
		// FIXME: this will override bank InitGenesis balance!
		k.SetBalance(ctx, account.Address, account.Balance)
		k.SetCode(ctx, account.Address, account.Code)
		for _, storage := range account.Storage {
			k.SetState(ctx, account.Address, storage.Key, storage.Value)
		}
	}

	var err error
	for _, txLog := range data.TxsLogs {
		err = k.SetLogs(ctx, txLog.Hash, txLog.Logs)
		if err != nil {
			panic(err)
		}
	}

	// set state objects and code to store
	_, err = k.Commit(ctx, false)
	if err != nil {
		panic(err)
	}

	// set storage to store
	err = k.Finalise(ctx, true)
	if err != nil {
		panic(err)
	}
	return []abci.ValidatorUpdate{}
}

// ExportGenesis exports genesis state
func ExportGenesis(ctx sdk.Context, k Keeper, ak types.AccountKeeper) GenesisState {
	// nolint: prealloc
	var ethGenAccounts []types.GenesisAccount
	accounts := ak.GetAllAccounts(ctx)

	var err error
	for _, account := range accounts {
		ethAccount, ok := account.(*emint.EthAccount)
		if !ok {
			continue
		}

		addr := common.BytesToAddress(ethAccount.GetAddress().Bytes())

		var storage []types.GenesisStorage
		err = k.CommitStateDB.ForEachStorage(addr, func(key, value common.Hash) bool {
			storage = append(storage, types.NewGenesisStorage(key, value))
			return false
		})
		if err != nil {
			panic(err)
		}

		genAccount := types.GenesisAccount{
			Address: addr,
			Balance: k.GetBalance(ctx, addr),
			Code:    k.GetCode(ctx, addr),
			Storage: storage,
		}

		ethGenAccounts = append(ethGenAccounts, genAccount)
	}

	return GenesisState{
		Accounts: ethGenAccounts,
		TxsLogs:  k.GetAllTxLogs(ctx),
	}
}
