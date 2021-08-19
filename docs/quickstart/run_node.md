<!--
order: 3
-->

# Run a Node

Configure and run an Ethermint node {synopsis}

## Pre-requisite Readings

- [Installation](./installation.md) {prereq}
- [`ethermintd`](./binary.md) {prereq}

## Automated deployment

Run the local node by running the `init.sh` script in the base directory of the repository.

::: warning
The script below will remove any pre-existing binaries installed. Use the manual deploy if you want
to keep your binaries and configuration files.
:::

```bash
./init.sh
```

## Manual deployment

The instructions for setting up a brand new full node from scratch are the the same as running a
[single node local testnet](./../guides/localnet/single_node.md#manual-localnet).

## Start node

To start your node, just type:

```bash
ethermintd start --json-rpc.enable=true --json-rpc.api="eth,web3,net"
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

To export your ethermint key as an Ethereum private key (for use with [Metamask](./../guides/keys-wallets/metamask) for example):

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

:::tip
For more information about the Keyring and its backend options, click [here](./../guides/keys-wallets/keyring).
:::

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

Data for the Daemon and CLI binaries should be stored at `~/.ethermintd`, respectively by default. To **delete** the existing binaries and configuration, run:

```bash
rm -rf ~/.ethermintd
```

To clear all data except key storage (if keyring backend chosen) and then you can rerun the full node installation commands from above to start the node again.

## Next {hide}

Learn about running a Ethermint [testnet](./testnet.md) {hide}
