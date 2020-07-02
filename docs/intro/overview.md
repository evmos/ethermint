<!--
order: 1
-->

# High-level Overview

## What is Ethermint

Ethermint is a high throughput PoS blockchain that is fully compatible and
interoperable with Ethereum. In other words, it allows for running vanilla Ethereum
on top of [Tendermint](https://github.com/tendermint/tendermint) consensus via
the [Cosmos SDK](https://github.com/cosmos/cosmos-sdk/). This allows developers
to have all the desired features of Ethereum, while at the same time benefit
from Tendermint’s PoS implementation. Also, because it is built on top of the
Cosmos SDK, it will be able to exchange value with the rest of the Cosmos Ecosystem.

Here’s a glance at some of the key features of Ethermint:

* Web3 compatibility
* High throughput
* Horizontal scalability
* Transaction finality

Ethermint enables these key features through:

* Implementing Tendermint's ABCI application interface to manage the base Blockchain
* Leveraging [modules](https://github.com/cosmos/cosmos-sdk/tree/master/x/) and other mechanisms implemented by the Cosmos SDK
* Utilizing [`geth`](https://github.com/ethereum/go-ethereum) as a library to avoid code reuse and improve maintainability
* Exposing a fully compatible Web3 RPC layer for interacting with the system

The sum of these features allows developers to leverage existing Ethereum ecosystem
tooling and software to seamlessly deploy smart contracts which interact with the rest of the Cosmos
ecosystem!

## Next {hide}

Learn about Ethermint's [architecture](./architectures.md) {hide}
