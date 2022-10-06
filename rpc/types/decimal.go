package types

import (
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum/common/hexutil"
)

// DecimalOrHex unmarshals a non-negative decimal or hex parameter into a uint64.
type DecimalOrHex uint64

// UnmarshalJSON implements json.Unmarshaler.
func (dh *DecimalOrHex) UnmarshalJSON(data []byte) error {
	input := strings.TrimSpace(string(data))
	if len(input) >= 2 && input[0] == '"' && input[len(input)-1] == '"' {
		input = input[1 : len(input)-1]
	}

	value, err := strconv.ParseUint(input, 10, 64)
	if err != nil {
		value, err = hexutil.DecodeUint64(input)
	}
	if err != nil {
		return err
	}
	*dh = DecimalOrHex(value)
	return nil
}
