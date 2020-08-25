<!--
order: 1
-->

# High-level Overview

## What is Ethermint

Ethermint is a scalable, high-throughput Proof-of-Stake blockchain that is fully compatible and
interoperable with Ethereum. It's built using the [Cosmos SDK](https://github.com/cosmos/cosmos-sdk/) which runs on top of [Tendermint Core](https://github.com/tendermint/tendermint) consensus engine.

Ethermint allows for running vanilla Ethereum as a [Cosmos](https://cosmos.network/) application-specific blockchain. This allows developers
to have all the desired features of Ethereum, while at the same time, benefit
from Tendermint’s PoS implementation. Also, because it is built on top of the
Cosmos SDK, it will be able to exchange value with the rest of the Cosmos Ecosystem through the Inter Blockchain Communication Protocol (IBC).

### Features

Here’s a glance at some of the key features of Ethermint:

* Web3 compatibility
* High throughput via [Tendermint Core](https://github.com/tendermint/tendermint)
* Horizontal scalability via [IBC](https://github.com/cosmos/ics)
* Fast transaction finality
* [Hard Spoon](https://blog.cosmos.network/introducing-the-hard-spoon-4a9288d3f0df)

Ethermint enables these key features through:

* Implementing Tendermint Core's ABCI application interface to manage the blockchain
* Leveraging [modules](https://github.com/cosmos/cosmos-sdk/tree/master/x/) and other mechanisms implemented by the Cosmos SDK
* Utilizing [`geth`](https://github.com/ethereum/go-ethereum) as a library to avoid code reuse and improve maintainability.
* Exposing a fully compatible Web3 RPC layer for interacting with existing Ethereum clients and tooling (Metamask, Remix, Truffle, etc).

The sum of these features allows developers to leverage existing Ethereum ecosystem tooling and
software to seamlessly deploy smart contracts which interact with the rest of the Cosmos ecosystem!

## Next {hide}

Learn about Ethermint's [architecture](./architectures.md) {hide}
