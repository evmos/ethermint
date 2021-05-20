package keeper

import (
	"bytes"
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

	// Journal of state modifications. This is the backbone of
	// Snapshot and RevertToSnapshot.
	journal        *types.Journal
	validRevisions []types.Revision
	nextRevisionID int

	// Per-transaction access list
	accessList *types.AccessListMappings

	// Transaction counter in a block. Used on StateSB's Prepare function.
	// It is reset to 0 every block on BeginBlock so there's no point in storing the counter
	// on the KVStore or adding it as a field on the EVM genesis state.
	txIndex int
	bloom   *big.Int

	// logs is a cache field that keeps mapping of contract address -> eth logs emitted
	// during EVM execution in the current block.
	logs map[common.Address][]*ethtypes.Log

	// accounts that are suicided will be returned as non-nil and will be deleted at every.
	suicided         map[common.Address]bool
	suicidedAccounts []common.Address
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
		// k.cache.journal.append(createObjectChange{account: &addr})
		log = "account created"
	} else {
		// TODO: add journal to ResetAccount func?
		// k.cache.journal.append(resetObjectChange{prev: account})
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

	prevBalance := k.bankKeeper.GetBalance(k.ctx, cosmosAddr, params.EvmDenom).Amount.BigInt()

	if err := k.bankKeeper.AddCoins(k.ctx, cosmosAddr, coins); err != nil {
		k.Logger(k.ctx).Error(
			"failed to add balance",
			"ethereum-address", addr.Hex(),
			"cosmos-address", cosmosAddr.String(),
			"error", err,
		)

		return
	}

	// k.cache.journal.append(balanceChange{
	// 	account: addr,
	// 	prev:    prevBalance,
	// })

	k.Logger(k.ctx).Debug(
		"balance addition",
		"ethereum-address", addr.Hex(),
		"cosmos-address", cosmosAddr.String(),
		"previous-balance", prevBalance.String(),
		"new-balance", big.NewInt(0).Add(prevBalance, amount).String(),
	)
}

// SubBalance calls CommitStateDB.SubBalance using the passed in context
func (k *Keeper) SubBalance(addr common.Address, amount *big.Int) {
	cosmosAddr := sdk.AccAddress(addr.Bytes())

	params := k.GetParams(k.ctx)
	coins := sdk.Coins{sdk.NewCoin(params.EvmDenom, sdk.NewIntFromBigInt(amount))}

	prevBalance := k.bankKeeper.GetBalance(k.ctx, cosmosAddr, params.EvmDenom).Amount.BigInt()

	if err := k.bankKeeper.SubtractCoins(k.ctx, cosmosAddr, coins); err != nil {
		k.Logger(k.ctx).Error(
			"failed to subtract balance",
			"ethereum-address", addr.Hex(),
			"cosmos-address", cosmosAddr.String(),
			"error", err,
		)

		return
	}

	// k.cache.journal.append(balanceChange{
	// 	account: addr,
	// 	prev:    prevBalance,
	// })

	// if k.Empty(addr) {
	//  NOTE: Ensure journal is updated with the changes
	// 	DeleteAccount, balance, code, storage
	// }

	k.Logger(k.ctx).Debug(
		"balance substraction",
		"ethereum-address", addr.Hex(),
		"cosmos-address", cosmosAddr.String(),
		"previous-balance", prevBalance.String(),
		"new-balance", big.NewInt(0).Sub(prevBalance, amount).String(),
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

	// k.cache.journal.append(nonceChange{
	// 	account: addr,
	// 	prev:    prevNonce,
	// })
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

	// prevHash := common.BytesToHash(ethAccount.CodeHash)
	// prevCode := k.GetCode(addr)

	ethAccount.CodeHash = hash.Bytes()
	k.accountKeeper.SetAccount(k.ctx, ethAccount)

	store := prefix.NewStore(k.ctx.KVStore(k.storeKey), types.KeyPrefixCode)
	store.Set(hash.Bytes(), code)

	k.Logger(k.ctx).Debug(
		"code updated",
		"ethereum-address", addr.Hex(),
		"code-hash", hash.Hex(),
	)

	// k.cache.journal.append(codeChange{
	// 	account:  addr,
	// 	prevHash: prevHash,
	// 	prevCode: prevCode,
	// })

	// if len(code) == 0 && k.Empty(addr) {
	//  TODO: Ensure journal is updated with the changes
	// 	DeleteAccount, balance, code, storage
	// }
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
	// k.cache.journal.append(refundChange{prev: k.refund})

	// TODO: refund to transaction gas meter or block gas meter?
	k.cache.refund += gas
}

// SubRefund calls CommitStateDB.SubRefund using the passed in context
func (k *Keeper) SubRefund(gas uint64) {
	// TODO: implement
	// k.cache.journal.append(refundChange{prev: k.refund})

	if gas > k.cache.refund {
		panic("refund counter below zero")
	}

	// TODO: refund to transaction gas meter or block gas meter?
	k.cache.refund -= gas
}

// GetRefund calls CommitStateDB.GetRefund using the passed in context
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
	// All state is commited directly
	return k.GetCommittedState(addr, hash)
}

