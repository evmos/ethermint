# ADR 001: State

## Changelog

- 2021-05-15: first draft

## Status

DRAFT, Not Implemented

## Abstract

The current ADR proposes a state machine breaking change to the EVM module state operations
(`Keeper`, `StateDB` and `StateTransition`) with the goal of reducing code maintainance, increase
performance, and document all the transaction and state cycles and flows.

## Context

<!-- > This section describes the forces at play, including technological, political, social, and project local. These forces are probably in tension, and should be called out as such. The language in this section is value-neutral. It is simply describing facts. It should clearly explain the problem and motivation that the proposal aims to resolve. -->

This ADR addresses the issues of 3 different components of the EVM state: the `StateDB` interface,
the live `stateObject` accounts, and the `StateTransition` functionality. These issues are outlined
below in the section for each corresponding component:

### `StateDB`

In order to execute state transitions, the EVM receives a reference to a database interface to
perform CRUD operations on accounts, balances, code and state storage, among other state queries.
This database interface is defined by go-ethereum's `vm.StateDB`, which is currently implemented
using the `CommitStateDB` concrete type.

The `CommitStateDB` performs state updates by having a direct access to the `sdk.Context`, the evm's
`sdk.StoreKey` and external `Keepers` for account and balances. Currently, the context field needs
to be set on every block or state transition using `WithContext(ctx)` in order to pass the updated
block and transaction data to the `CommitStateDB`.

However, traditionally in Cosmos SDK-based chains, the `Keeper` type has been the de-facto abstraction
that manages access the key-value store (`KVStore`) owned by the module through the store key.
`Keepers` usually hold a reference to external module `Keepers` to perform functionality outside of
the scope of their module.

In the existing architecture of the EVM module, both `CommitStateDB` and `Keeper` have access to
state.

### State Objects

The `CommitStateDB` also holds references of `stateObjects`, defined as "live ethereum consensus
accounts (i.e any balance, account nonces or storage) which will get modified while processing a
state transition".

Upon a state transition, these objects will be modified and marked as 'dirty' (a.k.a stateless
update) on the `CommitStateDB`. Then, at every `EndBlock`, the state of these modified objects will
be 'finalized' and commited to the store, resetting all the dirty list of objects.

The core issue arises when a chain that uses the EVM module can have also have their account and
balances updated through operations from other modules. This means that an EVM state object can be
modified through an EVM transaction (`evm.MsgEthereumTx`) and other transactions like `bank.MsgSend`
or `ibctransfer.MsgTransfer`. This can lead to unexpected behaviors like state overwrites, due to
the current behaviour that caches the dirty state on the EVM instead of commiting any changes
directly.

### State Transition

A general EVM state transition is performed by calling the ethereum `vm.EVM` `Create` or `Call` functions, depending on wheather the transaction creates a contract or performs a transfer or call to a given contract.

In the case of the `x/evm` module, it currently uses a modified version of Geth's `TransitionDB`, that wraps these two `vm.EVM` methods. The reason for using this modified function, is due to several reasons:

  1. The use of `sdk.Msg` (`MsgEthereumTx`) instead of the ethereum `core.Message` type for the `vm.EVM` functions, preventing the direct use of the `core.ApplyMessage`.
  2. The use of custom gas accounting through the transaction `GasMeter` available on the `sdk.Context` to consume the same amount of gas as on Ethereum.
  3. Simulate logic via ABCI `CheckTx`, that prevents the state from being finalized.

## Decision

<!-- > This section describes our response to these forces. It is stated in full sentences, with active voice. "We will ..." -->

### `StateDB`

The `CommitStateDB` type will be removed in favor turning the module's `Keeper` into a `StateDB`
concrete implementation.

```go
// Keeper now fully implements the StateDB interface
var _ vm.StateDB = (*Keeper)(nil)

// Keeper defines the EVM module state keeper for CRUD operations. 
// It also implements the go-ethereum vm.StateDB interface. Instead of using
// a trie and database for querying and persistence, the Keeper uses KVStores
// and external Keepers to facilitate state transitions for accounts and balance
// accounting.
type Keeper struct {
  // store key and encoding codec
  // external module keepers (account, bank, etc) and params subspace
  // cache fields and sdk.Context (reset every block)
  // other CommitStateDB fields (journal, accessList, etc)
}
```

