package types

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
)

func TestValidateSigner(t *testing.T) {
	msgs := []sdk.Msg{sdk.NewTestMsg(sdk.AccAddress(TestAddr1.Bytes()))}

	// create message signing structure
	signEtx := EmbeddedTxSign{TestChainID.String(), 0, 0, msgs, NewStdFee()}

	// create signing bytes and sign
	signBytes, err := signEtx.Bytes()
	require.NoError(t, err)

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
}
