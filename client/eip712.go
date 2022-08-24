package client

import (
	"github.com/evmos/ethermint/client/eip712"
	"github.com/spf13/cobra"
)

func EIP712Commands() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "eip712",
		Short: "EIP712 related utility commands",
	}
	cmd.AddCommand(eip712.DataTypeCommand())
	return cmd
}
