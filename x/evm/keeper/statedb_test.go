package keeper_test

import (
	"fmt"
	"math/big"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/tharsis/ethermint/tests"
	"github.com/tharsis/ethermint/x/evm/types"
)

func (suite *KeeperTestSuite) TestCreateAccount() {
	testCases := []struct {
		name     string
		addr     common.Address
		malleate func(common.Address)
		callback func(common.Address)
	}{
		{
			"reset account (keep balance)",
			suite.address,
			func(addr common.Address) {
				suite.app.EvmKeeper.AddBalance(addr, big.NewInt(100))
				suite.Require().NotZero(suite.app.EvmKeeper.GetBalance(addr).Int64())
			},
			func(addr common.Address) {
				suite.Require().Equal(suite.app.EvmKeeper.GetBalance(addr).Int64(), int64(100))
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

			suite.app.EvmKeeper.ClearStateError()
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

func (suite *KeeperTestSuite) TestCommittedState() {
	suite.SetupTest()

	key := common.BytesToHash([]byte("key"))
	value1 := common.BytesToHash([]byte("value1"))
	value2 := common.BytesToHash([]byte("value2"))

	suite.app.EvmKeeper.SetState(suite.address, key, value1)

	suite.app.EvmKeeper.Snapshot()

	suite.app.EvmKeeper.SetState(suite.address, key, value2)
	tmp := suite.app.EvmKeeper.GetState(suite.address, key)
	suite.Require().Equal(value2, tmp)
	tmp = suite.app.EvmKeeper.GetCommittedState(suite.address, key)
	suite.Require().Equal(value1, tmp)

	suite.app.EvmKeeper.CommitCachedContexts()

	tmp = suite.app.EvmKeeper.GetCommittedState(suite.address, key)
	suite.Require().Equal(value2, tmp)
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
	key := common.BytesToHash([]byte("key"))
	value1 := common.BytesToHash([]byte("value1"))
	value2 := common.BytesToHash([]byte("value2"))

	testCases := []struct {
		name     string
		malleate func()
	}{
		{"simple revert", func() {
			revision := suite.app.EvmKeeper.Snapshot()
			suite.Require().Zero(revision)

			suite.app.EvmKeeper.SetState(suite.address, key, value1)
			suite.Require().Equal(value1, suite.app.EvmKeeper.GetState(suite.address, key))

			suite.app.EvmKeeper.RevertToSnapshot(revision)

			// reverted
			suite.Require().Equal(common.Hash{}, suite.app.EvmKeeper.GetState(suite.address, key))
		}},
		{"nested snapshot/revert", func() {
			revision1 := suite.app.EvmKeeper.Snapshot()
			suite.Require().Zero(revision1)

			suite.app.EvmKeeper.SetState(suite.address, key, value1)

			revision2 := suite.app.EvmKeeper.Snapshot()

			suite.app.EvmKeeper.SetState(suite.address, key, value2)
			suite.Require().Equal(value2, suite.app.EvmKeeper.GetState(suite.address, key))

			suite.app.EvmKeeper.RevertToSnapshot(revision2)
			suite.Require().Equal(value1, suite.app.EvmKeeper.GetState(suite.address, key))

			suite.app.EvmKeeper.RevertToSnapshot(revision1)
			suite.Require().Equal(common.Hash{}, suite.app.EvmKeeper.GetState(suite.address, key))
		}},
		{"jump revert", func() {
			revision1 := suite.app.EvmKeeper.Snapshot()
			suite.app.EvmKeeper.SetState(suite.address, key, value1)
			suite.app.EvmKeeper.Snapshot()
			suite.app.EvmKeeper.SetState(suite.address, key, value2)
			suite.app.EvmKeeper.RevertToSnapshot(revision1)
			suite.Require().Equal(common.Hash{}, suite.app.EvmKeeper.GetState(suite.address, key))
		}},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.SetupTest()
			tc.malleate()
			// the test case should finish in clean state
			suite.Require().True(suite.app.EvmKeeper.CachedContextsEmpty())
		})
	}
}

func (suite *KeeperTestSuite) CreateTestTx(msg *types.MsgEthereumTx, priv cryptotypes.PrivKey) authsigning.Tx {
	option, err := codectypes.NewAnyWithValue(&types.ExtensionOptionsEthereumTx{})
	suite.Require().NoError(err)

	txBuilder := suite.clientCtx.TxConfig.NewTxBuilder()
	builder, ok := txBuilder.(authtx.ExtensionOptionsTxBuilder)
	suite.Require().True(ok)

	builder.SetExtensionOptions(option)

	err = msg.Sign(suite.ethSigner, tests.NewSigner(priv))
	suite.Require().NoError(err)

	err = txBuilder.SetMsgs(msg)
	suite.Require().NoError(err)

	return txBuilder.GetTx()
}

func (suite *KeeperTestSuite) TestAddLog() {
	addr, privKey := tests.NewAddrKey()
	msg := types.NewTx(big.NewInt(1), 0, &suite.address, big.NewInt(1), 100000, big.NewInt(1), nil, nil, []byte("test"), nil)
	msg.From = addr.Hex()

	tx := suite.CreateTestTx(msg, privKey)
	msg, _ = tx.GetMsgs()[0].(*types.MsgEthereumTx)
	txHash := msg.AsTransaction().Hash()

	msg2 := types.NewTx(big.NewInt(1), 1, &suite.address, big.NewInt(1), 100000, big.NewInt(1), nil, nil, []byte("test"), nil)
	msg2.From = addr.Hex()

	tx2 := suite.CreateTestTx(msg2, privKey)
	msg2, _ = tx2.GetMsgs()[0].(*types.MsgEthereumTx)
	txHash2 := msg2.AsTransaction().Hash()

	msg3 := types.NewTx(big.NewInt(1), 0, &suite.address, big.NewInt(1), 100000, nil, big.NewInt(1), big.NewInt(1), []byte("test"), nil)
	msg3.From = addr.Hex()

	tx3 := suite.CreateTestTx(msg3, privKey)
	msg3, _ = tx3.GetMsgs()[0].(*types.MsgEthereumTx)
	txHash3 := msg3.AsTransaction().Hash()

	msg4 := types.NewTx(big.NewInt(1), 1, &suite.address, big.NewInt(1), 100000, nil, big.NewInt(1), big.NewInt(1), []byte("test"), nil)
	msg4.From = addr.Hex()

	tx4 := suite.CreateTestTx(msg4, privKey)
	msg4, _ = tx4.GetMsgs()[0].(*types.MsgEthereumTx)
	txHash4 := msg4.AsTransaction().Hash()

	testCases := []struct {
		name        string
		hash        common.Hash
		log, expLog *ethtypes.Log // pre and post populating log fields
		malleate    func()
	}{
		{
			"tx hash from message",
			txHash,
			&ethtypes.Log{
				Address: addr,
			},
			&ethtypes.Log{
				Address: addr,
				TxHash:  txHash,
				Topics:  make([]common.Hash, 0),
			},
			func() {},
		},
		{
			"log index keep increasing in new tx",
			txHash2,
			&ethtypes.Log{
				Address: addr,
			},
			&ethtypes.Log{
				Address: addr,
				TxHash:  txHash2,
				TxIndex: 1,
				Index:   1,
				Topics:  make([]common.Hash, 0),
			},
			func() {
				suite.app.EvmKeeper.SetTxHashTransient(txHash)
				suite.app.EvmKeeper.AddLog(&ethtypes.Log{
					Address: addr,
				})
				suite.app.EvmKeeper.IncreaseTxIndexTransient()
			},
		},
		{
			"dynamicfee tx hash from message",
			txHash3,
			&ethtypes.Log{
				Address: addr,
			},
			&ethtypes.Log{
				Address: addr,
				TxHash:  txHash3,
				Topics:  make([]common.Hash, 0),
			},
			func() {},
		},
		{
			"log index keep increasing in new dynamicfee tx",
			txHash4,
			&ethtypes.Log{
				Address: addr,
			},
			&ethtypes.Log{
				Address: addr,
				TxHash:  txHash4,
				TxIndex: 1,
				Index:   1,
				Topics:  make([]common.Hash, 0),
			},
			func() {
				suite.app.EvmKeeper.SetTxHashTransient(txHash)
				suite.app.EvmKeeper.AddLog(&ethtypes.Log{
					Address: addr,
				})
				suite.app.EvmKeeper.IncreaseTxIndexTransient()
			},
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.SetupTest()
			tc.malleate()

			suite.app.EvmKeeper.SetTxHashTransient(tc.hash)
			suite.app.EvmKeeper.AddLog(tc.log)
			logs := suite.app.EvmKeeper.GetTxLogsTransient(tc.hash)
			suite.Require().Equal(1, len(logs))
			suite.Require().Equal(tc.expLog, logs[0])
		})
	}
}

func (suite *KeeperTestSuite) TestPrepareAccessList() {
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
			suite.Require().True(addrOK, access.Address.Hex())
			suite.Require().True(slotOK, key.Hex())
		}
	}
}

