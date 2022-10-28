## Seamless handling of Stateful Precompiles in Ethermint-based chains

### Goal

**Allow stateful precompiles to modify state in Cosmos KV stores seamlessly (i.e. through normal calls to native modules).**

To reach this goal, we propose several modifications to the dev ux:

- Remove the need to append modified state to journal object and later commit this state to Cosmos
  - This adds extra overhead in calling cosmos native logic, that already exists
- Allow precompile calls from solidity smart contracts to pass in arbitrary routes and args
  - A router will invoke the intended modules and calls, with corresponding args
- State changes to cosmos are made dynamically and if any subset of actions reverts during the eth
  tx, the entire state of cosmos is "automatically" restored.

### Approach

**Use cosmos sdk CacheContext to store 'snapshots' of cosmos state along the way.**

In the same way the ethermint statedb uses journal entries of evm state changes for every _evm call_, a journal entry holding the stateful precompiles' state changes will be added for every _precompile call_. This journal entry is essentially holding a cached cosmos sdk context in which temporary, dirty state changes are added to. Using the same statedb journaling logic, these `CacheContext`s are either reverted during the course of the transaction (by simply removing the cache context) or committed at the end. 

The cosmos `CacheContext` function takes 'snapshots' and is only committed if the calling Solidity call executes successfully. Cases in which restoring previous state are necessary, and their corresponding desired behaviors are below:

- Solidity revert called before calling precompile: effectively no changes are made to Cosmos at
  this point, so no actions are taken
- Precompile throws error: revert all changes to Cosmos state to the state immediately before
  calling the precompile, and then proceed with solidity execution
- Solidity revert call after calling precompile: revert all changes to Cosmos state to the state
  immediately before the Solidity call which invoked the prcompile.

Note, we also setup a `BasePrecompile` class abstraction and all stateful precompiles
should inherit from this base class, which has its own `Run` function. Stateful precompiles should only implement the `RunStateful` function, which is given a cosmos sdk context for state changes to be made on.

All of these revert behaviors are handled seamlessly and away from the perspective of a developer writing a stateful precompile. The `RunStateful` function can invoke module functions passing in the sdk context; there is no consideration of handling reverts from the perspective of a developer. 

### Cache Structure

**The max depth of caches is equal to the number of precompile calls per transaction**

Using this approach, cache contexts are nested, but never nested on invalid cache contexts. This is because when creating a new cache context for an additional precompile call, only the last valid (non-reverted) cosmos state context is used to nest from. Whenever a revert is called by the EVM interpreter, the statedb `RevertToSnapshot` will omit all necessary cache contexts and future precompile calls will only nest from the previous valid cache context. 

The overhead of this method is holding cache contexts proportional to the number of precompile calls in a single transaction. However, on reverts, unneeded cache contexts are discarded by the garbage collector (as the journal holds no pointer to these objects). Additionally, at the end of the transaction, only the latest valid journal entry (and its corresponding cache context) are committed to Cosmos state. 

With this approach, no unwrapping of contexts are necessary. By tracking the state of the call stack with a journal, we can handle all reverts in the same way the statedb currently does. 

### Necessary Changes

**In order to facilitate storing CacheContexts as a journal entry in the ethermint statedb, these crucial changes were made:**

- `ExtStateDB`: external statedb, implemented by the current ethermint statedb, which requires functions for getting cosmos sdk contexts for stateful precompiles and appending/reverting external journal entries
- `ExtJournalEntry`: external journal entry, a type that actually stores a cache context, for our use case
- `statedb.validExtJournalIndexes`: this is a stack holding the indexes of all external journal entries in the `statedb.journal.entries` slice. This is necessary to handle reverts and correctly return the latest valid cache context.
- `(e EVM) RunPrecompiledContract`: this function is modified to handle stateful precompiles, and properly sets up the external journal entry and appends it to the statedb. In the future, this can all be handled by the `BasePrecompile`, so all stateful precompiles automatically "inherit" this functionality. 
- Shadowing `Call`, `CallCode`, `StaticCall`, `DelegateCall`: these geth evm functions must be shadowed to hold custom precompiles and properly call stateful ones during execution. Additionally, their corresponding opcodes in the jump table (`opCall`, `opCallCode`, etc.) and the `EVMInterpreter` `Run` method must be shadowed. All of this adds up to essentially rewriting the geth core vm. A simple alternative is to maintain a simple fork which just allows the geth evm to call custom precompiles.