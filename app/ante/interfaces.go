package ante

import (
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"
	tx "github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/params"
	evmtypes "github.com/tharsis/ethermint/x/evm/types"
)

// EVMKeeper defines the expected keeper interface used on the Eth AnteHandler
type EVMKeeper interface {
	vm.StateDB

	ChainID() *big.Int
	GetParams(ctx sdk.Context) evmtypes.Params
	WithContext(ctx sdk.Context)
	ResetRefundTransient(ctx sdk.Context)
	NewEVM(msg core.Message, cfg *evmtypes.EVMConfig, tracer vm.Tracer) *vm.EVM
	GetCodeHash(addr common.Address) common.Hash
	DeductTxCostsFromUserBalance(
		ctx sdk.Context, msgEthTx evmtypes.MsgEthereumTx, txData evmtypes.TxData, denom string, homestead, istanbul, london bool,
	) (sdk.Coins, error)
	BaseFee(ctx sdk.Context, ethCfg *params.ChainConfig) *big.Int
}

type protoTxProvider interface {
	GetProtoTx() *tx.Tx
}
