package keeper_test

import (
	"fmt"
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"

	ethcmn "github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"

	"github.com/cosmos/ethermint/crypto/ethsecp256k1"
	ethermint "github.com/cosmos/ethermint/types"
	"github.com/cosmos/ethermint/x/evm/types"
)

func (suite *KeeperTestSuite) TestBloomFilter() {
	// Prepare db for logs
	tHash := ethcmn.BytesToHash([]byte{0x1})
	suite.app.EvmKeeper.Prepare(suite.ctx, tHash, 0)
	contractAddress := ethcmn.BigToAddress(big.NewInt(1))
	log := ethtypes.Log{Address: contractAddress, Topics: []ethcmn.Hash{}}

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
			bloomFilter := ethtypes.BytesToBloom(bloomInt)
			suite.Require().True(ethtypes.BloomLookup(bloomFilter, contractAddress), tc.name)
			suite.Require().False(ethtypes.BloomLookup(bloomFilter, ethcmn.BigToAddress(big.NewInt(2))), tc.name)
		}
	}
}

func (suite *KeeperTestSuite) TestStateDB_Balance() {
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

func (suite *KeeperTestSuite) TestStateDB_Error() {
	nonce := suite.app.EvmKeeper.GetNonce(suite.ctx, ethcmn.Address{})
	suite.Require().Equal(0, int(nonce))
	suite.Require().Error(suite.app.EvmKeeper.Error(suite.ctx))
}

func (suite *KeeperTestSuite) TestStateDB_Database() {
	suite.Require().Nil(suite.app.EvmKeeper.Database(suite.ctx))
}

func (suite *KeeperTestSuite) TestStateDB_State() {
	key := ethcmn.BytesToHash([]byte("foo"))
	val := ethcmn.BytesToHash([]byte("bar"))
	suite.app.EvmKeeper.SetState(suite.ctx, suite.address, key, val)

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
		value := suite.app.EvmKeeper.GetState(suite.ctx, tc.address, tc.key)
		suite.Require().Equal(tc.value, value, tc.name)
	}
}

func (suite *KeeperTestSuite) TestStateDB_Code() {
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
				suite.app.EvmKeeper.SetCode(suite.ctx, suite.address, []byte("code"))
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

		suite.Require().Equal(tc.code, suite.app.EvmKeeper.GetCode(suite.ctx, tc.address), tc.name)
		suite.Require().Equal(len(tc.code), suite.app.EvmKeeper.GetCodeSize(suite.ctx, tc.address), tc.name)
	}
}

func (suite *KeeperTestSuite) TestStateDB_Logs() {
	testCase := []struct {
		name string
		log  *ethtypes.Log
	}{
		{
			"state db log",
			&ethtypes.Log{
				Address:     suite.address,
				Topics:      []ethcmn.Hash{ethcmn.BytesToHash([]byte("topic"))},
				Data:        []byte("data"),
				BlockNumber: 1,
				TxHash:      ethcmn.Hash{},
				TxIndex:     1,
				BlockHash:   ethcmn.Hash{},
				Index:       0,
				Removed:     false,
			},
		},
	}

	for _, tc := range testCase {
		hash := ethcmn.BytesToHash([]byte("hash"))
		logs := []*ethtypes.Log{tc.log}

		err := suite.app.EvmKeeper.SetLogs(suite.ctx, hash, logs)
		suite.Require().NoError(err, tc.name)
		dbLogs, err := suite.app.EvmKeeper.GetLogs(suite.ctx, hash)
		suite.Require().NoError(err, tc.name)
		suite.Require().Equal(logs, dbLogs, tc.name)

		suite.app.EvmKeeper.DeleteLogs(suite.ctx, hash)
		dbLogs, err = suite.app.EvmKeeper.GetLogs(suite.ctx, hash)
		suite.Require().NoError(err, tc.name)
		suite.Require().Empty(dbLogs, tc.name)

		suite.app.EvmKeeper.AddLog(suite.ctx, tc.log)
		tc.log.Index = 0 // reset index
		suite.Require().Equal(logs, suite.app.EvmKeeper.AllLogs(suite.ctx), tc.name)

		//resets state but checking to see if storekey still persists.
		err = suite.app.EvmKeeper.Reset(suite.ctx, hash)
		suite.Require().NoError(err, tc.name)
		suite.Require().Equal(logs, suite.app.EvmKeeper.AllLogs(suite.ctx), tc.name)
	}
}

