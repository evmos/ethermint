package cli

import (
	"fmt"
	"strings"

	// "strings"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"

	// "github.com/cosmos/cosmos-sdk/client/flags"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Entangle-Protocol/entangle-blockchain/utils/cli"
	"github.com/Entangle-Protocol/entangle-blockchain/x/distributorsauth/types"
)

// GetQueryCmd returns the cli query commands for this module
func GetQueryCmd() *cobra.Command {
	// Group distributorsauth queries under a subcommand
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("Querying commands for the %s module", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		//CmdQueryParams(),
		GetCmdQueryDistributors(),
		GetCmdQueryDistributor(),
		GetCmdQueryAdmins(),
		GetCmdQueryAdmin(),
	)
	// this line is used by starport scaffolding # 1

	return cmd
}

// GetCmdQueryDistributors implements the query distributors info
// distributorsauth command.
func GetCmdQueryDistributors() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list-distributors",
		Args:  cobra.MinimumNArgs(0),
		Short: "Query valid distributors with expire dates",
		Long: strings.TrimSpace(`
Query valid distributor with expire dates
`),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)
			query := types.QueryDistributorsRequest{}

			res, err := queryClient.Distributors(cmd.Context(), &query)
			return cli.PrintOrErr(res, err, clientCtx)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// GetCmdQueryDistributor implements the query distributor with address
// distributor command.
func GetCmdQueryDistributor() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "query-distributor [distributor]",
		Args:  cobra.RangeArgs(1, 1),
		Short: "Query valid distributor",
		Long: strings.TrimSpace(`
Query valid distributor.
`),
		RunE: func(cmd *cobra.Command, args []string) error {

			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)
			query := types.QueryDistributorRequest{}

			if len(args) > 0 {
				distributor, err := sdk.AccAddressFromBech32(args[0])
				if err != nil {
					return err
				}
				query.DistributorAddr = distributor.String()
			}

			res, err := queryClient.Distributor(cmd.Context(), &query)

			return cli.PrintOrErr(res, err, clientCtx)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// GetCmdQueryAdmins implements the query admins info
// admins command.
func GetCmdQueryAdmins() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list-admins",
		Args:  cobra.MinimumNArgs(0),
		Short: "Query valid admins with expire dates",
		Long: strings.TrimSpace(`
Query valid distributor with expire dates
`),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)
			query := types.QueryAdminsRequest{}

			res, err := queryClient.Admins(cmd.Context(), &query)
			return cli.PrintOrErr(res, err, clientCtx)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// GetCmdQueryAdmin implements the query admin with address
// admin command.
func GetCmdQueryAdmin() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "query-admin [admin_address]",
		Args:  cobra.RangeArgs(1, 1),
		Short: "Query valid admin",
		Long: strings.TrimSpace(`
Query valid admin.
`),
		RunE: func(cmd *cobra.Command, args []string) error {

			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)
			query := types.QueryAdminRequest{}

			if len(args) > 0 {
				admin, err := sdk.AccAddressFromBech32(args[0])
				if err != nil {
					return err
				}
				query.AdminAddr = admin.String()
			}

			res, err := queryClient.Admin(cmd.Context(), &query)

			return cli.PrintOrErr(res, err, clientCtx)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}
