package types

import (
	"fmt"
	"strings"

	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
)

const (
	// ProposalTypeBaseFeeChange defines the type for a BaseFeeChangeProposal
	ProposalTypeBaseFeeChange = "BaseFeeChange"
)

// Assert TokenMappingChangeProposal implements govtypes.Content at compile-time
var _ govtypes.Content = &BaseFeeChangeProposal{}

func init() {
	govtypes.RegisterProposalType(ProposalTypeBaseFeeChange)
	govtypes.RegisterProposalTypeCodec(&BaseFeeChangeProposal{}, "feemarket/BaseFeeChangeProposal")
}

func NewBaseFeeChangeProposal(title, description string, baseFee uint64) *BaseFeeChangeProposal {
	return &BaseFeeChangeProposal{title, description, baseFee}
}

// GetTitle returns the title of a parameter change proposal.
func (bcp *BaseFeeChangeProposal) GetTitle() string { return bcp.Title }

// GetDescription returns the description of a parameter change proposal.
func (bcp *BaseFeeChangeProposal) GetDescription() string { return bcp.Description }

// ProposalRoute returns the routing key of a parameter change proposal.
func (bcp *BaseFeeChangeProposal) ProposalRoute() string { return RouterKey }

// ProposalType returns the type of a parameter change proposal.
func (bcp *BaseFeeChangeProposal) ProposalType() string { return ProposalTypeBaseFeeChange }

// ValidateBasic validates the parameter change proposal
func (bcp *BaseFeeChangeProposal) ValidateBasic() error {
	return nil
}

// String implements the Stringer interface.
func (bcp BaseFeeChangeProposal) String() string {
	var b strings.Builder

	b.WriteString(fmt.Sprintf(`BaseFee Change Proposal:
  Title:       %s
  Description: %s
  BaseFee:     %d
`, bcp.Title, bcp.Description, bcp.BaseFee))

	return b.String()
}
