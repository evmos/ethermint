package backend

import (
	"bytes"
	"errors"
	"math/big"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/tharsis/ethermint/ethereum/rpc/types"
	evmtypes "github.com/tharsis/ethermint/x/evm/types"
)

// setTxDefaults populates tx message with default values in case they are not
// provided on the args
func (e *EVMBackend) setTxDefaults(args types.SendTxArgs) (types.SendTxArgs, error) {
	if args.GasPrice == nil {
		// TODO: Suggest a gas price based on the previous included txs
		args.GasPrice = (*hexutil.Big)(new(big.Int).SetUint64(e.RPCGasCap()))
	}

	if args.Nonce == nil {
		// get the nonce from the account retriever
		// ignore error in case tge account doesn't exist yet
		nonce, _ := e.getAccountNonce(args.From, true, 0, e.logger)
		args.Nonce = (*hexutil.Uint64)(&nonce)
	}

	if args.Data != nil && args.Input != nil && !bytes.Equal(*args.Data, *args.Input) {
		return args, errors.New("both 'data' and 'input' are set and not equal. Please use 'input' to pass transaction call data")
	}

	if args.To == nil {
		// Contract creation
		var input []byte
		if args.Data != nil {
			input = *args.Data
		} else if args.Input != nil {
			input = *args.Input
		}

		if len(input) == 0 {
			return args, errors.New(`contract creation without any data provided`)
		}
	}

	if args.Gas == nil {
		// For backwards-compatibility reason, we try both input and data
		// but input is preferred.
		input := args.Input
		if input == nil {
			input = args.Data
		}

		callArgs := evmtypes.CallArgs{
			From:       &args.From, // From shouldn't be nil
			To:         args.To,
			Gas:        args.Gas,
			GasPrice:   args.GasPrice,
			Value:      args.Value,
			Data:       input,
			AccessList: args.AccessList,
		}
		blockNr := types.NewBlockNumber(big.NewInt(0))
		estimated, err := e.EstimateGas(callArgs, &blockNr)
		if err != nil {
			return args, err
		}
		args.Gas = &estimated
		e.logger.Debug("estimate gas usage automatically", "gas", args.Gas)
	}

	if args.ChainID == nil {
		args.ChainID = (*hexutil.Big)(e.chainID)
	}

	return args, nil
}

// getAccountNonce returns the account nonce for the given account address.
// If the pending value is true, it will iterate over the mempool (pending)
// txs in order to compute and return the pending tx sequence.
// Todo: include the ability to specify a blockNumber
func (e *EVMBackend) getAccountNonce(accAddr common.Address, pending bool, height int64, logger log.Logger) (uint64, error) {
	queryClient := authtypes.NewQueryClient(e.clientCtx)
	res, err := queryClient.Account(types.ContextWithHeight(height), &authtypes.QueryAccountRequest{Address: sdk.AccAddress(accAddr.Bytes()).String()})
	if err != nil {
		return 0, err
	}
	var acc authtypes.AccountI
	if err := e.clientCtx.InterfaceRegistry.UnpackAny(res.Account, &acc); err != nil {
		return 0, err
	}

	nonce := acc.GetSequence()

	if !pending {
		return nonce, nil
	}

	// the account retriever doesn't include the uncommitted transactions on the nonce so we need to
	// to manually add them.
	pendingTxs, err := e.PendingTransactions()
	if err != nil {
		logger.Error("failed to fetch pending transactions", "error", err.Error())
		return nonce, nil
	}

	// add the uncommitted txs to the nonce counter
	// only supports `MsgEthereumTx` style tx
	for _, tx := range pendingTxs {
		msg, err := evmtypes.UnwrapEthereumMsg(tx)
		if err != nil {
			// not ethereum tx
			continue
		}

		sender, err := msg.GetSender(e.chainID)
		if err != nil {
			continue
		}
		if sender == accAddr {
			nonce++
		}
	}

	return nonce, nil
}

// TxLogsFromEvents parses ethereum logs from cosmos events
func TxLogsFromEvents(codec codec.Codec, events []abci.Event) []*ethtypes.Log {
	logs := make([]*evmtypes.Log, 0)
	for _, event := range events {
		if event.Type != evmtypes.EventTypeTxLog {
			continue
		}
		for _, attr := range event.Attributes {
			if !bytes.Equal(attr.Key, []byte(evmtypes.AttributeKeyTxLog)) {
				continue
			}

			var log evmtypes.Log
			codec.MustUnmarshal(attr.Value, &log)
			logs = append(logs, &log)
		}
	}
	return evmtypes.LogsToEthereum(logs)
}
