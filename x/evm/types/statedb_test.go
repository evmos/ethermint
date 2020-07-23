package types_test

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"

	ethcmn "github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"

	"github.com/cosmos/ethermint/app"
	"github.com/cosmos/ethermint/crypto"
	ethermint "github.com/cosmos/ethermint/types"
	"github.com/cosmos/ethermint/x/evm/keeper"
	"github.com/cosmos/ethermint/x/evm/types"

	abci "github.com/tendermint/tendermint/abci/types"
)

type StateDBTestSuite struct {
	suite.Suite

	ctx         sdk.Context
	querier     sdk.Querier
	app         *app.EthermintApp
	stateDB     *types.CommitStateDB
	address     ethcmn.Address
	stateObject types.StateObject
}

func TestStateDBTestSuite(t *testing.T) {
	suite.Run(t, new(StateDBTestSuite))
}

func (suite *StateDBTestSuite) SetupTest() {
	checkTx := false

	suite.app = app.Setup(checkTx)
	suite.ctx = suite.app.BaseApp.NewContext(checkTx, abci.Header{Height: 1})
	suite.querier = keeper.NewQuerier(suite.app.EvmKeeper)
	suite.stateDB = suite.app.EvmKeeper.CommitStateDB.WithContext(suite.ctx)

	privkey, err := crypto.GenerateKey()
	suite.Require().NoError(err)

	suite.address = ethcmn.BytesToAddress(privkey.PubKey().Address().Bytes())
	acc := &ethermint.EthAccount{
		BaseAccount: auth.NewBaseAccount(sdk.AccAddress(suite.address.Bytes()), nil, 0, 0),
		CodeHash:    ethcrypto.Keccak256(nil),
	}

	suite.app.AccountKeeper.SetAccount(suite.ctx, acc)
	suite.stateObject = suite.stateDB.GetOrNewStateObject(suite.address)
}
func (suite *StateDBTestSuite) TestBloomFilter() {
	// Prepare db for logs
	tHash := ethcmn.BytesToHash([]byte{0x1})
	suite.stateDB.Prepare(tHash, ethcmn.Hash{}, 0)
	contractAddress := ethcmn.BigToAddress(big.NewInt(1))
	log := ethtypes.Log{Address: contractAddress}

	testCase := []struct {
		name     string
		malleate func()
		numLogs  int
		isBloom  bool
	}{
		{
			"no logs",
			func() {},
			0,
			false,
		},
		{
			"add log",
			func() {
				suite.stateDB.AddLog(&log)
			},
			1,
			false,
		},
		{
			"bloom",
			func() {},
			0,
			true,
		},
	}

	for _, tc := range testCase {
		tc.malleate()
		logs, err := suite.stateDB.GetLogs(tHash)
		if !tc.isBloom {
			suite.Require().NoError(err, tc.name)
			suite.Require().Len(logs, tc.numLogs, tc.name)
			if len(logs) != 0 {
				suite.Require().Equal(log, *logs[0], tc.name)
			}
		} else {
			// get logs bloom from the log
			bloomInt := ethtypes.LogsBloom(logs)
			bloomFilter := ethtypes.BytesToBloom(bloomInt.Bytes())
			suite.Require().True(ethtypes.BloomLookup(bloomFilter, contractAddress), tc.name)
			suite.Require().False(ethtypes.BloomLookup(bloomFilter, ethcmn.BigToAddress(big.NewInt(2))), tc.name)
		}
	}
}

func (suite *StateDBTestSuite) TestStateDBBalance() {
	testCase := []struct {
		name     string
		malleate func()
		balance  *big.Int
	}{
		{
			"set balance",
			func() {
				suite.stateDB.SetBalance(suite.address, big.NewInt(100))
			},
			big.NewInt(100),
		},
		{
			"sub balance",
			func() {
				suite.stateDB.SubBalance(suite.address, big.NewInt(100))
			},
			big.NewInt(0),
		},
		{
			"add balance",
			func() {
				suite.stateDB.AddBalance(suite.address, big.NewInt(200))
			},
			big.NewInt(200),
		},
		{
			"sub more than balance",
			func() {
				suite.stateDB.SubBalance(suite.address, big.NewInt(300))
			},
			big.NewInt(-100),
		},
	}

	for _, tc := range testCase {
		tc.malleate()
		suite.Require().Equal(tc.balance, suite.stateDB.GetBalance(suite.address), tc.name)
	}
}

