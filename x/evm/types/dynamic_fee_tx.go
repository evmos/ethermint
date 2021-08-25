package types

import (
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/tharsis/ethermint/types"
)

// TODO: remove once migrated to latest go-ethereum
const (
	BaseFeeChangeDenominator = 8          // Bounds the amount the base fee can change between blocks.
	ElasticityMultiplier     = 2          // Bounds the maximum gas limit an EIP-1559 block may have.
	InitialBaseFee           = 1000000000 // Initial base fee for EIP-1559 blocks.
)

// nolint: deadcode, unused
func newDynamicFeeTx(tx *ethtypes.Transaction) *DynamicFeeTx {
	txData := &DynamicFeeTx{
		Nonce:    tx.Nonce(),
		Data:     tx.Data(),
		GasLimit: tx.Gas(),
	}

	v, r, s := tx.RawSignatureValues()
	if tx.To() != nil {
		txData.To = tx.To().Hex()
	}

	if tx.Value() != nil {
		amountInt := sdk.NewIntFromBigInt(tx.Value())
		txData.Amount = &amountInt
	}
	// TODO:

	// if tx.GasFeeCap() != nil {
	// 	gasFeeCapInt := sdk.NewIntFromBigInt(tx.GasFeeCap())
	// 	txData.GasFeeCap = &gasFeeCapInt
	// }

	// if tx.GasTipCap() != nil {
	// 	gasTipCapInt := sdk.NewIntFromBigInt(tx.GasTipCap())
	// 	txData.GasTipCap = &gasTipCapInt
	// }

	if tx.AccessList() != nil {
		al := tx.AccessList()
		txData.Accesses = NewAccessList(&al)
	}

	txData.SetSignatureValues(tx.ChainId(), v, r, s)
	return txData
}

// TxType returns the tx type
func (tx *DynamicFeeTx) TxType() uint8 {
	// TODO
	return 0
	// return ethtypes.DynamicFeeTxType
}

// Copy returns an instance with the same field values
func (tx *DynamicFeeTx) Copy() TxData {
	return &DynamicFeeTx{
		ChainID:   tx.ChainID,
		Nonce:     tx.Nonce,
		GasTipCap: tx.GasTipCap,
		GasFeeCap: tx.GasFeeCap,
		GasLimit:  tx.GasLimit,
		To:        tx.To,
		Amount:    tx.Amount,
		Data:      common.CopyBytes(tx.Data),
		Accesses:  tx.Accesses,
		V:         common.CopyBytes(tx.V),
		R:         common.CopyBytes(tx.R),
		S:         common.CopyBytes(tx.S),
	}
}

// GetChainID returns the chain id field from the DynamicFeeTx
func (tx *DynamicFeeTx) GetChainID() *big.Int {
	if tx.ChainID == nil {
		return nil
	}

	return tx.ChainID.BigInt()
}

// GetAccessList returns the AccessList field.
func (tx *DynamicFeeTx) GetAccessList() ethtypes.AccessList {
	if tx.Accesses == nil {
		return nil
	}
	return *tx.Accesses.ToEthAccessList()
}

// GetData returns the a copy of the input data bytes.
func (tx *DynamicFeeTx) GetData() []byte {
	return common.CopyBytes(tx.Data)
}

// GetGas returns the gas limit.
func (tx *DynamicFeeTx) GetGas() uint64 {
	return tx.GasLimit
}

// GetGasPrice returns the gas fee cap field.
func (tx *DynamicFeeTx) GetGasPrice() *big.Int {
	return tx.GetGasFeeCap()
}

// GetGasTipCap returns the gas price field.
func (tx *DynamicFeeTx) GetGasTipCap() *big.Int {
	if tx.GasTipCap == nil {
		return nil
	}
	return tx.GasTipCap.BigInt()
}

// GetGasFeeCap returns the gas price field.
func (tx *DynamicFeeTx) GetGasFeeCap() *big.Int {
	if tx.GasFeeCap == nil {
		return nil
	}
	return tx.GasFeeCap.BigInt()
}

// GetValue returns the tx amount.
func (tx *DynamicFeeTx) GetValue() *big.Int {
	if tx.Amount == nil {
		return nil
	}

	return tx.Amount.BigInt()
}

