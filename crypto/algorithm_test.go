package crypto

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

	ethermint "github.com/cosmos/ethermint/types"
)

func TestEthermintKeygenFunc(t *testing.T) {
	privkey, err := GenerateKey()
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

	privkey := PrivKeySecp256k1(bz)
	addr := common.BytesToAddress(privkey.PubKey().Address().Bytes())

	wallet, err := hdwallet.NewFromMnemonic(mnemonic)
	require.NoError(t, err)

	path := hdwallet.MustParseDerivationPath(hdPath)

	account, err := wallet.Derive(path, false)
	require.NoError(t, err)
	require.Equal(t, addr.String(), account.Address.String())
}
