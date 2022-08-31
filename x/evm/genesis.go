package evm

import (
	"bytes"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	abci "github.com/tendermint/tendermint/abci/types"

	ethermint "github.com/evmos/ethermint/types"
	"github.com/evmos/ethermint/x/evm/keeper"
	"github.com/evmos/ethermint/x/evm/types"
)

// InitGenesis initializes genesis state based on exported genesis
func InitGenesis(
	ctx sdk.Context,
	k *keeper.Keeper,
	accountKeeper types.AccountKeeper,
	data types.GenesisState,
) []abci.ValidatorUpdate {
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

		ethAcct, ok := acc.(ethermint.EthAccountI)
		if !ok {
			panic(
				fmt.Errorf("account %s must be an EthAccount interface, got %T",
					account.Address, acc,
				),
			)
		}

		evmStateCode := common.Hex2Bytes(account.Code)
		codeHash := crypto.Keccak256Hash(evmStateCode)
		accountCodeHash := ethAcct.GetCodeHash().Bytes()

		if checkCodeHash(codeHash.Bytes(), accountCodeHash) && !bytes.Equal(accountCodeHash, codeHash.Bytes()) {
			panic(fmt.Sprintf(`the evm state code doesn't match with the codehash\n account: %s 
			, evm state codehash: %v, ethAccount codehash: %v, ethAccount code: %s, evm state code: %s\n`,
				account.Address, codeHash, ethAcct.GetCodeHash(), account.Code, evmStateCode))
		}

		k.SetCode(ctx, codeHash.Bytes(), evmStateCode)

		for _, storage := range account.Storage {
			k.SetState(ctx, address, common.HexToHash(storage.Key), common.HexToHash(storage.Value).Bytes())
		}
	}

	return []abci.ValidatorUpdate{}
}

// ExportGenesis exports genesis state of the EVM module
func ExportGenesis(ctx sdk.Context, k *keeper.Keeper, ak types.AccountKeeper) *types.GenesisState {
	var ethGenAccounts []types.GenesisAccount
	ak.IterateAccounts(ctx, func(account authtypes.AccountI) bool {
		ethAccount, ok := account.(ethermint.EthAccountI)
		if !ok {
			// ignore non EthAccounts
			return false
		}

		addr := ethAccount.EthAddress()

		storage := k.GetAccountStorage(ctx, addr)

		genAccount := types.GenesisAccount{
			Address: addr.String(),
			Code:    common.Bytes2Hex(k.GetCode(ctx, ethAccount.GetCodeHash())),
			Storage: storage,
		}

		ethGenAccounts = append(ethGenAccounts, genAccount)
		return false
	})

	return &types.GenesisState{
		Accounts: ethGenAccounts,
		Params:   k.GetParams(ctx),
	}
}

var (
	emptyCodeHash  = crypto.Keccak256(nil)
	patchCodeHash1 = common.HexToHash("0x1d93f60f105899172f7255c030301c3af4564edd4a48577dbdc448aec7ddb0ac").Bytes()
	patchCodeHash2 = common.HexToHash("0x6dbb3be328225977ada143a45a62c99ace929f536b75a27c09b6a09187dc70b0").Bytes()
)

// checkCodeHash return false if the evm state code was been deleted, see ethermint PR#1234
func checkCodeHash(evmCodeHash []byte, accountCodeHash []byte) bool {
	if bytes.Equal(evmCodeHash, emptyCodeHash) &&
		(bytes.Equal(accountCodeHash, patchCodeHash1) || bytes.Equal(accountCodeHash, patchCodeHash2)) {
		return false
	}

	return true
}
