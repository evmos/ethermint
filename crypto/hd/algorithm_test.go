package hd

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ethereum/go-ethereum/common"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"

	hdwallet "github.com/miguelmota/go-ethereum-hdwallet"

	"github.com/cosmos/cosmos-sdk/crypto/keys"
	"github.com/cosmos/cosmos-sdk/crypto/keys/hd"
	"github.com/cosmos/cosmos-sdk/tests"

	"github.com/cosmos/ethermint/crypto/ethsecp256k1"
	ethermint "github.com/cosmos/ethermint/types"
)

func TestEthermintKeygenFunc(t *testing.T) {
	privkey, err := ethsecp256k1.GenerateKey()
	require.NoError(t, err)

	testCases := []struct {
		name    string
		privKey []byte
		algo    keys.SigningAlgo
		expPass bool
	}{
		{
			"valid ECDSA privKey",
			ethcrypto.FromECDSA(privkey.ToECDSA()),
			EthSecp256k1,
			true,
		},
		{
			"nil bytes, valid algo",
			nil,
			EthSecp256k1,
			true,
		},
		{
			"empty bytes, valid algo",
			[]byte{},
			EthSecp256k1,
			true,
		},
		{
			"invalid algo",
			nil,
			keys.MultiAlgo,
			false,
		},
	}

	for _, tc := range testCases {
		privkey, err := EthermintKeygenFunc(tc.privKey, tc.algo)
		if tc.expPass {
			require.NoError(t, err, tc.name)
		} else {
			require.Error(t, err, tc.name)
			require.Nil(t, privkey, tc.name)
		}
	}
}

func TestKeyring(t *testing.T) {
	dir, cleanup := tests.NewTestCaseDir(t)
	mockIn := strings.NewReader("")
	t.Cleanup(cleanup)

	kr, err := keys.NewKeyring("ethermint", keys.BackendTest, dir, mockIn, EthSecp256k1Options()...)
	require.NoError(t, err)

	// fail in retrieving key
	info, err := kr.Get("foo")
	require.Error(t, err)
	require.Nil(t, info)

	mockIn.Reset("password\npassword\n")
	info, mnemonic, err := kr.CreateMnemonic("foo", keys.English, ethermint.BIP44HDPath, EthSecp256k1)
	require.NoError(t, err)
	require.NotEmpty(t, mnemonic)
	require.Equal(t, "foo", info.GetName())
	require.Equal(t, "local", info.GetType().String())
	require.Equal(t, EthSecp256k1, info.GetAlgo())

	params := *hd.NewFundraiserParams(0, ethermint.Bip44CoinType, 0)
	hdPath := params.String()

	bz, err := DeriveKey(mnemonic, keys.DefaultBIP39Passphrase, hdPath, keys.Secp256k1)
	require.NoError(t, err)
	require.NotEmpty(t, bz)

	bz, err = DeriveSecp256k1(mnemonic, keys.DefaultBIP39Passphrase, hdPath)
	require.NoError(t, err)
	require.NotEmpty(t, bz)

	bz, err = DeriveKey(mnemonic, keys.DefaultBIP39Passphrase, hdPath, keys.SigningAlgo(""))
	require.Error(t, err)
	require.Empty(t, bz)

	bz, err = DeriveSecp256k1(mnemonic, keys.DefaultBIP39Passphrase, "/wrong/hdPath")
	require.Error(t, err)
	require.Empty(t, bz)

	bz, err = DeriveKey(mnemonic, keys.DefaultBIP39Passphrase, hdPath, EthSecp256k1)
	require.NoError(t, err)
	require.NotEmpty(t, bz)

	privkey := ethsecp256k1.PrivKey(bz)
	addr := common.BytesToAddress(privkey.PubKey().Address().Bytes())

	wallet, err := hdwallet.NewFromMnemonic(mnemonic)
	require.NoError(t, err)

	path := hdwallet.MustParseDerivationPath(hdPath)

	account, err := wallet.Derive(path, false)
	require.NoError(t, err)
	require.Equal(t, addr.String(), account.Address.String())
}

func TestDerivation(t *testing.T) {
	mnemonic := "picnic rent average infant boat squirrel federal assault mercy purity very motor fossil wheel verify upset box fresh horse vivid copy predict square regret"

	bz, err := DeriveSecp256k1(mnemonic, keys.DefaultBIP39Passphrase, ethermint.BIP44HDPath)
	require.NoError(t, err)
	require.NotEmpty(t, bz)

	badBz, err := DeriveSecp256k1(mnemonic, keys.DefaultBIP39Passphrase, "44'/60'/0'/0/0")
	require.NoError(t, err)
	require.NotEmpty(t, badBz)

	require.NotEqual(t, bz, badBz)

	privkey, err := EthermintKeygenFunc(bz, EthSecp256k1)
	require.NoError(t, err)
	require.NotEmpty(t, privkey)

	badPrivKey, err := EthermintKeygenFunc(badBz, EthSecp256k1)
	require.NoError(t, err)
	require.NotEmpty(t, badPrivKey)

	require.NotEqual(t, privkey, badPrivKey)

	wallet, err := hdwallet.NewFromMnemonic(mnemonic)
	require.NoError(t, err)

	path := hdwallet.MustParseDerivationPath(ethermint.BIP44HDPath)
	account, err := wallet.Derive(path, false)
	require.NoError(t, err)

	badPath := hdwallet.MustParseDerivationPath("44'/60'/0'/0/0")
	badAccount, err := wallet.Derive(badPath, false)
	require.NoError(t, err)

	// Equality of Address BIP44
	require.Equal(t, account.Address.String(), "0xA588C66983a81e800Db4dF74564F09f91c026351")
	require.Equal(t, badAccount.Address.String(), "0xF8D6FDf2B8b488ea37e54903750dcd13F67E71cb")
	// Inequality of wrong derivation path address
	require.NotEqual(t, account.Address.String(), badAccount.Address.String())
	// Equality of Ethermint implementation
	require.Equal(t, common.BytesToAddress(privkey.PubKey().Address().Bytes()).String(), "0xA588C66983a81e800Db4dF74564F09f91c026351")
	require.Equal(t, common.BytesToAddress(badPrivKey.PubKey().Address().Bytes()).String(), "0xF8D6FDf2B8b488ea37e54903750dcd13F67E71cb")

	// Equality of Eth and Ethermint implementation
	require.Equal(t, common.BytesToAddress(privkey.PubKey().Address()).String(), account.Address.String())
	require.Equal(t, common.BytesToAddress(badPrivKey.PubKey().Address()).String(), badAccount.Address.String())

	// Inequality of wrong derivation path of Eth and Ethermint implementation
	require.NotEqual(t, common.BytesToAddress(privkey.PubKey().Address()).String(), badAccount.Address.String())
	require.NotEqual(t, common.BytesToAddress(badPrivKey.PubKey().Address()).String(), account.Address.Hex())
}
