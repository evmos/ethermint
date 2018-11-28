package app

import (
	"encoding/hex"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/tendermint/tendermint/crypto/secp256k1"

	evmtypes "github.com/cosmos/ethermint/x/evm/types"
)

var dummySecp256k1Pubkey secp256k1.PubKeySecp256k1 // used for tx simulation

const (
	memoCostPerByte     = 1
	maxMemoCharacters   = 100
	secp256k1VerifyCost = 100
)

func init() {
	bz, _ := hex.DecodeString("035AD6810A47F073553FF30D2FCC7E0D3B1C0B74B61A1AAA2582344037151E143A")
	copy(dummySecp256k1Pubkey[:], bz)
}

// NewAnteHandler returns an ante handelr responsible for attempting to route an
// Ethereum or SDK transaction to an internal ante handler for performing
// transaction-level processing (e.g. fee payment, signature verification) before
// being passed onto it's respective handler.
//
// NOTE: The EVM will already consume (intrinsic) gas for signature verification
// and covering input size as well as handling nonce incrementing.
func NewAnteHandler(ak auth.AccountKeeper, fck auth.FeeCollectionKeeper) sdk.AnteHandler {
	return func(
		ctx sdk.Context, tx sdk.Tx, sim bool,
	) (newCtx sdk.Context, res sdk.Result, abort bool) {

		stdTx, ok := tx.(auth.StdTx)
		if !ok {
			return ctx, sdk.ErrInternal("transaction type invalid: must be StdTx").Result(), true
		}

		// TODO: Handle gas/fee checking and spam prevention. We may need two
		// different models for SDK and Ethereum txs. The SDK currently supports a
		// primitive model where a constant gas price is used.
		//
		// Ref: #473

		if ethTx, ok := isEthereumTx(stdTx); ethTx != nil && ok {
			return ethAnteHandler(ctx, ethTx, ak)
		}

		return auth.NewAnteHandler(ak, fck)(ctx, stdTx, sim)
	}
}

// ----------------------------------------------------------------------------
// Ethereum Ante Handler

// ethAnteHandler defines an internal ante handler for an Ethereum transaction
// ethTx that implements the sdk.Msg interface. The Ethereum transaction is a
// single message inside a auth.StdTx.
//
// For now we simply pass the transaction on as the EVM shares common business
// logic of an ante handler. Anything not handled by the EVM that should be
// prior to transaction processing, should be done here.
func ethAnteHandler(
	ctx sdk.Context, ethTx *evmtypes.MsgEthereumTx, ak auth.AccountKeeper,
) (newCtx sdk.Context, res sdk.Result, abort bool) {

	return ctx, sdk.Result{}, false
}

// ----------------------------------------------------------------------------
// Auxiliary

// isEthereumTx returns a boolean if a given standard SDK transaction contains
// an Ethereum transaction. If so, the transaction is also returned. A standard
// SDK transaction contains an Ethereum transaction if it only has a single
// message and that embedded message if of type MsgEthereumTx.
func isEthereumTx(tx auth.StdTx) (*evmtypes.MsgEthereumTx, bool) {
	msgs := tx.GetMsgs()
	if len(msgs) == 1 {
		ethTx, ok := msgs[0].(*evmtypes.MsgEthereumTx)
		if ok {
			return ethTx, true
		}
	}

	return nil, false
}
