package types

import (
	"fmt"
	"math"

	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

const (
	DefaultBaseFeeChangeDenominator = 8
	DefaultElasticityMultiplier     = 2
	DefaultInitialBaseFee           = 1000000000
)

var _ paramtypes.ParamSet = &Params{}

// Parameter keys
var (
	ParamStoreKeyNoBaseFee                = []byte("NoBaseFee")
	ParamStoreKeyBaseFeeChangeDenominator = []byte("BaseFeeChangeDenominator")
	ParamStoreKeyElasticityMultiplier     = []byte("ElasticityMultiplier")
	ParamStoreKeyInitialBaseFee           = []byte("InitialBaseFee")
	ParamStoreKeyEnableHeight             = []byte("EnableHeight")
)

// ParamKeyTable returns the parameter key table.
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

// NewParams creates a new Params instance
func NewParams(noBaseFee bool, baseFeeChangeDenom, elasticityMultiplier uint32, initialBaseFee, enableHeight int64) Params {
	return Params{
		NoBaseFee:                noBaseFee,
		BaseFeeChangeDenominator: baseFeeChangeDenom,
		ElasticityMultiplier:     elasticityMultiplier,
		InitialBaseFee:           initialBaseFee,
		EnableHeight:             enableHeight,
	}
}

// DefaultParams returns default evm parameters
func DefaultParams() Params {
	// TODO: use geth parameters
	return Params{
		NoBaseFee:                true,
		BaseFeeChangeDenominator: DefaultBaseFeeChangeDenominator,
		ElasticityMultiplier:     DefaultElasticityMultiplier,
		InitialBaseFee:           DefaultInitialBaseFee,
		EnableHeight:             math.MaxInt64,
	}
}

// ParamSetPairs returns the parameter set pairs.
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(ParamStoreKeyNoBaseFee, &p.NoBaseFee, validateBool),
		paramtypes.NewParamSetPair(ParamStoreKeyBaseFeeChangeDenominator, &p.BaseFeeChangeDenominator, validateBaseFeeChangeDenominator),
		paramtypes.NewParamSetPair(ParamStoreKeyElasticityMultiplier, &p.ElasticityMultiplier, validateElasticityMultiplier),
		paramtypes.NewParamSetPair(ParamStoreKeyInitialBaseFee, &p.InitialBaseFee, validateInitialBaseFee),
		paramtypes.NewParamSetPair(ParamStoreKeyEnableHeight, &p.EnableHeight, validateEnableHeight),
	}
}

// Validate performs basic validation on fee market parameters.
func (p Params) Validate() error {
	if p.BaseFeeChangeDenominator == 0 {
		return fmt.Errorf("base fee change denominator cannot be 0")
	}

	if p.InitialBaseFee < 0 {
		return fmt.Errorf("initial base fee cannot be negative: %d", p.InitialBaseFee)
	}

	if p.EnableHeight < 0 {
		return fmt.Errorf("enable height cannot be negative: %d", p.EnableHeight)
	}

	return nil
}

func validateBool(i interface{}) error {
	_, ok := i.(bool)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	return nil
}

func validateBaseFeeChangeDenominator(i interface{}) error {
	value, ok := i.(uint32)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if value == 0 {
		return fmt.Errorf("base fee change denominator cannot be 0")
	}

	return nil
}

func validateElasticityMultiplier(i interface{}) error {
	_, ok := i.(uint32)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	return nil
}

func validateInitialBaseFee(i interface{}) error {
	value, ok := i.(int64)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if value < 0 {
		return fmt.Errorf("initial base fee cannot be negative: %d", value)
	}

	return nil
}

func validateEnableHeight(i interface{}) error {
	value, ok := i.(int64)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if value < 0 {
		return fmt.Errorf("enable height cannot be negative: %d", value)
	}

	return nil
}
