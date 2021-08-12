package config

// DefaultConfigTemplate defines the configuration template for the EVM RPC configuration
const DefaultConfigTemplate = `
###############################################################################
###                             EVM Configuration                           ###
###############################################################################

[json-rpc]

# tracer defines the tracer.
enable = {{ .EVMRPC.Enable }}

###############################################################################
###                           JSON RPC Configuration                        ###
###############################################################################

[json-rpc]

# Enable defines if the gRPC server should be enabled.
enable = {{ .JSONRPC.Enable }}

# Address defines the EVM RPC HTTP server address to bind to.
address = "{{ .JSONRPC.Address }}"

# Address defines the EVM WebSocket server address to bind to.
ws-address = "{{ .JSONRPC.WsAddress }}"

# API defines a list of JSON-RPC namespaces that should be enabled
# Example: "eth,txpool,personal,net,debug,web3"
api = "{{range $index, $elmt := .JSONRPC.API}}{{if $index}},{{$elmt}}{{else}}{{$elmt}}{{end}}{{end}}"

# EnableUnsafeCORS defines if CORS should be enabled (unsafe - use it at your own risk)
enable-unsafe-cors = "{{ .JSONRPC.EnableUnsafeCORS }}"
`
