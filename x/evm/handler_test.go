package evm_test

import (
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/ethereum/go-ethereum/common"
	ethcmn "github.com/ethereum/go-ethereum/common"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmos/ethermint/app"
	"github.com/cosmos/ethermint/crypto"
	"github.com/cosmos/ethermint/x/evm"
	"github.com/cosmos/ethermint/x/evm/keeper"
	"github.com/cosmos/ethermint/x/evm/types"

	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto/secp256k1"
)

type EvmTestSuite struct {
	suite.Suite

	ctx     sdk.Context
	handler sdk.Handler
	querier sdk.Querier
	app     *app.EthermintApp
	codec   *codec.Codec
}

func (suite *EvmTestSuite) SetupTest() {
	checkTx := false

	suite.app = app.Setup(checkTx)
	suite.ctx = suite.app.BaseApp.NewContext(checkTx, abci.Header{Height: 1, ChainID: "3", Time: time.Now().UTC()})
	suite.handler = evm.NewHandler(suite.app.EvmKeeper)
	suite.querier = keeper.NewQuerier(suite.app.EvmKeeper)
	suite.codec = codec.New()
}

func TestEvmTestSuite(t *testing.T) {
	suite.Run(t, new(EvmTestSuite))
}

func (suite *EvmTestSuite) TestHandleMsgEthereumTx() {
	privkey, err := crypto.GenerateKey()
	suite.Require().NoError(err)
	sender := ethcmn.HexToAddress(privkey.PubKey().Address().String())

	var (
		tx      types.MsgEthereumTx
		chainID *big.Int
	)

	testCases := []struct {
		msg      string
		malleate func()
		expPass  bool
	}{
		{
			"passed",
			func() {
				suite.app.EvmKeeper.SetBalance(suite.ctx, sender, big.NewInt(100))
				tx = types.NewMsgEthereumTx(0, &sender, big.NewInt(100), 0, big.NewInt(10000), nil)

				// parse context chain ID to big.Int
				var ok bool
				chainID, ok = new(big.Int).SetString(suite.ctx.ChainID(), 10)
				suite.Require().True(ok)

				// sign transaction
				err = tx.Sign(chainID, privkey.ToECDSA())
				suite.Require().NoError(err)
			},
			true,
		},
		{
			"insufficient balance",
			func() {
				tx = types.NewMsgEthereumTxContract(0, big.NewInt(100), 0, big.NewInt(10000), nil)

				// parse context chain ID to big.Int
				var ok bool
				chainID, ok = new(big.Int).SetString(suite.ctx.ChainID(), 10)
				suite.Require().True(ok)

				// sign transaction
				err = tx.Sign(chainID, privkey.ToECDSA())
				suite.Require().NoError(err)
			},
			false,
		},
		{
			"tx encoding failed",
			func() {
				tx = types.NewMsgEthereumTxContract(0, big.NewInt(100), 0, big.NewInt(10000), nil)
			},
			false,
		},
		{
			"invalid chain ID",
			func() {
				suite.ctx = suite.ctx.WithChainID("chainID")
			},
			false,
		},
		{
			"VerifySig failed",
			func() {
				tx = types.NewMsgEthereumTxContract(0, big.NewInt(100), 0, big.NewInt(10000), nil)
			},
			false,
		},
	}

	for _, tc := range testCases {
		suite.Run("", func() {
			suite.SetupTest() // reset
			//nolint
			tc.malleate()

			res, err := suite.handler(suite.ctx, tx)

			//nolint
			if tc.expPass {
				suite.Require().NoError(err)
				suite.Require().NotNil(res)
			} else {
				suite.Require().Error(err)
				suite.Require().Nil(res)
			}
		})
	}
}

