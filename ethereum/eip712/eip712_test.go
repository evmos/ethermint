package eip712_test

import (
	"fmt"
	"math/big"
	"reflect"
	"testing"

	"cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/simapp/params"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/evmos/ethermint/ethereum/eip712"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/evmos/ethermint/crypto/ethsecp256k1"

	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/evmos/ethermint/app"
	"github.com/evmos/ethermint/encoding"

	txtypes "github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"

	distributiontypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	govtypesv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/stretchr/testify/suite"
)

// Unit tests for single-signer EIP-712 signature verification. Multi-signer verification tests are included
// in ante_test.go.

type EIP712TestSuite struct {
	suite.Suite

	config                   params.EncodingConfig
	clientCtx                client.Context
	useLegacyEIP712TypedData bool
}

func TestEIP712TestSuite(t *testing.T) {
	suite.Run(t, &EIP712TestSuite{})
	suite.Run(t, &EIP712TestSuite{
		useLegacyEIP712TypedData: true,
	})
}

// Set up test env to replicate prod. environment
func (suite *EIP712TestSuite) SetupTest() {
	suite.config = encoding.MakeConfig(app.ModuleBasics)
	suite.clientCtx = client.Context{}.WithTxConfig(suite.config.TxConfig)

	sdk.GetConfig().SetBech32PrefixForAccount("ethm", "")
	eip712.SetEncodingConfig(suite.config)
}

// Helper to create random test addresses for messages
func (suite *EIP712TestSuite) createTestAddress() sdk.AccAddress {
	privkey, _ := ethsecp256k1.GenerateKey()
	key, err := privkey.ToECDSA()
	suite.Require().NoError(err)

	addr := crypto.PubkeyToAddress(key.PublicKey)

	return addr.Bytes()
}

// Helper to create random keypair for signing + verification
func (suite *EIP712TestSuite) createTestKeyPair() (*ethsecp256k1.PrivKey, *ethsecp256k1.PubKey) {
	privKey, err := ethsecp256k1.GenerateKey()
	suite.Require().NoError(err)

	pubKey := &ethsecp256k1.PubKey{
		Key: privKey.PubKey().Bytes(),
	}
	suite.Require().Implements((*cryptotypes.PubKey)(nil), pubKey)

	return privKey, pubKey
}

// Helper to create instance of sdk.Coins[] with single coin
func (suite *EIP712TestSuite) makeCoins(denom string, amount math.Int) sdk.Coins {
	return sdk.NewCoins(
		sdk.NewCoin(
			denom,
			amount,
		),
	)
}

