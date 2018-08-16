package handlers

import (
	"fmt"
	"testing"
	"math/big"
	"crypto/ecdsa"

	"github.com/cosmos/ethermint/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	stake "github.com/cosmos/cosmos-sdk/x/stake/types"

	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"

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
	require.Equal(t, sdk.ABCICodeType(0x10004), res.Code, fmt.Sprintf("Result has wrong code on bad tx: %s", res.Log))

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
	require.Equal(t, sdk.ABCICodeType(0x1000c), res.Code, fmt.Sprintf("Result has wrong code on bad tx: %s", res.Log))

}

func TestIncorrectNonce(t *testing.T) {
	tx := types.NewTestEthTxs(types.TestChainID, []*ecdsa.PrivateKey{types.TestPrivKey1}, []ethcmn.Address{types.TestAddr1})[0]

	tx.Data.AccountNonce = 12

	ms, key := SetupMultiStore()
	ctx := sdk.NewContext(ms, abci.Header{ChainID: types.TestChainID.String()}, false, log.NewNopLogger())

	accountMapper := auth.NewAccountMapper(types.NewTestCodec(), key, auth.ProtoBaseAccount)

	// Set account in accountMapper
	acc := accountMapper.NewAccountWithAddress(ctx, types.TestAddr1[:])
	accountMapper.SetAccount(ctx, acc)

	handler := AnteHandler(accountMapper)

	fmt.Println("start test")
	fmt.Println(accountMapper.GetAccount(ctx, types.TestAddr1[:]))
	fmt.Println(types.TestAddr1)
	fmt.Println("end test")

	_, res, abort := handler(ctx, tx)

	require.True(t, abort, "Antehandler did not abort")
	require.Equal(t, sdk.ABCICodeType(0x10003), res.Code, fmt.Sprintf("Result has wrong code on bad tx: %s", res.Log))

}

func TestValidTx(t *testing.T) {
	tx := types.NewTestEthTxs(types.TestChainID, []*ecdsa.PrivateKey{types.TestPrivKey1}, []ethcmn.Address{types.TestAddr1})[0]

	ms, key := SetupMultiStore()
	ctx := sdk.NewContext(ms, abci.Header{ChainID: types.TestChainID.String()}, false, nil)

	accountMapper := auth.NewAccountMapper(types.NewTestCodec(), key, auth.ProtoBaseAccount)
	handler := AnteHandler(accountMapper)

	_, res, abort := handler(ctx, tx)

	require.False(t, abort, "Antehandler abort on valid tx")
	require.Equal(t, sdk.CodeOK, res.Code, fmt.Sprintf("Result not OK on valid Tx: %s", res.Log))

}

func TestValidEmbeddedTx(t *testing.T) {
	cdc := types.NewTestCodec()
	// Create msg to be embedded
	msgs := []sdk.Msg{stake.NewMsgDelegate(types.TestAddr1[:], types.TestAddr2[:], sdk.Coin{"steak", sdk.NewInt(50)})}

	tx := types.NewTestSDKTxs(cdc, types.TestChainID, msgs, []*ecdsa.PrivateKey{types.TestPrivKey1}, []int64{0}, []int64{0}, types.NewStdFee())[0]

	ms, key := SetupMultiStore()
	ctx := sdk.NewContext(ms, abci.Header{ChainID: types.TestChainID.String()}, false, nil)

	accountMapper := auth.NewAccountMapper(types.NewTestCodec(), key, auth.ProtoBaseAccount)
	handler := AnteHandler(accountMapper)

	_, res, abort := handler(ctx, tx)

	require.False(t, abort, "Antehandler abort on valid embedded tx")
	require.Equal(t, sdk.CodeOK, res.Code, fmt.Sprintf("Result not OK on valid Tx: %s", res.Log))
}

/*
func TestInvalidEmbeddedTx(t *testing.T) {
	cdc := types.NewTestCodec()
	// Create msg to be embedded
	msgs := []sdk.Msg{stake.NewMsgCreateValidator(types.TestAddr1[:], nil, sdk.Coin{"steak", sdk.NewInt(50)}, stake.Description{})}

	tx := types.NewTestSDKTxs(cdc, types.TestChainID, msgs, []*ecdsa.PrivateKey{types.TestPrivKey1}, []int64{0}, []int64{0}, types.NewStdFee())[0]

	ms, key := SetupMultiStore()
	ctx := sdk.NewContext(ms, abci.Header{ChainID: types.TestChainID.String()}, false, nil)

	accountMapper := auth.NewAccountMapper(types.NewTestCodec(), key, auth.ProtoBaseAccount)
	handler := AnteHandler(accountMapper)

	_, res, abort := handler(ctx, tx)

	require.True(t, abort, "Antehandler did not abort on invalid embedded tx")
	require.Equal(t, sdk.ABCICodeType(0x1000c), res.Code, "Result is OK on bad tx")
}
*/