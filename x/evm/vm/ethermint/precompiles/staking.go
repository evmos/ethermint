package precompiles

import (
	"errors"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"

	"github.com/evmos/ethermint/x/evm/statedb"
	evm "github.com/evmos/ethermint/x/evm/vm"

	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

var (
	// DelegateMethod defines the ABI method signature for the staking Delegate
	// function.
	DelegateMethod abi.Method
	// UndelegateMethod defines the ABI method signature for the staking Undelegate
	// function.
	UndelegateMethod abi.Method

	_ evm.StatefulPrecompiledContract = (*StakingPrecompile)(nil)
)

func init() {
	addressType, _ := abi.NewType("address", "", nil)
	stringType, _ := abi.NewType("string", "", nil)
	uint256Type, _ := abi.NewType("uint256", "", nil)

	DelegateMethod = abi.NewMethod(
		"delegate", // name
		"delegate", // raw name
		abi.Function,
		"",
		false,
		false,
		abi.Arguments{
			{
				Name: "delegatorAddress",
				Type: addressType,
			},
			{
				Name: "validatorAddress",
				Type: stringType,
			},
			{
				Name: "amount",
				Type: uint256Type,
			},
		},
		abi.Arguments{},
	)

	UndelegateMethod = abi.NewMethod(
		"undelegate", // name
		"undelegate", // raw name
		abi.Function,
		"",
		false,
		false,
		abi.Arguments{
			{
				Name: "delegatorAddress",
				Type: addressType,
			},
			{
				Name: "validatorAddress",
				Type: stringType,
			},
			{
				Name: "amount",
				Type: uint256Type,
			},
		},
		abi.Arguments{},
	)
}

type StakingPrecompile struct {
	stakingKeeper stakingkeeper.Keeper
}

func NewStakingPrecompile(
	stakingKeeper stakingkeeper.Keeper,
) evm.StatefulPrecompiledContract {
	return &StakingPrecompile{
		stakingKeeper: stakingKeeper,
	}
}

// RequiredGas calculates the contract gas use
func (sp *StakingPrecompile) RequiredGas(input []byte) uint64 {
	// TODO estimate required gas
	return 0
}

func (sp *StakingPrecompile) Run(_ []byte) ([]byte, error) {
	return nil, errors.New("should run with RunStateful")
}

func (sp *StakingPrecompile) RunStateful(evm evm.EVM, caller common.Address, input []byte, value *big.Int) ([]byte, error) {
	stateDB, ok := evm.StateDB().(statedb.ExtStateDB)
	if !ok {
		return nil, errors.New("not run in ethermint")
	}

	ctx := stateDB.Context()

	methodID := string(input[:4])
	argsBz := input[4:]

	switch methodID {
	case string(DelegateMethod.ID):
		return sp.Delegate(ctx, argsBz, stateDB)
	case string(UndelegateMethod.ID):
		return sp.Undelegate(ctx, argsBz, stateDB)
	// TODO: redelegate
	// TODO: get delegation
	default:
		return nil, fmt.Errorf("unknown method '%s'", methodID)
	}
}

func (sp *StakingPrecompile) Delegate(ctx sdk.Context, argsBz []byte, stateDB statedb.ExtStateDB) ([]byte, error) {
	args, err := DelegateMethod.Inputs.Unpack(argsBz)
	if err != nil {
		return nil, errors.New("fail to unpack input arguments")
	}

	denom := sp.stakingKeeper.BondDenom(ctx)

	msg, err := sp.checkDelegateArgs(denom, args)
	if err != nil {
		return nil, err
	}

	msgSrv := stakingkeeper.NewMsgServerImpl(sp.stakingKeeper)

	cacheCtx, writeFn := ctx.CacheContext()

	if _, err := msgSrv.Delegate(sdk.WrapSDKContext(cacheCtx), msg); err != nil {
		return nil, err
	}

	writeFn()

	// FIXME: add entry to revert delegation behavior
	// stateDB.AppendJournalEntry(entry)

	return nil, nil
}

func (sp *StakingPrecompile) Undelegate(ctx sdk.Context, argsBz []byte, stateDB statedb.ExtStateDB) ([]byte, error) {
	args, err := UndelegateMethod.Inputs.Unpack(argsBz)
	if err != nil {
		return nil, errors.New("fail to unpack input arguments")
	}

	denom := sp.stakingKeeper.BondDenom(ctx)

	msg, err := sp.checkUndelegateArgs(denom, args)
	if err != nil {
		return nil, err
	}

	msgSrv := stakingkeeper.NewMsgServerImpl(sp.stakingKeeper)

	cacheCtx, writeFn := ctx.CacheContext()

	if _, err := msgSrv.Undelegate(sdk.WrapSDKContext(cacheCtx), msg); err != nil {
		return nil, err
	}

	writeFn()

	// FIXME: add entry to revert undelegation behavior
	// stateDB.AppendJournalEntry(entry)

	return nil, nil
}

func (sp *StakingPrecompile) checkDelegateArgs(denom string, args []interface{}) (*stakingtypes.MsgDelegate, error) {
	if len(args) != 3 {
		return nil, fmt.Errorf("invalid input arguments. Expected 3, got %d", len(args))
	}

	delegatorAddr, _ := args[0].(common.Address)
	validatorAddr, _ := args[1].(string)
	amount, ok := args[2].(*big.Int)
	if !ok || amount == nil {
		amount = big.NewInt(0)
	}

	coin := sdk.Coin{
		Denom:  denom,
		Amount: sdk.NewIntFromBigInt(amount),
	}

	delAddr := sdk.AccAddress(delegatorAddr.Bytes())

	msg := &stakingtypes.MsgDelegate{
		DelegatorAddress: delAddr.String(), // bech32 formatted
		ValidatorAddress: validatorAddr,
		Amount:           coin,
	}

	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	return msg, nil
}

func (sp *StakingPrecompile) checkUndelegateArgs(denom string, args []interface{}) (*stakingtypes.MsgUndelegate, error) {
	if len(args) != 3 {
		return nil, fmt.Errorf("invalid input arguments. Expected 3, got %d", len(args))
	}

	delegatorAddr, _ := args[0].(common.Address)
	validatorAddr, _ := args[1].(string)
	amount, ok := args[2].(*big.Int)
	if !ok || amount == nil {
		amount = big.NewInt(0)
	}

	coin := sdk.Coin{
		Denom:  denom,
		Amount: sdk.NewIntFromBigInt(amount),
	}

	delAddr := sdk.AccAddress(delegatorAddr.Bytes())

	msg := &stakingtypes.MsgUndelegate{
		DelegatorAddress: delAddr.String(), // bech32 formatted
		ValidatorAddress: validatorAddr,
		Amount:           coin,
	}

	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	return msg, nil
}
