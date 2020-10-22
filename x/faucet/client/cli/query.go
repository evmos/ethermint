package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"

	"github.com/cosmos/ethermint/x/faucet/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// GetQueryCmd defines evm module queries through the cli
func GetQueryCmd(cdc *codec.Codec) *cobra.Command {
	faucetQueryCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("Querying commands for the %s module", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}
	faucetQueryCmd.AddCommand(flags.GetCommands(
		GetCmdFunded(cdc),
	)...)
	return faucetQueryCmd
}

// GetCmdFunded queries the total amount funded by the faucet.
func GetCmdFunded(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "funded",
		Short: "Gets storage for an account at a given key",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := context.NewCLIContext().WithCodec(cdc)

			res, height, err := clientCtx.Query(fmt.Sprintf("custom/%s/%s", types.QuerierRoute, types.QueryFunded))
			if err != nil {
				return err
			}

			var out sdk.Coins
			cdc.MustUnmarshalJSON(res, &out)
			clientCtx = clientCtx.WithHeight(height)
			return clientCtx.PrintOutput(out)
		},
	}
}
