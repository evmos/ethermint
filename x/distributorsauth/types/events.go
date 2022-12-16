package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// distributorsauth module event types
const (
	EventTypeAddDistributor    = "AddDistributor"
	EventTypeRemoveDistributor = "RemoveDistributor"

	EventTypeAddAdmin    = "AddAdmin"
	EventTypeRemoveAdmin = "RemoveAdmin"

	EventTypeAddDistributorProposal = "AddDistributorProposal"

	AttributeKeyAddress    = "address"
	AttributeKeyEndDate    = "endDate"
	AttributeKeyEditOption = "editOption"
	//AttributeKeySender    = sdk.AttributeKeySender
)

// AddDistributorEvent constructs a new distributor add sdk.Event
func AddDistributorEvent(address string, endDate string) sdk.Event {
	return sdk.NewEvent(
		EventTypeAddDistributor,
		sdk.NewAttribute(AttributeKeyAddress, address),
		sdk.NewAttribute(AttributeKeyEndDate, endDate),
	)
}

// RemoveDistributorEvent constructs distributor remove sdk.Event
func RemoveDistributorEvent(address string) sdk.Event {
	return sdk.NewEvent(
		EventTypeRemoveDistributor,
		sdk.NewAttribute(AttributeKeyAddress, address),
	)
}

// AddAdminEvent construct a new admin add sdk.Event
func AddAdminEvent(address string, edit_option string) sdk.Event {
	return sdk.NewEvent(
		EventTypeAddAdmin,
		sdk.NewAttribute(AttributeKeyAddress, address),
		sdk.NewAttribute(AttributeKeyEditOption, edit_option),
	)
}

// RemoveAdminEvent construct admin remove sdk.Event
func RemoveAdminEvent(address string) sdk.Event {
	return sdk.NewEvent(
		EventTypeRemoveAdmin,
		sdk.NewAttribute(AttributeKeyAddress, address),
	)
}

// AddDistributoProposalEvent constructs a new distributor add proposal sdk.Event
func AddDistributoProposalEvent(address string, endDate string) sdk.Event {
	return sdk.NewEvent(
		EventTypeAddDistributor,
		sdk.NewAttribute(AttributeKeyAddress, address),
		sdk.NewAttribute(AttributeKeyEndDate, endDate),
	)
}
