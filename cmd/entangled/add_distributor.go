package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"

	"github.com/cosmos/cosmos-sdk/client/flags"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmos/cosmos-sdk/crypto/keyring"

	"github.com/Entangle-Protocol/entangle-blockchain/crypto/hd"
	distributorsauthtypes "github.com/Entangle-Protocol/entangle-blockchain/x/distributorsauth/types"
)

// AddDistributorCmd returns add-distributor cobra Command.
func AddDistributorCmd(defaultNodeHome string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add-distributor [address_or_key_name] [end_date]",
		Short: "Add a distributor account to genesis.json",
		Long: `Add a distrubutor account to genesis.json. 
		This account can send tx to evm module.
`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd).WithKeyringOptions(hd.EthSecp256k1Option())
			clientCtx, err := client.ReadPersistentCommandFlags(clientCtx, cmd.Flags())
			if err != nil {
				return err
			}

			serverCtx := server.GetServerContextFromCmd(cmd)
			config := serverCtx.Config

			config.SetRoot(clientCtx.HomeDir)

			var kr keyring.Keyring
			addr, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				inBuf := bufio.NewReader(cmd.InOrStdin())
				keyringBackend, _ := cmd.Flags().GetString(flags.FlagKeyringBackend)

				if keyringBackend != "" && clientCtx.Keyring == nil {
					var err error
					kr, err = keyring.New(sdk.KeyringServiceName(), keyringBackend, clientCtx.HomeDir, inBuf, clientCtx.Codec)
					if err != nil {
						return err
					}
				} else {
					kr = clientCtx.Keyring
				}

				k, err := kr.Key(args[0])
				if err != nil {
					return fmt.Errorf("failed to get address from Keyring: %w", err)
				}

				addr, err = k.GetAddress()
				if err != nil {
					return err
				}
			}

			end_date, err := strconv.ParseUint(args[1], 10, 64)
			if err != nil {
				return err
			}

			distributor := distributorsauthtypes.DistributorInfo{
				Address: addr.String(),
				EndDate: end_date,
			}

			// if err := genState.Validate(); err != nil {
			// 	return fmt.Errorf("failed to validate new genesis account: %w", err)
			// }

			genFile := config.GenesisFile()
			appState, genDoc, err := genutiltypes.GenesisStateFromGenFile(genFile)
			if err != nil {
				return fmt.Errorf("failed to unmarshal genesis state: %w", err)
			}

			distrGenState := distributorsauthtypes.GetGenesisStateFromAppState(clientCtx.Codec, appState)

			distributors := distrGenState.Distributors
			if distributorsauthtypes.ContainsGenesisDistributor(addr.String(), distributors) {
				return fmt.Errorf("cannot add account at existing address %s", addr)
			}

			distributors = append(distributors, distributor)
			distributors = distributorsauthtypes.SanitizeGenesisDistributor(distributors)

			distrGenState.Distributors = distributors

			distrGenStateBz, err := clientCtx.Codec.MarshalJSON(distrGenState)
			if err != nil {
				return fmt.Errorf("failed to marshal distributorsauth genesis state: %w", err)
			}

			appState[distributorsauthtypes.ModuleName] = distrGenStateBz

			appStateJSON, err := json.Marshal(appState)
			if err != nil {
				return fmt.Errorf("failed to marshal application genesis state: %w", err)
			}

			genDoc.AppState = appStateJSON
			return genutil.ExportGenesisFile(genDoc, genFile)
		},
	}

	return cmd
}
