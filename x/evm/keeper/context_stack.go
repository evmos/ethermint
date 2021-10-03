package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// cachedContext is a pair of cache context and its corresponding commit method.
// They are obtained from the return value of `context.CacheContext()`.
type cachedContext struct {
	ctx    sdk.Context
	commit func()
}

// ContextStack manages the initial context and a stack of cached contexts,
// to support the `StateDB.Snapshot` and `StateDB.RevertToSnapshot` methods.
type ContextStack struct {
	// Context of the initial state before transaction execution.
	// It's the context used by `StateDB.CommitedState`.
	initialCtx     sdk.Context
	cachedContexts []cachedContext
}

// CurrentContext returns the top context of cached stack,
// if the stack is empty, returns the initial context.
func (cs *ContextStack) CurrentContext() sdk.Context {
	l := len(cs.cachedContexts)
	if l == 0 {
		return cs.initialCtx
	}
	return cs.cachedContexts[l-1].ctx
}

// Reset sets the initial context and clear the cache context stack.
func (cs *ContextStack) Reset(ctx sdk.Context) {
	cs.initialCtx = ctx
	if len(cs.cachedContexts) > 0 {
		cs.cachedContexts = []cachedContext{}
	}
}

// IsEmpty returns true if the cache context stack is empty.
func (cs *ContextStack) IsEmpty() bool {
	return len(cs.cachedContexts) == 0
}

// Commit commits all the cached contexts from top to bottom in order and clears the stack by setting an empty slice of cache contexts.
func (cs *ContextStack) Commit() {
	// commit in order from top to bottom
	for i := len(cs.cachedContexts) - 1; i >= 0; i-- {
		// keep all the cosmos events
		cs.initialCtx.EventManager().EmitEvents(cs.cachedContexts[i].ctx.EventManager().Events())
		if cs.cachedContexts[i].commit == nil {
			panic(fmt.Sprintf("commit function at index %d should not be nil", i))
		} else {
			cs.cachedContexts[i].commit()
		}
	}
	cs.cachedContexts = []cachedContext{}
}

// CommitToRevision commit the cache after the target revision,
// to improve efficiency of db operations.
func (cs *ContextStack) CommitToRevision(target int) {
	if target < 0 || target >= len(cs.cachedContexts) {
		panic(fmt.Errorf("snapshot index %d out of bound [%d..%d)", target, 0, len(cs.cachedContexts)))
	}

	targetCtx := cs.cachedContexts[target].ctx
	// commit in order from top to bottom
	for i := len(cs.cachedContexts) - 1; i > target; i-- {
		// keep all the cosmos events
		targetCtx.EventManager().EmitEvents(cs.cachedContexts[i].ctx.EventManager().Events())
		if cs.cachedContexts[i].commit == nil {
			panic(fmt.Sprintf("commit function at index %d should not be nil", i))
		} else {
			cs.cachedContexts[i].commit()
		}
	}
	cs.cachedContexts = cs.cachedContexts[0 : target+1]
}

// Snapshot pushes a new cached context to the stack,
// and returns the index of it.
func (cs *ContextStack) Snapshot() int {
	i := len(cs.cachedContexts)
	ctx, commit := cs.CurrentContext().CacheContext()
	cs.cachedContexts = append(cs.cachedContexts, cachedContext{ctx: ctx, commit: commit})
	return i
}

// RevertToSnapshot pops all the cached contexts after the target index (inclusive).
// the target should be snapshot index returned by `Snapshot`.
// This function panics if the index is out of bounds.
func (cs *ContextStack) RevertToSnapshot(target int) {
	if target < 0 || target >= len(cs.cachedContexts) {
		panic(fmt.Errorf("snapshot index %d out of bound [%d..%d)", target, 0, len(cs.cachedContexts)))
	}
	cs.cachedContexts = cs.cachedContexts[:target]
}

// RevertAll discards all the cache contexts.
func (cs *ContextStack) RevertAll() {
	if len(cs.cachedContexts) > 0 {
		cs.RevertToSnapshot(0)
	}
}
