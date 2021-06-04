package encoding_test

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"

	ethtypes "github.com/ethereum/go-ethereum/core/types"

	"github.com/cosmos/ethermint/app"
	"github.com/cosmos/ethermint/encoding"
	"github.com/cosmos/ethermint/tests"
	evmtypes "github.com/cosmos/ethermint/x/evm/types"
)

func TestTxEncoding(t *testing.T) {
	addr, key := tests.NewAddrKey()
	signer := tests.NewSigner(key)

	msg := evmtypes.NewMsgEthereumTxContract(big.NewInt(1), 1, big.NewInt(10), 100000, big.NewInt(1), []byte{}, nil)
	msg.From = addr.Hex()

	ethSigner := ethtypes.LatestSignerForChainID(big.NewInt(1))
	err := msg.Sign(ethSigner, signer)
	require.NoError(t, err)

	cfg := encoding.MakeConfig(app.ModuleBasics)

	bz, err := cfg.TxConfig.TxEncoder()(msg)
	require.NoError(t, err, "encoding failed")

	tx, err := cfg.TxConfig.TxDecoder()(bz)
	require.NoError(t, err, "decoding failed")
	require.IsType(t, &evmtypes.MsgEthereumTx{}, tx)
	require.Equal(t, msg.Data, tx.(*evmtypes.MsgEthereumTx).Data)

	// FIXME: transaction hashing is hardcoded on Terndermint:
	// See https://github.com/tendermint/tendermint/issues/6539 for reference
	// txHash := msg.AsTransaction().Hash()
	// tmTx := tmtypes.Tx(bz)

	// require.Equal(t, txHash.Bytes(), tmTx.Hash())
}
