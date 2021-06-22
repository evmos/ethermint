<!--
order: 1
-->

# Concepts

## EVM

The Ethereum Virtual Machine [EVM](https://ethereum.org/en/developers/docs/evm/)  is a state machine
that provides the necessary tools to run or create a contract on a given state.

## State DB

The `StateDB` interface from geth represents an EVM database for full state querying of both
contracts and accounts. The concrete type that fulfills this interface on Ethermint is the
`CommitStateDB`.

## State Object

A `stateObject` represents an Ethereum account which is being modified.
The usage pattern is as follows:

* First you need to obtain a state object.
* Account values can be accessed and modified through the object.

## Genesis State

The `x/evm` module `GenesisState` defines the state necessary for initializing the chain from a previous exported height.

+++ https://github.com/tharsis/ethermint/blob/v0.3.1/x/evm/types/genesis.go#L14-L20

### Genesis Accounts

The `GenesisAccount` type corresponds to an adaptation of the Ethereum `GenesisAccount` type. Its
main difference is that the one on Ethermint uses a custom `Storage` type that uses a slice instead
of maps for the evm `State` (due to non-determinism), and that it doesn't contain the private key
field.

It is also important to note that since the `auth` and `bank` SDK modules manage the accounts and
balance state,  the `Address` must correspond to an `EthAccount` that is stored in the `auth`'s
module `AccountKeeper` and the balance must match the balance of the `EvmDenom` token denomination
defined on the `GenesisState`'s `Param`. The values for the address and the balance amount maintain
the same format as the ones from the SDK to make manual inspections easier on the genesis.json.

+++ https://github.com/tharsis/ethermint/blob/v0.3.1/x/evm/types/genesis.go#L22-L30

### Transaction Logs

On every Ethermint transaction, its result contains the Ethereum `Log`s from the state machine
execution that are used by the JSON-RPC Web3 server for for filter querying. Since Cosmos upgrades
don't persist the transactions on the blockchain state, we need to persist the logs the EVM module
state to prevent the queries from failing.

`TxsLogs` is the field that contains all the transaction logs that need to be persisted after an
upgrade. It uses an array instead of a map to ensure determinism on the iteration.

+++ https://github.com/tharsis/ethermint/blob/v0.3.1/x/evm/types/logs.go#L12-L18

### Chain Config

The `ChainConfig` is a custom type that contains the same fields as the go-ethereum `ChainConfig`
parameters, but using `sdk.Int` types instead of `*big.Int`. It also defines additional YAML tags
for pretty printing.

The `ChainConfig` type is not a configurable SDK `Param` since the SDK does not allow for validation
against a previous stored parameter values or `Context` fields. Since most of this type's fields
rely on the block height value, this limitation prevents the validation of of potential new
parameter values against the current block height (eg: to prevent updating the config block values
to a past block).

If you want to update the config values, use an software upgrade procedure.

+++ https://github.com/tharsis/ethermint/blob/v0.3.1/x/evm/types/chain_config.go#L16-L45

### Params

See the [params](07_params.md) document for further information about parameters.
