package rpc

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	sdkcontext "github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/crypto/keys"
	sdk "github.com/cosmos/cosmos-sdk/types"
	emintcrypto "github.com/cosmos/ethermint/crypto"
	params "github.com/cosmos/ethermint/rpc/args"
	"github.com/spf13/viper"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
)

// PersonalEthAPI is the eth_ prefixed set of APIs in the Web3 JSON-RPC spec.
type PersonalEthAPI struct {
	logger      log.Logger
	cliCtx      sdkcontext.CLIContext
	ethAPI      *PublicEthAPI
	nonceLock   *AddrLocker
	keys        []emintcrypto.PrivKeySecp256k1
	keyInfos    []keys.Info
	keybaseLock sync.Mutex
}

// NewPersonalEthAPI creates an instance of the public ETH Web3 API.
func NewPersonalEthAPI(cliCtx sdkcontext.CLIContext, ethAPI *PublicEthAPI, nonceLock *AddrLocker, keys []emintcrypto.PrivKeySecp256k1) *PersonalEthAPI {
	api := &PersonalEthAPI{
		logger:    log.NewTMLogger(log.NewSyncWriter(os.Stdout)).With("module", "json-rpc"),
		cliCtx:    cliCtx,
		ethAPI:    ethAPI,
		nonceLock: nonceLock,
		keys:      keys,
	}

	infos, err := api.getKeybaseInfo()
	if err != nil {
		return api
	}

	api.keyInfos = infos
	return api
}

func (e *PersonalEthAPI) getKeybaseInfo() ([]keys.Info, error) {
	e.keybaseLock.Lock()
	defer e.keybaseLock.Unlock()

	if e.cliCtx.Keybase == nil {
		keybase, err := keys.NewKeyring(
			sdk.KeyringServiceName(),
			viper.GetString(flags.FlagKeyringBackend),
			viper.GetString(flags.FlagHome),
			e.cliCtx.Input,
			emintcrypto.EthSecp256k1Options()...,
		)
		if err != nil {
			return nil, err
		}

		e.cliCtx.Keybase = keybase
	}

	return e.cliCtx.Keybase.List()
}

// ImportRawKey stores the given hex encoded ECDSA key into the key directory,
// encrypting it with the passphrase.
// Currently, this is not implemented since the feature is not supported by the keys.
func (e *PersonalEthAPI) ImportRawKey(privkey, password string) (common.Address, error) {
	e.logger.Debug("personal_importRawKey", "error", "not implemented")
	_, err := crypto.HexToECDSA(privkey)
	if err != nil {
		return common.Address{}, err
	}

	return common.Address{}, nil
}

// ListAccounts will return a list of addresses for accounts this node manages.
func (e *PersonalEthAPI) ListAccounts() ([]common.Address, error) {
	e.logger.Debug("personal_listAccounts")
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
	e.logger.Debug("personal_lockAccount", "address", address)
	for i, key := range e.keys {
		if !bytes.Equal(key.PubKey().Address().Bytes(), address.Bytes()) {
			continue
		}

		tmp := make([]emintcrypto.PrivKeySecp256k1, len(e.keys)-1)
		copy(tmp[:i], e.keys[:i])
		copy(tmp[i:], e.keys[i+1:])
		e.keys = tmp
		return true
	}

	for i, key := range e.ethAPI.keys {
		if !bytes.Equal(key.PubKey().Address().Bytes(), address.Bytes()) {
			continue
		}

		tmp := make([]emintcrypto.PrivKeySecp256k1, len(e.ethAPI.keys)-1)
		copy(tmp[:i], e.ethAPI.keys[:i])
		copy(tmp[i:], e.ethAPI.keys[i+1:])
		e.ethAPI.keys = tmp
		return true
	}

	return false
}

