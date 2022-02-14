package client

import (
	govclient "github.com/cosmos/cosmos-sdk/x/gov/client"
	"github.com/tharsis/ethermint/x/feemarket/client/rest"

	"github.com/tharsis/ethermint/x/feemarket/client/cli"
)

// ProposalHandler is the token mapping change proposal handler.
var ProposalHandler = govclient.NewProposalHandler(cli.NewSubmitBaseChangeProposalTxCmd, rest.ProposalRESTHandler)
