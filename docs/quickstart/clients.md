<!--
order: 6
-->

# Clients

Learn how to connect a client to a running node. {synopsis}

## Pre-requisite Readings

- [Run a Node](./run_node.md) {prereq}

## Client Integrations

### Command Line Interface

Ethermint is integrated with a CLI client that can be used to send transactions and query the state from each module.

```bash
# available query commands
ethermintcli query -h

# available transaction commands
ethermintcli tx -h
```

### Client Servers

The Ethermint client supports both [REST endpoints](https://cosmos.network/rpc) from the SDK and Ethereum's [JSON-RPC](https://eth.wiki/json-rpc/API).

#### REST and Tendermint RPC

Ethermint exposes REST endpoints for all the integrated Cosmos-SDK modules. This makes it easier for wallets and block explorers to interact with the proof-of-stake logic.

To run the REST Server, you need to run the Ethermint daemon (`ethermintd`) and then execute (in another
process):

```bash
ethermintcli rest-server --laddr "tcp://localhost:8545" --unlock-key $KEY --chain-id $CHAINID --trace
```

You should see the logs from the REST and the RPC server.

```bash
I[2020-07-17|16:54:35.037] Starting application REST service (chain-id: "8")... module=rest-server
I[2020-07-17|16:54:35.037] Starting RPC HTTP server on 127.0.0.1:8545   module=rest-server
```

#### Ethereum JSON-RPC server

Ethermint also supports most of the standard web3 [JSON-RPC
APIs](https://eth.wiki/json-rpc/API) to connect with existing web3 tooling.

::: tip
Some of the JSON-RPC API [namespaces](https://geth.ethereum.org/docs/rpc/server) are currently under development.
:::

To connect to the JSON-PRC server, use the `rest-server` command as shown on the section above. Then, you can point any Ethereum development tooling to `http://localhost:8545` or whatever port you choose with the listen address flag (`--laddr`).

For further information JSON-RPC calls, please refer to [this](../basics/json_rpc.md)  document.

## Next {hide}

Process and subscribe to [events](./events.md) via websockets {hide}
