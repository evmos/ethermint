package evm_test

import (
	"encoding/json"
	"math/big"
	"strings"
	"testing"
	"time"

	"github.com/status-im/keycard-go/hexutils"

	"github.com/stretchr/testify/suite"

	"github.com/ethereum/go-ethereum/common"
	ethcmn "github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmos/ethermint/app"
	"github.com/cosmos/ethermint/crypto/ethsecp256k1"
	"github.com/cosmos/ethermint/tests"
	"github.com/cosmos/ethermint/x/evm"
	"github.com/cosmos/ethermint/x/evm/types"

	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
)

type EvmTestSuite struct {
	suite.Suite

	ctx     sdk.Context
	handler sdk.Handler
	app     *app.EthermintApp
	codec   codec.BinaryMarshaler
	chainID *big.Int

	signer    keyring.Signer
	ethSigner ethtypes.Signer
	from      ethcmn.Address
	to        sdk.AccAddress
}

func (suite *EvmTestSuite) SetupTest() {
	checkTx := false

	suite.app = app.Setup(checkTx)
	suite.ctx = suite.app.BaseApp.NewContext(checkTx, tmproto.Header{Height: 1, ChainID: "ethermint-1", Time: time.Now().UTC()})
	suite.app.EvmKeeper.CommitStateDB.WithContext(suite.ctx)
	suite.handler = evm.NewHandler(suite.app.EvmKeeper)
	suite.codec = suite.app.AppCodec()
	suite.chainID = suite.app.EvmKeeper.ChainID()

	privKey, err := ethsecp256k1.GenerateKey()
	suite.Require().NoError(err)

	suite.to = sdk.AccAddress(privKey.PubKey().Address())

	privKey, err = ethsecp256k1.GenerateKey()
	suite.Require().NoError(err)

	suite.signer = tests.NewSigner(privKey)
	suite.ethSigner = ethtypes.LatestSignerForChainID(suite.chainID)
	suite.from = ethcmn.BytesToAddress(privKey.PubKey().Address().Bytes())

}

func TestEvmTestSuite(t *testing.T) {
	suite.Run(t, new(EvmTestSuite))
}

