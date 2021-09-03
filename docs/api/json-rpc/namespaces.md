<!--
order: 2
-->

# Namespaces

Check the JSON-RPC namespaces supported on Ethermint. {synopsis}

## Pre-requisite Readings

- [Geth JSON-RPC Namespaces](https://geth.ethereum.org/docs/rpc/server) {prereq}

## Ethereum Namespaces

| Namespace                                  | Description                                                                                                                                                                                                                  | Supported | Enabled by Default |
|--------------------------------------------|------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|-----------|--------------------|
| [`eth`](./endpoints#eth-methods)           | Ethermint provides several extensions to the standard `eth` JSON-RPC namespace.                                                                                                                                              | ✔         | ✔                  |
| [`web3`](./endpoints#web3-methods)         | The `web3` API provides utility functions for the web3 client.                                                                                                                                                               | ✔         | ✔                  |
| [`net`](./endpoints#net-methods)           | The `net` API provides access to network information of the node                                                                                                                                                             | ✔         | ✔                  |
| `clique`                                   | The `clique` API provides access to the state of the clique consensus engine. You can use this API to manage signer votes and to check the health of a private network.                                                      | ❌         |                    |
| `debug`                                    | The `debug` API gives you access to several non-standard RPC methods, which will allow you to inspect, debug and set certain debugging flags during runtime.                                                                 | ✔         |                    |
| `les`                                      | The `les` API allows you to manage LES server settings, including client parameters and payment settings for prioritized clients. It also provides functions to query checkpoint information in both server and client mode. | ❌         |                    |
| [`miner`](./endpoints#miner-methods)       | The `miner` API allows you to remote control the node’s mining operation and set various mining specific settings.                                                                                                           | ✔         | ❌                  |
| [`txpool`](./endpoints#txpool-methods)     | The `txpool` API gives you access to several non-standard RPC methods to inspect the contents of the transaction pool containing all the currently pending transactions as well as the ones queued for future processing.    | ✔         | ❌                  |
| `admin`                                    | The `admin` API gives you access to several non-standard RPC methods, which will allow you to have a fine grained control over your nodeinstance, including but not limited to network peer and RPC endpoint management.     | ❌         |                    |
| [`personal`](./endpoints#personal-methods) | The `personal` API manages private keys in the key store.                                                                                                                                                                    | ✔         | ❌                  |
