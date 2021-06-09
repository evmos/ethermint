package types

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
)

// var _ ethtypes.TxData = &TxData{}

func (data *TxData) txType() byte {
	if data.Accesses == nil {
		return ethtypes.LegacyTxType
	}
	return ethtypes.AccessListTxType
}

func (data *TxData) chainID() *big.Int {
	if data.txType() == ethtypes.LegacyTxType {
		v, _, _ := data.rawSignatureValues()
		return deriveChainId(v)
	}

	if data.ChainID == nil {
		return nil
	}

	return new(big.Int).SetBytes(data.ChainID)
}

func (data *TxData) accessList() ethtypes.AccessList {
	if data.Accesses == nil {
		return nil
	}
	return *data.Accesses.ToEthAccessList()
}

func (data *TxData) data() []byte {
	return common.CopyBytes(data.Input)
}

func (data *TxData) gas() uint64 {
	return data.GasLimit
}

func (data *TxData) gasPrice() *big.Int {
	if data.GasPrice == nil {
		return nil
	}
	return new(big.Int).SetBytes(data.GasPrice)
}

func (data *TxData) amount() *big.Int {
	if data.Amount == nil {
		return nil
	}
	return new(big.Int).SetBytes(data.Amount)
}

func (data *TxData) nonce() uint64 { return data.Nonce }

func (data *TxData) to() *common.Address {
	if data.To == "" {
		return nil
	}
	to := common.HexToAddress(data.To)
	return &to
}

// AsEthereumData returns an AccessListTx transaction data from the proto-formatted
// TxData defined on the Cosmos EVM.
func (data *TxData) AsEthereumData() ethtypes.TxData {
	v, r, s := data.rawSignatureValues()
	if data.Accesses == nil {
		return &ethtypes.LegacyTx{
			Nonce:    data.nonce(),
			GasPrice: data.gasPrice(),
			Gas:      data.gas(),
			To:       data.to(),
			Value:    data.amount(),
			Data:     data.data(),
			V:        v,
			R:        r,
			S:        s,
		}
	}

	return &ethtypes.AccessListTx{
		ChainID:    data.chainID(),
		Nonce:      data.nonce(),
		GasPrice:   data.gasPrice(),
		Gas:        data.gas(),
		To:         data.to(),
		Value:      data.amount(),
		Data:       data.data(),
		AccessList: data.accessList(),
		V:          v,
		R:          r,
		S:          s,
	}
}

// rawSignatureValues returns the V, R, S signature values of the transaction.
// The return values should not be modified by the caller.
func (data *TxData) rawSignatureValues() (v, r, s *big.Int) {
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

func (data *TxData) setSignatureValues(chainID, v, r, s *big.Int) {
	if v != nil {
		data.V = v.Bytes()
	}
	if r != nil {
		data.R = r.Bytes()
	}
	if s != nil {
		data.S = s.Bytes()
	}
	if data.txType() == ethtypes.AccessListTxType && chainID != nil {
		data.ChainID = chainID.Bytes()
	}
}

// deriveChainId derives the chain id from the given v parameter
func deriveChainId(v *big.Int) *big.Int {
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