// SetState calls CommitStateDB.SetState using the passed in context
func (k *Keeper) SetState(addr common.Address, key, value common.Hash) {
	store := prefix.NewStore(k.ctx.KVStore(k.storeKey), types.AddressStoragePrefix(addr))
	// TODO: document logic
	// prevValue := k.GetState(addr, key)

	key = types.GetStorageByAddressKey(addr, key)
	store.Set(key.Bytes(), value.Bytes())

	k.Logger(k.ctx).Debug(
		"state updated",
		"ethereum-address", addr.Hex(),
		"key", key.Hex(),
	)

	// since the new value is different, update and journal the change
	// k.cache.journal.append(storageChange{
	// 	account:   addr,
	// 	key:       key,
	// 	prevValue: prevValue,
	// })
}

// ----------------------------------------------------------------------------
// Suicide
// ----------------------------------------------------------------------------

// TODO: (@fedekunze) consider prunning suicide records after unbonding period?

// Suicide marks the given account as suicided and clears the account balance.
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

	k.cache.suicided[addr] = true
	k.cache.suicidedAccounts = append(k.cache.suicidedAccounts, addr)

	// k.cache.journal.append(suicideChange{
	// 	account:     &addr,
	// 	prev:        prev,
	// 	prevBalance: prevBalance,
	// })

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
	return k.cache.accessList.ContainsAddress(addr)
}

func (k *Keeper) SlotInAccessList(addr common.Address, slot common.Hash) (addressOk bool, slotOk bool) {
	return k.cache.accessList.Contains(addr, slot)
}

// AddAddressToAccessList adds the given address to the access list. This operation is safe to perform
// even if the feature/fork is not active yet
func (k *Keeper) AddAddressToAccessList(addr common.Address) {
	if k.cache.accessList.AddAddress(addr) {
		// 	k.cache.journal.append(accessListAddAccountChange{&addr})
	}
}

// AddSlotToAccessList adds the given (address,slot) to the access list. This operation is safe to perform
// even if the feature/fork is not active yet
func (k *Keeper) AddSlotToAccessList(addr common.Address, slot common.Hash) {
	addrMod, slotMod := k.cache.accessList.AddSlot(addr, slot)
	if addrMod {
		// In practice, this should not happen, since there is no way to enter the
		// scope of 'address' without having the 'address' become already added
		// to the access list (via call-variant, create, etc).
		// Better safe than sorry, though
		// k.cache.journal.append(accessListAddAccountChange{&addr})
	}
	if slotMod {
		// k.cache.journal.append(accessListAddSlotChange{
		// 	address: &addr,
		// 	slot:    &slot,
		// })
	}
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
	// 		journalIndex: k.cache.journal.length(),
	// 	},
	// )

	return id
}

// RevertToSnapshot calls CommitStateDB.RevertToSnapshot using the passed in context
func (k *Keeper) RevertToSnapshot(revID int) {
	// // find the snapshot in the stack of valid snapshots
	// idx := sort.Search(len(k.validRevisions), func(i int) bool {
	// 	return k.cache.validRevisionsi].id >= revID
	// })

	// if idx == len(k.validRevisions) || k.cache.validRevisionsidx].id != revID {
	// 	panic(fmt.Errorf("revision ID %v cannot be reverted", revID))
	// }

	// snapshot := k.cache.validRevisionsidx].journalIndex

	// // replay the journal to undo changes and remove invalidated snapshots
	// k.cache.journal.revert(csdb, snapshot)
	// k.validRevisions = k.cache.validRevisions:idx]

	k.Logger(k.ctx).Debug(
		"reverted to snapshot",
		"revision-id", revID,
	)
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
	log.TxIndex = uint(k.cache.txIndex)

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
