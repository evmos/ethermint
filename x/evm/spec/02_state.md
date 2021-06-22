<!--
order: 2
-->

# State

The `x/evm` module keeps the following objects in state:

|                 | Key                                               | Value                     |
|-----------------|---------------------------------------------------|---------------------------|
| Block Height    | `[]byte{1} + []byte(block.Hash)`                  | `BigEndian(block.Height)` |
| Bloom           | `[]byte{2} + []byte(block.Height)`                | `[]byte(Bloom)`           |
| Tx Logs         | `[]byte{3} + []byte(tx.Hash)`                     | `amino([]Log)`            |
| Account Code    | `[]byte{4} + []byte(code.Hash)`                   | `[]byte(Code)`            |
| Account Storage | `[]byte{5} + []byte(address) + []byte(state.Key)` | `[]byte(state.Value)`     |
| Chain Config    | `[]byte{6}`                                       | `amino(ChainConfig)`      |

## `CommitStateDB`

`StateDB`s within the ethereum protocol are used to store anything within the IAVL tree. `StateDB`s
take care of caching and storing nested states. It's the general query interface to retrieve
contracts and accounts

The Ethermint `CommitStateDB` is a concrete type that implements the EVM `StateDB` interface.
Instead of using a trie and database for querying and persistence, the `CommitStateDB` uses
`KVStores` (key-value stores) and Cosmos SDK `Keeper`s to facilitate state transitions.

The `CommitStateDB` contains a store key that allows the DB to write to a concrete subtree of the
multistore that is only accessible to the EVM module.

+++ https://github.com/tharsis/ethermint/blob/v0.3.1/x/evm/types/statedb.go#L33-L85

The functionalities provided by the Ethermint `StateDB` are:

* CRUD of `stateObject`s and accounts:
  * Balance
  * Code
  * Nonce
  * State
* EVM module parameter getter and setter
* State transition logic
  * Preparation: transaction index and hash, block hash
* CRUD of transaction logs
* Aggregate queries
* Snapshot state
  * Identify current state with a revision
  * Revert state to a given revision
* State transition and persistence
  * Preparation: tx and block context
  * Commit state objects
  * Finalise state objects
  * Export state for upgrades
* Auxiliary functions
  * Copy state
  * Reset state

## State Objects

State objects are used by the VM which is unable to deal with database-level errors. Any error that occurs during a database read is memoized here and will eventually be returned by `StateDB.Commit`.

The Ethermint `stateObject` is a concrete type that mimics the functionality from the `go-ethereum`
private `stateObject` type. It keeps track of the interim values for the contract bytecode, storage
state and balance of an `EthAccount`.

The storage entries (original and "dirty") for each state object are represented as slices instead
of maps since latter can cause non-deterministic block app hashes, which result in the chain
halting.

When a `stateObject` is committed during `EndBlock`. It sets sets the account contract code to store, as well as the dirty storage state. The account's nonce and the account balance are updated by calling the `auth` and `bank` module setter functions, respectively.

+++ https://github.com/tharsis/ethermint/blob/v0.3.1/x/evm/types/state_object.go#L49-L81

The functionalities provided by the Ethermint `stateObject` are:

* Storage state getter and setter (temporary)
* Contract bytecode getter and setter (temporary)
* Balance getter and setter (temporary)
* Balance accounting (temporary)
* Account nonce and address getter and setter (temporary)
* Auxiliary functions: copy, RLP encoding, empty
* Commit state object (final)
