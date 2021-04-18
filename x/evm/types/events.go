package types

// Evm module events
const (
	EventTypeEthereumTx = TypeMsgEthereumTx

	AttributeKeyContractAddress = "contract"
	AttributeKeyRecipient       = "recipient"
	AttributeKeyTxHash          = "txHash"
	AttributeValueCategory      = ModuleName
)
