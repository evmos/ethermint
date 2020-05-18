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
		csdb := k.CommitStateDB.WithContext(ctx)
		csdb.SetBalance(account.Address, account.Balance)
		csdb.SetCode(account.Address, account.Code)
		for _, storage := range account.Storage {
			csdb.SetState(account.Address, storage.Key, storage.Value)
		}
	}
	return []abci.ValidatorUpdate{}
}

// ExportGenesis exports genesis state
func ExportGenesis(ctx sdk.Context, k Keeper, ak types.AccountKeeper) GenesisState {
	// nolint: prealloc
	var ethGenAccounts []GenesisAccount
	accounts := ak.GetAllAccounts(ctx)

	var err error
	for _, account := range accounts {
		ethAccount, ok := account.(emint.EthAccount)
		if !ok {
			continue
		}

		addr := common.BytesToAddress(ethAccount.GetAddress().Bytes())

		var storage []GenesisStorage
		err = k.CommitStateDB.ForEachStorage(addr, func(key, value common.Hash) bool {
			storage = append(storage, NewGenesisStorage(key, value))
			return false
		})
		if err != nil {
			panic(err)
		}

		genAccount := GenesisAccount{
			Address: addr,
			Balance: k.GetBalance(ctx, addr),
			Code:    k.GetCode(ctx, addr),
			Storage: storage,
		}

		ethGenAccounts = append(ethGenAccounts, genAccount)
	}

	return GenesisState{Accounts: ethGenAccounts}
}
