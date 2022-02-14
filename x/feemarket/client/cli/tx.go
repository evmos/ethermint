package cli

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/tharsis/ethermint/x/feemarket/types"

	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/version"
	govcli "github.com/cosmos/cosmos-sdk/x/gov/client/cli"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	// "github.com/cosmos/cosmos-sdk/client/flags"
)

// NewSubmitBaseChangeProposalTxCmd returns a CLI command handler for creating
// a base fee change proposal governance transaction.
func NewSubmitBaseChangeProposalTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "base-fee-change [basefee]",
		Args:  cobra.ExactArgs(1),
		Short: "Submit a base fee change proposal",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Submit a base fee change proposal.

Example:
$ %s tx gov submit-proposal base-fee-change 500000 --from=<key_or_address>
`,
				version.AppName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			title, err := cmd.Flags().GetString(govcli.FlagTitle)
			if err != nil {
				return err
			}

			description, err := cmd.Flags().GetString(govcli.FlagDescription)
			if err != nil {
				return err
			}

			baseFeeValue, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return err
			}

			content := types.NewBaseFeeChangeProposal(
				title, description, baseFeeValue,
			)

			from := clientCtx.GetFromAddress()

			strDeposit, err := cmd.Flags().GetString(govcli.FlagDeposit)
			if err != nil {
				return err
			}

			deposit, err := sdk.ParseCoinsNormalized(strDeposit)
			if err != nil {
				return err
			}

			msg, err := govtypes.NewMsgSubmitProposal(content, deposit, from)
			if err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
	cmd.Flags().String(govcli.FlagTitle, "", "The proposal title")
	cmd.Flags().String(govcli.FlagDescription, "", "The proposal description")
	cmd.Flags().String(govcli.FlagDeposit, "", "The proposal deposit")

	return cmd
}
