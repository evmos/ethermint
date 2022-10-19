package eip712

import (
	"errors"
	"fmt"

	"github.com/cosmos/cosmos-sdk/simapp/params"
	"github.com/cosmos/cosmos-sdk/x/auth/migrations/legacytx"

	cosmosTypes "github.com/cosmos/cosmos-sdk/types"
	txTypes "github.com/cosmos/cosmos-sdk/types/tx"

	apitypes "github.com/ethereum/go-ethereum/signer/core/apitypes"
	ethermint "github.com/evmos/ethermint/types"

	"github.com/cosmos/cosmos-sdk/codec"
)

var ethermintProtoCodec codec.ProtoCodecMarshaler
var ethermintAminoCodec *codec.LegacyAmino

func SetEncodingConfig(cfg params.EncodingConfig) {
	ethermintAminoCodec = cfg.Amino
	ethermintProtoCodec = codec.NewProtoCodec(cfg.InterfaceRegistry)
}

func GetEIP712HashForMsg(signDocBytes []byte) ([]byte, error) {
	var typedData apitypes.TypedData

	typedDataAmino, errAmino := decodeAminoSignDoc(signDocBytes)
	typedDataProtobuf, errProtobuf := decodeProtobufSignDoc(signDocBytes)

	if errAmino == nil {
		typedData = typedDataAmino
	} else if errProtobuf == nil {
		typedData = typedDataProtobuf
	} else {
		return make([]byte, 0), errors.New(fmt.Sprintf("could not decode sign doc as either Amino or Protobuf.\n amino: %v\n protobuf: %v\n", errAmino, errProtobuf))
	}

	return typedData.TypeHash("Tx"), nil
}

// Attempt to decode the SignDoc bytes as an Amino SignDoc and return an error on failure
func decodeAminoSignDoc(signDocBytes []byte) (apitypes.TypedData, error) {
	var (
		aminoDoc legacytx.StdSignDoc
		err      error
	)

	err = ethermintAminoCodec.UnmarshalJSON(signDocBytes, &aminoDoc)
	if err != nil {
		return apitypes.TypedData{}, err
	}

	// Unwrap fees
	var fees legacytx.StdFee
	err = ethermintAminoCodec.UnmarshalJSON(aminoDoc.Fee, &fees)
	if err != nil {
		return apitypes.TypedData{}, err
	}

	if len(aminoDoc.Msgs) != 1 {
		return apitypes.TypedData{}, errors.New(fmt.Sprintf("Invalid number of messages in SignDoc, expected 1 but got %v\n", len(aminoDoc.Msgs)))
	}

	var msg cosmosTypes.Msg
	err = ethermintAminoCodec.UnmarshalJSON(aminoDoc.Msgs[0], &msg)
	if err != nil {
		fmt.Printf("Encountered err %v\n", err)
		return apitypes.TypedData{}, err
	}

	// By default, use first address in list of signers to cover fee
	// Currently, support only one signer
	if len(msg.GetSigners()) != 1 {
		return apitypes.TypedData{}, errors.New("Expected exactly one signer for message")
	}
	feePayer := msg.GetSigners()[0]
	feeDelegation := &FeeDelegationOptions{
		FeePayer: feePayer,
	}

	// Parse ChainID
	chainID, err := ethermint.ParseChainID(aminoDoc.ChainID)
	if err != nil {
		return apitypes.TypedData{}, errors.New("Invalid chain ID passed as argument")
	}

	typedData, err := WrapTxToTypedData(
		ethermintProtoCodec,
		chainID.Uint64(),
		msg,
		signDocBytes, // Amino StdSignDocBytes
		feeDelegation,
	)

	if err != nil {
		return apitypes.TypedData{}, errors.New(fmt.Sprintf("Could not convert to EIP712 representation: %v\n", err))
	}

	return typedData, nil
}

// Attempt to decode the SignDoc bytes as a Protobuf SignDoc and return an error on failure
func decodeProtobufSignDoc(signDocBytes []byte) (apitypes.TypedData, error) {
	// Decode sign doc
	signDoc := &txTypes.SignDoc{}
	err := signDoc.Unmarshal(signDocBytes)
	if err != nil {
		return apitypes.TypedData{}, err
	}

	// Decode auth info
	authInfo := &txTypes.AuthInfo{}
	err = authInfo.Unmarshal(signDoc.AuthInfoBytes)
	if err != nil {
		return apitypes.TypedData{}, err
	}

	// Decode body
	body := &txTypes.TxBody{}
	err = body.Unmarshal(signDoc.BodyBytes)
	if err != nil {
		return apitypes.TypedData{}, err
	}

	// Until support for these fields is added, throw an error at their presence
	if body.TimeoutHeight != 0 || len(body.ExtensionOptions) != 0 || len(body.NonCriticalExtensionOptions) != 0 {
		return apitypes.TypedData{}, errors.New("Body contains unsupported fields: TimeoutHeight, ExtensionOptions, or NonCriticalExtensionOptions")
	}

	// Verify single message
	if len(body.Messages) != 1 {
		return apitypes.TypedData{}, errors.New(fmt.Sprintf("Invalid number of messages, expected 1 got %v\n", len(body.Messages)))
	}

	// Verify single signature (single signer for now)
	if len(authInfo.SignerInfos) != 1 {
		return apitypes.TypedData{}, errors.New(fmt.Sprintf("Invalid number of signers, expected 1 got %v\n", len(authInfo.SignerInfos)))
	}

	// Decode signer info (single signer for now)
	signerInfo := authInfo.SignerInfos[0]

	// Parse ChainID
	chainID, err := ethermint.ParseChainID(signDoc.ChainId)
	if err != nil {
		return apitypes.TypedData{}, errors.New("Invalid chain ID passed as argument")
	}

	// Create StdFee
	stdFee := &legacytx.StdFee{
		Amount: authInfo.Fee.Amount,
		Gas:    authInfo.Fee.GasLimit,
	}

	// Parse Message (single message only)
	var msg cosmosTypes.Msg
	err = ethermintProtoCodec.UnpackAny(body.Messages[0], &msg)
	if err != nil {
		return apitypes.TypedData{}, errors.New(fmt.Sprintf("Could not unpack message object with error %v\n", err))
	}

	// Init fee payer
	feePayer := msg.GetSigners()[0]
	feeDelegation := &FeeDelegationOptions{
		FeePayer: feePayer,
	}

	// Get tip
	tip := authInfo.Tip

	// Create Legacy SignBytes (expected type for WrapTxToTypedData)
	signBytes := legacytx.StdSignBytes(
		signDoc.ChainId,
		signDoc.AccountNumber,
		signerInfo.Sequence,
		body.TimeoutHeight,
		*stdFee,
		[]cosmosTypes.Msg{msg},
		body.Memo,
		tip,
	)

	typedData, err := WrapTxToTypedData(
		ethermintProtoCodec,
		chainID.Uint64(),
		msg,
		signBytes,
		feeDelegation,
	)
	if err != nil {
		return apitypes.TypedData{}, err
	}

	return typedData, nil
}
