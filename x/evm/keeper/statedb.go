package keeper

import (
	"bytes"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/crypto"

	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"

	ethermint "github.com/tharsis/ethermint/types"
	"github.com/tharsis/ethermint/x/evm/types"
)

var _ vm.StateDB = &Keeper{}

// ----------------------------------------------------------------------------
// Account
// ----------------------------------------------------------------------------

// CreateAccount creates a new EthAccount instance from the provided address and
// sets the value to store.
func (k *Keeper) CreateAccount(addr common.Address) {
	cosmosAddr := sdk.AccAddress(addr.Bytes())

	account := k.accountKeeper.GetAccount(k.ctx, cosmosAddr)
	log := ""
	if account == nil {
		log = "account created"
	} else {
		log = "account overwritten"
		k.ResetAccount(addr)
	}

	account = k.accountKeeper.NewAccountWithAddress(k.ctx, cosmosAddr)
	k.accountKeeper.SetAccount(k.ctx, account)

	k.Logger(k.ctx).Debug(
		log,
		"ethereum-address", addr.Hex(),
		"cosmos-address", cosmosAddr.String(),
	)
}

// ----------------------------------------------------------------------------
// Balance
// ----------------------------------------------------------------------------

// AddBalance calls CommitStateDB.AddBalance using the passed in context
func (k *Keeper) AddBalance(addr common.Address, amount *big.Int) {
	if amount.Sign() != 1 {
		k.Logger(k.ctx).Debug(
			"ignored non-positive amount addition",
			"ethereum-address", addr.Hex(),
			"amount", amount.Int64(),
		)
		return
	}

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
		return
	}

	k.Logger(k.ctx).Debug(
		"balance addition",
		"ethereum-address", addr.Hex(),
		"cosmos-address", cosmosAddr.String(),
	)
}

// SubBalance calls CommitStateDB.SubBalance using the passed in context
func (k *Keeper) SubBalance(addr common.Address, amount *big.Int) {
	if amount.Sign() != 1 {
		k.Logger(k.ctx).Debug(
			"ignored non-positive amount addition",
			"ethereum-address", addr.Hex(),
			"amount", amount.Int64(),
		)
		return
	}

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

		return
	}

	k.Logger(k.ctx).Debug(
		"balance subtraction",
		"ethereum-address", addr.Hex(),
		"cosmos-address", cosmosAddr.String(),
	)
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

// SetNonce sets the given nonce as the sequence of the address' account. If the
// account doesn't exist, a new one will be created from the address.
func (k *Keeper) SetNonce(addr common.Address, nonce uint64) {
	cosmosAddr := sdk.AccAddress(addr.Bytes())
	account := k.accountKeeper.GetAccount(k.ctx, cosmosAddr)
	if account == nil {
		k.Logger(k.ctx).Debug(
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
			"nonce", nonce,
			"error", err,
		)

		return
	}

	k.accountKeeper.SetAccount(k.ctx, account)

	k.Logger(k.ctx).Debug(
		"nonce set",
		"ethereum-address", addr.Hex(),
		"cosmos-address", cosmosAddr.String(),
		"nonce", nonce,
	)
}

// ----------------------------------------------------------------------------
// Code
// ----------------------------------------------------------------------------

// GetCodeHash calls CommitStateDB.GetCodeHash using the passed in context
func (k *Keeper) GetCodeHash(addr common.Address) common.Hash {
	cosmosAddr := sdk.AccAddress(addr.Bytes())
	account := k.accountKeeper.GetAccount(k.ctx, cosmosAddr)
	if account == nil {
		return common.BytesToHash(types.EmptyCodeHash)
	}

	ethAccount, isEthAccount := account.(*ethermint.EthAccount)
	if !isEthAccount {
		return common.BytesToHash(types.EmptyCodeHash)
	}

	return common.BytesToHash(ethAccount.CodeHash)
}

// GetCode calls CommitStateDB.GetCode using the passed in context
func (k *Keeper) GetCode(addr common.Address) []byte {
	hash := k.GetCodeHash(addr)

	if bytes.Equal(hash.Bytes(), common.BytesToHash(types.EmptyCodeHash).Bytes()) {
		return nil
	}

	store := prefix.NewStore(k.ctx.KVStore(k.storeKey), types.KeyPrefixCode)
	code := store.Get(hash.Bytes())

	if len(code) == 0 {
		k.Logger(k.ctx).Debug(
			"code not found",
			"ethereum-address", addr.Hex(),
			"code-hash", hash.Hex(),
		)
	}

	return code
}

