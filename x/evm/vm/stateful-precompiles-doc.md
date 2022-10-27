## Seamless handling of Stateful Precompiles in Ethermint-based chains

### Goal

**Allow stateful precompiles to modify state in Cosmos KV stores seamlessly (i.e. through normal
calls to native modules).**

To reach this goal, we propose several modifications to the dev ux:

- Remove the need to append modified state to journal object and later commit this state to Cosmos
  - This adds extra overhead in calling cosmos native logic, that already exists
- Allow precompile calls from solidity smart contracts to pass in arbitrary routes and args
  - A router will invoke the intended modules and calls, with corresponding args
- State changes to cosmos are made dynamically and if any subset of actions reverts during the eth
  tx, the entire state of cosmos is "automatically" restored.

### Approach

**Use cosmos sdk CacheKVStore to store 'checkpoints' of cosmos state along the way.**

There will be a cached state taken before beginning the eth transaction. In case the EVM
runtime reverts, we can switch back to this state.

Rather than committing or reverting temporary dirty state (journal) to the cosmos store _after_ the
completion of the precompile `Run` function, we are proposing to cache the state _before_ any
precompile call. This way, if any precompile logic throws an error, we can easily revert state to
immediately before the precompile was called.

Use the cosmos `CacheContext` function to take 'snapshots' and commit the cache if the
corresponding logic executes successfully.

Cases in which restoring previous state are necessary, and their corresponding desired behaviors
are below.

- Solidity revert called before calling precompile: effectively no changes are made to Cosmos at
  this point, so no actions are taken
- Precompile throws error: revert all changes to Cosmos state to the state immediately before
  calling the precompile, and then proceed with solidity execution
- Solidity revert call after calling precompile: revert all changes to Cosmos state to the state
  immediately before the eth tx

To implement this, we setup a `BasePrecompile` class abstraction and all stateful precompiles
should inherit from this base class, which has its own `Run` function. The `Run` function, invoked
by Geth EVM interpreter, will call the custom stateful precompile's `RunStateful` function, which
includes the actual state-modifiying module-specific logic.

### Cache Structure

**The max depth of caches is 2**

The pseudocode of caching snapshots of cosmos state during an Eth transaction (specifically during
the execution of `ApplyTransaction()` in Ethermint) is as follows:

- `ethTxCtx, commitTxState := currCtx.CacheContext()`
- `ApplyMessageWithConfig(ethTxCtx): // calls into EVM interpreter with newly created ctx`
  - `for precompileCall in ethTxCtx.evmInterpreter():`
    - `pcCtx, commitPCState := ethTxCtx.CacheContext() // every call is done on ethTxCtx, NOT currCtx`
    - `if precompileCall.Success():`
      - `commitPCState(pcCtx) // commits precompile state to txCtx`
    - `else:`
      - `discard(pcCtx)`
  - `if ethTx.Success():`
    - `commitTxState(ethTxCtx) // commits eth tx state to currCtx`
  - `else:`
    - `discard(ethTxCtx)`

As seen here, the overhead with this approach includes storing extra dirty state in a cache and
writing/discarding this cache for every precompile (grows linearly with number of precompiles
called).