// NewAccount will create a new account and returns the address for the new account.
func (e *PersonalEthAPI) NewAccount(password string) (common.Address, error) {
	e.logger.Debug("personal_newAccount")
	_, err := e.getKeybaseInfo()
	if err != nil {
		return common.Address{}, err
	}

	name := "key_" + time.Now().UTC().Format(time.RFC3339)
	info, _, err := e.cliCtx.Keybase.CreateMnemonic(name, keys.English, password, emintcrypto.EthSecp256k1)
	if err != nil {
		return common.Address{}, err
	}

	e.keyInfos = append(e.keyInfos, info)

	// update ethAPI
	privKey, err := e.cliCtx.Keybase.ExportPrivateKeyObject(name, password)
	if err != nil {
		return common.Address{}, err
	}

	emintKey, ok := privKey.(emintcrypto.PrivKeySecp256k1)
	if !ok {
		return common.Address{}, fmt.Errorf("invalid private key type: %T", privKey)
	}
	e.ethAPI.keys = append(e.ethAPI.keys, emintKey)
	e.logger.Debug("personal_newAccount", "address", fmt.Sprintf("0x%x", emintKey.PubKey().Address().Bytes()))

	addr := common.BytesToAddress(info.GetPubKey().Address().Bytes())
	e.logger.Info("Your new key was generated", "address", addr)
	e.logger.Info("Please backup your key file!", "path", os.Getenv("HOME")+"/.ethermintcli/"+name)
	e.logger.Info("Please remember your password!")
	return addr, nil
}

// UnlockAccount will unlock the account associated with the given address with
// the given password for duration seconds. If duration is nil it will use a
// default of 300 seconds. It returns an indication if the account was unlocked.
// It exports the private key corresponding to the given address from the keyring and stores it in the API's local keys.
func (e *PersonalEthAPI) UnlockAccount(ctx context.Context, addr common.Address, password string, _ *uint64) (bool, error) {
	e.logger.Debug("personal_unlockAccount", "address", addr)
	// TODO: use duration

	name := ""
	for _, info := range e.keyInfos {
		addressBytes := info.GetPubKey().Address().Bytes()
		if bytes.Equal(addressBytes, addr[:]) {
			name = info.GetName()
		}
	}

	if name == "" {
		return false, fmt.Errorf("cannot find key with given address")
	}

	// TODO: this only works on local keys
	privKey, err := e.cliCtx.Keybase.ExportPrivateKeyObject(name, password)
	if err != nil {
		return false, err
	}

	emintKey, ok := privKey.(emintcrypto.PrivKeySecp256k1)
	if !ok {
		return false, fmt.Errorf("invalid private key type: %T", privKey)
	}

	e.keys = append(e.keys, emintKey)
	e.ethAPI.keys = append(e.ethAPI.keys, emintKey)
	e.logger.Debug("personal_unlockAccount", "address", fmt.Sprintf("0x%x", emintKey.PubKey().Address().Bytes()))

	return true, nil
}

// SendTransaction will create a transaction from the given arguments and
// tries to sign it with the key associated with args.To. If the given password isn't
// able to decrypt the key it fails.
func (e *PersonalEthAPI) SendTransaction(ctx context.Context, args params.SendTxArgs, passwd string) (common.Hash, error) {
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
func (e *PersonalEthAPI) Sign(ctx context.Context, data hexutil.Bytes, addr common.Address, passwd string) (hexutil.Bytes, error) {
	e.logger.Debug("personal_sign", "data", data, "address", addr)

	key, ok := checkKeyInKeyring(e.keys, addr)
	if !ok {
		return nil, fmt.Errorf("cannot find key with given address")
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
func (e *PersonalEthAPI) EcRecover(ctx context.Context, data, sig hexutil.Bytes) (common.Address, error) {
	e.logger.Debug("personal_ecRecover", "data", data, "sig", sig)

	if len(sig) != crypto.SignatureLength {
		return common.Address{}, fmt.Errorf("signature must be %d bytes long", crypto.SignatureLength)
	}
	if sig[crypto.RecoveryIDOffset] != 27 && sig[crypto.RecoveryIDOffset] != 28 {
		return common.Address{}, fmt.Errorf("invalid Ethereum signature (V is not 27 or 28)")
	}
	sig[crypto.RecoveryIDOffset] -= 27 // Transform yellow paper V from 27/28 to 0/1

	rpk, err := crypto.SigToPub(accounts.TextHash(data), sig)
	if err != nil {
		return common.Address{}, err
	}
	return crypto.PubkeyToAddress(*rpk), nil
}
