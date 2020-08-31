package types

import (
	"encoding/json"
	"fmt"

	"gopkg.in/yaml.v2"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/exported"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	tmamino "github.com/tendermint/tendermint/crypto/encoding/amino"

	ethcmn "github.com/ethereum/go-ethereum/common"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
)

var _ exported.Account = (*EthAccount)(nil)
var _ exported.GenesisAccount = (*EthAccount)(nil)

func init() {
	authtypes.RegisterAccountTypeCodec(&EthAccount{}, EthAccountName)
}

// ----------------------------------------------------------------------------
// Main Ethermint account
// ----------------------------------------------------------------------------

// EthAccount implements the auth.Account interface and embeds an
// auth.BaseAccount type. It is compatible with the auth.AccountKeeper.
type EthAccount struct {
	*authtypes.BaseAccount `json:"base_account" yaml:"base_account"`
	CodeHash               []byte `json:"code_hash" yaml:"code_hash"`
}

// ProtoAccount defines the prototype function for BaseAccount used for an
// AccountKeeper.
func ProtoAccount() exported.Account {
	return &EthAccount{
		BaseAccount: &auth.BaseAccount{},
		CodeHash:    ethcrypto.Keccak256(nil),
	}
}

// EthAddress returns the account address ethereum format.
func (acc EthAccount) EthAddress() ethcmn.Address {
	return ethcmn.BytesToAddress(acc.Address.Bytes())
}

// TODO: remove on SDK v0.40

// Balance returns the balance of an account.
func (acc EthAccount) Balance() sdk.Int {
	return acc.GetCoins().AmountOf(DenomDefault)
}

// SetBalance sets an account's balance of aphotons
func (acc *EthAccount) SetBalance(amt sdk.Int) {
	coins := acc.GetCoins()
	diff := amt.Sub(coins.AmountOf(DenomDefault))
	switch {
	case diff.IsPositive():
		// Increase coins to amount
		coins = coins.Add(sdk.NewCoin(DenomDefault, diff))
	case diff.IsNegative():
		// Decrease coins to amount
		coins = coins.Sub(sdk.NewCoins(sdk.NewCoin(DenomDefault, diff.Neg())))
	default:
		return
	}

	if err := acc.SetCoins(coins); err != nil {
		panic(fmt.Errorf("could not set coins for address %s: %w", acc.EthAddress().String(), err))
	}
}

type ethermintAccountPretty struct {
	Address       sdk.AccAddress `json:"address" yaml:"address"`
	Coins         sdk.Coins      `json:"coins" yaml:"coins"`
	PubKey        []byte         `json:"public_key" yaml:"public_key"`
	AccountNumber uint64         `json:"account_number" yaml:"account_number"`
	Sequence      uint64         `json:"sequence" yaml:"sequence"`
	CodeHash      string         `json:"code_hash" yaml:"code_hash"`
}

// MarshalYAML returns the YAML representation of an account.
func (acc EthAccount) MarshalYAML() (interface{}, error) {
	alias := ethermintAccountPretty{
		Address:       acc.Address,
		Coins:         acc.Coins,
		AccountNumber: acc.AccountNumber,
		Sequence:      acc.Sequence,
		CodeHash:      ethcmn.Bytes2Hex(acc.CodeHash),
	}

	if acc.PubKey != nil {
		alias.PubKey = acc.PubKey.Bytes()
	}

	bz, err := yaml.Marshal(alias)
	if err != nil {
		return nil, err
	}

	return string(bz), err
}

// MarshalJSON returns the JSON representation of an EthAccount.
func (acc EthAccount) MarshalJSON() ([]byte, error) {
	alias := ethermintAccountPretty{
		Address:       acc.Address,
		Coins:         acc.Coins,
		AccountNumber: acc.AccountNumber,
		Sequence:      acc.Sequence,
		CodeHash:      ethcmn.Bytes2Hex(acc.CodeHash),
	}

	if acc.PubKey != nil {
		alias.PubKey = acc.PubKey.Bytes()
	}

	return json.Marshal(alias)
}

// UnmarshalJSON unmarshals raw JSON bytes into an EthAccount.
func (acc *EthAccount) UnmarshalJSON(bz []byte) error {
	acc.BaseAccount = &authtypes.BaseAccount{}
	var alias ethermintAccountPretty
	if err := json.Unmarshal(bz, &alias); err != nil {
		return err
	}

	if alias.PubKey != nil {
		pubKey, err := tmamino.PubKeyFromBytes(alias.PubKey)
		if err != nil {
			return err
		}

		acc.BaseAccount.PubKey = pubKey
	}

	acc.BaseAccount.Coins = alias.Coins
	acc.BaseAccount.Address = alias.Address
	acc.BaseAccount.AccountNumber = alias.AccountNumber
	acc.BaseAccount.Sequence = alias.Sequence
	acc.CodeHash = ethcmn.Hex2Bytes(alias.CodeHash)

	return nil
}
