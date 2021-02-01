package rest

import (
	"encoding/hex"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/types/rest"
	authrest "github.com/cosmos/cosmos-sdk/x/auth/client/rest"
	"github.com/cosmos/cosmos-sdk/x/auth/client/utils"
	rpctypes "github.com/cosmos/ethermint/rpc/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gorilla/mux"
)

// RegisterRoutes - Central function to define routes that get registered by the main application
func RegisterRoutes(cliCtx context.CLIContext, r *mux.Router) {
	r.HandleFunc("/txs/{hash}", QueryTxRequestHandlerFn(cliCtx)).Methods("GET")
	r.HandleFunc("/txs", authrest.QueryTxsRequestHandlerFn(cliCtx)).Methods("GET")         // default from auth
	r.HandleFunc("/txs", authrest.BroadcastTxRequest(cliCtx)).Methods("POST")              // default from auth
	r.HandleFunc("/txs/encode", authrest.EncodeTxRequestHandlerFn(cliCtx)).Methods("POST") // default from auth
	r.HandleFunc("/txs/decode", authrest.DecodeTxRequestHandlerFn(cliCtx)).Methods("POST") // default from auth
}

func QueryTxRequestHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		hashHexStr := vars["hash"]

		cliCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
		if !ok {
			return
		}

		ethHashPrefix := "0x"
		if strings.HasPrefix(hashHexStr, ethHashPrefix) {
			// eth Tx
			ethHashPrefixLength := len(ethHashPrefix)
			output, err := getEthTransactionByHash(cliCtx, hashHexStr[ethHashPrefixLength:])
			if err != nil {
				rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
				return
			}
			rest.PostProcessResponseBare(w, cliCtx, output)
			return
		}

		output, err := utils.QueryTx(cliCtx, hashHexStr)
		if err != nil {
			if strings.Contains(err.Error(), hashHexStr) {
				rest.WriteErrorResponse(w, http.StatusNotFound, err.Error())
				return
			}
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}
		rest.PostProcessResponseBare(w, cliCtx, output)
	}

}

// GetTransactionByHash returns the transaction identified by hash.
func getEthTransactionByHash(cliCtx context.CLIContext, hashHex string) ([]byte, error) {
	hash, err := hex.DecodeString(hashHex)
	if err != nil {
		return nil, err
	}
	node, err := cliCtx.GetNode()
	if err != nil {
		return nil, err
	}
	tx, err := node.Tx(hash, false)
	if err != nil {
		return nil, err
	}

	// Can either cache or just leave this out if not necessary
	block, err := node.Block(&tx.Height)
	if err != nil {
		return nil, err
	}

	blockHash := common.BytesToHash(block.Block.Hash())

	ethTx, err := rpctypes.RawTxToEthTx(cliCtx, tx.Tx)
	if err != nil {
		return nil, err
	}

	height := uint64(tx.Height)
	res, err := rpctypes.NewTransaction(ethTx, common.BytesToHash(tx.Tx.Hash()), blockHash, height, uint64(tx.Index))
	if err != nil {
		return nil, err
	}
	return json.Marshal(res)
}
