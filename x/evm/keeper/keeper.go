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
	"github.com/palantir/stacktrace"
	"github.com/tendermint/tendermint/libs/log"

	ethermint "github.com/tharsis/ethermint/types"
	"github.com/tharsis/ethermint/x/evm/types"
)

// Keeper grants access to the EVM module state and implements the go-ethereum StateDB interface.
type Keeper struct {
	// Protobuf codec
	cdc codec.BinaryCodec
	// Store key required for the EVM Prefix KVStore. It is required by:
	// - storing account's Storage State
	// - storing account's Code
	// - storing transaction Logs
	// - storing Bloom filters by block height. Needed for the Web3 API.
	storeKey sdk.StoreKey

	// key to access the transient store, which is reset on every block during Commit
	transientKey sdk.StoreKey

	// module specific parameter space that can be configured through governance
	paramSpace paramtypes.Subspace
	// access to account state
	accountKeeper types.AccountKeeper
	// update balance and accounting operations with coins
	bankKeeper types.BankKeeper
	// access historical headers for EVM state transition execution
	stakingKeeper types.StakingKeeper

	// Manage the initial context and cache context stack for accessing the store,
	// emit events and log info.
	// It is kept as a field to make is accessible by the StateDb
	// functions. Resets on every transaction/block.
	ctxStack ContextStack

	// chain ID number obtained from the context's chain id
	eip155ChainID *big.Int

	// Tracer used to collect execution traces from the EVM transaction execution
	tracer string
	// trace EVM state transition execution. This value is obtained from the `--trace` flag.
	// For more info check https://geth.ethereum.org/docs/dapp/tracing
	debug bool

	// EVM Hooks for tx post-processing
	hooks types.EvmHooks

	// error from previous state operation
	stateErr error
}

// NewKeeper generates new evm module keeper
func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey, transientKey sdk.StoreKey, paramSpace paramtypes.Subspace,
	ak types.AccountKeeper, bankKeeper types.BankKeeper, sk types.StakingKeeper,
	tracer string, debug bool,
) *Keeper {
	// ensure evm module account is set
	if addr := ak.GetModuleAddress(types.ModuleName); addr == nil {
		panic("the EVM module account has not been set")
	}

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
		tracer:        tracer,
		debug:         debug,
		stateErr:      nil,
	}
}

// Ctx returns the current context from the context stack
func (k Keeper) Ctx() sdk.Context {
	return k.ctxStack.CurrentContext()
}

// CommitCachedContexts commit all the cache contexts created by `StateDB.Snapshot`.
func (k *Keeper) CommitCachedContexts() {
	k.ctxStack.Commit()
}

// CachedContextsEmpty returns true if there's no cache contexts.
func (k *Keeper) CachedContextsEmpty() bool {
	return k.ctxStack.IsEmpty()
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", types.ModuleName)
}

// WithContext clears the context stack, and set the initial context.
func (k *Keeper) WithContext(ctx sdk.Context) {
	k.ctxStack.Reset(ctx)
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

// EmitBlockBloomEvent emit block bloom events
func (k Keeper) EmitBlockBloomEvent(ctx sdk.Context, bloom ethtypes.Bloom) {
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeBlockBloom,
			sdk.NewAttribute(types.AttributeKeyEthereumBloom, string(bloom.Bytes())),
		),
	)
}

// GetBlockBloomTransient returns bloom bytes for the current block height
func (k Keeper) GetBlockBloomTransient() *big.Int {
	store := prefix.NewStore(k.Ctx().TransientStore(k.transientKey), types.KeyPrefixTransientBloom)
	heightBz := sdk.Uint64ToBigEndian(uint64(k.Ctx().BlockHeight()))
	bz := store.Get(heightBz)
	if len(bz) == 0 {
		return big.NewInt(0)
	}

	return new(big.Int).SetBytes(bz)
}

// SetBlockBloomTransient sets the given bloom bytes to the transient store. This value is reset on
// every block.
func (k Keeper) SetBlockBloomTransient(bloom *big.Int) {
	store := prefix.NewStore(k.Ctx().TransientStore(k.transientKey), types.KeyPrefixTransientBloom)
	heightBz := sdk.Uint64ToBigEndian(uint64(k.Ctx().BlockHeight()))
	store.Set(heightBz, bloom.Bytes())
}

// ----------------------------------------------------------------------------
// Tx
// ----------------------------------------------------------------------------

// GetTxHashTransient returns the hash of current processing transaction
func (k Keeper) GetTxHashTransient() common.Hash {
	store := k.Ctx().TransientStore(k.transientKey)
	bz := store.Get(types.KeyPrefixTransientTxHash)
	if len(bz) == 0 {
		return common.Hash{}
	}

	return common.BytesToHash(bz)
}

// SetTxHashTransient set the hash of processing transaction
func (k Keeper) SetTxHashTransient(hash common.Hash) {
	store := k.Ctx().TransientStore(k.transientKey)
	store.Set(types.KeyPrefixTransientTxHash, hash.Bytes())
}

// SetTxIndexTransient set the index of processing transaction
func (k Keeper) SetTxIndexTransient(index uint64) {
	store := k.Ctx().TransientStore(k.transientKey)
	store.Set(types.KeyPrefixTransientTxIndex, sdk.Uint64ToBigEndian(index))
}

