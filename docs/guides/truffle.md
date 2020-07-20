<!--
order: 1
-->

# Truffle

Set up a Truffle Ethermint local development environment. {synopsis}

## Pre-requisite Readings

- [Installation](./../quickstart/installation.md) {prereq}
- [Run a node](./../quickstart/run_node.md) {prereq}

[Truffle](https://www.trufflesuite.com/truffle) is a development framework for deploying and managing [Solidity](https://github.com/ethereum/solidity) smart contracts. In this guide, we will learn how to deploy a contract to a running Ethermint network.

## Install dependencies

First, install the latest Truffle version on your machine globally.

```bash
npm install truffle -g
```

You will also need to install Ethermint. Check this [document](./../quickstart/installation.md) for the full instructions.

## Create Truffle Project

In this step we will create a simple counter contract. Feel free to skip this step if you already have your own compiled contract.

Create a new directory to host the contracts and initialize it

```bash
mkdir ethermint-truffle
cd ethermint-truffle
```

Initialize the Truffle suite with:

```bash
truffle init
```

Create `contracts/Counter.sol` containing the following contract:

```javascript
pragma solidity ^0.5.11;

contract Counter {
  uint256 counter = 0;

  function add() public {
    counter++;
  }

  function subtract() public {
    counter--;
  }

  function getCounter() public view returns (uint256) {
    return counter;
  }
}
```

Compile the contract using the `compile` command:

```bash
truffle compile
```

Create `test/counter_test.js` containing the following tests in Javascript using [Mocha](https://mochajs.org/):

```javascript
const Counter = artifacts.require("Counter")

contract('Counter', accounts => {
  const from = accounts[0]
  let counter

  before(async() => {
    counter = await Counter.new()
  })

  it('should add', async() => {
    await counter.add()
    let count = await counter.getCounter()
    assert(count == 1, `count was ${count}`)
  })
})
```

## Truffle configuration

Open `truffle-config.js` and uncomment the `development` section in `networks`:

```javascript
    development: {
     host: "127.0.0.1",     // Localhost (default: none)
     port: 8545,            // Standard Ethereum port (default: none)
     network_id: "*",       // Any network (default: none)
    },
```

This will allow your contract to connect to your Ethermint local node.

## Start Node and REST server

Start your local node using the following command on the Terminal

```bash
# on the ~/ethermint/ directory
init.sh
```

::: tip
For further information on how to run a node, please refer to [this](./../quickstart/run_node.md) quickstart document.
:::

In another Terminal wintdow/tab, start the [REST and JSON-RPC server](./../quickstart/clients.md#rest-and-tendermint-rpc.md):

```bash
emintcli rest-server --laddr "tcp://localhost:8545" --unlock-key mykey--chain-id 8 --trace
```

## Deploy contract

Back in the Truffle terminal, migrate the contract using

```bash
truffle migrate --network development
```

You should see incoming deployment logs in the Ethermint daemon Terminal tab for each transaction (one to deploy `Migrations.sol` and the oether to deploy `Counter.sol`).

```bash
I[2020-07-15|17:35:59.934] Added good transaction                       module=mempool tx=22245B935689918D332F58E82690F02073F0453D54D5944B6D64AAF1F21974E2 res="&{CheckTx:log:\"[]\" gas_wanted:6721975 }" height=3 total=1
I[2020-07-15|17:36:02.065] Executed block                               module=state height=4 validTxs=1 invalidTxs=0
I[2020-07-15|17:36:02.068] Committed state                              module=state height=4 txs=1 appHash=76BA85365F10A59FE24ADCA87544191C2D72B9FB5630466C5B71E878F9C0A111
I[2020-07-15|17:36:02.981] Added good transaction                       module=mempool tx=84516B4588CBB21E6D562A6A295F1F8876076A0CFF2EF1B0EC670AD8D8BB5425 res="&{CheckTx:log:\"[]\" gas_wanted:6721975 }" height=4 total=1
```

## Run Truffle tests

Now, you can run the Truffle tests using the Ethermint node using the `test` command:

```bash
truffle test --network development

Using network 'development'.


Compiling your contracts...
===========================
> Everything is up to date, there is nothing to compile.



  Contract: Counter
    ✓ should add (5036ms)


  1 passing (10s)
```

## Next {hide}

Learn how to connect Ethermint to [Metamask](./../guides/metamask.md) {hide}
