package evm

import (
	"math/big"
	"testing"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/params"

	"github.com/cosmos/ethermint/types"
	evmtypes "github.com/cosmos/ethermint/x/evm/types"

	ethcmn "github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"

	abci "github.com/tendermint/tendermint/abci/types"
	tmlog "github.com/tendermint/tendermint/libs/log"
	dbm "github.com/tendermint/tm-db"
)

var (
	address = ethcmn.HexToAddress("0x756F45E3FA69347A9A973A725E3C98bC4db0b4c1")

	accKey     = sdk.NewKVStoreKey("acc")
	storageKey = sdk.NewKVStoreKey(evmtypes.EvmStoreKey)
	codeKey    = sdk.NewKVStoreKey(evmtypes.EvmCodeKey)

	logger = tmlog.NewNopLogger()
)

func newTestCodec() *codec.Codec {
	cdc := codec.New()

	evmtypes.RegisterCodec(cdc)
	types.RegisterCodec(cdc)
	auth.RegisterCodec(cdc)
	sdk.RegisterCodec(cdc)
	codec.RegisterCrypto(cdc)

	return cdc
}

func TestDBStorage(t *testing.T) {
	// create logger, codec and root multi-store
	cdc := newTestCodec()

	// The ParamsKeeper handles parameter storage for the application
	keyParams := sdk.NewKVStoreKey(params.StoreKey)
	tkeyParams := sdk.NewTransientStoreKey(params.TStoreKey)
	paramsKeeper := params.NewKeeper(cdc, keyParams, tkeyParams, params.DefaultCodespace)
	// Set specific supspaces
	authSubspace := paramsKeeper.Subspace(auth.DefaultParamspace)
	ak := auth.NewAccountKeeper(cdc, accKey, authSubspace, types.ProtoBaseAccount)
	ek := NewKeeper(ak, storageKey, codeKey, cdc)

	db := dbm.NewMemDB()
	cms := store.NewCommitMultiStore(db)
	// mount stores
	keys := []*sdk.KVStoreKey{accKey, storageKey, codeKey}
	for _, key := range keys {
		cms.MountStoreWithDB(key, sdk.StoreTypeIAVL, nil)
	}

	// load latest version (root)
	err := cms.LoadLatestVersion()
	require.NoError(t, err)

	// First execution
	ms := cms.CacheMultiStore()
	ctx := sdk.NewContext(ms, abci.Header{}, false, logger)
	ctx = ctx.WithBlockHeight(1)

	// Perform state transitions
	ek.SetBalance(ctx, address, big.NewInt(5))
	ek.SetNonce(ctx, address, 4)
	ek.SetState(ctx, address, ethcmn.HexToHash("0x2"), ethcmn.HexToHash("0x3"))
	ek.SetCode(ctx, address, []byte{0x1})

	// Get those state transitions
	require.Equal(t, ek.GetBalance(ctx, address).Cmp(big.NewInt(5)), 0)
	require.Equal(t, ek.GetNonce(ctx, address), uint64(4))
	require.Equal(t, ek.GetState(ctx, address, ethcmn.HexToHash("0x2")), ethcmn.HexToHash("0x3"))
	require.Equal(t, ek.GetCode(ctx, address), []byte{0x1})

	// commit stateDB
	_, err = ek.Commit(ctx, false)
	require.NoError(t, err, "failed to commit StateDB")

	// simulate BaseApp EndBlocker commitment
	ms.Write()
	cms.Commit()
}
