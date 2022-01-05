package statedb_test

import (
	"math/big"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/suite"
	"github.com/tharsis/ethermint/x/evm/statedb"
)

type StateDBTestSuite struct {
	suite.Suite
}

func (suite *StateDBTestSuite) TestAccounts() {
	addrErr := common.BigToAddress(big.NewInt(1))
	addr2 := common.BigToAddress(big.NewInt(2))
	testTxConfig := statedb.NewTxConfig(
		common.BigToHash(big.NewInt(10)), // tx hash
		common.BigToHash(big.NewInt(11)), // block hash
		1,                                // txIndex
		1,                                // logSize
	)

	testCases := []struct {
		msg  string
		test func(*statedb.StateDB)
	}{
		{
			"success,empty account",
			func(db *statedb.StateDB) {
				suite.Require().Equal(true, db.Empty(addr2))
				suite.Require().Equal(big.NewInt(0), db.GetBalance(addr2))
				suite.Require().Equal([]byte(nil), db.GetCode(addr2))
				suite.Require().Equal(uint64(0), db.GetNonce(addr2))
			},
		},
		{
			"success,GetBalance",
			func(db *statedb.StateDB) {
				db.AddBalance(addr2, big.NewInt(1))
				suite.Require().Equal(big.NewInt(1), db.GetBalance(addr2))
			},
		},
		{
			"fail,GetBalance dbErr",
			func(db *statedb.StateDB) {
				suite.Require().Equal(big.NewInt(0), db.GetBalance(addrErr))
				suite.Require().Error(db.Commit())
			},
		},
		{
			"success,change balance",
			func(db *statedb.StateDB) {
				db.AddBalance(addr2, big.NewInt(2))
				suite.Require().Equal(big.NewInt(2), db.GetBalance(addr2))
				db.SubBalance(addr2, big.NewInt(1))
				suite.Require().Equal(big.NewInt(1), db.GetBalance(addr2))

				suite.Require().NoError(db.Commit())

				// create a clean StateDB, check the balance is committed
				db = statedb.New(db.Context(), db.Keeper(), testTxConfig)
				suite.Require().Equal(big.NewInt(1), db.GetBalance(addr2))
			},
		},
		{
			"success,SetState",
			func(db *statedb.StateDB) {
				key := common.BigToHash(big.NewInt(1))
				value := common.BigToHash(big.NewInt(1))

				suite.Require().Equal(common.Hash{}, db.GetState(addr2, key))
				db.SetState(addr2, key, value)
				suite.Require().Equal(value, db.GetState(addr2, key))
				suite.Require().Equal(common.Hash{}, db.GetCommittedState(addr2, key))
			},
		},
		{
			"success,SetCode",
			func(db *statedb.StateDB) {
				code := []byte("hello world")
				codeHash := crypto.Keccak256Hash(code)
				db.SetCode(addr2, code)
				suite.Require().Equal(code, db.GetCode(addr2))
				suite.Require().Equal(codeHash, db.GetCodeHash(addr2))

				suite.Require().NoError(db.Commit())

				// create a clean StateDB, check the code is committed
				db = statedb.New(db.Context(), db.Keeper(), testTxConfig)
				suite.Require().Equal(code, db.GetCode(addr2))
				suite.Require().Equal(codeHash, db.GetCodeHash(addr2))
			},
		},
		{
			"success,CreateAccount",
			func(db *statedb.StateDB) {
				// test balance carry over when overwritten
				amount := big.NewInt(1)
				code := []byte("hello world")
				key := common.BigToHash(big.NewInt(1))
				value := common.BigToHash(big.NewInt(1))

				db.AddBalance(addr2, amount)
				db.SetCode(addr2, code)
				db.SetState(addr2, key, value)

				rev := db.Snapshot()

				db.CreateAccount(addr2)
				suite.Require().Equal(amount, db.GetBalance(addr2))
				suite.Require().Equal([]byte(nil), db.GetCode(addr2))
				suite.Require().Equal(common.Hash{}, db.GetState(addr2, key))

				db.RevertToSnapshot(rev)
				suite.Require().Equal(amount, db.GetBalance(addr2))
				suite.Require().Equal(code, db.GetCode(addr2))
				suite.Require().Equal(value, db.GetState(addr2, key))

				db.CreateAccount(addr2)
				suite.Require().NoError(db.Commit())
				db = statedb.New(db.Context(), db.Keeper(), testTxConfig)
				suite.Require().Equal(amount, db.GetBalance(addr2))
				suite.Require().Equal([]byte(nil), db.GetCode(addr2))
				suite.Require().Equal(common.Hash{}, db.GetState(addr2, key))
			},
		},
		{
			"success,nested snapshot revert",
			func(db *statedb.StateDB) {
				key := common.BigToHash(big.NewInt(1))
				value1 := common.BigToHash(big.NewInt(1))
				value2 := common.BigToHash(big.NewInt(2))

				rev1 := db.Snapshot()
				db.SetState(addr2, key, value1)

				rev2 := db.Snapshot()
				db.SetState(addr2, key, value2)
				suite.Require().Equal(value2, db.GetState(addr2, key))

				db.RevertToSnapshot(rev2)
				suite.Require().Equal(value1, db.GetState(addr2, key))

				db.RevertToSnapshot(rev1)
				suite.Require().Equal(common.Hash{}, db.GetState(addr2, key))
			},
		},
		{
			"success,nonce",
			func(db *statedb.StateDB) {
				suite.Require().Equal(uint64(0), db.GetNonce(addr2))
				db.SetNonce(addr2, 1)
				suite.Require().Equal(uint64(1), db.GetNonce(addr2))

				suite.Require().NoError(db.Commit())

				db = statedb.New(db.Context(), db.Keeper(), testTxConfig)
				suite.Require().Equal(uint64(1), db.GetNonce(addr2))
			},
		},
		{
			"success,logs",
			func(db *statedb.StateDB) {
				data := []byte("hello world")
				db.AddLog(&ethtypes.Log{
					Address:     addr2,
					Topics:      []common.Hash{},
					Data:        data,
					BlockNumber: 1,
				})
				suite.Require().Equal(1, len(db.Logs()))
				expecedLog := &ethtypes.Log{
					Address:     addr2,
					Topics:      []common.Hash{},
					Data:        data,
					BlockNumber: 1,
					BlockHash:   common.BigToHash(big.NewInt(10)),
					TxHash:      common.BigToHash(big.NewInt(11)),
					TxIndex:     1,
					Index:       1,
				}
				suite.Require().Equal(expecedLog, db.Logs()[0])

				rev := db.Snapshot()

				db.AddLog(&ethtypes.Log{
					Address:     addr2,
					Topics:      []common.Hash{},
					Data:        data,
					BlockNumber: 1,
				})
				suite.Require().Equal(2, len(db.Logs()))
				suite.Require().Equal(uint(2), db.Logs()[1].Index)

				db.RevertToSnapshot(rev)
				suite.Require().Equal(1, len(db.Logs()))
			},
		},
		{
			"success,refund",
			func(db *statedb.StateDB) {
				db.AddRefund(uint64(10))
				suite.Require().Equal(uint64(10), db.GetRefund())

				rev := db.Snapshot()

				db.SubRefund(uint64(5))
				suite.Require().Equal(uint64(5), db.GetRefund())

				db.RevertToSnapshot(rev)
				suite.Require().Equal(uint64(10), db.GetRefund())
			},
		},
		{
			"success,empty",
			func(db *statedb.StateDB) {
				suite.Require().False(db.Exist(addr2))
				suite.Require().True(db.Empty(addr2))

				db.AddBalance(addr2, big.NewInt(1))
				suite.Require().True(db.Exist(addr2))
				suite.Require().False(db.Empty(addr2))

				db.SubBalance(addr2, big.NewInt(1))
				suite.Require().True(db.Exist(addr2))
				suite.Require().True(db.Empty(addr2))
			},
		},
		{
			"success,suicide commit",
			func(db *statedb.StateDB) {
				code := []byte("hello world")
				db.SetCode(addr2, code)
				db.AddBalance(addr2, big.NewInt(1))

				suite.Require().True(db.Exist(addr2))
				suite.Require().False(db.Empty(addr2))

				db.Suicide(addr2)
				suite.Require().True(db.HasSuicided(addr2))
				suite.Require().True(db.Exist(addr2))
				suite.Require().Equal(new(big.Int), db.GetBalance(addr2))

				suite.Require().NoError(db.Commit())
				db = statedb.New(db.Context(), db.Keeper(), testTxConfig)
				suite.Require().True(db.Empty(addr2))
			},
		},
		{
			"success,suicide revert",
			func(db *statedb.StateDB) {
				code := []byte("hello world")
				db.SetCode(addr2, code)
				db.AddBalance(addr2, big.NewInt(1))

				rev := db.Snapshot()

				db.Suicide(addr2)
				suite.Require().True(db.HasSuicided(addr2))

				db.RevertToSnapshot(rev)

				suite.Require().False(db.HasSuicided(addr2))
				suite.Require().Equal(code, db.GetCode(addr2))
				suite.Require().Equal(big.NewInt(1), db.GetBalance(addr2))
			},
		},
		// TODO access lisForEachStorage
		// https://github.com/tharsis/ethermint/issues/876
	}
	for _, tc := range testCases {
		suite.Run(tc.msg, func() {
			db := statedb.New(
				sdk.Context{},
				NewMockKeeper(),
				testTxConfig,
			)
			tc.test(db)
		})
	}
}

func TestStateDBTestSuite(t *testing.T) {
	suite.Run(t, &StateDBTestSuite{})
}
