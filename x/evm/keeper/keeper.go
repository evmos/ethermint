package keeper

import (
	"bytes"
	"math/big"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/tendermint/tendermint/libs/log"

	ethermint "github.com/cosmos/ethermint/types"
	"github.com/cosmos/ethermint/x/evm/types"
)

// Keeper wraps the CommitStateDB, allowing us to pass in SDK context while adhering
// to the StateDB interface.
type Keeper struct {
	// Protobuf codec
	cdc codec.BinaryMarshaler
	// Store key required for the EVM Prefix KVStore. It is required by:
	// - storing Account's Storage State
	// - storing Account's Code
	// - storing transaction Logs
	// - storing block height -> bloom filter map. Needed for the Web3 API.
	storeKey sdk.StoreKey

	// key to access the transient store, which is reset on every block during Commit
	transientKey sdk.StoreKey

	paramSpace    paramtypes.Subspace
	accountKeeper types.AccountKeeper
	bankKeeper    types.BankKeeper
	stakingKeeper types.StakingKeeper

	ctx sdk.Context
	// chain ID number obtained from the context's chain id
	eip155ChainID *big.Int
	debug         bool
}

// NewKeeper generates new evm module keeper
func NewKeeper(
	cdc codec.BinaryMarshaler, storeKey, transientKey sdk.StoreKey, paramSpace paramtypes.Subspace,
	ak types.AccountKeeper, bankKeeper types.BankKeeper, sk types.StakingKeeper,
) *Keeper {
	// set KeyTable if it has not already been set
	if !paramSpace.HasKeyTable() {
		paramSpace = paramSpace.WithKeyTable(types.ParamKeyTable())
	}

	// NOTE: we pass in the parameter space to the CommitStateDB in order to use custom denominations for the EVM operations
	return &Keeper{
		cdc:           cdc,
		paramSpace:    paramSpace,
		accountKeeper: ak,
		bankKeeper:    bankKeeper,
		stakingKeeper: sk,
		storeKey:      storeKey,
		transientKey:  transientKey,
	}
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", types.ModuleName)
}

// WithContext sets an updated SDK context to the keeper
func (k *Keeper) WithContext(ctx sdk.Context) {
	k.ctx = ctx
}

// WithChainID sets the chain id to the local variable in the keeper
func (k *Keeper) WithChainID(ctx sdk.Context) {
	chainID, err := ethermint.ParseChainID(ctx.ChainID())
	if err != nil {
		panic(err)
	}

	if k.eip155ChainID != nil && k.eip155ChainID.Cmp(chainID) != 0 {
		panic("chain id already set")
	}

	k.eip155ChainID = chainID
}

// ChainID returns the EIP155 chain ID for the EVM context
func (k Keeper) ChainID() *big.Int {
	return k.eip155ChainID
}

// ----------------------------------------------------------------------------
// Block Bloom
// Required by Web3 API.
// ----------------------------------------------------------------------------

// GetBlockBloom gets bloombits from block height
func (k Keeper) GetBlockBloom(ctx sdk.Context, height int64) (ethtypes.Bloom, bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.BloomKey(height))
	if len(bz) == 0 {
		return ethtypes.Bloom{}, false
	}

	return ethtypes.BytesToBloom(bz), true
}

// SetBlockBloom sets the mapping from block height to bloom bits
func (k Keeper) SetBlockBloom(ctx sdk.Context, height int64, bloom ethtypes.Bloom) {
	store := ctx.KVStore(k.storeKey)

	key := types.BloomKey(height)
	store.Set(key, bloom.Bytes())
}

// GetBlockBloomTransient returns bloom bytes for the current block height
func (k Keeper) GetBlockBloomTransient() (*big.Int, bool) {
	store := k.ctx.TransientStore(k.transientKey)
	bz := store.Get(types.KeyPrefixTransientBloom)
	if len(bz) == 0 {
		return nil, false
	}

	return new(big.Int).SetBytes(bz), true
}

// SetBlockBloomTransient sets the given bloom bytes to the transient store. This value is reset on
// every block.
func (k Keeper) SetBlockBloomTransient(bloom *big.Int) {
	store := k.ctx.TransientStore(k.transientKey)
	store.Set(types.KeyPrefixTransientBloom, bloom.Bytes())
}

// ----------------------------------------------------------------------------
// Tx
// ----------------------------------------------------------------------------

// GetTxIndexTransient returns EVM transaction index on the current block.
func (k Keeper) GetTxIndexTransient() uint64 {
	store := k.ctx.TransientStore(k.transientKey)
	bz := store.Get(types.KeyPrefixTransientBloom)
	if len(bz) == 0 {
		return 0
	}

	return sdk.BigEndianToUint64(bz)
}

// IncreaseTxIndexTransient fetches the current EVM tx index from the transient store, increases its
// value by one and then sets the new index back to the transient store.
func (k Keeper) IncreaseTxIndexTransient() {
	txIndex := k.GetTxIndexTransient()
	store := k.ctx.TransientStore(k.transientKey)
	store.Set(types.KeyPrefixTransientBloom, sdk.Uint64ToBigEndian(txIndex+1))
}

