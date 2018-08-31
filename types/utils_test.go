package types

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
)

func TestValidateSigner(t *testing.T) {
	msgs := []sdk.Msg{sdk.NewTestMsg(sdk.AccAddress(TestAddr1.Bytes()))}

	// create message signing structure and bytes
	signBytes := GetStdTxSignBytes(TestChainID.String(), 0, 0, NewTestStdFee(), msgs, "")

	// require signing not to fail
	sig, err := ethcrypto.Sign(signBytes, TestPrivKey1)
	require.NoError(t, err)

	// require signature to be valid
	err = ValidateSigner(signBytes, sig, TestAddr1)
	require.NoError(t, err)

	sig, err = ethcrypto.Sign(signBytes, TestPrivKey2)
	require.NoError(t, err)

	// require signature to be invalid
	err = ValidateSigner(signBytes, sig, TestAddr1)
	require.Error(t, err)

	// require invalid signature bytes return an error
	err = ValidateSigner([]byte{}, sig, TestAddr2)
	require.Error(t, err)
}
