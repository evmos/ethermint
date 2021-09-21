package keeper_test

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"

	"github.com/tendermint/tendermint/crypto/tmhash"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	tmtypes "github.com/tendermint/tendermint/types"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/tharsis/ethermint/tests"
)

func (suite *KeeperTestSuite) TestGetHashFn() {
	header := suite.ctx.BlockHeader()
	h, _ := tmtypes.HeaderFromProto(&header)
	hash := h.Hash()

	testCases := []struct {
		msg      string
		height   uint64
		malleate func()
		expHash  common.Hash
	}{
		{
			"case 1.1: context hash cached",
			uint64(suite.ctx.BlockHeight()),
			func() {
				suite.ctx = suite.ctx.WithHeaderHash(tmhash.Sum([]byte("header")))
				suite.app.EvmKeeper.WithContext(suite.ctx)
			},
			common.BytesToHash(tmhash.Sum([]byte("header"))),
		},
		{
			"case 1.2: failed to cast Tendermint header",
			uint64(suite.ctx.BlockHeight()),
			func() {
				header := tmproto.Header{}
				header.Height = suite.ctx.BlockHeight()
				suite.ctx = suite.ctx.WithBlockHeader(header)
				suite.app.EvmKeeper.WithContext(suite.ctx)
			},
			common.Hash{},
		},
		{
			"case 1.3: hash calculated from Tendermint header",
			uint64(suite.ctx.BlockHeight()),
			func() {
				suite.ctx = suite.ctx.WithBlockHeader(header)
				suite.app.EvmKeeper.WithContext(suite.ctx)
			},
			common.BytesToHash(hash),
		},
		{
			"case 2.1: height lower than current one, hist info not found",
			1,
			func() {
				suite.ctx = suite.ctx.WithBlockHeight(10)
				suite.app.EvmKeeper.WithContext(suite.ctx)
			},
			common.Hash{},
		},
		{
			"case 2.2: height lower than current one, invalid hist info header",
			1,
			func() {
				suite.app.StakingKeeper.SetHistoricalInfo(suite.ctx, 1, &stakingtypes.HistoricalInfo{})
				suite.ctx = suite.ctx.WithBlockHeight(10)
				suite.app.EvmKeeper.WithContext(suite.ctx)
			},
			common.Hash{},
		},
		{
			"case 2.3: height lower than current one, calculated from hist info header",
			1,
			func() {
				histInfo := &stakingtypes.HistoricalInfo{
					Header: header,
				}
				suite.app.StakingKeeper.SetHistoricalInfo(suite.ctx, 1, histInfo)
				suite.ctx = suite.ctx.WithBlockHeight(10)
				suite.app.EvmKeeper.WithContext(suite.ctx)
			},
			common.BytesToHash(hash),
		},
		{
			"case 3: height greater than current one",
			200,
			func() {},
			common.Hash{},
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			suite.SetupTest() // reset

			tc.malleate()

			hash := suite.app.EvmKeeper.GetHashFn()(tc.height)
			suite.Require().Equal(tc.expHash, hash)
		})
	}
}

func (suite *KeeperTestSuite) TestGetCoinbaseAddress() {
	valOpAddr := tests.GenerateAddress()

	testCases := []struct {
		msg      string
		malleate func()
		expPass  bool
	}{
		{
			"validator not found",
			func() {
				header := suite.ctx.BlockHeader()
				header.ProposerAddress = []byte{}
				suite.ctx = suite.ctx.WithBlockHeader(header)
				suite.app.EvmKeeper.WithContext(suite.ctx)
			},
			false,
		},
		{
			"success",
			func() {
				valConsAddr, privkey := tests.NewAddrKey()

				pkAny, err := codectypes.NewAnyWithValue(privkey.PubKey())
				suite.Require().NoError(err)

				validator := stakingtypes.Validator{
					OperatorAddress: sdk.ValAddress(valOpAddr.Bytes()).String(),
					ConsensusPubkey: pkAny,
				}

				suite.app.StakingKeeper.SetValidator(suite.ctx, validator)
				err = suite.app.StakingKeeper.SetValidatorByConsAddr(suite.ctx, validator)
				suite.Require().NoError(err)

				header := suite.ctx.BlockHeader()
				header.ProposerAddress = valConsAddr.Bytes()
				suite.ctx = suite.ctx.WithBlockHeader(header)

				validator, found := suite.app.StakingKeeper.GetValidatorByConsAddr(suite.ctx, valConsAddr.Bytes())
				suite.Require().True(found)

				suite.app.EvmKeeper.WithContext(suite.ctx)
				suite.Require().NotEmpty(suite.ctx.BlockHeader().ProposerAddress)
			},
			true,
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			suite.SetupTest() // reset

			tc.malleate()

			coinbase, err := suite.app.EvmKeeper.GetCoinbaseAddress(suite.ctx)
			if tc.expPass {
				suite.Require().NoError(err)
				suite.Require().Equal(valOpAddr, coinbase)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}
