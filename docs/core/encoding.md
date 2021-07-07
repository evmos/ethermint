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
client and state serialization. All the EVM module types that are used for state and clients
(transaction messages, genesis, query services, etc) will be implemented as protocol buffer messages.

### Amino

The Cosmos SDK also supports the legacy Amino encoding format for backwards compatibility with
previous versions, specially for client encoding and signing with Ledger devices. Ethermint does not
support Amino in the EVM module, but it is supported for all other Cosmos SDK modules that enable it.

### RLP

Recursive Length Prefix ([RLP](https://eth.wiki/en/fundamentals/rlp)), is an encoding/decoding algorithm that serializes a message and
allows for quick reconstruction of encoded data. Ethermint uses RLP to encode/decode Ethereum
messages for JSON-RPC handling to conform messages to the proper Ethereum format. This allows
messages to be encoded and decoded in the exact format as Ethereum's.

The `x/evm` transactions (`MsgEthereumTx`) encoding is performed by casting the message to a go-ethereum's `Transaction` and then marshaling the transaction data using RLP:

```go
// TxEncoder overwrites sdk.TxEncoder to support MsgEthereumTx
func (g txConfig) TxEncoder() sdk.TxEncoder {
  return func(tx sdk.Tx) ([]byte, error) {
    msg, ok := tx.(*evmtypes.MsgEthereumTx)
    if ok {
      return msg.AsTransaction().MarshalBinary()
   }
    return g.TxConfig.TxEncoder()(tx)
  }
}

// TxDecoder overwrites sdk.TxDecoder to support MsgEthereumTx
func (g txConfig) TxDecoder() sdk.TxDecoder {
  return func(txBytes []byte) (sdk.Tx, error) {
    tx := &ethtypes.Transaction{}

    err := tx.UnmarshalBinary(txBytes)
    if err == nil {
      msg := &evmtypes.MsgEthereumTx{}
      msg.FromEthereumTx(tx)
      return msg, nil
    }

    return g.TxConfig.TxDecoder()(txBytes)
  }
}
```

## Next {hide}

Learn how [pending state](./pending_state.md) is handled on Ethermint. {hide}
