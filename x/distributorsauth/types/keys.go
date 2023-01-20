package types

const (
	// ModuleName defines the module name
	ModuleName = "distributorsauth"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey defines the module's message routing key
	RouterKey = ModuleName

	// MemStoreKey defines the in-memory store key
	MemStoreKey = "mem_distributorsauth"

	// QuerierRoute is the query router key for the distributorsauth module
	QuerierRoute = ModuleName
)

const (
	DistributorPrefix      = "distr-"
	DistributorAdminPrefix = "admin-"
)

func KeyPrefix(p string) []byte {
	return []byte(p)
}