func (suite *EIP712TestSuite) TestEIP712PayloadAndSignature() {
	suite.SetupTest()

	signModes := []signing.SignMode{
		signing.SignMode_SIGN_MODE_DIRECT,
		signing.SignMode_SIGN_MODE_LEGACY_AMINO_JSON,
	}

	// Fixed test address
	testAddress := suite.createTestAddress()

	testCases := []struct {
		title         string
		chainId       string
		fee           txtypes.Fee
		memo          string
		msgs          []sdk.Msg
		accountNumber uint64
		sequence      uint64
		timeoutHeight uint64
		expectSuccess bool
	}{
		{
			title: "Succeeds - Standard MsgSend",
			fee: txtypes.Fee{
				Amount:   suite.makeCoins("aphoton", math.NewInt(2000)),
				GasLimit: 20000,
			},
			memo: "",
			msgs: []sdk.Msg{
				banktypes.NewMsgSend(
					suite.createTestAddress(),
					suite.createTestAddress(),
					suite.makeCoins("photon", math.NewInt(1)),
				),
			},
			accountNumber: 8,
			sequence:      5,
			expectSuccess: true,
		},
		{
			title: "Succeeds - Standard MsgVote",
			fee: txtypes.Fee{
				Amount:   suite.makeCoins("aphoton", math.NewInt(2000)),
				GasLimit: 20000,
			},
			memo: "",
			msgs: []sdk.Msg{
				govtypes.NewMsgVote(
					suite.createTestAddress(),
					5,
					govtypes.OptionNo,
				),
			},
			accountNumber: 25,
			sequence:      78,
			expectSuccess: true,
		},
		{
			title: "Succeeds - Standard MsgDelegate",
			fee: txtypes.Fee{
				Amount:   suite.makeCoins("aphoton", math.NewInt(2000)),
				GasLimit: 20000,
			},
			memo: "",
			msgs: []sdk.Msg{
				stakingtypes.NewMsgDelegate(
					suite.createTestAddress(),
					sdk.ValAddress(suite.createTestAddress()),
					suite.makeCoins("photon", math.NewInt(1))[0],
				),
			},
			accountNumber: 25,
			sequence:      78,
			expectSuccess: true,
		},
		{
			title: "Succeeds - Standard MsgWithdrawDelegationReward",
			fee: txtypes.Fee{
				Amount:   suite.makeCoins("aphoton", math.NewInt(2000)),
				GasLimit: 20000,
			},
			memo: "",
			msgs: []sdk.Msg{
				distributiontypes.NewMsgWithdrawDelegatorReward(
					suite.createTestAddress(),
					sdk.ValAddress(suite.createTestAddress()),
				),
			},
			accountNumber: 25,
			sequence:      78,
			expectSuccess: true,
		},
		{
			title: "Succeeds - Two Single-Signer MsgDelegate",
			fee: txtypes.Fee{
				Amount:   suite.makeCoins("aphoton", math.NewInt(2000)),
				GasLimit: 20000,
			},
			memo: "",
			msgs: []sdk.Msg{
				stakingtypes.NewMsgDelegate(
					testAddress,
					sdk.ValAddress(suite.createTestAddress()),
					suite.makeCoins("photon", math.NewInt(1))[0],
				),
				stakingtypes.NewMsgDelegate(
					testAddress,
					sdk.ValAddress(suite.createTestAddress()),
					suite.makeCoins("photon", math.NewInt(5))[0],
				),
			},
			accountNumber: 25,
			sequence:      78,
			expectSuccess: true,
		},
		{
			title: "Succeeds - Single-Signer MsgVote V1 with omitted value",
			fee: txtypes.Fee{
				Amount:   suite.makeCoins("aphoton", math.NewInt(2000)),
				GasLimit: 20000,
			},
			memo: "",
			msgs: []sdk.Msg{
				govtypesv1.NewMsgVote(
					testAddress,
					5,
					govtypesv1.VoteOption_VOTE_OPTION_NO,
					"",
				),
			},
			accountNumber: 25,
			sequence:      78,
			expectSuccess: true,
		},
		{
			title: "Succeeds - Single-Signer MsgSend + MsgVote",
			fee: txtypes.Fee{
				Amount:   suite.makeCoins("aphoton", math.NewInt(2000)),
				GasLimit: 20000,
			},
			memo: "",
			msgs: []sdk.Msg{
				govtypes.NewMsgVote(
					testAddress,
					5,
					govtypes.OptionNo,
				),
				banktypes.NewMsgSend(
					testAddress,
					suite.createTestAddress(),
					suite.makeCoins("photon", math.NewInt(50)),
				),
			},
			accountNumber: 25,
			sequence:      78,
			expectSuccess: !suite.useLegacyEIP712TypedData,
		},
		{
			title: "Succeeds - Single-Signer 2x MsgVoteV1 with different schemas",
			fee: txtypes.Fee{
				Amount:   suite.makeCoins("aphoton", math.NewInt(2000)),
				GasLimit: 20000,
			},
			memo: "",
			msgs: []sdk.Msg{
				govtypesv1.NewMsgVote(
					testAddress,
					5,
					govtypesv1.VoteOption_VOTE_OPTION_NO,
					"",
				),
				govtypesv1.NewMsgVote(
					testAddress,
					10,
					govtypesv1.VoteOption_VOTE_OPTION_YES,
					"Has Metadata",
				),
			},
			accountNumber: 25,
			sequence:      78,
			expectSuccess: !suite.useLegacyEIP712TypedData,
		},
		{
			title: "Fails - Two MsgVotes with Different Signers",
			fee: txtypes.Fee{
				Amount:   suite.makeCoins("aphoton", math.NewInt(2000)),
				GasLimit: 20000,
			},
			memo: "",
			msgs: []sdk.Msg{
				govtypes.NewMsgVote(
					suite.createTestAddress(),
					5,
					govtypes.OptionNo,
				),
				govtypes.NewMsgVote(
					suite.createTestAddress(),
					25,
					govtypes.OptionAbstain,
				),
			},
			accountNumber: 25,
			sequence:      78,
			expectSuccess: false,
		},
		{
			title: "Fails - Empty transaction",
			fee: txtypes.Fee{
				Amount:   suite.makeCoins("aphoton", math.NewInt(2000)),
				GasLimit: 20000,
			},
			memo:          "",
			msgs:          []sdk.Msg{},
			accountNumber: 25,
			sequence:      78,
			expectSuccess: false,
		},
		{
			title:   "Fails - Invalid ChainID",
			chainId: "invalidchainid",
			fee: txtypes.Fee{
				Amount:   suite.makeCoins("aphoton", math.NewInt(2000)),
				GasLimit: 20000,
			},
			memo: "",
			msgs: []sdk.Msg{
				govtypes.NewMsgVote(
					suite.createTestAddress(),
					5,
					govtypes.OptionNo,
				),
			},
			accountNumber: 25,
			sequence:      78,
			expectSuccess: false,
		},
		{
			title:   "Fails - Large ChainID",
			chainId: fmt.Sprintf("ethm_%v-1", big.NewInt(10).Exp(big.NewInt(10), big.NewInt(1000), nil).String()),
			fee: txtypes.Fee{
				Amount:   suite.makeCoins("aphoton", math.NewInt(2000)),
				GasLimit: 20000,
			},
			memo: "",
			msgs: []sdk.Msg{
				govtypes.NewMsgVote(
					suite.createTestAddress(),
					5,
					govtypes.OptionNo,
				),
			},
			accountNumber: 25,
			sequence:      78,
			expectSuccess: false,
		},
		{
			title: "Fails - Includes TimeoutHeight",
			fee: txtypes.Fee{
				Amount:   suite.makeCoins("aphoton", math.NewInt(2000)),
				GasLimit: 20000,
			},
			memo: "",
			msgs: []sdk.Msg{
				govtypes.NewMsgVote(
					suite.createTestAddress(),
					5,
					govtypes.OptionNo,
				),
			},
			accountNumber: 25,
			sequence:      78,
			timeoutHeight: 1000,
			expectSuccess: false,
		},
		{
			title: "Fails - Single Message / Multi-Signer",
			fee: txtypes.Fee{
				Amount:   suite.makeCoins("aphoton", math.NewInt(2000)),
				GasLimit: 20000,
			},
			memo: "",
			msgs: []sdk.Msg{
				banktypes.NewMsgMultiSend(
					[]banktypes.Input{
						banktypes.NewInput(
							suite.createTestAddress(),
							suite.makeCoins("photon", math.NewInt(50)),
						),
						banktypes.NewInput(
							suite.createTestAddress(),
							suite.makeCoins("photon", math.NewInt(50)),
						),
					},
					[]banktypes.Output{
						banktypes.NewOutput(
							suite.createTestAddress(),
							suite.makeCoins("photon", math.NewInt(50)),
						),
						banktypes.NewOutput(
							suite.createTestAddress(),
							suite.makeCoins("photon", math.NewInt(50)),
						),
					},
				),
			},
			accountNumber: 25,
			sequence:      78,
			expectSuccess: false,
		},
	}

	for _, tc := range testCases {
		for _, signMode := range signModes {
			suite.Run(tc.title, func() {
				privKey, pubKey := suite.createTestKeyPair()

				// Init tx builder
				txBuilder := suite.clientCtx.TxConfig.NewTxBuilder()

				// Set gas and fees
				txBuilder.SetGasLimit(tc.fee.GasLimit)
				txBuilder.SetFeeAmount(tc.fee.Amount)

				// Set messages
				err := txBuilder.SetMsgs(tc.msgs...)
				suite.Require().NoError(err)

				// Set memo
				txBuilder.SetMemo(tc.memo)

				// Prepare signature field
				txSigData := signing.SingleSignatureData{
					SignMode:  signMode,
					Signature: nil,
				}
				txSig := signing.SignatureV2{
					PubKey:   pubKey,
					Data:     &txSigData,
					Sequence: tc.sequence,
				}

				err = txBuilder.SetSignatures([]signing.SignatureV2{txSig}...)
				suite.Require().NoError(err)

				chainId := "ethermint_9000-1"
				if tc.chainId != "" {
					chainId = tc.chainId
				}

				if tc.timeoutHeight != 0 {
					txBuilder.SetTimeoutHeight(tc.timeoutHeight)
				}

				// Declare signerData
				signerData := authsigning.SignerData{
					ChainID:       chainId,
					AccountNumber: tc.accountNumber,
					Sequence:      tc.sequence,
					PubKey:        pubKey,
					Address:       sdk.MustBech32ifyAddressBytes("ethm", pubKey.Bytes()),
				}

				bz, err := suite.clientCtx.TxConfig.SignModeHandler().GetSignBytes(
					signMode,
					signerData,
					txBuilder.GetTx(),
				)
				suite.Require().NoError(err)

				suite.verifyEIP712SignatureVerification(tc.expectSuccess, *privKey, *pubKey, bz)

				if signMode == signing.SignMode_SIGN_MODE_LEGACY_AMINO_JSON {
					suite.verifySignDocFlattening(bz)

					if tc.expectSuccess {
						feePayer := txBuilder.GetTx().FeePayer()
						suite.sanityVerifyTypedData(bz, feePayer)
					}
				}
			})
		}
	}
}

