package types

import (
	"encoding/json"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/exported"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	tmamino "github.com/tendermint/tendermint/crypto/encoding/amino"
	"gopkg.in/yaml.v2"

	ethcmn "github.com/ethereum/go-ethereum/common"
)

var _ exported.Account = (*Account)(nil)
var _ exported.GenesisAccount = (*Account)(nil)

const (
	// DenomDefault defines the single coin type/denomination supported in
	// Ethermint.
	DenomDefault = "photon"
)

// ----------------------------------------------------------------------------
// Main Ethermint account
// ----------------------------------------------------------------------------

func init() {
	authtypes.RegisterAccountTypeCodec(Account{}, EthermintAccountName)
}

// Account implements the auth.Account interface and embeds an
// auth.BaseAccount type. It is compatible with the auth.AccountMapper.
type Account struct {
	*auth.BaseAccount

	// merkle root of the storage trie
	//
	// TODO: add back root if needed (marshalling is broken if not initializing)
	// Root ethcmn.Hash

	CodeHash []byte
}

// ProtoBaseAccount defines the prototype function for BaseAccount used for an
// account mapper.
func ProtoBaseAccount() exported.Account {
	return &Account{BaseAccount: &auth.BaseAccount{}}
}

// Balance returns the balance of an account.
func (acc Account) Balance() sdk.Int {
	return acc.GetCoins().AmountOf(DenomDefault)
}

// SetBalance sets an account's balance of photons
func (acc Account) SetBalance(amt sdk.Int) {
	coins := acc.GetCoins()
	diff := amt.Sub(coins.AmountOf(DenomDefault))
	if diff.IsZero() {
		return
	} else if diff.IsPositive() {
		// Increase coins to amount
		coins = coins.Add(sdk.Coins{sdk.NewCoin(DenomDefault, diff)})
	} else {
		// Decrease coins to amount
		coins = coins.Sub(sdk.Coins{sdk.NewCoin(DenomDefault, diff.Neg())})
	}
	if err := acc.SetCoins(coins); err != nil {
		panic(fmt.Sprintf("Could not set coins for address %s", acc.GetAddress()))
	}
}

// ----------------------------------------------------------------------------
// Code & Storage
// ----------------------------------------------------------------------------

type (
	// Code is account Code type alias
	Code []byte
	// Storage is account storage type alias
	Storage map[ethcmn.Hash]ethcmn.Hash
)

func (c Code) String() string {
	return string(c)
}

func (c Storage) String() (str string) {
	for key, value := range c {
		str += fmt.Sprintf("%X : %X\n", key, value)
	}

	return
}

// Copy returns a copy of storage.
func (c Storage) Copy() Storage {
	cpy := make(Storage)
	for key, value := range c {
		cpy[key] = value
	}

	return cpy
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
func (acc Account) MarshalYAML() (interface{}, error) {
	alias := ethermintAccountPretty{
		Address:       acc.Address,
		Coins:         acc.Coins,
		AccountNumber: acc.AccountNumber,
		Sequence:      acc.Sequence,
		CodeHash:      ethcmn.Bytes2Hex(acc.CodeHash),
	}

	if acc.PubKey != nil {
		alias.PubKey = acc.PubKey.Bytes()
		fmt.Println(len(alias.PubKey), alias.PubKey)
	}

	bz, err := yaml.Marshal(alias)
	if err != nil {
		return nil, err
	}

	return string(bz), err
}

// MarshalJSON returns the JSON representation of an Account.
func (acc Account) MarshalJSON() ([]byte, error) {
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

// UnmarshalJSON unmarshals raw JSON bytes into an Account.
func (acc *Account) UnmarshalJSON(bz []byte) error {
	acc.BaseAccount = &authtypes.BaseAccount{}
	var alias ethermintAccountPretty
	if err := json.Unmarshal(bz, &alias); err != nil {
		return err
	}

	if alias.PubKey != nil {
		pubk, err := tmamino.PubKeyFromBytes(alias.PubKey)
		if err != nil {
			return err
		}

		acc.BaseAccount.PubKey = pubk
	}

	acc.BaseAccount.Address = alias.Address
	acc.BaseAccount.Coins = alias.Coins
	acc.BaseAccount.AccountNumber = alias.AccountNumber
	acc.BaseAccount.Sequence = alias.Sequence
	acc.CodeHash = ethcmn.Hex2Bytes(alias.CodeHash)

	return nil
}
