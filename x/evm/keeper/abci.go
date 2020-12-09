package keeper

import (
	"math/big"

	abci "github.com/tendermint/tendermint/abci/types"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
)

// BeginBlock sets the block hash -> block height map for the previous block height
// and resets the Bloom filter and the transaction count to 0.
func (k *Keeper) BeginBlock(ctx sdk.Context, req abci.RequestBeginBlock) {
	if req.Header.LastBlockId.GetHash() == nil || req.Header.GetHeight() < 1 {
		return
	}

	// Gas costs are handled within msg handler so costs should be ignored
	ctx = ctx.WithGasMeter(sdk.NewInfiniteGasMeter())

	// Set the hash -> height and height -> hash mapping.
	hash := req.Header.LastBlockId.GetHash()
	height := req.Header.GetHeight() - 1

	k.SetHeightHash(ctx, uint64(height), common.BytesToHash(hash))
	k.SetBlockHash(ctx, hash, height)

	// reset counters that are used on CommitStateDB.Prepare
	k.Bloom = big.NewInt(0)
	k.TxCount = 0
}

// EndBlock updates the accounts and commits state objects to the KV Store, while
// deleting the empty ones. It also sets the bloom filers for the request block to
// the store. The EVM end block logic doesn't update the validator set, thus it returns
// an empty slice.
func (k Keeper) EndBlock(ctx sdk.Context, req abci.RequestEndBlock) []abci.ValidatorUpdate {
	// Gas costs are handled within msg handler so costs should be ignored
	ctx = ctx.WithGasMeter(sdk.NewInfiniteGasMeter())

	// Update account balances before committing other parts of state
	k.UpdateAccounts(ctx)

	// Commit state objects to KV store
	if _, err := k.Commit(ctx, true); err != nil {
		k.Logger(ctx).Error("failed to commit state objects", "error", err, "height", ctx.BlockHeight())
		panic(err)
	}

	// Clear accounts cache after account data has been committed
	k.ClearStateObjects(ctx)

	// set the block bloom filter bytes to store
	bloom := ethtypes.BytesToBloom(k.Bloom.Bytes())
	k.SetBlockBloom(ctx, req.Height, bloom)

	return []abci.ValidatorUpdate{}
}
