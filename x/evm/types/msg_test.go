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
	"github.com/ethereum/go-ethereum/core/types"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
)

const invalidFromAddress = "0x0000"

type MsgsTestSuite struct {
	suite.Suite

	signer  keyring.Signer
	from    common.Address
	to      common.Address
	chainID *big.Int
}

func TestMsgsTestSuite(t *testing.T) {
	suite.Run(t, new(MsgsTestSuite))
}

func (suite *MsgsTestSuite) SetupTest() {
	from, privFrom := tests.NewAddrKey()

	suite.signer = tests.NewSigner(privFrom)
	suite.from = from
	suite.to = tests.GenerateAddress()
	suite.chainID = big.NewInt(1)
}

func (suite *MsgsTestSuite) TestMsgEthereumTx_Constructor() {
	msg := NewTx(nil, 0, &suite.to, nil, 100000, nil, nil, nil, []byte("test"), nil)

	// suite.Require().Equal(msg.Data.To, suite.to.Hex())
	suite.Require().Equal(msg.Route(), RouterKey)
	suite.Require().Equal(msg.Type(), TypeMsgEthereumTx)
	// suite.Require().NotNil(msg.To())
	suite.Require().Equal(msg.GetMsgs(), []sdk.Msg{msg})
	suite.Require().Panics(func() { msg.GetSigners() })
	suite.Require().Panics(func() { msg.GetSignBytes() })

	msg = NewTxContract(nil, 0, nil, 100000, nil, nil, nil, []byte("test"), nil)
	suite.Require().NotNil(msg)
	// suite.Require().Empty(msg.Data.To)
	// suite.Require().Nil(msg.To())
}

