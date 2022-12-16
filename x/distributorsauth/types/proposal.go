package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
)

const (
	// ProposalTypeChange defines the type for a ParameterChangeProposal
	ProposalTypeAddDistributor = "AddDistributorGov"
)

// Assert ParameterChangeProposal implements govtypes.Content at compile-time
var _ govtypes.Content = &AddDistributorProposal{}

func init() {
	govtypes.RegisterProposalType(ProposalTypeAddDistributor)
}

func NewAddDistributorProposal(title, description, address, end_date, deposit string) *AddDistributorProposal {
	return &AddDistributorProposal{title, description, address, end_date, deposit}
}

// GetTitle returns the title of a parameter change proposal.
func (pcp *AddDistributorProposal) GetTitle() string { return pcp.Title }

// GetDescription returns the description of a parameter change proposal.
func (pcp *AddDistributorProposal) GetDescription() string { return pcp.Description }

// ProposalRoute returns the routing key of a parameter change proposal.
func (pcp *AddDistributorProposal) ProposalRoute() string { return RouterKey }

// ProposalType returns the type of a parameter change proposal.
func (pcp *AddDistributorProposal) ProposalType() string { return ProposalTypeAddDistributor }

// ValidateBasic validates the parameter change proposal
func (pcp *AddDistributorProposal) ValidateBasic() error {
	err := govtypes.ValidateAbstract(pcp)
	if err != nil {
		return err
	}

	return ValidateDistributor(pcp.Address)
}

func ValidateDistributor(address string) error {
	_, err := sdk.AccAddressFromBech32(address)
	if err != nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "Wrong distributor address")
	}

	return nil
}
