package keeper_test

import (
	"errors"
	"math/big"
	"testing"

	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/params"
	"github.com/stretchr/testify/require"
	evmtypes "github.com/tharsis/ethermint/x/evm/types"
)

var templateAccessListTx = &ethtypes.AccessListTx{
	GasPrice: big.NewInt(1),
	Gas:      21000,
	To:       &common.Address{},
	Value:    big.NewInt(0),
	Data:     []byte{},
}

var templateLegacyTx = &ethtypes.LegacyTx{
	GasPrice: big.NewInt(1),
	Gas:      21000,
	To:       &common.Address{},
	Value:    big.NewInt(0),
	Data:     []byte{},
}

func newSignedEthTx(
	txData ethtypes.TxData,
	nonce uint64,
	addr sdk.Address,
	krSigner keyring.Signer,
	ethSigner ethtypes.Signer,
) (*ethtypes.Transaction, error) {
	var ethTx *ethtypes.Transaction
	switch txData := txData.(type) {
	case *ethtypes.AccessListTx:
		txData.Nonce = nonce
		ethTx = ethtypes.NewTx(txData)
	case *ethtypes.LegacyTx:
		txData.Nonce = nonce
		ethTx = ethtypes.NewTx(txData)
	default:
		return nil, errors.New("unknown transaction type!")
	}

	sig, _, err := krSigner.SignByAddress(addr, ethTx.Hash().Bytes())
	if err != nil {
		return nil, err
	}

	ethTx, err = ethTx.WithSignature(ethSigner, sig)
	if err != nil {
		return nil, err
	}

	return ethTx, nil
}

func newNativeMessage(
	nonce uint64,
	blockHeight int64,
	address common.Address,
	cfg *params.ChainConfig,
	krSigner keyring.Signer,
	ethSigner ethtypes.Signer,
	isLegacy bool,
) (core.Message, error) {
	msgSigner := ethtypes.MakeSigner(cfg, big.NewInt(blockHeight))

	var ethTx *ethtypes.Transaction
	if isLegacy {
		templateLegacyTx.Nonce = nonce
		ethTx = ethtypes.NewTx(templateLegacyTx)
	} else {
		templateAccessListTx.Nonce = nonce
		ethTx = ethtypes.NewTx(templateAccessListTx)
	}

	msg := &evmtypes.MsgEthereumTx{}
	msg.FromEthereumTx(ethTx)
	msg.From = address.Hex()

	if err := msg.Sign(ethSigner, krSigner); err != nil {
		return nil, err
	}

	m, err := msg.AsMessage(msgSigner) // TODO: add DynamicFeeTx
	if err != nil {
		return nil, err
	}

	return m, nil
}

func BenchmarkApplyTransaction(b *testing.B) {
	suite := KeeperTestSuite{}
	suite.DoSetupTest(b)

	ethSigner := ethtypes.LatestSignerForChainID(suite.app.EvmKeeper.ChainID())

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		tx, err := newSignedEthTx(templateAccessListTx,
			suite.app.EvmKeeper.GetNonce(suite.address),
			sdk.AccAddress(suite.address.Bytes()),
			suite.signer,
			ethSigner,
		)
		require.NoError(b, err)

		b.StartTimer()
		resp, err := suite.app.EvmKeeper.ApplyTransaction(tx)
		b.StopTimer()

		require.NoError(b, err)
		require.False(b, resp.Failed())
	}
}

func BenchmarkApplyTransactionWithLegacyTx(b *testing.B) {
	suite := KeeperTestSuite{}
	suite.DoSetupTest(b)

	ethSigner := ethtypes.LatestSignerForChainID(suite.app.EvmKeeper.ChainID())

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		tx, err := newSignedEthTx(templateLegacyTx,
			suite.app.EvmKeeper.GetNonce(suite.address),
			sdk.AccAddress(suite.address.Bytes()),
			suite.signer,
			ethSigner,
		)
		require.NoError(b, err)

		b.StartTimer()
		resp, err := suite.app.EvmKeeper.ApplyTransaction(tx)
		b.StopTimer()

		require.NoError(b, err)
		require.False(b, resp.Failed())
	}
}

func BenchmarkApplyNativeMessage(b *testing.B) {
	suite := KeeperTestSuite{}
	suite.DoSetupTest(b)

	params := suite.app.EvmKeeper.GetParams(suite.ctx)
	ethCfg := params.ChainConfig.EthereumConfig(suite.app.EvmKeeper.ChainID())
	signer := ethtypes.LatestSignerForChainID(suite.app.EvmKeeper.ChainID())

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		b.StopTimer()

		m, err := newNativeMessage(
			suite.app.EvmKeeper.GetNonce(suite.address),
			suite.ctx.BlockHeight(),
			suite.address,
			ethCfg,
			suite.signer,
			signer,
			false,
		)
		require.NoError(b, err)

		b.StartTimer()
		resp, err := suite.app.EvmKeeper.ApplyNativeMessage(m)
		b.StopTimer()

		require.NoError(b, err)
		require.False(b, resp.Failed())
	}
}

func BenchmarkApplyNativeMessageWithLegacyTx(b *testing.B) {
	suite := KeeperTestSuite{}
	suite.DoSetupTest(b)

	params := suite.app.EvmKeeper.GetParams(suite.ctx)
	ethCfg := params.ChainConfig.EthereumConfig(suite.app.EvmKeeper.ChainID())
	signer := ethtypes.LatestSignerForChainID(suite.app.EvmKeeper.ChainID())

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		b.StopTimer()

		m, err := newNativeMessage(
			suite.app.EvmKeeper.GetNonce(suite.address),
			suite.ctx.BlockHeight(),
			suite.address,
			ethCfg,
			suite.signer,
			signer,
			true,
		)
		require.NoError(b, err)

		b.StartTimer()
		resp, err := suite.app.EvmKeeper.ApplyNativeMessage(m)
		b.StopTimer()

		require.NoError(b, err)
		require.False(b, resp.Failed())
	}
}
