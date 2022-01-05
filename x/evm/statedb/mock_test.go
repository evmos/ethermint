package statedb_test

import (
	"errors"
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/tharsis/ethermint/x/evm/statedb"
)

var _ statedb.Keeper = &MockKeeper{}

type MockKeeper struct {
	errAddress common.Address

	accounts map[common.Address]statedb.Account
	states   map[common.Address]statedb.Storage
	codes    map[common.Hash][]byte
}

func NewMockKeeper() *MockKeeper {
	return &MockKeeper{
		errAddress: common.BigToAddress(big.NewInt(1)),

		accounts: make(map[common.Address]statedb.Account),
		states:   make(map[common.Address]statedb.Storage),
		codes:    make(map[common.Hash][]byte),
	}
}

func (k MockKeeper) GetAccount(ctx sdk.Context, addr common.Address) (*statedb.Account, error) {
	if addr == k.errAddress {
		return nil, errors.New("mock db error")
	}
	acct, ok := k.accounts[addr]
	if !ok {
		return nil, nil
	}
	return &acct, nil
}

func (k MockKeeper) GetState(ctx sdk.Context, addr common.Address, key common.Hash) common.Hash {
	return k.states[addr][key]
}

func (k MockKeeper) GetCode(ctx sdk.Context, codeHash common.Hash) []byte {
	return k.codes[codeHash]
}

func (k MockKeeper) ForEachStorage(ctx sdk.Context, addr common.Address, cb func(key, value common.Hash) bool) {
	for k, v := range k.states[addr] {
		if !cb(k, v) {
			return
		}
	}
}

func (k MockKeeper) SetAccount(ctx sdk.Context, addr common.Address, account statedb.Account) error {
	k.accounts[addr] = account
	return nil
}

func (k MockKeeper) SetState(ctx sdk.Context, addr common.Address, key common.Hash, value []byte) {
	if len(value) == 0 {
		delete(k.states[addr], key)
	} else {
		k.states[addr][key] = common.BytesToHash(value)
	}
}

func (k MockKeeper) SetCode(ctx sdk.Context, codeHash []byte, code []byte) {
	k.codes[common.BytesToHash(codeHash)] = code
}

func (k MockKeeper) DeleteAccount(ctx sdk.Context, addr common.Address) error {
	old := k.accounts[addr]
	delete(k.accounts, addr)
	delete(k.states, addr)
	if len(old.CodeHash) > 0 {
		delete(k.codes, common.BytesToHash(old.CodeHash))
	}
	return nil
}
