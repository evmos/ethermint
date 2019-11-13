package cli

import (
	"math/big"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/keys"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/client/utils"
	emintTypes "github.com/cosmos/ethermint/types"
	"github.com/cosmos/ethermint/x/evm/types"

	ethcmn "github.com/ethereum/go-ethereum/common"
)

// GetTxCmd defines the CLI commands regarding evm module transactions
func GetTxCmd(storeKey string, cdc *codec.Codec) *cobra.Command {
	evmTxCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "EVM transaction subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	evmTxCmd.AddCommand(client.PostCommands(
	// TODO: Add back generating cosmos tx for Ethereum tx message
	// GetCmdGenTx(cdc),
	)...)

	return evmTxCmd
}

// GetCmdGenTx generates an ethereum transaction wrapped in a Cosmos standard transaction
func GetCmdGenTx(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "generate-tx [ethaddress] [amount] [gaslimit] [gasprice] [payload]",
		Short: "generate eth tx wrapped in a Cosmos Standard tx",
		Args:  cobra.ExactArgs(5),
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: remove inputs and infer based on StdTx
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			txBldr := auth.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))

			kb, err := keys.NewKeyBaseFromHomeFlag()
			if err != nil {
				panic(err)
			}

			coins, err := sdk.ParseCoins(args[1])
			if err != nil {
				return err
			}

			gasLimit, err := strconv.ParseUint(args[2], 0, 64)
			if err != nil {
				return err
			}

			gasPrice, err := strconv.ParseUint(args[3], 0, 64)
			if err != nil {
				return err
			}

			payload := args[4]
			addr := ethcmn.HexToAddress(args[0])
			// TODO: Remove explicit photon check and check variables
			msg := types.NewEthereumTxMsg(0, &addr, big.NewInt(coins.AmountOf(emintTypes.DenomDefault).Int64()), gasLimit, new(big.Int).SetUint64(gasPrice), []byte(payload))
			err = msg.ValidateBasic()
			if err != nil {
				return err
			}

			// TODO: possibly overwrite gas values in txBldr
			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr.WithKeybase(kb), []sdk.Msg{msg})
		},
	}
}
