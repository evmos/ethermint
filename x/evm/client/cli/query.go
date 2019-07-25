package cli

import (
	"fmt"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/ethermint/x/evm/types"
	"github.com/spf13/cobra"
)

func GetQueryCmd(moduleName string, cdc *codec.Codec) *cobra.Command {
	evmQueryCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Querying commands for the evm module",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}
	evmQueryCmd.AddCommand(client.GetCommands(
		GetCmdGetBlockNumber(moduleName, cdc),
		GetCmdGetStorageAt(moduleName, cdc),
		GetCmdGetCode(moduleName, cdc),
	)...)
	return evmQueryCmd
}

// GetCmdGetBlockNumber queries information about the current block number
func GetCmdGetBlockNumber(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "block-number",
		Short: "Gets block number (block height)",
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/blockNumber", queryRoute), nil)
			if err != nil {
				fmt.Printf("could not resolve: %s\n", err)
				return nil
			}

			var out types.QueryResBlockNumber
			cdc.MustUnmarshalJSON(res, &out)
			return cliCtx.PrintOutput(out)
		},
	}
}

// GetCmdGetStorageAt queries a key in an accounts storage
func GetCmdGetStorageAt(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "storage [account] [key]",
		Short: "Gets storage for an account at a given key",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			// TODO: Validate args
			account := args[0]
			key := args[1]

			res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/storage/%s/%s", queryRoute, account, key), nil)
			if err != nil {
				fmt.Printf("could not resolve: %s\n", err)
				return nil
			}
			var out types.QueryResStorage
			cdc.MustUnmarshalJSON(res, &out)
			return cliCtx.PrintOutput(out)
		},
	}
}

// GetCmdGetCode queries the code field of a given address
func GetCmdGetCode(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "code [account]",
		Short: "Gets code from an account",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			// TODO: Validate args
			account := args[0]

			res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/code/%s", queryRoute, account), nil)
			if err != nil {
				fmt.Printf("could not resolve: %s\n", err)
				return nil
			}
			var out types.QueryResCode
			cdc.MustUnmarshalJSON(res, &out)
			return cliCtx.PrintOutput(out)
		},
	}
}
