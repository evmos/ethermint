[![CircleCI](https://circleci.com/gh/cosmos/ethermint.svg?style=svg)](https://circleci.com/gh/cosmos/ethermint)
[![](https://godoc.org/github.com/cosmos/ethermint?status.svg)](http://godoc.org/github.com/cosmos/ethermint) [![Go Report Card](https://goreportcard.com/badge/github.com/cosmos/ethermint)](https://goreportcard.com/report/github.com/cosmos/ethermint)

# Ethermint

__**WARNING:**__ Ethermint is under VERY ACTIVE DEVELOPMENT and should be treated as pre-alpha software. This means it is not meant to be run in production, its APIs are subject to change without warning and should not be relied upon, and it should not be used to hold any value. We will remove this warning when we have a release that is stable, secure, and properly tested.

## What is it?

`ethermint` will be an implementation of the EVM that runs on top of [`tendermint`](https://github.com/tendermint/tendermint) consensus, a Proof of Stake system. This project has as its primary goals:

- [Hard Spoon](https://blog.cosmos.network/introducing-the-hard-spoon-4a9288d3f0df) enablement: This is the ability to take a token from the Ethereum mainnet and "spoon" (shift) the balances over to another network. This feature is intended to make it easy for applications that require more transactions than the Ethereum main chain can provide to move their code over to a compatible chain with much more capacity.
- Web3 Compatibility: In order enable applications to be moved over to an ethermint chain existing tooling (i.e. web3 compatible clients) need to be able to interact with `ethermint`.

### Implementation

#### Completed

- Have a working implementation that can parse and validate the existing ETH Chain and persist it in a Tendermint store
- Implement Ethereum transactions in the CosmosSDK
- Implement web3 compatible API layer
- Implement the EVM as a CosmosSDK module
- Allow the Ethermint EVM to interact with other Cosmos SDK modules

#### Current Work

- Ethermint is a functioning Cosmos SDK application and can be deployed as its own zone
- Full web3 compatibility to enable existing Ethereum applications to use Ethermint

#### Next Steps

- Hard spoon enablement: The ability to export state from `geth` and import token balances into Ethermint

### Building Ethermint

To build, execute the following commands:

```bash
# To build the project and install it in $GOBIN
$ make install

# To build the binary and put the resulting binary in ./build
$ make build
```

### Starting a Ethermint daemon (node)

The following config steps can be performed all at once by executing the `init.sh` file located in the root directory like this:
```bash
./init.sh
```
> This bash file removes previous blockchain data from `~/.emintd` and `~/.emintcli`. It uses the `keyring-backend` called `test` that should prevent you from needing to enter a passkey. The `keyring-backend` `test` is unsecured and should not be used in production.

To initalize your chain manually, first create a key to use in signing the genesis transaction:

```bash
emintcli keys add mykey --keyring-backend test
```
> replace mykey with whatever you want to name the key

Then, run these commands to start up a node
```bash
# Set moniker and chain-id for Ethermint (Moniker can be anything, chain-id must be an integer)
emintd init mymoniker --chain-id 8

# Set up config for CLI
emintcli config keyring-backend test
emintcli config chain-id 8
emintcli config output json
emintcli config indent true
emintcli config trust-node true

# Allocate genesis accounts (cosmos formatted addresses)
emintd add-genesis-account $(emintcli keys show mykey -a) 1000000000000000000photon,1000000000000000000stake

# Sign genesis transaction
emintd gentx --name mykey --keyring-backend test

# Collect genesis tx
emintd collect-gentxs

# Run this to ensure everything worked and that the genesis file is setup correctly
emintd validate-genesis

# Start the node (remove the --pruning=nothing flag if historical queries are not needed)
emintd start --pruning=nothing
```
> Note: If you used `make build` instead of make install, and replace all `emintcli` and `emintd` references to `./build/emintcli` and `./build/emintd` respectively

### Starting Ethermint Web3 RPC API

After the daemon is started, run (in another process):

```bash
emintcli rest-server --laddr "tcp://localhost:8545" --unlock-key mykey
```

and to make sure the server has started correctly, try querying the current block number:

```
curl -X POST --data '{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}' -H "Content-Type: application/json" http://localhost:8545
```

or point any dev tooling at `http://localhost:8545` or whatever port is chosen just as you would with an Ethereum node

#### Clearing data from chain

Data for the CLI and Daemon should be stored at `~/.emintd` and `~/.emintcli` by default, to start the node with a fresh state, run:

```bash
rm -rf ~/.emint*
```

To clear all data except key storage (if keyring backend chosen) and then you can rerun the commands to start the node again.

#### Keyring backend options

The instructions above include commands to use `test` as the `keyring-backend`. This is an unsecured keyring that doesn't require entering a password and should not be used in production. Otherwise, Ethermint supports using a file or OS keyring backend for key storage. To create and use a file stored key instead of defaulting to the OS keyring, add the flag `--keyring-backend file` to any relevant command and the password prompt will occur through the command line. This can also be saved as a CLI config option with:

```bash
emintcli config keyring-backend file
```

### Exporting Ethereum private key from Ethermint

To export the private key from Ethermint to something like Metamask, run:

```bash
emintcli keys export-eth-key mykey
```

Import account through private key, and to verify that the Ethereum address is correct with:

```bash
emintcli keys parse $(emintcli keys show mykey -a)
```

### Tests

Integration tests are invoked via:

```bash
$ make test
```

To run CLI tests, execute:

```bash
$ make test-cli
```

#### Ethereum Mainnet Import

There is an included Ethereum mainnet exported blockchain file in `importer/blockchain`
that includes blocks up to height `97638`. To execute and test a full import of
these blocks using the EVM module, execute:

```bash
$ make test-import
```

You may also provide a custom blockchain export file to test importing more blocks
via the `--blockchain` flag. See `TestImportBlocks` for further documentation.

### Community

The following chat channels and forums are a great spot to ask questions about Ethermint:

- [Cosmos Discord](https://discord.gg/W8trcGV)
- Cosmos Forum [![Discourse status](https://img.shields.io/discourse/https/forum.cosmos.network/status.svg)](https://forum.cosmos.network)
