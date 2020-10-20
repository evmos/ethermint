package keeper

import (
	"math/big"

	abci "github.com/tendermint/tendermint/abci/types"

	sdk "github.com/cosmos/cosmos-sdk/types"

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

	k.SetBlockHash(ctx, req.Header.LastBlockId.GetHash(), req.Header.GetHeight()-1)

	// reset counters that are used on CommitStateDB.Prepare
	k.Bloom = big.NewInt(0)
	k.TxCount = 0
}

// EndBlock updates the accounts and commits state objects to the KV Store, while
// deleting the empty ones. It also sets the bloom filers for the request block to
// the store. The EVM end block loginc doesn't update the validator set, thus it returns
// an empty slice.
func (k Keeper) EndBlock(ctx sdk.Context, req abci.RequestEndBlock) []abci.ValidatorUpdate {
	// Gas costs are handled within msg handler so costs should be ignored
	ctx = ctx.WithGasMeter(sdk.NewInfiniteGasMeter())

	// Update account balances before committing other parts of state
	k.UpdateAccounts(ctx)

	// Commit state objects to KV store
	_, err := k.Commit(ctx, true)
	if err != nil {
		panic(err)
	}

	// Clear accounts cache after account data has been committed
	k.ClearStateObjects(ctx)

	// set the block bloom filter bytes to store
	bloom := ethtypes.BytesToBloom(k.Bloom.Bytes())
	k.SetBlockBloom(ctx, req.Height, bloom)

	return []abci.ValidatorUpdate{}
}
