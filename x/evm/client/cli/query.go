package cli

import (
	"math/big"
	"strings"

	rpctypes "github.com/cosmos/ethermint/ethereum/rpc/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"

	"github.com/InjectiveLabs/sdk-go/ethereum/rpc"
	"github.com/InjectiveLabs/sdk-go/wrappers"
	"github.com/cosmos/ethermint/x/evm/types"
)

// GetQueryCmd returns the parent command for all x/bank CLi query commands.
func GetQueryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Querying commands for the evm module",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		GetStorageCmd(),
		GetCodeCmd(),
		GetErc20Balance(),
		GetAccount(),
	)
	return cmd
}

// GetStorageCmd queries a key in an accounts storage
func GetStorageCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "storage [address] [key]",
		Short: "Gets storage for an account with a given key and height",
		Long:  "Gets storage for an account with a given key and height. If the height is not provided, it will use the latest height from context.",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			address, err := accountToHex(args[0])
			if err != nil {
				return err
			}

			key := formatKeyToHash(args[1])

			req := &types.QueryStorageRequest{
				Address: address,
				Key:     key,
			}

			res, err := queryClient.Storage(rpctypes.ContextWithHeight(clientCtx.Height), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// GetCodeCmd queries the code field of a given address
func GetCodeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "code [address]",
		Short: "Gets code from an account",
		Long:  "Gets code from an account. If the height is not provided, it will use the latest height from context.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			address, err := accountToHex(args[0])
			if err != nil {
				return err
			}

			req := &types.QueryCodeRequest{
				Address: address,
			}

			res, err := queryClient.Code(rpctypes.ContextWithHeight(clientCtx.Height), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// GetErc20Balance queries the erc20 balance of an address
func GetErc20Balance() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "erc20balance [contract] [address]",
		Short: "Gets erc20 balance of an account",
		Long:  "Gets erc20 balance of an account.",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			contract := args[0]
			address := args[1]

			erc20ABI, err := abi.JSON(strings.NewReader(wrappers.ERC20ABI))
			if err != nil {
				return err
			}
			input, err := erc20ABI.Pack("balanceOf", common.HexToAddress(address))
			if err != nil {
				return err
			}

			req := &types.QueryStaticCallRequest{
				Address: contract,
				Input:   input,
			}

			res, err := queryClient.StaticCall(rpc.ContextWithHeight(clientCtx.Height), req)
			if err != nil {
				return err
			}
			ret := big.NewInt(0)
			err = erc20ABI.UnpackIntoInterface(&ret, "balanceOf", res.Data)
			if err != nil {
				return err
			}

			return clientCtx.PrintString(ret.String())
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// GetAccount queries the account by address
func GetAccount() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "account [address]",
		Short: "Get an account by address",
		Long:  "Get an account by address",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)
			address := args[0]

			req := &types.QueryAccountRequest{
				Address: address,
			}

			res, err := queryClient.Account(rpc.ContextWithHeight(clientCtx.Height), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}
