<!--
order: 2
-->

# Run a Node

Run a local node and start the REST and JSON-RPC clients {synopsis}

## Pre-requisite Readings

- [Installation](./installation.md) {prereq}

## Automated deployment

Run the local node

::: warning
The script below will remove any pre-existing binaries installed. Use the manual deploy if you want
to keep your binaries and configuration files.
:::

```bash
./init.sh
```

In another terminal window or tab, run the Ethereum JSON-RPC server as well as the SDK REST server:

```bash
ethermintd rest-server --laddr "tcp://localhost:8545" --unlock-key mykey --chain-id 8
```

## Manual deployment

The instructions for setting up a brand new full node from scratch are the the same as running a
[single node local testnet](./testnet.md#single-node-local-manual-testnet).

## Start node

To start your node, just type:

```bash
ethermintd start
```

## Key Management

To run a node with the same key every time: replace `ethermintd keys add $KEY` in `./init.sh` with:

```bash
echo "your mnemonic here" | ethermintd keys add $KEY --recover
```

::: tip
Ethermint currently only supports 24 word mnemonics.
:::

You can generate a new key/mnemonic with:

```bash
ethermintd keys add $KEY
```

To export your ethermint key as an ethereum private key (for use with Metamask for example):

```bash
ethermintd keys unsafe-export-eth-key $KEY
```

For more about the available key commands, use the `--help` flag

```bash
ethermintd keys -h
```

### Keyring backend options

The instructions above include commands to use `test` as the `keyring-backend`. This is an unsecured
keyring that doesn't require entering a password and should not be used in production. Otherwise,
Ethermint supports using a file or OS keyring backend for key storage. To create and use a file
stored key instead of defaulting to the OS keyring, add the flag `--keyring-backend file` to any
relevant command and the password prompt will occur through the command line. This can also be saved
as a CLI config option with:

```bash
ethermintd config keyring-backend file
```

## Clearing data from chain

### Reset Data

Alternatively, you can **reset** the blockchain database, remove the node's address book files, and reset the `priv_validator.json` to the genesis state.

::: danger
If you are running a **validator node**, always be careful when doing `ethermintd unsafe-reset-all`. You should never use this command if you are not switching `chain-id`.
:::

::: danger
**IMPORTANT**: Make sure that every node has a unique `priv_validator.json`. **Do not** copy the `priv_validator.json` from an old node to multiple new nodes. Running two nodes with the same `priv_validator.json` will cause you to double sign!
:::

First, remove the outdated files and reset the data.

```bash
rm $HOME/.ethermintd/config/addrbook.json $HOME/.ethermintd/config/genesis.json
ethermintd unsafe-reset-all
```

Your node is now in a pristine state while keeping the original `priv_validator.json` and `config.toml`. If you had any sentry nodes or full nodes setup before, your node will still try to connect to them, but may fail if they haven't also been upgraded.

### Delete Data

Data for the Daemon and CLI binaries should be stored at `~/.ethermintd` and `~/.ethermintd`, respectively by default. To **delete** the existing binaries and configuration, run:

```bash
rm -rf ~/.emint*
```

To clear all data except key storage (if keyring backend chosen) and then you can rerun the full node installation commands from above to start the node again.

## Next {hide}

Learn about running a Ethermint [testnet](./testnet.md) {hide}
