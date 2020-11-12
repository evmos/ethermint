<!--
order: 3
-->

# Gas and Fees

Learn about the differences between `Gas` and `Fees` in Ethereum and Cosmos. {synopsis}

## Pre-requisite Readings

- [Cosmos SDK Gas](https://docs.cosmos.network/master/basics/gas-fees.html) {prereq}
- [Ethereum Gas](https://ethereum.org/en/developers/docs/gas/) {prereq}

The concept of Gas represents the amount of computational effort required to execute specific operations on the state machine.

Gas was created on Ethereum to disallow the EVM (Ethereum Virtual Machine) from running infinite
loops by allocating a small amount of monetary value into the system. A unit of gas, usually in a
form as a fraction of the native coin, is consumed for every operation on the EVM and requires a
user to pay for these operations. These operations consist in state transitions such as sending a
transaction or calling a contract.

Exactly like Ethereum, Cosmos utilizes the concept of gas and this is how Cosmos tracks the resource
usage of operations during execution. Operations on Cosmos are represented as read or writes done to the chain's store.

In Cosmos, a fee is calculated and charged to the user during a message execution. This fee is
calculated from the sum of all gas consumed in an message execution:

```
fee = gas * gas price
```

In both networks, gas is used to make sure that operations do not require an excess amount of
computational power to complete and as a way to deter bad-acting users from spamming the network.

## Cosmos SDK `Gas`

In the Cosmos SDK, gas is tracked in the main `GasMeter` and the `BlockGasMeter`:

- `GasMeter`: keeps track of the gas consumed during executions that lead to state transitions. It is reset on every transaction  execution.
- `BlockGasMeter`: keeps track of the gas consumed in a block and enforces that the gas does not go over a predefined limit. This limit is defined in the Tendermint consensus parameters and can be changed via governance parameter change proposals.

More information regarding gas in Cosmos SDK can be found [here](https://docs.cosmos.network/master/basics/gas-fees.html).

## Matching EVM Gas consumption

Ethermint is an EVM-compatible chain that supports Ethereum Web3 tooling. For this reason, gas
consumption must be equitable in order to accurately calculate the state transition hashes and exact
the behaviour that would be seen on the main Ethereum network (main net).

In Cosmos, there are types of operations that are not triggered by transactions that can also result in state transitions. Concrete examples are the  `BeginBlock` and `EndBlock` operations and the `AnteHandler` checks, which might also read and write to the store before running the state transition from a transaction.

### `BeginBlock` and `EndBlock`

These operations are defined by the Tendermint Core's Application Blockchain Interface (ABCI) and are defined by each Cosmos SDK module. As their name suggest, they are executed at the beginning and at the end of each block processing respectively (i.e pre and post transaction execution). Since these operations are not reflected on Ethereum, to match the the gas consumption we reset the main `GasMeter` to 0 on Ethermint's EVM module.

### `AnteHandler`

The Cosmos SDK [`AnteHandler`](https://docs.cosmos.network/master/basics/gas-fees.html#antehandler)
performs basic checks prior to transaction execution. These checks are usually signature
verification, transaction field validation, transaction fees, etc.

Because the gas calculated in Ethermint is done by the `IntrinsicGas` method from go-ethereum, a
special `AnteHandler` that is customized for EVM transaction fee verification is required. This
allows Ethermint to generate the expected gas costs for operations done in the network and scale the
gas costs as it would in the Ethereum network.

## Gas Refunds

In Ethereum, gas can be specified prior to execution and the remaining gas will be refunded back to
the user if any gas is left over - should fail with out of gas if not enough gas was provided. In
Ethermint, the concept of gas refunds does not exist and the fees paid is not refunded in part back
to the user. The fees exacted on a transaction will be collected by the validator and no refunds are
issued. Thus, it is extremely important to use the correct gas.

To prevent overspending on fees, providing the `--gas-adjustment` flag for a cosmos transactions
will determine the fees automatically. Also the `eth_estimateGas` rpc call can be used to manually
get the correct gas costs for a transaction.

## 0 Fee Transactions

In Cosmos, a minimum gas price is not enforced by the `AnteHandler` as the `min-gas-prices` is
checked against the local node/validator. In other words, the minimum fees accepted are determined
by the validators of the network, and each validator can specify a different value for their fees.
This potentially allows end users to submit 0 fee transactions if there is at least one single
validator that is willing to include transactions with `0` gas price in their blocks proposed.

For this same reason, in Ethermint it is possible to send transactions with `0` fees for transaction
types other than the ones defined by the `evm` module. EVM module transactions cannot have `0` fees
as gas is required inherently by Ethereum. This check is done by the evm transactions
`ValidateBasic` function as well as on the custom `AnteHandler` defined by Ethermint.

## Next {hide}

Learn about the [Photon](./photon.md) token {hide}
