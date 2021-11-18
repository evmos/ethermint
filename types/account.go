package types

import (
	"encoding/json"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

var (
	_ authtypes.AccountI                 = (*EthAccount)(nil)
	_ authtypes.GenesisAccount           = (*EthAccount)(nil)
	_ codectypes.UnpackInterfacesMessage = (*EthAccount)(nil)
)

// ----------------------------------------------------------------------------
// Main Ethermint account
// ----------------------------------------------------------------------------

// ProtoAccount defines the prototype function for BaseAccount used for an
// AccountKeeper.
func ProtoAccount() authtypes.AccountI {
	return &EthAccount{
		BaseAccount: &authtypes.BaseAccount{},
		CodeHash:    common.BytesToHash(crypto.Keccak256(nil)).String(),
	}
}

// EthAddress returns the account address ethereum format.
func (acc EthAccount) EthAddress() common.Address {
	return common.BytesToAddress(acc.GetAddress().Bytes())
}

// GetCodeHash returns the account code hash in byte format
func (acc EthAccount) GetCodeHash() common.Hash {
	return common.HexToHash(acc.CodeHash)
}

// MarshalJSON returns the JSON representation of a ModuleAccount.
func (acc EthAccount) MarshalJSON() ([]byte, error) {
	accAddr, err := sdk.AccAddressFromBech32(acc.Address)
	if err != nil {
		return nil, err
	}

	pubkey, ok := acc.PubKey.GetCachedValue().(cryptotypes.PubKey)
	if !ok {
		return []byte{}, nil
	}

	formatedKey := pubKeyPretty{
		Type:  acc.PubKey.TypeUrl,
		Value: pubkey.Bytes(),
	}

	return json.Marshal(ethAccountPretty{
		Address:       accAddr,
		PubKey:        formatedKey,
		AccountNumber: acc.AccountNumber,
		Sequence:      acc.Sequence,
		CodeHash:      acc.CodeHash,
	})
}

type ethAccountPretty struct {
	Address       sdk.AccAddress `json:"address" yaml:"address"`
	PubKey        pubKeyPretty   `json:"public_key" yaml:"public_key"`
	AccountNumber uint64         `json:"account_number" yaml:"account_number"`
	Sequence      uint64         `json:"sequence" yaml:"sequence"`
	CodeHash      string         `json:"code_hash" yaml:"code_hash"`
}

type pubKeyPretty struct {
	Type  string `json:"type" yaml:"type"`
	Value []byte `json:"value" yaml:"value"`
}
