package cachemulti

import (
	"fmt"
	"io"

	"github.com/cosmos/cosmos-sdk/store/tracekv"
	"github.com/cosmos/cosmos-sdk/store/types"
	"github.com/evmos/ethermint/store/cachekv"
)

// storeNameCtxKey is the TraceContext metadata key that identifies
// the store which emitted a given trace.
const storeNameCtxKey = "store_name"

//----------------------------------------
// Store

// Store holds many branched stores.
// Implements MultiStore.
// NOTE: a Store (and MultiStores in general) should never expose the
// keys for the substores.
type Store struct {
	stores map[types.StoreKey]*cachekv.Store

	traceWriter  io.Writer
	traceContext types.TraceContext
}

var _ types.CacheMultiStore = Store{}

// NewFromKVStore creates a new Store object from a mapping of store keys to
// CacheWrapper objects and a KVStore as the database. Each CacheWrapper store
// is a branched store.
func NewFromKVStore(
	stores map[types.StoreKey]types.KVStore,
	traceWriter io.Writer, traceContext types.TraceContext,
) Store {
	cms := Store{
		stores:       make(map[types.StoreKey]*cachekv.Store, len(stores)),
		traceWriter:  traceWriter,
		traceContext: traceContext,
	}

	for key, store := range stores {
		if cms.TracingEnabled() {
			tctx := cms.traceContext.Clone().Merge(types.TraceContext{
				storeNameCtxKey: key.Name(),
			})

			store = tracekv.NewStore(store, cms.traceWriter, tctx)
		}
		cms.stores[key] = cachekv.NewStore(store)
	}

	return cms
}

// NewStore creates a new Store object from parent rootmulti store, it branch out inner store of the specified keys.
func NewStore(
	parent types.MultiStore, keys map[string]*types.KVStoreKey,
) Store {
	stores := make(map[types.StoreKey]types.KVStore, len(keys))
	for _, key := range keys {
		stores[key] = parent.GetKVStore(key)
	}
	return NewFromKVStore(stores, nil, nil)
}

func newCacheMultiStoreFromCMS(cms Store) Store {
	stores := make(map[types.StoreKey]types.KVStore)
	for k, v := range cms.stores {
		stores[k] = v
	}

	return NewFromKVStore(stores, cms.traceWriter, cms.traceContext)
}

// SetTracer sets the tracer for the MultiStore that the underlying
// stores will utilize to trace operations. A MultiStore is returned.
func (cms Store) SetTracer(w io.Writer) types.MultiStore {
	cms.traceWriter = w
	return cms
}

// SetTracingContext updates the tracing context for the MultiStore by merging
// the given context with the existing context by key. Any existing keys will
// be overwritten. It is implied that the caller should update the context when
// necessary between tracing operations. It returns a modified MultiStore.
func (cms Store) SetTracingContext(tc types.TraceContext) types.MultiStore {
	if cms.traceContext != nil {
		for k, v := range tc {
			cms.traceContext[k] = v
		}
	} else {
		cms.traceContext = tc
	}

	return cms
}

// TracingEnabled returns if tracing is enabled for the MultiStore.
func (cms Store) TracingEnabled() bool {
	return cms.traceWriter != nil
}

// LatestVersion returns the branch version of the store
func (cms Store) LatestVersion() int64 {
	panic("cannot get latest version from branch cached multi-store")
}

// GetStoreType returns the type of the store.
func (cms Store) GetStoreType() types.StoreType {
	return types.StoreTypeMulti
}

// Write calls Write on each underlying store.
func (cms Store) Write() {
	for _, store := range cms.stores {
		store.Write()
	}
}

// Clone creates a snapshot of each store of the cache-multistore.
// Each copy is a copy-on-write operation and therefore is very fast.
func (cms Store) Clone() types.CacheMultiStore {
	stores := make(map[types.StoreKey]*cachekv.Store, len(cms.stores))
	for key, store := range cms.stores {
		stores[key] = store.Clone()
	}
	return Store{
		stores: stores,

		traceWriter:  cms.traceWriter,
		traceContext: cms.traceContext,
	}
}

// Restore restores the cache-multistore cache to a given snapshot.
func (cms Store) Restore(s types.CacheMultiStore) {
	ms := s.(Store)
	for key, store := range cms.stores {
		otherStore, ok := ms.stores[key]
		if !ok {
			panic("Invariant violation: Restore should only be called on a store cloned from itself")
		}
		store.Restore(otherStore)
	}
}

// Implements CacheWrapper.
func (cms Store) CacheWrap() types.CacheWrap {
	return cms.CacheMultiStore().(types.CacheWrap)
}

// CacheWrapWithTrace implements the CacheWrapper interface.
func (cms Store) CacheWrapWithTrace(_ io.Writer, _ types.TraceContext) types.CacheWrap {
	return cms.CacheWrap()
}

// Implements MultiStore.
func (cms Store) CacheMultiStore() types.CacheMultiStore {
	return newCacheMultiStoreFromCMS(cms)
}

// CacheMultiStoreWithVersion implements the MultiStore interface. It will panic
// as an already cached multi-store cannot load previous versions.
//
// TODO: The store implementation can possibly be modified to support this as it
// seems safe to load previous versions (heights).
func (cms Store) CacheMultiStoreWithVersion(_ int64) (types.CacheMultiStore, error) {
	panic("cannot branch cached multi-store with a version")
}

// GetStore returns an underlying Store by key.
func (cms Store) GetStore(key types.StoreKey) types.Store {
	s := cms.stores[key]
	if key == nil || s == nil {
		panic(fmt.Sprintf("kv store with key %v has not been registered in stores", key))
	}
	return types.Store(s)
}

// GetKVStore returns an underlying KVStore by key.
func (cms Store) GetKVStore(key types.StoreKey) types.KVStore {
	store := cms.stores[key]
	if key == nil || store == nil {
		panic(fmt.Sprintf("kv store with key %v has not been registered in stores", key))
	}
	return types.KVStore(store)
}
