package cosmos

import (
	"encoding/hex"
	"strconv"

	"github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"
	txtypes "github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/cosmos/cosmos-sdk/x/auth/legacy/legacytx"
	"github.com/tendermint/tendermint/libs/log"
)

// API is the personal_ prefixed set of APIs in the Web3 JSON-RPC spec
type WalletConnectAPI struct {
	clientCtx client.Context
	logger    log.Logger
}

// NewAPI creates an instance of the public cosmos WalletConnect API v2
func NewAPI(clientCtx client.Context, logger log.Logger) *WalletConnectAPI {
	return &WalletConnectAPI{
		clientCtx: clientCtx,
		logger:    logger.With("api", "cosmos"),
	}
}

type AccountsResponse struct {
	Algo    string `json:"algo"`
	Address string `json:"address"`
	PubKey  string `json:"pubkey"`
}

// GetAccounts returns an array of key pairs available to sign from the wallet
// mapped with an associated algorithm and address on the blockchain
func (api *WalletConnectAPI) GetAccounts() ([]AccountsResponse, error) {
	api.logger.Debug("cosmos_getAccounts")
	accs := []AccountsResponse{}

	list, err := api.clientCtx.Keyring.List()
	if err != nil {
		return nil, err
	}
	for _, info := range list {

		addr := sdk.AccAddress(info.GetAddress())
		acc := AccountsResponse{
			Algo:    string(info.GetAlgo()),
			Address: addr.String(),
			PubKey:  info.GetPubKey().String(),
		}
		accs = append(accs, acc)
	}

	return accs, nil
}

// SignDocDirect gets converted to txtypes.SignDoc
type SignDocDirect struct {
	// BodyBytes is protobuf serialization of a TxBody that matches the
	// representation in TxRaw.
	BodyBytes string `json:"bodyBytes,omitempty"`
	// AuthInfoBytes is a protobuf serialization of an AuthInfo that matches the
	// representation in TxRaw.
	AuthInfoBytes string `json:"authInfoBytes,omitempty"`
	// ChainId is the unique identifier of the chain this transaction targets.
	// It prevents signed transactions from being used on another chain by an
	// attacker
	ChainID string `json:"chainId,omitempty"`
	// AccountNumber is the account number of the account in state
	AccountNumber string `json:"accountNumber,omitempty"`
}

func (api *WalletConnectAPI) convertToTxType(signDoc SignDocDirect) (txtypes.SignDoc, error) {
	accountNumber, err := strconv.ParseUint(signDoc.AccountNumber, 10, 64)
	if err != nil {
		api.logger.Error("failed to parse account number: %s, err: %s", signDoc.AccountNumber, err.Error())
		return txtypes.SignDoc{}, err
	}
	return txtypes.SignDoc{
		BodyBytes:     []byte(signDoc.BodyBytes),
		AuthInfoBytes: []byte(signDoc.AuthInfoBytes),
		ChainId:       signDoc.ChainID,
		AccountNumber: accountNumber,
	}, nil
}

type SignDirectRequest struct {
	SignerAddress sdk.AccAddress `json:"signerAddress"`
	SignDoc       SignDocDirect  `json:"signDoc"`
}
type SignDirectResponse struct {
	Signature string        `json:"signature"`
	SignDoc   SignDocDirect `json:"signDoc"`
}

// This method returns a signature for the provided document to be signed
// targetting the requested signer address corresponding to the keypair returned
// by the account data.
func (api *WalletConnectAPI) SignDirect(req SignDirectRequest) (SignDirectResponse, error) {
	api.logger.Debug("cosmos_signDirect")

	_, err := api.clientCtx.Keyring.KeyByAddress(req.SignerAddress)
	if err != nil {
		api.logger.Error("failed to find key in keyring", "address", req.SignerAddress.String())
		return SignDirectResponse{}, err
	}
	signDoc, err := api.convertToTxType(req.SignDoc)
	if err != nil {
		api.logger.Error("failed to convert signDoc to txType")
		return SignDirectResponse{}, err
	}
	signBytes, err := signDoc.Marshal()
	if err != nil {
		api.logger.Error("failed to unpack tx data")
		return SignDirectResponse{}, err
	}
	signature, _, err := api.clientCtx.Keyring.SignByAddress(req.SignerAddress, signBytes)
	if err != nil {
		api.logger.Error("keyring.SignByAddress failed", "address", req.SignerAddress.String())
		return SignDirectResponse{}, err
	}
	return SignDirectResponse{
		Signature: hex.EncodeToString(signature),
		SignDoc:   req.SignDoc,
	}, nil
}

// SignDocAmino gets converted to txtypes.SignDoc
type SignDocAmino struct {
	AccountNumber string `json:"account_number"`
	ChainID       string `json:"chain_id"`
	Sequence      string `json:"sequence"`
	Memo          string `json:"memo"`
}

func (api *WalletConnectAPI) convertToLegacyTx(signDoc SignDocAmino) (txtypes.SignDoc, error) {
	accountNumber, err := strconv.ParseUint(signDoc.AccountNumber, 10, 64)
	if err != nil {
		api.logger.Error("failed to parse account number: %s, err: %s", signDoc.AccountNumber, err.Error())
		return txtypes.SignDoc{}, err
	}
	// TODO
	return legacytx.StdSignDoc{}
	// return txtypes.SignDoc{
	// 	BodyBytes:     []byte(signDoc.BodyBytes),
	// 	AuthInfoBytes: []byte(signDoc.AuthInfoBytes),
	// 	ChainId:       signDoc.ChainId,
	// 	AccountNumber: accountNumber,
	// }, nil
}

type SignAminoRequest struct {
	SignerAddress sdk.AccAddress `json:"signerAddress"`
	SignDoc       SignDocAmino   `json:"signDoc"`
}
type SignAminoResponse struct {
	Signature string       `json:"signature"`
	SignDoc   SignDocAmino `json:"signDoc"`
}

// This method returns a signature for the provided document to be signed
// targetting the requested signer address corresponding to the keypair returned
// by the account data.
func (api *WalletConnectAPI) SignAmino(req SignAminoRequest) (SignAminoResponse, error) {

}

// // Sign signs the provided data using the private key of address via Geth's signature standard.
// func (e *PublicAPI) Sign(address common.Address, data hexutil.Bytes) (hexutil.Bytes, error) {
// 	e.logger.Debug("eth_sign", "address", address.Hex(), "data", common.Bytes2Hex(data))

// 	from := sdk.AccAddress(address.Bytes())

// 	_, err := e.clientCtx.Keyring.KeyByAddress(from)
// 	if err != nil {
// 		e.logger.Error("failed to find key in keyring", "address", address.String())
// 		return nil, fmt.Errorf("%s; %s", keystore.ErrNoMatch, err.Error())
// 	}

// 	// Sign the requested hash with the wallet
// 	signature, _, err := e.clientCtx.Keyring.SignByAddress(from, data)
// 	if err != nil {
// 		e.logger.Error("keyring.SignByAddress failed", "address", address.Hex())
// 		return nil, err
// 	}

// 	signature[64] += 27 // Transform V from 0/1 to 27/28 according to the yellow paper
// 	return signature, nil
// }
