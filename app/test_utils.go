package app

import (
	"fmt"
	"math/big"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"

	"github.com/cosmos/ethermint/crypto"
	evmtypes "github.com/cosmos/ethermint/x/evm/types"

	abci "github.com/tendermint/tendermint/abci/types"
	tmcrypto "github.com/tendermint/tendermint/crypto"
	tmsecp256k1 "github.com/tendermint/tendermint/crypto/secp256k1"
	"github.com/tendermint/tendermint/libs/log"
)

var testDenom = "testcoin"

type testSetup struct {
	ctx         sdk.Context
	accKeeper   auth.AccountKeeper
	feeKeeper   auth.FeeCollectionKeeper
	anteHandler sdk.AnteHandler
}

func newTestSetup() testSetup {
	cdc := codec.New()
	ms, capKey, capKey2 := setupMultiStore()

	auth.RegisterBaseAccount(cdc)

	accKeeper := auth.NewAccountKeeper(cdc, capKey, auth.ProtoBaseAccount)
	feeKeeper := auth.NewFeeCollectionKeeper(cdc, capKey2)
	anteHandler := NewAnteHandler(accKeeper, feeKeeper)

	ctx := sdk.NewContext(
		ms,
		abci.Header{ChainID: "3", Time: time.Now().UTC()},
		false,
		log.NewNopLogger(),
	)

	return testSetup{
		ctx:         ctx,
		accKeeper:   accKeeper,
		feeKeeper:   feeKeeper,
		anteHandler: anteHandler,
	}
}

func newTestMsg(addrs ...sdk.AccAddress) *sdk.TestMsg {
	return sdk.NewTestMsg(addrs...)
}

func newTestCoins() sdk.Coins {
	return sdk.Coins{sdk.NewInt64Coin(testDenom, 10000000)}
}

func newTestStdFee() auth.StdFee {
	return auth.NewStdFee(5000, sdk.NewInt64Coin(testDenom, 150))
}

// GenerateAddress generates an Ethereum address.
func newTestAddrKey() (sdk.AccAddress, tmcrypto.PrivKey) {
	priv := tmsecp256k1.GenPrivKey()
	addr := sdk.AccAddress(priv.PubKey().Address())
	return addr, priv
}

func newTestSDKTx(
	ctx sdk.Context, msgs []sdk.Msg, privs []tmcrypto.PrivKey,
	accNums []int64, seqs []int64, fee auth.StdFee,
) sdk.Tx {

	sigs := make([]auth.StdSignature, len(privs))
	for i, priv := range privs {
		signBytes := auth.StdSignBytes(ctx.ChainID(), accNums[i], seqs[i], fee, msgs, "")
		sig, err := priv.Sign(signBytes)
		if err != nil {
			panic(err)
		}

		sigs[i] = auth.StdSignature{
			PubKey:        priv.PubKey(),
			Signature:     sig,
			AccountNumber: accNums[i],
			Sequence:      seqs[i],
		}
	}

	return auth.NewStdTx(msgs, fee, sigs, "")
}

func newTestEthTx(ctx sdk.Context, msg *evmtypes.MsgEthereumTx, priv tmcrypto.PrivKey) sdk.Tx {
	chainID, ok := new(big.Int).SetString(ctx.ChainID(), 10)
	if !ok {
		panic(fmt.Sprintf("invalid chainID: %s", ctx.ChainID()))
	}

	privKey, err := crypto.PrivKeyToSecp256k1(priv)
	if err != nil {
		panic(fmt.Sprintf("failed to convert private key: %s", err))
	}

	msg.Sign(chainID, privKey)
	return auth.NewStdTx([]sdk.Msg{msg}, auth.StdFee{}, nil, "")
}
