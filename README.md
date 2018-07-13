[![](https://godoc.org/github.com/cosmos/ethermint?status.svg)](http://godoc.org/github.com/cosmos/ethermint)  [![Go Report Card](https://goreportcard.com/badge/github.com/cosmos/ethermint)](https://goreportcard.com/report/github.com/cosmos/ethermint)

# Ethermint

__**WARNING:**__ Ethermint is under VERY ACTIVE DEVELOPMENT and should be treated as pre-alpha software. This means it is not meant to be run in production, its APIs are subject to change without warning and should not be relied upon, and it should not be used to hold any value. We will remove this warning when we have a release that is stable, secure, and properly tested.

### What is it?

`ethermint` will be an implementation of the EVM that runs on top of [`tendermint`](https://github.com/tendermint/tendermint) consensus, a Proof of Stake system. This project has as its primary goals:

- [Hard Spoon](https://blog.cosmos.network/introducing-the-hard-spoon-4a9288d3f0df) enablement: This is the ability to take a token from the Ethereum mainnet and "spoon" (shift) the balances over to another network. This feature is intended to make it easy for applications that require more transactions than the Ethereum main chain can provide to move their code over to a compatible chain with much more capacity.
-  Web3 Compatibility: In order enable applications to be moved over to an ethermint chain existing tooling (i.e. web3 compatable clients) need to be able to interact with `ethermint`.

### Implementation

- [x] Have a working implementation that can parse and validate the existing ETH Chain and persist it in a Tendermint store
- [ ] Benchmark this implementation to ensure performance
- [ ] Allow the Ethermint EVM to interact with other [Cosmos SDK modules](https://github.com/cosmos/cosmos-sdk/blob/master/docs/core/app3.md)
- [ ] Implement the Web3 APIs as a Cosmos Light Client for Ethermint
- [ ] Ethermint is a full Cosmos SDK application and can be deployed as it's own zone

### Building Ethermint

To build, execute the following commands:

```bash
# To build the binary and put the results in ./build
$ make tools deps build

# To build the project and install it in $GOBIN
$ make tools deps install
```

### Using Ethermint to parse Mainnet Ethereum blocks

There is an included Ethereum Mainnet blockchain file in `data/blockchain` that provides an easy way to run the demo of parsing Mainnet Ethereum blocks. The dump in `data/` only includes up to block `97638`. To run this, type the following command:

```bash
$ go run main.go copied.go
balance of 0x756F45E3FA69347A9A973A725E3C98bC4db0b5a0: 200000000000000000000
commitID after genesis: CommitID{[235 78 80 238 156 235 50 19 95 118 247 9 106 207 72 45 127 238 223 177]:1}
genesis state root hash: 0000000000000001000000000000000000000000000000000000000000000000
processed 10000 blocks, time so far: 3.076565301s
processed 20000 blocks, time so far: 6.437135359s
processed 30000 blocks, time so far: 10.145579892s
processed 40000 blocks, time so far: 14.538473343s
processed 50000 blocks, time so far: 20.924373963s
processed 60000 blocks, time so far: 28.486701074s
processed 70000 blocks, time so far: 40.029318946s
processed 80000 blocks, time so far: 48.056637508s
processed 90000 blocks, time so far: 57.049310537s
processed 97638 blocks
balance of one of the genesis investors: 200000000000000000000
root500: 00000000000001f5000000000000000000000000000000000000000000000000
investor's balance after block 500: 200000000000000000000
miner of block 501's balance after block 500: 0
miner of block 501's balance after block 501: 5000000000000000000
```

### Community

The following chat channels and forums are a great spot to ask questions about Ethermint:

- [Cosmos Riot Chat Channel](https://riot.im/app/#/group/+cosmos:matrix.org)
- Cosmos Forum [![Discourse status](https://img.shields.io/discourse/https/forum.cosmos.network/status.svg)](https://forum.cosmos.network)
