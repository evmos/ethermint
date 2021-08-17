<!--
order: 2
-->

# Metamask

Connect your Metamask wallet with Ethermint. {synopsis}

## Adding a custom Network for Ethermint

Open the Metamask extension on your browser, you may have to log in to your Metamask account if you
are not already. Then click the top right circle and go to `Settings` > `Networks` > `Add
Network` and fill the form as shown below.

::: tip
You can also find the full `ChainID` form the `genesis.json` file. To get the
[EIP155](https://github.com/ethereum/EIPs/blob/master/EIPS/eip-155.md) chain ID from the Cosmos
chain ID, check the [Chain ID](./../../basics/chain_id) documentation page.
:::

![metamask networks settings](./../img/metamask_network_settings.png)

Here is the list of fields that you can use to paste on Metamask:

- **Network Name**: Ethermint
<!-- TODO: add RPC URL -->
- **New RPC URL**: http://localhost:8545
- **Chain ID**: 9000
- **Currency Symbol (optional)**: PHOTON
- **Block Explorer URL (optional)**:

## Import Account to Metamask

Then close the settings, and go to `My Accounts` (top right circle) and select `Import Account`. You should see and image like the following one:

![metamask import account page](./../img/metamask_import.png)

Now you can export your private key from the terminal using the following command. Again, make sure
to replace `mykey` with the name of the key that you want to export and use the correct `keyring-backend`:

```bash
ethermintd keys unsafe-export-eth-key mykey --keyring-backend test
```

Go back to the browser and select the `Private Key` option. Then paste the private key exported from
the `unsafe-export-eth-key` command.

Your account balance should show up as `1 APHOTON` and do transfers as usual.

::: tip
If it takes some time to load the balance of the account, change the network to `Main Ethereum
Network` (or any other than `Localhost 8545` or `Ethermint`) and then switch back to `Ethermint`.
:::

## Downloading State

To see your Metamask logs, click the top right circle and go to `Settings` > `Advanced` > `Download State Logs`.
If you search through the JSON file for the account address you'll find the transaction history.
