package types

import (
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"

	"github.com/tharsis/ethermint/types"
)

var _ TxData = &AccessListTx{}

// TxData implements the Ethereum transaction tx structure. It is used
// solely as intended in Ethereum abiding by the protocol.
type TxData interface {
	// TODO: embed ethtypes.TxData. See https://github.com/ethereum/go-ethereum/issues/23154

	TxType() byte
	Copy() TxData
	GetChainID() *big.Int
	GetAccessList() ethtypes.AccessList
	GetData() []byte
	GetNonce() uint64
	GetGas() uint64
	GetGasPrice() *big.Int
	GetValue() *big.Int
	GetTo() *common.Address

	GetRawSignatureValues() (v, r, s *big.Int)
	SetSignatureValues(chainID, v, r, s *big.Int)

	AsEthereumData() ethtypes.TxData
	Validate() error
	Fee() *big.Int
	Cost() *big.Int
}

func NewTxDataFromTx(tx *ethtypes.Transaction) TxData {
	var txData TxData
	switch tx.Type() {
	case ethtypes.AccessListTxType:
		txData = newAccessListTx(tx)
	default:
		txData = newLegacyTx(tx)
	}

	return txData
}

func newAccessListTx(tx *ethtypes.Transaction) *AccessListTx {
	txData := &AccessListTx{
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

	if tx.GasPrice() != nil {
		gasPriceInt := sdk.NewIntFromBigInt(tx.GasPrice())
		txData.GasPrice = &gasPriceInt
	}

	if tx.AccessList() != nil {
		al := tx.AccessList()
		txData.Accesses = NewAccessList(&al)
	}

	txData.SetSignatureValues(tx.ChainId(), v, r, s)
	return txData
}

func newLegacyTx(tx *ethtypes.Transaction) *LegacyTx {
	txData := &LegacyTx{
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

	if tx.GasPrice() != nil {
		gasPriceInt := sdk.NewIntFromBigInt(tx.GasPrice())
		txData.GasPrice = &gasPriceInt
	}

	txData.SetSignatureValues(tx.ChainId(), v, r, s)
	return txData
}

// TxType returns the tx type
func (tx *AccessListTx) TxType() uint8 {
	return ethtypes.AccessListTxType
}

// Copy returns an instance with the same field values
func (tx *AccessListTx) Copy() TxData {
	return &AccessListTx{
		ChainID:  tx.ChainID,
		Nonce:    tx.Nonce,
		GasPrice: tx.GasPrice,
		GasLimit: tx.GasLimit,
		To:       tx.To,
		Amount:   tx.Amount,
		Data:     common.CopyBytes(tx.Data),
		Accesses: tx.Accesses,
		V:        common.CopyBytes(tx.V),
		R:        common.CopyBytes(tx.R),
		S:        common.CopyBytes(tx.S),
	}
}

// GetChainID
func (tx *AccessListTx) GetChainID() *big.Int {
	if tx.ChainID == nil {
		return nil
	}

	return tx.ChainID.BigInt()
}

// GetAccessList
func (tx *AccessListTx) GetAccessList() ethtypes.AccessList {
	if tx.Accesses == nil {
		return nil
	}
	return *tx.Accesses.ToEthAccessList()
}

func (tx *AccessListTx) GetData() []byte {
	return common.CopyBytes(tx.Data)
}

func (tx *AccessListTx) GetGas() uint64 {
	return tx.GasLimit
}

func (tx *AccessListTx) GetGasPrice() *big.Int {
	if tx.GasPrice == nil {
		return nil
	}
	return tx.GasPrice.BigInt()
}

func (tx *AccessListTx) GetValue() *big.Int {
	if tx.Amount == nil {
		return nil
	}

	return tx.Amount.BigInt()
}

func (tx *AccessListTx) GetNonce() uint64 { return tx.Nonce }

func (tx *AccessListTx) GetTo() *common.Address {
	if tx.To == "" {
		return nil
	}
	to := common.HexToAddress(tx.To)
	return &to
}

// AsEthereumData returns an AccessListTx transaction tx from the proto-formatted
// TxData defined on the Cosmos EVM.
func (tx *AccessListTx) AsEthereumData() ethtypes.TxData {
	v, r, s := tx.GetRawSignatureValues()
	return &ethtypes.AccessListTx{
		ChainID:    tx.GetChainID(),
		Nonce:      tx.GetNonce(),
		GasPrice:   tx.GetGasPrice(),
		Gas:        tx.GetGas(),
		To:         tx.GetTo(),
		Value:      tx.GetValue(),
		Data:       tx.GetData(),
		AccessList: tx.GetAccessList(),
		V:          v,
		R:          r,
		S:          s,
	}
}

// GetRawSignatureValues returns the V, R, S signature values of the transaction.
// The return values should not be modified by the caller.
func (tx *AccessListTx) GetRawSignatureValues() (v, r, s *big.Int) {
	return rawSignatureValues(tx.V, tx.R, tx.S)
}

func rawSignatureValues(vBz, rBz, sBz []byte) (v, r, s *big.Int) {
	if len(vBz) > 0 {
		v = new(big.Int).SetBytes(vBz)
	}
	if len(rBz) > 0 {
		r = new(big.Int).SetBytes(rBz)
	}
	if len(sBz) > 0 {
		s = new(big.Int).SetBytes(sBz)
	}
	return v, r, s
}

func (tx *AccessListTx) SetSignatureValues(chainID, v, r, s *big.Int) {
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

// Validate performs a basic validation of the tx tx fields.
func (tx AccessListTx) Validate() error {
	gasPrice := tx.GetGasPrice()
	if gasPrice == nil {
		return sdkerrors.Wrap(ErrInvalidGasPrice, "cannot be nil")
	}

	if gasPrice.Sign() == -1 {
		return sdkerrors.Wrapf(ErrInvalidGasPrice, "gas price cannot be negative %s", gasPrice)
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

// Fee returns gasprice * gaslimit.
func (tx AccessListTx) Fee() *big.Int {
	return fee(tx.GetGasPrice(), tx.GetGas())
}

// Cost returns amount + gasprice * gaslimit.
func (tx AccessListTx) Cost() *big.Int {
	return cost(tx.Fee(), tx.GetValue())
}

// TxType returns the tx type
func (tx *LegacyTx) TxType() uint8 {
	return ethtypes.LegacyTxType
}

// Copy returns an instance with the same field values
func (tx *LegacyTx) Copy() TxData {
	return &LegacyTx{
		Nonce:    tx.Nonce,
		GasPrice: tx.GasPrice,
		GasLimit: tx.GasLimit,
		To:       tx.To,
		Amount:   tx.Amount,
		Data:     common.CopyBytes(tx.Data),
		V:        common.CopyBytes(tx.V),
		R:        common.CopyBytes(tx.R),
		S:        common.CopyBytes(tx.S),
	}
}

// GetChainID
func (tx *LegacyTx) GetChainID() *big.Int {
	v, _, _ := tx.GetRawSignatureValues()
	return DeriveChainID(v)
}

// GetAccessList
func (tx *LegacyTx) GetAccessList() ethtypes.AccessList {
	return nil
}

func (tx *LegacyTx) GetData() []byte {
	return common.CopyBytes(tx.Data)
}

func (tx *LegacyTx) GetGas() uint64 {
	return tx.GasLimit
}

func (tx *LegacyTx) GetGasPrice() *big.Int {
	if tx.GasPrice == nil {
		return nil
	}
	return tx.GasPrice.BigInt()
}

func (tx *LegacyTx) GetValue() *big.Int {
	if tx.Amount == nil {
		return nil
	}
	return tx.Amount.BigInt()
}

func (tx *LegacyTx) GetNonce() uint64 { return tx.Nonce }

func (tx *LegacyTx) GetTo() *common.Address {
	if tx.To == "" {
		return nil
	}
	to := common.HexToAddress(tx.To)
	return &to
}

// AsEthereumData returns an AccessListTx transaction tx from the proto-formatted
// TxData defined on the Cosmos EVM.
func (tx *LegacyTx) AsEthereumData() ethtypes.TxData {
	v, r, s := tx.GetRawSignatureValues()
	return &ethtypes.LegacyTx{
		Nonce:    tx.GetNonce(),
		GasPrice: tx.GetGasPrice(),
		Gas:      tx.GetGas(),
		To:       tx.GetTo(),
		Value:    tx.GetValue(),
		Data:     tx.GetData(),
		V:        v,
		R:        r,
		S:        s,
	}
}

// GetRawSignatureValues returns the V, R, S signature values of the transaction.
// The return values should not be modified by the caller.
func (tx *LegacyTx) GetRawSignatureValues() (v, r, s *big.Int) {
	return rawSignatureValues(tx.V, tx.R, tx.S)
}

func (tx *LegacyTx) SetSignatureValues(_, v, r, s *big.Int) {
	if v != nil {
		tx.V = v.Bytes()
	}
	if r != nil {
		tx.R = r.Bytes()
	}
	if s != nil {
		tx.S = s.Bytes()
	}
}

// Validate performs a basic validation of the tx tx fields.
func (tx LegacyTx) Validate() error {
	gasPrice := tx.GetGasPrice()
	if gasPrice == nil {
		return sdkerrors.Wrap(ErrInvalidGasPrice, "cannot be nil")
	}

	if gasPrice.Sign() == -1 {
		return sdkerrors.Wrapf(ErrInvalidGasPrice, "gas price cannot be negative %s", gasPrice)
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

	return nil
}

// Fee returns gasprice * gaslimit.
func (tx LegacyTx) Fee() *big.Int {
	return fee(tx.GetGasPrice(), tx.GetGas())
}

// Cost returns amount + gasprice * gaslimit.
func (tx LegacyTx) Cost() *big.Int {
	return cost(tx.Fee(), tx.GetValue())
}

// DeriveChainID derives the chain id from the given v parameter
func DeriveChainID(v *big.Int) *big.Int {
	if v == nil {
		return nil
	}

	if v.BitLen() <= 64 {
		v := v.Uint64()
		if v == 27 || v == 28 {
			return new(big.Int)
		}
		return new(big.Int).SetUint64((v - 35) / 2)
	}
	v = new(big.Int).Sub(v, big.NewInt(35))
	return v.Div(v, big.NewInt(2))
}

func fee(gasPrice *big.Int, gas uint64) *big.Int {
	gasLimit := new(big.Int).SetUint64(gas)
	return new(big.Int).Mul(gasPrice, gasLimit)
}

func cost(fee, value *big.Int) *big.Int {
	if value != nil {
		return new(big.Int).Add(fee, value)
	}
	return fee
}