// SetCode calls CommitStateDB.SetCode using the passed in context
func (k *Keeper) SetCode(addr common.Address, code []byte) {
	if bytes.Equal(code, types.EmptyCodeHash) {
		k.Logger(k.ctx).Debug("passed in EmptyCodeHash, but expected empty code")
	}
	hash := crypto.Keccak256Hash(code)

	// update account code hash
	account := k.accountKeeper.GetAccount(k.ctx, addr.Bytes())
	if account == nil {
		account = k.accountKeeper.NewAccountWithAddress(k.ctx, addr.Bytes())
		k.accountKeeper.SetAccount(k.ctx, account)
	}

	ethAccount, isEthAccount := account.(*ethermint.EthAccount)
	if !isEthAccount {
		k.Logger(k.ctx).Error(
			"invalid account type",
			"ethereum-address", addr.Hex(),
			"code-hash", hash.Hex(),
		)
		return
	}

	ethAccount.CodeHash = hash.Bytes()
	k.accountKeeper.SetAccount(k.ctx, ethAccount)

	store := prefix.NewStore(k.ctx.KVStore(k.storeKey), types.KeyPrefixCode)

	action := "updated"

	// store or delete code
	if len(code) == 0 {
		store.Delete(hash.Bytes())
		action = "deleted"
	} else {
		store.Set(hash.Bytes(), code)
	}

	k.Logger(k.ctx).Debug(
		fmt.Sprintf("code %s", action),
		"ethereum-address", addr.Hex(),
		"code-hash", hash.Hex(),
	)
}

// GetCodeSize returns the size of the contract code associated with this object,
// or zero if none.
func (k *Keeper) GetCodeSize(addr common.Address) int {
	code := k.GetCode(addr)
	return len(code)
}

// ----------------------------------------------------------------------------
// Refund
// ----------------------------------------------------------------------------

// NOTE: gas refunded needs to be tracked and stored in a separate variable in
// order to add it subtract/add it from/to the gas used value after the EVM
// execution has finalized. The refund value is cleared on every transaction and
// at the end of every block.

// AddRefund adds the given amount of gas to the refund cached value.
func (k *Keeper) AddRefund(gas uint64) {
	refund := k.GetRefund()

	refund += gas

	store := k.ctx.TransientStore(k.transientKey)
	store.Set(types.KeyPrefixTransientRefund, sdk.Uint64ToBigEndian(refund))
}

// SubRefund subtracts the given amount of gas from the refund value. This function
// will panic if gas amount is greater than the stored refund.
func (k *Keeper) SubRefund(gas uint64) {
	refund := k.GetRefund()

	if gas > refund {
		// TODO: (@fedekunze) set to 0?? Geth panics here
		panic("refund counter below zero")
	}

	refund -= gas

	store := k.ctx.TransientStore(k.transientKey)
	store.Set(types.KeyPrefixTransientRefund, sdk.Uint64ToBigEndian(refund))
}

// GetRefund returns the amount of gas available for return after the tx execution
// finalises. This value is reset to 0 on every transaction.
func (k *Keeper) GetRefund() uint64 {
	store := k.ctx.TransientStore(k.transientKey)

	bz := store.Get(types.KeyPrefixTransientRefund)
	if len(bz) == 0 {
		return 0
	}

	return sdk.BigEndianToUint64(bz)
}

// ----------------------------------------------------------------------------
// State
// ----------------------------------------------------------------------------

// GetCommittedState returns the value set in store for the given key hash. If the key is not registered
// this function returns the empty hash.
func (k *Keeper) GetCommittedState(addr common.Address, hash common.Hash) common.Hash {
	store := prefix.NewStore(k.ctx.KVStore(k.storeKey), types.AddressStoragePrefix(addr))

	key := types.KeyAddressStorage(addr, hash)
	value := store.Get(key.Bytes())
	if len(value) == 0 {
		return common.Hash{}
	}

	return common.BytesToHash(value)
}

// GetState returns the committed state for the given key hash, as all changes are committed directly
// to the KVStore.
func (k *Keeper) GetState(addr common.Address, hash common.Hash) common.Hash {
	return k.GetCommittedState(addr, hash)
}

// SetState sets the given hashes (key, value) to the KVStore. If the value hash is empty, this
// function deletes the key from the store.
func (k *Keeper) SetState(addr common.Address, key, value common.Hash) {
	store := prefix.NewStore(k.ctx.KVStore(k.storeKey), types.AddressStoragePrefix(addr))
	key = types.KeyAddressStorage(addr, key)

	action := "updated"
	if ethermint.IsEmptyHash(value.Hex()) {
		store.Delete(key.Bytes())
		action = "deleted"
	} else {
		store.Set(key.Bytes(), value.Bytes())
	}

	k.Logger(k.ctx).Debug(
		fmt.Sprintf("state %s", action),
		"ethereum-address", addr.Hex(),
		"key", key.Hex(),
	)
}

// ----------------------------------------------------------------------------
// Suicide
// ----------------------------------------------------------------------------

