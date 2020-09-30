package rpc

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"time"

	"github.com/spf13/viper"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/crypto/keys"
	"github.com/cosmos/cosmos-sdk/crypto/keys/mintkey"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"

	emintcrypto "github.com/cosmos/ethermint/crypto"
	params "github.com/cosmos/ethermint/rpc/args"
)

// PersonalEthAPI is the personal_ prefixed set of APIs in the Web3 JSON-RPC spec.
type PersonalEthAPI struct {
	ethAPI   *PublicEthAPI
	keyInfos []keys.Info // all keys, both locked and unlocked. unlocked keys are stored in ethAPI.keys
}

// NewPersonalEthAPI creates an instance of the public Personal Eth API.
func NewPersonalEthAPI(ethAPI *PublicEthAPI) *PersonalEthAPI {
	api := &PersonalEthAPI{
		ethAPI: ethAPI,
	}

	infos, err := api.getKeybaseInfo()
	if err != nil {
		return api
	}

	api.keyInfos = infos
	return api
}

func (e *PersonalEthAPI) getKeybaseInfo() ([]keys.Info, error) {
	e.ethAPI.keybaseLock.Lock()
	defer e.ethAPI.keybaseLock.Unlock()

	if e.ethAPI.cliCtx.Keybase == nil {
		keybase, err := keys.NewKeyring(
			sdk.KeyringServiceName(),
			viper.GetString(flags.FlagKeyringBackend),
			viper.GetString(flags.FlagHome),
			e.ethAPI.cliCtx.Input,
			emintcrypto.EthSecp256k1Options()...,
		)
		if err != nil {
			return nil, err
		}

		e.ethAPI.cliCtx.Keybase = keybase
	}

	return e.ethAPI.cliCtx.Keybase.List()
}

// ImportRawKey armors and encrypts a given raw hex encoded ECDSA key and stores it into the key directory.
// The name of the key will have the format "personal_<length-keys>", where <length-keys> is the total number of
// keys stored on the keyring.
// NOTE: The key will be both armored and encrypted using the same passphrase.
func (e *PersonalEthAPI) ImportRawKey(privkey, password string) (common.Address, error) {
	e.ethAPI.logger.Debug("personal_importRawKey")
	priv, err := crypto.HexToECDSA(privkey)
	if err != nil {
		return common.Address{}, err
	}

	privKey := emintcrypto.PrivKeySecp256k1(crypto.FromECDSA(priv))

	armor := mintkey.EncryptArmorPrivKey(privKey, password, emintcrypto.EthSecp256k1Type)

	// ignore error as we only care about the length of the list
	list, _ := e.ethAPI.cliCtx.Keybase.List()
	privKeyName := fmt.Sprintf("personal_%d", len(list))

	if err := e.ethAPI.cliCtx.Keybase.ImportPrivKey(privKeyName, armor, password); err != nil {
		return common.Address{}, err
	}

	addr := common.BytesToAddress(privKey.PubKey().Address().Bytes())

	info, err := e.ethAPI.cliCtx.Keybase.Get(privKeyName)
	if err != nil {
		return common.Address{}, err
	}

	// append key and info to be able to lock and list the account
	//e.ethAPI.keys = append(e.ethAPI.keys, privKey)
	e.keyInfos = append(e.keyInfos, info)
	e.ethAPI.logger.Info("key successfully imported", "name", privKeyName, "address", addr.String())

	return addr, nil
}

// ListAccounts will return a list of addresses for accounts this node manages.
func (e *PersonalEthAPI) ListAccounts() ([]common.Address, error) {
	e.ethAPI.logger.Debug("personal_listAccounts")
	addrs := []common.Address{}
	for _, info := range e.keyInfos {
		addressBytes := info.GetPubKey().Address().Bytes()
		addrs = append(addrs, common.BytesToAddress(addressBytes))
	}

	return addrs, nil
}

// LockAccount will lock the account associated with the given address when it's unlocked.
// It removes the key corresponding to the given address from the API's local keys.
func (e *PersonalEthAPI) LockAccount(address common.Address) bool {
	e.ethAPI.logger.Debug("personal_lockAccount", "address", address.String())

	for i, key := range e.ethAPI.keys {
		if !bytes.Equal(key.PubKey().Address().Bytes(), address.Bytes()) {
			continue
		}

		tmp := make([]emintcrypto.PrivKeySecp256k1, len(e.ethAPI.keys)-1)
		copy(tmp[:i], e.ethAPI.keys[:i])
		copy(tmp[i:], e.ethAPI.keys[i+1:])
		e.ethAPI.keys = tmp

		e.ethAPI.logger.Debug("account unlocked", "address", address.String())
		return true
	}

	return false
}

