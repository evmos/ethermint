package types

import (
	"github.com/ethereum/go-ethereum/params"
	"math/big"

)

func IsLondon(ethConfig *params.ChainConfig, height int64) bool {
	rules := ethConfig.Rules(big.NewInt(height))
	return rules.IsLondon
}
