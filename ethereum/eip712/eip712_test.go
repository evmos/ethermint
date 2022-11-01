package eip712_test

import (
	"testing"

	"cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/simapp/params"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/evmos/ethermint/ethereum/eip712"

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
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/stretchr/testify/suite"
)

// Unit tests for single-signer EIP-712 signature verification. Multi-signer verification tests are included
// in ante_test.go.

type EIP712TestSuite struct {
	suite.Suite

	config    params.EncodingConfig
	clientCtx client.Context
}

func TestEIP712TestSuite(t *testing.T) {
	suite.Run(t, &EIP712TestSuite{})
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

func (suite *EIP712TestSuite) TestEIP712SignatureVerification() {
	suite.SetupTest()

	signModes := []signing.SignMode{
		signing.SignMode_SIGN_MODE_DIRECT,
		signing.SignMode_SIGN_MODE_LEGACY_AMINO_JSON,
	}

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
			title: "Fails - Two MsgVotes",
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
			expectSuccess: false, // Multiple messages are currently not allowed
		},
		{
			title: "Fails - MsgSend + MsgVote",
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
				banktypes.NewMsgSend(
					suite.createTestAddress(),
					suite.createTestAddress(),
					suite.makeCoins("photon", math.NewInt(50)),
				),
			},
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
			})
		}
	}
}

// Verify that the payload passes signature verification if signed as its EIP-712 representation.
func (suite *EIP712TestSuite) verifyEIP712SignatureVerification(expectedSuccess bool, privKey ethsecp256k1.PrivKey, pubKey ethsecp256k1.PubKey, signBytes []byte) {
	// Convert to EIP712 hash and sign
	eip712Hash, err := eip712.GetEIP712HashForMsg(signBytes)
	if !expectedSuccess {
		// Expect failure generating EIP-712 hash
		suite.Require().Error(err)
		return
	}

	suite.Require().NoError(err)

	sigHash := crypto.Keccak256Hash(eip712Hash)
	sig, err := privKey.Sign(sigHash.Bytes())
	suite.Require().NoError(err)

	// Verify against original payload bytes. This should pass, even though it is not
	// the original message that was signed.
	res := pubKey.VerifySignature(signBytes, sig)
	suite.Require().True(res)

	// Verify against the signed EIP-712 bytes. This should pass, since it is the message signed.
	res = pubKey.VerifySignature(eip712Hash, sig)
	suite.Require().True(res)

	// Verify against random bytes to ensure it does not pass unexpectedly (sanity check).
	randBytes := make([]byte, len(signBytes))
	copy(randBytes, signBytes)
	// Change the first element of signBytes to a different value
	randBytes[0] = (signBytes[0] + 10) % 128
	res = pubKey.VerifySignature(randBytes, sig)
	suite.Require().False(res)
}
