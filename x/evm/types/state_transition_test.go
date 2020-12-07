package types_test

import (
	"math/big"

	abci "github.com/tendermint/tendermint/abci/types"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmos/ethermint/crypto/ethsecp256k1"
	ethermint "github.com/cosmos/ethermint/types"
	"github.com/cosmos/ethermint/x/evm/types"

	"github.com/ethereum/go-ethereum/common"
	ethcmn "github.com/ethereum/go-ethereum/common"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
)

func (suite *StateDBTestSuite) TestGetHashFn() {
	testCase := []struct {
		name         string
		height       uint64
		malleate     func()
		expEmptyHash bool
	}{
		{
			"valid hash, case 1",
			1,
			func() {
				suite.ctx = suite.ctx.WithBlockHeader(
					abci.Header{
						ChainID:        "ethermint-1",
						Height:         1,
						ValidatorsHash: []byte("val_hash"),
					},
				)
			},
			false,
		},
		{
			"case 1, nil tendermint hash",
			1,
			func() {},
			true,
		},
		{
			"valid hash, case 2",
			1,
			func() {
				suite.ctx = suite.ctx.WithBlockHeader(
					abci.Header{
						ChainID:        "ethermint-1",
						Height:         100,
						ValidatorsHash: []byte("val_hash"),
					},
				)
				hash := types.HashFromContext(suite.ctx)
				suite.stateDB.WithContext(suite.ctx).SetHeightHash(1, hash)
			},
			false,
		},
		{
			"height not found, case 2",
			1,
			func() {
				suite.ctx = suite.ctx.WithBlockHeader(
					abci.Header{
						ChainID:        "ethermint-1",
						Height:         100,
						ValidatorsHash: []byte("val_hash"),
					},
				)
			},
			true,
		},
		{
			"empty hash, case 3",
			1000,
			func() {
				suite.ctx = suite.ctx.WithBlockHeader(
					abci.Header{
						ChainID:        "ethermint-1",
						Height:         100,
						ValidatorsHash: []byte("val_hash"),
					},
				)
			},
			true,
		},
	}

	for _, tc := range testCase {
		suite.Run(tc.name, func() {
			suite.SetupTest() // reset

			tc.malleate()

			hash := types.GetHashFn(suite.ctx, suite.stateDB)(tc.height)
			if tc.expEmptyHash {
				suite.Require().Equal(common.Hash{}.String(), hash.String())
			} else {
				suite.Require().NotEqual(common.Hash{}.String(), hash.String())
			}
		})
	}
}

