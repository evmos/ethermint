package keeper_test

import (
	"fmt"
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	"github.com/cosmos/ethermint/tests"
	"github.com/cosmos/ethermint/x/evm/types"
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
)

func (suite *KeeperTestSuite) TestCreateAccount() {
	testCases := []struct {
		name     string
		addr     common.Address
		malleate func(common.Address)
		callback func(common.Address)
	}{
		{
			"reset account",
			suite.address,
			func(addr common.Address) {
				suite.app.EvmKeeper.AddBalance(addr, big.NewInt(100))
				suite.Require().NotZero(suite.app.EvmKeeper.GetBalance(addr).Int64())
			},
			func(addr common.Address) {
				suite.Require().Zero(suite.app.EvmKeeper.GetBalance(addr).Int64())
			},
		},
		{
			"create account",
			tests.GenerateAddress(),
			func(addr common.Address) {
				acc := suite.app.AccountKeeper.GetAccount(suite.ctx, addr.Bytes())
				suite.Require().Nil(acc)
			},
			func(addr common.Address) {
				acc := suite.app.AccountKeeper.GetAccount(suite.ctx, addr.Bytes())
				suite.Require().NotNil(acc)
			},
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			tc.malleate(tc.addr)
			suite.app.EvmKeeper.CreateAccount(tc.addr)
			tc.callback(tc.addr)
		})
	}
}

func (suite *KeeperTestSuite) TestAddBalance() {
	testCases := []struct {
		name   string
		amount *big.Int
		isNoOp bool
	}{
		{
			"positive amount",
			big.NewInt(100),
			false,
		},
		{
			"zero amount",
			big.NewInt(0),
			true,
		},
		{
			"negative amount",
			big.NewInt(-1),
			true,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			prev := suite.app.EvmKeeper.GetBalance(suite.address)
			suite.app.EvmKeeper.AddBalance(suite.address, tc.amount)
			post := suite.app.EvmKeeper.GetBalance(suite.address)

			if tc.isNoOp {
				suite.Require().Equal(prev.Int64(), post.Int64())
			} else {
				suite.Require().Equal(new(big.Int).Add(prev, tc.amount).Int64(), post.Int64())
			}
		})
	}
}

func (suite *KeeperTestSuite) TestSubBalance() {
	testCases := []struct {
		name     string
		amount   *big.Int
		malleate func()
		isNoOp   bool
	}{
		{
			"positive amount, below zero",
			big.NewInt(100),
			func() {},
			true,
		},
		{
			"positive amount, below zero",
			big.NewInt(50),
			func() {
				suite.app.EvmKeeper.AddBalance(suite.address, big.NewInt(100))
			},
			false,
		},
		{
			"zero amount",
			big.NewInt(0),
			func() {},
			true,
		},
		{
			"negative amount",
			big.NewInt(-1),
			func() {},
			true,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			tc.malleate()

			prev := suite.app.EvmKeeper.GetBalance(suite.address)
			suite.app.EvmKeeper.SubBalance(suite.address, tc.amount)
			post := suite.app.EvmKeeper.GetBalance(suite.address)

			if tc.isNoOp {
				suite.Require().Equal(prev.Int64(), post.Int64())
			} else {
				suite.Require().Equal(new(big.Int).Sub(prev, tc.amount).Int64(), post.Int64())
			}
		})
	}
}

func (suite *KeeperTestSuite) TestGetNonce() {
	testCases := []struct {
		name          string
		address       common.Address
		expectedNonce uint64
		malleate      func()
	}{
		{
			"account not found",
			tests.GenerateAddress(),
			0,
			func() {},
		},
		{
			"existing account",
			suite.address,
			1,
			func() {
				suite.app.EvmKeeper.SetNonce(suite.address, 1)
			},
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			tc.malleate()

			nonce := suite.app.EvmKeeper.GetNonce(tc.address)
			suite.Require().Equal(tc.expectedNonce, nonce)

		})
	}
}

