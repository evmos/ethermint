package bank

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/spf13/cobra"

	proto "github.com/gogo/protobuf/proto"

	"github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

func MsgUtilityCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "data <sender> <recipient> <coin>",
		Short: "Generate encode data for msg send",
		Long: `Generate encode data for msg send.
Example:
	ethermintd bank data D09F7C8C4529CB5D387AA17E33D707C529A6F694 03EB2CBAE6754C6E459F444783D1557DCA0F4E1A 10evm/0x5003c1fcc043D2d81fF970266bf3fa6e8C5a1F3A
	`,
		Args: cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			coins, err := sdk.ParseCoinsNormalized(args[2])
			if err != nil {
				return fmt.Errorf("failed to parse coins: %w", err)
			}
			msg := banktypes.NewMsgSend(
				sdk.AccAddress(common.FromHex(args[0])),
				sdk.AccAddress(common.FromHex(args[1])),
				coins,
			)
			bytes, err := proto.Marshal(msg)
			if err != nil {
				return err
			}

			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			return clientCtx.PrintString(string(bytes))
		},
	}

	return cmd
}
