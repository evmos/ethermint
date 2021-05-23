package keeper_test

// func (suite *KeeperTestSuite) TestTransitionDb() {
// 	suite.app.EvmKeeper.SetNonce(suite.address, 123)

// 	addr := sdk.AccAddress(suite.address.Bytes())
// 	balance := ethermint.NewPhotonCoin(sdk.NewInt(5000))
// 	acc := suite.app.AccountKeeper.GetAccount(suite.ctx, addr)
// 	suite.app.AccountKeeper.SetAccount(suite.ctx, acc)
// 	suite.app.BankKeeper.SetBalance(suite.ctx, addr, balance)

// 	priv, err := ethsecp256k1.GenerateKey()
// 	suite.Require().NoError(err)
// 	recipient := ethcrypto.PubkeyToAddress(priv.ToECDSA().PublicKey)

// 	testCase := []struct {
// 		name     string
// 		malleate func()
// 		expPass  bool
// 	}{
// 		{
// 			"passing state transition",
// 			func() {},
// 			types.StateTransition{
// 				Message: ethtypes.NewMessage(
// 					suite.address,
// 					&recipient,
// 					123,
// 					big.NewInt(50),
// 					11,
// 					big.NewInt(10),
// 					[]byte("data"),
// 					nil,
// 					true,
// 				),
// 				ChainID:  big.NewInt(1),
// 				TxHash:   &ethcmn.Hash{},
// 				Simulate: suite.ctx.IsCheckTx(),
// 			},
// 			true,
// 		},
// 		{
// 			"contract creation",
// 			func() {},
// 			types.StateTransition{
// 				Message: ethtypes.NewMessage(
// 					suite.address,
// 					nil,
// 					123,
// 					big.NewInt(50),
// 					11,
// 					big.NewInt(10),
// 					[]byte("data"),
// 					nil,
// 					true,
// 				),
// 				ChainID:  big.NewInt(1),
// 				TxHash:   &ethcmn.Hash{},
// 				Simulate: true,
// 			},
// 			true,
// 		},
// 		{
// 			"state transition simulation",
// 			func() {},
// 			types.StateTransition{
// 				Message: ethtypes.NewMessage(
// 					suite.address,
// 					&recipient,
// 					123,
// 					big.NewInt(50),
// 					11,
// 					big.NewInt(10),
// 					[]byte("data"),
// 					nil,
// 					true,
// 				),
// 				ChainID:  big.NewInt(1),
// 				TxHash:   &ethcmn.Hash{},
// 				Simulate: true,
// 			},
// 			true,
// 		},
// 		{
// 			"fail by sending more than balance",
// 			func() {},
// 			types.StateTransition{
// 				Message: ethtypes.NewMessage(
// 					suite.address,
// 					&recipient,
// 					123,
// 					big.NewInt(50000000),
// 					11,
// 					big.NewInt(10),
// 					[]byte("data"),
// 					nil,
// 					true,
// 				),
// 				ChainID:  big.NewInt(1),
// 				TxHash:   &ethcmn.Hash{},
// 				Simulate: suite.ctx.IsCheckTx(),
// 			},
// 			false,
// 		},
// 		{
// 			"failed to Finalize",
// 			func() {},
// 			types.StateTransition{
// 				Message: ethtypes.NewMessage(
// 					suite.address,
// 					&recipient,
// 					123,
// 					big.NewInt(-5000),
// 					11,
// 					big.NewInt(10),
// 					[]byte("data"),
// 					nil,
// 					true,
// 				),
// 				ChainID:  big.NewInt(1),
// 				TxHash:   &ethcmn.Hash{},
// 				Simulate: false,
// 			},
// 			false,
// 		},
// 		{
// 			"nil gas price",
// 			func() {
// 				invalidGas := sdk.DecCoins{
// 					{Denom: ethermint.AttoPhoton},
// 				}
// 				suite.ctx = suite.ctx.WithMinGasPrices(invalidGas)
// 			},
// 			types.StateTransition{
// 				Message: ethtypes.NewMessage(
// 					suite.address,
// 					&recipient,
// 					123,
// 					big.NewInt(50),
// 					11,
// 					nil,
// 					[]byte("data"),
// 					nil,
// 					true,
// 				),
// 				ChainID:  big.NewInt(1),
// 				TxHash:   &ethcmn.Hash{},
// 				Simulate: suite.ctx.IsCheckTx(),
// 			},
// 			false,
// 		},
// 	}

// 	for _, tc := range testCase {
// 		tc.malleate()

// 		_, err = tc.state.TransitionDb(suite.ctx, types.DefaultChainConfig())

// 		if tc.expPass {
// 			suite.Require().NoError(err, tc.name)
// 			fromBalance := suite.app.EvmKeeper.GetBalance(suite.address)
// 			toBalance := suite.app.EvmKeeper.GetBalance(recipient)
// 			suite.Require().Equal(fromBalance, big.NewInt(4950), tc.name)
// 			suite.Require().Equal(toBalance, big.NewInt(50), tc.name)
// 		} else {
// 			suite.Require().Error(err, tc.name)
// 		}
// 	}
// }
