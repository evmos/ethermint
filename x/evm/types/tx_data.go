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
}

func newTxData(
	chainID *big.Int, nonce uint64, to *common.Address, amount *big.Int,
	gasLimit uint64, gasPrice *big.Int, input []byte, accesses *ethtypes.AccessList,
) TxData {
	txData := &AccessListTx{
		Nonce:    nonce,
		GasLimit: gasLimit,
	}

	txData.Data = common.CopyBytes(input)

	if to != nil {
		txData.To = to.Hex()
	}

	if accesses != nil {
		txData.Accesses = NewAccessList(accesses)

		// NOTE: we don't populate chain id on LegacyTx type
		if chainID != nil {
			txData.ChainID = sdk.NewIntFromBigInt(chainID)
		}
	}

	if amount != nil {
		txData.Amount = sdk.NewIntFromBigInt(amount)
	}
	if gasPrice != nil {
		txData.GasPrice = sdk.NewIntFromBigInt(gasPrice)
	}
	return txData
}

// TxType returns the tx type
func (tx *AccessListTx) TxType() uint8 {
	if tx.Accesses == nil {
		return ethtypes.LegacyTxType
	}
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
		V:        common.CopyBytes(tx.V),
		R:        common.CopyBytes(tx.R),
		S:        common.CopyBytes(tx.S),
	}
}

// GetChainID
func (tx *AccessListTx) GetChainID() *big.Int {
	if tx.TxType() == ethtypes.LegacyTxType {
		v, _, _ := tx.GetRawSignatureValues()
		return DeriveChainID(v)
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
	return tx.GasPrice.BigInt()
}

func (tx *AccessListTx) GetValue() *big.Int {
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
	if tx.Accesses == nil {
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
	if len(tx.V) > 0 {
		v = new(big.Int).SetBytes(tx.V)
	}
	if len(tx.R) > 0 {
		r = new(big.Int).SetBytes(tx.R)
	}
	if len(tx.S) > 0 {
		s = new(big.Int).SetBytes(tx.S)
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
	if tx.TxType() == ethtypes.AccessListTxType && chainID != nil {
		tx.ChainID = sdk.NewIntFromBigInt(chainID)
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

	if tx.TxType() == ethtypes.AccessListTxType && tx.GetChainID() == nil {
		return sdkerrors.Wrap(
			sdkerrors.ErrInvalidChainID,
			"chain ID must be present on AccessList txs",
		)
	}

	return nil
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