func (suite *StateDBTestSuite) TestTransitionDb() {
	suite.stateDB.SetNonce(suite.address, 123)

	addr := sdk.AccAddress(suite.address.Bytes())
	balance := ethermint.NewPhotonCoin(sdk.NewInt(5000))
	acc := suite.app.AccountKeeper.GetAccount(suite.ctx, addr)
	_ = acc.SetCoins(sdk.NewCoins(balance))
	suite.app.AccountKeeper.SetAccount(suite.ctx, acc)

	priv, err := ethsecp256k1.GenerateKey()
	suite.Require().NoError(err)
	recipient := ethcrypto.PubkeyToAddress(priv.ToECDSA().PublicKey)

	testCase := []struct {
		name     string
		malleate func()
		state    types.StateTransition
		expPass  bool
	}{
		{
			"passing state transition",
			func() {},
			types.StateTransition{
				AccountNonce: 123,
				Price:        big.NewInt(10),
				GasLimit:     11,
				Recipient:    &recipient,
				Amount:       big.NewInt(50),
				Payload:      []byte("data"),
				ChainID:      big.NewInt(1),
				Csdb:         suite.stateDB,
				TxHash:       &ethcmn.Hash{},
				Sender:       suite.address,
				Simulate:     suite.ctx.IsCheckTx(),
			},
			true,
		},
		{
			"contract creation",
			func() {},
			types.StateTransition{
				AccountNonce: 123,
				Price:        big.NewInt(10),
				GasLimit:     11,
				Recipient:    nil,
				Amount:       big.NewInt(10),
				Payload:      []byte("data"),
				ChainID:      big.NewInt(1),
				Csdb:         suite.stateDB,
				TxHash:       &ethcmn.Hash{},
				Sender:       suite.address,
				Simulate:     true,
			},
			true,
		},
		{
			"state transition simulation",
			func() {},
			types.StateTransition{
				AccountNonce: 123,
				Price:        big.NewInt(10),
				GasLimit:     11,
				Recipient:    &recipient,
				Amount:       big.NewInt(10),
				Payload:      []byte("data"),
				ChainID:      big.NewInt(1),
				Csdb:         suite.stateDB,
				TxHash:       &ethcmn.Hash{},
				Sender:       suite.address,
				Simulate:     true,
			},
			true,
		},
		{
			"fail by sending more than balance",
			func() {},
			types.StateTransition{
				AccountNonce: 123,
				Price:        big.NewInt(10),
				GasLimit:     11,
				Recipient:    &recipient,
				Amount:       big.NewInt(500000),
				Payload:      []byte("data"),
				ChainID:      big.NewInt(1),
				Csdb:         suite.stateDB,
				TxHash:       &ethcmn.Hash{},
				Sender:       suite.address,
				Simulate:     suite.ctx.IsCheckTx(),
			},
			false,
		},
		{
			"call disabled",
			func() {
				params := types.NewParams(ethermint.AttoPhoton, true, false)
				suite.stateDB.SetParams(params)
			},
			types.StateTransition{
				AccountNonce: 123,
				Price:        big.NewInt(10),
				GasLimit:     11,
				Recipient:    &recipient,
				Amount:       big.NewInt(50),
				Payload:      []byte("data"),
				ChainID:      big.NewInt(1),
				Csdb:         suite.stateDB,
				TxHash:       &ethcmn.Hash{},
				Sender:       suite.address,
				Simulate:     suite.ctx.IsCheckTx(),
			},
			false,
		},
		{
			"create disabled",
			func() {
				params := types.NewParams(ethermint.AttoPhoton, false, true)
				suite.stateDB.SetParams(params)
			},
			types.StateTransition{
				AccountNonce: 123,
				Price:        big.NewInt(10),
				GasLimit:     11,
				Recipient:    nil,
				Amount:       big.NewInt(50),
				Payload:      []byte("data"),
				ChainID:      big.NewInt(1),
				Csdb:         suite.stateDB,
				TxHash:       &ethcmn.Hash{},
				Sender:       suite.address,
				Simulate:     suite.ctx.IsCheckTx(),
			},
			false,
		},
		{
			"nil gas price",
			func() {
				suite.stateDB.SetParams(types.DefaultParams())
				invalidGas := sdk.DecCoins{
					{Denom: ethermint.AttoPhoton},
				}
				suite.ctx = suite.ctx.WithMinGasPrices(invalidGas)
			},
			types.StateTransition{
				AccountNonce: 123,
				Price:        big.NewInt(10),
				GasLimit:     11,
				Recipient:    &recipient,
				Amount:       big.NewInt(10),
				Payload:      []byte("data"),
				ChainID:      big.NewInt(1),
				Csdb:         suite.stateDB,
				TxHash:       &ethcmn.Hash{},
				Sender:       suite.address,
				Simulate:     suite.ctx.IsCheckTx(),
			},
			false,
		},
	}

	for _, tc := range testCase {
		tc.malleate()

		_, err = tc.state.TransitionDb(suite.ctx, types.DefaultChainConfig())

		if tc.expPass {
			suite.Require().NoError(err, tc.name)
			fromBalance := suite.app.EvmKeeper.GetBalance(suite.ctx, suite.address)
			toBalance := suite.app.EvmKeeper.GetBalance(suite.ctx, recipient)
			suite.Require().Equal(fromBalance, big.NewInt(4950), tc.name)
			suite.Require().Equal(toBalance, big.NewInt(50), tc.name)
		} else {
			suite.Require().Error(err, tc.name)
		}
	}
}
