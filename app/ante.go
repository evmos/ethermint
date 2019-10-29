package app

import (
	"fmt"
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/exported"
	"github.com/cosmos/cosmos-sdk/x/auth/types"

	"github.com/cosmos/ethermint/crypto"
	emint "github.com/cosmos/ethermint/types"
	evmtypes "github.com/cosmos/ethermint/x/evm/types"

	ethcore "github.com/ethereum/go-ethereum/core"

	tmcrypto "github.com/tendermint/tendermint/crypto"
)

const (
	memoCostPerByte     sdk.Gas = 3
	secp256k1VerifyCost uint64  = 21000
)

// NewAnteHandler returns an ante handler responsible for attempting to route an
// Ethereum or SDK transaction to an internal ante handler for performing
// transaction-level processing (e.g. fee payment, signature verification) before
// being passed onto it's respective handler.
//
// NOTE: The EVM will already consume (intrinsic) gas for signature verification
// and covering input size as well as handling nonce incrementing.
func NewAnteHandler(ak auth.AccountKeeper, sk types.SupplyKeeper) sdk.AnteHandler {
	return func(
		ctx sdk.Context, tx sdk.Tx, sim bool,
	) (newCtx sdk.Context, res sdk.Result, abort bool) {

		switch castTx := tx.(type) {
		case auth.StdTx:
			return sdkAnteHandler(ctx, ak, sk, castTx, sim)

		case *evmtypes.EthereumTxMsg:
			return ethAnteHandler(ctx, ak, sk, castTx, sim)

		default:
			return ctx, sdk.ErrInternal(fmt.Sprintf("transaction type invalid: %T", tx)).Result(), true
		}
	}
}

// ----------------------------------------------------------------------------
// SDK Ante Handler

func sdkAnteHandler(
	ctx sdk.Context, ak auth.AccountKeeper, sk types.SupplyKeeper, stdTx auth.StdTx, sim bool,
) (newCtx sdk.Context, res sdk.Result, abort bool) {
	// Ensure that the provided fees meet a minimum threshold for the validator,
	// if this is a CheckTx. This is only for local mempool purposes, and thus
	// is only ran on check tx.
	if ctx.IsCheckTx() && !sim {
		res := auth.EnsureSufficientMempoolFees(ctx, stdTx.Fee)
		if !res.IsOK() {
			return newCtx, res, true
		}
	}

	newCtx = auth.SetGasMeter(sim, ctx, stdTx.Fee.Gas)

	// AnteHandlers must have their own defer/recover in order for the BaseApp
	// to know how much gas was used! This is because the GasMeter is created in
	// the AnteHandler, but if it panics the context won't be set properly in
	// runTx's recover call.
	defer func() {
		if r := recover(); r != nil {
			switch rType := r.(type) {
			case sdk.ErrorOutOfGas:
				log := fmt.Sprintf("out of gas in location: %v", rType.Descriptor)
				res = sdk.ErrOutOfGas(log).Result()
				res.GasWanted = stdTx.Fee.Gas
				res.GasUsed = newCtx.GasMeter().GasConsumed()
				abort = true
			default:
				panic(r)
			}
		}
	}()

	if err := stdTx.ValidateBasic(); err != nil {
		return newCtx, err.Result(), true
	}

	newCtx.GasMeter().ConsumeGas(memoCostPerByte*sdk.Gas(len(stdTx.GetMemo())), "memo")

	// stdSigs contains the sequence number, account number, and signatures.
	// When simulating, this would just be a 0-length slice.
	signerAddrs := stdTx.GetSigners()
	signerAccs := make([]exported.Account, len(signerAddrs))
	isGenesis := ctx.BlockHeight() == 0

	// fetch first signer, who's going to pay the fees
	signerAccs[0], res = auth.GetSignerAcc(newCtx, ak, signerAddrs[0])
	if !res.IsOK() {
		return newCtx, res, true
	}

	// the first signer pays the transaction fees
	if !stdTx.Fee.Amount.IsZero() {
		res = auth.DeductFees(sk, newCtx, signerAccs[0], stdTx.Fee.Amount)
		if !res.IsOK() {
			return newCtx, res, true
		}

		// Reload account after fees deducted
		signerAccs[0] = ak.GetAccount(newCtx, signerAccs[0].GetAddress())
	}

	stdSigs := stdTx.GetSignatures()

	for i := 0; i < len(stdSigs); i++ {
		// skip the fee payer, account is cached and fees were deducted already
		if i != 0 {
			signerAccs[i], res = auth.GetSignerAcc(newCtx, ak, signerAddrs[i])
			if !res.IsOK() {
				return newCtx, res, true
			}
		}

		// check signature, return account with incremented nonce
		signBytes := auth.GetSignBytes(newCtx.ChainID(), stdTx, signerAccs[i], isGenesis)
		signerAccs[i], res = processSig(newCtx, signerAccs[i], stdSigs[i], signBytes, sim)
		if !res.IsOK() {
			return newCtx, res, true
		}

		ak.SetAccount(newCtx, signerAccs[i])
	}

	return newCtx, sdk.Result{GasWanted: stdTx.Fee.Gas}, false
}