// Suicide marks the given account as suicided and clears the account balance of
// the EVM tokens.
func (k *Keeper) Suicide(addr common.Address) bool {
	prev := k.HasSuicided(addr)
	if prev {
		return true
	}

	cosmosAddr := sdk.AccAddress(addr.Bytes())

	_, err := k.ClearBalance(cosmosAddr)
	if err != nil {
		k.Logger(k.ctx).Error(
			"failed to subtract balance on suicide",
			"ethereum-address", addr.Hex(),
			"cosmos-address", cosmosAddr.String(),
			"error", err,
		)

		return false
	}

	// TODO: (@fedekunze) do we also need to delete the storage state and the code?

	// Set a single byte to the transient store
	store := prefix.NewStore(k.ctx.TransientStore(k.transientKey), types.KeyPrefixTransientSuicided)
	store.Set(addr.Bytes(), []byte{1})

	k.Logger(k.ctx).Debug(
		"account suicided",
		"ethereum-address", addr.Hex(),
		"cosmos-address", cosmosAddr.String(),
	)

	return true
}

// HasSuicided queries the transient store to check if the account has been marked as suicided in the
// current block. Accounts that are suicided will be returned as non-nil during queries and "cleared"
// after the block has been committed.
func (k *Keeper) HasSuicided(addr common.Address) bool {
	store := prefix.NewStore(k.ctx.TransientStore(k.transientKey), types.KeyPrefixTransientSuicided)
	return store.Has(addr.Bytes())
}

// ----------------------------------------------------------------------------
// Account Exist / Empty
// ----------------------------------------------------------------------------

// Exist returns true if the given account exists in store or if it has been
// marked as suicided in the transient store.
func (k *Keeper) Exist(addr common.Address) bool {
	// return true if the account has suicided
	if k.HasSuicided(addr) {
		return true
	}

	cosmosAddr := sdk.AccAddress(addr.Bytes())
	account := k.accountKeeper.GetAccount(k.ctx, cosmosAddr)
	return account != nil
}

// Empty returns true if the address meets the following conditions:
// 	- nonce is 0
// 	- balance amount for evm denom is 0
// 	- account code hash is empty
//
// Non-ethereum accounts are considered not empty
func (k *Keeper) Empty(addr common.Address) bool {
	nonce := uint64(0)
	codeHash := types.EmptyCodeHash

	cosmosAddr := sdk.AccAddress(addr.Bytes())
	account := k.accountKeeper.GetAccount(k.ctx, cosmosAddr)

	if account != nil {
		nonce = account.GetSequence()
		ethAccount, isEthAccount := account.(*ethermint.EthAccount)
		if !isEthAccount {
			return false
		}

		codeHash = ethAccount.CodeHash
	}

	balance := k.GetBalance(addr)
	hasZeroBalance := balance.Sign() == 0
	hasEmptyCodeHash := bytes.Equal(codeHash, types.EmptyCodeHash)

	return hasZeroBalance && nonce == 0 && hasEmptyCodeHash
}

// ----------------------------------------------------------------------------
// Access List
// ----------------------------------------------------------------------------

// PrepareAccessList handles the preparatory steps for executing a state transition with
// regards to both EIP-2929 and EIP-2930:
//
// 	- Add sender to access list (2929)
// 	- Add destination to access list (2929)
// 	- Add precompiles to access list (2929)
// 	- Add the contents of the optional tx access list (2930)
//
// This method should only be called if Yolov3/Berlin/2929+2930 is applicable at the current number.
func (k *Keeper) PrepareAccessList(sender common.Address, dest *common.Address, precompiles []common.Address, txAccesses ethtypes.AccessList) {
	k.AddAddressToAccessList(sender)
	if dest != nil {
		k.AddAddressToAccessList(*dest)
		// If it's a create-tx, the destination will be added inside evm.create
	}
	for _, addr := range precompiles {
		k.AddAddressToAccessList(addr)
	}
	for _, tuple := range txAccesses {
		k.AddAddressToAccessList(tuple.Address)
		for _, key := range tuple.StorageKeys {
			k.AddSlotToAccessList(tuple.Address, key)
		}
	}
}

// AddressInAccessList returns true if the address is registered on the transient store.
func (k *Keeper) AddressInAccessList(addr common.Address) bool {
	ts := prefix.NewStore(k.ctx.TransientStore(k.transientKey), types.KeyPrefixTransientAccessListAddress)
	return ts.Has(addr.Bytes())
}

// SlotInAccessList checks if the address and the slots are registered in the transient store
func (k *Keeper) SlotInAccessList(addr common.Address, slot common.Hash) (addressOk bool, slotOk bool) {
	addressOk = k.AddressInAccessList(addr)
	slotOk = k.addressSlotInAccessList(addr, slot)
	return addressOk, slotOk
}

