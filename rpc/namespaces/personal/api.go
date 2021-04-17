package personal

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"time"

	"github.com/cosmos/ethermint/crypto/hd"
	ethermint "github.com/cosmos/ethermint/types"
	"github.com/tendermint/tendermint/libs/log"

	sdkcrypto "github.com/cosmos/cosmos-sdk/crypto"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/cosmos/ethermint/crypto/ethsecp256k1"
	"github.com/cosmos/ethermint/rpc/namespaces/eth"
	rpctypes "github.com/cosmos/ethermint/rpc/types"
)

// PrivateAccountAPI is the personal_ prefixed set of APIs in the Web3 JSON-RPC spec.
type PrivateAccountAPI struct {
	ethAPI   *eth.PublicEthereumAPI
	logger   log.Logger
	keyInfos []keyring.Info // all keys, both locked and unlocked. unlocked keys are stored in ethAPI.keys
}

// NewAPI creates an instance of the public Personal Eth API.
func NewAPI(ethAPI *eth.PublicEthereumAPI) *PrivateAccountAPI {
	api := &PrivateAccountAPI{
		ethAPI: ethAPI,
		logger: log.NewTMLogger(log.NewSyncWriter(os.Stdout)).With("module", "json-rpc", "namespace", "personal"),
	}

	err := api.ethAPI.GetKeyringInfo()
	if err != nil {
		return api
	}

	api.keyInfos, err = api.ethAPI.ClientCtx().Keyring.List()
	if err != nil {
		return api
	}

	return api
}

// ImportRawKey armors and encrypts a given raw hex encoded ECDSA key and stores it into the key directory.
// The name of the key will have the format "personal_<length-keys>", where <length-keys> is the total number of
// keys stored on the keyring.
// NOTE: The key will be both armored and encrypted using the same passphrase.
func (api *PrivateAccountAPI) ImportRawKey(privkey, password string) (common.Address, error) {
	api.logger.Debug("personal_importRawKey")
	priv, err := crypto.HexToECDSA(privkey)
	if err != nil {
		return common.Address{}, err
	}

	privKey := &ethsecp256k1.PrivKey{Key: crypto.FromECDSA(priv)}

	armor := sdkcrypto.EncryptArmorPrivKey(privKey, password, ethsecp256k1.KeyType)

	// ignore error as we only care about the length of the list
	list, _ := api.ethAPI.ClientCtx().Keyring.List()
	privKeyName := fmt.Sprintf("personal_%d", len(list))

	if err := api.ethAPI.ClientCtx().Keyring.ImportPrivKey(privKeyName, armor, password); err != nil {
		return common.Address{}, err
	}

	addr := common.BytesToAddress(privKey.PubKey().Address().Bytes())

	info, err := api.ethAPI.ClientCtx().Keyring.Key(privKeyName)
	if err != nil {
		return common.Address{}, err
	}

	// append key and info to be able to lock and list the account
	// api.ethAPI.keys = append(api.ethAPI.keys, privKey)
	api.keyInfos = append(api.keyInfos, info)
	api.logger.Info("key successfully imported", "name", privKeyName, "address", addr.String())

	return addr, nil
}

// ListAccounts will return a list of addresses for accounts this node manages.
func (api *PrivateAccountAPI) ListAccounts() ([]common.Address, error) {
	api.logger.Debug("personal_listAccounts")
	addrs := []common.Address{}
	for _, info := range api.keyInfos {
		addressBytes := info.GetPubKey().Address().Bytes()
		addrs = append(addrs, common.BytesToAddress(addressBytes))
	}

	return addrs, nil
}

// LockAccount will lock the account associated with the given address when it's unlocked.
// It removes the key corresponding to the given address from the API's local keys.
func (api *PrivateAccountAPI) LockAccount(address common.Address) bool {
	api.logger.Debug("personal_lockAccount", "address", address.String())

	keys := api.ethAPI.GetKeys()
	for i, key := range keys {
		if !bytes.Equal(key.PubKey().Address().Bytes(), address.Bytes()) {
			continue
		}

		tmp := make([]ethsecp256k1.PrivKey, len(keys)-1)
		copy(tmp[:i], keys[:i])
		copy(tmp[i:], keys[i+1:])
		api.ethAPI.SetKeys(tmp)

		api.logger.Debug("account locked", "address", address.String())
		return true
	}

	return false
}

