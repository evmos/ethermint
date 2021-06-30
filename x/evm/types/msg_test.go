package types

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tharsis/ethermint/crypto/ethsecp256k1"
	"github.com/tharsis/ethermint/tests"

	ethcmn "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
)

const invalidFromAddress = "0x0000"

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
		to         string
		amount     sdk.Int
		gasPrice   sdk.Int
		from       string
		accessList *ethtypes.AccessList
		chainID    sdk.Int
		expectPass bool
	}{
		{msg: "pass with recipient - Legacy Tx", to: suite.to.Hex(), amount: sdk.NewInt(100), gasPrice: sdk.NewInt(100000), expectPass: true},
		{msg: "pass with recipient - AccessList Tx", to: suite.to.Hex(), amount: sdk.NewInt(100), gasPrice: sdk.ZeroInt(), accessList: &ethtypes.AccessList{}, chainID: sdk.OneInt(), expectPass: true},
		{msg: "pass contract - Legacy Tx", to: "", amount: sdk.NewInt(100), gasPrice: sdk.NewInt(100000), expectPass: true},
		{msg: "invalid recipient", to: invalidFromAddress, amount: sdk.NewInt(-1), gasPrice: sdk.NewInt(1000), expectPass: false},
		{msg: "nil amount", to: suite.to.Hex(), amount: sdk.Int{}, gasPrice: sdk.NewInt(1000), expectPass: true},
		{msg: "negative amount", to: suite.to.Hex(), amount: sdk.NewInt(-1), gasPrice: sdk.NewInt(1000), expectPass: false},
		{msg: "nil gas price", to: suite.to.Hex(), amount: sdk.NewInt(100), gasPrice: sdk.Int{}, expectPass: false},
		{msg: "negative gas price", to: suite.to.Hex(), amount: sdk.NewInt(100), gasPrice: sdk.NewInt(-1), expectPass: false},
		{msg: "zero gas price", to: suite.to.Hex(), amount: sdk.NewInt(100), gasPrice: sdk.ZeroInt(), expectPass: true},
		{msg: "invalid from address", to: suite.to.Hex(), amount: sdk.NewInt(100), gasPrice: sdk.ZeroInt(), from: invalidFromAddress, expectPass: false},
		{msg: "chain ID not set on AccessListTx", to: suite.to.Hex(), amount: sdk.NewInt(100), gasPrice: sdk.ZeroInt(), accessList: &ethtypes.AccessList{}, chainID: sdk.Int{}, expectPass: false},
	}

	for i, tc := range testCases {
		// recreate txData
		txData := TxData{
			Nonce:    0,
			GasLimit: 0,
			To:       tc.to,
		}

		if tc.accessList != nil {
			txData.Accesses = NewAccessList(tc.accessList)
			if !tc.chainID.IsNil() {
				txData.ChainID = tc.chainID
			}
		}

		if !tc.amount.IsNil() {
			txData.Amount = tc.amount
		}

		if !tc.gasPrice.IsNil() {
			txData.GasPrice = tc.gasPrice
		}

		msg := MsgEthereumTx{
			Data: &txData,
		}

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
