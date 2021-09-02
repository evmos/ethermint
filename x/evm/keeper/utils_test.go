package keeper_test

import (
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	evmkeeper "github.com/tharsis/ethermint/x/evm/keeper"
	evmtypes "github.com/tharsis/ethermint/x/evm/types"
)

func (suite *KeeperTestSuite) TestCheckSenderBalance() {
	hundredInt := sdk.NewInt(100)
	zeroInt := sdk.ZeroInt()
	oneInt := sdk.NewInt(1)
	//minusOneInt := sdk.NewInt(-1)

	testCases := []struct {
		name       string
		msg        string
		balance    *sdk.Int
		to         string
		gasLimit   uint64
		gasPrice   *sdk.Int
		cost       *sdk.Int
		from       string
		accessList *ethtypes.AccessList
		chainID    *sdk.Int
		expectPass bool
	}{
		{name: "Enough balance", msg: "balance should be greater than fee + amount", balance: &hundredInt, to: suite.address.String(), gasLimit: 10, gasPrice: &oneInt, cost: &oneInt, from: suite.address.String(), accessList: &ethtypes.AccessList{}, chainID: &zeroInt, expectPass: true},
	}

	for i, tc := range testCases {
		suite.Run(tc.name, func() {
			// prev := suite.app.EvmKeeper.GetBalance(suite.address)
			suite.app.EvmKeeper.AddBalance(suite.address, tc.balance.BigInt())
			// post := suite.app.EvmKeeper.GetBalance(suite.address)

			to := common.HexToAddress(tc.from)

			var chainID, amount, gasPrice *big.Int
			if tc.chainID != nil {
				chainID = tc.chainID.BigInt()
			}
			if tc.cost != nil {
				amount = tc.cost.BigInt()
			}
			if tc.gasPrice != nil {
				gasPrice = tc.gasPrice.BigInt()
			}

			tx := evmtypes.NewTx(chainID, 1, &to, amount, tc.gasLimit, gasPrice, nil, tc.accessList)
			tx.From = tc.from

			txData, _ := evmtypes.UnpackTxData(tx.Data)

			err := evmkeeper.CheckSenderBalance(suite.app.EvmKeeper.Ctx(), suite.app.BankKeeper, suite.address[:], txData, "aphoton")

			if tc.expectPass {
				suite.Require().NoError(err, "valid test %d failed: %s", i, tc.msg)
			} else {
				suite.Require().Error(err, "invalid test %d passed: %s", i, tc.msg)
			}

		})
	}
}
