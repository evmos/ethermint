package types

import (
	"fmt"

	yaml "gopkg.in/yaml.v2"

	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/cosmos/ethermint/types"
	"github.com/ethereum/go-ethereum/core/vm"
)

var _ paramtypes.ParamSet = &Params{}

const (
	DefaultEVMDenom = types.AttoPhoton
)

// Parameter keys
var (
	ParamStoreKeyEVMDenom     = []byte("EVMDenom")
	ParamStoreKeyEnableCreate = []byte("EnableCreate")
	ParamStoreKeyEnableCall   = []byte("EnableCall")
	ParamStoreKeyExtraEIPs    = []byte("EnableExtraEIPs")
)

// ParamKeyTable returns the parameter key table.
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

// NewParams creates a new Params instance
func NewParams(evmDenom string, enableCreate, enableCall bool, extraEIPs ...int64) Params {
	return Params{
		EvmDenom:     evmDenom,
		EnableCreate: enableCreate,
		EnableCall:   enableCall,
		ExtraEIPs:    extraEIPs,
	}
}

// DefaultParams returns default evm parameters
func DefaultParams() Params {
	return Params{
		EvmDenom:     DefaultEVMDenom,
		EnableCreate: true,
		EnableCall:   true,
		ExtraEIPs:    []int64{2929, 2200, 1884, 1344},
	}
}

// String implements the fmt.Stringer interface
func (p Params) String() string {
	out, _ := yaml.Marshal(p)
	return string(out)
}

// ParamSetPairs returns the parameter set pairs.
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(ParamStoreKeyEVMDenom, &p.EvmDenom, validateEVMDenom),
		paramtypes.NewParamSetPair(ParamStoreKeyEnableCreate, &p.EnableCreate, validateBool),
		paramtypes.NewParamSetPair(ParamStoreKeyEnableCall, &p.EnableCall, validateBool),
		paramtypes.NewParamSetPair(ParamStoreKeyExtraEIPs, &p.ExtraEIPs, validateEIPs),
	}
}

// Validate performs basic validation on evm parameters.
func (p Params) Validate() error {
	if err := sdk.ValidateDenom(p.EvmDenom); err != nil {
		return err
	}

	return validateEIPs(p.ExtraEIPs)
}

// EIPs returns the ExtraEips as a int slice
func (p Params) EIPs() []int {
	eips := make([]int, len(p.ExtraEIPs))
	for i, eip := range p.ExtraEIPs {
		eips[i] = int(eip)
	}
	return eips
}

func validateEVMDenom(i interface{}) error {
	denom, ok := i.(string)
	if !ok {
		return fmt.Errorf("invalid parameter EVM denom type: %T", i)
	}

	return sdk.ValidateDenom(denom)
}

func validateBool(i interface{}) error {
	_, ok := i.(bool)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	return nil
}

func validateEIPs(i interface{}) error {
	eips, ok := i.([]int64)
	if !ok {
		return fmt.Errorf("invalid EIP slice type: %T", i)
	}

	for _, eip := range eips {
		if !vm.ValidEip(int(eip)) {
			return fmt.Errorf("EIP %d is not activateable, valid EIPS are: %s", eip, vm.ActivateableEips())
		}
	}

	return nil
}
