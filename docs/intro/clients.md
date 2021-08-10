<!--
order: 3
-->

# Clients

Learn how to connect a client to a running node. {synopsis}

## Pre-requisite Readings

- [Run a Node](./run_node.md) {prereq}
- [Interacting with the Node](https://docs.cosmos.network/v0.43/run-node/interact-node.html) {prereq}

### Client Servers

The Ethermint client supports both [gRPC endpoints](https://cosmos.network/rpc) from the SDK and Ethereum's [JSON-RPC](https://eth.wiki/json-rpc/API).

#### Cosmos gRPC and Tendermint RPC

Ethermint exposes gRPC endpoints (and REST) for all the integrated Cosmos-SDK modules. This makes it easier for
wallets and block explorers to interact with the proof-of-stake logic and native Cosmos transactions and queries:

#### Ethereum JSON-RPC server

Ethermint also supports most of the standard web3 [JSON-RPC
APIs](https://eth.wiki/json-rpc/API) to connect with existing web3 tooling.

::: tip
See the list of supported JSON-RPC API [namespaces](https://geth.ethereum.org/docs/rpc/server) and endpoints.
:::

To connect to the JSON-PRC server, start the node with the `--evm-rpc.enable=true` flag and define the namespaces that you would like to run using the `--evm.rpc.api` flag (e.g. `"txpool,eth,web3,net,personal"`. Then, you can point any Ethereum development tooling to `http://localhost:8545` or whatever port you choose with the listen address flag (`--evm-rpc.address`).

## Next {hide}

Process and subscribe to [events](./events.md) via websockets {hide}
