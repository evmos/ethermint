package app

import (
	"fmt"
	"math/big"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/mock"
	"github.com/cosmos/cosmos-sdk/x/params"

	"github.com/cosmos/ethermint/crypto"
	emint "github.com/cosmos/ethermint/types"
	evmtypes "github.com/cosmos/ethermint/x/evm/types"

	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	abci "github.com/tendermint/tendermint/abci/types"
	tmcrypto "github.com/tendermint/tendermint/crypto"
	cmn "github.com/tendermint/tendermint/libs/common"
	"github.com/tendermint/tendermint/libs/log"
	dbm "github.com/tendermint/tm-db"
)

type testSetup struct {
	ctx          sdk.Context
	cdc          *codec.Codec
	accKeeper    auth.AccountKeeper
	supplyKeeper types.SupplyKeeper
	anteHandler  sdk.AnteHandler
}

func newTestSetup() testSetup {
	db := dbm.NewMemDB()
	authCapKey := sdk.NewKVStoreKey("authCapKey")
	keySupply := sdk.NewKVStoreKey("keySupply")
	keyParams := sdk.NewKVStoreKey("params")
	tkeyParams := sdk.NewTransientStoreKey("transient_params")

	ms := store.NewCommitMultiStore(db)
	ms.MountStoreWithDB(authCapKey, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(keySupply, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(keyParams, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(tkeyParams, sdk.StoreTypeIAVL, db)

	if err := ms.LoadLatestVersion(); err != nil {
		cmn.Exit(err.Error())
	}

	cdc := MakeCodec()
	cdc.RegisterConcrete(&sdk.TestMsg{}, "test/TestMsg", nil)

	// Set params keeper and subspaces
	paramsKeeper := params.NewKeeper(cdc, keyParams, tkeyParams, params.DefaultCodespace)
	authSubspace := paramsKeeper.Subspace(auth.DefaultParamspace)

	ctx := sdk.NewContext(
		ms,
		abci.Header{ChainID: "3", Time: time.Now().UTC()},
		true,
		log.NewNopLogger(),
	)

	// Add keepers
	accKeeper := auth.NewAccountKeeper(cdc, authCapKey, authSubspace, auth.ProtoBaseAccount)
	accKeeper.SetParams(ctx, types.DefaultParams())
	supplyKeeper := mock.NewDummySupplyKeeper(accKeeper)
	anteHandler := NewAnteHandler(accKeeper, supplyKeeper)

	return testSetup{
		ctx:          ctx,
		cdc:          cdc,
		accKeeper:    accKeeper,
		supplyKeeper: supplyKeeper,
		anteHandler:  anteHandler,
	}
}

func newTestMsg(addrs ...sdk.AccAddress) *sdk.TestMsg {
	return sdk.NewTestMsg(addrs...)
}

func newTestCoins() sdk.Coins {
	return sdk.Coins{sdk.NewInt64Coin(emint.DenomDefault, 500000000)}
}

func newTestStdFee() auth.StdFee {
	return auth.NewStdFee(220000, sdk.NewCoins(sdk.NewInt64Coin(emint.DenomDefault, 150)))
}

// GenerateAddress generates an Ethereum address.
func newTestAddrKey() (sdk.AccAddress, tmcrypto.PrivKey) {
	privkey, _ := crypto.GenerateKey()
	addr := ethcrypto.PubkeyToAddress(privkey.ToECDSA().PublicKey)

	return sdk.AccAddress(addr.Bytes()), privkey
}

func newTestSDKTx(
	ctx sdk.Context, msgs []sdk.Msg, privs []tmcrypto.PrivKey,
	accNums []uint64, seqs []uint64, fee auth.StdFee,
) sdk.Tx {

	sigs := make([]auth.StdSignature, len(privs))
	for i, priv := range privs {
		signBytes := auth.StdSignBytes(ctx.ChainID(), accNums[i], seqs[i], fee, msgs, "")

		sig, err := priv.Sign(signBytes)
		if err != nil {
			panic(err)
		}

		sigs[i] = auth.StdSignature{
			PubKey:    priv.PubKey(),
			Signature: sig,
		}
	}

	return auth.NewStdTx(msgs, fee, sigs, "")
}

func newTestEthTx(ctx sdk.Context, msg *evmtypes.EthereumTxMsg, priv tmcrypto.PrivKey) sdk.Tx {
	chainID, ok := new(big.Int).SetString(ctx.ChainID(), 10)
	if !ok {
		panic(fmt.Sprintf("invalid chainID: %s", ctx.ChainID()))
	}

	privkey, ok := priv.(crypto.PrivKeySecp256k1)
	if !ok {
		panic(fmt.Sprintf("invalid private key type: %T", priv))
	}

	msg.Sign(chainID, privkey.ToECDSA())
	return msg
}
