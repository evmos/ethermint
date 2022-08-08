package backend

import (
	"fmt"
	"math/big"

	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdkcrypto "github.com/cosmos/cosmos-sdk/crypto"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
	distributiontypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/evmos/ethermint/crypto/ethsecp256k1"
	rpctypes "github.com/evmos/ethermint/rpc/types"
	evmtypes "github.com/evmos/ethermint/x/evm/types"
	"github.com/pkg/errors"
	tmtypes "github.com/tendermint/tendermint/types"
)

// Accounts returns the list of accounts available to this node.
func (b *Backend) Accounts() ([]common.Address, error) {
	addresses := make([]common.Address, 0) // return [] instead of nil if empty

	infos, err := b.clientCtx.Keyring.List()
	if err != nil {
		return addresses, err
	}

	for _, info := range infos {
		pubKey, err := info.GetPubKey()
		if err != nil {
			return nil, err
		}
		addressBytes := pubKey.Address().Bytes()
		addresses = append(addresses, common.BytesToAddress(addressBytes))
	}

	return addresses, nil
}

// Syncing returns false in case the node is currently not syncing with the network. It can be up to date or has not
// yet received the latest block headers from its pears. In case it is synchronizing:
// - startingBlock: block number this node started to synchronize from
// - currentBlock:  block number this node is currently importing
// - highestBlock:  block number of the highest block header this node has received from peers
// - pulledStates:  number of state entries processed until now
// - knownStates:   number of known state entries that still need to be pulled
func (b *Backend) Syncing() (interface{}, error) {
	status, err := b.clientCtx.Client.Status(b.ctx)
	if err != nil {
		return false, err
	}

	if !status.SyncInfo.CatchingUp {
		return false, nil
	}

	return map[string]interface{}{
		"startingBlock": hexutil.Uint64(status.SyncInfo.EarliestBlockHeight),
		"currentBlock":  hexutil.Uint64(status.SyncInfo.LatestBlockHeight),
		// "highestBlock":  nil, // NA
		// "pulledStates":  nil, // NA
		// "knownStates":   nil, // NA
	}, nil
}

// Sign signs the provided data using the private key of address via Geth's signature standard.
func (b *Backend) Sign(address common.Address, data hexutil.Bytes) (hexutil.Bytes, error) {
	from := sdk.AccAddress(address.Bytes())

	_, err := b.clientCtx.Keyring.KeyByAddress(from)
	if err != nil {
		b.logger.Error("failed to find key in keyring", "address", address.String())
		return nil, fmt.Errorf("%s; %s", keystore.ErrNoMatch, err.Error())
	}

	// Sign the requested hash with the wallet
	signature, _, err := b.clientCtx.Keyring.SignByAddress(from, data)
	if err != nil {
		b.logger.Error("keyring.SignByAddress failed", "address", address.Hex())
		return nil, err
	}

	signature[crypto.RecoveryIDOffset] += 27 // Transform V from 0/1 to 27/28 according to the yellow paper
	return signature, nil
}

