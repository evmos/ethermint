package keeper

import (
	"bytes"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/crypto"

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
// sets the value to store. If an account with the given address already exists,
// this function also resets any preexisting code and storage associated with that
// address.
func (k *Keeper) CreateAccount(addr common.Address) {
	cosmosAddr := sdk.AccAddress(addr.Bytes())
	ctx := k.Ctx()
	account := k.accountKeeper.GetAccount(ctx, cosmosAddr)
	log := ""
	if account == nil {
		log = "account created"
	} else {
		log = "account overwritten"
		k.ResetAccount(addr)
	}

	account = k.accountKeeper.NewAccountWithAddress(ctx, cosmosAddr)
	k.accountKeeper.SetAccount(ctx, account)

	k.Logger(ctx).Debug(
		log,
		"ethereum-address", addr.Hex(),
		"cosmos-address", cosmosAddr.String(),
	)
}

// ----------------------------------------------------------------------------
// Balance
// ----------------------------------------------------------------------------

// AddBalance adds the given amount to the address balance coin by minting new
// coins and transferring them to the address. The coin denomination is obtained
// from the module parameters.
func (k *Keeper) AddBalance(addr common.Address, amount *big.Int) {
	ctx := k.Ctx()

	if amount.Sign() != 1 {
		k.Logger(ctx).Debug(
			"ignored non-positive amount addition",
			"ethereum-address", addr.Hex(),
			"amount", amount.Int64(),
		)
		return
	}

	cosmosAddr := sdk.AccAddress(addr.Bytes())
	params := k.GetParams(ctx)

	// Coin denom and amount already validated
	coins := sdk.Coins{
		{
			Denom:  params.EvmDenom,
			Amount: sdk.NewIntFromBigInt(amount),
		},
	}

	if err := k.bankKeeper.MintCoins(ctx, types.ModuleName, coins); err != nil {
		k.Logger(ctx).Error(
			"failed to mint coins when adding balance",
			"ethereum-address", addr.Hex(),
			"cosmos-address", cosmosAddr.String(),
			"error", err,
		)
		return
	}

	if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, cosmosAddr, coins); err != nil {
		k.Logger(ctx).Error(
			"failed to send from module to account when adding balance",
			"ethereum-address", addr.Hex(),
			"cosmos-address", cosmosAddr.String(),
			"error", err,
		)
		return
	}

	k.Logger(ctx).Debug(
		"balance addition",
		"ethereum-address", addr.Hex(),
		"cosmos-address", cosmosAddr.String(),
	)
}

// SubBalance subtracts the given amount from the address balance by transferring the
// coins to an escrow account and then burning them. The coin denomination is obtained
// from the module parameters. This function performs a no-op if the amount is negative
// or the user doesn't have enough funds for the transfer.
func (k *Keeper) SubBalance(addr common.Address, amount *big.Int) {
	ctx := k.Ctx()

	if amount.Sign() != 1 {
		k.Logger(ctx).Debug(
			"ignored non-positive amount addition",
			"ethereum-address", addr.Hex(),
			"amount", amount.Int64(),
		)
		return
	}

	cosmosAddr := sdk.AccAddress(addr.Bytes())

	params := k.GetParams(ctx)

	// Coin denom and amount already validated
	coins := sdk.Coins{
		{
			Denom:  params.EvmDenom,
			Amount: sdk.NewIntFromBigInt(amount),
		},
	}

	if err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, cosmosAddr, types.ModuleName, coins); err != nil {
		k.Logger(ctx).Debug(
			"failed to send from account to module when subtracting balance",
			"ethereum-address", addr.Hex(),
			"cosmos-address", cosmosAddr.String(),
			"error", err,
		)

		return
	}

	if err := k.bankKeeper.BurnCoins(ctx, types.ModuleName, coins); err != nil {
		k.Logger(ctx).Error(
			"failed to burn coins when subtracting balance",
			"ethereum-address", addr.Hex(),
			"cosmos-address", cosmosAddr.String(),
			"error", err,
		)
		return
	}

	k.Logger(ctx).Debug(
		"balance subtraction",
		"ethereum-address", addr.Hex(),
		"cosmos-address", cosmosAddr.String(),
	)
}

// GetBalance returns the EVM denomination balance of the provided address. The
// denomination is obtained from the module parameters.
func (k *Keeper) GetBalance(addr common.Address) *big.Int {
	ctx := k.Ctx()

	cosmosAddr := sdk.AccAddress(addr.Bytes())
	params := k.GetParams(ctx)
	balance := k.bankKeeper.GetBalance(ctx, cosmosAddr, params.EvmDenom)

	return balance.Amount.BigInt()
}

