# ADR 001: State

## Changelog

- 2021-05-15: first draft

## Status

DRAFT, Not Implemented

> Please have a look at the [PROCESS](./PROCESS.md#adr-status) page.
> Use DRAFT if the ADR is in a draft stage (draft PR) or PROPOSED if it's in review.

## Abstract

> "If you can't explain it simply, you don't understand it well enough." Provide a simplified and layman-accessible explanation of the ADR.
> A short (~200 word) description of the issue being addressed.

## Context

> This section describes the forces at play, including technological, political, social, and project local. These forces are probably in tension, and should be called out as such. The language in this section is value-neutral. It is simply describing facts. It should clearly explain the problem and motivation that the proposal aims to resolve.

- EVM state with an instance of the go-ethereum `vm.StateDB` interface
- This StateDB defines the database interface for the state querying
  - Account, Balance, Code, State
  - Access Lists
  - Snapshot
  - Logs and preimages
- Some of these operations are executed in SDK-based chains by the auth module's `AccountKeeper` and bank module's `Keeper`.
- Current EVM module state design
  - EVM Keeper which contains an field for a CommitStateDB instance
    - This instance receives:
      - module store key on intitialization
      - account and bank keeper interfaces
      - parameter space to access relevant module params (eg: evm denomination
    - The `CommitStateDB` needs to receive the `sdk.Context` in every call in order to access the store
  - `CommitStateDB` is then passed to the new EVM instance that is created on every state transiton (i.e `StateTransition.TransitionDB`)
  - Maintenance of state transition logic
    - Custom gas accounting is not properly documented
    - Simulate logic (state is not commited due to finalizations at EndBlock)
  
### Dirty objects

The `CommitStateDB` holds references of `stateObjects`, live ethereum consensus accounts (i.e any
balance, account nonces or storage) which will get modified while processing a state transition.

Upon a state transition, these objects will be modified and marked as 'dirty'. Then, at every
`EndBlock`, the state of these modified objects will be commited to the store. While this model
works for Ethereum, asumming the state can only be modified through EVM transactions, a chain that
uses the EVM module can have their account balances updated through operations from other modules.
This means that an EVM state object can be modified through an EVM transaction (`evm.MsgEthereumTx`)
and other transactions like `bank.MsgSend` or `ibctransfer.MsgTransfer`. This can lead to unexpected
behaviors like state overwrites, due to the current behaviour that caches the dirty state on the EVM
instead of commiting any changes directly.

## Decision

<!-- > This section describes our response to these forces. It is stated in full sentences, with active voice. "We will ..." -->

```go
// Keeper now fully implements the StateDB interface
var _ vm.StateDB = (*Keeper)(nil)

type Keeper struct {
  // store key and encoding codec
  // ...

  // cache fields and context (reset every block)
  ctx sdk.Context

  // other CommitStateDB fields (journal, etc)
}
```


```go
func (k *Keeper) BeginBlock(ctx sdk.Context, req abci.RequestBeginBlock) {
  // reset cache values and context
  k.ResetCacheFields(ctx)
}
```

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
```

```go
func (k Keeper) VMConfig(debug bool) vm.Config{
    eips := make([]int, len(extraEIPs))
    for i, eip := range extraEIPs {
      eips[i] = int(eip)
    }

  return vm.Config{
    ExtraEips:  eips,
    Tracer:     vm.NewJSONLogger(&vm.LogConfig{Debug: debug}, os.Stderr),
    Debug:      debug,
  }
}
```

```go
func (k *Keeper) TransitionDb(ctx sdk.Context, msg core.Message) (*types.ExecutionResult, error) {
  defer telemetry.ModuleMeasureSince(types.ModuleName, time.Now(), types.MetricKeyTransitionDB)

  initialGasMeter := ctx.GasMeter()

  // NOTE: Since CRUD operations on the SDK store consume gasm we need to set up an infinite gas meter so that we only consume
  // the gas used by the Ethereum message execution.
  // Not setting the infinite gas meter here would mean that we are incurring in additional gas costs
  k.ctx = ctx.WithGasMeter(sdk.NewInfiniteGasMeter())
  evm := k.NewEVM(msg, config)
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

## Consequences

<!-- > This section describes the resulting context, after applying the decision. All consequences should be listed here, not just the "positive" ones. A particular decision may have positive, negative, and neutral consequences, but all of them affect the team and project in the future. -->

### Backwards Compatibility

<!-- All ADRs that introduce backwards incompatibilities must include a section describing these incompatibilities and their severity. The ADR must explain how the author proposes to deal with these incompatibilities. ADR submissions without a sufficient backwards compatibility treatise may be rejected outright. -->

The proposed ADR is a breaking state machine change and will not have any backwards compatibility
since no chain that uses this code is in a production ready-state (at the moment of writing).

### Positive

- Decreases maintenance of state transition code
- Creates a single option for accessing the store through the `Keeper`, thus removing the `CommitStateDB` type
- State operations and tests are now all located in the `evm/keeper/` package
- Removes the concept of `stateObject` by commiting to the store directly
- Deletes extra operations on `EndBlock`

### Negative

- Increases the dependency of external packages (eg: `go-ethereum`)
- Some state changes will have to be kept in store (eg: suicide state)

### Neutral

## Further Discussions

<!-- While an ADR is in the DRAFT or PROPOSED stage, this section should contain a summary of issues to be solved in future iterations (usually referencing comments from a pull-request discussion).
Later, this section can optionally list ideas or improvements the author or reviewers found during the analysis of this ADR. -->

## Test Cases [optional]

<!-- Test cases for an implementation are mandatory for ADRs that are affecting consensus changes. Other ADRs can choose to include links to test cases if applicable. -->

## References

<!-- - {reference link} -->