// SendTransaction sends transaction based on received args using Node's key to sign it
func (b *Backend) SendTransaction(args evmtypes.TransactionArgs) (common.Hash, error) {
	// Look up the wallet containing the requested signer
	_, err := b.clientCtx.Keyring.KeyByAddress(sdk.AccAddress(args.From.Bytes()))
	if err != nil {
		b.logger.Error("failed to find key in keyring", "address", args.From, "error", err.Error())
		return common.Hash{}, fmt.Errorf("%s; %s", keystore.ErrNoMatch, err.Error())
	}

	args, err = b.SetTxDefaults(args)
	if err != nil {
		return common.Hash{}, err
	}

	msg := args.ToTransaction()
	if err := msg.ValidateBasic(); err != nil {
		b.logger.Debug("tx failed basic validation", "error", err.Error())
		return common.Hash{}, err
	}

	bn, err := b.BlockNumber()
	if err != nil {
		b.logger.Debug("failed to fetch latest block number", "error", err.Error())
		return common.Hash{}, err
	}

	signer := ethtypes.MakeSigner(b.ChainConfig(), new(big.Int).SetUint64(uint64(bn)))

	// Sign transaction
	if err := msg.Sign(signer, b.clientCtx.Keyring); err != nil {
		b.logger.Debug("failed to sign tx", "error", err.Error())
		return common.Hash{}, err
	}

	// Query params to use the EVM denomination
	res, err := b.queryClient.QueryClient.Params(b.ctx, &evmtypes.QueryParamsRequest{})
	if err != nil {
		b.logger.Error("failed to query evm params", "error", err.Error())
		return common.Hash{}, err
	}

	// Assemble transaction from fields
	tx, err := msg.BuildTx(b.clientCtx.TxConfig.NewTxBuilder(), res.Params.EvmDenom)
	if err != nil {
		b.logger.Error("build cosmos tx failed", "error", err.Error())
		return common.Hash{}, err
	}

	// Encode transaction by default Tx encoder
	txEncoder := b.clientCtx.TxConfig.TxEncoder()
	txBytes, err := txEncoder(tx)
	if err != nil {
		b.logger.Error("failed to encode eth tx using default encoder", "error", err.Error())
		return common.Hash{}, err
	}

	ethTx := msg.AsTransaction()

	// check the local node config in case unprotected txs are disabled
	if !b.UnprotectedAllowed() && !ethTx.Protected() {
		// Ensure only eip155 signed transactions are submitted if EIP155Required is set.
		return common.Hash{}, errors.New("only replay-protected (EIP-155) transactions allowed over RPC")
	}

	txHash := ethTx.Hash()

	// Broadcast transaction in sync mode (default)
	// NOTE: If error is encountered on the node, the broadcast will not return an error
	syncCtx := b.clientCtx.WithBroadcastMode(flags.BroadcastSync)
	rsp, err := syncCtx.BroadcastTx(txBytes)
	if rsp != nil && rsp.Code != 0 {
		err = sdkerrors.ABCIError(rsp.Codespace, rsp.Code, rsp.RawLog)
	}
	if err != nil {
		b.logger.Error("failed to broadcast tx", "error", err.Error())
		return txHash, err
	}

	// Return transaction hash
	return txHash, nil
}

// SetEtherbase sets the etherbase of the miner
func (b *Backend) SetEtherbase(etherbase common.Address) bool {
	delAddr, err := b.GetCoinbase()
	if err != nil {
		b.logger.Debug("failed to get coinbase address", "error", err.Error())
		return false
	}

	withdrawAddr := sdk.AccAddress(etherbase.Bytes())
	msg := distributiontypes.NewMsgSetWithdrawAddress(delAddr, withdrawAddr)

	if err := msg.ValidateBasic(); err != nil {
		b.logger.Debug("tx failed basic validation", "error", err.Error())
		return false
	}

	// Assemble transaction from fields
	builder, ok := b.clientCtx.TxConfig.NewTxBuilder().(authtx.ExtensionOptionsTxBuilder)
	if !ok {
		b.logger.Debug("clientCtx.TxConfig.NewTxBuilder returns unsupported builder", "error", err.Error())
		return false
	}

	err = builder.SetMsgs(msg)
	if err != nil {
		b.logger.Error("builder.SetMsgs failed", "error", err.Error())
		return false
	}

	// Fetch minimun gas price to calculate fees using the configuration.
	minGasPrices := b.cfg.GetMinGasPrices()
	if len(minGasPrices) == 0 || minGasPrices.Empty() {
		b.logger.Debug("the minimun fee is not set")
		return false
	}
	minGasPriceValue := minGasPrices[0].Amount
	denom := minGasPrices[0].Denom

	delCommonAddr := common.BytesToAddress(delAddr.Bytes())
	nonce, err := b.GetTransactionCount(delCommonAddr, rpctypes.EthPendingBlockNumber)
	if err != nil {
		b.logger.Debug("failed to get nonce", "error", err.Error())
		return false
	}

	txFactory := tx.Factory{}
	txFactory = txFactory.
		WithChainID(b.clientCtx.ChainID).
		WithKeybase(b.clientCtx.Keyring).
		WithTxConfig(b.clientCtx.TxConfig).
		WithSequence(uint64(*nonce)).
		WithGasAdjustment(1.25)

	_, gas, err := tx.CalculateGas(b.clientCtx, txFactory, msg)
	if err != nil {
		b.logger.Debug("failed to calculate gas", "error", err.Error())
		return false
	}

	txFactory = txFactory.WithGas(gas)

	value := new(big.Int).SetUint64(gas * minGasPriceValue.Ceil().TruncateInt().Uint64())
	fees := sdk.Coins{sdk.NewCoin(denom, sdkmath.NewIntFromBigInt(value))}
	builder.SetFeeAmount(fees)
	builder.SetGasLimit(gas)

	keyInfo, err := b.clientCtx.Keyring.KeyByAddress(delAddr)
	if err != nil {
		b.logger.Debug("failed to get the wallet address using the keyring", "error", err.Error())
		return false
	}

	if err := tx.Sign(txFactory, keyInfo.Name, builder, false); err != nil {
		b.logger.Debug("failed to sign tx", "error", err.Error())
		return false
	}

	// Encode transaction by default Tx encoder
	txEncoder := b.clientCtx.TxConfig.TxEncoder()
	txBytes, err := txEncoder(builder.GetTx())
	if err != nil {
		b.logger.Debug("failed to encode eth tx using default encoder", "error", err.Error())
		return false
	}

	tmHash := common.BytesToHash(tmtypes.Tx(txBytes).Hash())

	// Broadcast transaction in sync mode (default)
	// NOTE: If error is encountered on the node, the broadcast will not return an error
	syncCtx := b.clientCtx.WithBroadcastMode(flags.BroadcastSync)
	rsp, err := syncCtx.BroadcastTx(txBytes)
	if rsp != nil && rsp.Code != 0 {
		err = sdkerrors.ABCIError(rsp.Codespace, rsp.Code, rsp.RawLog)
	}
	if err != nil {
		b.logger.Debug("failed to broadcast tx", "error", err.Error())
		return false
	}

	b.logger.Debug("broadcasted tx to set miner withdraw address (etherbase)", "hash", tmHash.String())
	return true
}

