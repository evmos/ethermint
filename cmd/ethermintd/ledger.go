package main

import (
	"fmt"
	"math/big"
	"strconv"

	"github.com/ethereum/go-ethereum/accounts/usbwallet"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/tharsis/ethermint/ethereum/rpc/backend"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/spf13/cobra"
	"github.com/tendermint/tendermint/libs/log"
	rpctypes "github.com/tharsis/ethermint/ethereum/rpc/types"
)

const (
	flagIndex = "index"
)

func runAddCmd(cmd *cobra.Command, _ []string) error {
	bip44Index, _ := cmd.Flags().GetUint32(flagIndex)
	ledgerHub, detectLedgerErr := usbwallet.NewLedgerHub()
	if detectLedgerErr != nil {
		return fmt.Errorf("ledger detect error %v", detectLedgerErr)
	}
	allLedgerWallets := ledgerHub.Wallets()
	firstLedgerWallet := allLedgerWallets[0]
	openWalletErr := firstLedgerWallet.Open("")
	if openWalletErr != nil {
		return fmt.Errorf("ledger open error %vn", openWalletErr)

	}

	// bip44(44), coin type(60), account, change ,index
	hdPath := []uint32{0x80000000 + 44, 0x80000000 + 60, 0x80000000 + 0, 0, bip44Index}
	derivedAddress, deriveErr := firstLedgerWallet.Derive(hdPath, true)
	if deriveErr != nil {
		return fmt.Errorf("derive address error %v", deriveErr)
	}

	bech32Address, _ := sdk.Bech32ifyAddressBytes(sdk.GetConfig().GetBech32AccountAddrPrefix(), derivedAddress.Address[:])
	fmt.Printf("Ledger Address Index= %d\nAddress= %s  %s\n", bip44Index, derivedAddress.Address.Hex(), bech32Address)
	closeWalletErr := firstLedgerWallet.Close()
	if closeWalletErr != nil {
		return fmt.Errorf("ledger close error %v", closeWalletErr)
	}
	return nil
}

func ledgerCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        "ledger",
		Aliases:                    []string{"l"},
		Short:                      "Ledger subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	// add
	addCmd := &cobra.Command{
		Use:   "add",
		Short: "Add ledger address",
		RunE:  runAddCmd,
	}
	addCmd.Flags().Uint32(flagIndex, 0, "Address index number for HD derivation")

	// send
	sendCmd := &cobra.Command{
		Use: "send [from_key_or_address] [to_address] [amount]",
		Short: `Send funds from one account to another. Note, the'--from' flag is
ignored as it is implied from [from_key_or_address].`,
		Args: cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, sendTransactionErr := client.GetClientTxContext(cmd)
			if sendTransactionErr != nil {
				return sendTransactionErr
			}

			gasLimitString, _ := cmd.Flags().GetString(flags.FlagGas)
			gasPriceString, _ := cmd.Flags().GetString(flags.FlagGasPrices)

			fromAddr := args[0]
			toAddr := args[1]
			coins := args[2]

			evmBackend := backend.NewEVMBackend(log.NewNopLogger(), clientCtx)
			queryClient := rpctypes.NewQueryClient(clientCtx)

			fromAddressBytes, fromAddressErr := hexutil.Decode(fromAddr)
			if fromAddressErr != nil {
				return fmt.Errorf("from address error %v", fromAddressErr)
			}
			fromAddress := common.BytesToAddress(fromAddressBytes)
			bech32FromAddress, _ := sdk.Bech32ifyAddressBytes(sdk.GetConfig().GetBech32AccountAddrPrefix(), fromAddressBytes)

			toAddressBytes, toAddressErr := hexutil.Decode(toAddr)
			if toAddressErr != nil {
				return fmt.Errorf("to address error %v", toAddressErr)
			}
			toAddress := common.BytesToAddress(toAddressBytes)
			bechh32ToAddress, _ := sdk.Bech32ifyAddressBytes(sdk.GetConfig().GetBech32AccountAddrPrefix(), toAddressBytes)

			fmt.Printf("From= %s %s\n", fromAddr, bech32FromAddress)
			fmt.Printf("To=   %s %s\n", toAddr, bechh32ToAddress)
			fmt.Printf("Amount= %s\n", coins)
			fmt.Printf("Gas Limit= %s  Gas Price= %s\n", gasLimitString, gasPriceString)

			gasLimitBytes, gasLimitErr := strconv.ParseUint(gasLimitString, 10, 64) // decial 64bit
			if gasLimitErr != nil {
				return fmt.Errorf("gas limit error %v", gasLimitErr)
			}
			gasPriceValue := big.NewInt(0)
			gasPriceValue.SetString(gasLimitString, 10) // decimal

			amount := big.NewInt(0)
			amount.SetString(coins, 10) // decimal

			sendArgs := rpctypes.SendTxArgs{
				From:     fromAddress,
				To:       &toAddress,
				Gas:      (*hexutil.Uint64)(&gasLimitBytes),
				GasPrice: (*hexutil.Big)(gasPriceValue),
				Value:    (*hexutil.Big)(amount),
				Data:     nil,
				Input:    nil,
			}

			transactionHash, sendTransactionErr := SendTransactionEth(clientCtx, evmBackend, queryClient, sendArgs)
			fmt.Printf("Txhash= %v\n", transactionHash)
			return sendTransactionErr
		},
	}
	sendCmd.Flags().String(flags.FlagGasPrices, "20", "Gas prices to determine the transaction fee")
	sendCmd.Flags().String(flags.FlagGas, "10000000", "Gas limit to determine the transaction fee")

	cmd.AddCommand(addCmd, sendCmd)

	return cmd
}