This means that a `Keeper` pointer will now directly be passed to the `vm.EVM` for accessing the state and performing state transitions.

The ABCI `BeginBlock` and `EndBlock` are have now been refactored to only (1) reset cached fields, and (2) keep track of internal mappings (hashes, height, etc).

```go
func (k *Keeper) BeginBlock(ctx sdk.Context, req abci.RequestBeginBlock) {
  // ...

  // reset cache values and context
  k.ResetCacheFields(ctx)
}

func (k Keeper) EndBlock(ctx sdk.Context, req abci.RequestEndBlock) []abci.ValidatorUpdate {
  // NOTE: UpdateAccounts, Commit and Reset execution steps have been removed in favor of directly
  // updating the state.

  // set the block bloom filter bytes to store
  bloom := ethtypes.BytesToBloom(k.Bloom.Bytes())
  k.SetBlockBloom(ctx, req.Height, bloom)

  return []abci.ValidatorUpdate{}
}
```

### State Objects

The `stateObject` type will be completely removed in favor of updating the store directly through
the use of the auth `AccountKeeper` and the bank `Keeper`. For the storage `State` and `Code`, the
evm module `Keeper` will store these values directly on the KVStore using the EVM module store key
and corresponding prefix keys.

For accounts marked as 'suicided', a new relationship will be added to the `Keeper` to map `Address
(bytes) -> suicided (bool)`.

```go
// HasSuicided implements the vm.StoreDB interface
func (k Keeper) HasSuicided(address common.Address) bool {
  store := prefix.NewStore(k.ctx.KVStore(csdb.storeKey), KeyPrefixSuicide)
  key := types.KeySuicide(address.Bytes())
  return store.Has(key)
}

// Suicide implements the vm.StoreDB interface
func (k Keeper) Suicide(address common.Address) bool {
  store := prefix.NewStore(k.ctx.KVStore(csdb.storeKey), KeyPrefixSuicide)
  key := types.KeySuicide(address.Bytes())
  store.Set(key, []byte{0x1})
  return true
}
```

### State Transition

The state transition logic will be refactored to use the `ApplyMessage` function from the `core/`
package of go-ethereum as the backbone. This method calls creates a go-ethereum `StateTransition`
instance and, as it name implies, applies a Ethereum message to execute it and update the state.
This `ApplyMessage` call will be wrapped in the `Keeper`'s `TransitionDb` function, which will
generate the required arguments for this call (EVM, chain config, and gas pool), thus performing the
same gas accounting as before.

This will lead to the switching from the existing Ethermint's evm `StateTransition` type to the
go-ethereum `vm.ApplyMessage` type, thus reducing code necessary perform a state transition.

```go
func (k *Keeper) TransitionDb(ctx sdk.Context, msg core.Message) (*types.ExecutionResult, error) {
  defer telemetry.ModuleMeasureSince(types.ModuleName, time.Now(), types.MetricKeyTransitionDB)

  initialGasMeter := ctx.GasMeter()

  // NOTE: Since CRUD operations on the SDK store consume gasm we need to set up an infinite gas meter so that we only consume
  // the gas used by the Ethereum message execution.
  // Not setting the infinite gas meter here would mean that we are incurring in additional gas costs
  k.ctx = ctx.WithGasMeter(sdk.NewInfiniteGasMeter())

  params := k.GetParams(ctx)
  cfg, found := k.GetChainConfig(ctx)
  if !found {
    // error
  }

  evm := k.NewEVM(msg, cfg.EthereumConfig(chainID))
  gasPool := &core.GasPool(ctx.BlockGasMeter().Limit()) // available gas left in the block for the tx execution

  // create an ethereum StateTransition instance and run TransitionDb
  result, err := core.ApplyMessage(evm, msg, gasPool)
  // return precheck errors (nonce, signature, balance and gas)
  // NOTE: these should be checked previously on the AnteHandler
  if err != nil {
    // log error
    return err
  }

  // The gas used on the state transition will 
  // be returned in the execution result so we need to deduct it from the transaction (?) GasMeter // TODO: double-check
  initialGasMeter.ConsumeGas(resp.UsedGas, "evm state transition")

  // set the gas meter to current_gas = initial_gas - used_gas
  k.ctx = k.ctx.WithGasMeter(initialGasMeter) 

  // return the VM Execution error (see go-ethereum/core/vm/errors.go)
  if result.Err != nil {
    // log error
    return result.Err
  }

  // return logs
  executionRes := &ExecutionResult{
    Response: &MsgEthereumTxResponse{
      Ret: result.ret,
    },
    GasInfo: GasInfo{
      GasConsumed: result.UsedGas,
      GasLimit:    gasPool,
    }
  
  return executionRes, nil
}
```