func (suite *StateDBTestSuite) TestStateDBNonce() {
	nonce := uint64(123)
	suite.stateDB.SetNonce(suite.address, nonce)
	suite.Require().Equal(nonce, suite.stateDB.GetNonce(suite.address))
}

func (suite *StateDBTestSuite) TestStateDBState() {
	key := ethcmn.BytesToHash([]byte("foo"))
	val := ethcmn.BytesToHash([]byte("bar"))
	suite.stateDB.SetState(suite.address, key, val)

	testCase := []struct {
		name    string
		address ethcmn.Address
		key     ethcmn.Hash
		value   ethcmn.Hash
	}{
		{
			"found state",
			suite.address,
			ethcmn.BytesToHash([]byte("foo")),
			ethcmn.BytesToHash([]byte("bar")),
		},
		{
			"state not found",
			suite.address,
			ethcmn.BytesToHash([]byte("key")),
			ethcmn.Hash{},
		},
		{
			"object not found",
			ethcmn.Address{},
			ethcmn.BytesToHash([]byte("foo")),
			ethcmn.Hash{},
		},
	}
	for _, tc := range testCase {
		value := suite.stateDB.GetState(tc.address, tc.key)
		suite.Require().Equal(tc.value, value, tc.name)
	}
}

func (suite *StateDBTestSuite) TestStateDBCode() {
	testCase := []struct {
		name     string
		address  ethcmn.Address
		code     []byte
		malleate func()
	}{
		{
			"no stored code for state object",
			suite.address,
			nil,
			func() {},
		},
		{
			"existing address",
			suite.address,
			[]byte("code"),
			func() {
				suite.stateDB.SetCode(suite.address, []byte("code"))
			},
		},
		{
			"state object not found",
			ethcmn.Address{},
			nil,
			func() {},
		},
	}

	for _, tc := range testCase {
		tc.malleate()

		suite.Require().Equal(tc.code, suite.stateDB.GetCode(tc.address), tc.name)
		suite.Require().Equal(len(tc.code), suite.stateDB.GetCodeSize(tc.address), tc.name)
	}
}

func (suite *StateDBTestSuite) TestStateDBLogs() {
	testCase := []struct {
		name string
		log  ethtypes.Log
	}{
		{
			"state db log",
			ethtypes.Log{
				Address:     suite.address,
				Topics:      []ethcmn.Hash{ethcmn.BytesToHash([]byte("topic"))},
				Data:        []byte("data"),
				BlockNumber: 1,
				TxHash:      ethcmn.Hash{},
				TxIndex:     1,
				BlockHash:   ethcmn.Hash{},
				Index:       1,
				Removed:     false,
			},
		},
	}

	for _, tc := range testCase {
		hash := ethcmn.BytesToHash([]byte("hash"))
		logs := []*ethtypes.Log{&tc.log}

		err := suite.stateDB.SetLogs(hash, logs)
		suite.Require().NoError(err, tc.name)
		dbLogs, err := suite.stateDB.GetLogs(hash)
		suite.Require().NoError(err, tc.name)
		suite.Require().Equal(logs, dbLogs, tc.name)

		suite.stateDB.DeleteLogs(hash)
		dbLogs, err = suite.stateDB.GetLogs(hash)
		suite.Require().NoError(err, tc.name)
		suite.Require().Empty(dbLogs, tc.name)

		suite.stateDB.AddLog(&tc.log)
		suite.Require().Equal(logs, suite.stateDB.AllLogs(), tc.name)

		//resets state but checking to see if storekey still persists.
		err = suite.stateDB.Reset(hash)
		suite.Require().NoError(err, tc.name)
		suite.Require().Equal(logs, suite.stateDB.AllLogs(), tc.name)
	}
}

func (suite *StateDBTestSuite) TestStateDBPreimage() {
	hash := ethcmn.BytesToHash([]byte("hash"))
	preimage := []byte("preimage")

	suite.stateDB.AddPreimage(hash, preimage)

	suite.Require().Equal(preimage, suite.stateDB.Preimages()[hash])
}