// ----------------------------------------------------------------------------
// Nonce
// ----------------------------------------------------------------------------

// GetNonce retrieves the account with the given address and returns the tx
// sequence (i.e nonce). The function performs a no-op if the account is not found.
func (k *Keeper) GetNonce(addr common.Address) uint64 {
	ctx := k.Ctx()

	cosmosAddr := sdk.AccAddress(addr.Bytes())
	nonce, err := k.accountKeeper.GetSequence(ctx, cosmosAddr)
	if err != nil {
		k.Logger(ctx).Error(
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
	ctx := k.Ctx()

	cosmosAddr := sdk.AccAddress(addr.Bytes())
	account := k.accountKeeper.GetAccount(ctx, cosmosAddr)
	if account == nil {
		k.Logger(ctx).Debug(
			"account not found",
			"ethereum-address", addr.Hex(),
			"cosmos-address", cosmosAddr.String(),
		)

		// create address if it doesn't exist
		account = k.accountKeeper.NewAccountWithAddress(ctx, cosmosAddr)
	}

	if err := account.SetSequence(nonce); err != nil {
		k.Logger(ctx).Error(
			"failed to set nonce",
			"ethereum-address", addr.Hex(),
			"cosmos-address", cosmosAddr.String(),
			"nonce", nonce,
			"error", err,
		)

		return
	}

	k.accountKeeper.SetAccount(ctx, account)

	k.Logger(ctx).Debug(
		"nonce set",
		"ethereum-address", addr.Hex(),
		"cosmos-address", cosmosAddr.String(),
		"nonce", nonce,
	)
}

// ----------------------------------------------------------------------------
// Code
// ----------------------------------------------------------------------------

// GetCodeHash fetches the account from the store and returns its code hash. If the account doesn't
// exist or is not an EthAccount type, GetCodeHash returns the empty code hash value.
func (k *Keeper) GetCodeHash(addr common.Address) common.Hash {
	ctx := k.Ctx()
	cosmosAddr := sdk.AccAddress(addr.Bytes())

	account := k.accountKeeper.GetAccount(ctx, cosmosAddr)
	if account == nil {
		return common.BytesToHash(types.EmptyCodeHash)
	}

	ethAccount, isEthAccount := account.(*ethermint.EthAccount)
	if !isEthAccount {
		return common.BytesToHash(types.EmptyCodeHash)
	}

	return common.HexToHash(ethAccount.CodeHash)
}

// GetCode returns the code byte array associated with the given address.
// If the code hash from the account is empty, this function returns nil.
func (k *Keeper) GetCode(addr common.Address) []byte {
	ctx := k.Ctx()
	hash := k.GetCodeHash(addr)

	if bytes.Equal(hash.Bytes(), common.BytesToHash(types.EmptyCodeHash).Bytes()) {
		return nil
	}

	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixCode)
	code := store.Get(hash.Bytes())

	if len(code) == 0 {
		k.Logger(ctx).Debug(
			"code not found",
			"ethereum-address", addr.Hex(),
			"code-hash", hash.Hex(),
		)
	}

	return code
}

// SetCode stores the code byte array to the application KVStore and sets the
// code hash to the given account. The code is deleted from the store if it is empty.
func (k *Keeper) SetCode(addr common.Address, code []byte) {
	ctx := k.Ctx()

	if bytes.Equal(code, types.EmptyCodeHash) {
		k.Logger(ctx).Debug("passed in EmptyCodeHash, but expected empty code")
	}
	hash := crypto.Keccak256Hash(code)

	// update account code hash
	account := k.accountKeeper.GetAccount(ctx, addr.Bytes())
	if account == nil {
		account = k.accountKeeper.NewAccountWithAddress(ctx, addr.Bytes())
		k.accountKeeper.SetAccount(ctx, account)
	}

	ethAccount, isEthAccount := account.(*ethermint.EthAccount)
	if !isEthAccount {
		k.Logger(ctx).Error(
			"invalid account type",
			"ethereum-address", addr.Hex(),
			"code-hash", hash.Hex(),
		)
		return
	}

	ethAccount.CodeHash = hash.Hex()
	k.accountKeeper.SetAccount(ctx, ethAccount)

	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixCode)

	action := "updated"

	// store or delete code
	if len(code) == 0 {
		store.Delete(hash.Bytes())
		action = "deleted"
	} else {
		store.Set(hash.Bytes(), code)
	}

	k.Logger(ctx).Debug(
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

// AddRefund adds the given amount of gas to the refund transient value.
func (k *Keeper) AddRefund(gas uint64) {
	ctx := k.Ctx()
	refund := k.GetRefund()

	refund += gas

	store := ctx.TransientStore(k.transientKey)
	store.Set(types.KeyPrefixTransientRefund, sdk.Uint64ToBigEndian(refund))
}

// SubRefund subtracts the given amount of gas from the transient refund value. This function
// will panic if gas amount is greater than the stored refund.
func (k *Keeper) SubRefund(gas uint64) {
	ctx := k.Ctx()
	refund := k.GetRefund()

	if gas > refund {
		// TODO: (@fedekunze) set to 0?? Geth panics here
		panic("refund counter below zero")
	}

	refund -= gas

	store := ctx.TransientStore(k.transientKey)
	store.Set(types.KeyPrefixTransientRefund, sdk.Uint64ToBigEndian(refund))
}

// GetRefund returns the amount of gas available for return after the tx execution
// finalizes. This value is reset to 0 on every transaction.
func (k *Keeper) GetRefund() uint64 {
	ctx := k.Ctx()
	store := ctx.TransientStore(k.transientKey)

	bz := store.Get(types.KeyPrefixTransientRefund)
	if len(bz) == 0 {
		return 0
	}

	return sdk.BigEndianToUint64(bz)
}

// ----------------------------------------------------------------------------
// State
// ----------------------------------------------------------------------------

func doGetState(ctx sdk.Context, storeKey sdk.StoreKey, addr common.Address, hash common.Hash) common.Hash {
	store := prefix.NewStore(ctx.KVStore(storeKey), types.AddressStoragePrefix(addr))

	key := types.KeyAddressStorage(addr, hash)
	value := store.Get(key.Bytes())
	if len(value) == 0 {
		return common.Hash{}
	}

	return common.BytesToHash(value)
}

// GetCommittedState returns the value set in store for the given key hash. If the key is not registered
// this function returns the empty hash.
func (k *Keeper) GetCommittedState(addr common.Address, hash common.Hash) common.Hash {
	return doGetState(k.ctxStack.initialCtx, k.storeKey, addr, hash)
}

// GetState returns the committed state for the given key hash, as all changes are committed directly
// to the KVStore.
func (k *Keeper) GetState(addr common.Address, hash common.Hash) common.Hash {
	ctx := k.Ctx()
	return doGetState(ctx, k.storeKey, addr, hash)
}

// SetState sets the given hashes (key, value) to the KVStore. If the value hash is empty, this
// function deletes the key from the store.
func (k *Keeper) SetState(addr common.Address, key, value common.Hash) {
	ctx := k.Ctx()
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.AddressStoragePrefix(addr))
	key = types.KeyAddressStorage(addr, key)

	action := "updated"
	if ethermint.IsEmptyHash(value.Hex()) {
		store.Delete(key.Bytes())
		action = "deleted"
	} else {
		store.Set(key.Bytes(), value.Bytes())
	}

	k.Logger(ctx).Debug(
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
	ctx := k.Ctx()

	prev := k.HasSuicided(addr)
	if prev {
		return true
	}

	cosmosAddr := sdk.AccAddress(addr.Bytes())

	_, err := k.ClearBalance(cosmosAddr)
	if err != nil {
		k.Logger(ctx).Error(
			"failed to subtract balance on suicide",
			"ethereum-address", addr.Hex(),
			"cosmos-address", cosmosAddr.String(),
			"error", err,
		)

		return false
	}

	// TODO: (@fedekunze) do we also need to delete the storage state and the code?
	k.setSuicided(ctx, addr)

	k.Logger(ctx).Debug(
		"account suicided",
		"ethereum-address", addr.Hex(),
		"cosmos-address", cosmosAddr.String(),
	)

	return true
}

// setSuicided sets a single byte to the transient store and marks the address as suicided
func (k Keeper) setSuicided(ctx sdk.Context, addr common.Address) {
	store := prefix.NewStore(ctx.TransientStore(k.transientKey), types.KeyPrefixTransientSuicided)
	store.Set(addr.Bytes(), []byte{1})
}

// HasSuicided queries the transient store to check if the account has been marked as suicided in the
// current block. Accounts that are suicided will be returned as non-nil during queries and "cleared"
// after the block has been committed.
func (k *Keeper) HasSuicided(addr common.Address) bool {
	ctx := k.Ctx()
	store := prefix.NewStore(ctx.TransientStore(k.transientKey), types.KeyPrefixTransientSuicided)
	return store.Has(addr.Bytes())
}

// ----------------------------------------------------------------------------
// Account Exist / Empty
// ----------------------------------------------------------------------------

// Exist returns true if the given account exists in store or if it has been
// marked as suicided in the transient store.
func (k *Keeper) Exist(addr common.Address) bool {
	ctx := k.Ctx()
	// return true if the account has suicided
	if k.HasSuicided(addr) {
		return true
	}

	cosmosAddr := sdk.AccAddress(addr.Bytes())
	account := k.accountKeeper.GetAccount(ctx, cosmosAddr)
	return account != nil
}

// Empty returns true if the address meets the following conditions:
// 	- nonce is 0
// 	- balance amount for evm denom is 0
// 	- account code hash is empty
//
// Non-ethereum accounts are considered not empty
func (k *Keeper) Empty(addr common.Address) bool {
	ctx := k.Ctx()
	nonce := uint64(0)
	codeHash := types.EmptyCodeHash

	cosmosAddr := sdk.AccAddress(addr.Bytes())
	account := k.accountKeeper.GetAccount(ctx, cosmosAddr)

	if account != nil {
		nonce = account.GetSequence()
		ethAccount, isEthAccount := account.(*ethermint.EthAccount)
		if !isEthAccount {
			return false
		}

		codeHash = common.HexToHash(ethAccount.CodeHash).Bytes()
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
	ctx := k.Ctx()
	ts := prefix.NewStore(ctx.TransientStore(k.transientKey), types.KeyPrefixTransientAccessListAddress)
	return ts.Has(addr.Bytes())
}

// SlotInAccessList checks if the address and the slots are registered in the transient store
func (k *Keeper) SlotInAccessList(addr common.Address, slot common.Hash) (addressOk, slotOk bool) {
	addressOk = k.AddressInAccessList(addr)
	slotOk = k.addressSlotInAccessList(addr, slot)
	return addressOk, slotOk
}

// addressSlotInAccessList returns true if the address's slot is registered on the transient store.
func (k *Keeper) addressSlotInAccessList(addr common.Address, slot common.Hash) bool {
	ctx := k.Ctx()
	ts := prefix.NewStore(ctx.TransientStore(k.transientKey), types.KeyPrefixTransientAccessListSlot)
	key := append(addr.Bytes(), slot.Bytes()...)
	return ts.Has(key)
}

// AddAddressToAccessList adds the given address to the access list. If the address is already
// in the access list, this function performs a no-op.
func (k *Keeper) AddAddressToAccessList(addr common.Address) {
	if k.AddressInAccessList(addr) {
		return
	}

	ctx := k.Ctx()
	ts := prefix.NewStore(ctx.TransientStore(k.transientKey), types.KeyPrefixTransientAccessListAddress)
	ts.Set(addr.Bytes(), []byte{0x1})
}

// AddSlotToAccessList adds the given (address, slot) to the access list. If the address and slot are
// already in the access list, this function performs a no-op.
func (k *Keeper) AddSlotToAccessList(addr common.Address, slot common.Hash) {
	k.AddAddressToAccessList(addr)
	if k.addressSlotInAccessList(addr, slot) {
		return
	}

	ctx := k.Ctx()
	ts := prefix.NewStore(ctx.TransientStore(k.transientKey), types.KeyPrefixTransientAccessListSlot)
	key := append(addr.Bytes(), slot.Bytes()...)
	ts.Set(key, []byte{0x1})
}

// ----------------------------------------------------------------------------
// Snapshotting
// ----------------------------------------------------------------------------

// Snapshot return the index in the cached context stack
func (k *Keeper) Snapshot() int {
	return k.ctxStack.Snapshot()
}

// RevertToSnapshot pop all the cached contexts after(including) the snapshot
func (k *Keeper) RevertToSnapshot(target int) {
	k.ctxStack.RevertToSnapshot(target)
}

// ----------------------------------------------------------------------------
// Log
// ----------------------------------------------------------------------------

// AddLog appends the given ethereum Log to the list of Logs associated with the transaction hash kept in the current
// context. This function also fills in the tx hash, block hash, tx index and log index fields before setting the log
// to store.
func (k *Keeper) AddLog(log *ethtypes.Log) {
	ctx := k.Ctx()

	log.BlockHash = common.BytesToHash(ctx.HeaderHash())
	log.TxIndex = uint(k.GetTxIndexTransient())
	log.TxHash = k.GetTxHashTransient()

	log.Index = uint(k.GetLogSizeTransient())
	k.IncreaseLogSizeTransient()
	k.SetLog(log)

	k.Logger(ctx).Debug(
		"log added",
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

// ForEachStorage uses the store iterator to iterate over all the state keys and perform a callback
// function on each of them.
func (k *Keeper) ForEachStorage(addr common.Address, cb func(key, value common.Hash) bool) error {
	ctx := k.Ctx()
	store := ctx.KVStore(k.storeKey)
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
