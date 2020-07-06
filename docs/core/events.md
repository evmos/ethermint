<!--
order: 2
-->

# Events

`Event`s are objects that contain information about the execution of the application. They are mainly used by service providers like block explorers and wallet to track the execution of various messages and index transactions. {synopsis}

## Pre-requisite Readings

- [Cosmos SDK Events](https://docs.cosmos.network/master/core/events.html) {prereq}
- [Ethereum's PubSub JSON-RPC API](https://geth.ethereum.org/docs/rpc/pubsub) {prereq}

## Subscribing to Events

### SDK and Tendermint Events

It is possible to subscribe to `Events` via Tendermint's [Websocket](https://tendermint.com/docs/app-dev/subscribing-to-events-via-websocket.html#subscribing-to-events-via-websocket).
This is done by calling the `subscribe` RPC method via Websocket:

```json
{
    "jsonrpc": "2.0",
    "method": "subscribe",
    "id": "0",
    "params": {
        "query": "tm.event='eventCategory' AND eventType.eventAttribute='attributeValue'"
    }
}
```

The main `eventCategory` you can subscribe to are:

- `NewBlock`: Contains `events` triggered during `BeginBlock` and `EndBlock`.
- `Tx`: Contains `events` triggered during `DeliverTx` (i.e. transaction processing).
- `ValidatorSetUpdates`: Contains validator set updates for the block.

These events are triggered from the `state` package after a block is committed. You can get the
full list of `event` categories [here](https://godoc.org/github.com/tendermint/tendermint/types#pkg-constants).

The `type` and `attribute` value of the `query` allow you to filter the specific `event` you are looking for. For example, a `MsgEthereumTx` transaction triggers an `event` of type `ethermint` and has `sender` and `recipient` as `attributes`. Subscribing to this `event` would be done like so:

```json
{
    "jsonrpc": "2.0",
    "method": "subscribe",
    "id": "0",
    "params": {
        "query": "tm.event='Tx' AND ethereum.recipient='hexAddress'"
    }
}
```

where `hexAddress` is an Ethereum hex address (eg: `0x1122334455667788990011223344556677889900`).

### Ethereum JSON-RPC Events

<!-- TODO: PublicFiltersAPI -->

## Websocket Connection

### Tendermint Websocket

To start a connection with the Tendermint websocket you need to define the address with the `--node` flag when initializing the REST server (default `tcp://localhost:26657`):

```bash
emintcli rest-server --laddr "tcp://localhost:8545" --node "tcp://localhost:8080" --unlock-key <my_key> --chain-id <chain_id>
```

Then, start a websocket subscription with [ws](https://github.com/hashrocket/ws)

```bash
# connect to tendermint websocet at port 8080 as defined above
ws ws://localhost:8080/websocket

# subscribe to new Tendermint block headers
> { "jsonrpc": "2.0", "method": "subscribe", "params": ["tm.event='NewBlockHeader'"], "id": 1 }
```

### Ethereum Websocket

Since Ethermint runs uses Tendermint Core as it's consensus Engine and it's built with the Cosmos SDK framework, it inherits the event format from them. However, in order to support the native Web3 compatibility for websockets of the [Ethereum's PubSubAPI](https://geth.ethereum.org/docs/rpc/pubsub), Ethermint needs to cast the Tendermint responses retreived into the Ethereum types.

You can start a connection with the Ethereum websocket using the `--websocket-port` flag when initializing the REST server (default `7545`):

```bash
emintcli rest-server --laddr "tcp://localhost:8545" --websocket-port 8546 --unlock-key <my_key> --chain-id <chain_id>
```

Then, start a websocket subscription with [ws](https://github.com/hashrocket/ws)

```bash
# connect to tendermint websocet at port 8546 as defined above
ws ws://localhost:8546/

# subscribe to new Ethereum-formatted block Headers
> {"id": 1, "method": "eth_subscribe", "params": ["newHeads", {}]}
< {"jsonrpc":"2.0","result":"0x44e010cb2c3161e9c02207ff172166ef","id":1}
```

## Next {hide}

Learn how to connect Ethermint to [Metamask](./../guides/metamask.md) {hide}