func (suite *StateDBTestSuite) TestStateDBRefund() {
	testCase := []struct {
		name      string
		addAmount uint64
		subAmount uint64
		expRefund uint64
		expPanic  bool
	}{
		{
			"refund 0",
			0, 0, 0,
			false,
		},
		{
			"refund positive amount",
			100, 0, 100,
			false,
		},
		{
			"refund panic",
			100, 200, 100,
			true,
		},
	}

	for _, tc := range testCase {
		suite.Run(tc.name, func() {
			suite.SetupTest() // reset

			suite.stateDB.AddRefund(tc.addAmount)
			suite.Require().Equal(tc.addAmount, suite.stateDB.GetRefund())

			if tc.expPanic {
				suite.Panics(func() {
					suite.stateDB.SubRefund(tc.subAmount)
				})
			} else {
				suite.stateDB.SubRefund(tc.subAmount)
				suite.Require().Equal(tc.expRefund, suite.stateDB.GetRefund())
			}
		})
	}
}

func (suite *StateDBTestSuite) TestStateDBCreateAcct() {
	prevBalance := big.NewInt(12)

	testCase := []struct {
		name     string
		address  ethcmn.Address
		malleate func()
	}{
		{
			"existing account",
			suite.address,
			func() {
				suite.stateDB.AddBalance(suite.address, prevBalance)
			},
		},
		{
			"new account",
			ethcmn.HexToAddress("0x756F45E3FA69347A9A973A725E3C98bC4db0b4c1"),
			func() {
				prevBalance = big.NewInt(0)
			},
		},
	}

	for _, tc := range testCase {
		tc.malleate()

		suite.stateDB.CreateAccount(tc.address)
		suite.Require().True(suite.stateDB.Exist(tc.address), tc.name)
		suite.Require().Equal(prevBalance, suite.stateDB.GetBalance(tc.address), tc.name)
	}
}

func (suite *StateDBTestSuite) TestStateDBClearStateOjb() {
	priv, err := crypto.GenerateKey()
	suite.Require().NoError(err)

	addr := ethcrypto.PubkeyToAddress(priv.ToECDSA().PublicKey)

	suite.stateDB.CreateAccount(addr)
	suite.Require().True(suite.stateDB.Exist(addr))

	suite.stateDB.ClearStateObjects()
	suite.Require().False(suite.stateDB.Exist(addr))
}

func (suite *StateDBTestSuite) TestStateDBReset() {
	priv, err := crypto.GenerateKey()
	suite.Require().NoError(err)

	addr := ethcrypto.PubkeyToAddress(priv.ToECDSA().PublicKey)

	suite.stateDB.CreateAccount(addr)
	suite.Require().True(suite.stateDB.Exist(addr))

	err = suite.stateDB.Reset(ethcmn.BytesToHash(nil))
	suite.Require().NoError(err)
	suite.Require().False(suite.stateDB.Exist(addr))
}

func (suite *StateDBTestSuite) TestSuiteDBPrepare() {
	thash := ethcmn.BytesToHash([]byte("thash"))
	bhash := ethcmn.BytesToHash([]byte("bhash"))
	txi := 1

	suite.stateDB.Prepare(thash, bhash, txi)

	suite.Require().Equal(txi, suite.stateDB.TxIndex())
	suite.Require().Equal(bhash, suite.stateDB.BlockHash())
}

func (suite *StateDBTestSuite) TestSuiteDBCopyState() {
	testCase := []struct {
		name string
		log  ethtypes.Log
	}{
		{
			"copy state",
			ethtypes.Log{
				Address:     suite.address,
				Topics:      []ethcmn.Hash{ethcmn.BytesToHash([]byte("topic"))},
				Data:        []byte("data"),
				BlockNumber: 1,
				TxHash:      ethcmn.Hash{},
				TxIndex:     1,
				BlockHash:   ethcmn.Hash{},
				Index:       1,
				Removed:     false,
			},
		},
	}

	for _, tc := range testCase {
		hash := ethcmn.BytesToHash([]byte("hash"))
		logs := []*ethtypes.Log{&tc.log}

		err := suite.stateDB.SetLogs(hash, logs)
		suite.Require().NoError(err, tc.name)

		copyDB := suite.stateDB.Copy()

		copiedDBLogs, err := copyDB.GetLogs(hash)
		suite.Require().NoError(err, tc.name)
		suite.Require().Equal(logs, copiedDBLogs, tc.name)
		suite.Require().Equal(suite.stateDB.Exist(suite.address), copyDB.Exist(suite.address), tc.name)
	}
}

func (suite *StateDBTestSuite) TestSuiteDBEmpty() {
	suite.Require().True(suite.stateDB.Empty(suite.address))

	suite.stateDB.SetBalance(suite.address, big.NewInt(100))

	suite.Require().False(suite.stateDB.Empty(suite.address))
}

