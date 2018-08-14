package handlers

import (
	"testing"
	"math/big"
	"crypto/ecdsa"

	"github.com/cosmos/ethermint/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	abci "github.com/tendermint/tendermint/abci/types"

	ethcmn "github.com/ethereum/go-ethereum/common"

	"github.com/stretchr/testify/require"
)

func TestBadSig(t *testing.T) {
	tx := types.NewTestEthTxs(types.TestChainID, []*ecdsa.PrivateKey{types.TestPrivKey1}, []ethcmn.Address{types.TestAddr1})[0]

	tx.Data.Signature = types.NewEthSignature(new(big.Int), new(big.Int), new(big.Int))

	ms, key := SetupMultiStore()
	ctx := sdk.NewContext(ms, abci.Header{ChainID: types.TestChainID.String()}, false, nil)

	accountMapper := auth.NewAccountMapper(types.NewTestCodec(), key, auth.ProtoBaseAccount)
	handler := AnteHandler(accountMapper)

	_, res, abort := handler(ctx, tx)

	require.True(t, abort, "Antehandler did not abort")
	require.Equal(t, sdk.ABCICodeType(0x10004), res.Code, "Result is OK on bad tx")

}

func TestInsufficientGas(t *testing.T) {
	tx := types.NewTestEthTxs(types.TestChainID, []*ecdsa.PrivateKey{types.TestPrivKey1}, []ethcmn.Address{types.TestAddr1})[0]

	tx.Data.GasLimit = 0

	ms, key := SetupMultiStore()
	ctx := sdk.NewContext(ms, abci.Header{ChainID: types.TestChainID.String()}, false, nil)

	accountMapper := auth.NewAccountMapper(types.NewTestCodec(), key, auth.ProtoBaseAccount)
	handler := AnteHandler(accountMapper)

	_, res, abort := handler(ctx, tx)

	require.True(t, abort, "Antehandler did not abort")
	require.Equal(t, sdk.ABCICodeType(0x1000c), res.Code, "Result is OK on bad tx")

}

func TestWrongNonce(t *testing.T) {
	tx := types.NewTestEthTxs(types.TestChainID, []*ecdsa.PrivateKey{types.TestPrivKey1}, []ethcmn.Address{types.TestAddr1})[0]

	tx.Data.AccountNonce = 12

	ms, key := SetupMultiStore()
	ctx := sdk.NewContext(ms, abci.Header{ChainID: types.TestChainID.String()}, false, nil)

	accountMapper := auth.NewAccountMapper(types.NewTestCodec(), key, auth.ProtoBaseAccount)
	handler := AnteHandler(accountMapper)

	_, res, abort := handler(ctx, tx)

	require.True(t, abort, "Antehandler did not abort")
	require.Equal(t, sdk.ABCICodeType(0x10004), res.Code, "Result is OK on bad tx")

}