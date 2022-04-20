package simulation

import (
	"bytes"
	"fmt"

	"github.com/cosmos/cosmos-sdk/types/kv"
	"github.com/ethereum/go-ethereum/common"
	"github.com/tharsis/ethermint/x/evm/types"
)

// NewDecodeStore returns a decoder function closure that unmarshals the KVPair's
// Value to the corresponding evm type.
func NewDecodeStore() func(kvA, kvB kv.Pair) string {
	return func(kvA, kvB kv.Pair) string {
		switch {
		case bytes.Equal(kvA.Key[:1], types.KeyPrefixStorage):
			storageHashA := common.BytesToHash(kvA.Value).Hex()
			storageHashB := common.BytesToHash(kvB.Value).Hex()

			return fmt.Sprintf("%v\n%v", storageHashA, storageHashB)
		case bytes.Equal(kvA.Key[:1], types.KeyPrefixCode):
			codeHashA := common.BytesToHash(kvA.Value).Hex()
			codeHashB := common.BytesToHash(kvB.Value).Hex()

			return fmt.Sprintf("%v\n%v", codeHashA, codeHashB)
		default:
			panic(fmt.Sprintf("invalid evm key prefix %X", kvA.Key[:1]))
		}
	}
}
