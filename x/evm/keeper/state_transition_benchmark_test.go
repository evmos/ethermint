package keeper_test

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/require"
	evmtypes "github.com/tharsis/ethermint/x/evm/types"
)

func BenchmarkApplyTransaction(b *testing.B) {
	suite := KeeperTestSuite{}
	suite.DoSetupTest(b)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		nonce := suite.app.EvmKeeper.GetNonce(suite.address)
		msg := evmtypes.NewTx(suite.app.EvmKeeper.ChainID(), nonce, &common.Address{}, big.NewInt(100), 21000, big.NewInt(1), nil, nil)
		msg.From = suite.address.Hex()
		err := msg.Sign(ethtypes.LatestSignerForChainID(suite.app.EvmKeeper.ChainID()), suite.signer)
		require.NoError(b, err)

		b.StartTimer()
		_, err = suite.app.EvmKeeper.ApplyTransaction(msg.AsTransaction())
		b.StopTimer()
		require.NoError(b, err)
	}
}

func BenchmarkApplyNativeMessage(b *testing.B) {
	suite := KeeperTestSuite{}
	suite.DoSetupTest(b)

	params := suite.app.EvmKeeper.GetParams(suite.ctx)
	ethCfg := params.ChainConfig.EthereumConfig(suite.app.EvmKeeper.ChainID())

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		nonce := suite.app.EvmKeeper.GetNonce(suite.address)
		msg := evmtypes.NewTx(suite.app.EvmKeeper.ChainID(), nonce, &common.Address{}, big.NewInt(100), 21000, big.NewInt(1), nil, nil)
		msg.From = suite.address.Hex()
		err := msg.Sign(ethtypes.LatestSignerForChainID(suite.app.EvmKeeper.ChainID()), suite.signer)
		require.NoError(b, err)

		blockNum := big.NewInt(suite.ctx.BlockHeight())
		signer := ethtypes.MakeSigner(ethCfg, blockNum)

		m, err := msg.AsMessage(signer)
		require.NoError(b, err)

		b.StartTimer()
		_, err = suite.app.EvmKeeper.ApplyNativeMessage(m)
		b.StopTimer()
		require.NoError(b, err)
	}
}
