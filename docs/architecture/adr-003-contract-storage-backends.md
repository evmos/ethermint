# ADR 003: Contract Storage Backends

## Changelog

- 2022-08-26: first draft

## Status

DRAFT

## Abstract

This ADR proposes a way to customize contract storage backends, to allow user to try more efficient solutions for contract storage.

## Context

In go-ethereum, contract storage is stored as a separate Merkle Trie, whose root hash is recorded in the account object. And the underlying Patricia Merkle Trees[^1] supports store dynamic number of tries at the same time.

In cosmos-sdk, the iavl tree roots are indexed with the block numbers, and user can't add arbitrary roots, and the state of the whole application is stored in a single tree, so in ethermint, the contract storage is also stored in this tree, with the slot keys prefixed with the contract address.

A simple benchmark[^2] shows that the result db size varies a lot between these two solutions, for example, commits 100 blocks with each block do 100 storage slot writes, the result db size difference is nearly 10 times.

|      | Number of nodes | Total size of key-value pairs |
| ---- | --------------- | ----------------------------- |
| IAVL | 280713          | 35939674                      |
| MPT  | 35981           | 4266564                       |

Another consequence of this is go-ethereum has constant complexity when destruct a contract, while ethermint need `O(N)` complexity to delete storage slots one by one.

## Decision

Add an interface for the ethereum style merkle trie implementations to store the contract storage:

```golang
type ContractStorage interface {
  // Open the storage trie for a contract
  OpenStorageTrie(addrHash common.Hash, root common.Hash) (ContractStorageTrie, error)
  // Delete the contract storage trie.
  DeleteContractTrie(addrHash common.Hash) error
  // Commit all the tries and returns the new root hash of the accounts trie.
  CommitTries() (common.Hash, error)
  // Flush the low level db to disk.
  Flush() error
}

type ContractStorageTrie interface {
  // TryGet returns the value for slot stored in the trie.
  // If a node was not found in the database, `common.Hash{}` is returned.
  TryGet(slot common.Hash) (common.Hash, error)

  // Commit writes dirty slots in batch, if the slot value is `common.Hash{}`, delete,
  // returns the new root hash if success.
  UpdateSlots(slots map[common.Hash]common.Hash) (common.Hash, error)
}
```

And implement go-ethereum or other high performance MPT implementations as backends.

### ABCI Events

- At the end of tx delivery, when commit statedb, calls the `UpdateSlots` to write the dirty slots to the storage trie.

- At the end blocker, calls the `CommitTries` and write the returned root hash to cosmos-sdk storage with a global key, so the root hash is committed into the app hash.
- At the ABCI commit event, calls the `Flush` to flush the db to disk. May need to patch cosmos-sdk to hook custom commit logic into the app.

### Snapshot

The custom db backend should implement the `ExtensionSnapshotter` interface to support state sync.

### Pruning

The custom db backend can support pruning to delete obsoleted storage tries, it could get hooked into cosmos-sdk's pruning procedure if there's a hook api for pruning.

## Consequences

### Positive

- Reduce app db size and may increase db access speed.
- Constant complexity to destruct a contract.

### Negative

- Need a storage migration when apply to an existing chain.

### Neutral

## References

[^1]: https://ethereum.org/en/developers/docs/data-structures-and-encoding/patricia-merkle-trie/
[^2]: https://github.com/yihuang/ethermint/tree/geth-mpt

