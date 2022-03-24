package simulation

import (
	"encoding/json"
	"fmt"
	"math/big"
	"math/rand"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	amino "github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/auth/tx"
	"github.com/cosmos/cosmos-sdk/x/simulation"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	ethtypes "github.com/ethereum/go-ethereum/core/types"

	"github.com/cosmos/cosmos-sdk/client"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/x/auth/signing"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/tharsis/ethermint/crypto/ethsecp256k1"
	"github.com/tharsis/ethermint/server/config"
	"github.com/tharsis/ethermint/tests"
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
		r *rand.Rand, bapp *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		simAccount, _ := simtypes.RandomAcc(r, accs)
		var receipient simtypes.Account
		if r.Intn(2) == 1 {
			receipient, _ = simtypes.RandomAcc(r, accs)
		} else {
			receipient = simtypes.RandomAccounts(r, 1)[0]
		}
		from := common.BytesToAddress(simAccount.Address)
		to := common.BytesToAddress(receipient.Address)

		estimateGas, err := EstimateGas(k, ctx, &from, &to, (*hexutil.Bytes)(&[]byte{}))
		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgEthereumTx, "can not estimate gas wanted"), nil, err
		}
		gasLimit := estimateGas
		ethChainID := k.ChainID()
		chainConfig := k.GetParams(ctx).ChainConfig.EthereumConfig(ethChainID)
		gasFeeCap := k.BaseFee(ctx, chainConfig)
		nonce := k.GetNonce(ctx, common.BytesToAddress(simAccount.Address))

		amount, err := RandomTransferableAmount(k, ctx, r, from, gasLimit, gasFeeCap)
		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgEthereumTx, "no enough transferable amount"), nil, nil
		}

		ethSimpleTransferTx := types.NewTx(ethChainID, nonce, &to, amount, gasLimit, gasFeeCap, gasFeeCap, nil, []byte{}, nil)

		ethAddress, err := GetEthAddress(simAccount)
		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgEthereumTx, "can not get ethaddress"), nil, err
		}
		ethSimpleTransferTx.From = ethAddress

		txConfig := NewTxConfig()
		txBuilder := txConfig.NewTxBuilder()
		signedTx, err := GetSignedTx(k, ctx, txBuilder, ethSimpleTransferTx, simAccount.PrivKey)
		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgEthereumTx, "can not sign ethereum tx"), nil, err
		}

		_, _, err = bapp.Deliver(txConfig.TxEncoder(), signedTx)
		// fmt.Printf("gas wanted: %v, used: %v\nlog: %v\n", gasInfo.GasWanted, gasInfo.GasUsed, result.Log)

		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgEthereumTx, "failed to deliver tx"), nil, err
		}

		return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgEthereumTx, "delivered"), nil, nil
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

func EstimateGas(k *keeper.Keeper, ctx sdk.Context, from, to *common.Address, data *hexutil.Bytes) (gas uint64, err error) {
	args, err := json.Marshal(&types.TransactionArgs{To: to, From: from, Data: data})
	if err != nil {
		return 0, err
	}

	res, err := k.EstimateGas(sdk.WrapSDKContext(ctx), &types.EthCallRequest{
		Args:   args,
		GasCap: config.DefaultGasCap,
	})
	if err != nil {
		return 0, err
	}
	return res.Gas, nil
}

func RandomTransferableAmount(k *keeper.Keeper, ctx sdk.Context, r *rand.Rand, address common.Address, gasLimit uint64, gasFeeCap *big.Int) (amount *big.Int, err error) {
	balance := k.GetBalance(ctx, address)
	feeLimit := new(big.Int).Mul(gasFeeCap, big.NewInt(int64(gasLimit)))
	if (feeLimit.Cmp(balance)) > 0 {
		return nil, fmt.Errorf("no enough balance")
	}
	spendable := new(big.Int).Sub(balance, feeLimit)
	if spendable.Cmp(big.NewInt(0)) == 0 {
		amount = new(big.Int).Set(spendable)
	} else {
		simAmount, err := simtypes.RandPositiveInt(r, sdk.NewIntFromBigInt(spendable))
		if err != nil {
			return nil, err
		}
		amount = simAmount.BigInt()
	}
	return amount, nil
}

func GetEthAddress(account simtypes.Account) (address string, err error) {
	prv, ok := account.PrivKey.(*ethsecp256k1.PrivKey)
	if !ok {
		return "", fmt.Errorf("require privkey type is ethsecp256k1.PrivKey")
	}
	key, err := prv.ToECDSA()
	if err != nil {
		return "", err
	}

	addr := crypto.PubkeyToAddress(key.PublicKey)
	return addr.String(), nil
}

func NewTxConfig() client.TxConfig {
	interfaceRegistry := codectypes.NewInterfaceRegistry()
	marshaler := amino.NewProtoCodec(interfaceRegistry)
	txConfig := tx.NewTxConfig(marshaler, tx.DefaultSignModes)
	return txConfig
}

func GetSignedTx(k *keeper.Keeper, ctx sdk.Context, txBuilder client.TxBuilder, msg *types.MsgEthereumTx, prv cryptotypes.PrivKey) (signedTx signing.Tx, err error) {
	builder, ok := txBuilder.(tx.ExtensionOptionsTxBuilder)
	if !ok {
		err = fmt.Errorf("can not initiate ExtensionOptionsTxBuilder")
		return nil, err
	}
	option, err := codectypes.NewAnyWithValue(&types.ExtensionOptionsEthereumTx{})
	if err != nil {
		return nil, err
	}
	builder.SetExtensionOptions(option)

	// err = ethSimpleTransferTx.Sign(ethtypes.LatestSignerForChainID(k.ChainID()), tests.NewSigner(prv))
	err = msg.Sign(ethtypes.LatestSignerForChainID(k.ChainID()), tests.NewSigner(prv))

	if err != nil {
		return nil, err
	}

	err = builder.SetMsgs(msg)
	if err != nil {
		return nil, err
	}

	txData, err := types.UnpackTxData(msg.Data)
	if err != nil {
		return nil, err
	}

	fees := sdk.NewCoins(sdk.NewCoin(k.GetParams(ctx).EvmDenom, sdk.NewIntFromBigInt(txData.Fee())))
	builder.SetFeeAmount(fees)
	builder.SetGasLimit(msg.GetGas())

	signedTx = builder.GetTx()
	return signedTx, nil
}
