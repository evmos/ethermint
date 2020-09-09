package types_test

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"

	tmamino "github.com/tendermint/tendermint/crypto/encoding/amino"
	"github.com/tendermint/tendermint/crypto/secp256k1"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"

	"github.com/cosmos/ethermint/crypto"
	"github.com/cosmos/ethermint/types"
)

func init() {
	tmamino.RegisterKeyType(crypto.PubKeySecp256k1{}, crypto.PubKeyAminoName)
	tmamino.RegisterKeyType(crypto.PrivKeySecp256k1{}, crypto.PrivKeyAminoName)
}

type AccountTestSuite struct {
	suite.Suite

	account *types.EthAccount
}

func (suite *AccountTestSuite) SetupTest() {
	pubkey := secp256k1.GenPrivKey().PubKey()
	addr := sdk.AccAddress(pubkey.Address())
	balance := sdk.NewCoins(types.NewPhotonCoin(sdk.OneInt()))
	baseAcc := auth.NewBaseAccount(addr, balance, pubkey, 10, 50)
	suite.account = &types.EthAccount{
		BaseAccount: baseAcc,
		CodeHash:    []byte{1, 2},
	}
}

func TestAccountTestSuite(t *testing.T) {
	suite.Run(t, new(AccountTestSuite))
}

func (suite *AccountTestSuite) TestEthAccount_Balance() {

	testCases := []struct {
		name         string
		denom        string
		initialCoins sdk.Coins
		amount       sdk.Int
	}{
		{"positive diff", types.AttoPhoton, sdk.Coins{}, sdk.OneInt()},
		{"zero diff, same coin", types.AttoPhoton, sdk.NewCoins(types.NewPhotonCoin(sdk.ZeroInt())), sdk.ZeroInt()},
		{"zero diff, other coin", sdk.DefaultBondDenom, sdk.NewCoins(types.NewPhotonCoin(sdk.ZeroInt())), sdk.ZeroInt()},
		{"negative diff", types.AttoPhoton, sdk.NewCoins(types.NewPhotonCoin(sdk.NewInt(10))), sdk.NewInt(1)},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.SetupTest() // reset values
			suite.account.SetCoins(tc.initialCoins)

			suite.account.SetBalance(tc.denom, tc.amount)
			suite.Require().Equal(tc.amount, suite.account.Balance(tc.denom))
		})
	}

}

func (suite *AccountTestSuite) TestEthermintAccountJSON() {
	bz, err := json.Marshal(suite.account)
	suite.Require().NoError(err)

	bz1, err := suite.account.MarshalJSON()
	suite.Require().NoError(err)
	suite.Require().Equal(string(bz1), string(bz))

	var a types.EthAccount
	suite.Require().NoError(a.UnmarshalJSON(bz))
	suite.Require().Equal(suite.account.String(), a.String())
	suite.Require().Equal(suite.account.PubKey, a.PubKey)
}

func (suite *AccountTestSuite) TestEthermintPubKeyJSON() {
	privkey, err := crypto.GenerateKey()
	suite.Require().NoError(err)
	bz := privkey.PubKey().Bytes()

	pubk, err := tmamino.PubKeyFromBytes(bz)
	suite.Require().NoError(err)
	suite.Require().Equal(pubk, privkey.PubKey())
}

func (suite *AccountTestSuite) TestSecpPubKeyJSON() {
	pubkey := secp256k1.GenPrivKey().PubKey()
	bz := pubkey.Bytes()

	pubk, err := tmamino.PubKeyFromBytes(bz)
	suite.Require().NoError(err)
	suite.Require().Equal(pubk, pubkey)
}

func (suite *AccountTestSuite) TestEthermintAccount_String() {
	config := sdk.GetConfig()
	types.SetBech32Prefixes(config)

	bech32pubkey, err := sdk.Bech32ifyPubKey(sdk.Bech32PubKeyTypeAccPub, suite.account.PubKey)
	suite.Require().NoError(err)

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
`, suite.account.Address, suite.account.EthAddress().String(), bech32pubkey)

	suite.Require().Equal(accountStr, suite.account.String())

	i, err := suite.account.MarshalYAML()
	suite.Require().NoError(err)

	var ok bool
	accountStr, ok = i.(string)
	suite.Require().True(ok)
	suite.Require().Contains(accountStr, suite.account.Address.String())
	suite.Require().Contains(accountStr, bech32pubkey)
}

func (suite *AccountTestSuite) TestEthermintAccount_MarshalJSON() {
	bz, err := suite.account.MarshalJSON()
	suite.Require().NoError(err)
	suite.Require().Contains(string(bz), suite.account.EthAddress().String())

	res := new(types.EthAccount)
	err = res.UnmarshalJSON(bz)
	suite.Require().NoError(err)
	suite.Require().Equal(suite.account, res)

	bech32pubkey, err := sdk.Bech32ifyPubKey(sdk.Bech32PubKeyTypeAccPub, suite.account.PubKey)
	suite.Require().NoError(err)

	// test that the sdk.AccAddress is populated from the hex address
	jsonAcc := fmt.Sprintf(
		`{"address":"","eth_address":"%s","coins":[{"denom":"aphoton","amount":"1"}],"public_key":"%s","account_number":10,"sequence":50,"code_hash":"0102"}`,
		suite.account.EthAddress().String(), bech32pubkey,
	)

	res = new(types.EthAccount)
	err = res.UnmarshalJSON([]byte(jsonAcc))
	suite.Require().NoError(err)
	suite.Require().Equal(suite.account.Address.String(), res.Address.String())

	jsonAcc = fmt.Sprintf(
		`{"address":"","eth_address":"","coins":[{"denom":"aphoton","amount":"1"}],"public_key":"%s","account_number":10,"sequence":50,"code_hash":"0102"}`,
		bech32pubkey,
	)

	res = new(types.EthAccount)
	err = res.UnmarshalJSON([]byte(jsonAcc))
	suite.Require().Error(err, "should fail if both address are empty")

	// test that the sdk.AccAddress is populated from the hex address
	jsonAcc = fmt.Sprintf(
		`{"address": "%s","eth_address":"0x0000000000000000000000000000000000000000","coins":[{"denom":"aphoton","amount":"1"}],"public_key":"%s","account_number":10,"sequence":50,"code_hash":"0102"}`,
		suite.account.Address.String(), bech32pubkey,
	)

	res = new(types.EthAccount)
	err = res.UnmarshalJSON([]byte(jsonAcc))
	suite.Require().Error(err, "should fail if addresses mismatch")
}
