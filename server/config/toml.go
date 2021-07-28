package config

// DefaultConfigTemplate defines the configuration template for the EVM RPC configuration
const DefaultConfigTemplate = `
###############################################################################
###                           EVM RPC Configuration                         ###
###############################################################################

[evm-rpc]

# Enable defines if the gRPC server should be enabled.
enable = {{ .EVMRPC.Enable }}

# Address defines the EVM RPC HTTP server address to bind to.
address = "{{ .EVMRPC.RPCAddress }}"

# Address defines the EVM WebSocket server address to bind to.
ws-address = "{{ .EVMRPC.WsAddress }}"

# API defines a list of JSON-RPC namespaces that should be enabled
api = "{{ .EVMRPC.API }}"

# EnableUnsafeCORS defines if CORS should be enabled (unsafe - use it at your own risk)
enable-unsafe-cors = "{{ .EVMRPC.EnableUnsafeCORS }}"
`
