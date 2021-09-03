package evm

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/ethereum/go-ethereum/common"
	abci "github.com/tendermint/tendermint/abci/types"

	ethermint "github.com/tharsis/ethermint/types"
	"github.com/tharsis/ethermint/x/evm/keeper"
	"github.com/tharsis/ethermint/x/evm/types"
)

// InitGenesis initializes genesis state based on exported genesis
func InitGenesis(
	ctx sdk.Context,
	k *keeper.Keeper,
	accountKeeper types.AccountKeeper, // nolint: interfacer
	data types.GenesisState,
) []abci.ValidatorUpdate {
	k.WithContext(ctx)
	k.WithChainID(ctx)

	k.SetParams(ctx, data.Params)

	// ensure evm module account is set
	if addr := accountKeeper.GetModuleAddress(types.ModuleName); addr == nil {
		panic("the EVM module account has not been set")
	}

	for _, account := range data.Accounts {
		address := common.HexToAddress(account.Address)
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

		k.SetCode(address, common.Hex2Bytes(account.Code))

		for _, storage := range account.Storage {
			k.SetState(address, common.HexToHash(storage.Key), common.HexToHash(storage.Value))
		}
	}

	for _, txLog := range data.TxsLogs {
		k.SetLogs(common.HexToHash(txLog.Hash), txLog.EthLogs())
	}

	return []abci.ValidatorUpdate{}
}

// ExportGenesis exports genesis state of the EVM module
func ExportGenesis(ctx sdk.Context, k *keeper.Keeper, ak types.AccountKeeper) *types.GenesisState {
	k.WithContext(ctx)

	var ethGenAccounts []types.GenesisAccount
	ak.IterateAccounts(ctx, func(account authtypes.AccountI) bool {
		ethAccount, ok := account.(*ethermint.EthAccount)
		if !ok {
			// ignore non EthAccounts
			return false
		}

		addr := ethAccount.EthAddress()

		storage, err := k.GetAccountStorage(ctx, addr)
		if err != nil {
			panic(err)
		}

		genAccount := types.GenesisAccount{
			Address: addr.String(),
			Code:    common.Bytes2Hex(k.GetCode(addr)),
			Storage: storage,
		}

		ethGenAccounts = append(ethGenAccounts, genAccount)
		return false
	})

	return &types.GenesisState{
		Accounts: ethGenAccounts,
		TxsLogs:  k.GetAllTxLogs(ctx),
		Params:   k.GetParams(ctx),
	}
}
