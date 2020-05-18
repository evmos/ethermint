package keeper

import (
	"errors"
	"fmt"
	"time"

	"github.com/tendermint/tendermint/libs/log"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	supplyexported "github.com/cosmos/cosmos-sdk/x/supply/exported"

	"github.com/cosmos/ethermint/x/faucet/types"
)

// Keeper defines the faucet Keeper.
type Keeper struct {
	cdc          *codec.Codec
	storeKey     sdk.StoreKey
	supplyKeeper types.SupplyKeeper

	// History of users and their funding timeouts. They are reset if the app is reinitialized.
	timeouts map[string]time.Time
}

// NewKeeper creates a new faucet Keeper instance.
func NewKeeper(
	cdc *codec.Codec, storeKey sdk.StoreKey, supplyKeeper types.SupplyKeeper,
) Keeper {
	return Keeper{
		cdc:          cdc,
		storeKey:     storeKey,
		supplyKeeper: supplyKeeper,
		timeouts:     make(map[string]time.Time),
	}
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// GetFaucetAccount returns the faucet ModuleAccount
func (k Keeper) GetFaucetAccount(ctx sdk.Context) supplyexported.ModuleAccountI {
	return k.supplyKeeper.GetModuleAccount(ctx, types.ModuleName)
}

// Fund checks for timeout and max thresholds and then mints coins and transfers
// coins to the recipient.
func (k Keeper) Fund(ctx sdk.Context, amount sdk.Coins, recipient sdk.AccAddress) error {
	if !k.IsEnabled(ctx) {
		return errors.New("faucet is not enabled. Restart the application and set faucet's 'enable_faucet' genesis field to true")
	}

	if err := k.rateLimit(ctx, recipient.String()); err != nil {
		return err
	}

	totalRequested := sdk.ZeroInt()
	for _, coin := range amount {
		totalRequested = totalRequested.Add(coin.Amount)
	}

	maxPerReq := k.GetMaxPerRequest(ctx)
	if totalRequested.GT(maxPerReq) {
		return fmt.Errorf("canot fund more than %s per request. requested %s", maxPerReq, totalRequested)
	}

	funded := k.GetFunded(ctx)
	totalFunded := sdk.ZeroInt()
	for _, coin := range funded {
		totalFunded = totalFunded.Add(coin.Amount)
	}

	cap := k.GetCap(ctx)

	if totalFunded.Add(totalRequested).GT(cap) {
		return fmt.Errorf("maximum cap of %s reached. Cannot continue funding", cap)
	}

	if err := k.supplyKeeper.MintCoins(ctx, types.ModuleName, amount); err != nil {
		return err
	}

	if err := k.supplyKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, recipient, amount); err != nil {
		return err
	}

	k.SetFunded(ctx, funded.Add(amount...))

	k.Logger(ctx).Info(fmt.Sprintf("funded %s to %s", amount, recipient))
	return nil
}

func (k Keeper) GetTimeout(ctx sdk.Context) time.Duration {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.TimeoutKey)
	if len(bz) == 0 {
		return time.Duration(0)
	}

	var timeout time.Duration
	k.cdc.MustUnmarshalBinaryBare(bz, &timeout)

	return timeout
}

func (k Keeper) SetTimout(ctx sdk.Context, timeout time.Duration) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshalBinaryBare(timeout)
	store.Set(types.TimeoutKey, bz)
}

func (k Keeper) IsEnabled(ctx sdk.Context) bool {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.EnableFaucetKey)
	if len(bz) == 0 {
		return false
	}

	var enabled bool
	k.cdc.MustUnmarshalBinaryBare(bz, &enabled)
	return enabled
}

func (k Keeper) SetEnabled(ctx sdk.Context, enabled bool) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshalBinaryBare(enabled)
	store.Set(types.EnableFaucetKey, bz)
}

func (k Keeper) GetCap(ctx sdk.Context) sdk.Int {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.CapKey)
	if len(bz) == 0 {
		return sdk.ZeroInt()
	}

	var cap sdk.Int
	k.cdc.MustUnmarshalBinaryBare(bz, &cap)

	return cap
}

func (k Keeper) SetCap(ctx sdk.Context, cap sdk.Int) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshalBinaryBare(cap)
	store.Set(types.CapKey, bz)
}

func (k Keeper) GetMaxPerRequest(ctx sdk.Context) sdk.Int {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.MaxPerRequestKey)
	if len(bz) == 0 {
		return sdk.ZeroInt()
	}

	var maxPerReq sdk.Int
	k.cdc.MustUnmarshalBinaryBare(bz, &maxPerReq)

	return maxPerReq
}

func (k Keeper) SetMaxPerRequest(ctx sdk.Context, maxPerReq sdk.Int) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshalBinaryBare(maxPerReq)
	store.Set(types.MaxPerRequestKey, bz)
}

func (k Keeper) GetFunded(ctx sdk.Context) sdk.Coins {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.FundedKey)
	if len(bz) == 0 {
		return nil
	}

	var funded sdk.Coins
	k.cdc.MustUnmarshalBinaryBare(bz, &funded)

	return funded
}

func (k Keeper) SetFunded(ctx sdk.Context, funded sdk.Coins) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshalBinaryBare(funded)
	store.Set(types.FundedKey, bz)
}

func (k Keeper) rateLimit(ctx sdk.Context, address string) error {
	// first time requester, can send request
	lastRequest, ok := k.timeouts[address]
	if !ok {
		k.timeouts[address] = time.Now().UTC()
		return nil
	}

	defaultTimeout := k.GetTimeout(ctx)
	sinceLastRequest := time.Since(lastRequest)

	if defaultTimeout > sinceLastRequest {
		wait := defaultTimeout - sinceLastRequest
		return fmt.Errorf("%s has requested funds within the last %s, wait %s before trying again", address, defaultTimeout.String(), wait.String())
	}

	// user able to send funds since they have waited for period
	k.timeouts[address] = time.Now().UTC()
	return nil
}
