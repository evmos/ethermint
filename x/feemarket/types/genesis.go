package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// DefaultGenesisState sets default fee market genesis state.
func DefaultGenesisState() *GenesisState {
	return &GenesisState{
		Params:   DefaultParams(),
		BaseFee:  sdk.ZeroInt(),
		BlockGas: 0,
	}
}

// Validate performs basic genesis state validation returning an error upon any
// failure.
func (gs GenesisState) Validate() error {
	if gs.BaseFee.IsNegative() {
		return fmt.Errorf("base fee cannot be negative: %s", gs.BaseFee)
	}

	return gs.Params.Validate()
}
