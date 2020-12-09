<!--
order: 6
-->

# Events

The EVM module emits the Cosmos SDK [events](./../../../docs/quickstart/events.md#sdk-and-tendermint-events) after a state execution. It can be expected that the type `message`, with an
attribute key of `action` will represent the first event for each message being processed as emitted
by the Cosmos SDK's `Baseapp` (i.e the the basic application that implements Tendermint Core's ABCI
interface).

## MsgEthereumTx

| Type     | Attribute Key | Attribute Value |
|----------|---------------|-----------------|
| ethereum | `"amount"`    | `{amount}`      |
| ethereum | `"recipient"` | `{eth_address}` |
| message  | `"sender"`    | `{eth_address}` |
| message  | `"action"`    | `"ethereum"`    |
| message  | `"module"`    | `"evm"`         |