func (suite *StateDBTestSuite) TestSuiteDBSuicide() {

	testCase := []struct {
		name    string
		amount  *big.Int
		expPass bool
		delete  bool
	}{
		{
			"suicide zero balance",
			big.NewInt(0),
			false, false,
		},
		{
			"suicide with balance",
			big.NewInt(100),
			true, false,
		},
		{
			"delete",
			big.NewInt(0),
			true, true,
		},
	}

	for _, tc := range testCase {
		if tc.delete {
			_, err := suite.stateDB.Commit(tc.delete)
			suite.Require().NoError(err, tc.name)
			suite.Require().False(suite.stateDB.Exist(suite.address), tc.name)
			continue
		}

		if tc.expPass {
			suite.stateDB.SetBalance(suite.address, tc.amount)
			suicide := suite.stateDB.Suicide(suite.address)
			suite.Require().True(suicide, tc.name)
			suite.Require().True(suite.stateDB.HasSuicided(suite.address), tc.name)
		} else {
			//Suicide only works for an account with non-zero balance/nonce
			priv, err := crypto.GenerateKey()
			suite.Require().NoError(err)

			addr := ethcrypto.PubkeyToAddress(priv.ToECDSA().PublicKey)
			suicide := suite.stateDB.Suicide(addr)
			suite.Require().False(suicide, tc.name)
			suite.Require().False(suite.stateDB.HasSuicided(addr), tc.name)
		}
	}
}

func (suite *StateDBTestSuite) TestCommitStateDB_Commit() {
	testCase := []struct {
		name       string
		malleate   func()
		deleteObjs bool
		expPass    bool
	}{
		{
			"commit suicided",
			func() {
				ok := suite.stateDB.Suicide(suite.address)
				suite.Require().True(ok)
			},
			true, true,
		},
		{
			"commit with dirty value",
			func() {
				suite.stateDB.SetCode(suite.address, []byte("code"))
			},
			false, true,
		},
		{
			"faled to update state object",
			func() {
				suite.stateDB.SubBalance(suite.address, big.NewInt(10))
			},
			false, false,
		},
	}

	for _, tc := range testCase {
		tc.malleate()

		hash, err := suite.stateDB.Commit(tc.deleteObjs)
		suite.Require().Equal(ethcmn.Hash{}, hash)

		if !tc.expPass {
			suite.Require().Error(err, tc.name)
			continue
		}

		suite.Require().NoError(err, tc.name)
		acc := suite.app.AccountKeeper.GetAccount(suite.ctx, sdk.AccAddress(suite.address.Bytes()))

		if tc.deleteObjs {
			suite.Require().Nil(acc, tc.name)
			continue
		}

		suite.Require().NotNil(acc, tc.name)
		ethAcc, ok := acc.(*ethermint.EthAccount)
		suite.Require().True(ok)
		suite.Require().Equal(ethcrypto.Keccak256([]byte("code")), ethAcc.CodeHash)
	}
}

func (suite *StateDBTestSuite) TestCommitStateDB_Finalize() {
	testCase := []struct {
		name       string
		malleate   func()
		deleteObjs bool
		expPass    bool
	}{
		{
			"finalize suicided",
			func() {
				ok := suite.stateDB.Suicide(suite.address)
				suite.Require().True(ok)
			},
			true, true,
		},
		{
			"finalize, not suicided",
			func() {
				suite.stateDB.AddBalance(suite.address, big.NewInt(5))
			},
			false, true,
		},
		{
			"finalize, dirty storage",
			func() {
				suite.stateDB.SetState(suite.address, ethcmn.BytesToHash([]byte("key")), ethcmn.BytesToHash([]byte("value")))
			},
			false, true,
		},
		{
			"faled to update state object",
			func() {
				suite.stateDB.SubBalance(suite.address, big.NewInt(10))
			},
			false, false,
		},
	}

	for _, tc := range testCase {
		tc.malleate()

		err := suite.stateDB.Finalise(tc.deleteObjs)

		if !tc.expPass {
			suite.Require().Error(err, tc.name)
			continue
		}

		suite.Require().NoError(err, tc.name)
		acc := suite.app.AccountKeeper.GetAccount(suite.ctx, sdk.AccAddress(suite.address.Bytes()))

		if tc.deleteObjs {
			suite.Require().Nil(acc, tc.name)
			continue
		}

		suite.Require().NotNil(acc, tc.name)
	}
}
