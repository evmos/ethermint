package types

import (
	"bytes"
	"encoding/json"

	yaml "gopkg.in/yaml.v2"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	ethcmn "github.com/ethereum/go-ethereum/common"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
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
		CodeHash:    ethcrypto.Keccak256(nil),
	}
}

// EthAddress returns the account address ethereum format.
func (acc EthAccount) EthAddress() ethcmn.Address {
	return ethcmn.BytesToAddress(acc.GetAddress().Bytes())
}

type ethermintAccountPretty struct {
	Address       string `json:"address" yaml:"address"`
	EthAddress    string `json:"eth_address" yaml:"eth_address"`
	PubKey        string `json:"public_key" yaml:"public_key"`
	AccountNumber uint64 `json:"account_number" yaml:"account_number"`
	Sequence      uint64 `json:"sequence" yaml:"sequence"`
	CodeHash      string `json:"code_hash" yaml:"code_hash"`
}

// MarshalYAML returns the YAML representation of an account.
func (acc EthAccount) MarshalYAML() (interface{}, error) {
	alias := ethermintAccountPretty{
		Address:       acc.Address,
		EthAddress:    acc.EthAddress().String(),
		AccountNumber: acc.AccountNumber,
		Sequence:      acc.Sequence,
		CodeHash:      ethcmn.Bytes2Hex(acc.CodeHash),
	}

	var err error

	if acc.PubKey != nil {
		alias.PubKey, err = sdk.Bech32ifyPubKey(sdk.Bech32PubKeyTypeAccPub, acc.GetPubKey())
		if err != nil {
			return nil, err
		}
	}

	bz, err := yaml.Marshal(alias)
	if err != nil {
		return nil, err
	}

	return string(bz), err
}

// MarshalJSON returns the JSON representation of an EthAccount.
func (acc EthAccount) MarshalJSON() ([]byte, error) {
	var ethAddress = ""

	if acc.BaseAccount != nil && acc.Address != "" {
		ethAddress = acc.EthAddress().String()
	}

	alias := ethermintAccountPretty{
		Address:       acc.Address,
		EthAddress:    ethAddress,
		AccountNumber: acc.AccountNumber,
		Sequence:      acc.Sequence,
		CodeHash:      ethcmn.Bytes2Hex(acc.CodeHash),
	}

	var err error

	if acc.PubKey != nil {
		alias.PubKey, err = sdk.Bech32ifyPubKey(sdk.Bech32PubKeyTypeAccPub, acc.GetPubKey())
		if err != nil {
			return nil, err
		}
	}

	return json.Marshal(alias)
}

// UnmarshalJSON unmarshals raw JSON bytes into an EthAccount.
func (acc *EthAccount) UnmarshalJSON(bz []byte) error {
	var (
		alias ethermintAccountPretty
		err   error
	)

	if err := json.Unmarshal(bz, &alias); err != nil {
		return err
	}

	switch {
	case alias.Address != "" && alias.EthAddress != "":
		// Both addresses provided. Verify correctness
		ethAddress := ethcmn.HexToAddress(alias.EthAddress)

		var address sdk.AccAddress
		address, err = sdk.AccAddressFromBech32(alias.Address)
		if err != nil {
			return err
		}

		ethAddressFromAccAddress := ethcmn.BytesToAddress(address.Bytes())

		if !bytes.Equal(ethAddress.Bytes(), address.Bytes()) {
			err = sdkerrors.Wrapf(
				sdkerrors.ErrInvalidAddress,
				"expected %s, got %s",
				ethAddressFromAccAddress.String(), ethAddress.String(),
			)
		}

	case alias.Address != "" && alias.EthAddress == "":
		// unmarshal sdk.AccAddress only. Do nothing here
	case alias.Address == "" && alias.EthAddress != "":
		// retrieve sdk.AccAddress from ethereum address
		ethAddress := ethcmn.HexToAddress(alias.EthAddress)
		alias.Address = sdk.AccAddress(ethAddress.Bytes()).String()
	case alias.Address == "" && alias.EthAddress == "":
		err = sdkerrors.Wrapf(
			sdkerrors.ErrInvalidAddress,
			"account must contain address in Ethereum Hex or Cosmos Bech32 format",
		)
	}

	if err != nil {
		return err
	}

	acc.BaseAccount = &authtypes.BaseAccount{
		Address:       alias.Address,
		AccountNumber: alias.AccountNumber,
		Sequence:      alias.Sequence,
	}

	if alias.PubKey != "" {
		pubkey, err := sdk.GetPubKeyFromBech32(sdk.Bech32PubKeyTypeAccPub, alias.PubKey)
		if err != nil {
			return err
		}
		if err := acc.SetPubKey(pubkey); err != nil {
			return err
		}
	}

	acc.CodeHash = ethcmn.HexToHash(alias.CodeHash).Bytes()

	return nil
}

// String implements the fmt.Stringer interface
func (acc EthAccount) String() string {
	out, _ := yaml.Marshal(acc)
	return string(out)
}