func (suite *MsgsTestSuite) TestMsgEthereumTx_ValidateBasic() {
	hundredInt := sdk.NewInt(100)
	zeroInt := sdk.ZeroInt()
	minusOneInt := sdk.NewInt(-1)
	exp_2_255 := sdk.NewIntFromBigInt(new(big.Int).Exp(big.NewInt(2), big.NewInt(255), nil))

	testCases := []struct {
		msg        string
		to         string
		amount     *sdk.Int
		gasPrice   *sdk.Int
		from       string
		accessList *types.AccessList
		chainID    *sdk.Int
		expectPass bool
	}{
		{msg: "pass with recipient - Legacy Tx", to: suite.to.Hex(), amount: &hundredInt, gasPrice: &hundredInt, expectPass: true},
		{msg: "pass with recipient - AccessList Tx", to: suite.to.Hex(), amount: &hundredInt, gasPrice: &zeroInt, accessList: &types.AccessList{}, chainID: &hundredInt, expectPass: true},
		{msg: "pass contract - Legacy Tx", to: "", amount: &hundredInt, gasPrice: &hundredInt, expectPass: true},
		// {msg: "invalid recipient", to: invalidFromAddress, amount: &minusOneInt, gasPrice: &hundredInt, expectPass: false},
		{msg: "nil amount - Legacy Tx", to: suite.to.Hex(), amount: nil, gasPrice: &hundredInt, expectPass: true},
		{msg: "negative amount - Legacy Tx", to: suite.to.Hex(), amount: &minusOneInt, gasPrice: &hundredInt, expectPass: false},
		{msg: "nil gas price - Legacy Tx", to: suite.to.Hex(), amount: &hundredInt, gasPrice: nil, expectPass: false},
		{msg: "negative gas price - Legacy Tx", to: suite.to.Hex(), amount: &hundredInt, gasPrice: &minusOneInt, expectPass: false},
		{msg: "zero gas price - Legacy Tx", to: suite.to.Hex(), amount: &hundredInt, gasPrice: &zeroInt, expectPass: true},
		{msg: "invalid from address - Legacy Tx", to: suite.to.Hex(), amount: &hundredInt, gasPrice: &zeroInt, from: invalidFromAddress, expectPass: false},
		{msg: "out of bound gas fee - Legacy Tx", to: suite.to.Hex(), amount: &hundredInt, gasPrice: &exp_2_255, expectPass: false},
		{msg: "nil amount - AccessListTx", to: suite.to.Hex(), amount: nil, gasPrice: &hundredInt, accessList: &types.AccessList{}, chainID: &hundredInt, expectPass: true},
		{msg: "negative amount - AccessListTx", to: suite.to.Hex(), amount: &minusOneInt, gasPrice: &hundredInt, accessList: &types.AccessList{}, chainID: nil, expectPass: false},
		{msg: "nil gas price - AccessListTx", to: suite.to.Hex(), amount: &hundredInt, gasPrice: nil, accessList: &types.AccessList{}, chainID: &hundredInt, expectPass: false},
		{msg: "negative gas price - AccessListTx", to: suite.to.Hex(), amount: &hundredInt, gasPrice: &minusOneInt, accessList: &types.AccessList{}, chainID: nil, expectPass: false},
		{msg: "zero gas price - AccessListTx", to: suite.to.Hex(), amount: &hundredInt, gasPrice: &zeroInt, accessList: &types.AccessList{}, chainID: &hundredInt, expectPass: true},
		{msg: "invalid from address - AccessListTx", to: suite.to.Hex(), amount: &hundredInt, gasPrice: &zeroInt, from: invalidFromAddress, accessList: &types.AccessList{}, chainID: &hundredInt, expectPass: false},
		{msg: "chain ID not set on AccessListTx", to: suite.to.Hex(), amount: &hundredInt, gasPrice: &zeroInt, accessList: &types.AccessList{}, chainID: nil, expectPass: false},
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

		tx := NewTx(chainID, 1, &to, amount, 1000, gasPrice, nil, nil, nil, tc.accessList)
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
		ethSigner  types.Signer
		malleate   func(tx *MsgEthereumTx)
		expectPass bool
	}{
		{
			"pass - EIP2930 signer",
			NewTx(suite.chainID, 0, &suite.to, nil, 100000, nil, nil, nil, []byte("test"), &types.AccessList{}),
			types.NewEIP2930Signer(suite.chainID),
			func(tx *MsgEthereumTx) { tx.From = suite.from.Hex() },
			true,
		},
		{
			"pass - EIP155 signer",
			NewTx(suite.chainID, 0, &suite.to, nil, 100000, nil, nil, nil, []byte("test"), nil),
			types.NewEIP155Signer(suite.chainID),
			func(tx *MsgEthereumTx) { tx.From = suite.from.Hex() },
			true,
		},
		{
			"pass - Homestead signer",
			NewTx(suite.chainID, 0, &suite.to, nil, 100000, nil, nil, nil, []byte("test"), nil),
			types.HomesteadSigner{},
			func(tx *MsgEthereumTx) { tx.From = suite.from.Hex() },
			true,
		},
		{
			"pass - Frontier signer",
			NewTx(suite.chainID, 0, &suite.to, nil, 100000, nil, nil, nil, []byte("test"), nil),
			types.FrontierSigner{},
			func(tx *MsgEthereumTx) { tx.From = suite.from.Hex() },
			true,
		},
		{
			"no from address ",
			NewTx(suite.chainID, 0, &suite.to, nil, 100000, nil, nil, nil, []byte("test"), &types.AccessList{}),
			types.NewEIP2930Signer(suite.chainID),
			func(tx *MsgEthereumTx) { tx.From = "" },
			false,
		},
		{
			"from address ≠ signer address",
			NewTx(suite.chainID, 0, &suite.to, nil, 100000, nil, nil, nil, []byte("test"), &types.AccessList{}),
			types.NewEIP2930Signer(suite.chainID),
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

func (suite *MsgsTestSuite) TestFromEthereumTx() {
	privkey, _ := ethsecp256k1.GenerateKey()
	ethPriv, err := privkey.ToECDSA()
	suite.Require().NoError(err)

	// 10^80 is more than 256 bits
	exp_10_80 := new(big.Int).Mul(big.NewInt(1), new(big.Int).Exp(big.NewInt(10), big.NewInt(80), nil))

	testCases := []struct {
		msg        string
		expectPass bool
		buildTx    func() *ethtypes.Transaction
	}{
		{"success, normal tx", true, func() *ethtypes.Transaction {
			tx := ethtypes.NewTransaction(
				0,
				common.BigToAddress(big.NewInt(1)),
				big.NewInt(10),
				21000, big.NewInt(0),
				nil,
			)
			tx, err := ethtypes.SignTx(tx, types.NewEIP2930Signer(suite.chainID), ethPriv)
			suite.Require().NoError(err)
			return tx
		}},
		{"fail, value bigger than 256bits", false, func() *ethtypes.Transaction {
			tx := ethtypes.NewTransaction(
				0,
				common.BigToAddress(big.NewInt(1)),
				exp_10_80,
				21000, big.NewInt(0),
				nil,
			)
			tx, err := ethtypes.SignTx(tx, types.NewEIP2930Signer(suite.chainID), ethPriv)
			suite.Require().NoError(err)
			return tx
		}},
		{"fail, gas price bigger than 256bits", false, func() *ethtypes.Transaction {
			tx := ethtypes.NewTransaction(
				0,
				common.BigToAddress(big.NewInt(1)),
				big.NewInt(10),
				21000, exp_10_80,
				nil,
			)
			tx, err := ethtypes.SignTx(tx, types.NewEIP2930Signer(suite.chainID), ethPriv)
			suite.Require().NoError(err)
			return tx
		}},
	}

	for _, tc := range testCases {
		ethTx := tc.buildTx()
		tx := &MsgEthereumTx{}
		err := tx.FromEthereumTx(ethTx)
		if tc.expectPass {
			suite.Require().NoError(err)

			// round-trip test
			suite.assertEthTxEqual(tx.AsTransaction(), ethTx)
		} else {
			suite.Require().Error(err)
		}
	}
}

func (suite *MsgsTestSuite) assertEthTxEqual(tx1 *ethtypes.Transaction, tx2 *ethtypes.Transaction) {
	suite.Require().Equal(tx1.Hash(), tx2.Hash())
	suite.Require().Equal(tx1.Size(), tx2.Size())

	bin1, err := tx1.MarshalBinary()
	suite.Require().NoError(err)
	bin2, err := tx2.MarshalBinary()
	suite.Require().NoError(err)
	suite.Require().Equal(bin1, bin2)
}
