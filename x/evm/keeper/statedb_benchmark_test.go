package keeper_test

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/tharsis/ethermint/tests"
)

func BenchmarkCreateAccountNew(b *testing.B) {
	suite := KeeperTestSuite{}
	suite.DoSetupTest(b)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		b.StopTimer()
		addr := tests.GenerateAddress()
		b.StartTimer()
		suite.app.EvmKeeper.CreateAccount(addr)
	}
}

func BenchmarkCreateAccountExisting(b *testing.B) {
	suite := KeeperTestSuite{}
	suite.DoSetupTest(b)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		suite.app.EvmKeeper.CreateAccount(suite.address)
	}
}

func BenchmarkAddBalance(b *testing.B) {
	suite := KeeperTestSuite{}
	suite.DoSetupTest(b)

	amt := big.NewInt(10)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		suite.app.EvmKeeper.AddBalance(suite.address, amt)
	}
}

func BenchmarkSetCode(b *testing.B) {
	suite := KeeperTestSuite{}
	suite.DoSetupTest(b)

	hash := crypto.Keccak256Hash([]byte("code")).Bytes()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		suite.app.EvmKeeper.SetCode(suite.address, hash)
	}
}
