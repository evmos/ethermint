<!--
order: 1
-->

# Single Node

## Pre-requisite Readings

- [Install Binary](./../../quickstart/installation)  {prereq}
- [Install Starport](https://docs.starport.network/intro/install.html)  {prereq}

## Automated Localnet (script)

```bash
init.sh
```

## Manual Localnet

This guide helps you create a single validator node that runs a network locally for testing and other development related uses.

### Initialize node

```bash
$MONIKER=testing
$KEY=mykey
$CHAINID="ethermint_9000-1"

ethermintd init $MONIKER --chain-id=$CHAINID
```

::: warning
Monikers can contain only ASCII characters. Using Unicode characters will render your node unreachable.
:::

You can edit this `moniker` later, in the `$(HOME)/.ethermintd/config/config.toml` file:

```toml
# A custom human readable name for this node
moniker = "<your_custom_moniker>"
```

You can edit the `$HOME/.ethermintd/config/app.toml` file in order to enable the anti spam mechanism and reject incoming transactions with less than the minimum gas prices:

```toml
# This is a TOML config file.
# For more information, see https://github.com/toml-lang/toml

##### main base config options #####

# The minimum gas prices a validator is willing to accept for processing a
# transaction. A transaction's fees must meet the minimum of any denomination
# specified in this config (e.g. 10aphoton).

minimum-gas-prices = ""
```

### Genesis Procedure

```bash
# Create a key to hold your account
ethermintd keys add $KEY

# Add that key into the genesis.app_state.accounts array in the genesis file
# NOTE: this command lets you set the number of coins. Make sure this account has some coins
# with the genesis.app_state.staking.params.bond_denom denom, the default is staking
ethermintd add-genesis-account $(ethermintd keys show validator -a) 1000000000stake,10000000000aphoton

# Generate the transaction that creates your validator
ethermintd gentx --name $KEY

# Add the generated bonding transaction to the genesis file
ethermintd collect-gentxs

# Finally, check the correctness of the genesis.json file
ethermintd validate-genesis
```

### Run Testnet

Now its safe to start the daemon:

```bash
ethermintd start
```

You can then stop the node using Ctrl+C.