// NewAccount will create a new account and returns the address for the new account.
func (api *PrivateAccountAPI) NewAccount(password string) (common.Address, error) {
	api.logger.Debug("personal_newAccount")

	name := "key_" + time.Now().UTC().Format(time.RFC3339)
	info, _, err := api.ethAPI.ClientCtx().Keyring.NewMnemonic(name, keyring.English, ethermint.BIP44HDPath, hd.EthSecp256k1)
	if err != nil {
		return common.Address{}, err
	}

	// TODO: rpc_test_fix: if we use NewMnemonic, it will create an account automatically, and
	// the new account encrypted with default empty string in cosmos-SDK,
	// if we create mnemonic then create new account, we basically generated two accounts

	// nolint
	//name = "key_" + time.Now().UTC().Format(time.RFC3339)
	//info, err := api.ethAPI.ClientCtx().Keyring.NewAccount(name, mnemonic, password, ethermint.BIP44HDPath, hd.EthSecp256k1)
	//if err != nil {
	//	return common.Address{}, err
	//}
	//
	//// encrypt the key info with given password
	//armor, err := api.ethAPI.ClientCtx().Keyring.ExportPrivKeyArmor(info.GetName(), password)
	//if err != nil {
	//	return common.Address{}, err
	//}

	api.keyInfos = append(api.keyInfos, info)

	addr := common.BytesToAddress(info.GetPubKey().Address().Bytes())
	api.logger.Info("Your new key was generated", "address", addr.String())
	api.logger.Info("Please backup your key file!", "path", os.Getenv("HOME")+"/.ethermint/"+name)
	api.logger.Info("Please remember your password!")
	return addr, nil
}

// UnlockAccount will unlock the account associated with the given address with
// the given password for duration seconds. If duration is nil it will use a
// default of 300 seconds. It returns an indication if the account was unlocked.
// It exports the private key corresponding to the given address from the keyring and stores it in the API's local keys.
func (api *PrivateAccountAPI) UnlockAccount(_ context.Context, addr common.Address, password string, _ *uint64) (bool, error) { // nolint: interfacer
	api.logger.Debug("personal_unlockAccount", "address", addr.String())
	// TODO: use duration

	var keyInfo keyring.Info

	for _, info := range api.keyInfos {
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
	if keyInfo.GetType() != keyring.TypeLocal {
		return false, fmt.Errorf("key type must be %s, got %s", keyring.TypeLedger.String(), keyInfo.GetType().String())
	}

	// TODO: rpc_test_fix: password is used for encrypt and then decrypt, what's the point of locking the account ?
	armor, err := api.ethAPI.ClientCtx().Keyring.ExportPrivKeyArmor(keyInfo.GetName(), password)
	if err != nil {
		return false, err
	}

	privKey, algo, err := sdkcrypto.UnarmorDecryptPrivKey(armor, password)
	if err != nil {
		return false, err
	}

	// TODO: rpc_test_fix: we need to support both ethsecp256k1 and secp256k1, however, cosmos-SDK only support secp256k1 when init with keys add CLI
	// so that the very first account has key type as secp256k1, thus, we need to support both key types here
	if algo != ethsecp256k1.KeyType {
		return false, fmt.Errorf("invalid key algorithm, got %s, expected %s", algo, ethsecp256k1.KeyType)
	}

	ethermintPrivKey, ok := privKey.(*ethsecp256k1.PrivKey)
	if !ok {
		return false, fmt.Errorf("invalid private key type %T, expected %T", privKey, &ethsecp256k1.PrivKey{})
	}

	api.ethAPI.SetKeys(append(api.ethAPI.GetKeys(), *ethermintPrivKey))
	api.logger.Debug("account unlocked", "address", addr.String())
	return true, nil
}

// SendTransaction will create a transaction from the given arguments and
// tries to sign it with the key associated with args.To. If the given password isn't
// able to decrypt the key it fails.
func (api *PrivateAccountAPI) SendTransaction(_ context.Context, args rpctypes.SendTxArgs, _ string) (common.Hash, error) {
	return api.ethAPI.SendTransaction(args)
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
func (api *PrivateAccountAPI) Sign(_ context.Context, data hexutil.Bytes, addr common.Address, _ string) (hexutil.Bytes, error) {
	api.logger.Debug("personal_sign", "data", data, "address", addr.String())

	// TODO: rpc_test_fix: because the default account is secp256k1, it can not be unlocked and put into api.ethAPI array
	key, ok := rpctypes.GetKeyByAddress(api.ethAPI.GetKeys(), addr)
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
func (api *PrivateAccountAPI) EcRecover(_ context.Context, data, sig hexutil.Bytes) (common.Address, error) {
	api.logger.Debug("personal_ecRecover", "data", data, "sig", sig)

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
