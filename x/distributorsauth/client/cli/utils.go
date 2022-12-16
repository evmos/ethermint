package cli

import (
	"os"

	"github.com/Entangle-Protocol/entangle-blockchain/x/distributorsauth/types"
	"github.com/cosmos/cosmos-sdk/codec"
)

// ParseAddDistributorProposal reads and parses a ParseAddDistributorProposal from a file.
func ParseAddDistributorProposal(cdc codec.JSONCodec, proposalFile string) (types.AddDistributorProposal, error) {
	proposal := types.AddDistributorProposal{}

	contents, err := os.ReadFile(proposalFile)
	if err != nil {
		return proposal, err
	}

	if err = cdc.UnmarshalJSON(contents, &proposal); err != nil {
		return proposal, err
	}

	return proposal, nil
}
