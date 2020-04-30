package types

import (
	"fmt"
	"math/big"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/ethermint/crypto"
	ethcmn "github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/rlp"

	"github.com/pkg/errors"
	"golang.org/x/crypto/sha3"
)

// GenerateEthAddress generates an Ethereum address.
func GenerateEthAddress() ethcmn.Address {
	priv, err := crypto.GenerateKey()
	if err != nil {
		panic(err)
	}

	return ethcrypto.PubkeyToAddress(priv.ToECDSA().PublicKey)
}

// ValidateSigner attempts to validate a signer for a given slice of bytes over
// which a signature and signer is given. An error is returned if address
// derived from the signature and bytes signed does not match the given signer.
func ValidateSigner(signBytes, sig []byte, signer ethcmn.Address) error {
	pk, err := ethcrypto.SigToPub(signBytes, sig)

	if err != nil {
		return errors.Wrap(err, "failed to derive public key from signature")
	} else if ethcrypto.PubkeyToAddress(*pk) != signer {
		return fmt.Errorf("invalid signature for signer: %s", signer)
	}

	return nil
}

func rlpHash(x interface{}) (hash ethcmn.Hash) {
	hasher := sha3.NewLegacyKeccak256()
	//nolint:gosec,errcheck
	rlp.Encode(hasher, x)
	hasher.Sum(hash[:0])

	return hash
}

// ResultData represents the data returned in an sdk.Result
type ResultData struct {
	Address ethcmn.Address  `json:"address"`
	Bloom   ethtypes.Bloom  `json:"bloom"`
	Logs    []*ethtypes.Log `json:"logs"`
	Ret     []byte          `json:"ret"`
	TxHash  ethcmn.Hash     `json:"tx_hash"`
}

// EncodeResultData takes all of the necessary data from the EVM execution
// and returns the data as a byte slice encoded with amino
func EncodeResultData(data *ResultData) ([]byte, error) {
	return ModuleCdc.MarshalBinaryLengthPrefixed(data)
}

// DecodeResultData decodes an amino-encoded byte slice into ResultData
func DecodeResultData(in []byte) (ResultData, error) {
	data := new(ResultData)
	err := ModuleCdc.UnmarshalBinaryLengthPrefixed(in, data)
	if err != nil {
		return ResultData{}, err
	}
	return *data, nil
}

// EncodeLogs encodes an array of logs using amino
func EncodeLogs(logs []*ethtypes.Log) ([]byte, error) {
	return ModuleCdc.MarshalBinaryLengthPrefixed(logs)
}

// DecodeLogs decodes an amino-encoded byte array into an array of logs
func DecodeLogs(in []byte) ([]*ethtypes.Log, error) {
	logs := []*ethtypes.Log{}
	err := ModuleCdc.UnmarshalBinaryLengthPrefixed(in, &logs)
	if err != nil {
		return nil, err
	}
	return logs, nil
}

// ----------------------------------------------------------------------------
// Auxiliary

// TxDecoder returns an sdk.TxDecoder that can decode both auth.StdTx and
// MsgEthereumTx transactions.
func TxDecoder(cdc *codec.Codec) sdk.TxDecoder {
	return func(txBytes []byte) (sdk.Tx, error) {
		var tx sdk.Tx

		if len(txBytes) == 0 {
			return nil, sdkerrors.Wrap(sdkerrors.ErrTxDecode, "tx bytes are empty")
		}

		// sdk.Tx is an interface. The concrete message types
		// are registered by MakeTxCodec
		err := cdc.UnmarshalBinaryBare(txBytes, &tx)
		if err != nil {
			return nil, sdkerrors.Wrap(sdkerrors.ErrTxDecode, err.Error())
		}

		return tx, nil
	}
}

// recoverEthSig recovers a signature according to the Ethereum specification and
// returns the sender or an error.
//
// Ref: Ethereum Yellow Paper (BYZANTIUM VERSION 69351d5) Appendix F
// nolint: gocritic
func recoverEthSig(R, S, Vb *big.Int, sigHash ethcmn.Hash) (ethcmn.Address, error) {
	if Vb.BitLen() > 8 {
		return ethcmn.Address{}, errors.New("invalid signature")
	}

	V := byte(Vb.Uint64() - 27)
	if !ethcrypto.ValidateSignatureValues(V, R, S, true) {
		return ethcmn.Address{}, errors.New("invalid signature")
	}

	// encode the signature in uncompressed format
	r, s := R.Bytes(), S.Bytes()
	sig := make([]byte, 65)

	copy(sig[32-len(r):32], r)
	copy(sig[64-len(s):64], s)
	sig[64] = V

	// recover the public key from the signature
	pub, err := ethcrypto.Ecrecover(sigHash[:], sig)
	if err != nil {
		return ethcmn.Address{}, err
	}

	if len(pub) == 0 || pub[0] != 4 {
		return ethcmn.Address{}, errors.New("invalid public key")
	}

	var addr ethcmn.Address
	copy(addr[:], ethcrypto.Keccak256(pub[1:])[12:])

	return addr, nil
}
