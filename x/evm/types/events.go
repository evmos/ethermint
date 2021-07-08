package types

// Evm module events
const (
	EventTypeEthereumTx = TypeMsgEthereumTx

	AttributeKeyContractAddress    = "contract"
	AttributeKeyRecipient          = "recipient"
	AttributeKeyTxHash             = "txHash"
	AttributeKeyEthereumTxHash     = "ethereumTxHash"
	AttributeKeyEthereumTxReverted = "ethereumTxReverted"
	AttributeValueCategory         = ModuleName

	MetricKeyTransitionDB = "transition_db"
	MetricKeyStaticCall   = "static_call"
)
