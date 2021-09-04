<!--
order: 1
-->

# JSON-RPC Server

Learn about the JSON-RPC server to interact with the EVM. {synopsis}

## Pre-requisite Readings

- [EthWiki JSON-RPC API](https://eth.wiki/json-rpc/API) {prereq}
- [Geth JSON-RPC Server](https://geth.ethereum.org/docs/rpc/server) {prereq}

## JSON-RPC API

[JSON](https://json.org/) is a lightweight data-interchange format. It can represent numbers, strings, ordered sequences of values, and collections of name/value pairs.

[JSON-RPC](http://www.jsonrpc.org/specification) is a stateless, light-weight remote procedure call (RPC) protocol. Primarily this specification defines several data structures and the rules around their processing. It is transport agnostic in that the concepts can be used within the same process, over sockets, over HTTP, or in many various message passing environments. It uses JSON ([RFC 4627](https://www.ietf.org/rfc/rfc4627.txt)) as data format.

## JSON-RPC Support

Ethermint supports all standard web3 JSON-RPC APIs. You can find documentation for these APIs on the [`JSON-RPC Methods`](./endpoints.md) page.

JSON-RPC is provided on multiple transports. Ethermint supports JSON-RPC over HTTP and WebSocket. Transports must be enabled through command-line flags or through the `app.toml` configuration file. For more details see the []

Ethereum JSON-RPC APIs use a name-space system. RPC methods are grouped into several categories depending on their purpose. All method names are composed of the namespace, an underscore, and the actual method name within the namespace. For example, the eth_call method resides in the eth namespace.

Access to RPC methods can be enabled on a per-namespace basis. Find documentation for individual namespaces in the [Namespaces](./namespaces.md) page.

## HEX value encoding

At present there are two key datatypes that are passed over JSON: unformatted byte arrays and quantities. Both are passed with a hex encoding, however with different requirements to formatting:

When encoding **QUANTITIES** (integers, numbers): encode as hex, prefix with `"0x"`, the most compact representation (slight exception: zero should be represented as `"0x0"`). Examples:

- `0x41` (65 in decimal)
- `0x400` (1024 in decimal)
- WRONG: `0x` (should always have at least one digit - zero is `"0x0"`)
- WRONG: `0x0400` (no leading zeroes allowed)
- WRONG: `ff` (must be prefixed `0x`)

When encoding **UNFORMATTED DATA** (byte arrays, account addresses, hashes, bytecode arrays): encode as hex, prefix with `"0x"`, two hex digits per byte. Examples:

- `0x41` (size 1, `"A"`)
- `0x004200` (size 3, `"\0B\0"`)
- `0x` (size 0, `""`)
- WRONG: `0xf0f0f` (must be even number of digits)
- WRONG: `004200` (must be prefixed `0x`)

## Default block parameter

The following methods have an extra default block parameter:

- [`eth_getBalance`](./endpoints.md#eth-getbalance)
- [`eth_getCode`](./endpoints.md#eth-getcode)
- [`eth_getTransactionCount`](./endpoints.md#eth-gettransactioncount)
- [`eth_getStorageAt`](./endpoints.md#eth-getstorageat)
- [`eth_call`](./endpoints.md#eth-call)

When requests are made that act on the state of Ethermint, the last default block parameter determines the height of the block.

The following options are possible for the `defaultBlock` parameter:

- `HEX String` - an integer block number
- `String "earliest"` for the earliest/genesis block
- `String "latest"` - for the latest mined block
- `String "pending"` - for the pending state/transactions

## Curl Examples Explained

The curl options below might return a response where the node complains about the content type, this is because the `--data` option sets the content type to `application/x-www-form-urlencoded`. If your node does complain, manually set the header by placing `-H "Content-Type: application/json"` at the start of the call.

The examples also do not include the URL/IP & port combination which must be the last argument given to curl e.x. `127.0.0.1:8545`
