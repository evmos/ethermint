package types

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/params"
)

// GenerateChainConfig returns an Ethereum chainconfig for EVM state transitions
func GenerateChainConfig(chainID *big.Int) *params.ChainConfig {
	// TODO: Update chainconfig to take in parameters for fork blocks
	return &params.ChainConfig{
		ChainID:             chainID,
		HomesteadBlock:      big.NewInt(0),
		DAOForkBlock:        big.NewInt(0),
		DAOForkSupport:      true,
		EIP150Block:         big.NewInt(0),
		EIP150Hash:          common.HexToHash("0x2086799aeebeae135c246c65021c82b4e15a2c451340993aacfd2751886514f0"),
		EIP155Block:         big.NewInt(0),
		EIP158Block:         big.NewInt(0),
		ByzantiumBlock:      big.NewInt(0),
		ConstantinopleBlock: big.NewInt(0),
		PetersburgBlock:     big.NewInt(0),
	}
}
