package app

import (
	"fmt"
	"math/big"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"

	"github.com/cosmos/ethermint/crypto"
	"github.com/cosmos/ethermint/types"
	evmtypes "github.com/cosmos/ethermint/x/evm/types"

	ethcrypto "github.com/ethereum/go-ethereum/crypto"

	abci "github.com/tendermint/tendermint/abci/types"
	tmcrypto "github.com/tendermint/tendermint/crypto"
	dbm "github.com/tendermint/tendermint/libs/db"
	"github.com/tendermint/tendermint/libs/log"
)

type testSetup struct {
	ctx         sdk.Context
	cdc         *codec.Codec
	accKeeper   auth.AccountKeeper
	feeKeeper   auth.FeeCollectionKeeper
	anteHandler sdk.AnteHandler
}

func newTestSetup() testSetup {
	db := dbm.NewMemDB()
	authCapKey := sdk.NewKVStoreKey("authCapKey")
	feeCapKey := sdk.NewKVStoreKey("feeCapKey")
	keyParams := sdk.NewKVStoreKey("params")
	tkeyParams := sdk.NewTransientStoreKey("transient_params")

	ms := store.NewCommitMultiStore(db)
	ms.MountStoreWithDB(authCapKey, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(feeCapKey, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(keyParams, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(tkeyParams, sdk.StoreTypeIAVL, db)
	ms.LoadLatestVersion()

	cdc := CreateCodec()
	cdc.RegisterConcrete(&sdk.TestMsg{}, "test/TestMsg", nil)

	accKeeper := auth.NewAccountKeeper(cdc, authCapKey, auth.ProtoBaseAccount)
	feeKeeper := auth.NewFeeCollectionKeeper(cdc, feeCapKey)
	anteHandler := NewAnteHandler(accKeeper, feeKeeper)

	ctx := sdk.NewContext(
		ms,
		abci.Header{ChainID: "3", Time: time.Now().UTC()},
		true,
		log.NewNopLogger(),
	)

	return testSetup{
		ctx:         ctx,
		cdc:         cdc,
		accKeeper:   accKeeper,
		feeKeeper:   feeKeeper,
		anteHandler: anteHandler,
	}
}

func newTestMsg(addrs ...sdk.AccAddress) *sdk.TestMsg {
	return sdk.NewTestMsg(addrs...)
}

func newTestCoins() sdk.Coins {
	return sdk.Coins{sdk.NewInt64Coin(types.DenomDefault, 500000000)}
}

func newTestStdFee() auth.StdFee {
	return auth.NewStdFee(220000, sdk.NewInt64Coin(types.DenomDefault, 150))
}

// GenerateAddress generates an Ethereum address.
func newTestAddrKey() (sdk.AccAddress, tmcrypto.PrivKey) {
	privkey, _ := crypto.GenerateKey()
	addr := ethcrypto.PubkeyToAddress(privkey.PublicKey)

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