// processSig verifies the signature and increments the nonce. If the account
// doesn't have a pubkey, set it.
func processSig(
	ctx sdk.Context, acc auth.Account, sig auth.StdSignature, signBytes []byte, sim bool,
) (updatedAcc auth.Account, res sdk.Result) {

	pubKey, res := auth.ProcessPubKey(acc, sig, sim)
	if !res.IsOK() {
		return nil, res
	}

	err := acc.SetPubKey(pubKey)
	if err != nil {
		return nil, sdk.ErrInternal("failed to set PubKey on signer account").Result()
	}

	consumeSigGas(ctx.GasMeter(), pubKey)
	if !sim && !pubKey.VerifyBytes(signBytes, sig.Signature) {
		return nil, sdk.ErrUnauthorized("signature verification failed").Result()
	}

	err = acc.SetSequence(acc.GetSequence() + 1)
	if err != nil {
		return nil, sdk.ErrInternal("failed to set account nonce").Result()
	}

	return acc, res
}

func consumeSigGas(meter sdk.GasMeter, pubkey tmcrypto.PubKey) {
	switch pubkey.(type) {
	case crypto.PubKeySecp256k1:
		meter.ConsumeGas(secp256k1VerifyCost, "ante verify: secp256k1")
	// TODO: Remove allowing tm Pub key to sign transactions (if intended in final release)
	// or until genesis utils are built into the evm or as their own module
	case tmcrypto.PubKey:
		meter.ConsumeGas(secp256k1VerifyCost, "ante verify: tendermint secp256k1")
	default:
		panic("Unrecognized signature type")
	}
}

// ----------------------------------------------------------------------------
// Ethereum Ante Handler