// addressSlotInAccessList returns true if the address's slot is registered on the transient store.
func (k *Keeper) addressSlotInAccessList(addr common.Address, slot common.Hash) bool {
	ts := prefix.NewStore(k.ctx.TransientStore(k.transientKey), types.KeyPrefixTransientAccessListSlot)
	key := append(addr.Bytes(), slot.Bytes()...)
	return ts.Has(key)
}

// AddAddressToAccessList adds the given address to the access list. If the address is already
// in the access list, this function performs a no-op.
func (k *Keeper) AddAddressToAccessList(addr common.Address) {
	if k.AddressInAccessList(addr) {
		return
	}

	ts := prefix.NewStore(k.ctx.TransientStore(k.transientKey), types.KeyPrefixTransientAccessListAddress)
	ts.Set(addr.Bytes(), []byte{0x1})
}

// AddSlotToAccessList adds the given (address, slot) to the access list. If the address and slot are
// already in the access list, this function performs a no-op.
func (k *Keeper) AddSlotToAccessList(addr common.Address, slot common.Hash) {
	k.AddAddressToAccessList(addr)
	if k.addressSlotInAccessList(addr, slot) {
		return
	}

	ts := prefix.NewStore(k.ctx.TransientStore(k.transientKey), types.KeyPrefixTransientAccessListSlot)
	key := append(addr.Bytes(), slot.Bytes()...)
	ts.Set(key, []byte{0x1})
}

// ----------------------------------------------------------------------------
// Snapshotting
// ----------------------------------------------------------------------------

// Snapshot return zero as the state changes won't be committed if the state transition fails. So there
// is no need to snapshot before the VM execution.
// See Cosmos SDK docs for more info: https://docs.cosmos.network/master/core/baseapp.html#delivertx-state-updates
func (k *Keeper) Snapshot() int {
	return 0
}

// RevertToSnapshot performs a no-op because when a transaction execution fails on the EVM, the state
// won't be persisted during ABCI DeliverTx.
func (k *Keeper) RevertToSnapshot(_ int) {}

// ----------------------------------------------------------------------------
// Log
// ----------------------------------------------------------------------------

// AddLog appends the given ethereum Log to the list of Logs associated with the transaction hash kept in the current
// context. This function also fills in the tx hash, block hash, tx index and log index fields before setting the log
// to store.
func (k *Keeper) AddLog(log *ethtypes.Log) {
	tx, err := k.txDecoder(k.ctx.TxBytes())
	if err != nil {
		// safety check, should be checked when processing the tx
		panic(err)
	}
	// NOTE: tx length checked on AnteHandler
	ethTx, ok := tx.GetMsgs()[0].(*types.MsgEthereumTx)
	if !ok {
		panic("invalid ethereum tx")
	}

	// NOTE: we set up the transaction hash from tendermint as it is the format expected by the application:
	// Remove once hashing is fixed on Tendermint. See https://github.com/tendermint/tendermint/issues/6539
	key := common.BytesToHash(tmtypes.Tx(k.ctx.TxBytes()).Hash())
	log.TxHash = common.HexToHash(ethTx.Hash)

	log.BlockHash = common.BytesToHash(k.ctx.HeaderHash())
	log.TxIndex = uint(k.GetTxIndexTransient())

	logs := k.GetTxLogs(log.TxHash)

	log.Index = uint(len(logs))
	logs = append(logs, log)

	k.SetLogs(key, logs)

	k.Logger(k.ctx).Debug(
		"log added",
		"tx-hash-tendermint", key.Hex(),
		"tx-hash-ethereum", log.TxHash.Hex(),
		"log-index", int(log.Index),
	)
}

// ----------------------------------------------------------------------------
// Trie
// ----------------------------------------------------------------------------

// AddPreimage performs a no-op since the EnablePreimageRecording flag is disabled
// on the vm.Config during state transitions. No store trie preimages are written
// to the database.
func (k *Keeper) AddPreimage(_ common.Hash, _ []byte) {}

// ----------------------------------------------------------------------------
// Iterator
// ----------------------------------------------------------------------------

// ForEachStorage calls CommitStateDB.ForEachStorage using passed in context
func (k *Keeper) ForEachStorage(addr common.Address, cb func(key, value common.Hash) bool) error {
	store := k.ctx.KVStore(k.storeKey)
	prefix := types.AddressStoragePrefix(addr)

	iterator := sdk.KVStorePrefixIterator(store, prefix)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {

		// TODO: check if the key prefix needs to be trimmed
		key := common.BytesToHash(iterator.Key())
		value := common.BytesToHash(iterator.Value())

		// check if iteration stops
		if cb(key, value) {
			return nil
		}
	}

	return nil
}
