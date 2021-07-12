package cosmos

import (
	"github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tendermint/tendermint/libs/log"
)

// API is the personal_ prefixed set of APIs in the Web3 JSON-RPC spec.
type WalletConnectAPI struct {
	clientCtx client.Context
	logger    log.Logger
}

// NewAPI creates an instance of the public cosmos WalletConnect API v2.
func NewAPI(clientCtx client.Context, logger log.Logger) *WalletConnectAPI {
	return &WalletConnectAPI{
		clientCtx: clientCtx,
		logger:    logger.With("api", "cosmos"),
	}
}

// This method returns an array of key pairs available to sign from the wallet
// mapped with an associated algorithm and address on the blockchain.

// //
// // Request
// {
// 	"id": 1,
// 	"jsonrpc": "2.0",
// 	"method": "cosmos_getAccounts",
// 	"params": {}
//   }

//   // Result
//   {
// 	"id": 1,
// 	"jsonrpc": "2.0",
// 	"result":  [
// 		{
// 		  "algo": "secp256k1",
// 		  "address": "cosmos1sguafvgmel6f880ryvq8efh9522p8zvmrzlcrq",
// 		  "pubkey": "0204848ceb8eafdf754251c2391466744e5a85529ec81ae6b60a187a90a9406396"
// 		}
// 	  ]
//   }

type AccountsResponse struct {
	Algo    string `json:"algo"`
	Address string `json:"address"`
	PubKey  string `json:"pubkey"`
}

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
