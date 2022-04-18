package app

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"

	ethermint "github.com/tharsis/ethermint/types"

	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var maxTestingAccounts = 100
var seed = int64(233)

func TestRandomGenesisAccounts(t *testing.T) {
	r := rand.New(rand.NewSource(seed))
	accs := RandomAccounts(r, rand.Intn(maxTestingAccounts))

	encodingConfig := MakeEncodingConfig()
	appCodec := encodingConfig.Marshaler
	cdc := encodingConfig.Amino

	paramsKeeper := initParamsKeeper(appCodec, cdc, sdk.NewKVStoreKey(paramstypes.StoreKey), sdk.NewTransientStoreKey(paramstypes.StoreKey))
	subSpace, find := paramsKeeper.GetSubspace(authtypes.ModuleName)
	require.True(t, find)
	accountKeeper := authkeeper.NewAccountKeeper(
		appCodec, sdk.NewKVStoreKey(authtypes.StoreKey), subSpace, ethermint.ProtoAccount, maccPerms,
	)
	authModule := auth.NewAppModule(appCodec, accountKeeper, RandomGenesisAccounts)

	genesisState := simapp.NewDefaultGenesisState(appCodec)
	simState := &module.SimulationState{Accounts: accs, Cdc: appCodec, Rand: r, GenState: genesisState}
	authModule.GenerateGenesisState(simState)

	authStateBz, find := genesisState[authtypes.ModuleName]
	require.True(t, find)

	authState := new(authtypes.GenesisState)
	appCodec.MustUnmarshalJSON(authStateBz, authState)
	accounts, err := authtypes.UnpackAccounts(authState.Accounts)
	require.NoError(t, err)
	for _, acc := range accounts {
		_, ok := acc.(ethermint.EthAccountI)
		require.True(t, ok)
	}
}
