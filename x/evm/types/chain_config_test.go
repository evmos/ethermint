package types

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/ethereum/go-ethereum/common"
)

var defaultEIP150Hash = common.Hash{}.String()

func TestChainConfigValidate(t *testing.T) {
	testCases := []struct {
		name     string
		config   ChainConfig
		expError bool
	}{
		{"default", DefaultChainConfig(), false},
		{
			"valid",
			ChainConfig{
				HomesteadBlock:      sdk.OneInt(),
				DAOForkBlock:        sdk.OneInt(),
				EIP150Block:         sdk.OneInt(),
				EIP150Hash:          defaultEIP150Hash,
				EIP155Block:         sdk.OneInt(),
				EIP158Block:         sdk.OneInt(),
				ByzantiumBlock:      sdk.OneInt(),
				ConstantinopleBlock: sdk.OneInt(),
				PetersburgBlock:     sdk.OneInt(),
				IstanbulBlock:       sdk.OneInt(),
				MuirGlacierBlock:    sdk.OneInt(),
				YoloV3Block:         sdk.OneInt(),
				EWASMBlock:          sdk.OneInt(),
				CatalystBlock:       sdk.OneInt(),
			},
			false,
		},
		{
			"empty",
			ChainConfig{},
			true,
		},
		{
			"invalid HomesteadBlock",
			ChainConfig{
				HomesteadBlock: sdk.Int{},
			},
			true,
		},
		{
			"invalid DAOForkBlock",
			ChainConfig{
				HomesteadBlock: sdk.OneInt(),
				DAOForkBlock:   sdk.Int{},
			},
			true,
		},
		{
			"invalid EIP150Block",
			ChainConfig{
				HomesteadBlock: sdk.OneInt(),
				DAOForkBlock:   sdk.OneInt(),
				EIP150Block:    sdk.Int{},
			},
			true,
		},
		{
			"invalid EIP150Hash",
			ChainConfig{
				HomesteadBlock: sdk.OneInt(),
				DAOForkBlock:   sdk.OneInt(),
				EIP150Block:    sdk.OneInt(),
				EIP150Hash:     "  ",
			},
			true,
		},
		{
			"invalid EIP155Block",
			ChainConfig{
				HomesteadBlock: sdk.OneInt(),
				DAOForkBlock:   sdk.OneInt(),
				EIP150Block:    sdk.OneInt(),
				EIP150Hash:     defaultEIP150Hash,
				EIP155Block:    sdk.Int{},
			},
			true,
		},
		{
			"invalid EIP158Block",
			ChainConfig{
				HomesteadBlock: sdk.OneInt(),
				DAOForkBlock:   sdk.OneInt(),
				EIP150Block:    sdk.OneInt(),
				EIP150Hash:     defaultEIP150Hash,
				EIP155Block:    sdk.OneInt(),
				EIP158Block:    sdk.Int{},
			},
			true,
		},
		{
			"invalid ByzantiumBlock",
			ChainConfig{
				HomesteadBlock: sdk.OneInt(),
				DAOForkBlock:   sdk.OneInt(),
				EIP150Block:    sdk.OneInt(),
				EIP150Hash:     defaultEIP150Hash,
				EIP155Block:    sdk.OneInt(),
				EIP158Block:    sdk.OneInt(),
				ByzantiumBlock: sdk.Int{},
			},
			true,
		},
		{
			"invalid ConstantinopleBlock",
			ChainConfig{
				HomesteadBlock:      sdk.OneInt(),
				DAOForkBlock:        sdk.OneInt(),
				EIP150Block:         sdk.OneInt(),
				EIP150Hash:          defaultEIP150Hash,
				EIP155Block:         sdk.OneInt(),
				EIP158Block:         sdk.OneInt(),
				ByzantiumBlock:      sdk.OneInt(),
				ConstantinopleBlock: sdk.Int{},
			},
			true,
		},
		{
			"invalid PetersburgBlock",
			ChainConfig{
				HomesteadBlock:      sdk.OneInt(),
				DAOForkBlock:        sdk.OneInt(),
				EIP150Block:         sdk.OneInt(),
				EIP150Hash:          defaultEIP150Hash,
				EIP155Block:         sdk.OneInt(),
				EIP158Block:         sdk.OneInt(),
				ByzantiumBlock:      sdk.OneInt(),
				ConstantinopleBlock: sdk.OneInt(),
				PetersburgBlock:     sdk.Int{},
			},
			true,
		},
		{
			"invalid IstanbulBlock",
			ChainConfig{
				HomesteadBlock:      sdk.OneInt(),
				DAOForkBlock:        sdk.OneInt(),
				EIP150Block:         sdk.OneInt(),
				EIP150Hash:          defaultEIP150Hash,
				EIP155Block:         sdk.OneInt(),
				EIP158Block:         sdk.OneInt(),
				ByzantiumBlock:      sdk.OneInt(),
				ConstantinopleBlock: sdk.OneInt(),
				PetersburgBlock:     sdk.OneInt(),
				IstanbulBlock:       sdk.Int{},
			},
			true,
		},
		{
			"invalid MuirGlacierBlock",
			ChainConfig{
				HomesteadBlock:      sdk.OneInt(),
				DAOForkBlock:        sdk.OneInt(),
				EIP150Block:         sdk.OneInt(),
				EIP150Hash:          defaultEIP150Hash,
				EIP155Block:         sdk.OneInt(),
				EIP158Block:         sdk.OneInt(),
				ByzantiumBlock:      sdk.OneInt(),
				ConstantinopleBlock: sdk.OneInt(),
				PetersburgBlock:     sdk.OneInt(),
				IstanbulBlock:       sdk.OneInt(),
				MuirGlacierBlock:    sdk.Int{},
			},
			true,
		},
		{
			"invalid YoloV2Block",
			ChainConfig{
				HomesteadBlock:      sdk.OneInt(),
				DAOForkBlock:        sdk.OneInt(),
				EIP150Block:         sdk.OneInt(),
				EIP150Hash:          defaultEIP150Hash,
				EIP155Block:         sdk.OneInt(),
				EIP158Block:         sdk.OneInt(),
				ByzantiumBlock:      sdk.OneInt(),
				ConstantinopleBlock: sdk.OneInt(),
				PetersburgBlock:     sdk.OneInt(),
				IstanbulBlock:       sdk.OneInt(),
				MuirGlacierBlock:    sdk.OneInt(),
				YoloV3Block:         sdk.Int{},
			},
			true,
		},
		{
			"invalid EWASMBlock",
			ChainConfig{
				HomesteadBlock:      sdk.OneInt(),
				DAOForkBlock:        sdk.OneInt(),
				EIP150Block:         sdk.OneInt(),
				EIP150Hash:          defaultEIP150Hash,
				EIP155Block:         sdk.OneInt(),
				EIP158Block:         sdk.OneInt(),
				ByzantiumBlock:      sdk.OneInt(),
				ConstantinopleBlock: sdk.OneInt(),
				PetersburgBlock:     sdk.OneInt(),
				IstanbulBlock:       sdk.OneInt(),
				MuirGlacierBlock:    sdk.OneInt(),
				YoloV3Block:         sdk.OneInt(),
				EWASMBlock:          sdk.Int{},
			},
			true,
		},
		{
			"invalid CatalystBlock",
			ChainConfig{
				HomesteadBlock:      sdk.OneInt(),
				DAOForkBlock:        sdk.OneInt(),
				EIP150Block:         sdk.OneInt(),
				EIP150Hash:          defaultEIP150Hash,
				EIP155Block:         sdk.OneInt(),
				EIP158Block:         sdk.OneInt(),
				ByzantiumBlock:      sdk.OneInt(),
				ConstantinopleBlock: sdk.OneInt(),
				PetersburgBlock:     sdk.OneInt(),
				IstanbulBlock:       sdk.OneInt(),
				MuirGlacierBlock:    sdk.OneInt(),
				YoloV3Block:         sdk.OneInt(),
				EWASMBlock:          sdk.OneInt(),
				CatalystBlock:       sdk.Int{},
			},
			true,
		},
	}

	for _, tc := range testCases {
		err := tc.config.Validate()

		if tc.expError {
			require.Error(t, err, tc.name)
		} else {
			require.NoError(t, err, tc.name)
		}
	}
}
