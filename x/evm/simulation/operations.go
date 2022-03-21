package simulation

import (
	"fmt"
	"math/rand"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"
	ethermint "github.com/tharsis/ethermint/types"
	"github.com/tharsis/ethermint/x/evm/keeper"
	"github.com/tharsis/ethermint/x/evm/types"
)

const (
	OpWeightMsgEthSimpleTransfer = "op_weight_msg_eth_simple_transfer"
	OpWeightMsgEthCreateContract = "op_weight_msg_eth_create_contract"
	OpWeightMsgEthCallContract   = "op_weight_msg_eth_call_contract"
)

const (
	WeightMsgEthSimpleTransfer = 100
	WeightMsgEthCreateContract = 100
	WeightMsgEthCallContract   = 100
)

func WeightedOperations(
	appParams simtypes.AppParams, cdc codec.JSONCodec, ak types.AccountKeeper, k *keeper.Keeper) simulation.WeightedOperations {
	var (
		weightMsgEthSimpleTransfer int
		weightMsgEthCreateContract int
		weightMsgEthCallContract   int
	)

	appParams.GetOrGenerate(cdc, OpWeightMsgEthSimpleTransfer, &weightMsgEthSimpleTransfer, nil,
		func(_ *rand.Rand) {
			weightMsgEthSimpleTransfer = WeightMsgEthSimpleTransfer
		},
	)

	appParams.GetOrGenerate(cdc, OpWeightMsgEthCreateContract, &weightMsgEthCreateContract, nil,
		func(_ *rand.Rand) {
			weightMsgEthCreateContract = WeightMsgEthCreateContract
		},
	)

	appParams.GetOrGenerate(cdc, OpWeightMsgEthCallContract, &weightMsgEthCallContract, nil,
		func(_ *rand.Rand) {
			weightMsgEthCallContract = WeightMsgEthCallContract
		},
	)

	return simulation.WeightedOperations{
		simulation.NewWeightedOperation(
			weightMsgEthSimpleTransfer,
			SimulateEthSimpleTransfer(ak, k),
		),
		simulation.NewWeightedOperation(
			weightMsgEthCreateContract,
			SimulateEthCreateContract(ak, k),
		),
		simulation.NewWeightedOperation(
			weightMsgEthCallContract,
			SimulateEthCallContract(ak, k),
		),
	}
}

func SimulateEthSimpleTransfer(ak types.AccountKeeper, k *keeper.Keeper) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		simAccount, _ := simtypes.RandomAcc(r, accs)
		account := ak.GetAccount(ctx, simAccount.Address)
		if account == nil {
			err := fmt.Errorf("account not found")
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgEthereumTx, "account not found"), nil, err
		}
		_, ok := account.(*ethermint.EthAccount)
		if !ok {
			err := fmt.Errorf("not EthAccount")
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgEthereumTx, "not EthAccount"), nil, err
		}

		return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgEthereumTx, "not EthAccount"), nil, nil
	}
}

func SimulateEthCreateContract(ak types.AccountKeeper, k *keeper.Keeper) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		return simtypes.OperationMsg{}, nil, nil
	}
}

func SimulateEthCallContract(ak types.AccountKeeper, k *keeper.Keeper) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		return simtypes.OperationMsg{}, nil, nil
	}
}
