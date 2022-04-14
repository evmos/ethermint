package simulation

import (
	"encoding/json"
	"fmt"

	"github.com/cosmos/cosmos-sdk/types/module"

	"github.com/tharsis/ethermint/x/evm/types"
)

// RandomizedGenState generates a random GenesisState for nft
func RandomizedGenState(simState *module.SimulationState) {
	params := types.NewParams(types.DefaultEVMDenom, true, true, types.DefaultChainConfig())
	if simState.Rand.Uint32()%2 == 0 {
		params = types.NewParams(types.DefaultEVMDenom, true, true, types.DefaultChainConfig(), 1344, 1884, 2200, 2929, 3198, 3529)
	}
	evmGenesis := types.NewGenesisState(params, []types.GenesisAccount{})

	bz, err := json.MarshalIndent(evmGenesis, "", " ")
	if err != nil {
		panic(err)
	}
	fmt.Printf("Selected randomly generated %s parameters:\n%s\n", types.ModuleName, bz)

	simState.GenState[types.ModuleName] = simState.Cdc.MustMarshalJSON(evmGenesis)
}
