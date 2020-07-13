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

	// Generate and add a log to test
	log := ethtypes.Log{Address: contractAddress}
	suite.stateDB.AddLog(&log)

	// Get log from db
	logs, err := suite.stateDB.GetLogs(tHash)
	suite.Require().NoError(err)
	suite.Require().Len(logs, 1)
	suite.Require().Equal(log, *logs[0])

	// get logs bloom from the log
	bloomInt := ethtypes.LogsBloom(logs)
	bloomFilter := ethtypes.BytesToBloom(bloomInt.Bytes())

	// Check to make sure bloom filter will succeed on
	suite.Require().True(ethtypes.BloomLookup(bloomFilter, contractAddress))
	suite.Require().False(ethtypes.BloomLookup(bloomFilter, ethcmn.BigToAddress(big.NewInt(2))))
}

func (suite *StateDBTestSuite) TestStateDBBalance() {
	priv, err := crypto.GenerateKey()
	suite.Require().NoError(err)

	addr := ethcrypto.PubkeyToAddress(priv.ToECDSA().PublicKey)
	value := big.NewInt(100)
	suite.stateDB.SetBalance(addr, value)
	suite.Require().Equal(value, suite.stateDB.GetBalance(addr))

	suite.stateDB.SubBalance(addr, value)
	suite.Require().Equal(big.NewInt(0), suite.stateDB.GetBalance(addr))

	suite.stateDB.AddBalance(addr, value)
	suite.Require().Equal(value, suite.stateDB.GetBalance(addr))
}

func (suite *StateDBTestSuite) TestStateDBNonce() {
	priv, err := crypto.GenerateKey()
	suite.Require().NoError(err)
	addr := ethcrypto.PubkeyToAddress(priv.ToECDSA().PublicKey)

	nonce := uint64(123)
	suite.stateDB.SetNonce(addr, nonce)

	suite.Require().Equal(nonce, suite.stateDB.GetNonce(addr))
}

func (suite *StateDBTestSuite) TestStateDBState() {
	priv, err := crypto.GenerateKey()
	suite.Require().NoError(err)

	addr := ethcrypto.PubkeyToAddress(priv.ToECDSA().PublicKey)
	key := ethcmn.BytesToHash([]byte("foo"))
	val := ethcmn.BytesToHash([]byte("bar"))

	suite.stateDB.SetState(addr, key, val)

	suite.Require().Equal(val, suite.stateDB.GetState(addr, key))
}

func (suite *StateDBTestSuite) TestStateDBCode() {
	priv, err := crypto.GenerateKey()
	suite.Require().NoError(err)

	addr := ethcrypto.PubkeyToAddress(priv.ToECDSA().PublicKey)
	code := []byte("foobar")

	suite.stateDB.SetCode(addr, code)

	suite.Require().Equal(code, suite.stateDB.GetCode(addr))

	codelen := len(code)
	suite.Require().Equal(codelen, suite.stateDB.GetCodeSize(addr))
}

func (suite *StateDBTestSuite) TestStateDBLogs() {
	priv, err := crypto.GenerateKey()
	suite.Require().NoError(err)

	addr := ethcrypto.PubkeyToAddress(priv.ToECDSA().PublicKey)

	hash := ethcmn.BytesToHash([]byte("hash"))
	log := ethtypes.Log{
		Address:     addr,
		Topics:      []ethcmn.Hash{ethcmn.BytesToHash([]byte("topic"))},
		Data:        []byte("data"),
		BlockNumber: 1,
		TxHash:      ethcmn.Hash{},
		TxIndex:     1,
		BlockHash:   ethcmn.Hash{},
		Index:       1,
		Removed:     false,
	}
	logs := []*ethtypes.Log{&log}

	err = suite.stateDB.SetLogs(hash, logs)
	suite.Require().NoError(err)
	dbLogs, err := suite.stateDB.GetLogs(hash)
	suite.Require().NoError(err)
	suite.Require().Equal(logs, dbLogs)

	suite.stateDB.DeleteLogs(hash)
	dbLogs, err = suite.stateDB.GetLogs(hash)
	suite.Require().NoError(err)
	suite.Require().Empty(dbLogs)

	suite.stateDB.AddLog(&log)
	suite.Require().Equal(logs, suite.stateDB.AllLogs())

	//resets state but checking to see if storekey still persists.
	err = suite.stateDB.Reset(hash)
	suite.Require().NoError(err)
	suite.Require().Equal(logs, suite.stateDB.AllLogs())
}

