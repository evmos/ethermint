package types

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"

	ethcmn "github.com/ethereum/go-ethereum/common"
)

func TestValidateGenesis(t *testing.T) {

	testCases := []struct {
		name     string
		genState GenesisState
		expPass  bool
	}{
		{
			name:     "default",
			genState: DefaultGenesisState(),
			expPass:  true,
		},
		{
			name: "empty account address bytes",
			genState: GenesisState{
				Accounts: []GenesisAccount{
					{
						Address: ethcmn.Address{},
						Balance: big.NewInt(1),
					},
				},
			},
			expPass: false,
		},
		{
			name: "nil account balance",
			genState: GenesisState{
				Accounts: []GenesisAccount{
					{
						Address: ethcmn.BytesToAddress([]byte{1, 2, 3, 4, 5}),
						Balance: nil,
					},
				},
			},
			expPass: false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		err := tc.genState.Validate()
		if tc.expPass {
			require.NoError(t, err, tc.name)
		} else {
			require.Error(t, err, tc.name)
		}
	}
}
