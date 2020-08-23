package evm_test

import (
	"encoding/json"

	"github.com/cosmos/ethermint/x/evm"

	"github.com/ethereum/go-ethereum/common"
)

var testJSON = `{
      "accounts": [
        {
          "address": "0x00cabdd44664b73cfc3194b9d32eb6c351ef7652",
          "balance": 34
        },
        {
          "address": "0x2cc7fdf9fde6746731d7f11979609d455c2c197a",
          "balance": 0,
          "code": "0x60806040"
        }
      ]
 	}`

func (suite *EvmTestSuite) TestInitGenesis() {
	am := evm.NewAppModule(suite.app.EvmKeeper, suite.app.AccountKeeper)
	in := json.RawMessage([]byte(testJSON))
	_ = am.InitGenesis(suite.ctx, in)

	testAddr := common.HexToAddress("0x2cc7fdf9fde6746731d7f11979609d455c2c197a")

	res := suite.app.EvmKeeper.CommitStateDB.WithContext(suite.ctx).GetCode(testAddr)
	expectedCode := common.FromHex("0x60806040")
	suite.Require().Equal(expectedCode, res)
}
