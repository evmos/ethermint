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

- `GasMeter`: keeps track of the gas consumed during executions that lead to state transitions. It is reset on every transaction execution.
- `BlockGasMeter`: keeps track of the gas consumed in a block and enforces that the gas does not go over a predefined limit. This limit is defined in the Tendermint consensus parameters and can be changed via governance parameter change proposals.

More information regarding gas in Cosmos SDK can be found [here](https://docs.cosmos.network/master/basics/gas-fees.html).

## Matching EVM Gas consumption

Ethermint is an EVM-compatible chain that supports Ethereum Web3 tooling. For this reason, gas
consumption must be equitable with other EVMs, most importantly Ethereum.

The main difference between EVM and Cosmos state transitions, is that the EVM uses a [gas table](https://github.com/ethereum/go-ethereum/blob/master/params/protocol_params.go) for each OPCODE, whereas Cosmos uses a `GasConfig` that charges gas for each CRUD operation by setting a flat and per-byte cost for accessing the database.

+++ https://github.com/cosmos/cosmos-sdk/blob/3fd376bd5659f076a4dc79b644573299fd1ec1bf/store/types/gas.go#L187-L196

In order to match the the gas consumed by the EVM, the gas consumption logic from the SDK is ignored, and instead the gas consumed is calculated by subtracting the state transition leftover gas plus refund from the gas limit defined on the message.

To ignore the SDK gas consumption, we reset the transaction `GasMeter` count to 0 and manually set it to the `gasUsed` value computed by the EVM module at the end of the execution.

+++ https://github.com/tharsis/ethermint/blob/098da6d0cc0e0c4cefbddf632df1057383973e4a/x/evm/keeper/state_transition.go#L188

### `AnteHandler`

The Cosmos SDK [`AnteHandler`](https://docs.cosmos.network/master/basics/gas-fees.html#antehandler)
performs basic checks prior to transaction execution. These checks are usually signature
verification, transaction field validation, transaction fees, etc.

Regarding gas consumption and fees, the `AnteHandler` checks that the user has enough balance to
cover for the tx cost (amount plus fees) as well as checking that the gas limit defined in the
message is greater or equal than the computed intrinsic gas for the message.

## Gas Refunds

In the EVM, gas can be specified prior to execution. The totality of the gas specified is consumed at the beginning of the execution (during the `AnteHandler` step) and the remaining gas is refunded back to
the user if any gas is left over after the execution. Additionally the EVM can also define gas to be refunded back to the user but those will be capped to a fraction of the used gas depending on the fork/version being used.

## 0 Fee Transactions

In Cosmos, a minimum gas price is not enforced by the `AnteHandler` as the `min-gas-prices` is
checked against the local node/validator. In other words, the minimum fees accepted are determined
by the validators of the network, and each validator can specify a different minimum value for their fees.
This potentially allows end users to submit 0 fee transactions if there is at least one single
validator that is willing to include transactions with `0` gas price in their blocks proposed.

For this same reason, in Ethermint it is possible to send transactions with `0` fees for transaction
types other than the ones defined by the `evm` module. EVM module transactions cannot have `0` fees
as gas is required inherently by the EVM. This check is done by the EVM transactions stateless validation
(i.e `ValidateBasic`) function as well as on the custom `AnteHandler` defined by Ethermint.

## Gas estimation

Ethereum provides a JSON-RPC endpoint `eth_estimateGas` to help users set up a correct gas limit in their transactions. 

Unfortunately, we cannot make use of the SDK `tx simulation` for gas estimation because the pre-check in the Ante Handlers would require a valid signature, and the sender balance to be enough to pay for the gas. But in Ethereum, this endpoint can be called without specifying any sender address.

For that reason, a specific query API `EstimateGas` is implemented in Ethermint. It will apply the transaction against the current block/state and perform a binary search in order to find the optimal gas value to return to the user (the same transaction will be applied over and over until we find the minimum gas needed before it fails). The reason we need to use a binary search is that the gas required for the
transaction might be higher than the value returned by the EVM after applying the transaction, so we need to try until we find the optimal value.

A cache context will be used during the whole execution to avoid changes be persisted in the state.

+++ https://github.com/tharsis/ethermint/blob/098da6d0cc0e0c4cefbddf632df1057383973e4a/x/evm/keeper/grpc_query.go#L100

## Next {hide}

Learn about the different types of [tokens](./tokens.md) available {hide}
