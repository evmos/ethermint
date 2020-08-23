package cli

import (
	"bufio"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authclient "github.com/cosmos/cosmos-sdk/x/auth/client/utils"

	"github.com/cosmos/ethermint/x/faucet/types"
)

// GetTxCmd return faucet sub-command for tx
func GetTxCmd(cdc *codec.Codec) *cobra.Command {
	faucetTxCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "faucet transaction subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	faucetTxCmd.AddCommand(flags.PostCommands(
		GetCmdRequest(cdc),
	)...)

	return faucetTxCmd
}

// GetCmdRequest is the CLI command to fund an address with the requested coins
func GetCmdRequest(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "request [amount] [other-recipient (optional)]",
		Short: "request an address with the requested coins",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			inBuf := bufio.NewReader(cmd.InOrStdin())
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			txBldr := auth.NewTxBuilderFromCLI(inBuf).WithTxEncoder(authclient.GetTxEncoder(cdc))

			amount, err := sdk.ParseCoins(args[0])
			if err != nil {
				return err
			}

			var recipient sdk.AccAddress
			if len(args) == 1 {
				recipient = cliCtx.GetFromAddress()
			} else {
				recipient, err = sdk.AccAddressFromBech32(args[1])
			}

			if err != nil {
				return err
			}

			msg := types.NewMsgFund(amount, cliCtx.GetFromAddress(), recipient)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return authclient.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
}
