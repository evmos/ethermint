package backend

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/tharsis/ethermint/rpc/ethereum/types"
	evmtypes "github.com/tharsis/ethermint/x/evm/types"
)

// SetTxDefaults populates tx message with default values in case they are not
// provided on the args
func (e *EVMBackend) SetTxDefaults(args evmtypes.TransactionArgs) (evmtypes.TransactionArgs, error) {
	if args.GasPrice != nil && (args.MaxFeePerGas != nil || args.MaxPriorityFeePerGas != nil) {
		return args, errors.New("both gasPrice and (maxFeePerGas or maxPriorityFeePerGas) specified")
	}

	head := e.CurrentHeader()
	if head == nil {
		return args, errors.New("latest header is nil")
	}

	// If user specifies both maxPriorityfee and maxFee, then we do not
	// need to consult the chain for defaults. It's definitely a London tx.
	if args.MaxPriorityFeePerGas == nil || args.MaxFeePerGas == nil {
		// In this clause, user left some fields unspecified.
		if head.BaseFee != nil && args.GasPrice == nil {
			if args.MaxPriorityFeePerGas == nil {
				tip, err := e.SuggestGasTipCap(head.BaseFee)
				if err != nil {
					return args, err
				}
				args.MaxPriorityFeePerGas = (*hexutil.Big)(tip)
			}

			if args.MaxFeePerGas == nil {
				gasFeeCap := new(big.Int).Add(
					(*big.Int)(args.MaxPriorityFeePerGas),
					new(big.Int).Mul(head.BaseFee, big.NewInt(2)),
				)
				args.MaxFeePerGas = (*hexutil.Big)(gasFeeCap)
			}

			if args.MaxFeePerGas.ToInt().Cmp(args.MaxPriorityFeePerGas.ToInt()) < 0 {
				return args, fmt.Errorf("maxFeePerGas (%v) < maxPriorityFeePerGas (%v)", args.MaxFeePerGas, args.MaxPriorityFeePerGas)
			}

		} else {
			if args.MaxFeePerGas != nil || args.MaxPriorityFeePerGas != nil {
				return args, errors.New("maxFeePerGas or maxPriorityFeePerGas specified but london is not active yet")
			}

			if args.GasPrice == nil {
				price, err := e.SuggestGasTipCap(head.BaseFee)
				if err != nil {
					return args, err
				}
				if head.BaseFee != nil {
					// The legacy tx gas price suggestion should not add 2x base fee
					// because all fees are consumed, so it would result in a spiral
					// upwards.
					price.Add(price, head.BaseFee)
				}
				args.GasPrice = (*hexutil.Big)(price)
			}
		}
	} else {
		// Both maxPriorityfee and maxFee set by caller. Sanity-check their internal relation
		if args.MaxFeePerGas.ToInt().Cmp(args.MaxPriorityFeePerGas.ToInt()) < 0 {
			return args, fmt.Errorf("maxFeePerGas (%v) < maxPriorityFeePerGas (%v)", args.MaxFeePerGas, args.MaxPriorityFeePerGas)
		}
	}

	if args.Value == nil {
		args.Value = new(hexutil.Big)
	}
	if args.Nonce == nil {
		// get the nonce from the account retriever
		// ignore error in case tge account doesn't exist yet
		nonce, _ := e.getAccountNonce(*args.From, true, 0, e.logger)
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
			return args, errors.New("contract creation without any data provided")
		}
	}

	if args.Gas == nil {
		// For backwards-compatibility reason, we try both input and data
		// but input is preferred.
		input := args.Input
		if input == nil {
			input = args.Data
		}

		callArgs := evmtypes.TransactionArgs{
			From:                 args.From,
			To:                   args.To,
			Gas:                  args.Gas,
			GasPrice:             args.GasPrice,
			MaxFeePerGas:         args.MaxFeePerGas,
			MaxPriorityFeePerGas: args.MaxPriorityFeePerGas,
			Value:                args.Value,
			Data:                 input,
			AccessList:           args.AccessList,
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
		for _, msg := range (*tx).GetMsgs() {
			ethMsg, ok := msg.(*evmtypes.MsgEthereumTx)
			if !ok {
				// not ethereum tx
				break
			}

			sender, err := ethMsg.GetSender(e.chainID)
			if err != nil {
				continue
			}
			if sender == accAddr {
				nonce++
			}
		}
	}

	return nonce, nil
}

// AllTxLogsFromEvents parses all ethereum logs from cosmos events
func AllTxLogsFromEvents(events []abci.Event) ([][]*ethtypes.Log, error) {
	allLogs := make([][]*ethtypes.Log, 0, 4)
	for _, event := range events {
		if event.Type != evmtypes.EventTypeTxLog {
			continue
		}

		logs, err := ParseTxLogsFromEvent(event)
		if err != nil {
			return nil, err
		}

		allLogs = append(allLogs, logs)
	}
	return allLogs, nil
}

// TxLogsFromEvents parses ethereum logs from cosmos events for specific msg index
func TxLogsFromEvents(events []abci.Event, msgIndex int) ([]*ethtypes.Log, error) {
	for _, event := range events {
		if event.Type != evmtypes.EventTypeTxLog {
			continue
		}

		if msgIndex > 0 {
			// not the eth tx we want
			msgIndex--
			continue
		}

		return ParseTxLogsFromEvent(event)
	}
	return nil, fmt.Errorf("eth tx logs not found for message index %d", msgIndex)
}

// ParseTxLogsFromEvent parse tx logs from one event
func ParseTxLogsFromEvent(event abci.Event) ([]*ethtypes.Log, error) {
	logs := make([]*evmtypes.Log, 0, len(event.Attributes))
	for _, attr := range event.Attributes {
		if !bytes.Equal(attr.Key, []byte(evmtypes.AttributeKeyTxLog)) {
			continue
		}

		var log evmtypes.Log
		if err := json.Unmarshal(attr.Value, &log); err != nil {
			return nil, err
		}

		logs = append(logs, &log)
	}
	return evmtypes.LogsToEthereum(logs), nil
}