func (suite *EvmTestSuite) TestHandleMsgEthereumTx() {

	var tx *types.MsgEthereumTx

	testCases := []struct {
		msg      string
		malleate func()
		expPass  bool
	}{
		{
			"passed",
			func() {
				suite.app.EvmKeeper.CommitStateDB.SetBalance(suite.from, big.NewInt(100))
				to := ethcmn.BytesToAddress(suite.to)
				tx = types.NewMsgEthereumTx(suite.chainID, 0, &to, big.NewInt(100), 0, big.NewInt(10000), nil, nil)
				tx.From = suite.from.String()

				// sign transaction
				err := tx.Sign(suite.ethSigner, suite.signer)
				suite.Require().NoError(err)
			},
			true,
		},
		{
			"insufficient balance",
			func() {
				tx = types.NewMsgEthereumTxContract(suite.chainID, 0, big.NewInt(100), 0, big.NewInt(10000), nil, nil)

				// sign transaction
				err := tx.Sign(suite.ethSigner, suite.signer)
				suite.Require().NoError(err)
			},
			false,
		},
		{
			"tx encoding failed",
			func() {
				tx = types.NewMsgEthereumTxContract(suite.chainID, 0, big.NewInt(100), 0, big.NewInt(10000), nil, nil)
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
				tx = types.NewMsgEthereumTxContract(suite.chainID, 0, big.NewInt(100), 0, big.NewInt(10000), nil, nil)
			},
			false,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.msg, func() {
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

	bytecode := common.FromHex("0x6080604052348015600f57600080fd5b5060117f775a94827b8fd9b519d36cd827093c664f93347070a554f65e4a6f56cd73889860405160405180910390a2603580604b6000396000f3fe6080604052600080fdfea165627a7a723058206cab665f0f557620554bb45adf266708d2bd349b8a4314bdff205ee8440e3c240029")
	tx := types.NewMsgEthereumTx(suite.chainID, 1, nil, big.NewInt(0), gasLimit, gasPrice, bytecode, nil)
	tx.From = suite.from.String()

	err := tx.Sign(suite.ethSigner, suite.signer)
	suite.Require().NoError(err)

	result, err := suite.handler(suite.ctx, tx)
	suite.Require().NoError(err, "failed to handle eth tx msg")

	txResponse, err := types.DecodeTxResponse(result.Data)
	suite.Require().NoError(err, "failed to decode result data")

	suite.Require().Equal(len(txResponse.TxLogs.Logs), 1)
	suite.Require().Equal(len(txResponse.TxLogs.Logs[0].Topics), 2)

	hash := []byte{1}
	err = suite.app.EvmKeeper.CommitStateDB.SetLogs(ethcmn.BytesToHash(hash), txResponse.TxLogs.EthLogs())
	suite.Require().NoError(err)

	logs, err := suite.app.EvmKeeper.CommitStateDB.GetLogs(ethcmn.BytesToHash(hash))
	suite.Require().NoError(err, "failed to get logs")

	suite.Require().Equal(logs, txResponse.TxLogs.Logs)
}

func (suite *EvmTestSuite) TestQueryTxLogs() {
	gasLimit := uint64(100000)
	gasPrice := big.NewInt(1000000)

	// send contract deployment transaction with an event in the constructor
	bytecode := common.FromHex("0x6080604052348015600f57600080fd5b5060117f775a94827b8fd9b519d36cd827093c664f93347070a554f65e4a6f56cd73889860405160405180910390a2603580604b6000396000f3fe6080604052600080fdfea165627a7a723058206cab665f0f557620554bb45adf266708d2bd349b8a4314bdff205ee8440e3c240029")
	tx := types.NewMsgEthereumTx(suite.chainID, 1, nil, big.NewInt(0), gasLimit, gasPrice, bytecode, nil)
	tx.From = suite.from.String()

	err := tx.Sign(suite.ethSigner, suite.signer)
	suite.Require().NoError(err)

	result, err := suite.handler(suite.ctx, tx)
	suite.Require().NoError(err)
	suite.Require().NotNil(result)

	txResponse, err := types.DecodeTxResponse(result.Data)
	suite.Require().NoError(err, "failed to decode result data")

	suite.Require().Equal(len(txResponse.TxLogs.Logs), 1)
	suite.Require().Equal(len(txResponse.TxLogs.Logs[0].Topics), 2)

	// get logs by tx hash
	hash := txResponse.TxLogs.Hash

	logs, err := suite.app.EvmKeeper.CommitStateDB.GetLogs(ethcmn.HexToHash(hash))
	suite.Require().NoError(err, "failed to get logs")

	suite.Require().Equal(logs, txResponse.TxLogs.EthLogs())
}

func (suite *EvmTestSuite) TestDeployAndCallContract() {
	// Test contract:
	//http://remix.ethereum.org/#optimize=false&evmVersion=istanbul&version=soljson-v0.5.15+commit.6a57276f.js
	//2_Owner.sol
	//
	//pragma solidity >=0.4.22 <0.7.0;
	//
	///**
	// * @title Owner
	// * @dev Set & change owner
	// */
	//contract Owner {
	//
	//	address private owner;
	//
	//	// event for EVM logging
	//	event OwnerSet(address indexed oldOwner, address indexed newOwner);
	//
	//	// modifier to check if caller is owner
	//	modifier isOwner() {
	//	// If the first argument of 'require' evaluates to 'false', execution terminates and all
	//	// changes to the state and to Ether balances are reverted.
	//	// This used to consume all gas in old EVM versions, but not anymore.
	//	// It is often a good idea to use 'require' to check if functions are called correctly.
	//	// As a second argument, you can also provide an explanation about what went wrong.
	//	require(msg.sender == owner, "Caller is not owner");
	//	_;
	//}
	//
	//	/**
	//	 * @dev Set contract deployer as owner
	//	 */
	//	constructor() public {
	//	owner = msg.sender; // 'msg.sender' is sender of current call, contract deployer for a constructor
	//	emit OwnerSet(address(0), owner);
	//}
	//
	//	/**
	//	 * @dev Change owner
	//	 * @param newOwner address of new owner
	//	 */
	//	function changeOwner(address newOwner) public isOwner {
	//	emit OwnerSet(owner, newOwner);
	//	owner = newOwner;
	//}
	//
	//	/**
	//	 * @dev Return owner address
	//	 * @return address of owner
	//	 */
	//	function getOwner() external view returns (address) {
	//	return owner;
	//}
	//}

	// Deploy contract - Owner.sol
	gasLimit := uint64(100000000)
	gasPrice := big.NewInt(10000)

	bytecode := common.FromHex("0x608060405234801561001057600080fd5b50336000806101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff1602179055506000809054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16600073ffffffffffffffffffffffffffffffffffffffff167f342827c97908e5e2f71151c08502a66d44b6f758e3ac2f1de95f02eb95f0a73560405160405180910390a36102c4806100dc6000396000f3fe608060405234801561001057600080fd5b5060043610610053576000357c010000000000000000000000000000000000000000000000000000000090048063893d20e814610058578063a6f9dae1146100a2575b600080fd5b6100606100e6565b604051808273ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200191505060405180910390f35b6100e4600480360360208110156100b857600080fd5b81019080803573ffffffffffffffffffffffffffffffffffffffff16906020019092919050505061010f565b005b60008060009054906101000a900473ffffffffffffffffffffffffffffffffffffffff16905090565b6000809054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff16146101d1576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004018080602001828103825260138152602001807f43616c6c6572206973206e6f74206f776e65720000000000000000000000000081525060200191505060405180910390fd5b8073ffffffffffffffffffffffffffffffffffffffff166000809054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff167f342827c97908e5e2f71151c08502a66d44b6f758e3ac2f1de95f02eb95f0a73560405160405180910390a3806000806101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff1602179055505056fea265627a7a72315820f397f2733a89198bc7fed0764083694c5b828791f39ebcbc9e414bccef14b48064736f6c63430005100032")
	tx := types.NewMsgEthereumTx(suite.chainID, 1, nil, big.NewInt(0), gasLimit, gasPrice, bytecode, nil)
	tx.From = suite.from.String()

	err := tx.Sign(suite.ethSigner, suite.signer)
	suite.Require().NoError(err)

	result, err := suite.handler(suite.ctx, tx)
	suite.Require().NoError(err, "failed to handle eth tx msg")

	txResponse, err := types.DecodeTxResponse(result.Data)
	suite.Require().NoError(err, "failed to decode result data")

	// store - changeOwner
	gasLimit = uint64(100000000000)
	gasPrice = big.NewInt(100)
	receiver := common.HexToAddress(txResponse.ContractAddress)

	storeAddr := "0xa6f9dae10000000000000000000000006a82e4a67715c8412a9114fbd2cbaefbc8181424"
	bytecode = common.FromHex(storeAddr)
	tx = types.NewMsgEthereumTx(suite.chainID, 2, &receiver, big.NewInt(0), gasLimit, gasPrice, bytecode, nil)
	tx.From = suite.from.String()

	err = tx.Sign(suite.ethSigner, suite.signer)
	suite.Require().NoError(err)

	result, err = suite.handler(suite.ctx, tx)
	suite.Require().NoError(err, "failed to handle eth tx msg")

	txResponse, err = types.DecodeTxResponse(result.Data)
	suite.Require().NoError(err, "failed to decode result data")

	// query - getOwner
	bytecode = common.FromHex("0x893d20e8")
	tx = types.NewMsgEthereumTx(suite.chainID, 2, &receiver, big.NewInt(0), gasLimit, gasPrice, bytecode, nil)
	tx.From = suite.from.String()
	err = tx.Sign(suite.ethSigner, suite.signer)
	suite.Require().NoError(err)

	result, err = suite.handler(suite.ctx, tx)
	suite.Require().NoError(err, "failed to handle eth tx msg")

	txResponse, err = types.DecodeTxResponse(result.Data)
	suite.Require().NoError(err, "failed to decode result data")

	getAddr := strings.ToLower(hexutils.BytesToHex(txResponse.Ret))
	suite.Require().Equal(true, strings.HasSuffix(storeAddr, getAddr), "Fail to query the address")
}

func (suite *EvmTestSuite) TestSendTransaction() {
	gasLimit := uint64(21000)
	gasPrice := big.NewInt(0x55ae82600)

	suite.app.EvmKeeper.CommitStateDB.SetBalance(suite.from, big.NewInt(100))

	// send simple value transfer with gasLimit=21000
	tx := types.NewMsgEthereumTx(suite.chainID, 1, &ethcmn.Address{0x1}, big.NewInt(1), gasLimit, gasPrice, nil, nil)
	tx.From = suite.from.String()
	err := tx.Sign(suite.ethSigner, suite.signer)
	suite.Require().NoError(err)

	result, err := suite.handler(suite.ctx, tx)
	suite.Require().NoError(err)
	suite.Require().NotNil(result)
}

func (suite *EvmTestSuite) TestOutOfGasWhenDeployContract() {
	// Test contract:
	//http://remix.ethereum.org/#optimize=false&evmVersion=istanbul&version=soljson-v0.5.15+commit.6a57276f.js
	//2_Owner.sol
	//
	//pragma solidity >=0.4.22 <0.7.0;
	//
	///**
	// * @title Owner
	// * @dev Set & change owner
	// */
	//contract Owner {
	//
	//	address private owner;
	//
	//	// event for EVM logging
	//	event OwnerSet(address indexed oldOwner, address indexed newOwner);
	//
	//	// modifier to check if caller is owner
	//	modifier isOwner() {
	//	// If the first argument of 'require' evaluates to 'false', execution terminates and all
	//	// changes to the state and to Ether balances are reverted.
	//	// This used to consume all gas in old EVM versions, but not anymore.
	//	// It is often a good idea to use 'require' to check if functions are called correctly.
	//	// As a second argument, you can also provide an explanation about what went wrong.
	//	require(msg.sender == owner, "Caller is not owner");
	//	_;
	//}
	//
	//	/**
	//	 * @dev Set contract deployer as owner
	//	 */
	//	constructor() public {
	//	owner = msg.sender; // 'msg.sender' is sender of current call, contract deployer for a constructor
	//	emit OwnerSet(address(0), owner);
	//}
	//
	//	/**
	//	 * @dev Change owner
	//	 * @param newOwner address of new owner
	//	 */
	//	function changeOwner(address newOwner) public isOwner {
	//	emit OwnerSet(owner, newOwner);
	//	owner = newOwner;
	//}
	//
	//	/**
	//	 * @dev Return owner address
	//	 * @return address of owner
	//	 */
	//	function getOwner() external view returns (address) {
	//	return owner;
	//}
	//}

	// Deploy contract - Owner.sol
	gasLimit := uint64(1)
	suite.ctx = suite.ctx.WithGasMeter(sdk.NewGasMeter(gasLimit))
	gasPrice := big.NewInt(10000)

	bytecode := common.FromHex("0x608060405234801561001057600080fd5b50336000806101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff1602179055506000809054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16600073ffffffffffffffffffffffffffffffffffffffff167f342827c97908e5e2f71151c08502a66d44b6f758e3ac2f1de95f02eb95f0a73560405160405180910390a36102c4806100dc6000396000f3fe608060405234801561001057600080fd5b5060043610610053576000357c010000000000000000000000000000000000000000000000000000000090048063893d20e814610058578063a6f9dae1146100a2575b600080fd5b6100606100e6565b604051808273ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200191505060405180910390f35b6100e4600480360360208110156100b857600080fd5b81019080803573ffffffffffffffffffffffffffffffffffffffff16906020019092919050505061010f565b005b60008060009054906101000a900473ffffffffffffffffffffffffffffffffffffffff16905090565b6000809054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff16146101d1576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004018080602001828103825260138152602001807f43616c6c6572206973206e6f74206f776e65720000000000000000000000000081525060200191505060405180910390fd5b8073ffffffffffffffffffffffffffffffffffffffff166000809054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff167f342827c97908e5e2f71151c08502a66d44b6f758e3ac2f1de95f02eb95f0a73560405160405180910390a3806000806101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff1602179055505056fea265627a7a72315820f397f2733a89198bc7fed0764083694c5b828791f39ebcbc9e414bccef14b48064736f6c63430005100032")
	tx := types.NewMsgEthereumTx(suite.chainID, 1, nil, big.NewInt(0), gasLimit, gasPrice, bytecode, nil)
	tx.From = suite.from.String()

	err := tx.Sign(suite.ethSigner, suite.signer)
	suite.Require().NoError(err)

	snapshotCommitStateDBJson, err := json.Marshal(suite.app.EvmKeeper.CommitStateDB)
	suite.Require().Nil(err)

	defer func() {
		if r := recover(); r != nil {
			currentCommitStateDBJson, err := json.Marshal(suite.app.EvmKeeper.CommitStateDB)
			suite.Require().Nil(err)
			suite.Require().Equal(snapshotCommitStateDBJson, currentCommitStateDBJson)
		} else {
			suite.Require().Fail("panic did not happen")
		}
	}()

	suite.handler(suite.ctx, tx)
	suite.Require().Fail("panic did not happen")
}

func (suite *EvmTestSuite) TestErrorWhenDeployContract() {
	gasLimit := uint64(1000000)
	gasPrice := big.NewInt(10000)

	bytecode := common.FromHex("0xa6f9dae10000000000000000000000006a82e4a67715c8412a9114fbd2cbaefbc8181424")

	tx := types.NewMsgEthereumTx(suite.chainID, 1, nil, big.NewInt(0), gasLimit, gasPrice, bytecode, nil)
	tx.From = suite.from.String()

	err := tx.Sign(suite.ethSigner, suite.signer)
	suite.Require().NoError(err)

	snapshotCommitStateDBJson, err := json.Marshal(suite.app.EvmKeeper.CommitStateDB)
	suite.Require().Nil(err)

	_, sdkErr := suite.handler(suite.ctx, tx)
	suite.Require().NotNil(sdkErr)

	currentCommitStateDBJson, err := json.Marshal(suite.app.EvmKeeper.CommitStateDB)
	suite.Require().Nil(err)
	suite.Require().Equal(snapshotCommitStateDBJson, currentCommitStateDBJson)
}
