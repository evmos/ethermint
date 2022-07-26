package backend

import (
	"math/big"
	"testing"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/suite"

	"github.com/evmos/ethermint/app"
	"github.com/evmos/ethermint/encoding"
	"github.com/evmos/ethermint/rpc/backend/mocks"
	rpc "github.com/evmos/ethermint/rpc/types"
	evmtypes "github.com/evmos/ethermint/x/evm/types"
)

type BackendTestSuite struct {
	suite.Suite
	backend *Backend
}

func TestBackendTestSuite(t *testing.T) {
	suite.Run(t, new(BackendTestSuite))
}

// SetupTest is executed before every BackendTestSuite test
func (suite *BackendTestSuite) SetupTest() {
	ctx := server.NewDefaultContext()
	ctx.Viper.Set("telemetry.global-labels", []interface{}{})

	encodingConfig := encoding.MakeConfig(app.ModuleBasics)
	clientCtx := client.Context{}.WithChainID("ethermint_9000-1").
		WithHeight(1).
		WithTxConfig(encodingConfig.TxConfig)

	allowUnprotectedTxs := false

	suite.backend = NewBackend(ctx, ctx.Logger, clientCtx, allowUnprotectedTxs)
	suite.backend.queryClient.QueryClient = mocks.NewQueryClient(suite.T())
	suite.backend.clientCtx.Client = mocks.NewClient(suite.T())
	suite.backend.ctx = rpc.ContextWithHeight(1)
}

// buildEthereumTx returns an example legacy Ethereum transaction
func (suite *BackendTestSuite) buildEthereumTx() (*evmtypes.MsgEthereumTx, []byte) {
	msgEthereumTx := evmtypes.NewTx(
		big.NewInt(1),
		uint64(0),
		&common.Address{},
		big.NewInt(0),
		100000,
		big.NewInt(1),
		nil,
		nil,
		nil,
		nil,
	)

	// A valid msg should have empty `From`
	msgEthereumTx.From = ""

	txBuilder := suite.backend.clientCtx.TxConfig.NewTxBuilder()
	err := txBuilder.SetMsgs(msgEthereumTx)
	suite.Require().NoError(err)

	bz, err := suite.backend.clientCtx.TxConfig.TxEncoder()(txBuilder.GetTx())
	suite.Require().NoError(err)
	return msgEthereumTx, bz
}
