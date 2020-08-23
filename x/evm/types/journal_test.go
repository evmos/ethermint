package types

import (
	"os"
	"testing"

	"github.com/stretchr/testify/suite"

	abci "github.com/tendermint/tendermint/abci/types"
	tmlog "github.com/tendermint/tendermint/libs/log"
	tmdb "github.com/tendermint/tm-db"

	sdkcodec "github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/params"

	ethcmn "github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"

	"github.com/cosmos/ethermint/crypto"
	ethermint "github.com/cosmos/ethermint/types"
)

type JournalTestSuite struct {
	suite.Suite

	address ethcmn.Address
	journal *journal
	ctx     sdk.Context
	stateDB *CommitStateDB
}

func newTestCodec() *sdkcodec.Codec {
	cdc := sdkcodec.New()

	RegisterCodec(cdc)
	sdk.RegisterCodec(cdc)
	crypto.RegisterCodec(cdc)
	sdkcodec.RegisterCrypto(cdc)
	auth.RegisterCodec(cdc)
	ethermint.RegisterCodec(cdc)

	return cdc
}

func (suite *JournalTestSuite) SetupTest() {
	suite.setup()

	privkey, err := crypto.GenerateKey()
	suite.Require().NoError(err)

	suite.address = ethcmn.BytesToAddress(privkey.PubKey().Address().Bytes())
	suite.journal = newJournal()

	balance := sdk.NewCoins(sdk.NewCoin(ethermint.DenomDefault, sdk.NewInt(100)))
	acc := &ethermint.EthAccount{
		BaseAccount: auth.NewBaseAccount(sdk.AccAddress(suite.address.Bytes()), balance, nil, 0, 0),
		CodeHash:    ethcrypto.Keccak256(nil),
	}

	suite.stateDB.accountKeeper.SetAccount(suite.ctx, acc)
	// suite.stateDB.bankKeeper.SetBalance(suite.ctx, sdk.AccAddress(suite.address.Bytes()), balance)
	suite.stateDB.SetLogs(ethcmn.BytesToHash([]byte("txhash")), []*ethtypes.Log{
		{
			Address:     suite.address,
			Topics:      []ethcmn.Hash{ethcmn.BytesToHash([]byte("topic_0"))},
			Data:        []byte("data_0"),
			BlockNumber: 1,
			TxHash:      ethcmn.BytesToHash([]byte("tx_hash")),
			TxIndex:     1,
			BlockHash:   ethcmn.BytesToHash([]byte("block_hash")),
			Index:       1,
			Removed:     false,
		},
		{
			Address:     suite.address,
			Topics:      []ethcmn.Hash{ethcmn.BytesToHash([]byte("topic_1"))},
			Data:        []byte("data_1"),
			BlockNumber: 10,
			TxHash:      ethcmn.BytesToHash([]byte("tx_hash")),
			TxIndex:     0,
			BlockHash:   ethcmn.BytesToHash([]byte("block_hash")),
			Index:       0,
			Removed:     false,
		},
	})
}

// setup performs a manual setup of the GoLevelDB and mounts the required IAVL stores. We use the manual
// setup here instead of the Ethermint app test setup because the journal methods are private and using
// the latter would result in a cycle dependency. We also want to avoid declaring the journal methods public
// to maintain consistency with the Geth implementation.
func (suite *JournalTestSuite) setup() {
	authKey := sdk.NewKVStoreKey(auth.StoreKey)
	// bankKey := sdk.NewKVStoreKey(bank.StoreKey)
	storeKey := sdk.NewKVStoreKey(StoreKey)

	db := tmdb.NewDB("state", tmdb.GoLevelDBBackend, "temp")
	defer func() {
		os.RemoveAll("temp")
	}()

	cms := store.NewCommitMultiStore(db)
	cms.MountStoreWithDB(authKey, sdk.StoreTypeIAVL, db)
	// cms.MountStoreWithDB(bankKey, sdk.StoreTypeIAVL, db)
	cms.MountStoreWithDB(storeKey, sdk.StoreTypeIAVL, db)

	err := cms.LoadLatestVersion()
	suite.Require().NoError(err)

	cdc := newTestCodec()

	keyParams := sdk.NewKVStoreKey(params.StoreKey)
	tkeyParams := sdk.NewTransientStoreKey(params.TStoreKey)
	paramsKeeper := params.NewKeeper(cdc, keyParams, tkeyParams)

	authSubspace := paramsKeeper.Subspace(auth.DefaultParamspace)

	ak := auth.NewAccountKeeper(cdc, authKey, authSubspace, ethermint.ProtoAccount)

	suite.ctx = sdk.NewContext(cms, abci.Header{ChainID: "8"}, false, tmlog.NewNopLogger())
	suite.stateDB = NewCommitStateDB(suite.ctx, storeKey, ak).WithContext(suite.ctx)
}

func TestJournalTestSuite(t *testing.T) {
	suite.Run(t, new(JournalTestSuite))
}

func (suite *JournalTestSuite) TestJournal_append_revert() {
	testCases := []struct {
		name  string
		entry journalEntry
	}{
		{
			"createObjectChange",
			createObjectChange{
				account: &suite.address,
			},
		},
		{
			"resetObjectChange",
			resetObjectChange{
				prev: &stateObject{
					address: suite.address,
				},
			},
		},
		{
			"suicideChange",
			suicideChange{
				account:     &suite.address,
				prev:        false,
				prevBalance: sdk.OneInt(),
			},
		},
		{
			"balanceChange",
			balanceChange{
				account: &suite.address,
				prev:    sdk.OneInt(),
			},
		},
		{
			"nonceChange",
			nonceChange{
				account: &suite.address,
				prev:    1,
			},
		},
		{
			"storageChange",
			storageChange{
				account:   &suite.address,
				key:       ethcmn.BytesToHash([]byte("key")),
				prevValue: ethcmn.BytesToHash([]byte("value")),
			},
		},
		{
			"codeChange",
			codeChange{
				account:  &suite.address,
				prevCode: []byte("code"),
				prevHash: []byte("hash"),
			},
		},
		{
			"touchChange",
			touchChange{
				account: &suite.address,
			},
		},
		{
			"refundChange",
			refundChange{
				prev: 1,
			},
		},
		{
			"addPreimageChange",
			addPreimageChange{
				hash: ethcmn.BytesToHash([]byte("hash")),
			},
		},
		{
			"addLogChange",
			addLogChange{
				txhash: ethcmn.BytesToHash([]byte("hash")),
			},
		},
		{
			"addLogChange - 2 logs",
			addLogChange{
				txhash: ethcmn.BytesToHash([]byte("txhash")),
			},
		},
	}
	var dirtyCount int
	for i, tc := range testCases {
		suite.journal.append(tc.entry)
		suite.Require().Equal(suite.journal.length(), i+1, tc.name)
		if tc.entry.dirtied() != nil {
			dirtyCount++

			suite.Require().Equal(dirtyCount, suite.journal.getDirty(suite.address), tc.name)
		}
	}

	// revert to the initial journal state
	suite.journal.revert(suite.stateDB, 0)

	// verify the dirty entry has been deleted
	idx, ok := suite.journal.addressToJournalIndex[suite.address]
	suite.Require().False(ok)
	suite.Require().Zero(idx)
}

func (suite *JournalTestSuite) TestJournal_dirty() {
	// dirty entry hasn't been set
	idx, ok := suite.journal.addressToJournalIndex[suite.address]
	suite.Require().False(ok)
	suite.Require().Zero(idx)

	// update dirty count
	suite.journal.dirty(suite.address)
	suite.Require().Equal(1, suite.journal.getDirty(suite.address))
}
