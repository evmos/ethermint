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

var (
	ethermintProtoCodec codec.ProtoCodecMarshaler
	ethermintAminoCodec *codec.LegacyAmino
)

// SetEncodingConfig set the encoding config to the singleton codecs (Amino and Protobuf).
// The process of unmarshaling SignDoc bytes into a SignDoc object requires having a codec
// populated with all relevant message types. As a result, we must call this method on app
// initialization with the app's encoding config.
func SetEncodingConfig(cfg params.EncodingConfig) {
	ethermintAminoCodec = cfg.Amino
	ethermintProtoCodec = codec.NewProtoCodec(cfg.InterfaceRegistry)
}

// Get the EIP-712 object hash for the given SignDoc bytes by first decoding the bytes into
// an EIP-712 object, then hashing the EIP-712 object to create the bytes to be signed.
// See https://eips.ethereum.org/EIPS/eip-712 for more.
func GetEIP712HashForMsg(signDocBytes []byte) ([]byte, error) {
	typedData, err := GetEIP712TypedDataForMsg(signDocBytes)
	if err != nil {
		return nil, err
	}

	domainSeparator, err := typedData.HashStruct("EIP712Domain", typedData.Domain.Map())
	if err != nil {
		return nil, fmt.Errorf("could not hash EIP-712 domain: %w", err)
	}

	typedDataHash, err := typedData.HashStruct(typedData.PrimaryType, typedData.Message)
	if err != nil {
		return nil, fmt.Errorf("could not hash EIP-712 primary type: %w", err)
	}
	rawData := []byte(fmt.Sprintf("\x19\x01%s%s", string(domainSeparator), string(typedDataHash)))

	return rawData, nil
}

// GetEIP712TypedDataForMsg returns the EIP-712 TypedData representation for either
// Amino or Protobuf encoded signature doc bytes.
func GetEIP712TypedDataForMsg(signDocBytes []byte) (apitypes.TypedData, error) {
	// Attempt to decode as both Amino and Protobuf since the message format is unknown.
	// If either decode works, we can move forward with the corresponding typed data.
	typedDataAmino, errAmino := decodeAminoSignDoc(signDocBytes)
	if errAmino == nil && verifyEIP712Payload(typedDataAmino) {
		return typedDataAmino, nil
	}
	typedDataProtobuf, errProtobuf := decodeProtobufSignDoc(signDocBytes)
	if errProtobuf == nil && verifyEIP712Payload(typedDataProtobuf) {
		return typedDataProtobuf, nil
	}

	return apitypes.TypedData{}, fmt.Errorf("could not decode sign doc as either Amino or Protobuf. amino: %v protobuf: %v", errAmino, errProtobuf)
}

// verifyEIP712Payload ensures that the given TypedData does not contain empty fields from
// an improper initialization.
func verifyEIP712Payload(typedData apitypes.TypedData) bool {
	return len(typedData.Message) != 0 && len(typedData.Types) != 0 && typedData.PrimaryType != "" && typedData.Domain != apitypes.TypedDataDomain{}
}

// Attempt to decode the SignDoc bytes as an Amino SignDoc and return an error on failure
func decodeAminoSignDoc(signDocBytes []byte) (apitypes.TypedData, error) {
	var aminoDoc legacytx.StdSignDoc

	if err := ethermintAminoCodec.UnmarshalJSON(signDocBytes, &aminoDoc); err != nil {
		return apitypes.TypedData{}, err
	}

	// Unwrap fees
	var fees legacytx.StdFee
	if err := ethermintAminoCodec.UnmarshalJSON(aminoDoc.Fee, &fees); err != nil {
		return apitypes.TypedData{}, err
	}

	if len(aminoDoc.Msgs) != 1 {
		return apitypes.TypedData{}, fmt.Errorf("invalid number of messages in SignDoc, expected 1 but got %v", len(aminoDoc.Msgs))
	}

	var msg cosmosTypes.Msg
	if err := ethermintAminoCodec.UnmarshalJSON(aminoDoc.Msgs[0], &msg); err != nil {
		return apitypes.TypedData{}, fmt.Errorf("failed to unmarshal first message: %w", err)
	}

	// By default, use first address in list of signers to cover fee
	// Currently, support only one signer
	if len(msg.GetSigners()) != 1 {
		return apitypes.TypedData{}, errors.New("expected exactly one signer for message")
	}
	feePayer := msg.GetSigners()[0]
	feeDelegation := &FeeDelegationOptions{
		FeePayer: feePayer,
	}

	// Parse ChainID
	chainID, err := ethermint.ParseChainID(aminoDoc.ChainID)
	if err != nil {
		return apitypes.TypedData{}, errors.New("invalid chain ID passed as argument")
	}

	typedData, err := WrapTxToTypedData(
		ethermintProtoCodec,
		chainID.Uint64(),
		msg,
		signDocBytes, // Amino StdSignDocBytes
		feeDelegation,
	)
	if err != nil {
		return apitypes.TypedData{}, fmt.Errorf("could not convert to EIP712 representation: %w", err)
	}

	return typedData, nil
}

// Attempt to decode the SignDoc bytes as a Protobuf SignDoc and return an error on failure
func decodeProtobufSignDoc(signDocBytes []byte) (apitypes.TypedData, error) {
	// Decode sign doc
	signDoc := &txTypes.SignDoc{}
	if err := signDoc.Unmarshal(signDocBytes); err != nil {
		return apitypes.TypedData{}, err
	}

	// Decode auth info
	authInfo := &txTypes.AuthInfo{}
	if err := authInfo.Unmarshal(signDoc.AuthInfoBytes); err != nil {
		return apitypes.TypedData{}, err
	}

	// Decode body
	body := &txTypes.TxBody{}
	if err := body.Unmarshal(signDoc.BodyBytes); err != nil {
		return apitypes.TypedData{}, err
	}

	// Until support for these fields is added, throw an error at their presence
	if body.TimeoutHeight != 0 || len(body.ExtensionOptions) != 0 || len(body.NonCriticalExtensionOptions) != 0 {
		return apitypes.TypedData{}, errors.New("body contains unsupported fields: TimeoutHeight, ExtensionOptions, or NonCriticalExtensionOptions")
	}

	// Verify single message
	if len(body.Messages) != 1 {
		return apitypes.TypedData{}, fmt.Errorf("invalid number of messages, expected 1 got %v", len(body.Messages))
	}

	// Decode signer info (single signer for now)
	signerInfo := authInfo.SignerInfos[0]

	// Parse ChainID
	chainID, err := ethermint.ParseChainID(signDoc.ChainId)
	if err != nil {
		return apitypes.TypedData{}, fmt.Errorf("invalid chain ID passed as argument: %w", err)
	}

	// Create StdFee
	stdFee := &legacytx.StdFee{
		Amount: authInfo.Fee.Amount,
		Gas:    authInfo.Fee.GasLimit,
	}

	// Parse Message (single message only)
	var msg cosmosTypes.Msg
	if err := ethermintProtoCodec.UnpackAny(body.Messages[0], &msg); err != nil {
		return apitypes.TypedData{}, fmt.Errorf("could not unpack message object with error %w", err)
	}

	// Verify single signer (single signer for now)
	if len(msg.GetSigners()) != 1 {
		return apitypes.TypedData{}, fmt.Errorf("invalid number of signers, expected 1 got %v", len(authInfo.SignerInfos))
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
