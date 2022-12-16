package cli

import (
	"github.com/Entangle-Protocol/entangle-blockchain/x/distributorsauth/client/cli"
	govclient "github.com/cosmos/cosmos-sdk/x/gov/client"
)

// ProposalHandler is the add distributor proposal handler.
var (
	DistributorsAuthProposalHandler = govclient.NewProposalHandler(cli.GetCmdSubmitDistributorsAuthProposal)
)
