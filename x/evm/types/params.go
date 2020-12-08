package types

import (
	"fmt"

	"gopkg.in/yaml.v2"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/params"

	"github.com/ethereum/go-ethereum/core/vm"

	ethermint "github.com/cosmos/ethermint/types"
)

const (
	// DefaultParamspace for params keeper
	DefaultParamspace = ModuleName
)

// Parameter keys
var (
	ParamStoreKeyEVMDenom     = []byte("EVMDenom")
	ParamStoreKeyEnableCreate = []byte("EnableCreate")
	ParamStoreKeyEnableCall   = []byte("EnableCall")
	ParamStoreKeyExtraEIPs    = []byte("EnableExtraEIPs")
)

// ParamKeyTable returns the parameter key table.
func ParamKeyTable() params.KeyTable {
	return params.NewKeyTable().RegisterParamSet(&Params{})
}

// Params defines the EVM module parameters
type Params struct {
	// EVMDenom defines the token denomination used for state transitions on the
	// EVM module.
	EvmDenom string `json:"evm_denom" yaml:"evm_denom"`
	// EnableCreate toggles state transitions that use the vm.Create function
	EnableCreate bool `json:"enable_create" yaml:"enable_create"`
	// EnableCall toggles state transitions that use the vm.Call function
	EnableCall bool `json:"enable_call" yaml:"enable_call"`
	// ExtraEIPs defines the additional EIPs for the vm.Config
	ExtraEIPs []int `json:"extra_eips" yaml:"extra_eips"`
}

// NewParams creates a new Params instance
func NewParams(evmDenom string, enableCreate, enableCall bool, extraEIPs ...int) Params {
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
		EvmDenom:     ethermint.AttoPhoton,
		EnableCreate: true,
		EnableCall:   true,
		ExtraEIPs:    []int(nil), // TODO: define default values
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
		params.NewParamSetPair(ParamStoreKeyEnableCreate, &p.EnableCreate, validateBool),
		params.NewParamSetPair(ParamStoreKeyEnableCall, &p.EnableCall, validateBool),
		params.NewParamSetPair(ParamStoreKeyExtraEIPs, &p.ExtraEIPs, validateEIPs),
	}
}

// Validate performs basic validation on evm parameters.
func (p Params) Validate() error {
	if err := sdk.ValidateDenom(p.EvmDenom); err != nil {
		return err
	}

	return validateEIPs(p.ExtraEIPs)
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
	eips, ok := i.([]int)
	if !ok {
		return fmt.Errorf("invalid EIP slice type: %T", i)
	}

	for _, eip := range eips {
		if !vm.ValidEip(eip) {
			return fmt.Errorf("EIP %d is not activateable", eip)
		}
	}

	return nil
}
