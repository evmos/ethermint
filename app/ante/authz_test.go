package ante_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	sdkvesting "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/evmos/ethermint/testutil"
	utiltx "github.com/evmos/ethermint/testutil/tx"

	"github.com/evmos/ethermint/app/ante"

	"github.com/evmos/ethermint/app"
	"github.com/evmos/ethermint/crypto/ethsecp256k1"
	"github.com/evmos/ethermint/encoding"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	evmtypes "github.com/evmos/ethermint/x/evm/types"
)

func generatePrivKeyAddressPairs(accCount int) ([]*ethsecp256k1.PrivKey, []sdk.AccAddress, error) {
	var (
		err           error
		testPrivKeys  = make([]*ethsecp256k1.PrivKey, accCount)
		testAddresses = make([]sdk.AccAddress, accCount)
	)

	for i := range testPrivKeys {
		testPrivKeys[i], err = ethsecp256k1.GenerateKey()
		if err != nil {
			return nil, nil, err
		}
		testAddresses[i] = testPrivKeys[i].PubKey().Address().Bytes()
	}
	return testPrivKeys, testAddresses, nil
}

func newMsgExec(grantee sdk.AccAddress, msgs []sdk.Msg) *authz.MsgExec {
	msg := authz.NewMsgExec(grantee, msgs)
	return &msg
}

func createNestedExecMsgSend(testAddresses []sdk.AccAddress, depth int) *authz.MsgExec {
	return createNestedMsgExec(
		testAddresses[1],
		depth,
		[]sdk.Msg{
			createMsgSend(testAddresses),
		},
	)
}

func createMsgSend(testAddresses []sdk.AccAddress) *banktypes.MsgSend {
	return banktypes.NewMsgSend(
		testAddresses[0],
		testAddresses[3],
		sdk.NewCoins(sdk.NewInt64Coin(evmtypes.DefaultEVMDenom, 1e8)),
	)
}

func newMsgGrant(testAddresses []sdk.AccAddress, auth authz.Authorization) *authz.MsgGrant {
	expiration := time.Date(9000, 1, 1, 0, 0, 0, 0, time.UTC)
	msg, err := authz.NewMsgGrant(testAddresses[0], testAddresses[1], auth, &expiration)
	if err != nil {
		panic(err)
	}
	return msg
}

func newGenericMsgGrant(testAddresses []sdk.AccAddress, typeUrl string) *authz.MsgGrant {
	auth := authz.NewGenericAuthorization(typeUrl)
	return newMsgGrant(testAddresses, auth)
}

func createNestedMsgExec(grantee sdk.AccAddress, numLevels int, msgsToExec []sdk.Msg) *authz.MsgExec {
	msgs := make([]*authz.MsgExec, numLevels)
	for i := range msgs {
		if i == 0 {
			msgs[i] = newMsgExec(grantee, msgsToExec)
			continue
		}
		msgs[i] = newMsgExec(grantee, []sdk.Msg{msgs[i-1]})
	}
	return msgs[numLevels-1]
}

