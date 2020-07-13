package types

import (
	"bytes"
	"fmt"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	ethcmn "github.com/ethereum/go-ethereum/common"
)

// Storage represents the account Storage map as a slice of single key value
// State pairs. This is to prevent non determinism at genesis initialization or export.
type Storage []State

// Validate performs a basic validation of the Storage fields.
func (s Storage) Validate() error {
	seenStorage := make(map[string]bool)
	for i, state := range s {
		if seenStorage[state.Key.String()] {
			return sdkerrors.Wrapf(ErrInvalidState, "duplicate state key %d", i)
		}

		if err := state.Validate(); err != nil {
			return err
		}

		seenStorage[state.Key.String()] = true
	}
	return nil
}

// String implements the stringer interface
func (s Storage) String() string {
	var str string
	for _, state := range s {
		str += fmt.Sprintf("%s: %s\n", state.Key.String(), state.Value.String())
	}

	return str
}

// Copy returns a copy of storage.
func (s Storage) Copy() Storage {
	cpy := make(Storage, len(s))
	copy(cpy, s)

	return cpy
}

// State represents a single Storage key value pair item.
type State struct {
	Key   ethcmn.Hash `json:"key"`
	Value ethcmn.Hash `json:"value"`
}

// Validate performs a basic validation of the State fields.
func (s State) Validate() error {
	if bytes.Equal(s.Key.Bytes(), ethcmn.Hash{}.Bytes()) {
		return sdkerrors.Wrap(ErrInvalidState, "state key hash cannot be empty")
	}
	// NOTE: state value can be empty
	return nil
}

// NewState creates a new State instance
func NewState(key, value ethcmn.Hash) State {
	return State{
		Key:   key,
		Value: value,
	}
}
