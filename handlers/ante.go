package handlers

import (
	"fmt"
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"

	"github.com/cosmos/ethermint/types"
	ethcmn "github.com/ethereum/go-ethereum/common"
)

const (
	// TODO: Ported from the SDK and may have a different context/value for
	// Ethermint.
	verifySigCost = 100
)

// internalAnteHandler reflects a function signature an internal ante handler
// must implementing. Internal ante handlers will be dependant upon the
// transaction type.
type internalAnteHandler func(
	sdkCtx sdk.Context, tx sdk.Tx, am auth.AccountMapper,
) (newCtx sdk.Context, res sdk.Result, abort bool)

// AnteHandler handles Ethereum transactions and passes SDK transactions to the
// embeddedAnteHandler if it's an Ethermint transaction. The ante handler gets
// invoked after the BaseApp performs the runTx. At this point, the transaction
// should be properly decoded via the TxDecoder and should be of a proper type,
// Transaction or EmbeddedTx.
func AnteHandler(am auth.AccountMapper) sdk.AnteHandler {
	return func(sdkCtx sdk.Context, tx sdk.Tx) (newCtx sdk.Context, res sdk.Result, abort bool) {
		var (
			handler  internalAnteHandler
			gasLimit int64
		) 

		switch tx := tx.(type) {
		case types.Transaction:
			gasLimit = int64(tx.Data.GasLimit)
			handler = handleEthTx
		case types.EmbeddedTx:
			gasLimit = tx.Fee.Gas
			handler = handleEmbeddedTx
		default:
			return sdkCtx, sdk.ErrInternal(fmt.Sprintf("invalid transaction: %T", tx)).Result(), true
		}

		newCtx = sdkCtx.WithGasMeter(sdk.NewGasMeter(gasLimit))

		// AnteHandlers must have their own defer/recover in order for the
		// BaseApp to know how much gas was used! This is because the GasMeter
		// is created in the AnteHandler, but if it panics the context won't be
		// set properly in runTx's recover.
		defer func() {
			if r := recover(); r != nil {
				switch rType := r.(type) {
				case sdk.ErrorOutOfGas:
					log := fmt.Sprintf("out of gas in location: %v", rType.Descriptor)
					res = sdk.ErrOutOfGas(log).Result()
					res.GasWanted = gasLimit
					res.GasUsed = newCtx.GasMeter().GasConsumed()
					abort = true
				default:
					panic(r)
				}
			}
		}()

		return handler(newCtx, tx, am)
	}
}

// handleEthTx implements an ante handler for an Ethereum transaction. It
// validates the signature and if valid returns an OK result.
//
// TODO: Do we need to do any further validation or account manipulation
// (e.g. increment nonce)?
func handleEthTx(sdkCtx sdk.Context, tx sdk.Tx, am auth.AccountMapper) (sdk.Context, sdk.Result, bool) {
	ethTx, ok := tx.(types.Transaction)
	if !ok {
		return sdkCtx, sdk.ErrInternal(fmt.Sprintf("invalid transaction: %T", tx)).Result(), true
	}

	// the SDK chainID is a string representation of integer
	chainID, ok := new(big.Int).SetString(sdkCtx.ChainID(), 10)
	if !ok {
		// TODO: ErrInternal may not be correct error to throw here?
		return sdkCtx, sdk.ErrInternal(fmt.Sprintf("invalid chainID: %s", sdkCtx.ChainID())).Result(), true
	}

	// validate signature
	sdkCtx.GasMeter().ConsumeGas(verifySigCost, "ante verify")
	addr, err := ethTx.VerifySig(chainID)

	if err != nil {
		return sdkCtx, sdk.ErrUnauthorized("signature verification failed").Result(), true
	}

	// validate AccountNonce (called Sequence in AccountMapper)
	acc := am.GetAccount(sdkCtx, addr[:])
	seq := acc.GetSequence()
	if ethTx.Data.AccountNonce != uint64(seq) {
		return sdkCtx, sdk.ErrInvalidSequence(fmt.Sprintf("Wrong AccountNonce: expected %d", seq)).Result(), true
	}
	err = acc.SetSequence(seq + 1)
	if err != nil {
		panic(err)
	}
	am.SetAccount(sdkCtx, acc)

	return sdkCtx, sdk.Result{GasWanted: int64(ethTx.Data.GasLimit)}, false
}

// handleEmbeddedTx implements an ante handler for an SDK transaction. It
// validates the signature and if valid returns an OK result.
func handleEmbeddedTx(sdkCtx sdk.Context, tx sdk.Tx, am auth.AccountMapper) (sdk.Context, sdk.Result, bool) {
	etx, ok := tx.(types.EmbeddedTx)
	if !ok {
		return sdkCtx, sdk.ErrInternal(fmt.Sprintf("invalid transaction: %T", tx)).Result(), true
	}

	if err := validateEmbeddedTxBasic(etx); err != nil {
		return sdkCtx, err.Result(), true
	}

	signerAddrs := etx.GetRequiredSigners()
	signerAccs := make([]auth.Account, len(signerAddrs))

	// validate signatures
	for i, sig := range etx.Signatures {
		signer := ethcmn.BytesToAddress(signerAddrs[i].Bytes())

		signerAcc, err := validateSignature(sdkCtx, etx, signer, sig, am)
		if err != nil {
			return sdkCtx, err.Result(), true
		}

		// TODO: Fees!

		am.SetAccount(sdkCtx, signerAcc)
		signerAccs[i] = signerAcc
	}

	newCtx := auth.WithSigners(sdkCtx, signerAccs)

	return newCtx, sdk.Result{GasWanted: etx.Fee.Gas}, false
}

// validateEmbeddedTxBasic validates an EmbeddedTx based on things that don't
// depend on the context.
func validateEmbeddedTxBasic(etx types.EmbeddedTx) (err sdk.Error) {
	sigs := etx.Signatures
	if len(sigs) == 0 {
		return sdk.ErrUnauthorized("transaction missing signatures")
	}

	signerAddrs := etx.GetRequiredSigners()
	if len(sigs) != len(signerAddrs) {
		return sdk.ErrUnauthorized("invalid number of transaction signers")
	}

	return nil
}

func validateSignature(
	sdkCtx sdk.Context, etx types.EmbeddedTx, signer ethcmn.Address,
	sig []byte, am auth.AccountMapper,
) (acc auth.Account, sdkErr sdk.Error) {

	chainID := sdkCtx.ChainID()

	acc = am.GetAccount(sdkCtx, signer.Bytes())
	if acc == nil {
		return nil, sdk.ErrUnknownAddress(fmt.Sprintf("no account with address %s found", signer))
	}

	signEtx := types.EmbeddedTxSign{
		ChainID:       chainID,
		AccountNumber: acc.GetAccountNumber(),
		Sequence:      acc.GetSequence(),
		Messages:      etx.Messages,
		Fee:           etx.Fee,
	}

	err := acc.SetSequence(signEtx.Sequence + 1)
	if err != nil {
		return nil, sdk.ErrInternal(err.Error())
	}

	signBytes, err := signEtx.Bytes()
	if err != nil {
		return nil, sdk.ErrInternal(err.Error())
	}

	// consume gas for signature verification
	sdkCtx.GasMeter().ConsumeGas(verifySigCost, "ante verify")

	if err := types.ValidateSigner(signBytes, sig, signer); err != nil {
		return nil, sdk.ErrUnauthorized(err.Error())
	}

	return
}
