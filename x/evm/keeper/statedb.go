package keeper

import (
	"bytes"
	"math/big"
	"strconv"

	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	ethermint "github.com/cosmos/ethermint/types"
	"github.com/cosmos/ethermint/x/evm/types"
)

var _ vm.StateDB = &Keeper{}

// ----------------------------------------------------------------------------
// Account
// ----------------------------------------------------------------------------

// CreateAccount creates a new EthAccount instance from the provided address and
// sets the value to store.
func (k *Keeper) CreateAccount(addr common.Address) {
	cosmosAddr := sdk.AccAddress(addr.Bytes())

	_ = k.accountKeeper.NewAccountWithAddress(k.ctx, cosmosAddr)

	k.Logger(k.ctx).Debug(
		"account created",
		"ethereum-address", addr.Hex(),
		"cosmos-address", cosmosAddr.String(),
	)
}

// ----------------------------------------------------------------------------
// Balance
// ----------------------------------------------------------------------------

// AddBalance calls CommitStateDB.AddBalance using the passed in context
func (k *Keeper) AddBalance(addr common.Address, amount *big.Int) {
	cosmosAddr := sdk.AccAddress(addr.Bytes())

	params := k.GetParams(k.ctx)
	coins := sdk.Coins{sdk.NewCoin(params.EvmDenom, sdk.NewIntFromBigInt(amount))}

	if err := k.bankKeeper.AddCoins(k.ctx, cosmosAddr, coins); err != nil {
		k.Logger(k.ctx).Error(
			"failed to add balance",
			"ethereum-address", addr.Hex(),
			"cosmos-address", cosmosAddr.String(),
			"error", err,
		)
	}
}

// SubBalance calls CommitStateDB.SubBalance using the passed in context
func (k *Keeper) SubBalance(addr common.Address, amount *big.Int) {
	cosmosAddr := sdk.AccAddress(addr.Bytes())

	params := k.GetParams(k.ctx)
	coins := sdk.Coins{sdk.NewCoin(params.EvmDenom, sdk.NewIntFromBigInt(amount))}

	if err := k.bankKeeper.SubtractCoins(k.ctx, cosmosAddr, coins); err != nil {
		k.Logger(k.ctx).Error(
			"failed to subtract balance",
			"ethereum-address", addr.Hex(),
			"cosmos-address", cosmosAddr.String(),
			"error", err,
		)
	}
}

// GetBalance calls CommitStateDB.GetBalance using the passed in context
func (k *Keeper) GetBalance(addr common.Address) *big.Int {
	cosmosAddr := sdk.AccAddress(addr.Bytes())
	params := k.GetParams(k.ctx)
	balance := k.bankKeeper.GetBalance(k.ctx, cosmosAddr, params.EvmDenom)

	return balance.Amount.BigInt()
}

// ----------------------------------------------------------------------------
// Nonce
// ----------------------------------------------------------------------------

// GetNonce calls CommitStateDB.GetNonce using the passed in context
func (k *Keeper) GetNonce(addr common.Address) uint64 {
	cosmosAddr := sdk.AccAddress(addr.Bytes())
	nonce, err := k.accountKeeper.GetSequence(k.ctx, cosmosAddr)
	if err != nil {
		k.Logger(k.ctx).Error(
			"account not found",
			"ethereum-address", addr.Hex(),
			"cosmos-address", cosmosAddr.String(),
			"error", err,
		)
	}

	return nonce
}

// SetNonce calls CommitStateDB.SetNonce using the passed in context
func (k *Keeper) SetNonce(addr common.Address, nonce uint64) {
	cosmosAddr := sdk.AccAddress(addr.Bytes())
	account := k.accountKeeper.GetAccount(k.ctx, cosmosAddr)
	if account == nil {
		k.Logger(k.ctx).Error(
			"account not found",
			"ethereum-address", addr.Hex(),
			"cosmos-address", cosmosAddr.String(),
		)

		// create address if it doesn't exist
		account = k.accountKeeper.NewAccountWithAddress(k.ctx, cosmosAddr)
	}

	if err := account.SetSequence(nonce); err != nil {
		k.Logger(k.ctx).Error(
			"failed to set nonce",
			"ethereum-address", addr.Hex(),
			"cosmos-address", cosmosAddr.String(),
			"nonce", strconv.FormatUint(nonce, 64),
		)
	}

	k.accountKeeper.SetAccount(k.ctx, account)
}

// ----------------------------------------------------------------------------
// Code
// ----------------------------------------------------------------------------

// GetCodeHash calls CommitStateDB.GetCodeHash using the passed in context
func (k *Keeper) GetCodeHash(addr common.Address) common.Hash {
	// TODO: implement
	return common.Hash{}
}

// GetCode calls CommitStateDB.GetCode using the passed in context
func (k *Keeper) GetCode(addr common.Address) []byte {
	// TODO: implement
	return nil
}

// SetCode calls CommitStateDB.SetCode using the passed in context
func (k *Keeper) SetCode(addr common.Address, code []byte) {
	// TODO: implement
}

// GetCodeSize calls CommitStateDB.GetCodeSize using the passed in context
func (k *Keeper) GetCodeSize(addr common.Address) int {
	// TODO: implement
	return 0
}

// ----------------------------------------------------------------------------
// Refund
// ----------------------------------------------------------------------------

