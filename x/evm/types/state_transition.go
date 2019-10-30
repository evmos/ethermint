package types

import (
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"

	sdk "github.com/cosmos/cosmos-sdk/types"
	emint "github.com/cosmos/ethermint/types"
)

// StateTransition defines data to transitionDB in evm
type StateTransition struct {
	Sender       common.Address
	AccountNonce uint64
	Price        *big.Int
	GasLimit     uint64
	Recipient    *common.Address
	Amount       *big.Int
	Payload      []byte
	Csdb         *CommitStateDB
	ChainID      *big.Int
	THash        *common.Hash
	Simulate     bool
}

// TransitionCSDB performs an evm state transition from a transaction
func (st StateTransition) TransitionCSDB(ctx sdk.Context) (*big.Int, sdk.Result) {
	contractCreation := st.Recipient == nil

	if res := st.checkNonce(); !res.IsOK() {
		return nil, res
	}

	cost, err := core.IntrinsicGas(st.Payload, contractCreation, true)
	if err != nil {
		return nil, sdk.ErrOutOfGas("invalid intrinsic gas for transaction").Result()
	}

	// This gas limit the the transaction gas limit with intrinsic gas subtracted
	gasLimit := st.GasLimit - ctx.GasMeter().GasConsumed()

	var snapshot int
	if st.Simulate {
		// gasLimit is set here because stdTxs incur gaskv charges in the ante handler, but for eth_call
		// the cost needs to be the same as an Ethereum transaction sent through the web3 API
		gasLimit = st.GasLimit - cost

		snapshot = st.Csdb.Snapshot()
	}

	// Create context for evm
	context := vm.Context{
		CanTransfer: core.CanTransfer,
		Transfer:    core.Transfer,
		Origin:      st.Sender,
		Coinbase:    common.Address{},
		BlockNumber: big.NewInt(ctx.BlockHeight()),
		Time:        big.NewInt(time.Now().Unix()),
		Difficulty:  big.NewInt(0x30000), // unused
		GasLimit:    gasLimit,
		GasPrice:    ctx.MinGasPrices().AmountOf(emint.DenomDefault).Int,
	}

	// This gas meter is set up to consume gas from gaskv during evm execution and be ignored
	evmGasMeter := sdk.NewInfiniteGasMeter()

	vmenv := vm.NewEVM(
		context, st.Csdb.WithContext(ctx.WithGasMeter(evmGasMeter)),
		GenerateChainConfig(st.ChainID), vm.Config{},
	)

	var (
		ret         []byte
		leftOverGas uint64
		addr        common.Address
		vmerr       error
		senderRef   = vm.AccountRef(st.Sender)
	)

	if contractCreation {
		ret, addr, leftOverGas, vmerr = vmenv.Create(senderRef, st.Payload, gasLimit, st.Amount)
	} else {
		// Increment the nonce for the next transaction
		st.Csdb.SetNonce(st.Sender, st.Csdb.GetNonce(st.Sender)+1)
		ret, leftOverGas, vmerr = vmenv.Call(senderRef, *st.Recipient, st.Payload, gasLimit, st.Amount)
	}

	// Generate bloom filter to be saved in tx receipt data
	bloomInt := big.NewInt(0)
	var bloomFilter ethtypes.Bloom
	if st.THash != nil {
		logs := st.Csdb.GetLogs(*st.THash)
		bloomInt = ethtypes.LogsBloom(logs)
		bloomFilter = ethtypes.BytesToBloom(bloomInt.Bytes())
	}

	// Encode all necessary data into slice of bytes to return in sdk result
	returnData := EncodeReturnData(addr, bloomFilter, ret)

	// handle errors
	if vmerr != nil {
		res := emint.ErrVMExecution(vmerr.Error()).Result()
		if vmerr == vm.ErrOutOfGas || vmerr == vm.ErrCodeStoreOutOfGas {
			res = sdk.ErrOutOfGas("EVM execution went out of gas").Result()
		}
		res.Data = returnData
		// Consume gas before returning
		ctx.GasMeter().ConsumeGas(gasLimit-leftOverGas, "EVM execution consumption")
		return nil, res
	}

	if st.Simulate {
		st.Csdb.RevertToSnapshot(snapshot)
	}

	// TODO: Refund unused gas here, if intended in future

	st.Csdb.Finalise(true) // Change to depend on config

	// Consume gas from evm execution
	// Out of gas check does not need to be done here since it is done within the EVM execution
	ctx.GasMeter().ConsumeGas(gasLimit-leftOverGas, "EVM execution consumption")

	return bloomInt, sdk.Result{Data: returnData, GasUsed: st.GasLimit - leftOverGas}
}

func (st *StateTransition) checkNonce() sdk.Result {
	// If simulated transaction, don't verify nonce
	if !st.Simulate {
		// Make sure this transaction's nonce is correct.
		nonce := st.Csdb.GetNonce(st.Sender)
		if nonce < st.AccountNonce {
			return emint.ErrInvalidNonce("nonce too high").Result()
		} else if nonce > st.AccountNonce {
			return emint.ErrInvalidNonce("nonce too low").Result()
		}
	}

	return sdk.Result{}
}
