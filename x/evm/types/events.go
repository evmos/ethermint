package types

// Evm module events
const (
	EventTypeEthereumTx = TypeMsgEthereumTx

	AttributeKeyContractAddress = "contract"
	AttributeKeyRecipient       = "recipient"
	AttributeKeyTxHash          = "txHash"
	AttributeKeyEthereumTxHash  = "ethereumTxHash"
	AttributeKeyTxType          = "txType"
	// tx failed in eth vm execution
	AttributeKeyEthereumTxFailed = "ethereumTxFailed"
	AttributeValueCategory       = ModuleName

	MetricKeyTransitionDB = "transition_db"
	MetricKeyStaticCall   = "static_call"
)
