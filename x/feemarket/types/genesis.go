package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/params"
)

// DefaultGenesisState sets default fee market genesis state.
func DefaultGenesisState() *GenesisState {
	return &GenesisState{
		Params: DefaultParams(),
		// the default base fee should be initialized because the default enable height is zero.
		DefaultBaseFee: sdk.NewIntFromUint64(params.InitialBaseFee),
		BlockGas:       0,
	}
}

// NewGenesisState creates a new genesis state.
func NewGenesisState(params Params, baseFee sdk.Int, blockGas uint64) *GenesisState {
	return &GenesisState{
		Params:         params,
		DefaultBaseFee: baseFee,
		BlockGas:       blockGas,
	}
}

// Validate performs basic genesis state validation returning an error upon any
// failure.
func (gs GenesisState) Validate() error {
	if gs.DefaultBaseFee.IsNegative() {
		return fmt.Errorf("base fee cannot be negative: %s", gs.DefaultBaseFee)
	}

	return gs.Params.Validate()
}
