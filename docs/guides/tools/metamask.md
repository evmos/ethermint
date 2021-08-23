<!--
order: 1
-->

# MetaMask: Connecting to Ethermint

Learn how to connect to the Ethermint testnet as well as a local node using MetaMask {synopsis}

## Pre-requisite Readings

- [Installation](./../../quickstart/installation.md) {prereq}
- [Run a node](./../../quickstart/run_node.md) {prereq}

The MetaMask browser extension is a wallet for accessing Ethereum-enabled applications and manage user identities. It can be used to connect to Ethermint through the official testnet or via a locally-running Ethermint node.

::: tip
If you haven’t already set up your own local node, refer to [the quickstart tutorial](../../quickstart/run_node/), or follow the instructions in the [GitHub repository](https://github.com/tharsis/ethermint/).
:::

## Set Up the MetaMask Extension

After downloading, installing, and initializing [MetaMask](https://metamask.io/), create a wallet, set a password, and store your secret backup phrase securely.

## Connect to Ethermint

After starting the Ethermint daemon, to connect to the local Ethermint node (it should be running) navigate to MetaMask's `Settings` in the top-right corner, followed by the `Networks` &rarr; `Add Network` option.

The network details are as follows:

:::: tabs
::: tab Local Node
* Network Name: `{{ $themeConfig.project.name }} Local`
* RPC URL: `{{ $themeConfig.project.rpc_url_local }}`
* ChainID: `n/a`
* Symbol (Optional): `{{ $themeConfig.project.ticker }}-LOCAL`
* Block Explorer (Optional): `n/a`
:::
::: tab Testnet
* Network Name: `{{ $themeConfig.project.name }}`
* RPC URL: `{{ $themeConfig.project.rpc_url }}`
* ChainID: `{{ $themeConfig.project.chain_id }}`
* Symbol (Optional): `{{ $themeConfig.project.ticker }}`
* Block Explorer (Optional): `{{ $themeConfig.project.block_explorer_url }}`
:::
::::

{IMAGE HERE}

MetaMask should now be connected to your local Ethermint node via its web3 RPC, and you should have a balance of 0 {{ $themeConfig.project.ticker }}.

{IMAGE HERE}

## Import Existing Account into MetaMask

If you would like to use an existing local account, you can export its private key using the following command:
```shell
$ ethermintd keys unsafe-export-eth-key mykey --keyring-backend test
```
You should be given the option to `Import Account` under MetaMask's `My Accounts` menu. Upon completing the import, you should see your account's balance.

{IMAGE HERE}
