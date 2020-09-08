package types

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"

	tmamino "github.com/tendermint/tendermint/crypto/encoding/amino"
	"github.com/tendermint/tendermint/crypto/secp256k1"

	emintcrypto "github.com/cosmos/ethermint/crypto"
)

func init() {
	tmamino.RegisterKeyType(emintcrypto.PubKeySecp256k1{}, emintcrypto.PubKeyAminoName)
	tmamino.RegisterKeyType(emintcrypto.PrivKeySecp256k1{}, emintcrypto.PrivKeyAminoName)
}

func TestEthermintAccountJSON(t *testing.T) {
	pubkey := secp256k1.GenPrivKey().PubKey()
	addr := sdk.AccAddress(pubkey.Address())
	balance := sdk.NewCoins(NewPhotonCoin(sdk.OneInt()))
	baseAcc := auth.NewBaseAccount(addr, balance, pubkey, 10, 50)
	ethAcc := EthAccount{BaseAccount: baseAcc, CodeHash: []byte{1, 2}}

	bz, err := json.Marshal(ethAcc)
	require.NoError(t, err)

	bz1, err := ethAcc.MarshalJSON()
	require.NoError(t, err)
	require.Equal(t, string(bz1), string(bz))

	var a EthAccount
	require.NoError(t, a.UnmarshalJSON(bz))
	require.Equal(t, ethAcc.String(), a.String())
	require.Equal(t, ethAcc.PubKey, a.PubKey)
}

func TestEthermintPubKeyJSON(t *testing.T) {
	privkey, err := emintcrypto.GenerateKey()
	require.NoError(t, err)
	bz := privkey.PubKey().Bytes()

	pubk, err := tmamino.PubKeyFromBytes(bz)
	require.NoError(t, err)
	require.Equal(t, pubk, privkey.PubKey())
}

func TestSecpPubKeyJSON(t *testing.T) {
	pubkey := secp256k1.GenPrivKey().PubKey()
	bz := pubkey.Bytes()

	pubk, err := tmamino.PubKeyFromBytes(bz)
	require.NoError(t, err)
	require.Equal(t, pubk, pubkey)
}

func TestEthermintAccount_String(t *testing.T) {
	pubkey := secp256k1.GenPrivKey().PubKey()
	addr := sdk.AccAddress(pubkey.Address())
	balance := sdk.NewCoins(NewPhotonCoin(sdk.OneInt()))
	baseAcc := auth.NewBaseAccount(addr, balance, pubkey, 10, 50)
	ethAcc := EthAccount{BaseAccount: baseAcc, CodeHash: []byte{1, 2}}

	config := sdk.GetConfig()
	SetBech32Prefixes(config)

	bech32pubkey, err := sdk.Bech32ifyPubKey(sdk.Bech32PubKeyTypeAccPub, pubkey)
	require.NoError(t, err)

	accountStr := fmt.Sprintf(`|
  address: %s
  eth_address: %s
  coins:
  - denom: aphoton
    amount: "1"
  public_key: %s
  account_number: 10
  sequence: 50
  code_hash: "0102"
`, addr, ethAcc.EthAddress().String(), bech32pubkey)

	require.Equal(t, accountStr, ethAcc.String())

	i, err := ethAcc.MarshalYAML()
	require.NoError(t, err)

	var ok bool
	accountStr, ok = i.(string)
	require.True(t, ok)
	require.Contains(t, accountStr, addr.String())
	require.Contains(t, accountStr, bech32pubkey)
}

func TestEthermintAccount_MarshalJSON(t *testing.T) {
	pubkey := secp256k1.GenPrivKey().PubKey()
	addr := sdk.AccAddress(pubkey.Address())
	balance := sdk.NewCoins(NewPhotonCoin(sdk.OneInt()))
	baseAcc := auth.NewBaseAccount(addr, balance, pubkey, 10, 50)
	ethAcc := &EthAccount{BaseAccount: baseAcc, CodeHash: []byte{1, 2}}

	bz, err := ethAcc.MarshalJSON()
	require.NoError(t, err)

	res := new(EthAccount)
	err = res.UnmarshalJSON(bz)
	require.NoError(t, err)
	require.Equal(t, ethAcc, res)
}