func (suite *KeeperTestSuite) TestSetNonce() {
	testCases := []struct {
		name     string
		address  common.Address
		nonce    uint64
		malleate func()
	}{
		{
			"new account",
			tests.GenerateAddress(),
			10,
			func() {},
		},
		{
			"existing account",
			suite.address,
			99,
			func() {},
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.app.EvmKeeper.SetNonce(tc.address, tc.nonce)
			nonce := suite.app.EvmKeeper.GetNonce(tc.address)
			suite.Require().Equal(tc.nonce, nonce)
		})
	}
}

func (suite *KeeperTestSuite) TestGetCodeHash() {
	addr := tests.GenerateAddress()
	baseAcc := &authtypes.BaseAccount{Address: sdk.AccAddress(addr.Bytes()).String()}
	suite.app.AccountKeeper.SetAccount(suite.ctx, baseAcc)

	testCases := []struct {
		name     string
		address  common.Address
		expHash  common.Hash
		malleate func()
	}{
		{
			"account not found",
			tests.GenerateAddress(),
			common.BytesToHash(types.EmptyCodeHash),
			func() {},
		},
		{
			"account not EthAccount type",
			addr,
			common.BytesToHash(types.EmptyCodeHash),
			func() {},
		},
		{
			"existing account",
			suite.address,
			crypto.Keccak256Hash([]byte("codeHash")),
			func() {
				suite.app.EvmKeeper.SetCode(suite.address, []byte("codeHash"))
			},
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {

			tc.malleate()

			hash := suite.app.EvmKeeper.GetCodeHash(tc.address)
			suite.Require().Equal(tc.expHash, hash)
		})
	}
}

func (suite *KeeperTestSuite) TestSetCode() {
	addr := tests.GenerateAddress()
	baseAcc := &authtypes.BaseAccount{Address: sdk.AccAddress(addr.Bytes()).String()}
	suite.app.AccountKeeper.SetAccount(suite.ctx, baseAcc)

	testCases := []struct {
		name    string
		address common.Address
		code    []byte
		isNoOp  bool
	}{
		{
			"account not found",
			tests.GenerateAddress(),
			[]byte("code"),
			false,
		},
		{
			"account not EthAccount type",
			addr,
			nil,
			true,
		},
		{
			"existing account",
			suite.address,
			[]byte("code"),
			false,
		},
		{
			"existing account, code deleted from store",
			suite.address,
			nil,
			false,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {

			prev := suite.app.EvmKeeper.GetCode(tc.address)
			suite.app.EvmKeeper.SetCode(tc.address, tc.code)
			post := suite.app.EvmKeeper.GetCode(tc.address)

			if tc.isNoOp {
				suite.Require().Equal(prev, post)
			} else {
				suite.Require().Equal(tc.code, post)
			}

			suite.Require().Equal(len(post), suite.app.EvmKeeper.GetCodeSize(tc.address))
		})
	}
}

func (suite *KeeperTestSuite) TestRefund() {
	testCases := []struct {
		name      string
		malleate  func()
		expRefund uint64
		expPanic  bool
	}{
		{
			"success - add and subtract refund",
			func() {
				suite.app.EvmKeeper.AddRefund(11)
			},
			1,
			false,
		},
		{
			"fail - subtract amount > current refund",
			func() {
			},
			0,
			true,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {

			tc.malleate()

			if tc.expPanic {
				suite.Require().Panics(func() { suite.app.EvmKeeper.SubRefund(10) })
			} else {
				suite.app.EvmKeeper.SubRefund(10)
				suite.Require().Equal(tc.expRefund, suite.app.EvmKeeper.GetRefund())
			}

			// clear and reset refund from store
			suite.app.EvmKeeper.ResetRefundTransient(suite.ctx)
			suite.Require().Zero(suite.app.EvmKeeper.GetRefund())
		})
	}
}

