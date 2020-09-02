package types

import (
	"fmt"

	"gopkg.in/yaml.v2"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/params"

	ethermint "github.com/cosmos/ethermint/types"
)

const (
	// DefaultParamspace for params keeper
	DefaultParamspace = ModuleName
)

// Parameter keys
var (
	ParamStoreKeyEVMDenom = []byte("EVMDenom")
)

// ParamKeyTable returns the parameter key table.
func ParamKeyTable() params.KeyTable {
	return params.NewKeyTable().RegisterParamSet(&Params{})
}

// Params defines the EVM module parameters
type Params struct {
	EvmDenom string `json:"evm_denom" yaml:"evm_denom"`
}

// NewParams creates a new Params instance
func NewParams(evmDenom string) Params {
	return Params{
		EvmDenom: evmDenom,
	}
}

// DefaultParams returns default evm parameters
func DefaultParams() Params {
	return Params{
		EvmDenom: ethermint.DenomDefault,
	}
}

// String implements the fmt.Stringer interface
func (p Params) String() string {
	out, _ := yaml.Marshal(p)
	return string(out)
}

// ParamSetPairs returns the parameter set pairs.
func (p *Params) ParamSetPairs() params.ParamSetPairs {
	return params.ParamSetPairs{
		params.NewParamSetPair(ParamStoreKeyEVMDenom, &p.EvmDenom, validateEVMDenom),
	}
}

// Validate performs basic validation on evm parameters.
func (p Params) Validate() error {
	return sdk.ValidateDenom(p.EvmDenom)
}

func validateEVMDenom(i interface{}) error {
	denom, ok := i.(string)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	return sdk.ValidateDenom(denom)
}
