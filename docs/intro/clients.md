<!--
order: 3
-->

# Clients

Learn about the client supported by your Ethermint node. {synopsis}

## Client Servers

The Ethermint client supports both [gRPC endpoints](https://cosmos.network/rpc) from the SDK and [Ethereum's JSON-RPC](https://eth.wiki/json-rpc/API).

### Cosmos gRPC and Tendermint RPC

Ethermint exposes gRPC endpoints (and REST) for all the integrated Cosmos-SDK modules. This makes it easier for
wallets and block explorers to interact with the proof-of-stake logic and native Cosmos transactions and queries:

### Ethereum JSON-RPC server

Ethermint also supports most of the standard web3 [JSON-RPC
APIs](./../api/JSON-RPC/running_server) to connect with existing web3 tooling.

::: tip
See the list of supported JSON-RPC API [endpoints](./../api/JSON-RPC/endpoints) and [namespaces](./../api/JSON-RPC/namespaces).
:::

To connect to the JSON-PRC server, start the node with the `--json-rpc.enable=true` flag and define the namespaces that you would like to run using the `--evm.rpc.api` flag (e.g. `"txpool,eth,web3,net,personal"`. Then, you can point any Ethereum development tooling to `http://localhost:8545` or whatever port you choose with the listen address flag (`--json-rpc.address`).

<!-- TODO: add Rosetta -->