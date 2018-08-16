# Introduction

## Preliminary

### Tendermint Core & the Application Blockchain Interface (ABCI)

Tendermint consists of two chief technical components: a blockchain consensus
engine and a generic application interface. The consensus engine, called
Tendermint Core, ensures that the same transactions are recorded on every machine
in the same order. The application interface, called the Application Blockchain
Interface (ABCI), enables the transactions to be processed in any programming
language.

Tendermint has evolved to be a general purpose blockchain consensus engine that
can host arbitrary application states. Since Tendermint can replicate arbitrary
applications, it can be used as a plug-and-play replacement for the consensus
engines of other blockchains. Ethermint is such an example of an ABCI application
replacing Ethereum's PoW via Tendermint's consensus engine.

Another example of a cryptocurrency application built on Tendermint is the Cosmos
network. Tendermint is able to decompose the blockchain design by offering a very
simple API (ie. the ABCI) between the application process and consensus process.

## What is Ethermint

Ethermint is a high throughput PoS blockchain that is fully compatible and
interoperable with Ethereum. In other words, it allows for running vanilla Ethereum
on top of [Tendermint](https://github.com/tendermint/tendermint) consensus via
the [Cosmos SDK](https://github.com/cosmos/cosmos-sdk/). This allows developers
to have all the desired features of Ethereum, while at the same time benefit
from Tendermint’s PoS implementation.

Here’s a glance at some of the key features of Ethermint:

* Web3 compatibility
* High throughput
* Horizontal scalability
* Transaction finality

Ethermint achieves these key features by implementing Tendermint's ABCI application
interface, leveraging modules and mechanisms implemented by the Cosmos SDK, utilizing
[Geth](https://github.com/ethereum/go-ethereum) as a library by implementing all
necessary interfaces, and finally by exposing a fully compatible Web3 RPC layer
allowing developers to leverage existing Ethereum ecosystem tooling and software
to seamlessly deploy smart contracts and interact with the rest of the Cosmos
ecosystem!
