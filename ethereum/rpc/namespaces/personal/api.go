package personal

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/tharsis/ethermint/ethereum/rpc/backend"

	"github.com/cosmos/cosmos-sdk/client"

	"github.com/tharsis/ethermint/crypto/hd"
	ethermint "github.com/tharsis/ethermint/types"

	"github.com/tendermint/tendermint/libs/log"

	sdkcrypto "github.com/cosmos/cosmos-sdk/crypto"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/tharsis/ethermint/crypto/ethsecp256k1"
	rpctypes "github.com/tharsis/ethermint/ethereum/rpc/types"
)

// PrivateAccountAPI is the personal_ prefixed set of APIs in the Web3 JSON-RPC spec.
type PrivateAccountAPI struct {
	clientCtx  client.Context
	backend    backend.Backend
	logger     log.Logger
	hdPathIter ethermint.HDPathIterator
}

// NewAPI creates an instance of the public Personal Eth API.
func NewAPI(logger log.Logger, clientCtx client.Context, backend backend.Backend) *PrivateAccountAPI {
	cfg := sdk.GetConfig()
	basePath := cfg.GetFullBIP44Path()

	iterator, err := ethermint.NewHDPathIterator(basePath, true)
	if err != nil {
		panic(err)
	}

	return &PrivateAccountAPI{
		clientCtx:  clientCtx,
		logger:     logger.With("api", "personal"),
		hdPathIter: iterator,
		backend:    backend,
	}
}

// ImportRawKey armors and encrypts a given raw hex encoded ECDSA key and stores it into the key directory.
// The name of the key will have the format "personal_<length-keys>", where <length-keys> is the total number of
// keys stored on the keyring.
//
// NOTE: The key will be both armored and encrypted using the same passphrase.
func (api *PrivateAccountAPI) ImportRawKey(privkey, password string) (common.Address, error) {
	api.logger.Debug("personal_importRawKey")
	priv, err := crypto.HexToECDSA(privkey)
	if err != nil {
		return common.Address{}, err
	}

	privKey := &ethsecp256k1.PrivKey{Key: crypto.FromECDSA(priv)}

	addr := sdk.AccAddress(privKey.PubKey().Address().Bytes())
	ethereumAddr := common.BytesToAddress(addr)

	// return if the key has already been imported
	if _, err := api.clientCtx.Keyring.KeyByAddress(addr); err == nil {
		return ethereumAddr, nil
	}

	// ignore error as we only care about the length of the list
	list, _ := api.clientCtx.Keyring.List()
	privKeyName := fmt.Sprintf("personal_%d", len(list))

	armor := sdkcrypto.EncryptArmorPrivKey(privKey, password, ethsecp256k1.KeyType)

	if err := api.clientCtx.Keyring.ImportPrivKey(privKeyName, armor, password); err != nil {
		return common.Address{}, err
	}

	api.logger.Info("key successfully imported", "name", privKeyName, "address", ethereumAddr.String())

	return ethereumAddr, nil
}

// ListAccounts will return a list of addresses for accounts this node manages.
func (api *PrivateAccountAPI) ListAccounts() ([]common.Address, error) {
	api.logger.Debug("personal_listAccounts")
	addrs := []common.Address{}

	list, err := api.clientCtx.Keyring.List()
	if err != nil {
		return nil, err
	}

	for _, info := range list {
		addrs = append(addrs, common.BytesToAddress(info.GetPubKey().Address()))
	}

	return addrs, nil
}

// LockAccount will lock the account associated with the given address when it's unlocked.
// It removes the key corresponding to the given address from the API's local keys.
func (api *PrivateAccountAPI) LockAccount(address common.Address) bool { // nolint: interfacer
	api.logger.Debug("personal_lockAccount", "address", address.String())
	api.logger.Info("personal_lockAccount not supported")
	// TODO: Not supported. See underlying issue  https://github.com/99designs/keyring/issues/85
	return false
}

// NewAccount will create a new account and returns the address for the new account.
func (api *PrivateAccountAPI) NewAccount(password string) (common.Address, error) {
	api.logger.Debug("personal_newAccount")

	name := "key_" + time.Now().UTC().Format(time.RFC3339)

	// create the mnemonic and save the account
	hdPath := api.hdPathIter()

	info, _, err := api.clientCtx.Keyring.NewMnemonic(name, keyring.English, hdPath.String(), password, hd.EthSecp256k1)
	if err != nil {
		return common.Address{}, err
	}

	addr := common.BytesToAddress(info.GetPubKey().Address().Bytes())
	api.logger.Info("Your new key was generated", "address", addr.String())
	api.logger.Info("Please backup your key file!", "path", os.Getenv("HOME")+"/.ethermint/"+name) // TODO: pass the correct binary
	api.logger.Info("Please remember your password!")
	return addr, nil
}

// UnlockAccount will unlock the account associated with the given address with
// the given password for duration seconds. If duration is nil it will use a
// default of 300 seconds. It returns an indication if the account was unlocked.
func (api *PrivateAccountAPI) UnlockAccount(_ context.Context, addr common.Address, _ string, _ *uint64) (bool, error) { // nolint: interfacer
	api.logger.Debug("personal_unlockAccount", "address", addr.String())
	// TODO: Not supported. See underlying issue  https://github.com/99designs/keyring/issues/85
	return false, nil
}

// SendTransaction will create a transaction from the given arguments and
// tries to sign it with the key associated with args.To. If the given password isn't
// able to decrypt the key it fails.
func (api *PrivateAccountAPI) SendTransaction(_ context.Context, args rpctypes.SendTxArgs, pwrd string) (common.Hash, error) {
	api.logger.Debug("personal_sendTransaction", "address", args.To.String())
	return api.backend.SendTransaction(args)
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
func (api *PrivateAccountAPI) Sign(_ context.Context, data hexutil.Bytes, addr common.Address, pwrd string) (hexutil.Bytes, error) {
	api.logger.Debug("personal_sign", "data", data, "address", addr.String())

	cosmosAddr := sdk.AccAddress(addr.Bytes())

	sig, _, err := api.clientCtx.Keyring.SignByAddress(cosmosAddr, accounts.TextHash(data))
	if err != nil {
		api.logger.Error("failed to sign with key", "data", data, "address", addr.String(), "error", err.Error())
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
