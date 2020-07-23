package keeper_test

import (
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"

	ethcmn "github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"

	"github.com/cosmos/ethermint/crypto"
	ethermint "github.com/cosmos/ethermint/types"
)

func (suite *KeeperTestSuite) TestBloomFilter() {
	// Prepare db for logs
	tHash := ethcmn.BytesToHash([]byte{0x1})
	suite.app.EvmKeeper.Prepare(suite.ctx, tHash, ethcmn.Hash{}, 0)
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
				suite.app.EvmKeeper.AddLog(suite.ctx, &log)
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
		logs, err := suite.app.EvmKeeper.GetLogs(suite.ctx, tHash)
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

func (suite *KeeperTestSuite) TestStateDBBalance() {
	testCase := []struct {
		name     string
		malleate func()
		balance  *big.Int
	}{
		{
			"set balance",
			func() {
				suite.app.EvmKeeper.SetBalance(suite.ctx, suite.address, big.NewInt(100))
			},
			big.NewInt(100),
		},
		{
			"sub balance",
			func() {
				suite.app.EvmKeeper.SubBalance(suite.ctx, suite.address, big.NewInt(100))
			},
			big.NewInt(0),
		},
		{
			"add balance",
			func() {
				suite.app.EvmKeeper.AddBalance(suite.ctx, suite.address, big.NewInt(200))
			},
			big.NewInt(200),
		},
		{
			"sub more than balance",
			func() {
				suite.app.EvmKeeper.SubBalance(suite.ctx, suite.address, big.NewInt(300))
			},
			big.NewInt(-100),
		},
	}

	for _, tc := range testCase {
		tc.malleate()
		suite.Require().Equal(tc.balance, suite.app.EvmKeeper.GetBalance(suite.ctx, suite.address), tc.name)

	}
}

func (suite *KeeperTestSuite) TestStateDBNonce() {
	nonce := uint64(123)
	suite.app.EvmKeeper.SetNonce(suite.ctx, suite.address, nonce)
	suite.Require().Equal(nonce, suite.app.EvmKeeper.GetNonce(suite.ctx, suite.address))
}

func (suite *KeeperTestSuite) TestStateDBState() {
	key := ethcmn.BytesToHash([]byte("foo"))
	val := ethcmn.BytesToHash([]byte("bar"))

	suite.app.EvmKeeper.SetState(suite.ctx, suite.address, key, val)
	suite.Require().Equal(val, suite.app.EvmKeeper.GetState(suite.ctx, suite.address, key))
}

func (suite *KeeperTestSuite) TestStateDBCode() {
	code := []byte("foobar")

	suite.app.EvmKeeper.SetCode(suite.ctx, suite.address, code)

	suite.Require().Equal(code, suite.app.EvmKeeper.GetCode(suite.ctx, suite.address))

	codelen := len(code)
	suite.Require().Equal(codelen, suite.app.EvmKeeper.GetCodeSize(suite.ctx, suite.address))
}

func (suite *KeeperTestSuite) TestStateDBLogs() {
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

		err := suite.app.EvmKeeper.SetLogs(suite.ctx, hash, logs)
		suite.Require().NoError(err, tc.name)
		dbLogs, err := suite.app.EvmKeeper.GetLogs(suite.ctx, hash)
		suite.Require().NoError(err, tc.name)
		suite.Require().Equal(logs, dbLogs, tc.name)
		suite.Require().Equal(logs, suite.app.EvmKeeper.AllLogs(suite.ctx), tc.name)

		//resets state but checking to see if storekey still persists.
		err = suite.app.EvmKeeper.Reset(suite.ctx, hash)
		suite.Require().NoError(err, tc.name)
		suite.Require().Equal(logs, suite.app.EvmKeeper.AllLogs(suite.ctx), tc.name)
	}
}

func (suite *KeeperTestSuite) TestStateDBPreimage() {
	hash := ethcmn.BytesToHash([]byte("hash"))
	preimage := []byte("preimage")

	suite.app.EvmKeeper.AddPreimage(suite.ctx, hash, preimage)

	suite.Require().Equal(preimage, suite.app.EvmKeeper.Preimages(suite.ctx)[hash])
}

func (suite *KeeperTestSuite) TestStateDBRefund() {
	testCase := []struct {
		name   string
		amount uint64
	}{
		{
			"refund",
			100,
		},
	}

	for _, tc := range testCase {
		suite.app.EvmKeeper.AddRefund(suite.ctx, tc.amount)
		suite.Require().Equal(tc.amount, suite.app.EvmKeeper.GetRefund(suite.ctx), tc.name)

		suite.app.EvmKeeper.SubRefund(suite.ctx, tc.amount)
		suite.Require().Equal(uint64(0), suite.app.EvmKeeper.GetRefund(suite.ctx), tc.name)
	}
}

func (suite *KeeperTestSuite) TestStateDBCreateAcct() {
	suite.app.EvmKeeper.CreateAccount(suite.ctx, suite.address)
	suite.Require().True(suite.app.EvmKeeper.Exist(suite.ctx, suite.address))

	value := big.NewInt(100)
	suite.app.EvmKeeper.AddBalance(suite.ctx, suite.address, value)

	suite.app.EvmKeeper.CreateAccount(suite.ctx, suite.address)
	suite.Require().Equal(value, suite.app.EvmKeeper.GetBalance(suite.ctx, suite.address))
}

