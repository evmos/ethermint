package eip712_test

import (
	"encoding/hex"
	"strings"
	"testing"

	"cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/client"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	"github.com/cosmos/cosmos-sdk/x/auth/ante"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/evmos/ethermint/app"
	"github.com/evmos/ethermint/encoding"
	"github.com/evmos/ethermint/ethereum/eip712"
	"github.com/evmos/ethermint/tests"
	"github.com/evmos/ethermint/types"
	evmtypes "github.com/evmos/ethermint/x/evm/types"
	"github.com/stretchr/testify/require"
)

// Testing Constants
var (
	chainId = "ethermint_9000-1"
	ctx     = client.Context{}.WithTxConfig(
		encoding.MakeConfig(app.ModuleBasics).TxConfig,
	)
)
var feePayerAddress = "ethm17xpfvakm2amg962yls6f84z3kell8c5lthdzgl"

type TestCaseStruct struct {
	txBuilder              client.TxBuilder
	expectedFeePayer       string
	expectedGas            uint64
	expectedFee            math.Int
	expectedMemo           string
	expectedMsg            string
	expectedSignatureBytes []byte
}

func TestLedgerPreprocessing(t *testing.T) {
	// Update bech32 prefix
	sdk.GetConfig().SetBech32PrefixForAccount("ethm", "")

	testCases := []TestCaseStruct{
		createBasicTestCase(t),
		createPopulatedTestCase(t),
	}

	for _, tc := range testCases {
		// Run pre-processing
		err := eip712.PreprocessLedgerTx(
			chainId,
			keyring.TypeLedger,
			tc.txBuilder,
		)

		require.NoError(t, err)

		// Verify Web3 extension matches expected
		hasExtOptsTx, ok := tc.txBuilder.(ante.HasExtensionOptionsTx)
		require.True(t, ok)
		require.True(t, len(hasExtOptsTx.GetExtensionOptions()) == 1)

		expectedExt := types.ExtensionOptionsWeb3Tx{
			TypedDataChainID: 9000,
			FeePayer:         feePayerAddress,
			FeePayerSig:      tc.expectedSignatureBytes,
		}

		expectedExtAny, err := codectypes.NewAnyWithValue(&expectedExt)
		require.NoError(t, err)

		actualExtAny := hasExtOptsTx.GetExtensionOptions()[0]
		require.Equal(t, expectedExtAny, actualExtAny)

		// Verify signature type matches expected
		signatures, err := tc.txBuilder.GetTx().GetSignaturesV2()
		require.NoError(t, err)
		require.Equal(t, len(signatures), 1)

		txSig := signatures[0].Data.(*signing.SingleSignatureData)
		require.Equal(t, txSig.SignMode, signing.SignMode_SIGN_MODE_LEGACY_AMINO_JSON)

		// Verify signature is blank
		require.Equal(t, len(txSig.Signature), 0)

		// Verify tx fields are unchanged
		tx := tc.txBuilder.GetTx()

		require.Equal(t, tx.FeePayer().String(), tc.expectedFeePayer)
		require.Equal(t, tx.GetGas(), tc.expectedGas)
		require.Equal(t, tx.GetFee().AmountOf(evmtypes.DefaultParams().EvmDenom), tc.expectedFee)
		require.Equal(t, tx.GetMemo(), tc.expectedMemo)

		// Verify message is unchanged
		if tc.expectedMsg != "" {
			require.Equal(t, len(tx.GetMsgs()), 1)
			require.Equal(t, tx.GetMsgs()[0].String(), tc.expectedMsg)
		} else {
			require.Equal(t, len(tx.GetMsgs()), 0)
		}
	}
}

func TestBlankTxBuilder(t *testing.T) {
	txBuilder := ctx.TxConfig.NewTxBuilder()

	err := eip712.PreprocessLedgerTx(
		chainId,
		keyring.TypeLedger,
		txBuilder,
	)

	require.Error(t, err)
}

func TestNonLedgerTxBuilder(t *testing.T) {
	txBuilder := ctx.TxConfig.NewTxBuilder()

	err := eip712.PreprocessLedgerTx(
		chainId,
		keyring.TypeLocal,
		txBuilder,
	)

	require.NoError(t, err)
}

func TestInvalidChainId(t *testing.T) {
	txBuilder := ctx.TxConfig.NewTxBuilder()

	err := eip712.PreprocessLedgerTx(
		"invalid-chain-id",
		keyring.TypeLedger,
		txBuilder,
	)

	require.Error(t, err)
}

func createBasicTestCase(t *testing.T) TestCaseStruct {
	t.Helper()
	txBuilder := ctx.TxConfig.NewTxBuilder()

	feePayer, err := sdk.AccAddressFromBech32(feePayerAddress)
	require.NoError(t, err)

	txBuilder.SetFeePayer(feePayer)

	// Create signature unrelated to payload for testing
	signatureHex := strings.Repeat("01", 65)
	signatureBytes, err := hex.DecodeString(signatureHex)
	require.NoError(t, err)

	_, privKey := tests.NewAddrKey()
	sigsV2 := signing.SignatureV2{
		PubKey: privKey.PubKey(), // Use unrelated public key for testing
		Data: &signing.SingleSignatureData{
			SignMode:  signing.SignMode_SIGN_MODE_DIRECT,
			Signature: signatureBytes,
		},
		Sequence: 0,
	}

	txBuilder.SetSignatures(sigsV2)
	return TestCaseStruct{
		txBuilder:              txBuilder,
		expectedFeePayer:       feePayer.String(),
		expectedGas:            0,
		expectedFee:            math.NewInt(0),
		expectedMemo:           "",
		expectedMsg:            "",
		expectedSignatureBytes: signatureBytes,
	}
}

func createPopulatedTestCase(t *testing.T) TestCaseStruct {
	t.Helper()
	basicTestCase := createBasicTestCase(t)
	txBuilder := basicTestCase.txBuilder

	gasLimit := uint64(200000)
	memo := ""
	denom := evmtypes.DefaultParams().EvmDenom
	feeAmount := math.NewInt(2000)

	txBuilder.SetFeeAmount(sdk.NewCoins(
		sdk.NewCoin(
			denom,
			feeAmount,
		)))

	txBuilder.SetGasLimit(gasLimit)
	txBuilder.SetMemo(memo)

	msgSend := banktypes.MsgSend{
		FromAddress: feePayerAddress,
		ToAddress:   "ethm12luku6uxehhak02py4rcz65zu0swh7wjun6msa",
		Amount: sdk.NewCoins(
			sdk.NewCoin(
				evmtypes.DefaultParams().EvmDenom,
				math.NewInt(10000000),
			),
		),
	}

	txBuilder.SetMsgs(&msgSend)

	return TestCaseStruct{
		txBuilder:              txBuilder,
		expectedFeePayer:       basicTestCase.expectedFeePayer,
		expectedGas:            gasLimit,
		expectedFee:            feeAmount,
		expectedMemo:           memo,
		expectedMsg:            msgSend.String(),
		expectedSignatureBytes: basicTestCase.expectedSignatureBytes,
	}
}
