<!--
order: 2
-->

# Architecture

Learn how Ethermint's architecture leverages the Cosmos SDK Proof-of-Stake functionality, EVM compatibility and fast-finality from Tendermint Core's BFT consensus. {synopsis}

## Cosmos SDK

Ethermint enables the full composability and modularity of the [Cosmos SDK](https://docs.cosmos.network/).

## Tendermint Core & the Application Blockchain Interface (ABCI)

Tendermint consists of two chief technical components: a blockchain consensus
engine and a generic application interface. The consensus engine, called
[Tendermint Core](https://docs.tendermint.com/), ensures that the same transactions are recorded on every machine
in the same order. The application interface, called the [Application Blockchain Interface (ABCI)](https://docs.tendermint.com/master/spec/abci/), enables the transactions to be processed in any programming
language.

Tendermint has evolved to be a general purpose blockchain consensus engine that
can host arbitrary application states. Since Tendermint can replicate arbitrary
applications, it can be used as a plug-and-play replacement for the consensus
engines of other blockchains. Ethermint is such an example of an ABCI application
replacing Ethereum's PoW via Tendermint's consensus engine.

Another example of a cryptocurrency application built on Tendermint is the Cosmos
network. Tendermint is able to decompose the blockchain design by offering a very
simple API (ie. the ABCI) between the application process and consensus process.

## EVM module

Ethermint enables EVM compatibility by implementing various components that together support all the EVM state transitions while ensuring the same developer experience as Ethereum:

- Ethereum transaction format as a Cosmos SDK `Tx` and `Msg` interface
- Ethereum's `secp256k1` curve for the Cosmos Keyring
- `StateDB` interface for state updates and queries
- JSON-RPC client for interacting with the EVM

## Next {hide}

Check the available Ethermint [resources](./resources.md) {hide}
