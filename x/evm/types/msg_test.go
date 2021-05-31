package types

import (
	"bytes"
	"math/big"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/ethermint/crypto/ethsecp256k1"
	"github.com/cosmos/ethermint/tests"

	ethcmn "github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/rlp"
)

type MsgsTestSuite struct {
	suite.Suite

	signer  keyring.Signer
	from    ethcmn.Address
	to      ethcmn.Address
	chainID *big.Int
}

func TestMsgsTestSuite(t *testing.T) {
	suite.Run(t, new(MsgsTestSuite))
}

func (suite *MsgsTestSuite) SetupTest() {
	privFrom, err := ethsecp256k1.GenerateKey()
	suite.Require().NoError(err)

	privTo, err := ethsecp256k1.GenerateKey()
	suite.Require().NoError(err)

	suite.signer = tests.NewSigner(privFrom)
	suite.from = crypto.PubkeyToAddress(privFrom.ToECDSA().PublicKey)
	suite.to = crypto.PubkeyToAddress(privTo.ToECDSA().PublicKey)
	suite.chainID = big.NewInt(1)
}

func (suite *MsgsTestSuite) TestMsgEthereumTx_Constructor() {
	msg := NewMsgEthereumTx(nil, 0, &suite.to, nil, 100000, nil, []byte("test"), nil)

	suite.Require().Equal(msg.Data.To, suite.to.Hex())
	suite.Require().Equal(msg.Route(), RouterKey)
	suite.Require().Equal(msg.Type(), TypeMsgEthereumTx)
	suite.Require().NotNil(msg.To())
	suite.Require().Equal(msg.GetMsgs(), []sdk.Msg{msg})
	suite.Require().Panics(func() { msg.GetSigners() })
	suite.Require().Panics(func() { msg.GetSignBytes() })

	msg = NewMsgEthereumTxContract(nil, 0, nil, 100000, nil, []byte("test"), nil)
	suite.Require().NotNil(msg)
	suite.Require().Empty(msg.Data.To)
	suite.Require().Nil(msg.To())
}

func (suite *MsgsTestSuite) TestMsgEthereumTx_ValidateBasic() {
	testCases := []struct {
		msg        string
		to         *ethcmn.Address
		amount     *big.Int
		gasPrice   *big.Int
		expectPass bool
	}{
		{msg: "pass with recipient", to: &suite.to, amount: big.NewInt(100), gasPrice: big.NewInt(100000), expectPass: true},
		{msg: "pass contract", to: nil, amount: big.NewInt(100), gasPrice: big.NewInt(100000), expectPass: true},
		{msg: "invalid recipient", to: &ethcmn.Address{}, amount: big.NewInt(-1), gasPrice: big.NewInt(1000), expectPass: false},
		// NOTE: these can't be effectively tested because the SetBytes function from big.Int only sets
		// the absolute value
		{msg: "negative amount", to: &suite.to, amount: big.NewInt(-1), gasPrice: big.NewInt(1000), expectPass: true},
		{msg: "negative gas price", to: &suite.to, amount: big.NewInt(100), gasPrice: big.NewInt(-1), expectPass: true},
		{msg: "zero gas price", to: &suite.to, amount: big.NewInt(100), gasPrice: big.NewInt(0), expectPass: true},
	}

	for i, tc := range testCases {
		msg := NewMsgEthereumTx(suite.chainID, 0, tc.to, tc.amount, 0, tc.gasPrice, nil, nil)
		err := msg.ValidateBasic()

		if tc.expectPass {
			suite.Require().NoError(err, "valid test %d failed: %s", i, tc.msg)
		} else {
			suite.Require().Error(err, "invalid test %d passed: %s", i, tc.msg)
		}
	}
}

func (suite *MsgsTestSuite) TestMsgEthereumTx_EncodeRLP() {
	expMsg := NewMsgEthereumTx(suite.chainID, 0, &suite.to, nil, 100000, nil, []byte("test"), nil)

	raw, err := rlp.EncodeToBytes(&expMsg)
	suite.Require().NoError(err)

	msg := &MsgEthereumTx{}
	err = rlp.Decode(bytes.NewReader(raw), &msg)
	suite.Require().NoError(err)
	suite.Require().Equal(expMsg.Data, msg.Data)
}

func (suite *MsgsTestSuite) TestMsgEthereumTx_RLPSignBytes() {
	msg := NewMsgEthereumTx(suite.chainID, 0, &suite.to, nil, 100000, nil, []byte("test"), nil)
	suite.NotPanics(func() { _ = msg.RLPSignBytes(suite.chainID) })
}

func (suite *MsgsTestSuite) TestMsgEthereumTx_Sign() {
	msg := NewMsgEthereumTx(suite.chainID, 0, &suite.to, nil, 100000, nil, []byte("test"), nil)

	testCases := []struct {
		msg        string
		malleate   func()
		expectPass bool
	}{
		{
			"pass",
			func() { msg.From = suite.from.Hex() },
			true,
		},
		{
			"no from address ",
			func() { msg.From = "" },
			false,
		},
		{
			"from address ≠ signer address",
			func() { msg.From = suite.to.Hex() },
			false,
		},
	}

	for i, tc := range testCases {
		tc.malleate()
		err := msg.Sign(suite.chainID, suite.signer)
		if tc.expectPass {
			suite.Require().NoError(err, "valid test %d failed: %s", i, tc.msg)

			tx := msg.AsTransaction()
			signer := ethtypes.NewEIP2930Signer(suite.chainID)

			sender, err := ethtypes.Sender(signer, tx)
			suite.Require().NoError(err, tc.msg)
			suite.Require().Equal(msg.From, sender.Hex(), tc.msg)
		} else {
			suite.Require().Error(err, "invalid test %d passed: %s", i, tc.msg)
		}
	}
}
