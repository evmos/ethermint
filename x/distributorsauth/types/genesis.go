package types

import (
	"encoding/json"
	fmt "fmt"
	"sort"

	"github.com/cosmos/cosmos-sdk/codec"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// this line is used by starport scaffolding # genesis/types/import

// DefaultIndex is the default global index
const DefaultIndex uint64 = 1

// DefaultGenesis returns the default genesis state
func DefaultGenesis() *GenesisState {
	fmt.Println("DefaultGenesis")
	return &GenesisState{
		Admins:       []Admin{},
		Distributors: []DistributorInfo{},
	}
}

// NewGenesisState creates a new genesis state.
func NewGenesisState(admins []Admin, distributors []DistributorInfo) *GenesisState {
	fmt.Println("NewGenesisState")
	return &GenesisState{
		Admins:       admins,
		Distributors: distributors,
	}
}

// Validate performs basic genesis state validation returning an error upon any
// failure.
func (gs GenesisState) Validate() error {
	for _, admin := range gs.Admins {
		if _, err := sdk.AccAddressFromBech32(admin.Address); err != nil {
			return sdkerrors.Wrap(ErrWrongAdminAddress, admin.Address)
		}
	}

	for _, distributor := range gs.Distributors {
		if _, err := sdk.AccAddressFromBech32(distributor.Address); err != nil {
			return sdkerrors.Wrap(ErrWrongAdminAddress, distributor.Address)
		}
	}

	return nil
}

func GetGenesisStateFromAppState(cdc codec.JSONCodec, appState map[string]json.RawMessage) *GenesisState {
	var genesisState GenesisState

	if appState[ModuleName] != nil {
		cdc.MustUnmarshalJSON(appState[ModuleName], &genesisState)
	}

	return &genesisState
}

// GenesisAdminIterator implements genesis admin iteration.
type GenesisAdminIterator struct{}

// IterateGenesisAccounts iterates over all the genesis admins found in
// appGenesis and invokes a callback on each genesis admin. If any call
// returns true, iteration stops.

type GenesisAdmins []Admin
type GenesisDistributors []DistributorInfo

func ContainsGenesisAdmin(addr string, admins []Admin) bool {
	for _, adm := range admins {
		if adm.Address == addr {
			return true
		}
	}

	return false
}

func ContainsGenesisDistributor(addr string, distributors []DistributorInfo) bool {
	for _, adm := range distributors {
		if adm.Address == addr {
			return true
		}
	}

	return false
}

func SanitizeGenesisAdmin(genAdmins GenesisAdmins) GenesisAdmins {
	sort.Slice(genAdmins, func(i, j int) bool {
		return genAdmins[i].GetAddress() < genAdmins[j].GetAddress()
	})

	return genAdmins
}

func SanitizeGenesisDistributor(genDistr GenesisDistributors) GenesisDistributors {
	sort.Slice(genDistr, func(i, j int) bool {
		return genDistr[i].GetAddress() < genDistr[j].GetAddress()
	})

	return genDistr
}
