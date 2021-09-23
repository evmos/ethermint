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
	oneInt := sdk.OneInt()
	fiveInt := sdk.NewInt(5)
	fiftyInt := sdk.NewInt(50)
	negInt := sdk.NewInt(-10)

	testCases := []struct {
		name       string
		to         string
		gasLimit   uint64
		gasPrice   *sdk.Int
		cost       *sdk.Int
		from       string
		accessList *ethtypes.AccessList
		expectPass bool
	}{
		{
			name:       "Enough balance",
			to:         suite.address.String(),
			gasLimit:   10,
			gasPrice:   &oneInt,
			cost:       &oneInt,
			from:       suite.address.String(),
			accessList: &ethtypes.AccessList{},
			expectPass: true,
		},
		{
			name:       "Equal balance",
			to:         suite.address.String(),
			gasLimit:   99,
			gasPrice:   &oneInt,
			cost:       &oneInt,
			from:       suite.address.String(),
			accessList: &ethtypes.AccessList{},
			expectPass: true,
		},
		{
			name:       "negative cost",
			to:         suite.address.String(),
			gasLimit:   1,
			gasPrice:   &oneInt,
			cost:       &negInt,
			from:       suite.address.String(),
			accessList: &ethtypes.AccessList{},
			expectPass: false,
		},
		{
			name:       "Higher gas limit, not enough balance",
			to:         suite.address.String(),
			gasLimit:   100,
			gasPrice:   &oneInt,
			cost:       &oneInt,
			from:       suite.address.String(),
			accessList: &ethtypes.AccessList{},
			expectPass: false,
		},
		{
			name:       "Higher gas price, enough balance",
			to:         suite.address.String(),
			gasLimit:   10,
			gasPrice:   &fiveInt,
			cost:       &oneInt,
			from:       suite.address.String(),
			accessList: &ethtypes.AccessList{},
			expectPass: true,
		},
		{
			name:       "Higher gas price, not enough balance",
			to:         suite.address.String(),
			gasLimit:   20,
			gasPrice:   &fiveInt,
			cost:       &oneInt,
			from:       suite.address.String(),
			accessList: &ethtypes.AccessList{},
			expectPass: false,
		},
		{
			name:       "Higher cost, enough balance",
			to:         suite.address.String(),
			gasLimit:   10,
			gasPrice:   &fiveInt,
			cost:       &fiftyInt,
			from:       suite.address.String(),
			accessList: &ethtypes.AccessList{},
			expectPass: true,
		},
		{
			name:       "Higher cost, not enough balance",
			to:         suite.address.String(),
			gasLimit:   10,
			gasPrice:   &fiveInt,
			cost:       &hundredInt,
			from:       suite.address.String(),
			accessList: &ethtypes.AccessList{},
			expectPass: false,
		},
	}

	suite.app.EvmKeeper.AddBalance(suite.address, hundredInt.BigInt())
	balance := suite.app.EvmKeeper.GetBalance(suite.address)
	suite.Require().Equal(balance, hundredInt.BigInt())

	for i, tc := range testCases {
		suite.Run(tc.name, func() {
			to := common.HexToAddress(tc.from)

			var amount, gasPrice *big.Int
			if tc.cost != nil {
				amount = tc.cost.BigInt()
			}
			if tc.gasPrice != nil {
				gasPrice = tc.gasPrice.BigInt()
			}

			tx := evmtypes.NewTx(zeroInt.BigInt(), 1, &to, amount, tc.gasLimit, gasPrice, nil, tc.accessList)
			tx.From = tc.from

			txData, _ := evmtypes.UnpackTxData(tx.Data)

			err := evmkeeper.CheckSenderBalance(
				suite.app.EvmKeeper.Ctx(),
				suite.app.BankKeeper,
				suite.address[:],
				txData,
				evmtypes.DefaultEVMDenom,
			)

			if tc.expectPass {
				suite.Require().NoError(err, "valid test %d failed", i)
			} else {
				suite.Require().Error(err, "invalid test %d passed", i)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestDeductTxCostsFromUserBalance() {
	hundredInt := sdk.NewInt(100)
	zeroInt := sdk.ZeroInt()
	oneInt := sdk.NewInt(1)
	fiveInt := sdk.NewInt(5)
	fiftyInt := sdk.NewInt(50)

	testCases := []struct {
		name       string
		gasLimit   uint64
		gasPrice   *sdk.Int
		cost       *sdk.Int
		accessList *ethtypes.AccessList
		expectPass bool
	}{
		{
			name:       "Enough balance",
			gasLimit:   10,
			gasPrice:   &oneInt,
			cost:       &oneInt,
			accessList: &ethtypes.AccessList{},
			expectPass: true,
		},
		{
			name:       "Equal balance",
			gasLimit:   100,
			gasPrice:   &oneInt,
			cost:       &oneInt,
			accessList: &ethtypes.AccessList{},
			expectPass: true,
		},
		{
			name:       "Higher gas limit, not enough balance",
			gasLimit:   105,
			gasPrice:   &oneInt,
			cost:       &oneInt,
			accessList: &ethtypes.AccessList{},
			expectPass: false,
		},
		{
			name:       "Higher gas price, enough balance",
			gasLimit:   20,
			gasPrice:   &fiveInt,
			cost:       &oneInt,
			accessList: &ethtypes.AccessList{},
			expectPass: true,
		},
		{
			name:       "Higher gas price, not enough balance",
			gasLimit:   20,
			gasPrice:   &fiftyInt,
			cost:       &oneInt,
			accessList: &ethtypes.AccessList{},
			expectPass: false,
		},
		// This case is expected to be true because the fees can be deducted, but the tx
		// execution is going to fail because there is no more balance to pay the cost
		{
			name:       "Higher cost, enough balance",
			gasLimit:   100,
			gasPrice:   &oneInt,
			cost:       &fiftyInt,
			accessList: &ethtypes.AccessList{},
			expectPass: true,
		},
	}

	for i, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.SetupTest()
			suite.app.EvmKeeper.AddBalance(suite.address, hundredInt.BigInt())
			balance := suite.app.EvmKeeper.GetBalance(suite.address)
			suite.Require().Equal(balance, hundredInt.BigInt())

			var amount, gasPrice *big.Int
			if tc.cost != nil {
				amount = tc.cost.BigInt()
			}
			if tc.gasPrice != nil {
				gasPrice = tc.gasPrice.BigInt()
			}

			tx := evmtypes.NewTx(zeroInt.BigInt(), 1, &suite.address, amount, tc.gasLimit, gasPrice, nil, tc.accessList)
			tx.From = suite.address.String()

			txData, _ := evmtypes.UnpackTxData(tx.Data)

			fees, err := suite.app.EvmKeeper.DeductTxCostsFromUserBalance(
				suite.app.EvmKeeper.Ctx(),
				*tx,
				txData,
				evmtypes.DefaultEVMDenom,
				false,
				false,
			)

			if tc.expectPass {
				suite.Require().NoError(err, "valid test %d failed", i)
				suite.Require().Equal(
					fees,
					sdk.NewCoins(
						sdk.NewCoin(evmtypes.DefaultEVMDenom, tc.gasPrice.Mul(sdk.NewIntFromUint64(tc.gasLimit))),
					),
					"valid test %d failed, fee value is wrong ", i,
				)

			} else {
				suite.Require().Error(err, "invalid test %d passed", i)
				suite.Require().Nil(fees, "invalid test %d passed. fees value must be nil", i)
			}
		})
	}
}
