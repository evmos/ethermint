package rest

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/cosmos/cosmos-sdk/client/context"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/rest"
	authclient "github.com/cosmos/cosmos-sdk/x/auth/client/utils"

	"github.com/cosmos/ethermint/x/faucet/types"
)

// RegisterRoutes register REST endpoints for the faucet module
func RegisterRoutes(clientCtx context.CLIContext, r *mux.Router) {
	r.HandleFunc(fmt.Sprintf("/%s/request", types.ModuleName), requestHandler(clientCtx)).Methods("POST")
	r.HandleFunc(fmt.Sprintf("/%s/funded", types.ModuleName), fundedHandlerFn(clientCtx)).Methods("GET")
}

// PostRequestBody defines fund request's body.
type PostRequestBody struct {
	BaseReq   rest.BaseReq `json:"base_req" yaml:"base_req"`
	Amount    sdk.Coins    `json:"amount" yaml:"amount"`
	Recipient string       `json:"receipient" yaml:"receipient"`
}

func requestHandler(clientCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req PostRequestBody
		if !rest.ReadRESTReq(w, r, clientCtx.Codec, &req) {
			rest.WriteErrorResponse(w, http.StatusBadRequest, "failed to parse request")
			return
		}

		baseReq := req.BaseReq.Sanitize()
		if !baseReq.ValidateBasic(w) {
			return
		}

		sender, err := sdk.AccAddressFromBech32(baseReq.From)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		var recipient sdk.AccAddress
		if req.Recipient == "" {
			recipient = sender
		} else {
			recipient, err = sdk.AccAddressFromBech32(req.Recipient)
		}

		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		msg := types.NewMsgFund(req.Amount, sender, recipient)
		err = msg.ValidateBasic()
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		authclient.WriteGenerateStdTxResponse(w, clientCtx, baseReq, []sdk.Msg{msg})
	}
}

func fundedHandlerFn(clientCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		res, height, err := clientCtx.Query(fmt.Sprintf("custom/%s/%s", types.QuerierRoute, types.QueryFunded))
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		clientCtx = clientCtx.WithHeight(height)
		rest.PostProcessResponse(w, clientCtx, res)
	}
}
