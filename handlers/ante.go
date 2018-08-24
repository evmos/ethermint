package handlers

import (
	"fmt"
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"

	"github.com/cosmos/ethermint/types"
	ethcmn "github.com/ethereum/go-ethereum/common"

	ethtypes "github.com/ethereum/go-ethereum/core/types"
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

// AnteHandler is responsible for attempting to route an Ethereum or SDK
// transaction to an internal ante handler for performing transaction-level
// processing (e.g. fee payment, signature verification) before being passed
// onto it's respective handler.
func AnteHandler(am auth.AccountMapper) sdk.AnteHandler {
	return func(sdkCtx sdk.Context, tx sdk.Tx) (newCtx sdk.Context, res sdk.Result, abort bool) {
		var (
			gasLimit int64
			handler  internalAnteHandler
		)

		switch tx := tx.(type) {
		case types.Transaction:
			gasLimit = int64(tx.Data.GasLimit)
			handler = handleEthTx
		case auth.StdTx:
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
	gethTx := ethTx.ConvertTx(chainID)
	signer := ethtypes.NewEIP155Signer(chainID)

	_, err := signer.Sender(&gethTx)
	if err != nil {
		return sdkCtx, sdk.ErrUnauthorized("signature verification failed").Result(), true
	}

	return sdkCtx, sdk.Result{GasWanted: int64(ethTx.Data.GasLimit)}, false
}

// handleEmbeddedTx implements an ante handler for an SDK transaction. It
// validates the signature and if valid returns an OK result.
func handleEmbeddedTx(sdkCtx sdk.Context, tx sdk.Tx, am auth.AccountMapper) (sdk.Context, sdk.Result, bool) {
	stdTx, ok := tx.(auth.StdTx)
	if !ok {
		return sdkCtx, sdk.ErrInternal(fmt.Sprintf("invalid transaction: %T", tx)).Result(), true
	}

	if err := validateStdTxBasic(stdTx); err != nil {
		return sdkCtx, err.Result(), true
	}

	signerAddrs := stdTx.GetSigners()
	signerAccs := make([]auth.Account, len(signerAddrs))

	// validate signatures
	for i, sig := range stdTx.Signatures {
		signer := ethcmn.BytesToAddress(signerAddrs[i].Bytes())

		signerAcc, err := validateSignature(sdkCtx, stdTx, signer, sig, am)
		if err.Code() != sdk.CodeOK {
			return sdkCtx, err.Result(), false
		}

		// TODO: Fees!

		am.SetAccount(sdkCtx, signerAcc)
		signerAccs[i] = signerAcc
	}

	newCtx := auth.WithSigners(sdkCtx, signerAccs)

	return newCtx, sdk.Result{GasWanted: stdTx.Fee.Gas}, false
}

// validateStdTxBasic validates an auth.StdTx based on parameters that do not
// depend on the context.
func validateStdTxBasic(stdTx auth.StdTx) (err sdk.Error) {
	sigs := stdTx.Signatures
	if len(sigs) == 0 {
		return sdk.ErrUnauthorized("transaction missing signatures")
	}

	signerAddrs := stdTx.GetSigners()
	if len(sigs) != len(signerAddrs) {
		return sdk.ErrUnauthorized("invalid number of transaction signers")
	}

	return nil
}

func validateSignature(
	sdkCtx sdk.Context, stdTx auth.StdTx, signer ethcmn.Address,
	sig auth.StdSignature, am auth.AccountMapper,
) (acc auth.Account, sdkErr sdk.Error) {

	chainID := sdkCtx.ChainID()

	acc = am.GetAccount(sdkCtx, signer.Bytes())
	if acc == nil {
		return nil, sdk.ErrUnknownAddress(fmt.Sprintf("no account with address %s found", signer))
	}

	accNum := acc.GetAccountNumber()
	if accNum != sig.AccountNumber {
		return nil, sdk.ErrInvalidSequence(
			fmt.Sprintf("invalid account number; got %d, expected %d", sig.AccountNumber, accNum))
	}

	accSeq := acc.GetSequence()
	if accSeq != sig.Sequence {
		return nil, sdk.ErrInvalidSequence(
			fmt.Sprintf("invalid account sequence; got %d, expected %d", sig.Sequence, accSeq))
	}

	err := acc.SetSequence(accSeq + 1)
	if err != nil {
		return nil, sdk.ErrInternal(err.Error())
	}

	signBytes := types.GetStdTxSignBytes(chainID, accNum, accSeq, stdTx.Fee, stdTx.GetMsgs(), stdTx.Memo)

	// consume gas for signature verification
	sdkCtx.GasMeter().ConsumeGas(verifySigCost, "ante signature verification")

	if err := types.ValidateSigner(signBytes, sig.Signature, signer); err != nil {
		return nil, sdk.ErrUnauthorized(err.Error())
	}

	return
}
