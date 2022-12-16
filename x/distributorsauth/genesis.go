package distributorsauth

import (
	"fmt"

	"github.com/Entangle-Protocol/entangle-blockchain/x/distributorsauth/keeper"
	"github.com/Entangle-Protocol/entangle-blockchain/x/distributorsauth/types"

	// cosmostypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// InitGenesis initializes the module's state from a provided genesis state.
func InitGenesis(ctx sdk.Context, k keeper.Keeper, accountKeeper types.AccountKeeper, genState types.GenesisState) {
	fmt.Println("func InitGenesis")
	for _, admin := range genState.Admins {
		k.AddAdmin(ctx, admin)
	}
	for _, distributor := range genState.Distributors {
		k.AddDistributor(ctx, distributor)
	}
}

// ExportGenesis returns the module's exported genesis
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	fmt.Println("ExportGenesis")
	var admins []types.Admin
	k.IterateAdmins(ctx, func(address string, editOptions bool) bool {
		admin := types.Admin{
			Address:    address,
			EditOption: editOptions,
		}
		admins = append(admins, admin)
		return false
	})

	var distributors []types.DistributorInfo
	k.IterateDistributors(ctx, func(address string, end_date uint64) bool {
		distributor := types.DistributorInfo{
			Address: address,
			EndDate: end_date,
		}
		distributors = append(distributors, distributor)
		return false
	})

	return types.NewGenesisState(admins, distributors)
}
