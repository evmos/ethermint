package keeper

import (
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"

	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmos/ethermint/x/evm/types"
)

// BeginBlock sets the block hash -> block height map for the previous block height
// and resets the Bloom filter and the transaction count to 0.
func (k *Keeper) BeginBlock(ctx sdk.Context, req abci.RequestBeginBlock) {
	defer telemetry.ModuleMeasureSince(types.ModuleName, time.Now(), telemetry.MetricKeyBeginBlocker)
	k.WithContext(ctx)
	k.headerHash = common.BytesToHash(req.Hash)
}

// EndBlock updates the accounts and commits state objects to the KV Store, while
// deleting the empty ones. It also sets the bloom filers for the request block to
// the store. The EVM end block logic doesn't update the validator set, thus it returns
// an empty slice.
func (k Keeper) EndBlock(ctx sdk.Context, req abci.RequestEndBlock) []abci.ValidatorUpdate {
	defer telemetry.ModuleMeasureSince(types.ModuleName, time.Now(), telemetry.MetricKeyEndBlocker)

	// Gas costs are handled within msg handler so costs should be ignored
	ctx = ctx.WithGasMeter(sdk.NewInfiniteGasMeter())
	k.CommitStateDB.WithContext(ctx)
	k.WithContext(ctx)

	// Update account balances before committing other parts of state
	k.CommitStateDB.UpdateAccounts()

	root, err := k.CommitStateDB.Commit(true)
	// Commit state objects to KV store
	if err != nil {
		k.Logger(ctx).Error("failed to commit state objects", "error", err, "height", ctx.BlockHeight())
		panic(err)
	}

	// reset all cache after account data has been committed, that make sure node state consistent
	if err = k.CommitStateDB.Reset(root); err != nil {
		panic(err)
	}

	// get the block bloom bytes from the transient store and set it to the persistent storage
	bloomBig, found := k.GetBlockBloomTransient()
	if !found {
		bloomBig = big.NewInt(0)
	}

	bloom := ethtypes.BytesToBloom(bloomBig.Bytes())
	k.SetBlockBloom(ctx, req.Height, bloom)

	return []abci.ValidatorUpdate{}
}
