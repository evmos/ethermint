package types

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/ethermint/crypto/ethsecp256k1"
	"github.com/cosmos/ethermint/tests"

	ethcmn "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
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
		from       string
		accessList *ethtypes.AccessList
		chainID    *big.Int
		expectPass bool
	}{
		{msg: "pass with recipient - Legacy Tx", to: &suite.to, amount: big.NewInt(100), gasPrice: big.NewInt(100000), expectPass: true},
		{msg: "pass with recipient - AccessList Tx", to: &suite.to, amount: big.NewInt(100), gasPrice: big.NewInt(0), accessList: &ethtypes.AccessList{}, chainID: big.NewInt(1), expectPass: true},
		{msg: "pass contract - Legacy Tx", to: nil, amount: big.NewInt(100), gasPrice: big.NewInt(100000), expectPass: true},
		{msg: "invalid recipient", to: &ethcmn.Address{}, amount: big.NewInt(-1), gasPrice: big.NewInt(1000), expectPass: false},
		{msg: "nil amount", to: &suite.to, amount: nil, gasPrice: big.NewInt(1000), expectPass: true},
		{msg: "negative amount", to: &suite.to, amount: big.NewInt(-1), gasPrice: big.NewInt(1000), expectPass: true},
		{msg: "nil gas price", to: &suite.to, amount: big.NewInt(100), gasPrice: nil, expectPass: false},
		{msg: "negative gas price", to: &suite.to, amount: big.NewInt(100), gasPrice: big.NewInt(-1), expectPass: true},
		{msg: "zero gas price", to: &suite.to, amount: big.NewInt(100), gasPrice: big.NewInt(0), expectPass: true},
		{msg: "invalid from address", to: &suite.to, amount: big.NewInt(100), gasPrice: big.NewInt(0), from: ethcmn.Address{}.Hex(), expectPass: false},
		{msg: "chain ID not set on AccessListTx", to: &suite.to, amount: big.NewInt(100), gasPrice: big.NewInt(0), accessList: &ethtypes.AccessList{}, chainID: nil, expectPass: false},
	}

	for i, tc := range testCases {
		msg := NewMsgEthereumTx(tc.chainID, 0, tc.to, tc.amount, 0, tc.gasPrice, nil, tc.accessList)
		msg.From = tc.from
		err := msg.ValidateBasic()

		if tc.expectPass {
			suite.Require().NoError(err, "valid test %d failed: %s, %v", i, tc.msg, msg)
		} else {
			suite.Require().Error(err, "invalid test %d passed: %s, %v", i, tc.msg, msg.Data)
		}
	}
}

func (suite *MsgsTestSuite) TestMsgEthereumTx_Sign() {
	testCases := []struct {
		msg        string
		tx         *MsgEthereumTx
		ethSigner  ethtypes.Signer
		malleate   func(tx *MsgEthereumTx)
		expectPass bool
	}{
		{
			"pass - EIP2930 signer",
			NewMsgEthereumTx(suite.chainID, 0, &suite.to, nil, 100000, nil, []byte("test"), &types.AccessList{}),
			ethtypes.NewEIP2930Signer(suite.chainID),
			func(tx *MsgEthereumTx) { tx.From = suite.from.Hex() },
			true,
		},
		{
			"pass - EIP155 signer",
			NewMsgEthereumTx(suite.chainID, 0, &suite.to, nil, 100000, nil, []byte("test"), nil),
			ethtypes.NewEIP155Signer(suite.chainID),
			func(tx *MsgEthereumTx) { tx.From = suite.from.Hex() },
			true,
		},
		{
			"pass - Homestead signer",
			NewMsgEthereumTx(suite.chainID, 0, &suite.to, nil, 100000, nil, []byte("test"), nil),
			ethtypes.HomesteadSigner{},
			func(tx *MsgEthereumTx) { tx.From = suite.from.Hex() },
			true,
		},
		{
			"pass - Frontier signer",
			NewMsgEthereumTx(suite.chainID, 0, &suite.to, nil, 100000, nil, []byte("test"), nil),
			ethtypes.FrontierSigner{},
			func(tx *MsgEthereumTx) { tx.From = suite.from.Hex() },
			true,
		},
		{
			"no from address ",
			NewMsgEthereumTx(suite.chainID, 0, &suite.to, nil, 100000, nil, []byte("test"), &types.AccessList{}),
			ethtypes.NewEIP2930Signer(suite.chainID),
			func(tx *MsgEthereumTx) { tx.From = "" },
			false,
		},
		{
			"from address ≠ signer address",
			NewMsgEthereumTx(suite.chainID, 0, &suite.to, nil, 100000, nil, []byte("test"), &types.AccessList{}),
			ethtypes.NewEIP2930Signer(suite.chainID),
			func(tx *MsgEthereumTx) { tx.From = suite.to.Hex() },
			false,
		},
	}

	for i, tc := range testCases {
		tc.malleate(tc.tx)

		err := tc.tx.Sign(tc.ethSigner, suite.signer)
		if tc.expectPass {
			suite.Require().NoError(err, "valid test %d failed: %s", i, tc.msg)

			tx := tc.tx.AsTransaction()

			sender, err := ethtypes.Sender(tc.ethSigner, tx)
			suite.Require().NoError(err, tc.msg)
			suite.Require().Equal(tc.tx.From, sender.Hex(), tc.msg)
		} else {
			suite.Require().Error(err, "invalid test %d passed: %s", i, tc.msg)
		}
	}
}
