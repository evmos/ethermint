package keys

import (
	"fmt"

	cosmosKeys "github.com/cosmos/cosmos-sdk/crypto/keys"
	emintCrypto "github.com/cosmos/ethermint/crypto"
	"github.com/tendermint/tendermint/crypto"

	"github.com/cosmos/cosmos-sdk/crypto/keys/hd"
	"github.com/cosmos/cosmos-sdk/types"
)

// KeyType reflects a human-readable type for key listing.
type KeyType uint

// Info KeyTypes
const (
	TypeLocal   KeyType = 0
	TypeLedger  KeyType = 1
	TypeOffline KeyType = 2
	TypeMulti   KeyType = 3
)

var keyTypes = map[KeyType]string{
	TypeLocal:   "local",
	TypeLedger:  "ledger",
	TypeOffline: "offline",
	TypeMulti:   "multi",
}

// String implements the stringer interface for KeyType.
func (kt KeyType) String() string {
	return keyTypes[kt]
}

var (
	_ cosmosKeys.Info = &localInfo{}
	_ cosmosKeys.Info = &ledgerInfo{}
	_ cosmosKeys.Info = &offlineInfo{}
)

// localInfo is the public information about a locally stored key
type localInfo struct {
	Name         string                      `json:"name"`
	PubKey       emintCrypto.PubKeySecp256k1 `json:"pubkey"`
	PrivKeyArmor string                      `json:"privkey.armor"`
}

func newLocalInfo(name string, pub emintCrypto.PubKeySecp256k1, privArmor string) cosmosKeys.Info {
	return &localInfo{
		Name:         name,
		PubKey:       pub,
		PrivKeyArmor: privArmor,
	}
}

func (i localInfo) GetType() cosmosKeys.KeyType {
	return cosmosKeys.TypeLocal
}

func (i localInfo) GetName() string {
	return i.Name
}

func (i localInfo) GetPubKey() crypto.PubKey {
	return i.PubKey
}

func (i localInfo) GetAddress() types.AccAddress {
	return i.PubKey.Address().Bytes()
}

func (i localInfo) GetPath() (*hd.BIP44Params, error) {
	return nil, fmt.Errorf("BIP44 Paths are not available for this type")
}

// ledgerInfo is the public information about a Ledger key
type ledgerInfo struct {
	Name   string                      `json:"name"`
	PubKey emintCrypto.PubKeySecp256k1 `json:"pubkey"`
	Path   hd.BIP44Params              `json:"path"`
}

func newLedgerInfo(name string, pub emintCrypto.PubKeySecp256k1, path hd.BIP44Params) cosmosKeys.Info {
	return &ledgerInfo{
		Name:   name,
		PubKey: pub,
		Path:   path,
	}
}

func (i ledgerInfo) GetType() cosmosKeys.KeyType {
	return cosmosKeys.TypeLedger
}

func (i ledgerInfo) GetName() string {
	return i.Name
}

func (i ledgerInfo) GetPubKey() crypto.PubKey {
	return i.PubKey
}

func (i ledgerInfo) GetAddress() types.AccAddress {
	return i.PubKey.Address().Bytes()
}

func (i ledgerInfo) GetPath() (*hd.BIP44Params, error) {
	tmp := i.Path
	return &tmp, nil
}

// offlineInfo is the public information about an offline key
type offlineInfo struct {
	Name   string                      `json:"name"`
	PubKey emintCrypto.PubKeySecp256k1 `json:"pubkey"`
}

func newOfflineInfo(name string, pub emintCrypto.PubKeySecp256k1) cosmosKeys.Info {
	return &offlineInfo{
		Name:   name,
		PubKey: pub,
	}
}

func (i offlineInfo) GetType() cosmosKeys.KeyType {
	return cosmosKeys.TypeOffline
}

func (i offlineInfo) GetName() string {
	return i.Name
}

func (i offlineInfo) GetPubKey() crypto.PubKey {
	return i.PubKey
}

func (i offlineInfo) GetAddress() types.AccAddress {
	return i.PubKey.Address().Bytes()
}

func (i offlineInfo) GetPath() (*hd.BIP44Params, error) {
	return nil, fmt.Errorf("BIP44 Paths are not available for this type")
}

// encoding info
func writeInfo(i cosmosKeys.Info) []byte {
	return cdc.MustMarshalBinaryLengthPrefixed(i)
}

// decoding info
func readInfo(bz []byte) (info cosmosKeys.Info, err error) {
	err = cdc.UnmarshalBinaryLengthPrefixed(bz, &info)
	return
}