func (suite *KeeperTestSuite) TestStateDB_Preimage() {
	hash := ethcmn.BytesToHash([]byte("hash"))
	preimage := []byte("preimage")

	suite.app.EvmKeeper.AddPreimage(suite.ctx, hash, preimage)
	suite.Require().Equal(preimage, suite.app.EvmKeeper.Preimages(suite.ctx)[hash])
}

func (suite *KeeperTestSuite) TestStateDB_Refund() {
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

			suite.app.EvmKeeper.AddRefund(suite.ctx, tc.addAmount)
			suite.Require().Equal(tc.addAmount, suite.app.EvmKeeper.GetRefund(suite.ctx))

			if tc.expPanic {
				suite.Panics(func() {
					suite.app.EvmKeeper.SubRefund(suite.ctx, tc.subAmount)
				})
			} else {
				suite.app.EvmKeeper.SubRefund(suite.ctx, tc.subAmount)
				suite.Require().Equal(tc.expRefund, suite.app.EvmKeeper.GetRefund(suite.ctx))
			}
		})
	}
}

func (suite *KeeperTestSuite) TestStateDB_CreateAccount() {
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
				suite.app.EvmKeeper.AddBalance(suite.ctx, suite.address, prevBalance)
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
		suite.Run(tc.name, func() {
			suite.SetupTest() // reset
			tc.malleate()

			suite.app.EvmKeeper.CreateAccount(suite.ctx, tc.address)
			suite.Require().True(suite.app.EvmKeeper.Exist(suite.ctx, tc.address))
			suite.Require().Equal(prevBalance, suite.app.EvmKeeper.GetBalance(suite.ctx, tc.address))
		})
	}
}

func (suite *KeeperTestSuite) TestStateDB_ClearStateObj() {
	priv, err := ethsecp256k1.GenerateKey()
	suite.Require().NoError(err)

	addr := ethcrypto.PubkeyToAddress(priv.ToECDSA().PublicKey)

	suite.app.EvmKeeper.CreateAccount(suite.ctx, addr)
	suite.Require().True(suite.app.EvmKeeper.Exist(suite.ctx, addr))

	suite.app.EvmKeeper.ClearStateObjects(suite.ctx)
	suite.Require().False(suite.app.EvmKeeper.Exist(suite.ctx, addr))
}

func (suite *KeeperTestSuite) TestStateDB_Reset() {
	priv, err := ethsecp256k1.GenerateKey()
	suite.Require().NoError(err)

	addr := ethcrypto.PubkeyToAddress(priv.ToECDSA().PublicKey)

	suite.app.EvmKeeper.CreateAccount(suite.ctx, addr)
	suite.Require().True(suite.app.EvmKeeper.Exist(suite.ctx, addr))

	err = suite.app.EvmKeeper.Reset(suite.ctx, ethcmn.BytesToHash(nil))
	suite.Require().NoError(err)
	suite.Require().False(suite.app.EvmKeeper.Exist(suite.ctx, addr))
}

func (suite *KeeperTestSuite) TestSuiteDB_Prepare() {
	thash := ethcmn.BytesToHash([]byte("thash"))
	bhash := ethcmn.BytesToHash([]byte("bhash"))
	txi := 1

	suite.app.EvmKeeper.Prepare(suite.ctx, thash, txi)
	suite.app.EvmKeeper.CommitStateDB.SetBlockHash(bhash)

	suite.Require().Equal(txi, suite.app.EvmKeeper.TxIndex(suite.ctx))
	suite.Require().Equal(bhash, suite.app.EvmKeeper.BlockHash(suite.ctx))
}

func (suite *KeeperTestSuite) TestSuiteDB_CopyState() {
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
				Index:       0,
				Removed:     false,
			},
		},
	}

	for _, tc := range testCase {
		hash := ethcmn.BytesToHash([]byte("hash"))
		logs := []*ethtypes.Log{&tc.log}

		err := suite.app.EvmKeeper.SetLogs(suite.ctx, hash, logs)
		suite.Require().NoError(err, tc.name)

		copyDB := suite.app.EvmKeeper.Copy(suite.ctx)
		suite.Require().Equal(suite.app.EvmKeeper.Exist(suite.ctx, suite.address), copyDB.Exist(suite.address), tc.name)
	}
}

func (suite *KeeperTestSuite) TestSuiteDB_Empty() {
	suite.Require().True(suite.app.EvmKeeper.Empty(suite.ctx, suite.address))

	suite.app.EvmKeeper.SetBalance(suite.ctx, suite.address, big.NewInt(100))
	suite.Require().False(suite.app.EvmKeeper.Empty(suite.ctx, suite.address))
}