// NewAccount will create a new account and returns the address for the new account.
func (e *PersonalEthAPI) NewAccount(password string) (common.Address, error) {
	e.ethAPI.logger.Debug("personal_newAccount")
	_, err := e.getKeybaseInfo()
	if err != nil {
		return common.Address{}, err
	}

	name := "key_" + time.Now().UTC().Format(time.RFC3339)
	info, _, err := e.ethAPI.cliCtx.Keybase.CreateMnemonic(name, keys.English, password, emintcrypto.EthSecp256k1)
	if err != nil {
		return common.Address{}, err
	}

	e.keyInfos = append(e.keyInfos, info)

	addr := common.BytesToAddress(info.GetPubKey().Address().Bytes())
	e.ethAPI.logger.Info("Your new key was generated", "address", addr.String())
	e.ethAPI.logger.Info("Please backup your key file!", "path", os.Getenv("HOME")+"/.ethermintcli/"+name)
	e.ethAPI.logger.Info("Please remember your password!")
	return addr, nil
}

// UnlockAccount will unlock the account associated with the given address with
// the given password for duration seconds. If duration is nil it will use a
// default of 300 seconds. It returns an indication if the account was unlocked.
// It exports the private key corresponding to the given address from the keyring and stores it in the API's local keys.
func (e *PersonalEthAPI) UnlockAccount(_ context.Context, addr common.Address, password string, _ *uint64) (bool, error) { // nolint: interfacer
	e.ethAPI.logger.Debug("personal_unlockAccount", "address", addr.String())
	// TODO: use duration

	var keyInfo keys.Info

	for _, info := range e.keyInfos {
		addressBytes := info.GetPubKey().Address().Bytes()
		if bytes.Equal(addressBytes, addr[:]) {
			keyInfo = info
			break
		}
	}

	if keyInfo == nil {
		return false, fmt.Errorf("cannot find key with given address %s", addr.String())
	}

	// exporting private key only works on local keys
	if keyInfo.GetType() != keys.TypeLocal {
		return false, fmt.Errorf("key type must be %s, got %s", keys.TypeLedger.String(), keyInfo.GetType().String())
	}

	privKey, err := e.ethAPI.cliCtx.Keybase.ExportPrivateKeyObject(keyInfo.GetName(), password)
	if err != nil {
		return false, err
	}

	emintKey, ok := privKey.(emintcrypto.PrivKeySecp256k1)
	if !ok {
		return false, fmt.Errorf("invalid private key type: %T", privKey)
	}

	e.ethAPI.keys = append(e.ethAPI.keys, emintKey)
	e.ethAPI.logger.Debug("account unlocked", "address", addr.String())
	return true, nil
}

// SendTransaction will create a transaction from the given arguments and
// tries to sign it with the key associated with args.To. If the given password isn't
// able to decrypt the key it fails.
func (e *PersonalEthAPI) SendTransaction(_ context.Context, args params.SendTxArgs, _ string) (common.Hash, error) {
	return e.ethAPI.SendTransaction(args)
}

// Sign calculates an Ethereum ECDSA signature for:
// keccak256("\x19Ethereum Signed Message:\n" + len(message) + message))
//
// Note, the produced signature conforms to the secp256k1 curve R, S and V values,
// where the V value will be 27 or 28 for legacy reasons.
//
// The key used to calculate the signature is decrypted with the given password.
//
// https://github.com/ethereum/go-ethereum/wiki/Management-APIs#personal_sign
func (e *PersonalEthAPI) Sign(_ context.Context, data hexutil.Bytes, addr common.Address, _ string) (hexutil.Bytes, error) {
	e.ethAPI.logger.Debug("personal_sign", "data", data, "address", addr.String())

	key, ok := checkKeyInKeyring(e.ethAPI.keys, addr)
	if !ok {
		return nil, fmt.Errorf("cannot find key with address %s", addr.String())
	}

	sig, err := crypto.Sign(accounts.TextHash(data), key.ToECDSA())
	if err != nil {
		return nil, err
	}

	sig[crypto.RecoveryIDOffset] += 27 // transform V from 0/1 to 27/28
	return sig, nil
}

// EcRecover returns the address for the account that was used to create the signature.
// Note, this function is compatible with eth_sign and personal_sign. As such it recovers
// the address of:
// hash = keccak256("\x19Ethereum Signed Message:\n"${message length}${message})
// addr = ecrecover(hash, signature)
//
// Note, the signature must conform to the secp256k1 curve R, S and V values, where
// the V value must be 27 or 28 for legacy reasons.
//
// https://github.com/ethereum/go-ethereum/wiki/Management-APIs#personal_ecRecove
func (e *PersonalEthAPI) EcRecover(_ context.Context, data, sig hexutil.Bytes) (common.Address, error) {
	e.ethAPI.logger.Debug("personal_ecRecover", "data", data, "sig", sig)

	if len(sig) != crypto.SignatureLength {
		return common.Address{}, fmt.Errorf("signature must be %d bytes long", crypto.SignatureLength)
	}
	if sig[crypto.RecoveryIDOffset] != 27 && sig[crypto.RecoveryIDOffset] != 28 {
		return common.Address{}, fmt.Errorf("invalid Ethereum signature (V is not 27 or 28)")
	}
	sig[crypto.RecoveryIDOffset] -= 27 // Transform yellow paper V from 27/28 to 0/1

	pubkey, err := crypto.SigToPub(accounts.TextHash(data), sig)
	if err != nil {
		return common.Address{}, err
	}
	return crypto.PubkeyToAddress(*pubkey), nil
}