func (suite *StateDBTestSuite) TestStateDBPreimage() {
	hash := ethcmn.BytesToHash([]byte("hash"))
	preimage := []byte("preimage")

	suite.stateDB.AddPreimage(hash, preimage)

	suite.Require().Equal(preimage, suite.stateDB.Preimages()[hash])
}

func (suite *StateDBTestSuite) TestStateDBRefund() {
	value := uint64(100)

	suite.stateDB.AddRefund(value)
	suite.Require().Equal(value, suite.stateDB.GetRefund())

	suite.stateDB.SubRefund(value)
	suite.Require().Equal(uint64(0), suite.stateDB.GetRefund())
}

func (suite *StateDBTestSuite) TestStateDBCreateAcct() {
	priv, err := crypto.GenerateKey()
	suite.Require().NoError(err)

	addr := ethcrypto.PubkeyToAddress(priv.ToECDSA().PublicKey)

	suite.stateDB.CreateAccount(addr)
	suite.Require().True(suite.stateDB.Exist(addr))

	value := big.NewInt(100)
	suite.stateDB.AddBalance(addr, value)

	suite.stateDB.CreateAccount(addr)
	suite.Require().Equal(value, suite.stateDB.GetBalance(addr))
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

	hash := ethcmn.BytesToHash([]byte("hash"))

	suite.stateDB.CreateAccount(addr)
	suite.Require().True(suite.stateDB.Exist(addr))

	err = suite.stateDB.Reset(hash)
	suite.Require().NoError(err)
	suite.Require().False(suite.stateDB.Exist(addr))
}

func (suite *StateDBTestSuite) TestStateDBUpdateAcct() {

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
	priv, err := crypto.GenerateKey()
	suite.Require().NoError(err)

	addr := ethcrypto.PubkeyToAddress(priv.ToECDSA().PublicKey)

	hash := ethcmn.BytesToHash([]byte("hash"))
	log := ethtypes.Log{
		Address:     addr,
		Topics:      []ethcmn.Hash{ethcmn.BytesToHash([]byte("topic"))},
		Data:        []byte("data"),
		BlockNumber: 1,
		TxHash:      ethcmn.Hash{},
		TxIndex:     1,
		BlockHash:   ethcmn.Hash{},
		Index:       1,
		Removed:     false,
	}
	logs := []*ethtypes.Log{&log}

	err = suite.stateDB.SetLogs(hash, logs)
	suite.Require().NoError(err)

	copyDB := suite.stateDB.Copy()

	copiedDBLogs, err := copyDB.GetLogs(hash)
	suite.Require().NoError(err)
	suite.Require().Equal(logs, copiedDBLogs)
	suite.Require().Equal(suite.stateDB.Exist(addr), copyDB.Exist(addr))
}

func (suite *StateDBTestSuite) TestSuiteDBEmpty() {
	priv, err := crypto.GenerateKey()
	suite.Require().NoError(err)

	addr := ethcrypto.PubkeyToAddress(priv.ToECDSA().PublicKey)

	suite.Require().True(suite.stateDB.Empty(addr))

	suite.stateDB.SetBalance(addr, big.NewInt(100))

	suite.Require().False(suite.stateDB.Empty(addr))
}

func (suite *StateDBTestSuite) TestSuiteDBSuicide() {
	priv, err := crypto.GenerateKey()
	suite.Require().NoError(err)

	addr := ethcrypto.PubkeyToAddress(priv.ToECDSA().PublicKey)

	suicide := suite.stateDB.Suicide(addr)
	suite.Require().False(suicide)
	suite.Require().False(suite.stateDB.HasSuicided(addr))

	//Suicide only works for an account with non-zero balance/nonce
	suite.stateDB.SetBalance(addr, big.NewInt(100))
	suicide = suite.stateDB.Suicide(addr)

	suite.Require().True(suicide)
	suite.Require().True(suite.stateDB.HasSuicided(addr))

	delete := true
	_, err = suite.stateDB.Commit(delete)
	suite.Require().NoError(err)
	suite.Require().False(suite.stateDB.Exist(addr))
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