// Verify that the payload passes signature verification if signed as its EIP-712 representation.
func (suite *EIP712TestSuite) verifyEIP712SignatureVerification(expectedSuccess bool, privKey ethsecp256k1.PrivKey, pubKey ethsecp256k1.PubKey, signBytes []byte) {
	// Convert to EIP712 bytes and sign
	eip712Bytes, err := eip712.GetEIP712BytesForMsg(signBytes)

	if suite.useLegacyEIP712TypedData {
		eip712Bytes, err = eip712.LegacyGetEIP712BytesForMsg(signBytes)
	}

	if !expectedSuccess {
		// Expect failure generating EIP-712 bytes
		suite.Require().Error(err)
		return
	}

	suite.Require().NoError(err)

	sig, err := privKey.Sign(eip712Bytes)
	suite.Require().NoError(err)

	// Verify against original payload bytes. This should pass, even though it is not
	// the original message that was signed.
	res := pubKey.VerifySignature(signBytes, sig)
	suite.Require().True(res)

	// Verify against the signed EIP-712 bytes. This should pass, since it is the message signed.
	res = pubKey.VerifySignature(eip712Bytes, sig)
	suite.Require().True(res)

	// Verify against random bytes to ensure it does not pass unexpectedly (sanity check).
	randBytes := make([]byte, len(signBytes))
	copy(randBytes, signBytes)
	// Change the first element of signBytes to a different value
	randBytes[0] = (signBytes[0] + 10) % 128
	res = pubKey.VerifySignature(randBytes, sig)
	suite.Require().False(res)
}