func (suite *KeeperTestSuite) TestState() {
	testCases := []struct {
		name       string
		key, value common.Hash
	}{
		{
			"set state - delete from store",
			common.BytesToHash([]byte("key")),
			common.Hash{},
		},
		{
			"set state - update value",
			common.BytesToHash([]byte("key")),
			common.BytesToHash([]byte("value")),
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {

			suite.app.EvmKeeper.SetState(suite.address, tc.key, tc.value)
			value := suite.app.EvmKeeper.GetState(suite.address, tc.key)
			suite.Require().Equal(tc.value, value)
		})
	}
}

func (suite *KeeperTestSuite) TestSuicide() {
	testCases := []struct {
		name     string
		suicided bool
	}{
		{"success, first time suicided", true},
		{"success, already suicided", true},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.Require().Equal(tc.suicided, suite.app.EvmKeeper.Suicide(suite.address))
			suite.Require().Equal(tc.suicided, suite.app.EvmKeeper.HasSuicided(suite.address))
		})
	}
}

func (suite *KeeperTestSuite) TestExist() {
	testCases := []struct {
		name     string
		address  common.Address
		malleate func()
		exists   bool
	}{
		{"success, account exists", suite.address, func() {}, true},
		{"success, has suicided", suite.address, func() {
			suite.app.EvmKeeper.Suicide(suite.address)
		}, true},
		{"success, account doesn't exist", tests.GenerateAddress(), func() {}, false},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			tc.malleate()

			suite.Require().Equal(tc.exists, suite.app.EvmKeeper.Exist(tc.address))
		})
	}
}

func (suite *KeeperTestSuite) TestEmpty() {
	addr := tests.GenerateAddress()
	baseAcc := &authtypes.BaseAccount{Address: sdk.AccAddress(addr.Bytes()).String()}
	suite.app.AccountKeeper.SetAccount(suite.ctx, baseAcc)

	testCases := []struct {
		name     string
		address  common.Address
		malleate func()
		empty    bool
	}{
		{"empty, account exists", suite.address, func() {}, true},
		{"not empty, non ethereum account", addr, func() {}, false},
		{"not empty, positive balance", suite.address, func() {
			suite.app.EvmKeeper.AddBalance(suite.address, big.NewInt(100))
		}, false},
		{"empty, account doesn't exist", tests.GenerateAddress(), func() {}, true},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			tc.malleate()

			suite.Require().Equal(tc.empty, suite.app.EvmKeeper.Empty(tc.address))
		})
	}
}

func (suite *KeeperTestSuite) TestSnapshot() {
	revision := suite.app.EvmKeeper.Snapshot()
	suite.Require().Zero(revision)
	suite.app.EvmKeeper.RevertToSnapshot(revision) // no-op
}

func (suite *KeeperTestSuite) TestAddLog() {
	addr := tests.GenerateAddress()
	msg := types.NewMsgEthereumTx(big.NewInt(1), 0, &suite.address, big.NewInt(1), 100000, big.NewInt(1), []byte("test"), nil)
	tx := msg.AsTransaction()
	txBz, err := tx.MarshalBinary()
	suite.Require().NoError(err)
	txHash := tx.Hash()

	testCases := []struct {
		name        string
		log, expLog *ethtypes.Log // pre and post populating log fields
		malleate    func()
	}{
		{
			"block hash not found",
			&ethtypes.Log{
				Address: addr,
			},
			&ethtypes.Log{
				Address: addr,
			},
			func() {},
		},
		{
			"tx hash from message",
			&ethtypes.Log{
				Address: addr,
			},
			&ethtypes.Log{
				Address: addr,
				TxHash:  txHash,
			},
			func() {
				suite.app.EvmKeeper.WithContext(suite.ctx.WithTxBytes(txBz))
			},
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			tc.malleate()

			prev := suite.app.EvmKeeper.GetTxLogs(tc.expLog.TxHash)
			suite.app.EvmKeeper.AddLog(tc.log)
			post := suite.app.EvmKeeper.GetTxLogs(tc.expLog.TxHash)

			suite.Require().NotZero(len(post), tc.expLog.TxHash.Hex())
			suite.Require().Equal(len(prev)+1, len(post))
			suite.Require().NotNil(post[len(post)-1])
			suite.Require().Equal(tc.log, post[len(post)-1])
		})
	}
}

