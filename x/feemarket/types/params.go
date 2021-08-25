package types

const (
	BaseFeeChangeDenominator = 8          // Bounds the amount the base fee can change between blocks.
	ElasticityMultiplier     = 2          // Bounds the maximum gas limit an EIP-1559 block may have.
	InitialBaseFee           = 1000000000 // Initial base fee for EIP-1559 blocks.
)