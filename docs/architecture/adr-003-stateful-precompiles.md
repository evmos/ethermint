# ADR 003: Stateful Precompiled Contracts

## Changelog

- 2022-06-16: first draft

## Status

DRAFT

## Abstract

Support stateful precompiled contracts to improve inter-operabilities between EVM smart contracts and native functionalities.

## Context

We need inter-operabilities to allow EVM smart contracts to access native functionalities, like manage native tokens through bank module, send/receive IBC messages through IBC modules.

The EVM hooks solution is not ideal, because the message processing is asynchronously, so the caller can't get the return value.

Precompiled contract can provide a much better interface for smart contract developers. But the default precompiled contract implementation in go-ethereum is stateless, we need to do some patches to support stateful ones properly.

## Decision

### Interface Changes

Change the `PrecompiledContract` interface like this (need to patch `go-ethereum`):

```
 type PrecompiledContract interface {
  	 RequiredGas(input []byte) uint64
-  	 Run(input []byte) ([]byte, error)
+  	 Run(input []byte, caller common.Address, value *big.Int, readonly bool) ([]byte, error)
 }
```

There are extra parameters passed to the precompiled contract:

- `caller`: aka. `msg.sender`.
- `value`: aka. `msg.value`.
- `readonly`: it's set to `true` for both `staticcall` and `delegatecall`, in these call types, the callee contract is not supposed to modify states. A stateful contract normally should just fail if it's `true`.

### Snapshot and Revert

To implement a stateful precompiled contract, one should be aware of the semantics of `StateDB` itself, basically it keeps all the state writes in memory and maintains a list of journal logs for the write operations, and it supports snapshot and revert by undoing the journal logs backward to a certain point in history. The dirty states are written into cosmos-sdk storage when commit at the end of the tx execution.

The precompiled contract must not write to cosmos-sdk storage directly, because the side effects can't be reverted by `StateDB` when exception happens. You should always maintain the dirty states in memory, and append a journal entry to `StateDB` for each modification which can undo it when called.

When reading from cosmos-sdk storage, you are actually reading the committed states, you need to read the in memory caches for the dirty states, for example the accounts and EVM contract storage are cached in `StateDB` itself, and different precompiled contracts may cache different native states.

It's also tricky if not impossible to let two precompiled contracts to write to the same piece of native states, because their in-memory states would be in conflicts.

### Example

TODO

## Consequences

> This section describes the resulting context, after applying the decision. All consequences should be listed here, not just the "positive" ones. A particular decision may have positive, negative, and neutral consequences, but all of them affect the team and project in the future.

### Backwards Compatibility

- State machine breaking
- Need to patch `go-ethereum`

### Positive

- Better interface for interaction between EVM contract and native functionalities.

### Negative

- Need to patch `go-ethereum`.

### Neutral

- Precompiled contract implementation need to be careful with the in memory dirty states.

## Further Discussions

## Test Cases [optional]

Test cases for an implementation are mandatory for ADRs that are affecting consensus changes. Other ADRs can choose to include links to test cases if applicable.

## References

- [ADR-001 EVM Hooks](adr-002-evm-hooks.md)
