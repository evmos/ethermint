package server

// Tendermint full-node start flags
const (
	flagWithTendermint = "with-tendermint"
	flagAddress        = "address"
	flagTransport      = "transport"
	flagTraceStore     = "trace-store"
	flagCPUProfile     = "cpu-profile"
)

// GRPC-related flags.
const (
	flagGRPCEnable  = "grpc.enable"
	flagGRPCAddress = "grpc.address"
)

// RPCAPI-related flags.
const (
	flagRPCAPI = "rpc-api"
)

// Ethereum-related flags.
const (
	flagJSONRPCEnable            = "json-rpc.enable"
	flagJSONRPCAddress           = "json-rpc.address"
	flagEthereumWebsocketEnable  = "ethereum-websocket.enable"
	flagEthereumWebsocketAddress = "ethereum-websocket.address"
)
