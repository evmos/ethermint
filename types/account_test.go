package types_test

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	ethermintcodec "github.com/cosmos/ethermint/codec"
	cryptocodec "github.com/cosmos/ethermint/crypto/codec"
	"github.com/cosmos/ethermint/crypto/ethsecp256k1"
	"github.com/cosmos/ethermint/types"
)

func init() {
	amino := codec.NewLegacyAmino()
	cryptocodec.RegisterCrypto(amino)
}

type AccountTestSuite struct {
	suite.Suite

	account *types.EthAccount
	cdc     codec.JSONMarshaler
}

func (suite *AccountTestSuite) SetupTest() {
	privKey, err := ethsecp256k1.GenerateKey()
	suite.Require().NoError(err)
	pubKey := privKey.PubKey()
	addr := sdk.AccAddress(pubKey.Address())
	baseAcc := authtypes.NewBaseAccount(addr, pubKey, 10, 50)
	suite.account = &types.EthAccount{
		BaseAccount: baseAcc,
		CodeHash:    []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 2},
	}

	interfaceRegistry := codectypes.NewInterfaceRegistry()
	ethermintcodec.RegisterInterfaces(interfaceRegistry)
	suite.cdc = codec.NewProtoCodec(interfaceRegistry)
}

func TestAccountTestSuite(t *testing.T) {
	suite.Run(t, new(AccountTestSuite))
}

func (suite *AccountTestSuite) TestEthermintAccountJSON() {
	bz, err := json.Marshal(suite.account)
	suite.Require().NoError(err)

	bz1, err := suite.account.MarshalJSON()
	suite.Require().NoError(err)
	suite.Require().Equal(string(bz1), string(bz))

	var a types.EthAccount
	suite.Require().NoError(a.UnmarshalJSON(bz1))
	suite.Require().Equal(suite.account.String(), a.String())
	suite.Require().Equal(suite.account.GetPubKey(), a.GetPubKey())
}

func (suite *AccountTestSuite) TestEthermintAccount_String() {
	config := sdk.GetConfig()
	types.SetBech32Prefixes(config)

	bech32pubkey, err := sdk.Bech32ifyPubKey(sdk.Bech32PubKeyTypeAccPub, suite.account.GetPubKey())
	suite.Require().NoError(err)

	accountStr := fmt.Sprintf(`|
  address: %s
  eth_address: %s
  public_key: %s
  account_number: 10
  sequence: 50
  code_hash: "0000000000000000000000000000000000000000000000000000000000000102"
`, suite.account.Address, suite.account.EthAddress().String(), bech32pubkey)

	suite.Require().Equal(accountStr, suite.account.String())

	i, err := suite.account.MarshalYAML()
	suite.Require().NoError(err)

	var ok bool
	accountStr, ok = i.(string)
	suite.Require().True(ok)
	suite.Require().Contains(accountStr, suite.account.Address)
	suite.Require().Contains(accountStr, bech32pubkey)
}

func (suite *AccountTestSuite) TestEthermintAccount_MarshalJSON() {
	bz, err := json.Marshal(suite.account)
	suite.Require().NoError(err)

	bz1, err := suite.account.MarshalJSON()
	suite.Require().NoError(err)
	suite.Require().Equal(string(bz1), string(bz))

	var a *types.EthAccount
	suite.Require().NoError(json.Unmarshal(bz, &a))
	suite.Require().Equal(suite.account.String(), a.String())

	bech32pubkey, err := sdk.Bech32ifyPubKey(sdk.Bech32PubKeyTypeAccPub, suite.account.GetPubKey())
	suite.Require().NoError(err)

	// test that the sdk.AccAddress is populated from the hex address
	jsonAcc := fmt.Sprintf(
		`{"address":"","eth_address":"%s","public_key":"%s","account_number":10,"sequence":50,"code_hash":"0102"}`,
		suite.account.EthAddress().String(), bech32pubkey,
	)

	res := new(types.EthAccount)
	err = res.UnmarshalJSON([]byte(jsonAcc))
	suite.Require().NoError(err)
	suite.Require().Equal(suite.account.Address, res.Address)

	jsonAcc = fmt.Sprintf(
		`{"address":"","eth_address":"","public_key":"%s","account_number":10,"sequence":50,"code_hash":"0102"}`,
		bech32pubkey,
	)

	res = new(types.EthAccount)
	err = res.UnmarshalJSON([]byte(jsonAcc))
	suite.Require().Error(err, "should fail if both address are empty")

	// test that the sdk.AccAddress is populated from the hex address
	jsonAcc = fmt.Sprintf(
		`{"address": "%s","eth_address":"0x0000000000000000000000000000000000000000","public_key":"%s","account_number":10,"sequence":50,"code_hash":"0102"}`,
		suite.account.Address, bech32pubkey,
	)

	res = new(types.EthAccount)
	err = res.UnmarshalJSON([]byte(jsonAcc))
	suite.Require().Error(err, "should fail if addresses mismatch")
}