The EVM is created then as follows:

```go
func (k *Keeper) NewEVM(msg core.Message, config *params.ChainConfig) *vm.EVM {
  blockCtx := vm.BlockContext{
    CanTransfer: core.CanTransfer,
    Transfer:    core.Transfer,
    GetHash:     k.GetHashFn(),
    Coinbase:    common.Address{}, // there's no beneficiary since we're not mining
    BlockNumber: big.NewInt(k.ctx.BlockHeight()),
    Time:        big.NewInt(k.ctx.BlockHeader().Time.Unix()),
    Difficulty:  big.NewInt(0), // unused. Only required in PoW context
    GasLimit:    gasLimit,
  }

  txCtx := core.NewEVMTxContext(msg)
  vmConfig := k.VMConfig(st.Debug)

  return vm.NewEVM(blockCtx, txCtx, k, config, vmConfig)
}

func (k Keeper) VMConfig(debug bool) vm.Config{
  params := k.GetParams(ctx)

  eips := make([]int, len(params.ExtraEIPs))
  for i, eip := range params.ExtraEIPs {
    eips[i] = int(eip)
  }

  return vm.Config{
    ExtraEips:  eips,
    Tracer:     vm.NewJSONLogger(&vm.LogConfig{Debug: debug}, os.Stderr),
    Debug:      debug,
  }
}
```

## Consequences

<!-- > This section describes the resulting context, after applying the decision. All consequences should be listed here, not just the "positive" ones. A particular decision may have positive, negative, and neutral consequences, but all of them affect the team and project in the future. -->

### Backwards Compatibility

<!-- All ADRs that introduce backwards incompatibilities must include a section describing these incompatibilities and their severity. The ADR must explain how the author proposes to deal with these incompatibilities. ADR submissions without a sufficient backwards compatibility treatise may be rejected outright. -->

The proposed ADR is a breaking state machine change and will not have any backwards compatibility
since no chain that uses this code is in a production ready-state (at the moment of writing).

### Positive

- Improve maintenance by simplifying the state transition logic
- Defines a single option for accessing the store through the `Keeper`, thus removing the
  `CommitStateDB` type.
- State operations and tests are now all located in the `evm/keeper/` package
- Removes the concept of `stateObject` by commiting to the store directly
- Delete operations on `EndBlock` for updating and commiting dirty state objects.
- Split the state transition functionality (`NewEVM` from `TransitionDb`) allows to further
  modularize certain components that can be beneficial for customization (eg: using other EVMs other
  than Geth's)

### Negative

- Increases the dependency of external packages (eg: `go-ethereum`)
- Some state changes will have to be kept in store (eg: suicide state)

### Neutral

- Some of the fields from the `CommitStateDB` will have to be added to the `Keeper`

## Further Discussions

<!-- While an ADR is in the DRAFT or PROPOSED stage, this section should contain a summary of issues to be solved in future iterations (usually referencing comments from a pull-request discussion).
Later, this section can optionally list ideas or improvements the author or reviewers found during the analysis of this ADR. -->

## Test Cases [optional]

<!-- Test cases for an implementation are mandatory for ADRs that are affecting consensus changes. Other ADRs can choose to include links to test cases if applicable. -->

## References

<!-- - {reference link} -->
