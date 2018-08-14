package handlers

import (
	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
	dbm "github.com/tendermint/tendermint/libs/db"
)

func SetupMultiStore() (sdk.MultiStore, *sdk.KVStoreKey) {
	db := dbm.NewMemDB()
	capKey := sdk.NewKVStoreKey("capkey")
	ms := store.NewCommitMultiStore(db)
	ms.MountStoreWithDB(capKey, sdk.StoreTypeIAVL, db)
	ms.LoadLatestVersion()
	return ms, capKey
}