func (suite *EvmTestSuite) TestMsgEthermint() {
	var (
		tx   types.MsgEthermint
		from = sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
		to   = sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	)

	testCases := []struct {
		msg      string
		malleate func()
		expPass  bool
	}{
		{
			"passed",
			func() {
				tx = types.NewMsgEthermint(0, &to, sdk.NewInt(1), 100000, sdk.NewInt(2), []byte("test"), from)
				suite.app.EvmKeeper.SetBalance(suite.ctx, ethcmn.BytesToAddress(from.Bytes()), big.NewInt(100))
			},
			true,
		},
		{
			"invalid state transition",
			func() {
				tx = types.NewMsgEthermint(0, &to, sdk.NewInt(1), 100000, sdk.NewInt(2), []byte("test"), from)
			},
			false,
		},
		{
			"invalid chain ID",
			func() {
				suite.ctx = suite.ctx.WithChainID("chainID")
			},
			false,
		},
	}

	for _, tc := range testCases {
		suite.Run("", func() {
			suite.SetupTest() // reset
			//nolint
			tc.malleate()

			res, err := suite.handler(suite.ctx, tx)

			//nolint
			if tc.expPass {
				suite.Require().NoError(err)
				suite.Require().NotNil(res)
			} else {
				suite.Require().Error(err)
				suite.Require().Nil(res)
			}
		})
	}
}

func (suite *EvmTestSuite) TestHandlerLogs() {
	// Test contract:

	// pragma solidity ^0.5.1;

	// contract Test {
	//     event Hello(uint256 indexed world);

	//     constructor() public {
	//         emit Hello(17);
	//     }
	// }

	// {
	// 	"linkReferences": {},
	// 	"object": "6080604052348015600f57600080fd5b5060117f775a94827b8fd9b519d36cd827093c664f93347070a554f65e4a6f56cd73889860405160405180910390a2603580604b6000396000f3fe6080604052600080fdfea165627a7a723058206cab665f0f557620554bb45adf266708d2bd349b8a4314bdff205ee8440e3c240029",
	// 	"opcodes": "PUSH1 0x80 PUSH1 0x40 MSTORE CALLVALUE DUP1 ISZERO PUSH1 0xF JUMPI PUSH1 0x0 DUP1 REVERT JUMPDEST POP PUSH1 0x11 PUSH32 0x775A94827B8FD9B519D36CD827093C664F93347070A554F65E4A6F56CD738898 PUSH1 0x40 MLOAD PUSH1 0x40 MLOAD DUP1 SWAP2 SUB SWAP1 LOG2 PUSH1 0x35 DUP1 PUSH1 0x4B PUSH1 0x0 CODECOPY PUSH1 0x0 RETURN INVALID PUSH1 0x80 PUSH1 0x40 MSTORE PUSH1 0x0 DUP1 REVERT INVALID LOG1 PUSH6 0x627A7A723058 KECCAK256 PUSH13 0xAB665F0F557620554BB45ADF26 PUSH8 0x8D2BD349B8A4314 0xbd SELFDESTRUCT KECCAK256 0x5e 0xe8 DIFFICULTY 0xe EXTCODECOPY 0x24 STOP 0x29 ",
	// 	"sourceMap": "25:119:0:-;;;90:52;8:9:-1;5:2;;;30:1;27;20:12;5:2;90:52:0;132:2;126:9;;;;;;;;;;25:119;;;;;;"
	// }

	gasLimit := uint64(100000)
	gasPrice := big.NewInt(1000000)

	priv, err := crypto.GenerateKey()
	suite.Require().NoError(err, "failed to create key")

	bytecode := common.FromHex("0x6080604052348015600f57600080fd5b5060117f775a94827b8fd9b519d36cd827093c664f93347070a554f65e4a6f56cd73889860405160405180910390a2603580604b6000396000f3fe6080604052600080fdfea165627a7a723058206cab665f0f557620554bb45adf266708d2bd349b8a4314bdff205ee8440e3c240029")
	tx := types.NewMsgEthereumTx(1, nil, big.NewInt(0), gasLimit, gasPrice, bytecode)
	err = tx.Sign(big.NewInt(3), priv.ToECDSA())
	suite.Require().NoError(err)

	result, err := suite.handler(suite.ctx, tx)
	suite.Require().NoError(err, "failed to handle eth tx msg")

	resultData, err := types.DecodeResultData(result.Data)
	suite.Require().NoError(err, "failed to decode result data")

	suite.Require().Equal(len(resultData.Logs), 1)
	suite.Require().Equal(len(resultData.Logs[0].Topics), 2)

	hash := []byte{1}
	err = suite.app.EvmKeeper.SetLogs(suite.ctx, ethcmn.BytesToHash(hash), resultData.Logs)
	suite.Require().NoError(err)

	logs, err := suite.app.EvmKeeper.GetLogs(suite.ctx, ethcmn.BytesToHash(hash))
	suite.Require().NoError(err, "failed to get logs")

	suite.Require().Equal(logs, resultData.Logs)
}

