# Snapshot and Revert in Ethermint

EVM uses state-reverting exceptions to handle errors. Such an exception will undo all changes made to the state in the current call (and all its sub-calls), and the caller could handle the error and don't propagate. We need to implement the `Snapshot` and `RevertToSnapshot` apis in `StateDB` interfaces to support this feature.

[go-ethereum implementation](https://github.com/ethereum/go-ethereum/blob/master/core/state/journal.go#L39) manages transient states in memory, and uses a list of journal logs to record all the state modification operations done so far, snapshot is an index in the log list, and to revert to a snapshot it just undo the journal logs after the snapshot index in reversed order.

Ethermint uses cosmos-sdk's storage api to manage states, fortunately the storage api supports creating cached overlays, it works like this:

```golang
// create a cached overlay storage on top of ctx storage.
overlayCtx, commit := ctx.CacheContext()
// Modify states using the overlayed storage
err := doCall(overlayCtx)
if err != nil {
  return err
}
// commit will write the dirty states into the underlying storage
commit()

// Now, just drop the overlayCtx and keep using ctx
```

And it can be used in a nested way, like this:

```golang
overlayCtx1, commit1 := ctx.CacheContext()
doCall1(overlayCtx1)
{
    overlayCtx2, commit2 := overlayCtx1.CacheContext()
    doCall2(overlayCtx2)
    commit2()
}
commit1()
```

With this feature, we can use a stake of overlayed contexts to implement nested `Snapshot` and `RevertToSnapshot` calls.

```golang
type cachedContext struct {
	ctx    sdk.Context
	commit func()
}
var contextStack []cachedContext
func Snapshot() int {
  ctx, commit := contextStack.Top().CacheContext()
  contextStack.Push(cachedContext{ctx, commit})
  return len(contextStack) - 1
}
func RevertToSnapshot(int snapshot) {
  contextStack = contextStack[:snapshot]
}
func Commit() {
  for i := len(contextStack) - 1; i >= 0; i-- {
    contextStack[i].commit()
  }
  contextStack = {}
}
```

