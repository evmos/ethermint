package cli

import (
	"bufio"
	"fmt"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authclient "github.com/cosmos/cosmos-sdk/x/auth/client/utils"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	emint "github.com/cosmos/ethermint/types"
	"github.com/cosmos/ethermint/x/evm/types"
)

// GetTxCmd defines the CLI commands regarding evm module transactions
func GetTxCmd(cdc *codec.Codec) *cobra.Command {
	evmTxCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "EVM transaction subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	evmTxCmd.AddCommand(flags.PostCommands(
		GetCmdSendTx(cdc),
		GetCmdGenCreateTx(cdc),
	)...)

	return evmTxCmd
}

// GetCmdSendTx generates an Ethermint transaction (excludes create operations)
func GetCmdSendTx(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "send [to_address] [amount (in aphotons)] [<data>]",
		Short: "send transaction to address (call operations included)",
		Args:  cobra.RangeArgs(2, 3),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			inBuf := bufio.NewReader(cmd.InOrStdin())

			txBldr := auth.NewTxBuilderFromCLI(inBuf).WithTxEncoder(authclient.GetTxEncoder(cdc))

			toAddr, err := cosmosAddressFromArg(args[0])
			if err != nil {
				return errors.Wrap(err, "must provide a valid Bech32 address for to_address")
			}

			// Ambiguously decode amount from any base
			amount, err := strconv.ParseInt(args[1], 0, 64)
			if err != nil {
				return err
			}

			var data []byte
			if len(args) > 2 {
				payload := args[2]
				if !strings.HasPrefix(payload, "0x") {
					payload = "0x" + payload
				}

				data, err = hexutil.Decode(payload)
				if err != nil {
					return err
				}
			}

			from := cliCtx.GetFromAddress()

			_, seq, err := authtypes.NewAccountRetriever(cliCtx).GetAccountNumberSequence(from)
			if err != nil {
				return errors.Wrap(err, "Could not retrieve account sequence")
			}

			// TODO: Potentially allow overriding of gas price and gas limit
			msg := types.NewMsgEthermint(seq, &toAddr, sdk.NewInt(amount), txBldr.Gas(),
				sdk.NewInt(emint.DefaultGasPrice), data, from)

			err = msg.ValidateBasic()
			if err != nil {
				return err
			}

			return authclient.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
}

// GetCmdGenCreateTx generates an Ethermint transaction (excludes create operations)
func GetCmdGenCreateTx(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "create [contract bytecode] [<amount (in aphotons)>]",
		Short: "create contract through the evm using compiled bytecode",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			inBuf := bufio.NewReader(cmd.InOrStdin())

			txBldr := auth.NewTxBuilderFromCLI(inBuf).WithTxEncoder(authclient.GetTxEncoder(cdc))

			payload := args[0]
			if !strings.HasPrefix(payload, "0x") {
				payload = "0x" + payload
			}

			data, err := hexutil.Decode(payload)
			if err != nil {
				return err
			}

			var amount int64
			if len(args) > 1 {
				// Ambiguously decode amount from any base
				amount, err = strconv.ParseInt(args[1], 0, 64)
				if err != nil {
					return errors.Wrap(err, "invalid amount")
				}
			}

			from := cliCtx.GetFromAddress()

			_, seq, err := authtypes.NewAccountRetriever(cliCtx).GetAccountNumberSequence(from)
			if err != nil {
				return errors.Wrap(err, "Could not retrieve account sequence")
			}

			// TODO: Potentially allow overriding of gas price and gas limit
			msg := types.NewMsgEthermint(seq, nil, sdk.NewInt(amount), txBldr.Gas(),
				sdk.NewInt(emint.DefaultGasPrice), data, from)

			err = msg.ValidateBasic()
			if err != nil {
				return err
			}

			if err = authclient.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg}); err != nil {
				return err
			}

			contractAddr := ethcrypto.CreateAddress(common.BytesToAddress(from.Bytes()), seq)
			fmt.Printf(
				"Contract will be deployed to: \nHex: %s\nCosmos Address: %s\n",
				contractAddr.Hex(),
				sdk.AccAddress(contractAddr.Bytes()),
			)
			return nil
		},
	}
}