func (suite *EvmTestSuite) TestQueryTxLogs() {
	gasLimit := uint64(100000)
	gasPrice := big.NewInt(1000000)

	priv, err := crypto.GenerateKey()
	suite.Require().NoError(err, "failed to create key")

	// send contract deployment transaction with an event in the constructor
	bytecode := common.FromHex("0x6080604052348015600f57600080fd5b5060117f775a94827b8fd9b519d36cd827093c664f93347070a554f65e4a6f56cd73889860405160405180910390a2603580604b6000396000f3fe6080604052600080fdfea165627a7a723058206cab665f0f557620554bb45adf266708d2bd349b8a4314bdff205ee8440e3c240029")
	tx := types.NewMsgEthereumTx(1, nil, big.NewInt(0), gasLimit, gasPrice, bytecode)
	err = tx.Sign(big.NewInt(3), priv.ToECDSA())
	suite.Require().NoError(err)

	result, err := suite.handler(suite.ctx, tx)
	suite.Require().NoError(err)
	suite.Require().NotNil(result)

	resultData, err := types.DecodeResultData(result.Data)
	suite.Require().NoError(err, "failed to decode result data")

	suite.Require().Equal(len(resultData.Logs), 1)
	suite.Require().Equal(len(resultData.Logs[0].Topics), 2)

	// get logs by tx hash
	hash := resultData.TxHash.Bytes()

	logs, err := suite.app.EvmKeeper.GetLogs(suite.ctx, ethcmn.BytesToHash(hash))
	suite.Require().NoError(err, "failed to get logs")

	suite.Require().Equal(logs, resultData.Logs)

	// query tx logs
	path := []string{"transactionLogs", fmt.Sprintf("0x%x", hash)}
	res, err := suite.querier(suite.ctx, path, abci.RequestQuery{})
	suite.Require().NoError(err, "failed to query txLogs")

	var txLogs types.QueryETHLogs
	suite.codec.MustUnmarshalJSON(res, &txLogs)

	// amino decodes an empty byte array as nil, whereas JSON decodes it as []byte{} causing a discrepancy
	resultData.Logs[0].Data = []byte{}
	suite.Require().Equal(txLogs.Logs[0], resultData.Logs[0])
}

func (suite *EvmTestSuite) TestSendTransaction() {
	gasLimit := uint64(21000)
	gasPrice := big.NewInt(1)

	priv, err := crypto.GenerateKey()
	suite.Require().NoError(err, "failed to create key")
	pub := priv.ToECDSA().Public().(*ecdsa.PublicKey)

	suite.app.EvmKeeper.SetBalance(suite.ctx, ethcrypto.PubkeyToAddress(*pub), big.NewInt(100))

	// send simple value transfer with gasLimit=21000
	tx := types.NewMsgEthereumTx(1, &ethcmn.Address{0x1}, big.NewInt(1), gasLimit, gasPrice, nil)
	err = tx.Sign(big.NewInt(3), priv.ToECDSA())
	suite.Require().NoError(err)

	result, err := suite.handler(suite.ctx, tx)
	suite.Require().NoError(err)
	suite.Require().NotNil(result)
}