// GetTxIndexTransient returns EVM transaction index on the current block.
func (k Keeper) GetTxIndexTransient() uint64 {
	store := k.Ctx().TransientStore(k.transientKey)
	bz := store.Get(types.KeyPrefixTransientTxIndex)
	if len(bz) == 0 {
		return 0
	}

	return sdk.BigEndianToUint64(bz)
}

// IncreaseTxIndexTransient fetches the current EVM tx index from the transient store, increases its
// value by one and then sets the new index back to the transient store.
func (k Keeper) IncreaseTxIndexTransient() {
	txIndex := k.GetTxIndexTransient()
	k.SetTxIndexTransient(txIndex + 1)
}

// ResetRefundTransient resets the available refund amount to 0
func (k Keeper) ResetRefundTransient(ctx sdk.Context) {
	store := ctx.TransientStore(k.transientKey)
	store.Set(types.KeyPrefixTransientRefund, sdk.Uint64ToBigEndian(0))
}

// ----------------------------------------------------------------------------
// Log
// ----------------------------------------------------------------------------

// GetTxLogsTransient returns the current logs for a given transaction hash from the KVStore.
// This function returns an empty, non-nil slice if no logs are found.
func (k Keeper) GetTxLogsTransient(txHash common.Hash) []*ethtypes.Log {
	store := prefix.NewStore(k.Ctx().TransientStore(k.transientKey), types.KeyPrefixTransientTxLogs)

	// We store the logs with key equal to txHash.Bytes() | sdk.Uint64ToBigEndian(uint64(log.Index)),
	// therefore, we set the end boundary(excluded) to txHash.Bytes() | uint64.Max -> []byte
	end := txHash.Bytes()
	end = append(end, []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}...)

	iter := store.Iterator(txHash.Bytes(), end)
	defer iter.Close()

	logs := []*ethtypes.Log{}
	for ; iter.Valid(); iter.Next() {
		var log types.Log
		k.cdc.MustUnmarshal(iter.Value(), &log)
		logs = append(logs, log.ToEthereum())
	}

	return logs
}

// SetLog sets the log for a transaction in the KVStore.
func (k Keeper) AddLogTransient(log *ethtypes.Log) {
	store := prefix.NewStore(k.Ctx().TransientStore(k.transientKey), types.KeyPrefixTransientTxLogs)

	key := log.TxHash.Bytes()
	key = append(key, sdk.Uint64ToBigEndian(uint64(log.Index))...)

	txIndexLog := types.NewLogFromEth(log)
	bz := k.cdc.MustMarshal(txIndexLog)
	store.Set(key, bz)
}

// GetLogSizeTransient returns EVM log index on the current block.
func (k Keeper) GetLogSizeTransient() uint64 {
	store := k.Ctx().TransientStore(k.transientKey)
	bz := store.Get(types.KeyPrefixTransientLogSize)
	if len(bz) == 0 {
		return 0
	}

	return sdk.BigEndianToUint64(bz)
}

// IncreaseLogSizeTransient fetches the current EVM log index from the transient store, increases its
// value by one and then sets the new index back to the transient store.
func (k Keeper) IncreaseLogSizeTransient() {
	logSize := k.GetLogSizeTransient()
	store := k.Ctx().TransientStore(k.transientKey)
	store.Set(types.KeyPrefixTransientLogSize, sdk.Uint64ToBigEndian(logSize+1))
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
	store := prefix.NewStore(k.Ctx().KVStore(k.storeKey), types.AddressStoragePrefix(addr))
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

	store := prefix.NewStore(k.Ctx().KVStore(k.storeKey), types.KeyPrefixCode)
	store.Delete(hash.Bytes())
}

// ClearBalance subtracts the EVM all the balance denomination from the address
// balance while also updating the total supply.
func (k Keeper) ClearBalance(addr sdk.AccAddress) (prevBalance sdk.Coin, err error) {
	params := k.GetParams(k.Ctx())

	prevBalance = k.bankKeeper.GetBalance(k.Ctx(), addr, params.EvmDenom)
	if prevBalance.IsPositive() {
		if err := k.bankKeeper.SendCoinsFromAccountToModule(k.Ctx(), addr, types.ModuleName, sdk.Coins{prevBalance}); err != nil {
			return sdk.Coin{}, stacktrace.Propagate(err, "failed to transfer to module account")
		}

		if err := k.bankKeeper.BurnCoins(k.Ctx(), types.ModuleName, sdk.Coins{prevBalance}); err != nil {
			return sdk.Coin{}, stacktrace.Propagate(err, "failed to burn coins from evm module account")
		}
	}

	return prevBalance, nil
}

// ResetAccount removes the code, storage state, but keep all the native tokens stored
// with the given address.
func (k Keeper) ResetAccount(addr common.Address) {
	k.DeleteCode(addr)
	k.DeleteAccountStorage(addr)
}

// SetHooks sets the hooks for the EVM module
func (k *Keeper) SetHooks(eh types.EvmHooks) *Keeper {
	if k.hooks != nil {
		panic("cannot set evm hooks twice")
	}

	k.hooks = eh
	return k
}

// PostTxProcessing delegate the call to the hooks. If no hook has been registered, this function returns with a `nil` error
func (k *Keeper) PostTxProcessing(txHash common.Hash, logs []*ethtypes.Log) error {
	if k.hooks == nil {
		return nil
	}
	return k.hooks.PostTxProcessing(k.Ctx(), txHash, logs)
}
