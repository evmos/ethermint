package types

import (
	"crypto/ecdsa"
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/wire"
	"github.com/cosmos/cosmos-sdk/x/auth"
	stake "github.com/cosmos/cosmos-sdk/x/stake/types"

	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	ethcmn "github.com/ethereum/go-ethereum/common"
)

var (
	TestChainID = sdk.NewInt(3)

	TestPrivKey1, _ = ethcrypto.GenerateKey()
	TestPrivKey2, _ = ethcrypto.GenerateKey()

	TestAddr1 = PrivKeyToEthAddress(TestPrivKey1)
	TestAddr2 = PrivKeyToEthAddress(TestPrivKey2)

	TestSDKAddress = GenerateEthAddress()
)

func NewTestCodec() *wire.Codec {
	codec := wire.NewCodec()

	RegisterWire(codec)
	codec.RegisterConcrete(&sdk.TestMsg{}, "test/TestMsg", nil)

	// Register any desired SDK msgs to be embedded
	stake.RegisterWire(codec)

	return codec
}

func NewStdFee() auth.StdFee {
	return auth.NewStdFee(5000, sdk.NewCoin("photon", 150))
}

func NewTestEmbeddedTx(
	chainID sdk.Int, msgs []sdk.Msg, pKeys []*ecdsa.PrivateKey,
	accNums []int64, seqs []int64, fee auth.StdFee,
) sdk.Tx {

	sigs := make([][]byte, len(pKeys))

	for i, priv := range pKeys {
		signEtx := EmbeddedTxSign{chainID.String(), accNums[i], seqs[i], msgs, fee}

		signBytes, err := signEtx.Bytes()
		if err != nil {
			panic(err)
		}

		sig, err := ethcrypto.Sign(signBytes, priv)
		if err != nil {
			panic(err)
		}

		sigs[i] = sig
	}

	return EmbeddedTx{msgs, fee, sigs}
}

func NewTestGethTxs(chainID sdk.Int, pKeys []*ecdsa.PrivateKey, addrs []ethcmn.Address) []ethtypes.Transaction {
	txs := make([]ethtypes.Transaction, len(pKeys))

	for i, priv := range pKeys {
		ethTx := ethtypes.NewTransaction(
			uint64(i), addrs[i], big.NewInt(10), 100, big.NewInt(100), nil,
		)

		signer := ethtypes.NewEIP155Signer(chainID.BigInt())
		ethTx, _ = ethtypes.SignTx(ethTx, signer, priv)

		txs[i] = *ethTx
	}

	return txs
}

func NewTestEthTxs(chainID sdk.Int, pKeys []*ecdsa.PrivateKey, addrs []ethcmn.Address) []Transaction {
	txs := make([]Transaction, len(pKeys))

	for i, priv := range pKeys {
		emintTx := NewTransaction(
			uint64(i), addrs[i], sdk.NewInt(10), 1000, sdk.NewInt(100), nil,
		)

		emintTx.Sign(chainID, priv)

		txs[i] = emintTx
	}

	return txs
}

func NewTestSDKTxs(
	codec *wire.Codec, chainID sdk.Int, msgs []sdk.Msg, pKeys []*ecdsa.PrivateKey,
	accNums []int64, seqs []int64, fee auth.StdFee,
) []Transaction {

	txs := make([]Transaction, len(pKeys))
	etx := NewTestEmbeddedTx(chainID, msgs, pKeys, accNums, seqs, fee)

	for i, priv := range pKeys {
		payload := codec.MustMarshalBinary(etx)

		emintTx := NewTransaction(
			uint64(i), TestSDKAddress, sdk.NewInt(10), 1000,
			sdk.NewInt(100), payload,
		)

		emintTx.Sign(TestChainID, priv)

		txs[i] = emintTx
	}

	return txs
}