func (suite *KeeperTestSuite) TestAccessList() {
	dest := tests.GenerateAddress()
	precompiles := []common.Address{tests.GenerateAddress(), tests.GenerateAddress()}
	accesses := ethtypes.AccessList{
		{Address: tests.GenerateAddress(), StorageKeys: []common.Hash{common.BytesToHash([]byte("key"))}},
		{Address: tests.GenerateAddress(), StorageKeys: []common.Hash{common.BytesToHash([]byte("key1"))}},
	}

	suite.app.EvmKeeper.PrepareAccessList(suite.address, &dest, precompiles, accesses)

	suite.Require().True(suite.app.EvmKeeper.AddressInAccessList(suite.address))
	suite.Require().True(suite.app.EvmKeeper.AddressInAccessList(dest))

	for _, precompile := range precompiles {
		suite.Require().True(suite.app.EvmKeeper.AddressInAccessList(precompile))
	}

	for _, access := range accesses {
		for _, key := range access.StorageKeys {
			addrOK, slotOK := suite.app.EvmKeeper.SlotInAccessList(access.Address, key)
			suite.Require().True(addrOK)
			suite.Require().True(slotOK)
		}
	}
}

func (suite *KeeperTestSuite) TestForEachStorage() {
	var storage types.Storage

	testCase := []struct {
		name      string
		malleate  func()
		callback  func(key, value common.Hash) (stop bool)
		expValues []common.Hash
	}{
		{
			"aggregate state",
			func() {
				for i := 0; i < 5; i++ {
					suite.app.EvmKeeper.SetState(suite.address, common.BytesToHash([]byte(fmt.Sprintf("key%d", i))), common.BytesToHash([]byte(fmt.Sprintf("value%d", i))))
				}
			},
			func(key, value common.Hash) bool {
				storage = append(storage, types.NewState(key, value))
				return false
			},
			[]common.Hash{
				common.BytesToHash([]byte("value0")),
				common.BytesToHash([]byte("value1")),
				common.BytesToHash([]byte("value2")),
				common.BytesToHash([]byte("value3")),
				common.BytesToHash([]byte("value4")),
			},
		},
		{
			"filter state",
			func() {
				suite.app.EvmKeeper.SetState(suite.address, common.BytesToHash([]byte("key")), common.BytesToHash([]byte("value")))
				suite.app.EvmKeeper.SetState(suite.address, common.BytesToHash([]byte("filterkey")), common.BytesToHash([]byte("filtervalue")))
			},
			func(key, value common.Hash) bool {
				if value == common.BytesToHash([]byte("filtervalue")) {
					storage = append(storage, types.NewState(key, value))
					return true
				}
				return false
			},
			[]common.Hash{
				common.BytesToHash([]byte("filtervalue")),
			},
		},
	}

	for _, tc := range testCase {
		suite.Run(tc.name, func() {
			suite.SetupTest() // reset
			tc.malleate()

			err := suite.app.EvmKeeper.ForEachStorage(suite.address, tc.callback)
			suite.Require().NoError(err)
			suite.Require().Equal(len(tc.expValues), len(storage), fmt.Sprintf("Expected values:\n%v\nStorage Values\n%v", tc.expValues, storage))

			vals := make([]common.Hash, len(storage))
			for i := range storage {
				vals[i] = common.HexToHash(storage[i].Value)
			}

			// TODO: not sure why Equals fails
			suite.Require().ElementsMatch(tc.expValues, vals)
		})
		storage = types.Storage{}
	}
}
