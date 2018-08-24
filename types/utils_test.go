package types

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
)

func TestValidateSigner(t *testing.T) {
	msgs := []sdk.Msg{sdk.NewTestMsg(sdk.AccAddress(testAddr1.Bytes()))}

	// create message signing structure and bytes
	signBytes := GetStdTxSignBytes(testChainID.String(), 0, 0, newStdFee(), msgs, "")

	// require signing not to fail
	sig, err := ethcrypto.Sign(signBytes, testPrivKey1)
	require.NoError(t, err)

	// require signature to be valid
	err = ValidateSigner(signBytes, sig, testAddr1)
	require.NoError(t, err)

	sig, err = ethcrypto.Sign(signBytes, testPrivKey2)
	require.NoError(t, err)

	// require signature to be invalid
	err = ValidateSigner(signBytes, sig, testAddr1)
	require.Error(t, err)
}
