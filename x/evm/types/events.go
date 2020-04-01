package types

// Evm module events
const (
	EventTypeEthermint  = TypeMsgEthermint
	EventTypeEthereumTx = TypeMsgEthereumTx

	AttributeKeyContractAddress = "contract"
	AttributeKeyRecipient       = "recipient"
	AttributeValueCategory      = ModuleName
)
