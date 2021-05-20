package keeper

import (
	"bytes"
	"math/big"
	"strconv"

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

	// Journal of state modifications. This is the backbone of
	// Snapshot and RevertToSnapshot.
	journal        *types.Journal
	validRevisions []types.Revision
	nextRevisionID int

	// Per-transaction access list
	accessList *types.AccessListMappings
}

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
	store.Set(hash.Bytes(), code)

	k.Logger(k.ctx).Debug(
		"code updated",
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

// AddRefund calls CommitStateDB.AddRefund using the passed in context
func (k *Keeper) AddRefund(gas uint64) {
	// TODO: implement
	// k.journal.append(refundChange{prev: k.refund})

	// TODO: refund to transaction gas meter or block gas meter?
	// k.refund += gas
}

// SubRefund calls CommitStateDB.SubRefund using the passed in context
func (k *Keeper) SubRefund(gas uint64) {
	// TODO: implement
	// k.journal.append(refundChange{prev: k.refund})

	// if gas > k.refund {
	// 	panic("refund counter below zero")
	// }

	// TODO: refund to transaction gas meter or block gas meter?
	// k.refund -= gas
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
	// All state is commited directly
	return k.GetCommittedState(addr, hash)
}

// SetState calls CommitStateDB.SetState using the passed in context
func (k *Keeper) SetState(addr common.Address, key, value common.Hash) {
	store := prefix.NewStore(k.ctx.KVStore(k.storeKey), types.AddressStoragePrefix(addr))
	// TODO: document logic
	key = types.GetStorageByAddressKey(addr, key)
	store.Set(key.Bytes(), value.Bytes())
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

// PrepareAccessList handles the preparatory steps for executing a state transition with
// regards to both EIP-2929 and EIP-2930:
//
// - Add sender to access list (2929)
// - Add destination to access list (2929)
// - Add precompiles to access list (2929)
// - Add the contents of the optional tx access list (2930)
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
	// TODO: implement
	// return k.accessList.ContainsAddress(addr)
	return false
}

func (k *Keeper) SlotInAccessList(addr common.Address, slot common.Hash) (addressOk bool, slotOk bool) {
	// TODO: implement
	// return k.accessList.Contains(addr, slot)
	return false, false
}

// AddAddressToAccessList adds the given address to the access list. This operation is safe to perform
// even if the feature/fork is not active yet
func (k *Keeper) AddAddressToAccessList(addr common.Address) {
	// TODO: implement
	// if k.accessList.AddAddress(addr) {
	// 	k.journal.append(accessListAddAccountChange{&addr})
	// }
}

// AddSlotToAccessList adds the given (address,slot) to the access list. This operation is safe to perform
// even if the feature/fork is not active yet
func (k *Keeper) AddSlotToAccessList(addr common.Address, slot common.Hash) {
	// // TODO: implement
	// addrMod, slotMod := k.accessList.AddSlot(addr, slot)
	// if addrMod {
	// 	// In practice, this should not happen, since there is no way to enter the
	// 	// scope of 'address' without having the 'address' become already added
	// 	// to the access list (via call-variant, create, etc).
	// 	// Better safe than sorry, though
	// 	k.journal.append(accessListAddAccountChange{&addr})
	// }
	// if slotMod {
	// 	k.journal.append(accessListAddSlotChange{
	// 		address: &addr,
	// 		slot:    &slot,
	// 	})
	// }
}

// ----------------------------------------------------------------------------
// Snapshotting
// ----------------------------------------------------------------------------

// TODO: (@fedekunze) The Cosmos SDK has a Snapshotter (https://github.com/cosmos/cosmos-sdk/blob/v0.42.4/snapshots/types/snapshotter.go#L8)
// interface that allows for state snapshots and reverts to a given snapshot.
// Unfortunately, this doesn't allow for the snapshot of a subtree (in this case the EVM). Ideally
// this funcationality should be included in the SMT work. See https://github.com/cosmos/cosmos-sdk/discussions/8297 for
// more details.
// Coordinate with the LazyLedger and Regen teams on this work

// Snapshot calls CommitStateDB.Snapshot using the passed in context
func (k *Keeper) Snapshot() int {
	id := 0
	// id := k.nextRevisionID
	// k.nextRevisionID++

	// k.validRevisions = append(
	// 	k.validRevisions,
	// 	revision{
	// 		id:           id,
	// 		journalIndex: k.journal.length(),
	// 	},
	// )

	return id
}

// RevertToSnapshot calls CommitStateDB.RevertToSnapshot using the passed in context
func (k *Keeper) RevertToSnapshot(revID int) {
	// // find the snapshot in the stack of valid snapshots
	// idx := sort.Search(len(k.validRevisions), func(i int) bool {
	// 	return k.validRevisions[i].id >= revID
	// })

	// if idx == len(k.validRevisions) || k.validRevisions[idx].id != revID {
	// 	panic(fmt.Errorf("revision ID %v cannot be reverted", revID))
	// }

	// snapshot := k.validRevisions[idx].journalIndex

	// // replay the journal to undo changes and remove invalidated snapshots
	// k.journal.revert(csdb, snapshot)
	// k.validRevisions = k.validRevisions[:idx]
}

// ----------------------------------------------------------------------------
// Log
// ----------------------------------------------------------------------------

// AddLog calls CommitStateDB.AddLog using the passed in context
func (k *Keeper) AddLog(log *ethtypes.Log) {
	txHash := common.BytesToHash(tmtypes.Tx(k.ctx.TxBytes()).Hash())
	blockHash, found := k.GetBlockHashFromHeight(k.ctx, k.ctx.BlockHeight())
	if found {
		log.BlockHash = blockHash
	}

	log.TxHash = txHash

	log.TxIndex = uint(k.TxIndex)

	// k.journal.append(addLogChange{txhash: txHash})

	logs := k.GetTxLogs(txHash)
	log.Index = uint(len(logs))
	logs = append(logs, log)
	k.SetLogs(txHash, logs)
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