func TestAuthzLimiterDecorator(t *testing.T) {
	testPrivKeys, testAddresses, err := generatePrivKeyAddressPairs(5)
	require.NoError(t, err)

	validator := sdk.ValAddress(testAddresses[4])
	stakingAuthDelegate, err := stakingtypes.NewStakeAuthorization([]sdk.ValAddress{validator}, nil, stakingtypes.AuthorizationType_AUTHORIZATION_TYPE_DELEGATE, nil)
	require.NoError(t, err)

	stakingAuthUndelegate, err := stakingtypes.NewStakeAuthorization([]sdk.ValAddress{validator}, nil, stakingtypes.AuthorizationType_AUTHORIZATION_TYPE_UNDELEGATE, nil)
	require.NoError(t, err)

	decorator := ante.NewAuthzLimiterDecorator(
		sdk.MsgTypeURL(&evmtypes.MsgEthereumTx{}),
		sdk.MsgTypeURL(&stakingtypes.MsgUndelegate{}),
	)

	testMsgSend := createMsgSend(testAddresses)
	testMsgEthereumTx := &evmtypes.MsgEthereumTx{}

	testCases := []struct {
		name        string
		msgs        []sdk.Msg
		checkTx     bool
		expectedErr error
	}{
		{
			"enabled msg - non blocked msg",
			[]sdk.Msg{
				testMsgSend,
			},
			false,
			nil,
		},
		{
			"enabled msg MsgEthereumTx - blocked msg not wrapped in MsgExec",
			[]sdk.Msg{
				testMsgEthereumTx,
			},
			false,
			nil,
		},
		{
			"enabled msg - blocked msg not wrapped in MsgExec",
			[]sdk.Msg{
				&stakingtypes.MsgCancelUnbondingDelegation{},
			},
			false,
			nil,
		},
		{
			"enabled msg - MsgGrant contains a non blocked msg",
			[]sdk.Msg{
				newGenericMsgGrant(
					testAddresses,
					sdk.MsgTypeURL(&banktypes.MsgSend{}),
				),
			},
			false,
			nil,
		},
		{
			"enabled msg - MsgGrant contains a non blocked msg",
			[]sdk.Msg{
				newMsgGrant(
					testAddresses,
					stakingAuthDelegate,
				),
			},
			false,
			nil,
		},
		{
			"disabled msg - MsgGrant contains a blocked msg",
			[]sdk.Msg{
				newGenericMsgGrant(
					testAddresses,
					sdk.MsgTypeURL(&evmtypes.MsgEthereumTx{}),
				),
			},
			false,
			sdkerrors.ErrUnauthorized,
		},
		{
			"disabled msg - MsgGrant contains a blocked msg",
			[]sdk.Msg{
				newMsgGrant(
					testAddresses,
					stakingAuthUndelegate,
				),
			},
			false,
			sdkerrors.ErrUnauthorized,
		},
		{
			"allowed msg - when a MsgExec contains a non blocked msg",
			[]sdk.Msg{
				newMsgExec(
					testAddresses[1],
					[]sdk.Msg{
						testMsgSend,
					}),
			},
			false,
			nil,
		},
		{
			"disabled msg - MsgExec contains a blocked msg",
			[]sdk.Msg{
				newMsgExec(
					testAddresses[1],
					[]sdk.Msg{
						testMsgEthereumTx,
					},
				),
			},
			false,
			sdkerrors.ErrUnauthorized,
		},
		{
			"disabled msg - surrounded by valid msgs",
			[]sdk.Msg{
				newMsgGrant(
					testAddresses,
					stakingAuthDelegate,
				),
				newMsgExec(
					testAddresses[1],
					[]sdk.Msg{
						testMsgSend,
						testMsgEthereumTx,
					},
				),
			},
			false,
			sdkerrors.ErrUnauthorized,
		},
		{
			"disabled msg - nested MsgExec containing a blocked msg",
			[]sdk.Msg{
				createNestedMsgExec(
					testAddresses[1],
					2,
					[]sdk.Msg{
						testMsgEthereumTx,
					},
				),
			},
			false,
			sdkerrors.ErrUnauthorized,
		},
		{
			"disabled msg - nested MsgGrant containing a blocked msg",
			[]sdk.Msg{
				newMsgExec(
					testAddresses[1],
					[]sdk.Msg{
						newGenericMsgGrant(
							testAddresses,
							sdk.MsgTypeURL(&evmtypes.MsgEthereumTx{}),
						),
					},
				),
			},
			false,
			sdkerrors.ErrUnauthorized,
		},
		{
			"disabled msg - nested MsgExec NOT containing a blocked msg but has more nesting levels than the allowed",
			[]sdk.Msg{
				createNestedExecMsgSend(testAddresses, 6),
			},
			false,
			sdkerrors.ErrUnauthorized,
		},
		{
			"disabled msg - multiple two nested MsgExec messages NOT containing a blocked msg over the limit",
			[]sdk.Msg{
				createNestedExecMsgSend(testAddresses, 5),
				createNestedExecMsgSend(testAddresses, 5),
			},
			false,
			sdkerrors.ErrUnauthorized,
		},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Case %s", tc.name), func(t *testing.T) {
			ctx := sdk.Context{}.WithIsCheckTx(tc.checkTx)
			tx, err := createTx(testPrivKeys[0], tc.msgs...)
			require.NoError(t, err)

			_, err = decorator.AnteHandle(ctx, tx, false, NextFn)
			if tc.expectedErr != nil {
				require.Error(t, err)
				require.ErrorIs(t, err, tc.expectedErr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

var chainID = testutil.TestnetChainID + "-1"

func createTx(priv cryptotypes.PrivKey, msgs ...sdk.Msg) (sdk.Tx, error) {
	encodingConfig := encoding.MakeConfig(app.ModuleBasics)
	txBuilder := encodingConfig.TxConfig.NewTxBuilder()

	txBuilder.SetGasLimit(1000000)
	if err := txBuilder.SetMsgs(msgs...); err != nil {
		return nil, err
	}

	// First round: we gather all the signer infos. We use the "set empty
	// signature" hack to do that.
	sigV2 := signing.SignatureV2{
		PubKey: priv.PubKey(),
		Data: &signing.SingleSignatureData{
			SignMode:  encodingConfig.TxConfig.SignModeHandler().DefaultMode(),
			Signature: nil,
		},
		Sequence: 0,
	}

	sigsV2 := []signing.SignatureV2{sigV2}

	if err := txBuilder.SetSignatures(sigsV2...); err != nil {
		return nil, err
	}

	signerData := authsigning.SignerData{
		ChainID:       chainID,
		AccountNumber: 0,
		Sequence:      0,
	}
	sigV2, err := tx.SignWithPrivKey(
		encodingConfig.TxConfig.SignModeHandler().DefaultMode(), signerData,
		txBuilder, priv, encodingConfig.TxConfig,
		0,
	)
	if err != nil {
		return nil, err
	}

	sigsV2 = []signing.SignatureV2{sigV2}
	err = txBuilder.SetSignatures(sigsV2...)
	if err != nil {
		return nil, err
	}

	return txBuilder.GetTx(), nil
}

func (suite *AnteTestSuite) TestRejectMsgsInAuthz() {
	_, testAddresses, err := generatePrivKeyAddressPairs(10)
	suite.Require().NoError(err)

	testcases := []struct {
		name         string
		msgs         []sdk.Msg
		expectedCode uint32
		isEIP712     bool
	}{
		{
			name: "a MsgGrant with MsgEthereumTx typeURL on the authorization field is blocked",
			msgs: []sdk.Msg{
				newGenericMsgGrant(
					testAddresses,
					sdk.MsgTypeURL(&evmtypes.MsgEthereumTx{}),
				),
			},
			expectedCode: sdkerrors.ErrUnauthorized.ABCICode(),
		},
		{
			name: "a MsgGrant with MsgCreateVestingAccount typeURL on the authorization field is blocked",
			msgs: []sdk.Msg{
				newGenericMsgGrant(
					testAddresses,
					sdk.MsgTypeURL(&sdkvesting.MsgCreateVestingAccount{}),
				),
			},
			expectedCode: sdkerrors.ErrUnauthorized.ABCICode(),
		},
		{
			name: "a MsgGrant with MsgEthereumTx typeURL on the authorization field included on EIP712 tx is blocked",
			msgs: []sdk.Msg{
				newGenericMsgGrant(
					testAddresses,
					sdk.MsgTypeURL(&evmtypes.MsgEthereumTx{}),
				),
			},
			expectedCode: sdkerrors.ErrUnauthorized.ABCICode(),
			isEIP712:     true,
		},
		{
			name: "a MsgExec with nested messages (valid: MsgSend and invalid: MsgEthereumTx) is blocked",
			msgs: []sdk.Msg{
				newMsgExec(
					testAddresses[1],
					[]sdk.Msg{
						createMsgSend(testAddresses),
						&evmtypes.MsgEthereumTx{},
					},
				),
			},
			expectedCode: sdkerrors.ErrUnauthorized.ABCICode(),
		},
		{
			name: "a MsgExec with nested MsgExec messages that has invalid messages is blocked",
			msgs: []sdk.Msg{
				createNestedMsgExec(
					testAddresses[1],
					2,
					[]sdk.Msg{
						&evmtypes.MsgEthereumTx{},
					},
				),
			},
			expectedCode: sdkerrors.ErrUnauthorized.ABCICode(),
		},
		{
			name: "a MsgExec with more nested MsgExec messages than allowed and with valid messages is blocked",
			msgs: []sdk.Msg{
				createNestedExecMsgSend(testAddresses, 6),
			},
			expectedCode: sdkerrors.ErrUnauthorized.ABCICode(),
		},
		{
			name: "two MsgExec messages NOT containing a blocked msg but between the two have more nesting than the allowed. Then, is blocked",
			msgs: []sdk.Msg{
				createNestedExecMsgSend(testAddresses, 5),
				createNestedExecMsgSend(testAddresses, 5),
			},
			expectedCode: sdkerrors.ErrUnauthorized.ABCICode(),
		},
	}

	for _, tc := range testcases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.SetupTest()
			var (
				tx  sdk.Tx
				err error
			)

			if tc.isEIP712 {
				coinAmount := sdk.NewCoin(evmtypes.DefaultEVMDenom, sdk.NewInt(20))
				fees := sdk.NewCoins(coinAmount)
				cosmosTxArgs := utiltx.CosmosTxArgs{
					TxCfg:   suite.clientCtx.TxConfig,
					Priv:    suite.priv,
					ChainID: suite.ctx.ChainID(),
					Gas:     200000,
					Fees:    fees,
					Msgs:    tc.msgs,
				}

				tx, err = utiltx.CreateEIP712CosmosTx(
					suite.ctx,
					suite.app,
					utiltx.EIP712TxArgs{
						CosmosTxArgs:       cosmosTxArgs,
						UseLegacyExtension: true,
					},
				)
			} else {
				tx, err = createTx(suite.priv, tc.msgs...)
			}
			suite.Require().NoError(err)

			txEncoder := suite.clientCtx.TxConfig.TxEncoder()
			bz, err := txEncoder(tx)
			suite.Require().NoError(err)

			resCheckTx := suite.app.CheckTx(
				abci.RequestCheckTx{
					Tx:   bz,
					Type: abci.CheckTxType_New,
				},
			)
			suite.Require().Equal(resCheckTx.Code, tc.expectedCode, resCheckTx.Log)

			resDeliverTx := suite.app.DeliverTx(
				abci.RequestDeliverTx{
					Tx: bz,
				},
			)
			suite.Require().Equal(resDeliverTx.Code, tc.expectedCode, resDeliverTx.Log)
		})
	}
}
