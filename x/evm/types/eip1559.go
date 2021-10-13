package types

import (
	"math/big"

	"github.com/ethereum/go-ethereum/params"
)

func IsLondon(ethConfig *params.ChainConfig, height int64) bool {
	rules := ethConfig.Rules(big.NewInt(height))
	return rules.IsLondon
}