// GetNonce returns the account sequence for the transaction.
func (tx *DynamicFeeTx) GetNonce() uint64 { return tx.Nonce }

// GetTo returns the pointer to the recipient address.
func (tx *DynamicFeeTx) GetTo() *common.Address {
	if tx.To == "" {
		return nil
	}
	to := common.HexToAddress(tx.To)
	return &to
}

// AsEthereumData returns an DynamicFeeTx transaction tx from the proto-formatted
// TxData defined on the Cosmos EVM.
func (tx *DynamicFeeTx) AsEthereumData() ethtypes.TxData {
	return nil
	// TODO:
	// v, r, s := tx.GetRawSignatureValues()
	// return &ethtypes.DynamicFeeTx{
	// 	ChainID:    tx.GetChainID(),
	// 	Nonce:      tx.GetNonce(),
	// 	GasTipCap:  tx.GetGasTipCap(),
	// 	GasFeeCap:  tx.GetGasFeeCap(),
	// 	Gas:        tx.GetGas(),
	// 	To:         tx.GetTo(),
	// 	Value:      tx.GetValue(),
	// 	Data:       tx.GetData(),
	// 	AccessList: tx.GetAccessList(),
	// 	V:          v,
	// 	R:          r,
	// 	S:          s,
	// }
}

// GetRawSignatureValues returns the V, R, S signature values of the transaction.
// The return values should not be modified by the caller.
func (tx *DynamicFeeTx) GetRawSignatureValues() (v, r, s *big.Int) {
	return rawSignatureValues(tx.V, tx.R, tx.S)
}

// SetSignatureValues sets the signature values to the transaction.
func (tx *DynamicFeeTx) SetSignatureValues(chainID, v, r, s *big.Int) {
	if v != nil {
		tx.V = v.Bytes()
	}
	if r != nil {
		tx.R = r.Bytes()
	}
	if s != nil {
		tx.S = s.Bytes()
	}
	if chainID != nil {
		chainIDInt := sdk.NewIntFromBigInt(chainID)
		tx.ChainID = &chainIDInt
	}
}

// Validate performs a stateless validation of the tx fields.
func (tx DynamicFeeTx) Validate() error {
	if tx.GasTipCap == nil {
		return sdkerrors.Wrap(ErrInvalidGasCap, "gas tip cap cannot nil")
	}

	if tx.GasFeeCap == nil {
		return sdkerrors.Wrap(ErrInvalidGasCap, "gas fee cap cannot nil")
	}

	if tx.GasTipCap.IsNegative() {
		return sdkerrors.Wrapf(ErrInvalidGasCap, "gas tip cap cannot be negative %s", tx.GasTipCap)
	}

	if tx.GasFeeCap.IsNegative() {
		return sdkerrors.Wrapf(ErrInvalidGasCap, "gas fee cap cannot be negative %s", tx.GasFeeCap)
	}

	if tx.GasFeeCap.LT(*tx.GasTipCap) {
		return sdkerrors.Wrapf(
			ErrInvalidGasCap, "max priority fee per gas higher than max fee per gas (%s > %s)",
			tx.GasTipCap, tx.GasFeeCap,
		)
	}

	amount := tx.GetValue()
	// Amount can be 0
	if amount != nil && amount.Sign() == -1 {
		return sdkerrors.Wrapf(ErrInvalidAmount, "amount cannot be negative %s", amount)
	}

	if tx.To != "" {
		if err := types.ValidateAddress(tx.To); err != nil {
			return sdkerrors.Wrap(err, "invalid to address")
		}
	}

	if tx.GetChainID() == nil {
		return sdkerrors.Wrap(
			sdkerrors.ErrInvalidChainID,
			"chain ID must be present on AccessList txs",
		)
	}

	return nil
}

// Fee panics as it requires the base fee amount to calculate
func (tx DynamicFeeTx) Fee() *big.Int {
	panic("fee can only be called manually by providing the base fee")
}

// Cost returns amount + gasprice * gaslimit.
func (tx DynamicFeeTx) Cost() *big.Int {
	panic("cost can only be called manually by providing the base fee")
}