// AddRefund calls CommitStateDB.AddRefund using the passed in context
func (k *Keeper) AddRefund(gas uint64) {
	// TODO: implement
}

// SubRefund calls CommitStateDB.SubRefund using the passed in context
func (k *Keeper) SubRefund(gas uint64) {
	// TODO: implement
}

// GetRefund calls CommitStateDB.GetRefund using the passed in context
func (k *Keeper) GetRefund() uint64 {
	// TODO: implement
	return 0
}

// ----------------------------------------------------------------------------
// State
// ----------------------------------------------------------------------------

// GetCommittedState calls CommitStateDB.GetCommittedState using the passed in context
func (k *Keeper) GetCommittedState(addr common.Address, hash common.Hash) common.Hash {
	// TODO: implement
	return common.Hash{}
}

// GetState calls CommitStateDB.GetState using the passed in context
func (k *Keeper) GetState(addr common.Address, hash common.Hash) common.Hash {
	// TODO: implement
	return common.Hash{}
}

// SetState calls CommitStateDB.SetState using the passed in context
func (k *Keeper) SetState(addr common.Address, key, value common.Hash) {
	// TODO: implement
}

// ----------------------------------------------------------------------------
// Suicide
// ----------------------------------------------------------------------------

// Suicide implements the vm.StoreDB interface
func (k *Keeper) Suicide(addr common.Address) bool {
	store := prefix.NewStore(k.ctx.KVStore(k.storeKey), types.KeyPrefixSuicide)
	store.Set(addr.Bytes(), []byte{0x1})
	return true
}

// HasSuicided implements the vm.StoreDB interface
func (k *Keeper) HasSuicided(addr common.Address) bool {
	store := prefix.NewStore(k.ctx.KVStore(k.storeKey), types.KeyPrefixSuicide)
	return store.Has(addr.Bytes())
}

// ----------------------------------------------------------------------------
// Account Exist / Empty
// ----------------------------------------------------------------------------

// Exist calls CommitStateDB.Exist using the passed in context
func (k *Keeper) Exist(addr common.Address) bool {
	cosmosAddr := sdk.AccAddress(addr.Bytes())
	account := k.accountKeeper.GetAccount(k.ctx, cosmosAddr)
	if account != nil {
		return true
	}

	// return true if the account doesn't exist but has suicided
	return k.HasSuicided(addr)
}

// Empty calls CommitStateDB.Empty using the passed in context
func (k *Keeper) Empty(addr common.Address) bool {
	cosmosAddr := sdk.AccAddress(addr.Bytes())
	account := k.accountKeeper.GetAccount(k.ctx, cosmosAddr)
	if account == nil {
		// CONTRACT: we assume that if the account doesn't exist in store, it doesn't
		// have a balance
		return true
	}

	ethAccount, isEthAccount := account.(*ethermint.EthAccount)
	if !isEthAccount {
		// NOTE: non-ethereum accounts are considered empty
		return true
	}

	balance := k.GetBalance(addr)
	hasZeroBalance := balance.Sign() == 0
	hasEmptyHash := bytes.Equal(ethAccount.CodeHash, types.EmptyCodeHash)

	return hasZeroBalance && account.GetSequence() == 0 && hasEmptyHash
}

// ----------------------------------------------------------------------------
// Access List
// ----------------------------------------------------------------------------

func (k *Keeper) PrepareAccessList(sender common.Address, dest *common.Address, precompiles []common.Address, txAccesses ethtypes.AccessList) {
	// TODO: implement
}

func (k *Keeper) AddressInAccessList(addr common.Address) bool {
	// TODO: implement
	return false
}

func (k *Keeper) SlotInAccessList(addr common.Address, slot common.Hash) (addressOk bool, slotOk bool) {
	// TODO: implement
	return false, false
}

// AddAddressToAccessList adds the given address to the access list. This operation is safe to perform
// even if the feature/fork is not active yet
func (k *Keeper) AddAddressToAccessList(addr common.Address) {
	// TODO: implement
}

// AddSlotToAccessList adds the given (address,slot) to the access list. This operation is safe to perform
// even if the feature/fork is not active yet
func (k *Keeper) AddSlotToAccessList(addr common.Address, slot common.Hash) {
	// TODO: implement
}

// ----------------------------------------------------------------------------
// Snapshotting
// ----------------------------------------------------------------------------

// Snapshot calls CommitStateDB.Snapshot using the passed in context
func (k *Keeper) Snapshot() int {
	// TODO: implement
	return 0
}

// RevertToSnapshot calls CommitStateDB.RevertToSnapshot using the passed in context
func (k *Keeper) RevertToSnapshot(revID int) {
	// TODO: implement
}

// ----------------------------------------------------------------------------
// Log
// ----------------------------------------------------------------------------

// AddLog calls CommitStateDB.AddLog using the passed in context
func (k *Keeper) AddLog(log *ethtypes.Log) {
	// TODO: implement
}

// AddPreimage calls CommitStateDB.AddPreimage using the passed in context
func (k *Keeper) AddPreimage(hash common.Hash, preimage []byte) {
	// TODO: implement
}

// ----------------------------------------------------------------------------
// Iterator
// ----------------------------------------------------------------------------

// ForEachStorage calls CommitStateDB.ForEachStorage using passed in context
func (k *Keeper) ForEachStorage(addr common.Address, cb func(key, value common.Hash) bool) error {
	// TODO: implement
	return nil
}