func (suite *EIP712TestSuite) verifySignDocFlattening(signDoc []byte) {
	payload := gjson.ParseBytes(signDoc)
	suite.Require().True(payload.IsObject())

	flattened, _, err := eip712.FlattenPayloadMessages(payload)
	suite.Require().NoError(err)

	suite.verifyPayloadAgainstFlattened(payload, flattened)
}

func (suite *EIP712TestSuite) verifyPayloadAgainstFlattened(payload gjson.Result, flattened gjson.Result) {
	payloadMap, ok := payload.Value().(map[string]interface{})
	suite.Require().True(ok)
	flattenedMap, ok := flattened.Value().(map[string]interface{})
	suite.Require().True(ok)

	suite.verifyPayloadMapAgainstFlattenedMap(payloadMap, flattenedMap)
}

// Verify that the payload matches the expected flattened version
func (suite *EIP712TestSuite) verifyPayloadMapAgainstFlattenedMap(original map[string]interface{}, flattened map[string]interface{}) {
	interfaceMessages, ok := original["msgs"]
	suite.Require().True(ok)

	messages, ok := interfaceMessages.([]interface{})
	suite.Require().True(ok)

	// Verify message contents
	for i, msg := range messages {
		flattenedMsg, ok := flattened[fmt.Sprintf("msg%d", i)]
		suite.Require().True(ok)

		flattenedMsgJSON, ok := flattenedMsg.(map[string]interface{})
		suite.Require().True(ok)

		suite.Require().True(reflect.DeepEqual(flattenedMsgJSON, msg))
	}

	// Verify new payload does not have msgs field
	_, ok = flattened["msgs"]
	suite.Require().False(ok)

	// Verify number of total keys
	numKeysOriginal := len(original)
	numKeysFlattened := len(flattened)
	numMessages := len(messages)

	// + N keys, -1 for msgs
	suite.Require().Equal(numKeysFlattened, numKeysOriginal+numMessages-1)

	// Verify contents of remaining keys
	for k, obj := range original {
		if k == "msgs" {
			continue
		}

		flattenedObj, ok := flattened[k]
		suite.Require().True(ok)

		suite.Require().Equal(obj, flattenedObj)
		suite.Require().True(reflect.DeepEqual(obj, flattenedObj))
	}
}

