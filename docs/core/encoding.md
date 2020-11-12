<!--
order: 1
-->

# Encoding

Learn about the encoding formats used on Ethermint. {synopsis}

## Pre-requisite Readings

- [Cosmos SDK Encoding](https://docs.cosmos.network/master/core/encoding.html) {prereq}
- [Ethereum RLP](https://eth.wiki/en/fundamentals/rlp) {prereq}

## Encoding Formats

### Protocol Buffers

The Cosmos [Stargate](https://stargate.cosmos.network/) release introduces
[protobuf](https://developers.google.com/protocol-buffers) as the main encoding format for both
client and state serialization. All the EVM module structs that are used for state and clients
(transaction messages, genesis, query services, etc) will be implemented as protocol buffer messages.

### Amino

The Cosmos SDK also supports the legacy Amino encoding format for backwards compatibility with
previous versions, specially for client encoding. Ethermint will not support Amino in the EVM module
once the migration to SDK `v0.40` is finalized.

### RLP

Recursive Length Prefix ([RLP](https://eth.wiki/en/fundamentals/rlp)), is an encoding/decoding algorithm that serializes a message and
allows for quick reconstruction of encoded data. Ethermint uses RLP to encode/decode Ethereum
messages for JSON-RPC handling to conform messages to the proper Ethereum format. This allows
messages to be encoded and decoded in the exact format as Ethereum's.

Each message type defined on the EVM module define the `EncodeRLP` and `DecodeRLP` methods which
implement the `rlp.Encoder` and `rlp.Decoder` interfaces respectively. The RLP encode method is used
to sign bytes and transactions in `RLPSignBytes` and `Sign`.

## Next {hide}

Learn how [pending state](./pending_state.md) is handled on Ethermint. {hide}
