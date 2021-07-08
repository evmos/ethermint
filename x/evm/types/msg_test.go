package types

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tharsis/ethermint/crypto/ethsecp256k1"
	"github.com/tharsis/ethermint/tests"

	"github.com/ethereum/go-ethereum/common"
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
	msg := NewTx(nil, 0, &suite.to, nil, 100000, nil, []byte("test"), nil)

	// suite.Require().Equal(msg.Data.To, suite.to.Hex())
	suite.Require().Equal(msg.Route(), RouterKey)
	suite.Require().Equal(msg.Type(), TypeMsgEthereumTx)
	// suite.Require().NotNil(msg.To())
	suite.Require().Equal(msg.GetMsgs(), []sdk.Msg{msg})
	suite.Require().Panics(func() { msg.GetSigners() })
	suite.Require().Panics(func() { msg.GetSignBytes() })

	msg = NewTxContract(nil, 0, nil, 100000, nil, []byte("test"), nil)
	suite.Require().NotNil(msg)
	// suite.Require().Empty(msg.Data.To)
	// suite.Require().Nil(msg.To())
}

func (suite *MsgsTestSuite) TestMsgEthereumTx_ValidateBasic() {
	hundredInt := sdk.NewInt(100)
	zeroInt := sdk.ZeroInt()
	minusOneInt := sdk.NewInt(-1)

	testCases := []struct {
		msg        string
		to         string
		amount     *sdk.Int
		gasPrice   *sdk.Int
		from       string
		accessList *ethtypes.AccessList
		chainID    *sdk.Int
		expectPass bool
	}{
		{msg: "pass with recipient - Legacy Tx", to: suite.to.Hex(), amount: &hundredInt, gasPrice: &hundredInt, expectPass: true},
		{msg: "pass with recipient - AccessList Tx", to: suite.to.Hex(), amount: &hundredInt, gasPrice: &zeroInt, accessList: &ethtypes.AccessList{}, chainID: &hundredInt, expectPass: true},
		{msg: "pass contract - Legacy Tx", to: "", amount: &hundredInt, gasPrice: &hundredInt, expectPass: true},
		// {msg: "invalid recipient", to: invalidFromAddress, amount: &minusOneInt, gasPrice: &hundredInt, expectPass: false},
		{msg: "nil amount - Legacy Tx", to: suite.to.Hex(), amount: nil, gasPrice: &hundredInt, expectPass: true},
		{msg: "negative amount - Legacy Tx", to: suite.to.Hex(), amount: &minusOneInt, gasPrice: &hundredInt, expectPass: false},
		{msg: "nil gas price - Legacy Tx", to: suite.to.Hex(), amount: &hundredInt, gasPrice: nil, expectPass: false},
		{msg: "negative gas price - Legacy Tx", to: suite.to.Hex(), amount: &hundredInt, gasPrice: &minusOneInt, expectPass: false},
		{msg: "zero gas price - Legacy Tx", to: suite.to.Hex(), amount: &hundredInt, gasPrice: &zeroInt, expectPass: true},
		{msg: "invalid from address - Legacy Tx", to: suite.to.Hex(), amount: &hundredInt, gasPrice: &zeroInt, from: invalidFromAddress, expectPass: false},
		{msg: "nil amount - AccessListTx", to: suite.to.Hex(), amount: nil, gasPrice: &hundredInt, accessList: &ethtypes.AccessList{}, chainID: &hundredInt, expectPass: true},
		{msg: "negative amount - AccessListTx", to: suite.to.Hex(), amount: &minusOneInt, gasPrice: &hundredInt, accessList: &ethtypes.AccessList{}, chainID: nil, expectPass: false},
		{msg: "nil gas price - AccessListTx", to: suite.to.Hex(), amount: &hundredInt, gasPrice: nil, accessList: &ethtypes.AccessList{}, chainID: &hundredInt, expectPass: false},
		{msg: "negative gas price - AccessListTx", to: suite.to.Hex(), amount: &hundredInt, gasPrice: &minusOneInt, accessList: &ethtypes.AccessList{}, chainID: nil, expectPass: false},
		{msg: "zero gas price - AccessListTx", to: suite.to.Hex(), amount: &hundredInt, gasPrice: &zeroInt, accessList: &ethtypes.AccessList{}, chainID: &hundredInt, expectPass: true},
		{msg: "invalid from address - AccessListTx", to: suite.to.Hex(), amount: &hundredInt, gasPrice: &zeroInt, from: invalidFromAddress, accessList: &ethtypes.AccessList{}, chainID: &hundredInt, expectPass: false},
		{msg: "chain ID not set on AccessListTx", to: suite.to.Hex(), amount: &hundredInt, gasPrice: &zeroInt, accessList: &ethtypes.AccessList{}, chainID: nil, expectPass: false},
	}

	for i, tc := range testCases {
		to := common.HexToAddress(tc.from)

		var chainID, amount, gasPrice *big.Int
		if tc.chainID != nil {
			chainID = tc.chainID.BigInt()
		}
		if tc.amount != nil {
			amount = tc.amount.BigInt()
		}
		if tc.gasPrice != nil {
			gasPrice = tc.gasPrice.BigInt()
		}

		tx := NewTx(chainID, 1, &to, amount, 1000, gasPrice, nil, tc.accessList)
		tx.From = tc.from

		err := tx.ValidateBasic()

		if tc.expectPass {
			suite.Require().NoError(err, "valid test %d failed: %s, %v", i, tc.msg)
		} else {
			suite.Require().Error(err, "invalid test %d passed: %s, %v", i, tc.msg)
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
			NewTx(suite.chainID, 0, &suite.to, nil, 100000, nil, []byte("test"), &types.AccessList{}),
			ethtypes.NewEIP2930Signer(suite.chainID),
			func(tx *MsgEthereumTx) { tx.From = suite.from.Hex() },
			true,
		},
		{
			"pass - EIP155 signer",
			NewTx(suite.chainID, 0, &suite.to, nil, 100000, nil, []byte("test"), nil),
			ethtypes.NewEIP155Signer(suite.chainID),
			func(tx *MsgEthereumTx) { tx.From = suite.from.Hex() },
			true,
		},
		{
			"pass - Homestead signer",
			NewTx(suite.chainID, 0, &suite.to, nil, 100000, nil, []byte("test"), nil),
			ethtypes.HomesteadSigner{},
			func(tx *MsgEthereumTx) { tx.From = suite.from.Hex() },
			true,
		},
		{
			"pass - Frontier signer",
			NewTx(suite.chainID, 0, &suite.to, nil, 100000, nil, []byte("test"), nil),
			ethtypes.FrontierSigner{},
			func(tx *MsgEthereumTx) { tx.From = suite.from.Hex() },
			true,
		},
		{
			"no from address ",
			NewTx(suite.chainID, 0, &suite.to, nil, 100000, nil, []byte("test"), &types.AccessList{}),
			ethtypes.NewEIP2930Signer(suite.chainID),
			func(tx *MsgEthereumTx) { tx.From = "" },
			false,
		},
		{
			"from address ≠ signer address",
			NewTx(suite.chainID, 0, &suite.to, nil, 100000, nil, []byte("test"), &types.AccessList{}),
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

			sender, err := tc.tx.GetSender(suite.chainID)
			suite.Require().NoError(err, tc.msg)
			suite.Require().Equal(tc.tx.From, sender.Hex(), tc.msg)
		} else {
			suite.Require().Error(err, "invalid test %d passed: %s", i, tc.msg)
		}
	}
}
