<!--
order: 1
-->

# Joining a Testnet

This document outlines the steps to join an existing testnet {synopsis}

## Install `ethermintd`

Follow the [installation](./../quickstart/installation) document to install the Ethermint binary `ethermintd`.

:::warning
Make sure you have the right version of `ethermintd` installed
:::

## Initialize Node

We need to initialize the node to create all the necessary validator and node configuration files:

```bash
ethermintd init <your_custom_moniker>
```

::: danger
Monikers can contain only ASCII characters. Using Unicode characters will render your node unreachable.
:::

By default, the `init` command creates your `~/.ethermintd` directory with subfolders `config/` and `data/`.
In the `config` directory, the most important files for configuration are `app.toml` and `config.toml`.

## Genesis & Seeds

### Copy the Genesis File

Check the genesis file from the [`testnets`](https://github.com/tharsis/testnets) repository and copy it over to the directory `~/.ethermintd/config/genesis.json`

### Add Seed Nodes

Your node needs to know how to find peers. You'll need to add healthy seed nodes to `$HOME/.ethermintd/config/config.toml`. The [`testnets`](https://github.com/tharsis/testnets) repo contains links to some seed nodes.

Edit the file located in `~/.ethermintd/config/config.toml` and the `seeds` to the following:

```toml
#######################################################
###           P2P Configuration Options             ###
#######################################################
[p2p]

# ...

# Comma separated list of seed nodes to connect to
seeds = ""
```

Validate genesis and start the Ethermint network

```bash
ethermintd validate-genesis
```

```bash
ethermintd start
```
