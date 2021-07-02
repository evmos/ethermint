package types

import (
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"

	"github.com/tharsis/ethermint/types"
)

var _ TxDataI = &TxData{}

type TxDataI interface {
	TxType() byte
	Copy() TxDataI
	GetChainID() *big.Int
	GetAccessList() ethtypes.AccessList
	GetData() []byte
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
) *TxData {
	txData := &TxData{
		Nonce:    nonce,
		GasLimit: gasLimit,
	}

	txData.Input = common.CopyBytes(input)

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
func (data *TxData) TxType() uint8 {
	if data.Accesses == nil {
		return ethtypes.LegacyTxType
	}
	return ethtypes.AccessListTxType
}

// Copy returns an instance with the same field values
func (data *TxData) Copy() TxDataI {
	return &TxData{
		ChainID:  data.ChainID,
		Nonce:    data.Nonce,
		GasPrice: data.GasPrice,
		GasLimit: data.GasLimit,
		To:       data.To,
		Amount:   data.Amount,
		Input:    common.CopyBytes(data.Input),
		V:        common.CopyBytes(data.V),
		R:        common.CopyBytes(data.R),
		S:        common.CopyBytes(data.S),
	}
}

// GetChainID
func (data *TxData) GetChainID() *big.Int {
	if data.TxType() == ethtypes.LegacyTxType {
		v, _, _ := data.GetRawSignatureValues()
		return DeriveChainID(v)
	}

	return data.ChainID.BigInt()
}

// GetAccessList
func (data *TxData) GetAccessList() ethtypes.AccessList {
	if data.Accesses == nil {
		return nil
	}
	return *data.Accesses.ToEthAccessList()
}

func (data *TxData) GetData() []byte {
	return common.CopyBytes(data.Input)
}

func (data *TxData) GetGas() uint64 {
	return data.GasLimit
}

func (data *TxData) GetGasPrice() *big.Int {
	return data.GasPrice.BigInt()
}

func (data *TxData) GetValue() *big.Int {
	return data.Amount.BigInt()
}

func (data *TxData) GetNonce() uint64 { return data.Nonce }

func (data *TxData) GetTo() *common.Address {
	if data.To == "" {
		return nil
	}
	to := common.HexToAddress(data.To)
	return &to
}

// AsEthereumData returns an AccessListTx transaction data from the proto-formatted
// TxData defined on the Cosmos EVM.
func (data *TxData) AsEthereumData() ethtypes.TxData {
	v, r, s := data.GetRawSignatureValues()
	if data.Accesses == nil {
		return &ethtypes.LegacyTx{
			Nonce:    data.GetNonce(),
			GasPrice: data.GetGasPrice(),
			Gas:      data.GetGas(),
			To:       data.GetTo(),
			Value:    data.GetValue(),
			Data:     data.GetData(),
			V:        v,
			R:        r,
			S:        s,
		}
	}

	return &ethtypes.AccessListTx{
		ChainID:    data.GetChainID(),
		Nonce:      data.GetNonce(),
		GasPrice:   data.GetGasPrice(),
		Gas:        data.GetGas(),
		To:         data.GetTo(),
		Value:      data.GetValue(),
		Data:       data.GetData(),
		AccessList: data.GetAccessList(),
		V:          v,
		R:          r,
		S:          s,
	}
}

// GetRawSignatureValues returns the V, R, S signature values of the transaction.
// The return values should not be modified by the caller.
func (data *TxData) GetRawSignatureValues() (v, r, s *big.Int) {
	if len(data.V) > 0 {
		v = new(big.Int).SetBytes(data.V)
	}
	if len(data.R) > 0 {
		r = new(big.Int).SetBytes(data.R)
	}
	if len(data.S) > 0 {
		s = new(big.Int).SetBytes(data.S)
	}
	return v, r, s
}

func (data *TxData) SetSignatureValues(chainID, v, r, s *big.Int) {
	if v != nil {
		data.V = v.Bytes()
	}
	if r != nil {
		data.R = r.Bytes()
	}
	if s != nil {
		data.S = s.Bytes()
	}
	if data.TxType() == ethtypes.AccessListTxType && chainID != nil {
		data.ChainID = sdk.NewIntFromBigInt(chainID)
	}
}

// Validate performs a basic validation of the tx data fields.
func (data TxData) Validate() error {
	gasPrice := data.GetGasPrice()
	if gasPrice == nil {
		return sdkerrors.Wrap(ErrInvalidGasPrice, "cannot be nil")
	}

	if gasPrice.Sign() == -1 {
		return sdkerrors.Wrapf(ErrInvalidGasPrice, "gas price cannot be negative %s", gasPrice)
	}

	amount := data.GetValue()
	// Amount can be 0
	if amount != nil && amount.Sign() == -1 {
		return sdkerrors.Wrapf(ErrInvalidAmount, "amount cannot be negative %s", amount)
	}

	if data.To != "" {
		if err := types.ValidateAddress(data.To); err != nil {
			return sdkerrors.Wrap(err, "invalid to address")
		}
	}

	if data.TxType() == ethtypes.AccessListTxType && data.GetChainID() == nil {
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