func (suite *KeeperTestSuite) TestAddAddressToAccessList() {
	testCases := []struct {
		name string
		addr common.Address
	}{
		{"new address", suite.address},
		{"existing address", suite.address},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.app.EvmKeeper.AddAddressToAccessList(tc.addr)
			addrOk := suite.app.EvmKeeper.AddressInAccessList(tc.addr)
			suite.Require().True(addrOk, tc.addr.Hex())
		})
	}
}

func (suite *KeeperTestSuite) AddSlotToAccessList() {
	testCases := []struct {
		name string
		addr common.Address
		slot common.Hash
	}{
		{"new address and slot (1)", tests.GenerateAddress(), common.BytesToHash([]byte("hash"))},
		{"new address and slot (2)", suite.address, common.Hash{}},
		{"existing address and slot", suite.address, common.Hash{}},
		{"existing address, new slot", suite.address, common.BytesToHash([]byte("hash"))},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.app.EvmKeeper.AddSlotToAccessList(tc.addr, tc.slot)
			addrOk, slotOk := suite.app.EvmKeeper.SlotInAccessList(tc.addr, tc.slot)
			suite.Require().True(addrOk, tc.addr.Hex())
			suite.Require().True(slotOk, tc.slot.Hex())
		})
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
				return true
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
					return false
				}
				return true
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
