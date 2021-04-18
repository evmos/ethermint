package keeper

import (
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	abci "github.com/tendermint/tendermint/abci/types"
)

// BeginBlock sets the block hash -> block height map for the previous block height
// and resets the Bloom filter and the transaction count to 0.
func (k *Keeper) BeginBlock(ctx sdk.Context, req abci.RequestBeginBlock) {
	if req.Header.Height < 1 {
		return
	}

	// Gas costs are handled within msg handler so costs should be ignored
	ctx = ctx.WithGasMeter(sdk.NewInfiniteGasMeter())

	k.SetBlockHash(ctx, req.Hash, req.Header.Height)
	k.SetBlockHeightToHash(ctx, req.Hash, req.Header.Height)

	// special setter for csdb
	k.SetHeightHash(ctx, uint64(req.Header.Height), common.BytesToHash(req.Hash))

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

	root, err := k.Commit(ctx, true)
	// Commit state objects to KV store
	if err != nil {
		k.Logger(ctx).Error("failed to commit state objects", "error", err, "height", ctx.BlockHeight())
		panic(err)
	}

	// reset all cache after account data has been committed, that make sure node state consistent
	if err = k.Reset(ctx, root); err != nil {
		panic(err)
	}

	// set the block bloom filter bytes to store
	bloom := ethtypes.BytesToBloom(k.Bloom.Bytes())
	k.SetBlockBloom(ctx, req.Height, bloom)

	return []abci.ValidatorUpdate{}
}
