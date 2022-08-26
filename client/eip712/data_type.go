package eip712

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/evmos/ethermint/client/flags"
	"github.com/evmos/ethermint/ethereum/eip712"
)

func DataTypeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "data-type <msg-url>",
		Short: "Generate EIP-712 data types for specified msg url",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)

			feeDelegation, err := cmd.Flags().GetBool(flags.FlagFeeDelegation)
			if err != nil {
				return err
			}

			typeURL := args[0]
			protoMsg, err := clientCtx.InterfaceRegistry.Resolve(typeURL)
			if err != nil {
				return err
			}
			msg, ok := protoMsg.(sdk.Msg)
			if !ok {
				return fmt.Errorf("the type is not a msg %s", typeURL)
			}

			typeData, err := eip712.ExtractMsgTypes(clientCtx.Codec, "MsgValue", msg, feeDelegation)
			if err != nil {
				return err
			}

			bz, err := json.Marshal(typeData)
			if err != nil {
				return err
			}

			fmt.Println(string(bz))
			return nil
		},
	}
	cmd.Flags().Bool(flags.FlagFeeDelegation, false, "enable fee delegation")
	return cmd
}
