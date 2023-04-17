package client

import (
	"github.com/evmos/ethermint/client/bank"
	"github.com/spf13/cobra"
)

func BankCommands() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "bank",
		Short: "Bank msg utility commands",
	}
	cmd.AddCommand(bank.MsgUtilityCommand())
	return cmd
}
