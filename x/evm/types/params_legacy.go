/*
NOTE: Usage of x/params to manage parameters is deprecated in favor of x/gov
controlled execution of MsgUpdateParams messages. These types remains solely
for migration purposes and will be removed in a future release.
*/
package types

import paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"

// Parameter keys
var (
	ParamStoreKeyEVMDenom            = []byte("EVMDenom")
	ParamStoreKeyEnableCreate        = []byte("EnableCreate")
	ParamStoreKeyEnableCall          = []byte("EnableCall")
	ParamStoreKeyExtraEIPs           = []byte("EnableExtraEIPs")
	ParamStoreKeyChainConfig         = []byte("ChainConfig")
	ParamStoreKeyAllowUnprotectedTxs = []byte("AllowUnprotectedTxs")
)

// ParamKeyTable returns the parameter key table.
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

// ParamSetPairs returns the parameter set pairs.
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(ParamStoreKeyEVMDenom, &p.EvmDenom, validateEVMDenom),
		paramtypes.NewParamSetPair(ParamStoreKeyEnableCreate, &p.EnableCreate, validateBool),
		paramtypes.NewParamSetPair(ParamStoreKeyEnableCall, &p.EnableCall, validateBool),
		paramtypes.NewParamSetPair(ParamStoreKeyExtraEIPs, &p.ExtraEIPs.EIPs, validateEIPs),
		paramtypes.NewParamSetPair(ParamStoreKeyChainConfig, &p.ChainConfig, validateChainConfig),
		paramtypes.NewParamSetPair(ParamStoreKeyAllowUnprotectedTxs, &p.AllowUnprotectedTxs, validateBool),
	}
}