func (suite *EIP712TestSuite) sanityVerifyTypedData(signDoc []byte, feePayer sdk.AccAddress) {
	typedData, err := eip712.GetEIP712TypedDataForMsg(signDoc)

	suite.Require().NoError(err)

	jsonPayload := gjson.ParseBytes(signDoc)
	suite.Require().True(jsonPayload.IsObject())

	flattened, _, err := eip712.FlattenPayloadMessages(jsonPayload)
	suite.Require().NoError(err)

	// Add feePayer field
	flattenedRaw, err := sjson.Set(flattened.Raw, "fee.feePayer", feePayer.String())
	suite.Require().NoError(err)

	flattened = gjson.Parse(flattenedRaw)
	suite.Require().True(flattened.IsObject())

	originalFlattenedMsg, ok := flattened.Value().(map[string]interface{})
	suite.Require().True(ok)

	suite.Require().True(reflect.DeepEqual(typedData.Message, originalFlattenedMsg))
}

func (suite *EIP712TestSuite) TestErrorHandling() {
	// Flatten Payload:
	// No msgs
	_, _, err := eip712.FlattenPayloadMessages(gjson.Parse(""))
	suite.Require().Error(err)

	// Non-array Msgs
	_, _, err = eip712.FlattenPayloadMessages(gjson.Parse(`{"msgs": 10}`))
	suite.Require().Error(err)

	// Array with non-object items
	_, _, err = eip712.FlattenPayloadMessages(gjson.Parse(`{"msgs": [10, 20]}`))
	suite.Require().Error(err)

	// Malformed payload
	malformed, err := sjson.Set(suite.generateRandomPayload(2).Raw, "msg0", 20)
	suite.Require().NoError(err)
	_, _, err = eip712.FlattenPayloadMessages(gjson.Parse(malformed))
	suite.Require().Error(err)

	// TypedData
	// Empty JSON
	_, err = eip712.WrapTxToTypedData(0, make([]byte, 0), nil)
	suite.Require().Error(err)

	_, err = eip712.WrapTxToTypedData(0, []byte(gjson.Parse(`{"msgs": 10}`).Raw), nil)
	suite.Require().Error(err)
}