// ImportRawKey armors and encrypts a given raw hex encoded ECDSA key and stores it into the key directory.
// The name of the key will have the format "personal_<length-keys>", where <length-keys> is the total number of
// keys stored on the keyring.
//
// NOTE: The key will be both armored and encrypted using the same passphrase.
func (b *Backend) ImportRawKey(privkey, password string) (common.Address, error) {
	priv, err := crypto.HexToECDSA(privkey)
	if err != nil {
		return common.Address{}, err
	}

	privKey := &ethsecp256k1.PrivKey{Key: crypto.FromECDSA(priv)}

	addr := sdk.AccAddress(privKey.PubKey().Address().Bytes())
	ethereumAddr := common.BytesToAddress(addr)

	// return if the key has already been imported
	if _, err := b.clientCtx.Keyring.KeyByAddress(addr); err == nil {
		return ethereumAddr, nil
	}

	// ignore error as we only care about the length of the list
	list, _ := b.clientCtx.Keyring.List()
	privKeyName := fmt.Sprintf("personal_%d", len(list))

	armor := sdkcrypto.EncryptArmorPrivKey(privKey, password, ethsecp256k1.KeyType)

	if err := b.clientCtx.Keyring.ImportPrivKey(privKeyName, armor, password); err != nil {
		return common.Address{}, err
	}

	b.logger.Info("key successfully imported", "name", privKeyName, "address", ethereumAddr.String())

	return ethereumAddr, nil
}

// ListAccounts will return a list of addresses for accounts this node manages.
func (b *Backend) ListAccounts() ([]common.Address, error) {
	addrs := []common.Address{}

	list, err := b.clientCtx.Keyring.List()
	if err != nil {
		return nil, err
	}

	for _, info := range list {
		pubKey, err := info.GetPubKey()
		if err != nil {
			return nil, err
		}
		addrs = append(addrs, common.BytesToAddress(pubKey.Address()))
	}

	return addrs, nil
}

// NewAccount will create a new account and returns the address for the new account.
func (b *Backend) NewMnemonic(uid string, language keyring.Language, hdPath, bip39Passphrase string, algo keyring.SignatureAlgo) (*keyring.Record, error) {
	info, _, err := b.clientCtx.Keyring.NewMnemonic(uid, keyring.English, bip39Passphrase, bip39Passphrase, algo)
	if err != nil {
		return nil, err
	}
	return info, err
}

// UnprotectedAllowed returns the node configuration value for allowing
// unprotected transactions (i.e not replay-protected)
func (b Backend) UnprotectedAllowed() bool {
	return b.allowUnprotectedTxs
}
