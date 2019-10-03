package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/ethermint/types"
	ethcmn "github.com/ethereum/go-ethereum/common"
)

var (
	_ sdk.Msg = EmintMsg{}
)

const (
	// TypeEmintMsg defines the type string of Emint message
	TypeEmintMsg = "emint_tx"
)

// EmintMsg implements a cosmos equivalent structure for Ethereum transactions
type EmintMsg struct {
	AccountNonce uint64          `json:"nonce"`
	Price        sdk.Int         `json:"gasPrice"`
	GasLimit     uint64          `json:"gas"`
	Recipient    *sdk.AccAddress `json:"to" rlp:"nil"` // nil means contract creation
	Amount       sdk.Int         `json:"value"`
	Payload      []byte          `json:"input"`

	// From address (formerly derived from signature)
	From sdk.AccAddress `json:"from"`
}

// NewEmintMsg returns a reference to a new Ethermint transaction
func NewEmintMsg(
	nonce uint64, to *sdk.AccAddress, amount sdk.Int,
	gasLimit uint64, gasPrice sdk.Int, payload []byte, from sdk.AccAddress,
) EmintMsg {
	return EmintMsg{
		AccountNonce: nonce,
		Price:        gasPrice,
		GasLimit:     gasLimit,
		Recipient:    to,
		Amount:       amount,
		Payload:      payload,
		From:         from,
	}
}

// Route should return the name of the module
func (msg EmintMsg) Route() string { return RouterKey }

// Type returns the action of the message
func (msg EmintMsg) Type() string { return TypeEmintMsg }

// GetSignBytes encodes the message for signing
func (msg EmintMsg) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

// ValidateBasic runs stateless checks on the message
func (msg EmintMsg) ValidateBasic() sdk.Error {
	if msg.Price.Sign() != 1 {
		return types.ErrInvalidValue(fmt.Sprintf("Price must be positive: %x", msg.Price))
	}

	// Amount can be 0
	if msg.Amount.Sign() == -1 {
		return types.ErrInvalidValue(fmt.Sprintf("amount cannot be negative: %x", msg.Amount))
	}

	return nil
}

// GetSigners defines whose signature is required
func (msg EmintMsg) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.From}
}

// To returns the recipient address of the transaction. It returns nil if the
// transaction is a contract creation.
func (msg EmintMsg) To() *ethcmn.Address {
	if msg.Recipient == nil {
		return nil
	}

	addr := ethcmn.BytesToAddress(msg.Recipient.Bytes())
	return &addr
}
