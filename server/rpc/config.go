package rpc

// Config contains configuration fields that determine the
// behavior of the RPC HTTP server.
type Config struct {
	// EnableRPC defines whether or not to enable the RPC server
	EnableRPC      bool
	// RPCAddr defines the IP address to listen on
	RPCAddr        string
	// RPCPort defines the port to listen on
	RPCPort        int
	// RPCCORSDomains defines list of domains to enable CORS headers for (used by browsers)
	RPCCORSDomains []string
	// RPCVhosts defines list of domains to listen on (useful if Tendermint is addressable via DNS)
	RPCVHosts      []string
}