func (suite *KeeperTestSuite) TestSuiteDB_Suicide() {
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
			priv, err := ethsecp256k1.GenerateKey()
			suite.Require().NoError(err)

			addr := ethcrypto.PubkeyToAddress(priv.ToECDSA().PublicKey)
			suicide := suite.app.EvmKeeper.Suicide(suite.ctx, addr)
			suite.Require().False(suicide, tc.name)
			suite.Require().False(suite.app.EvmKeeper.HasSuicided(suite.ctx, addr), tc.name)
		}
	}
}

func (suite *KeeperTestSuite) TestCommitStateDB_Commit() {
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
	}

	for _, tc := range testCase {
		tc.malleate()

		err := suite.app.EvmKeeper.Finalise(suite.ctx, tc.deleteObjs)

		if !tc.expPass {
			suite.Require().Error(err, tc.name)
			hash := suite.app.EvmKeeper.GetCommittedState(suite.ctx, suite.address, ethcmn.BytesToHash([]byte("key")))
			suite.Require().NotEqual(ethcmn.Hash{}, hash, tc.name)
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
func (suite *KeeperTestSuite) TestCommitStateDB_GetCommittedState() {
	hash := suite.app.EvmKeeper.GetCommittedState(suite.ctx, ethcmn.Address{}, ethcmn.BytesToHash([]byte("key")))
	suite.Require().Equal(ethcmn.Hash{}, hash)
}

func (suite *KeeperTestSuite) TestCommitStateDB_Snapshot() {
	id := suite.app.EvmKeeper.Snapshot(suite.ctx)
	suite.Require().NotPanics(func() {
		suite.app.EvmKeeper.RevertToSnapshot(suite.ctx, id)
	})

	suite.Require().Panics(func() {
		suite.app.EvmKeeper.RevertToSnapshot(suite.ctx, -1)
	}, "invalid revision should panic")
}

func (suite *KeeperTestSuite) TestCommitStateDB_ForEachStorage() {
	var storage types.Storage

	testCase := []struct {
		name      string
		malleate  func()
		callback  func(key, value ethcmn.Hash) (stop bool)
		expValues []string
	}{
		{
			"aggregate state",
			func() {
				for i := 0; i < 5; i++ {
					suite.app.EvmKeeper.SetState(suite.ctx, suite.address, ethcmn.BytesToHash([]byte(fmt.Sprintf("key%d", i))), ethcmn.BytesToHash([]byte(fmt.Sprintf("value%d", i))))
				}
			},
			func(key, value ethcmn.Hash) bool {
				storage = append(storage, types.NewState(key, value))
				return false
			},
			[]string{
				ethcmn.BytesToHash([]byte("value0")).String(),
				ethcmn.BytesToHash([]byte("value1")).String(),
				ethcmn.BytesToHash([]byte("value2")).String(),
				ethcmn.BytesToHash([]byte("value3")).String(),
				ethcmn.BytesToHash([]byte("value4")).String(),
			},
		},
		{
			"filter state",
			func() {
				suite.app.EvmKeeper.SetState(suite.ctx, suite.address, ethcmn.BytesToHash([]byte("key")), ethcmn.BytesToHash([]byte("value")))
				suite.app.EvmKeeper.SetState(suite.ctx, suite.address, ethcmn.BytesToHash([]byte("filterkey")), ethcmn.BytesToHash([]byte("filtervalue")))
			},
			func(key, value ethcmn.Hash) bool {
				if value == ethcmn.BytesToHash([]byte("filtervalue")) {
					storage = append(storage, types.NewState(key, value))
					return true
				}
				return false
			},
			[]string{
				ethcmn.BytesToHash([]byte("filtervalue")).String(),
			},
		},
	}

	for _, tc := range testCase {
		suite.Run(tc.name, func() {
			suite.SetupTest() // reset
			tc.malleate()
			suite.app.EvmKeeper.Finalise(suite.ctx, false)

			err := suite.app.EvmKeeper.ForEachStorage(suite.ctx, suite.address, tc.callback)
			suite.Require().NoError(err)
			suite.Require().Equal(len(tc.expValues), len(storage), fmt.Sprintf("Expected values:\n%v\nStorage Values\n%v", tc.expValues, storage))

			vals := make([]string, len(storage))
			for i := range storage {
				vals[i] = storage[i].Value
			}

			suite.Require().ElementsMatch(tc.expValues, vals)
		})
		storage = types.Storage{}
	}
}
