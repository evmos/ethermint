<!--
order: 1
-->

# Encoding

The `codec` is used everywhere in the Cosmos SDK to encode and decode structs and interfaces. The specific codec used in the Cosmos SDK is called `go-amino`. {synopsis}

## Pre-requisite Readings

- [Cosmos SDK Encoding](https://docs.cosmos.network/master/core/encoding.html) {prereq}

## Encoding Formats

The Cosmos SDK utilizes two binary wire encoding protocols, [Amino](https://github.com/tendermint/go-amino/)
and [Protocol Buffers](https://developers.google.com/protocol-buffers), where Amino
is an object encoding specification. It is a subset of Proto3 with an extension for
interface support. See the [Proto3 spec](https://developers.google.com/protocol-buffers/docs/proto3)
for more information on Proto3, which Amino is largely compatible with (but not with Proto2).

Due to Amino having significant performance drawbacks, being reflection-based, and
not having any meaningful cross-language/client support, Protocol Buffers, specifically
[gogoprotobuf](https://github.com/gogo/protobuf/), is being used in place of Amino.
Note, this process of using Protocol Buffers over Amino is still an ongoing process.

Binary wire encoding of types in the Cosmos SDK can be broken down into two main
categories, client encoding and store encoding. Client encoding mainly revolves
around transaction processing and signing, whereas store encoding revolves around
types used in state-machine transitions and what is ultimately stored in the Merkle
tree.

For store encoding, protobuf definitions can exist for any type and will typically
have an Amino-based "intermediary" type. Specifically, the protobuf-based type
definition is used for serialization and persistence, whereas the Amino-based type
is used for business logic in the state-machine where they may converted back-n-forth.
Note, the Amino-based types may slowly be phased-out in the future so developers
should take note to use the protobuf message definitions where possible.

In the `codec` package, there exists two core interfaces, `Marshaler` and `ProtoMarshaler`,
where the former encapsulates the current Amino interface except it operates on
types implementing the latter instead of generic `interface{}` types.

In addition, there exists three implementations of `Marshaler`. The first being
`AminoCodec`, where both binary and JSON serialization is handled via Amino. The
second being `ProtoCodec`, where both binary and JSON serialization is handled
via Protobuf. Finally, `HybridCodec`, a codec that utilizes Protobuf for binary
serialization and Amino for JSON serialization. The `HybridCodec` is typically
the codec that used in majority in situations as it's easier to use for client
and state serialization.

This means that modules may use Amino or Protobuf encoding but the types must
implement `ProtoMarshaler`. If modules wish to avoid implementing this interface
for their types, they may use an Amino codec directly.

### Amino

Every module uses an Amino codec to serialize types and interfaces. This codec typically
has types and interfaces registered in that module's domain only (e.g. messages),
but there are exceptions like `x/gov`. Each module exposes a `RegisterCodec` function
that allows a user to provide a codec and have all the types registered. An application
will call this method for each necessary module.

### Protobuf

<!-- TODO: -->

## RLP

<!-- TODO: -->

## Next {hide}

Learn how to deploy a Solidity smart contract on Ethermint using [Truffle](./../guides/truffle.md) {hide}