// ethAnteHandler defines an internal ante handler for an Ethereum transaction
// ethTxMsg. During CheckTx, the transaction is passed through a series of
// pre-message execution validation checks such as signature and account
// verification in addition to minimum fees being checked. Otherwise, during
// DeliverTx, the transaction is simply passed to the EVM which will also
// perform the same series of checks. The distinction is made in CheckTx to
// prevent spam and DoS attacks.
func ethAnteHandler(
	ctx sdk.Context, ak auth.AccountKeeper, sk types.SupplyKeeper,
	ethTxMsg *evmtypes.EthereumTxMsg, sim bool,
) (newCtx sdk.Context, res sdk.Result, abort bool) {

	var senderAddr sdk.AccAddress

	// This is done to ignore costs in Ante handler checks
	ctx = ctx.WithBlockGasMeter(sdk.NewInfiniteGasMeter())

	if ctx.IsCheckTx() {
		// Only perform pre-message (Ethereum transaction) execution validation
		// during CheckTx. Otherwise, during DeliverTx the EVM will handle them.
		if senderAddr, res = validateEthTxCheckTx(ctx, ak, ethTxMsg); !res.IsOK() {
			return ctx, res, true
		}
	} else {
		// This is still currently needed to retrieve the sender address
		if senderAddr, res = validateSignature(ctx, ethTxMsg); !res.IsOK() {
			return ctx, res, true
		}

		// Explicit nonce check is also needed in case of multiple txs with same nonce not being handled
		if res := checkNonce(ctx, ak, ethTxMsg, senderAddr); !res.IsOK() {
			return ctx, res, true
		}
	}

	// Recover and catch out of gas error
	defer func() {
		if r := recover(); r != nil {
			switch rType := r.(type) {
			case sdk.ErrorOutOfGas:
				log := fmt.Sprintf("out of gas in location: %v", rType.Descriptor)
				res = sdk.ErrOutOfGas(log).Result()
				res.GasWanted = ethTxMsg.Data.GasLimit
				res.GasUsed = ctx.GasMeter().GasConsumed()
				abort = true
			default:
				panic(r)
			}
		}
	}()

	// Fetch sender account from signature
	senderAcc, res := auth.GetSignerAcc(ctx, ak, senderAddr)
	if !res.IsOK() {
		return ctx, res, true
	}

	// Charge sender for gas up to limit
	if ethTxMsg.Data.GasLimit != 0 {
		// Cost calculates the fees paid to validators based on gas limit and price
		cost := new(big.Int).Mul(ethTxMsg.Data.Price, new(big.Int).SetUint64(ethTxMsg.Data.GasLimit))

		res = auth.DeductFees(sk, ctx, senderAcc, sdk.Coins{
			sdk.NewCoin(emint.DenomDefault, sdk.NewIntFromBigInt(cost)),
		})

		if !res.IsOK() {
			return ctx, res, true
		}
	}

	// Set gas meter after ante handler to ignore gaskv costs
	newCtx = auth.SetGasMeter(sim, ctx, ethTxMsg.Data.GasLimit)

	gas, _ := ethcore.IntrinsicGas(ethTxMsg.Data.Payload, ethTxMsg.To() == nil, true)
	newCtx.GasMeter().ConsumeGas(gas, "eth intrinsic gas")

	return newCtx, sdk.Result{}, false
}

func validateEthTxCheckTx(
	ctx sdk.Context, ak auth.AccountKeeper, ethTxMsg *evmtypes.EthereumTxMsg,
) (sdk.AccAddress, sdk.Result) {
	// Validate sufficient fees have been provided that meet a minimum threshold
	// defined by the proposer (for mempool purposes during CheckTx).
	if res := ensureSufficientMempoolFees(ctx, ethTxMsg); !res.IsOK() {
		return nil, res
	}

	// validate enough intrinsic gas
	if res := validateIntrinsicGas(ethTxMsg); !res.IsOK() {
		return nil, res
	}

	signer, res := validateSignature(ctx, ethTxMsg)
	if !res.IsOK() {
		return nil, res
	}

	// validate account (nonce and balance checks)
	if res := validateAccount(ctx, ak, ethTxMsg, signer); !res.IsOK() {
		return nil, res
	}

	return sdk.AccAddress(signer.Bytes()), sdk.Result{}
}

// Validates signature and returns sender address
func validateSignature(ctx sdk.Context, ethTxMsg *evmtypes.EthereumTxMsg) (sdk.AccAddress, sdk.Result) {
	// parse the chainID from a string to a base-10 integer
	chainID, ok := new(big.Int).SetString(ctx.ChainID(), 10)
	if !ok {
		return nil, emint.ErrInvalidChainID(fmt.Sprintf("invalid chainID: %s", ctx.ChainID())).Result()
	}

	// validate sender/signature
	signer, err := ethTxMsg.VerifySig(chainID)
	if err != nil {
		return nil, sdk.ErrUnauthorized(fmt.Sprintf("signature verification failed: %s", err)).Result()
	}

	return sdk.AccAddress(signer.Bytes()), sdk.Result{}
}

