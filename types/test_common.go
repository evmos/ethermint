// nolint
package types

import (
	"crypto/ecdsa"
	"math/big"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	ethcmn "github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
)

// test variables
var (
	TestSDKAddr = GenerateEthAddress()
	TestChainID = big.NewInt(3)

	TestPrivKey1, _ = ethcrypto.GenerateKey()
	TestPrivKey2, _ = ethcrypto.GenerateKey()

	TestAddr1 = PrivKeyToEthAddress(TestPrivKey1)
	TestAddr2 = PrivKeyToEthAddress(TestPrivKey2)
)

func NewTestCodec() *codec.Codec {
	cdc := codec.New()

	RegisterCodec(cdc)
	auth.RegisterCodec(cdc)
	codec.RegisterCrypto(cdc)
	cdc.RegisterConcrete(&sdk.TestMsg{}, "test/TestMsg", nil)

	return cdc
}

func NewTestStdFee() auth.StdFee {
	return auth.NewStdFee(5000, sdk.NewCoin("photon", sdk.NewInt(150)))
}

func NewTestStdTx(
	chainID *big.Int, msgs []sdk.Msg, accNums, seqs []int64, pKeys []*ecdsa.PrivateKey, fee auth.StdFee,
) sdk.Tx {

	sigs := make([]auth.StdSignature, len(pKeys))

	for i, priv := range pKeys {
		signBytes := GetStdTxSignBytes(chainID.String(), accNums[i], seqs[i], NewTestStdFee(), msgs, "")

		sig, err := ethcrypto.Sign(signBytes, priv)
		if err != nil {
			panic(err)
		}

		sigs[i] = auth.StdSignature{Signature: sig, AccountNumber: accNums[i], Sequence: seqs[i]}
	}

	return auth.NewStdTx(msgs, fee, sigs, "")
}

func NewTestGethTxs(
	chainID *big.Int, seqs []int64, addrs []ethcmn.Address, pKeys []*ecdsa.PrivateKey,
) []*ethtypes.Transaction {

	txs := make([]*ethtypes.Transaction, len(pKeys))

	for i, privKey := range pKeys {
		ethTx := ethtypes.NewTransaction(
			uint64(seqs[i]), addrs[i], big.NewInt(10), 1000, big.NewInt(100), []byte{},
		)

		signer := ethtypes.NewEIP155Signer(chainID)

		ethTx, err := ethtypes.SignTx(ethTx, signer, privKey)
		if err != nil {
			panic(err)
		}

		txs[i] = ethTx
	}

	return txs
}

func NewTestEthTxs(
	chainID *big.Int, seqs []int64, addrs []ethcmn.Address, pKeys []*ecdsa.PrivateKey,
) []*Transaction {

	txs := make([]*Transaction, len(pKeys))

	for i, privKey := range pKeys {
		ethTx := NewTransaction(
			uint64(seqs[i]), addrs[i], big.NewInt(10), 1000, big.NewInt(100), []byte{},
		)

		ethTx.Sign(chainID, privKey)
		txs[i] = ethTx
	}

	return txs
}
