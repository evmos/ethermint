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
	ethermint "github.com/cosmos/ethermint/types"
	"github.com/cosmos/ethermint/x/evm/types"
)

var _ vm.StateDB = &Keeper{}

// csdb defines the internal field values used on the StateDB operations
type csdb struct {
	// The refund counter, also used by state transitioning.
	refund uint64

	// Per-transaction access list
	accessList *types.AccessListMappings

	// Transaction counter in a block. Used on StateSB's Prepare function.
	// It is reset to 0 every block on BeginBlock so there's no point in storing the counter
	// on the KVStore or adding it as a field on the EVM genesis state.
	txIndex   int
	bloom     *big.Int
	blockHash common.Hash

	// logs is a cache field that keeps mapping of contract address -> eth logs emitted
	// during EVM execution in the current block.
	logs map[common.Address][]*ethtypes.Log

	// accounts that are suicided will be returned as non-nil during queries at "cleared" at
	// every end block.
	suicided map[common.Address]bool
}

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

	_ = k.accountKeeper.NewAccountWithAddress(k.ctx, cosmosAddr)

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

	// if k.Empty(addr) {
	// 	DeleteAccount, balance, code, storage
	// }

	k.Logger(k.ctx).Debug(
		"balance substraction",
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

// SetNonce calls CommitStateDB.SetNonce using the passed in context
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

		// if nonce == 0 && k.Empty(addr) {

		// }
	}

	// prevNonce := account.GetSequence()

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
	hash := crypto.Keccak256Hash(code)

	// update account code hash
	account := k.accountKeeper.GetAccount(k.ctx, addr.Bytes())
	if account == nil {
		account = k.accountKeeper.NewAccountWithAddress(k.ctx, addr.Bytes())
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

// GetCodeSize calls CommitStateDB.GetCodeSize using the passed in context
func (k *Keeper) GetCodeSize(addr common.Address) int {
	return len(k.GetCode(addr))
}

// ----------------------------------------------------------------------------
// Refund
// ----------------------------------------------------------------------------

// NOTE: gas refunded needs to be tracked and stored in a separate variable in
// order to add it subtract/add it from/to the gas used value adter the EVM
// execution has finalised.

// AddRefund adds the given amount of gas to the refund cached value
func (k *Keeper) AddRefund(gas uint64) {
	k.cache.refund += gas
}

// SubRefund subtracts the given amount of gas from the refund value. This function
// will panic if gas amount is greater than the stored refund.
func (k *Keeper) SubRefund(gas uint64) {
	if gas > k.cache.refund {
		// TODO: (@fedekunze) set to 0?? Geth panics here
		panic("refund counter below zero")
	}

	k.cache.refund -= gas
}

// GetRefund returns the amount of gas available for return after the tx execution
// finalises.
func (k *Keeper) GetRefund() uint64 {
	return k.cache.refund
}

// ----------------------------------------------------------------------------
// State
// ----------------------------------------------------------------------------

// GetCommittedState calls CommitStateDB.GetCommittedState using the passed in context
func (k *Keeper) GetCommittedState(addr common.Address, hash common.Hash) common.Hash {
	store := prefix.NewStore(k.ctx.KVStore(k.storeKey), types.AddressStoragePrefix(addr))

	// TODO: document logic
	key := types.GetStorageByAddressKey(addr, hash)
	value := store.Get(key.Bytes())
	if len(value) == 0 {
		return common.Hash{}
	}

	return common.BytesToHash(value)
}

// GetState calls CommitStateDB.GetState using the passed in context
func (k *Keeper) GetState(addr common.Address, hash common.Hash) common.Hash {
	// All state is committed directly
	return k.GetCommittedState(addr, hash)
}

// SetState calls CommitStateDB.SetState using the passed in context
func (k *Keeper) SetState(addr common.Address, key, value common.Hash) {
	store := prefix.NewStore(k.ctx.KVStore(k.storeKey), types.AddressStoragePrefix(addr))
	// TODO: document logic

	key = types.GetStorageByAddressKey(addr, key)

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

// TODO: (@fedekunze) consider removing the state immediately once Suicide has been
// called and store the map value in case address is queried again during the same
// execution. This will prevent us from having to iterate over the accounts and we
// can just reset the map during begin block.

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

	k.cache.suicided[addr] = true

	k.Logger(k.ctx).Debug(
		"account suicided",
		"ethereum-address", addr.Hex(),
		"cosmos-address", cosmosAddr.String(),
	)

	return true
}

// HasSuicided implements the vm.StoreDB interface
func (k *Keeper) HasSuicided(addr common.Address) bool {
	return k.cache.suicided[addr]
}

// ----------------------------------------------------------------------------
// Account Exist / Empty
// ----------------------------------------------------------------------------

// Exist returns true if the given account exists in store or if it has been
// marked as suicided.
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
func (k *Keeper) Empty(addr common.Address) bool {
	nonce := uint64(0)
	codeHash := types.EmptyCodeHash

	cosmosAddr := sdk.AccAddress(addr.Bytes())
	account := k.accountKeeper.GetAccount(k.ctx, cosmosAddr)

	if account != nil {
		nonce = account.GetSequence()
		ethAccount, isEthAccount := account.(*ethermint.EthAccount)
		if !isEthAccount {
			// NOTE: non-ethereum accounts are considered not empty
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

func (k *Keeper) AddressInAccessList(addr common.Address) bool {
	return k.cache.accessList.ContainsAddress(addr)
}

func (k *Keeper) SlotInAccessList(addr common.Address, slot common.Hash) (addressOk bool, slotOk bool) {
	return k.cache.accessList.Contains(addr, slot)
}

// AddAddressToAccessList adds the given address to the access list. This operation is safe to perform
// even if the feature/fork is not active yet
func (k *Keeper) AddAddressToAccessList(addr common.Address) {
	_ = k.cache.accessList.AddAddress(addr)
}

// AddSlotToAccessList adds the given (address,slot) to the access list. This operation is safe to perform
// even if the feature/fork is not active yet
func (k *Keeper) AddSlotToAccessList(addr common.Address, slot common.Hash) {
	_, _ = k.cache.accessList.AddSlot(addr, slot)
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

// AddLog calls CommitStateDB.AddLog using the passed in context
func (k *Keeper) AddLog(log *ethtypes.Log) {
	txHash := common.BytesToHash(tmtypes.Tx(k.ctx.TxBytes()).Hash())

	log.TxHash = txHash
	log.TxIndex = uint(k.cache.txIndex)
	log.BlockHash = k.cache.blockHash

	logs := k.GetTxLogs(txHash)
	log.Index = uint(len(logs))
	logs = append(logs, log)
	k.SetLogs(txHash, logs)

	// k.cache.journal.append(addLogChange{txhash: txHash})

	k.Logger(k.ctx).Debug(
		"log added",
		"tx-hash", txHash.Hex(),
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