// validateIntrinsicGas validates that the Ethereum tx message has enough to
// cover intrinsic gas. Intrinsic gas for a transaction is the amount of gas
// that the transaction uses before the transaction is executed. The gas is a
// constant value of 21000 plus any cost inccured by additional bytes of data
// supplied with the transaction.
func validateIntrinsicGas(ethTxMsg *evmtypes.EthereumTxMsg) sdk.Result {
	gas, err := ethcore.IntrinsicGas(ethTxMsg.Data.Payload, ethTxMsg.To() == nil, true)
	if err != nil {
		return sdk.ErrInternal(fmt.Sprintf("failed to compute intrinsic gas cost: %s", err)).Result()
	}

	if ethTxMsg.Data.GasLimit < gas {
		return sdk.ErrInternal(
			fmt.Sprintf("intrinsic gas too low; %d < %d", ethTxMsg.Data.GasLimit, gas),
		).Result()
	}

	return sdk.Result{}
}

// validateAccount validates the account nonce and that the account has enough
// funds to cover the tx cost.
func validateAccount(
	ctx sdk.Context, ak auth.AccountKeeper, ethTxMsg *evmtypes.EthereumTxMsg, signer sdk.AccAddress,
) sdk.Result {

	acc := ak.GetAccount(ctx, signer)

	// on InitChain make sure account number == 0
	if ctx.BlockHeight() == 0 && acc.GetAccountNumber() != 0 {
		return sdk.ErrInternal(
			fmt.Sprintf(
				"invalid account number for height zero; got %d, expected 0", acc.GetAccountNumber(),
			)).Result()
	}

	// Validate nonce is correct
	if res := checkNonce(ctx, ak, ethTxMsg, signer); !res.IsOK() {
		return res
	}

	// validate sender has enough funds
	balance := acc.GetCoins().AmountOf(emint.DenomDefault)
	if balance.BigInt().Cmp(ethTxMsg.Cost()) < 0 {
		return sdk.ErrInsufficientFunds(
			fmt.Sprintf("insufficient funds: %s < %s", balance, ethTxMsg.Cost()),
		).Result()
	}

	return sdk.Result{}
}

func checkNonce(
	ctx sdk.Context, ak auth.AccountKeeper, ethTxMsg *evmtypes.EthereumTxMsg, signer sdk.AccAddress,
) sdk.Result {
	acc := ak.GetAccount(ctx, signer)
	// Validate the transaction nonce is valid (equivalent to the sender account’s
	// current nonce).
	seq := acc.GetSequence()
	if ethTxMsg.Data.AccountNonce != seq {
		return sdk.ErrInvalidSequence(
			fmt.Sprintf("invalid nonce; got %d, expected %d", ethTxMsg.Data.AccountNonce, seq)).Result()
	}

	return sdk.Result{}
}

// ensureSufficientMempoolFees verifies that enough fees have been provided by the
// Ethereum transaction that meet the minimum threshold set by the block
// proposer.
//
// NOTE: This should only be ran during a CheckTx mode.
func ensureSufficientMempoolFees(ctx sdk.Context, ethTxMsg *evmtypes.EthereumTxMsg) sdk.Result {
	// fee = GP * GL
	fee := sdk.NewDecCoinFromCoin(sdk.NewInt64Coin(emint.DenomDefault, ethTxMsg.Fee().Int64()))

	minGasPrices := ctx.MinGasPrices()
	allGTE := true
	for _, v := range minGasPrices {
		if !fee.IsGTE(v) {
			allGTE = false
		}
	}

	// it is assumed that the minimum fees will only include the single valid denom
	if !ctx.MinGasPrices().IsZero() && !allGTE {
		// reject the transaction that does not meet the minimum fee
		return sdk.ErrInsufficientFee(
			fmt.Sprintf("insufficient fee, got: %q required: %q", fee, ctx.MinGasPrices()),
		).Result()
	}

	return sdk.Result{}
}