func (suite *KeeperTestSuite) TestStateDBClearStateOjb() {

	suite.app.EvmKeeper.CreateAccount(suite.ctx, suite.address)
	suite.Require().True(suite.app.EvmKeeper.Exist(suite.ctx, suite.address))

	suite.app.EvmKeeper.ClearStateObjects(suite.ctx)
	suite.Require().False(suite.app.EvmKeeper.Exist(suite.ctx, suite.address))
}

func (suite *KeeperTestSuite) TestStateDBReset() {
	hash := ethcmn.BytesToHash([]byte("hash"))

	suite.app.EvmKeeper.CreateAccount(suite.ctx, suite.address)
	suite.Require().True(suite.app.EvmKeeper.Exist(suite.ctx, suite.address))

	err := suite.app.EvmKeeper.Reset(suite.ctx, hash)
	suite.Require().NoError(err)
	suite.Require().False(suite.app.EvmKeeper.Exist(suite.ctx, suite.address))

}

func (suite *KeeperTestSuite) TestStateDBUpdateAcct() {

}

func (suite *KeeperTestSuite) TestSuiteDBPrepare() {
	thash := ethcmn.BytesToHash([]byte("thash"))
	bhash := ethcmn.BytesToHash([]byte("bhash"))
	txi := 1

	suite.app.EvmKeeper.Prepare(suite.ctx, thash, bhash, txi)

	suite.Require().Equal(txi, suite.app.EvmKeeper.TxIndex(suite.ctx))
	suite.Require().Equal(bhash, suite.app.EvmKeeper.BlockHash(suite.ctx))

}

func (suite *KeeperTestSuite) TestSuiteDBCopyState() {
	copyDB := suite.app.EvmKeeper.Copy(suite.ctx)
	suite.Require().Equal(suite.app.EvmKeeper.Exist(suite.ctx, suite.address), copyDB.Exist(suite.address))
}

func (suite *KeeperTestSuite) TestSuiteDBEmpty() {
	suite.Require().True(suite.app.EvmKeeper.Empty(suite.ctx, suite.address))

	suite.app.EvmKeeper.SetBalance(suite.ctx, suite.address, big.NewInt(100))

	suite.Require().False(suite.app.EvmKeeper.Empty(suite.ctx, suite.address))
}

func (suite *KeeperTestSuite) TestSuiteDBSuicide() {

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
			_, err := suite.app.EvmKeeper.Commit(suite.ctx, tc.delete)
			suite.Require().NoError(err, tc.name)
			suite.Require().False(suite.app.EvmKeeper.Exist(suite.ctx, suite.address), tc.name)
			continue
		}

		if tc.expPass {
			suite.app.EvmKeeper.SetBalance(suite.ctx, suite.address, tc.amount)
			suicide := suite.app.EvmKeeper.Suicide(suite.ctx, suite.address)
			suite.Require().True(suicide, tc.name)
			suite.Require().True(suite.app.EvmKeeper.HasSuicided(suite.ctx, suite.address), tc.name)
		} else {
			//Suicide only works for an account with non-zero balance/nonce
			priv, err := crypto.GenerateKey()
			suite.Require().NoError(err)

			addr := ethcrypto.PubkeyToAddress(priv.ToECDSA().PublicKey)
			suicide := suite.app.EvmKeeper.Suicide(suite.ctx, addr)
			suite.Require().False(suicide, tc.name)
			suite.Require().False(suite.app.EvmKeeper.HasSuicided(suite.ctx, addr), tc.name)
		}
	}
}

func (suite *KeeperTestSuite) TestCommitStateDB_Commit() {
	suite.app.EvmKeeper.AddBalance(suite.ctx, suite.address, big.NewInt(100))
	testCase := []struct {
		name       string
		malleate   func()
		deleteObjs bool
		expPass    bool
	}{
		{
			"commit suicided",
			func() {
				ok := suite.app.EvmKeeper.Suicide(suite.ctx, suite.address)
				suite.Require().True(ok)
			},
			true, true,
		},
		{
			"commit with dirty value",
			func() {
				suite.app.EvmKeeper.SetCode(suite.ctx, suite.address, []byte("code"))
			},
			false, true,
		},
		{
			"faled to update state object",
			func() {
				suite.app.EvmKeeper.SubBalance(suite.ctx, suite.address, big.NewInt(10))
			},
			false, false,
		},
	}

	for _, tc := range testCase {
		tc.malleate()

		hash, err := suite.app.EvmKeeper.Commit(suite.ctx, tc.deleteObjs)
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

func (suite *KeeperTestSuite) TestCommitStateDB_Finalize() {
	suite.app.EvmKeeper.AddBalance(suite.ctx, suite.address, big.NewInt(100))
	testCase := []struct {
		name       string
		malleate   func()
		deleteObjs bool
		expPass    bool
	}{
		{
			"finalize suicided",
			func() {
				ok := suite.app.EvmKeeper.Suicide(suite.ctx, suite.address)
				suite.Require().True(ok)
			},
			true, true,
		},
		{
			"finalize, not suicided",
			func() {
				suite.app.EvmKeeper.AddBalance(suite.ctx, suite.address, big.NewInt(5))
			},
			false, true,
		},
		{
			"finalize, dirty storage",
			func() {
				suite.app.EvmKeeper.SetState(suite.ctx, suite.address, ethcmn.BytesToHash([]byte("key")), ethcmn.BytesToHash([]byte("value")))
			},
			false, true,
		},
		{
			"faled to update state object",
			func() {
				suite.app.EvmKeeper.SubBalance(suite.ctx, suite.address, big.NewInt(10))
			},
			false, false,
		},
	}

	for _, tc := range testCase {
		tc.malleate()

		err := suite.app.EvmKeeper.Finalise(suite.ctx, tc.deleteObjs)

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