// ResetRefundTransient resets the available refund amount to 0
func (k Keeper) ResetRefundTransient(ctx sdk.Context) {
	store := ctx.TransientStore(k.transientKey)
	store.Set(types.KeyPrefixTransientRefund, sdk.Uint64ToBigEndian(0))
}

// ----------------------------------------------------------------------------
// Log
// ----------------------------------------------------------------------------

// GetAllTxLogs return all the transaction logs from the store.
func (k Keeper) GetAllTxLogs(ctx sdk.Context) []types.TransactionLogs {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, types.KeyPrefixLogs)
	defer iterator.Close()

	txsLogs := []types.TransactionLogs{}
	for ; iterator.Valid(); iterator.Next() {
		var txLog types.TransactionLogs
		k.cdc.MustUnmarshalBinaryBare(iterator.Value(), &txLog)

		// add a new entry
		txsLogs = append(txsLogs, txLog)
	}
	return txsLogs
}

// GetLogs returns the current logs for a given transaction hash from the KVStore.
// This function returns an empty, non-nil slice if no logs are found.
func (k Keeper) GetTxLogs(txHash common.Hash) []*ethtypes.Log {
	store := prefix.NewStore(k.ctx.KVStore(k.storeKey), types.KeyPrefixLogs)

	bz := store.Get(txHash.Bytes())
	if len(bz) == 0 {
		return []*ethtypes.Log{}
	}

	var logs types.TransactionLogs
	k.cdc.MustUnmarshalBinaryBare(bz, &logs)

	return logs.EthLogs()
}

// SetLogs sets the logs for a transaction in the KVStore.
func (k Keeper) SetLogs(txHash common.Hash, logs []*ethtypes.Log) {
	store := prefix.NewStore(k.ctx.KVStore(k.storeKey), types.KeyPrefixLogs)

	txLogs := types.NewTransactionLogsFromEth(txHash, logs)
	bz := k.cdc.MustMarshalBinaryBare(&txLogs)

	store.Set(txHash.Bytes(), bz)
}

// DeleteLogs removes the logs from the KVStore. It is used during journal.Revert.
func (k Keeper) DeleteTxLogs(ctx sdk.Context, txHash common.Hash) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixLogs)
	store.Delete(txHash.Bytes())
}

// ----------------------------------------------------------------------------
// Storage
// ----------------------------------------------------------------------------

// GetAccountStorage return state storage associated with an account
func (k Keeper) GetAccountStorage(ctx sdk.Context, address common.Address) (types.Storage, error) {
	storage := types.Storage{}

	err := k.ForEachStorage(address, func(key, value common.Hash) bool {
		storage = append(storage, types.NewState(key, value))
		return false
	})

	if err != nil {
		return types.Storage{}, err
	}

	return storage, nil
}

// ----------------------------------------------------------------------------
// Account
// ----------------------------------------------------------------------------

func (k Keeper) DeleteState(addr common.Address, key common.Hash) {
	store := prefix.NewStore(k.ctx.KVStore(k.storeKey), types.AddressStoragePrefix(addr))
	key = types.KeyAddressStorage(addr, key)
	store.Delete(key.Bytes())
}

// DeleteAccountStorage clears all the storage state associated with the given address.
func (k Keeper) DeleteAccountStorage(addr common.Address) {
	_ = k.ForEachStorage(addr, func(key, _ common.Hash) bool {
		k.DeleteState(addr, key)
		return false
	})
}

// DeleteCode removes the contract code byte array from the store associated with
// the given address.
func (k Keeper) DeleteCode(addr common.Address) {
	hash := k.GetCodeHash(addr)
	if bytes.Equal(hash.Bytes(), common.BytesToHash(types.EmptyCodeHash).Bytes()) {
		return
	}

	store := prefix.NewStore(k.ctx.KVStore(k.storeKey), types.KeyPrefixCode)
	store.Delete(hash.Bytes())
}

// ClearBalance subtracts the EVM all the balance denomination from the address
// balance while also updating the total supply.
func (k Keeper) ClearBalance(addr sdk.AccAddress) (prevBalance sdk.Coin, err error) {
	params := k.GetParams(k.ctx)

	prevBalance = k.bankKeeper.GetBalance(k.ctx, addr, params.EvmDenom)
	if prevBalance.IsPositive() {
		err := k.bankKeeper.SubtractCoins(k.ctx, addr, sdk.Coins{prevBalance})
		if err != nil {
			return sdk.Coin{}, err
		}
	}

	return prevBalance, nil
}

// ResetAccount removes the code, storage state and evm denom balance coins stored
// with the given address.
func (k Keeper) ResetAccount(addr common.Address) {
	k.DeleteCode(addr)
	k.DeleteAccountStorage(addr)
	_, err := k.ClearBalance(addr.Bytes())
	if err != nil {
		k.Logger(k.ctx).Error(
			"failed to clear balance during account reset",
			"ethereum-address", addr.Hex(),
		)
	}
}
