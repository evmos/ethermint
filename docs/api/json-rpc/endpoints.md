<!--
order: 3
-->

# JSON-RPC Methods

Check the JSON-RPC methods supported on Ethermint. {synopsis}

## Pre-requisite Readings

- [Ethereum JSON-RPC](https://eth.wiki/json-rpc/API) {prereq}
- [Geth JSON-RPC APIs](https://geth.ethereum.org/docs/rpc/server) {prereq}

## Endpoints

| Method                                                                            | Namespace | Implemented | Notes                     |
|-----------------------------------------------------------------------------------|-----------|-------------|---------------------------|
| [`web3_clientVersion`](#web3-clientversion)                                       | Web3      | ✔           |                           |
| [`web3_sha3`](#web3-sha3)                                                         | Web3      | ✔           |                           |
| [`net_version`](#net-version)                                                     | Net       | ✔           |                           |
| [`net_peerCount`](#net-peerCount)                                                 | Net       | ✔           |                           |
| [`net_listening`](#net-listening)                                                 | Net       | ✔           |                           |
| [`eth_protocolVersion`](#eth-protocolversion)                                     | Eth       | ✔           |                           |
| [`eth_syncing`](#eth-syncing)                                                     | Eth       | ✔           |                           |
| [`eth_gasPrice`](#eth-gasprice)                                                   | Eth       | ✔           |                           |
| [`eth_accounts`](#eth-accounts)                                                   | Eth       | ✔           |                           |
| [`eth_blockNumber`](#eth-blocknumber)                                             | Eth       | ✔           |                           |
| [`eth_getBalance`](#eth-getbalance)                                               | Eth       | ✔           |                           |
| [`eth_getStorageAt`](#eth-getstorageat)                                           | Eth       | ✔           |                           |
| [`eth_getTransactionCount`](#eth-gettransactioncount)                             | Eth       | ✔           |                           |
| [`eth_getBlockTransactionCountByNumber`](#eth-getblocktransactioncountbynumber)   | Eth       | ✔           |                           |
| [`eth_getBlockTransactionCountByHash`](#eth-getblocktransactioncountbyhash)       | Eth       | ✔           |                           |
| [`eth_getCode`](#eth-getcode)                                                     | Eth       | ✔           |                           |
| [`eth_sign`](#eth-sign)                                                           | Eth       | ✔           |                           |
| [`eth_sendTransaction`](#eth-sendtransaction)                                     | Eth       | ✔           |                           |
| [`eth_sendRawTransaction`](#eth-sendrawtransaction)                               | Eth       | ✔           |                           |
| [`eth_call`](#eth-call)                                                           | Eth       | ✔           |                           |
| [`eth_estimateGas`](#eth-estimategas)                                             | Eth       | ✔           |                           |
| [`eth_getBlockByNumber`](#eth-getblockbynumber)                                   | Eth       | ✔           |                           |
| [`eth_getBlockByHash`](#eth-getblockbyhash)                                       | Eth       | ✔           |                           |
| [`eth_getTransactionByHash`](#eth-gettransactionbyhash)                           | Eth       | ✔           |                           |
| [`eth_getTransactionByBlockHashAndIndex`](#eth-gettransactionbyblockhashandindex) | Eth       | ✔           |                           |
| [`eth_getTransactionReceipt`](#eth-gettransactionreceipt)                         | Eth       | ✔           |                           |
| [`eth_newFilter`](#eth-newfilter)                                                 | Eth       | ✔           |                           |
| [`eth_newBlockFilter`](#eth-newblockfilter)                                       | Eth       | ✔           |                           |
| [`eth_newPendingTransactionFilter`](#eth-newpendingtransactionfilter)             | Eth       | ✔           |                           |
| [`eth_uninstallFilter`](#eth-uninstallfilter)                                     | Eth       | ✔           |                           |
| [`eth_getFilterChanges`](#eth-getfilterchanges)                                   | Eth       | ✔           |                           |
| [`eth_getLogs`](#eth-getlogs)                                                     | Eth       | ✔           |                           |
| `eth_getTransactionbyBlockNumberAndIndex`                                         | Eth       |             |                           |
| `eth_getWork`                                                                     | Eth       |             |                           |
| `eth_submitWork`                                                                  | Eth       |             |                           |
| `eth_submitHashrate`                                                              | Eth       |             |                           |
| `eth_getCompilers`                                                                | Eth       |             |                           |
| `eth_compileLLL`                                                                  | Eth       |             |                           |
| `eth_compileSolidity`                                                             | Eth       |             |                           |
| `eth_compileSerpent`                                                              | Eth       |             |                           |
| `eth_signTransaction`                                                             | Eth       |             |                           |
| `eth_mining`                                                                      | Eth       | N/A         | Not relevant to Ethermint |
| [`eth_coinbase`](#eth-coinbase)                                                   | Eth       | ✔           |                           |
| `eth_hashrate`                                                                    | Eth       | N/A         | Not relevant to Ethermint |
| `eth_getUncleCountByBlockHash`                                                    | Eth       | N/A         | Not relevant to Ethermint |
| `eth_getUncleCountByBlockNumber`                                                  | Eth       | N/A         | Not relevant to Ethermint |
| `eth_getUncleByBlockHashAndIndex`                                                 | Eth       | N/A         | Not relevant to Ethermint |
| `eth_getUncleByBlockNumberAndIndex`                                               | Eth       | N/A         | Not relevant to Ethermint |
| [`eth_getProof`](#eth-getProof)                                                   | Eth       | ✔           |                           |
| [`eth_subscribe`](#eth-subscribe)                                                 | Websocket | ✔           |                           |
| [`eth_unsubscribe`](#eth-unsubscribe)                                             | Websocket | ✔           |                           |
| [`personal_importRawKey`](#personal-importrawkey)                                 | Personal  | ✔           |                           |
| [`personal_listAccounts`](#personal-listaccounts)                                 | Personal  | ✔           |                           |
| [`personal_lockAccount`](#personal-lockaccount)                                   | Personal  | ✔           |                           |
| [`personal_newAccount`](#personal-newaccount)                                     | Personal  | ✔           |                           |
| [`personal_unlockAccount`](#personal-unlockaccount)                               | Personal  | ✔           |                           |
| [`personal_sendTransaction`](#personal-sendtransaction)                           | Personal  | ✔           |                           |
| [`personal_sign`](#personal-sign)                                                 | Personal  | ✔           |                           |
| [`personal_ecRecover`](#personal-ecrecover)                                       | Personal  | ✔           |                           |
| `db_putString`                                                                    | DB        |             |                           |
| `db_getString`                                                                    | DB        |             |                           |
| `db_putHex`                                                                       | DB        |             |                           |
| `db_getHex`                                                                       | DB        |             |                           |
| `shh_post`                                                                        | SSH       |             |                           |
| `shh_version`                                                                     | SSH       |             |                           |
| `shh_newIdentity`                                                                 | SSH       |             |                           |
| `shh_hasIdentity`                                                                 | SSH       |             |                           |
| `shh_newGroup`                                                                    | SSH       |             |                           |
| `shh_addToGroup`                                                                  | SSH       |             |                           |
| `shh_newFilter`                                                                   | SSH       |             |                           |
| `shh_uninstallFilter`                                                             | SSH       |             |                           |
| `shh_getFilterChanges`                                                            | SSH       |             |                           |
| `shh_getMessages`                                                                 | SSH       |             |                           |
| `admin_addPeer`                                                                   | Admin     |             |                           |
| `admin_datadir`                                                                   | Admin     |             |                           |
| `admin_nodeInfo`                                                                  | Admin     |             |                           |
| `admin_peers`                                                                     | Admin     |             |                           |
| `admin_startRPC`                                                                  | Admin     |             |                           |
| `admin_startWS`                                                                   | Admin     |             |                           |
| `admin_stopRPC`                                                                   | Admin     |             |                           |
| `admin_stopWS`                                                                    | Admin     |             |                           |
| `clique_getSnapshot`                                                              | Clique    |             |                           |
| `clique_getSnapshotAtHash`                                                        | Clique    |             |                           |
| `clique_getSigners`                                                               | Clique    |             |                           |
| `clique_proposals`                                                                | Clique    |             |                           |
| `clique_propose`                                                                  | Clique    |             |                           |
| `clique_discard`                                                                  | Clique    |             |                           |
| `clique_status`                                                                   | Clique    |             |                           |
| `debug_backtraceAt`                                                               | Debug     |             |                           |
| `debug_blockProfile`                                                              | Debug     | ✔           |                           |
| `debug_cpuProfile`                                                                | Debug     | ✔           |                           |
| `debug_dumpBlock`                                                                 | Debug     |             |                           |
| `debug_gcStats`                                                                   | Debug     | ✔           |                           |
| `debug_getBlockRlp`                                                               | Debug     |             |                           |
| `debug_goTrace`                                                                   | Debug     | ✔           |                           |
| `debug_freeOSMemory`                                                              | Debug     | ✔           |                           |
| `debug_memStats`                                                                  | Debug     | ✔           |                           |
| `debug_mutexProfile`                                                              | Debug     | ✔           |                           |
| `debug_seedHash`                                                                  | Debug     |             |                           |
| `debug_setHead`                                                                   | Debug     |             |                           |
| `debug_setBlockProfileRate`                                                       | Debug     | ✔           |                           |
| `debug_setGCPercent`                                                              | Debug     | ✔           |                           |
| `debug_setMutexProfileFraction`                                                   | Debug     | ✔           |                           |
| `debug_stacks`                                                                    | Debug     | ✔           |                           |
| `debug_startCPUProfile`                                                           | Debug     | ✔           |                           |
| `debug_startGoTrace`                                                              | Debug     | ✔           |                           |
| `debug_stopCPUProfile`                                                            | Debug     | ✔           |                           |
| `debug_stopGoTrace`                                                               | Debug     | ✔           |                           |
| `debug_traceBlock`                                                                | Debug     |             |                           |
| `debug_traceBlockByNumber`                                                        | Debug     |             |                           |
| `debug_traceBlockByHash`                                                          | Debug     |             |                           |
| `debug_traceBlockFromFile`                                                        | Debug     |             |                           |
| `debug_standardTraceBlockToFile`                                                  | Debug     |             |                           |
| `debug_standardTraceBadBlockToFile`                                               | Debug     |             |                           |
| `debug_traceTransaction`                                                          | Debug     |             |                           |
| `debug_verbosity`                                                                 | Debug     |             |                           |
| `debug_vmodule`                                                                   | Debug     |             |                           |
| `debug_writeBlockProfile`                                                         | Debug     | ✔           |                           |
| `debug_writeMemProfile`                                                           | Debug     | ✔           |                           |
| `debug_writeMutexProfile`                                                         | Debug     | ✔           |                           |
| `les_serverInfo`                                                                  | Les       |             |                           |
| `les_clientInfo`                                                                  | Les       |             |                           |
| `les_priorityClientInfo`                                                          | Les       |             |                           |
| `les_addBalance`                                                                  | Les       |             |                           |
| `les_setClientParams`                                                             | Les       |             |                           |
| `les_setDefaultParams`                                                            | Les       |             |                           |
| `les_latestCheckpoint`                                                            | Les       |             |                           |
| `les_getCheckpoint`                                                               | Les       |             |                           |
| `les_getCheckpointContractAddress`                                                | Les       |             |                           |
| `miner_getHashrate`                                                               | Miner     | N/A         | Not relevant to Ethermint |
| `miner_setExtra`                                                                  | Miner     | N/A         | Not relevant to Ethermint |
| [`miner_setGasPrice`](#miner-setgasprice)                                         | Miner     | ✔           |                           |
| `miner_start`                                                                     | Miner     | N/A         | Not relevant to Ethermint |
| `miner_stop`                                                                      | Miner     | N/A         | Not relevant to Ethermint |
| `miner_setGasLimit`                                                               | Miner     | N/A         | Not relevant to Ethermint |
| [`miner_setEtherbase`](#miner-setetherbase)                                       | Miner     | ✔           |                           |
| `txpool_content`                                                                  | TXPool    | ✔           |                           |
| `txpool_inspect`                                                                  | TXPool    | ✔           |                           |
| `txpool_status`                                                                   | TXPool    | ✔           |                           |


:::tip
Block Number can be entered as a Hex string, `"earliest"`, `"latest"` or `"pending"`.
:::

Below is a list of the RPC methods, the parameters and an example response from the namespaces.

## Web3 Methods

### `web3_clientVersion`

Get the web3 client version.

```json
// Request
curl -X POST --data '{"jsonrpc":"2.0","method":"web3_clientVersion","params":[],"id":67}' -H "Content-Type: application/json" http://localhost:8545

// Result
 {"jsonrpc":"2.0","id":1,"result":"Ethermint/0.0.0+/linux/go1.14"}
```

### `web3_sha3`

Returns Keccak-256 (not the standardized SHA3-256) of the given data.

#### Parameters

- The data to convert into a SHA3 hash

```json
// Request
curl -X POST --data '{"jsonrpc":"2.0","method":"web3_sha3","params":["0x67656c6c6f20776f726c64"],"id":1}' -H "Content-Type: application/json" http://localhost:8545

// Result
{"jsonrpc":"2.0","id":1,"result":"0x1b84adea42d5b7d192fd8a61a85b25abe0757e9a65cab1da470258914053823f"}
```

## Net Methods

### `net_version`

Returns the current network id.

```json
// Request
curl -X POST --data '{"jsonrpc":"2.0","method":"net_version","params":[],"id":1}' -H "Content-Type: application/json" http://localhost:8545

// Result
{"jsonrpc":"2.0","id":1,"result":"8"}
```

### `net_peerCount`

Returns the number of peers currently connected to the client.

```json
// Request
curl -X POST --data '{"jsonrpc":"2.0","method":"net_peerCount","params":[],"id":1}' -H "Content-Type: application/json" http://localhost:8545

// Result
{"jsonrpc":"2.0","id":1,"result":23}
```

### `net_listening`

Returns if client is actively listening for network connections.

```json
// Request
curl -X POST --data '{"jsonrpc":"2.0","method":"net_listening","params":[],"id":1}' -H "Content-Type: application/json" http://localhost:8545

// Result
{"jsonrpc":"2.0","id":1,"result":true}
```

## Eth Methods

### `eth_protocolVersion`

Returns the current ethereum protocol version.

```json
// Request
curl -X POST --data '{"jsonrpc":"2.0","method":"eth_protocolVersion","params":[],"id":1}' -H "Content-Type: application/json" http://localhost:8545

// Result
{"jsonrpc":"2.0","id":1,"result":"0x3f"}
```

### `eth_syncing`

The sync status object may need to be different depending on the details of Tendermint's sync protocol. However, the 'synced' result is simply a boolean, and can easily be derived from Tendermint's internal sync state.

```json
// Request
curl -X POST --data '{"jsonrpc":"2.0","method":"eth_syncing","params":[],"id":1}' -H "Content-Type: application/json" http://localhost:8545

// Result
{"jsonrpc":"2.0","id":1,"result":false}
```

### `eth_gasPrice`

Returns the current gas price in aphotons.

```json
// Request
curl -X POST --data '{"jsonrpc":"2.0","method":"eth_gasPrice","params":[],"id":1}' -H "Content-Type: application/json" http://localhost:8545

// Result
{"jsonrpc":"2.0","id":1,"result":"0x0"}
```

### `eth_accounts`

Returns array of all eth accounts.

```json
// Request
curl -X POST --data '{"jsonrpc":"2.0","method":"eth_accounts","params":[],"id":1}' -H "Content-Type: application/json" http://localhost:8545

// Result
{"jsonrpc":"2.0","id":1,"result":["0x3b7252d007059ffc82d16d022da3cbf9992d2f70","0xddd64b4712f7c8f1ace3c145c950339eddaf221d","0x0f54f47bf9b8e317b214ccd6a7c3e38b893cd7f0"]}
```

### `eth_blockNumber`

Returns the current block height.

```json
// Request
curl -X POST --data '{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}' -H "Content-Type: application/json" http://localhost:8545

// Result
{"jsonrpc":"2.0","id":1,"result":"0x66"}
```

### `eth_getBalance`

Returns the account balance for a given account address and Block Number.

#### Parameters

- Account Address

- Block Number or Block Hash [(eip-1898)](https://github.com/ethereum/EIPs/blob/master/EIPS/eip-1898.md)

```json
// Request
curl -X POST --data '{"jsonrpc":"2.0","method":"eth_getBalance","params":["0x0f54f47bf9b8e317b214ccd6a7c3e38b893cd7f0", "0x0"],"id":1}' -H "Content-Type: application/json" http://localhost:8545

// Result
{"jsonrpc":"2.0","id":1,"result":"0x36354d5575577c8000"}
```

### `eth_getStorageAt`

Returns the storage address for a given account address.

#### Parameters

- Account Address

- Integer of the position in the storage

- Block Number  or Block Hash [(eip-1898)](https://github.com/ethereum/EIPs/blob/master/EIPS/eip-1898.md)

```json
// Request
curl -X POST --data '{"jsonrpc":"2.0","method":"eth_getStorageAt","params":["0x0f54f47bf9b8e317b214ccd6a7c3e38b893cd7f0", "0", "latest"],"id":1}' -H "Content-Type: application/json" http://localhost:8545

// Result
{"jsonrpc":"2.0","id":1,"result":"0x0000000000000000000000000000000000000000000000000000000000000000"}
```

### `eth_getTransactionCount`

Returns the total transaction for a given account address and Block Number.

#### Parameters

- Account Address

- Block Number  or Block Hash [(eip-1898)](https://github.com/ethereum/EIPs/blob/master/EIPS/eip-1898.md)

```json
// Request
curl -X POST --data '{"jsonrpc":"2.0","method":"eth_getTransactionCount","params":["0x7bf7b17da59880d9bcca24915679668db75f9397", "0x0"],"id":1}' -H "Content-Type: application/json" http://localhost:8545

// Result
{"jsonrpc":"2.0","id":1,"result":"0x8"}
```

### `eth_getBlockTransactionCountByNumber`

Returns the total transaction count for a given block number.

#### Parameters

- Block number

```json
// Request
curl -X POST --data '{"jsonrpc":"2.0","method":"eth_getBlockTransactionCountByNumber","params":["0x1"],"id":1}' -H "Content-Type: application/json" http://localhost:8545

// Result
 {"jsonrpc":"2.0","id":1,"result":{"difficulty":null,"extraData":"0x0","gasLimit":"0xffffffff","gasUsed":"0x0","hash":"0x8101cc04aea3341a6d4b3ced715e3f38de1e72867d6c0db5f5247d1a42fbb085","logsBloom":"0x00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000","miner":"0x0000000000000000000000000000000000000000","nonce":null,"number":"0x17d","parentHash":"0x70445488069d2584fea7d18c829e179322e2b2185b25430850deced481ca2e77","sha3Uncles":null,"size":"0x1df","stateRoot":"0x269bb17fe7adb8dd5f15f57b717979f82078d6b7a675c1ba1b0da2d27e415fcc","timestamp":"0x5f5ba97c","totalDifficulty":null,"transactions":[],"transactionsRoot":"0x","uncles":[]}}
```

### `eth_getBlockTransactionCountByHash`

Returns the total transaction count for a given block hash.

#### Parameters

- Block Hash

```json
// Request
curl -X POST --data '{"jsonrpc":"2.0","method":"eth_getBlockTransactionCountByHash","params":["0x8101cc04aea3341a6d4b3ced715e3f38de1e72867d6c0db5f5247d1a42fbb085"],"id":1}' -H "Content-Type: application/json" http://localhost:8545

// Result
{"jsonrpc":"2.0","id":1,"result":"0x3"}
```

### `eth_getCode`

Returns the code for a given account address and Block Number.

#### Parameters

- Account Address

- Block Number  or Block Hash [(eip-1898)](https://github.com/ethereum/EIPs/blob/master/EIPS/eip-1898.md)

```json
// Request
curl -X POST --data '{"jsonrpc":"2.0","method":"eth_getCode","params":["0x7bf7b17da59880d9bcca24915679668db75f9397", "0x0"],"id":1}' -H "Content-Type: application/json" http://localhost:8545

// Result
{"jsonrpc":"2.0","id":1,"result":"0xef616c92f3cfc9e92dc270d6acff9cea213cecc7020a76ee4395af09bdceb4837a1ebdb5735e11e7d3adb6104e0c3ac55180b4ddf5e54d022cc5e8837f6a4f971b"}
```

### `eth_sign`

The sign method calculates an Ethereum specific signature with: sign(keccak256("\x19Ethereum Signed Message:\n" + len(message) + message))).

By adding a prefix to the message makes the calculated signature recognisable as an Ethereum specific signature. This prevents misuse where a malicious DApp can sign arbitrary data (e.g. transaction) and use the signature to impersonate the victim.

::: warning
the address to sign with must be unlocked.
:::

#### Parameters

- Account Address

- Message to sign

```json
// Request
curl -X POST --data '{"jsonrpc":"2.0","method":"eth_sign","params":["0x3b7252d007059ffc82d16d022da3cbf9992d2f70", "0xdeadbeaf"],"id":1}' -H "Content-Type: application/json" http://localhost:8545

// Result
{"jsonrpc":"2.0","id":1,"result":"0x909809c76ed2a5d38733de39207d0f411222b9b49c64a192bf649cb13f63f37b45acb4f6939facb4f1c277bc70fb00407564140c0f18600ac44388f2c1dfd1dc1b"}
```

### `eth_sendTransaction`

Sends transaction from given account to a given account.

#### Parameters

- Object containing:

    from: DATA, 20 Bytes - The address the transaction is send from.

    to: DATA, 20 Bytes - (optional when creating new contract) The address the transaction is directed to.

    gas: QUANTITY - (optional, default: 90000) Integer of the gas provided for the transaction execution. It will return unused gas.

    gasPrice: QUANTITY - (optional, default: To-Be-Determined) Integer of the gasPrice used for each paid gas

    value: QUANTITY - value sent with this transaction

    data: DATA - The compiled code of a contract OR the hash of the invoked method signature and encoded parameters. For details see Ethereum Contract ABI

    nonce: QUANTITY - (optional) Integer of a nonce. This allows to overwrite your own pending transactions that use the same nonce.

```json
// Request
curl -X POST --data '{"jsonrpc":"2.0","method":"eth_sendTransaction","params":[{"from":"0x3b7252d007059ffc82d16d022da3cbf9992d2f70", "to":"0x0f54f47bf9b8e317b214ccd6a7c3e38b893cd7f0", "value":"0x16345785d8a0000", "gasLimit":"0x5208", "gasPrice":"0x55ae82600"}],"id":1}'  -H "Content-Type: application/json" http://localhost:8545

// Result
{"jsonrpc":"2.0","id":1,"result":"0x33653249db68ebe5c7ae36d93c9b2abc10745c80a72f591e296f598e2d4709f6"}
```

### `eth_sendRawTransaction`

Creates new message call transaction or a contract creation for signed transactions.

You can get signed transaction data using the personal_sign method

#### Parameters

- The signed transaction data

```json
// Request
curl -X POST --data '{"jsonrpc":"2.0","method":"eth_sendRawTransaction","params":["0xf9ff74c86aefeb5f6019d77280bbb44fb695b4d45cfe97e6eed7acd62905f4a85034d5c68ed25a2e7a8eeb9baf1b8401e4f865d92ec48c1763bf649e354d900b1c"],"id":1}'  -H "Content-Type: application/json" http://localhost:8545

// Result
{"jsonrpc":"2.0","id":1,"result":"0x0000000000000000000000000000000000000000000000000000000000000000"}
```

### `eth_call`

Executes a new message call immediately without creating a transaction on the block chain.

#### Parameters

- Object containing:

    from: DATA, 20 Bytes - (optional) The address the transaction is sent from.

    to: DATA, 20 Bytes - The address the transaction is directed to.

    gas: QUANTITY - gas provided for the transaction execution. eth_call consumes zero gas, but this parameter may be needed by some executions.

    gasPrice: QUANTITY - gasPrice used for each paid gas

    value: QUANTITY - value sent with this transaction

    data: DATA - (optional) Hash of the method signature and encoded parameters. For details see Ethereum Contract ABI in the Solidity documentation

- Block number  or Block Hash [(eip-1898)](https://github.com/ethereum/EIPs/blob/master/EIPS/eip-1898.md)

```json
// Request
curl -X POST --data '{"jsonrpc":"2.0","method":"eth_call","params":[{"from":"0x3b7252d007059ffc82d16d022da3cbf9992d2f70", "to":"0xddd64b4712f7c8f1ace3c145c950339eddaf221d", "gas":"0x5208", "gasPrice":"0x55ae82600", "value":"0x16345785d8a0000", "data": "0xd46e8dd67c5d32be8d46e8dd67c5d32be8058bb8eb970870f072445675058bb8eb970870f072445675"}, "0x0"],"id":1}'  -H "Content-Type: application/json" http://localhost:8545

// Result
{"jsonrpc":"2.0","id":1,"result":"0x"}
```

### `eth_estimateGas`

Returns an estimate value of the gas required to send the transaction.

#### Parameters

- Object containing:

    from: DATA, 20 Bytes - The address the transaction is send from.

    to: DATA, 20 Bytes - (optional when creating new contract) The address the transaction is directed to.

    value: QUANTITY - value sent with this transaction

```json
// Request
curl -X POST --data '{"jsonrpc":"2.0","method":"eth_estimateGas","params":[{"from":"0x0f54f47bf9b8e317b214ccd6a7c3e38b893cd7f0", "to":"0x3b7252d007059ffc82d16d022da3cbf9992d2f70", "value":"0x16345785d8a00000"}],"id":1}'  -H "Content-Type: application/json" http://localhost:8545

// Result
{"jsonrpc":"2.0","id":1,"result":"0x1199b"}
```

### `eth_getBlockByNumber`

Returns information about a block by block number.

#### Parameters

- Block Number

- If true it returns the full transaction objects, if false only the hashes of the transactions.

```json
// Request
curl -X POST --data '{"jsonrpc":"2.0","method":"eth_getBlockByNumber","params":["0x1", false],"id":1}' -H "Content-Type: application/json" http://localhost:8545

// Result
{"jsonrpc":"2.0","id":1,"result":{"difficulty":null,"extraData":"0x0","gasLimit":"0xffffffff","gasUsed":null,"hash":"0xabac6416f737a0eb54f47495b60246d405d138a6a64946458cf6cbeae0d48465","logsBloom":"0x00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000","miner":"0x0000000000000000000000000000000000000000","nonce":null,"number":"0x1","parentHash":"0x","sha3Uncles":null,"size":"0x9b","stateRoot":"0x","timestamp":"0x5f5bd3e5","totalDifficulty":null,"transactions":[],"transactionsRoot":"0x","uncles":[]}}
```

### `eth_getBlockByHash`

Returns the block info given the hash found in the command above and a bool.

#### Parameters 

- Hash of a block.

- If true it returns the full transaction objects, if false only the hashes of the transactions.

```json
// Request
curl -X POST --data '{"jsonrpc":"2.0","method":"eth_getBlockByHash","params":["0x1b9911f57c13e5160d567ea6cf5b545413f96b95e43ec6e02787043351fb2cc4", false],"id":1}' -H "Content-Type: application/json" http://localhost:8545

// Result
{"jsonrpc":"2.0","id":1,"result":{"difficulty":null,"extraData":"0x0","gasLimit":"0xffffffff","gasUsed":null,"hash":"0x1b9911f57c13e5160d567ea6cf5b545413f96b95e43ec6e02787043351fb2cc4","logsBloom":"0x00000000100000000000000000000000000000000000000000000000000000000000020000000000000000000000000000000000000000000000000000000000000040000000000000000000000000200000000000000000000000000000000000000000000000000000000000000000100000000000000000000000000000000000000000000000000000000000000000000000000000000004000000000000002000000000000000000000000000000000000000004000000000000000000000000000000000000000000000000000000000000000000000010000000000000000000000000000000000000000000000000000000000000000000000000000","miner":"0x0000000000000000000000000000000000000000","nonce":null,"number":"0xc","parentHash":"0x404e58f31a9ede1b614b98701d6b0fbf1450f186842dbcf6426dd16811a5ca0d","sha3Uncles":null,"size":"0x307","stateRoot":"0x599ccdb111fc62c6398dc39be957df8e97bf8ab72ce6c06ff10641a92b754627","timestamp":"0x5f5fdbbd","totalDifficulty":null,"transactions":["0xae64961cb206a9773a6e5efeb337773a6fd0a2085ce480a174135a029afea615"],"transactionsRoot":"0x4764dba431128836fa919b83d314ba9cc000e75f38e1c31a60484409acea777b","uncles":[]}}
```

### `eth_getTransactionByHash`

Returns transaction details given the ethereum tx something.

#### Parameters

- hash of a transaction

```json
// Request
curl -X POST --data '{"jsonrpc":"2.0","method":"eth_getTransactionByHash","params":["0xec5fa15e1368d6ac314f9f64118c5794f076f63c02e66f97ea5fe1de761a8973"],"id":1}' -H "Content-Type: application/json" http://localhost:8545
 
// Result
{"jsonrpc":"2.0","id":1,"result":{"blockHash":"0x7a7398cc11d9c4c8e6f53e0c73824297aceafdab62db9e4b867a0da694384864","blockNumber":"0x188","from":"0x3b7252d007059ffc82d16d022da3cbf9992d2f70","gas":"0x147ee","gasPrice":"0x3b9aca00","hash":"0xec5fa15e1368d6ac314f9f64118c5794f076f63c02e66f97ea5fe1de761a8973","input":"0x6dba746c","nonce":"0x18","to":"0xa655256f589060437e5ffe2246dec385d040f148","transactionIndex":"0x0","value":"0x0","v":"0xa96","r":"0x6db399d694a452fb4106419140a6e5dbbe6817743a0f6f695a651e6576e59a5e","s":"0x25dd6ab1f936d0280d2fed0caeb0ebe5b9a46de6d8cb08ad8fd2c88deb55fc31"}}
```

### `eth_getTransactionByBlockHashAndIndex`

Returns transaction details given the block hash and the transaction index.

#### Parameters

- Hash of a block.

- Transaction index position.

```json
// Request
curl -X POST --data '{"jsonrpc":"2.0","method":"eth_getTransactionByBlockHashAndIndex","params":["0x1b9911f57c13e5160d567ea6cf5b545413f96b95e43ec6e02787043351fb2cc4", "0x0"],"id":1}' -H "Content-Type: application/json" http://localhost:8545

// Result
{"jsonrpc":"2.0","id":1,"result":{"blockHash":"0x1b9911f57c13e5160d567ea6cf5b545413f96b95e43ec6e02787043351fb2cc4","blockNumber":"0xc","from":"0xddd64b4712f7c8f1ace3c145c950339eddaf221d","gas":"0x4c4b40","gasPrice":"0x3b9aca00","hash":"0xae64961cb206a9773a6e5efeb337773a6fd0a2085ce480a174135a029afea615","input":"0x4f2be91f","nonce":"0x0","to":"0x439c697e0742a0ddb124a376efd62a72a94ac35a","transactionIndex":"0x0","value":"0x0","v":"0xa96","r":"0xced57d973e58b0f634f776d57daf41d3d3387ceb347a3a72ca0746e5ec2b709e","s":"0x384e89e209a5eb147a2bac3a4e399507400ac7b29cd155531f9d6203a89db3f2"}}
```

### `eth_getTransactionReceipt`

Returns the receipt of a transaction by transaction hash.

Note: Tx Code from Tendermint and the Ethereum receipt status are switched:
|         | Tendermint | Ethereum |
|---------|------------|----------|
| Success | 0          | 1        |
| Fail    | 1          | 0        |

#### Parameters

- Hash of a transaction

```json
// Request
curl -X POST --data '{"jsonrpc":"2.0","method":"eth_getTransactionReceipt","params":["0xae64961cb206a9773a6e5efeb337773a6fd0a2085ce480a174135a029afea614"],"id":1}' -H "Content-Type: application/json" http://localhost:8545

// Result
{"jsonrpc":"2.0","id":1,"result":{"blockHash":"0x1b9911f57c13e5160d567ea6cf5b545413f96b95e43ec6e02787043351fb2cc4","blockNumber":"0xc","contractAddress":"0x0000000000000000000000000000000000000000","cumulativeGasUsed":null,"from":"0xddd64b4712f7c8f1ace3c145c950339eddaf221d","gasUsed":"0x5289","logs":[{"address":"0x439c697e0742a0ddb124a376efd62a72a94ac35a","topics":["0x64a55044d1f2eddebe1b90e8e2853e8e96931cefadbfa0b2ceb34bee36061941"],"data":"0x0000000000000000000000000000000000000000000000000000000000000002","blockNumber":"0xc","transactionHash":"0xae64961cb206a9773a6e5efeb337773a6fd0a2085ce480a174135a029afea615","transactionIndex":"0x0","blockHash":"0x0000000000000000000000000000000000000000000000000000000000000000","logIndex":"0x0","removed":false},{"address":"0x439c697e0742a0ddb124a376efd62a72a94ac35a","topics":["0x938d2ee5be9cfb0f7270ee2eff90507e94b37625d9d2b3a61c97d30a4560b829"],"data":"0x0000000000000000000000000000000000000000000000000000000000000002","blockNumber":"0xc","transactionHash":"0xae64961cb206a9773a6e5efeb337773a6fd0a2085ce480a174135a029afea615","transactionIndex":"0x0","blockHash":"0x0000000000000000000000000000000000000000000000000000000000000000","logIndex":"0x1","removed":false}],"logsBloom":"0x00000000100000000000000000000000000000000000000000000000000000000000020000000000000000000000000000000000000000000000000000000000000040000000000000000000000000200000000000000000000000000000000000000000000000000000000000000000100000000000000000000000000000000000000000000000000000000000000000000000000000000004000000000000002000000000000000000000000000000000000000004000000000000000000000000000000000000000000000000000000000000000000000010000000000000000000000000000000000000000000000000000000000000000000000000000","status":"0x1","to":"0x439c697e0742a0ddb124a376efd62a72a94ac35a","transactionHash":"0xae64961cb206a9773a6e5efeb337773a6fd0a2085ce480a174135a029afea615","transactionIndex":"0x0"}}
```

### `eth_newFilter`

Create new filter using topics of some kind.

#### Parameters

- hash of a transaction

```json
// Request
curl -X POST --data '{"jsonrpc":"2.0","method":"eth_newFilter","params":[{"topics":["0x0000000000000000000000000000000000000000000000000000000012341234"]}],"id":1}' -H "Content-Type: application/json" http://localhost:8545

// Result
{"jsonrpc":"2.0","id":1,"result":"0xdc714a4a2e3c39dc0b0b84d66a3ccb00"}
```

### `eth_newBlockFilter`

Creates a filter in the node, to notify when a new block arrives.
 
```json
// Request
curl -X POST --data '{"jsonrpc":"2.0","method":"eth_newBlockFilter","params":[],"id":1}' -H "Content-Type: application/json" http://localhost:8545

// Result
{"jsonrpc":"2.0","id":1,"result":"0x3503de5f0c766c68f78a03a3b05036a5"}
```

### `eth_newPendingTransactionFilter`

Creates a filter in the node, to notify when new pending transactions arrive.

```json
// Request
curl -X POST --data '{"jsonrpc":"2.0","method":"eth_newPendingTransactionFilter","params":[],"id":1}' -H "Content-Type: application/json" http://localhost:8545

// Result
{"jsonrpc":"2.0","id":1,"result":"0x9daacfb5893d946997d3801ea18e9902"}
```

### `eth_uninstallFilter`

Removes the filter with the given filter id. Returns true if the filter was successfully uninstalled, otherwise false.

#### Parameters

- The filter id

```json
// Request
curl -X POST --data '{"jsonrpc":"2.0","method":"eth_uninstallFilter","params":["0xb91b6608b61bf56288a661a1bd5eb34a"],"id":1}' -H "Content-Type: application/json" http://localhost:8545

// Result
{"jsonrpc":"2.0","id":1,"result":true}
```

### `eth_getFilterChanges`

Polling method for a filter, which returns an array of logs which occurred since last poll.

#### Parameters

- The filter id

```json
// Request
curl -X POST --data '{"jsonrpc":"2.0","method":"eth_getFilterChanges","params":["0x127e9eca4f7751fb4e5cb5291ad8b455"],"id":1}' -H "Content-Type: application/json" http://localhost:8545

// Result
{"jsonrpc":"2.0","id":1,"result":["0xc6f08d183a81e149896fc5317c872f9092068e88e956ca1864e9bd4c81c09b44","0x3ca6dfb5be15549d721d1b3d10c1bec50ed6217c9ac7b61df361fac9692a27e5","0x776fffac134171acb1ebf2e59856625501ad5ccc5c4c8fe0359e0d4dff8919f2","0x84123103704dbd738c089276ab2b04b5936330b24f6e78453c4ba8bf4848aaf9","0xffddbe5bd8e8aa41e44002daa9ea89ade9e6980a0d83f51d104cf16498827eca","0x53430e49963e8ae32605d8f22dec2e757a691e6436d593854ca4d9383eeab86a","0x975948058c9351a91fbec332ca00dda39d1a919f5f16b996a4c7e30c38ba423b","0x619e37e32024c8efef7f7220e6caff4ee1d682ea78b2ac91e0a6b30850dc0677","0x31a5d985a40d08303ac68000ce008df512bcd1a911c497415c97f0624b4a271a","0x91dcf1fce4503a8dbb3e6fb61073f25cd31d69c766ecba639fefde4436e59d07","0x606d9e0143cfdb410a6812c590a8135b5c6b5c59eec26d760d5cd930aa47257d","0xd3c00b859b29b20ba654415eef648ef58251389c73a138580db87675b0d5465f","0x954391f0eb50888be90489898016ebb54f750f612f3adec2a00854955d5e52d8","0x698905f06aff921a9e9fcef39b8b0d107747c3e6204d2ea79cf4c12debf8d253","0x9fcafec5721938a06eb8e2951ede4b6ef8fae54a8c8f85f3166ec9782a0032b5","0xaec6d3364e47a5716ba69e4705f3c705d017f81298859589591183bfea87be7a","0x91bf2ee13319b6eaca96ed89c126437b66c4df1b13560c6a9bb18556ee3b7e1f","0x4f426dc1fc0ea8149052033065b237892d2d34927b2d558ab50c5a7fb98d6e79","0xdd809fb07e5aab638fef5311371b4e2b27c9c9a6183fde0cdd2b7724f6d2a89b","0x7e12fc92ab953e233a304959a2a8474d96195e71efd9388fdceb1326a577811a","0x30618ef6b490c3cc9979c47163459db37c1a1e0aa5793c56accd417f9d89973b","0x614609f06ee24bae7408e45895b1a25e6b19a8159aeea7a95c9d1339d9ba286f","0x115ddc6d533620040791d241f01f1c5ae3d9d1a8f64b15af5e9793e4d9096e22","0xb7458c9323beeca2cd54f32a6af5671f3cd5a7a251aed9d82bdd6ebe5f56305b","0x573dd48a5ba7bf4cc3d49597cd7419f75ecc9897258f1ebadebd670446d0d358","0xcb6670918439f9698413b53f3b5336d82ca4be152fdefaacf45e052fff6262fc","0xf3fe2a8945abafd269ab97bfdc80b3dbff2202ffdce59a227f952874b966b230","0x989980707007533cc0840a079f77f261a2e818abae1a1ffd3af02f3fff1d35fd","0x886b6ae365fec996be8a9a2c31cf4cda97ff8352908be2c83f17abd66ef1591e","0xfd90df68706ef95a62b317de93d6899a9bd6c80416e42d007f5c30fcdedfce24","0x7af8491fbb0373886d9032bb74e0ef52ed9e100f260b79bd15f46126b38cbede","0x91d1e2cd55533cf7dd5de86c9aa73295e811b1279be193d429bbd6ba83810e16","0x6b65b3128c2104005a04923288fe2aa33a2477a4962bef70532f94cab582f2a7"]}
```

<!-- 
### eth_getFilterLogs

Polling method for a filter, which returns an array of logs which occurred since last poll.

#### Parameters

- The filter id

```json
// Request
curl -X POST --data '{"jsonrpc":"2.0","method":"eth_getFilterLogs","params":["0x127e9eca4f7751fb4e5cb5291ad8b455"],"id":1}' -H "Content-Type: application/json" http://localhost:8545

// Result
{"jsonrpc":"2.0","id":1,"error":{"code":-32000,"message":"filter 0x35b64c227ce30e84fc5c7bd347be380e doesn't have a LogsSubscription type: got 5"}} 
``` -->

### `eth_getLogs`

Returns an array of all logs matching a given filter object.

#### Parameters

- Object containing:

    fromBlock: QUANTITY|TAG - (optional, default: "latest") Integer block number, or "latest" for the last mined block or "pending", "earliest" for not yet mined transactions.

    toBlock: QUANTITY|TAG - (optional, default: "latest") Integer block number, or "latest" for the last mined block or "pending", "earliest" for not yet mined transactions.

    address: DATA|Array, 20 Bytes - (optional) Contract address or a list of addresses from which logs should originate.

    topics: Array of DATA, - (optional) Array of 32 Bytes DATA topics. Topics are order-dependent. Each topic can also be an array of DATA with “or” options.

    blockhash: (optional, future) With the addition of EIP-234, blockHash will be a new filter option which restricts the logs returned to the single block with the 32-byte hash blockHash. Using blockHash is equivalent to fromBlock = toBlock = the block number with hash blockHash. If blockHash is present in in the filter criteria, then neither fromBlock nor toBlock are allowed.

```json
// Request
curl -X POST --data '{"jsonrpc":"2.0","method":"eth_getLogs","params":[{"topics":["0x775a94827b8fd9b519d36cd827093c664f93347070a554f65e4a6f56cd738898","0x0000000000000000000000000000000000000000000000000000000000000011"], "fromBlock":"latest"}],"id":1}' -H "Content-Type: application/json" http://localhost:8545

// Result
{"jsonrpc":"2.0","id":1,"result":[]}
```

## TxPool Methods

### `txpool_content`

Returns a list of the exact details of all the transactions currently pending for inclusion in the next block(s), as well as the ones that are being scheduled for future execution only.

```json
// Request
curl -X POST --data '{"jsonrpc":"2.0","method":"txpool_content","params":[],"id":1}' -H "Content-Type: application/json" http://localhost:8545

// Result
{"jsonrpc":"2.0","id":1,"result":{"pending":{},"queued":{}}
```

### `txpool_inspect`

Returns a list on text format to summarize all the transactions currently pending for inclusion in the next block(s), as well as the ones that are being scheduled for future execution only. This is a method specifically tailored to developers to quickly see the transactions in the pool and find any potential issues.

```json
// Request
curl -X POST --data '{"jsonrpc":"2.0","method":"txpool_inspect","params":[],"id":1}' -H "Content-Type: application/json" http://localhost:8545

// Result
{"jsonrpc":"2.0","id":1,"result":{"pending":{},"queued":{}}
```

### `txpool_status`

Returns the number of transactions currently pending for inclusion in the next block(s), as well as the ones that are being scheduled for future execution only.

```json
// Request
curl -X POST --data '{"jsonrpc":"2.0","method":"txpool_status","params":[],"id":1}' -H "Content-Type: application/json" http://localhost:8545

// Result
{"jsonrpc":"2.0","id":1,"result":{"pending":"0x0","queued":"0x0"}}
```

### eth_coinbase

Returns the account the mining rewards will be send to.

```json
// Request
curl -X POST --data '{"jsonrpc":"2.0","method":"eth_coinbase","params":[],"id":1}' -H "Content-Type: application/json" http://localhost:8545

// Result
{"jsonrpc":"2.0","id":1,"result":"0x7cB61D4117AE31a12E393a1Cfa3BaC666481D02E"}
```

### eth_getProof

Returns the account- and storage-values of the specified account including the Merkle-proof.

#### Parameters

- Address of account or contract

- Integer of the position in the storage

- Block Number  or Block Hash [(eip-1898)](https://github.com/ethereum/EIPs/blob/master/EIPS/eip-1898.md)

```json
// Request
curl -X POST --data '{"jsonrpc":"2.0","method":"eth_getProof","params":["0x1234567890123456789012345678901234567890",["0x0000000000000000000000000000000000000000000000000000000000000000","0x0000000000000000000000000000000000000000000000000000000000000001"],"latest"],"id":1}' -H "Content-type:application/json" http://localhost:8545

// Result
{"jsonrpc": "2.0", "id": 1, "result": {"address": "0x1234567890123456789012345678901234567890", "accountProof": ["0xf90211a090dcaf88c40c7bbc95a912cbdde67c175767b31173df9ee4b0d733bfdd511c43a0babe369f6b12092f49181ae04ca173fb68d1a5456f18d20fa32cba73954052bda0473ecf8a7e36a829e75039a3b055e51b8332cbf03324ab4af2066bbd6fbf0021a0bbda34753d7aa6c38e603f360244e8f59611921d9e1f128372fec0d586d4f9e0a04e44caecff45c9891f74f6a2156735886eedf6f1a733628ebc802ec79d844648a0a5f3f2f7542148c973977c8a1e154c4300fec92f755f7846f1b734d3ab1d90e7a0e823850f50bf72baae9d1733a36a444ab65d0a6faaba404f0583ce0ca4dad92da0f7a00cbe7d4b30b11faea3ae61b7f1f2b315b61d9f6bd68bfe587ad0eeceb721a07117ef9fc932f1a88e908eaead8565c19b5645dc9e5b1b6e841c5edbdfd71681a069eb2de283f32c11f859d7bcf93da23990d3e662935ed4d6b39ce3673ec84472a0203d26456312bbc4da5cd293b75b840fc5045e493d6f904d180823ec22bfed8ea09287b5c21f2254af4e64fca76acc5cd87399c7f1ede818db4326c98ce2dc2208a06fc2d754e304c48ce6a517753c62b1a9c1d5925b89707486d7fc08919e0a94eca07b1c54f15e299bd58bdfef9741538c7828b5d7d11a489f9c20d052b3471df475a051f9dd3739a927c89e357580a4c97b40234aa01ed3d5e0390dc982a7975880a0a089d613f26159af43616fd9455bb461f4869bfede26f2130835ed067a8b967bfb80", "0xf90211a0395d87a95873cd98c21cf1df9421af03f7247880a2554e20738eec2c7507a494a0bcf6546339a1e7e14eb8fb572a968d217d2a0d1f3bc4257b22ef5333e9e4433ca012ae12498af8b2752c99efce07f3feef8ec910493be749acd63822c3558e6671a0dbf51303afdc36fc0c2d68a9bb05dab4f4917e7531e4a37ab0a153472d1b86e2a0ae90b50f067d9a2244e3d975233c0a0558c39ee152969f6678790abf773a9621a01d65cd682cc1be7c5e38d8da5c942e0a73eeaef10f387340a40a106699d494c3a06163b53d956c55544390c13634ea9aa75309f4fd866f312586942daf0f60fb37a058a52c1e858b1382a8893eb9c1f111f266eb9e21e6137aff0dddea243a567000a037b4b100761e02de63ea5f1fcfcf43e81a372dafb4419d126342136d329b7a7ba032472415864b08f808ba4374092003c8d7c40a9f7f9fe9cc8291f62538e1cc14a074e238ff5ec96b810364515551344100138916594d6af966170ff326a092fab0a0d31ac4eef14a79845200a496662e92186ca8b55e29ed0f9f59dbc6b521b116fea090607784fe738458b63c1942bba7c0321ae77e18df4961b2bc66727ea996464ea078f757653c1b63f72aff3dcc3f2a2e4c8cb4a9d36d1117c742833c84e20de994a0f78407de07f4b4cb4f899dfb95eedeb4049aeb5fc1635d65cf2f2f4dfd25d1d7a0862037513ba9d45354dd3e36264aceb2b862ac79d2050f14c95657e43a51b85c80", "0xf90171a04ad705ea7bf04339fa36b124fa221379bd5a38ffe9a6112cb2d94be3a437b879a08e45b5f72e8149c01efcb71429841d6a8879d4bbe27335604a5bff8dfdf85dcea00313d9b2f7c03733d6549ea3b810e5262ed844ea12f70993d87d3e0f04e3979ea0b59e3cdd6750fa8b15164612a5cb6567cdfb386d4e0137fccee5f35ab55d0efda0fe6db56e42f2057a071c980a778d9a0b61038f269dd74a0e90155b3f40f14364a08538587f2378a0849f9608942cf481da4120c360f8391bbcc225d811823c6432a026eac94e755534e16f9552e73025d6d9c30d1d7682a4cb5bd7741ddabfd48c50a041557da9a74ca68da793e743e81e2029b2835e1cc16e9e25bd0c1e89d4ccad6980a041dda0a40a21ade3a20fcd1a4abb2a42b74e9a32b02424ff8db4ea708a5e0fb9a09aaf8326a51f613607a8685f57458329b41e938bb761131a5747e066b81a0a16808080a022e6cef138e16d2272ef58434ddf49260dc1de1f8ad6dfca3da5d2a92aaaadc58080", "0xf851808080a009833150c367df138f1538689984b8a84fc55692d3d41fe4d1e5720ff5483a6980808080808080808080a0a319c1c415b271afc0adcb664e67738d103ac168e0bc0b7bd2da7966165cb9518080"], "balance": "0x0", "codeHash": "0xc5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470", "nonce": "0x0", "storageHash": "0x56e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b421", "storageProof": [{"key": "0x0000000000000000000000000000000000000000000000000000000000000000", "value": "0x0", "proof": []}, {"key": "0x0000000000000000000000000000000000000000000000000000000000000001", "value": "0x0", "proof": []}]}}
```

## WebSocket Methods

Read about websockets in [events](./../quickstart/events.md) {hide}

### `eth_subscribe`

subscribe using JSON-RPC notifications. This allows clients to wait for events instead of polling for them.

It works by subscribing to particular events. The node will return a subscription id. For each event that matches the subscription a notification with relevant data is send together with the subscription id.

#### Parameters

- Subscription Name

- Optional Arguments

```json
// Request
{"id": 1, "method": "eth_subscribe", "params": ["newHeads", {"includeTransactions": true}]}

// Result
< {"jsonrpc":"2.0","result":"0x34da6f29e3e953af4d0c7c58658fd525","id":1}
```

### `eth_unsubscribe`

Unsubscribe from an event using the subscription id

#### Parameters

- Subscription ID

```json
// Request
{"id": 1, "method": "eth_unsubscribe", "params": ["0x34da6f29e3e953af4d0c7c58658fd525"]}

// Result
{"jsonrpc":"2.0","result":true,"id":1}
```

## Personal Methods

### `personal_importRawKey`

Imports the given unencrypted private key (hex string) into the key store, encrypting it with the passphrase.

Returns the address of the new account.

:::warning
Currently, this is not implemented since the feature is not supported by the keys
:::

#### Parameters

- Hex encoded ECDSA key

- Passphrase

```json
// Request
curl -X POST --data '{"jsonrpc":"2.0","method":"personal_importRawKey","params":["c5bd76cd0cd948de17a31261567d219576e992d9066fe1a6bca97496dec634e2c8e06f8949773b300b9f73fabbbc7710d5d6691e96bcf3c9145e15daf6fe07b9", "the key is this"],"id":1}' -H "Content-Type: application/json" http://localhost:8545

```

### `personal_listAccounts`

Returns a list of addresses for accounts this node manages.

```json
// Request
curl -X POST --data '{"jsonrpc":"2.0","method":"personal_listAccounts","params":[],"id":1}' -H "Content-Type: application/json" http://localhost:8545

// Result
{"jsonrpc":"2.0","id":1,"result":["0x3b7252d007059ffc82d16d022da3cbf9992d2f70","0xddd64b4712f7c8f1ace3c145c950339eddaf221d","0x0f54f47bf9b8e317b214ccd6a7c3e38b893cd7f0"]}
```

### `personal_lockAccount`

Removes the private key with given address from memory. The account can no longer be used to send transactions.

#### Parameters

- Account Address

```json
// Request
curl -X POST --data '{"jsonrpc":"2.0","method":"personal_lockAccount","params":["0x0f54f47bf9b8e317b214ccd6a7c3e38b893cd7f0"],"id":1}' -H "Content-Type: application/json" http://localhost:8545

// Result
{"jsonrpc":"2.0","id":1,"result":true}
```

### `personal_newAccount`

Generates a new private key and stores it in the key store directory. The key file is encrypted with the given passphrase. Returns the address of the new account.

#### Parameters

- Passphrase

```json
// Request
curl -X POST --data '{"jsonrpc":"2.0","method":"personal_newAccount","params":["This is the passphrase"],"id":1}' -H "Content-Type: application/json" http://localhost:8545

// Result
{"jsonrpc":"2.0","id":1,"result":"0xf0e4086ad1c6aab5d42161d5baaae2f9ad0571c0"}
```

### `personal_unlockAccount`

Decrypts the key with the given address from the key store.

Both passphrase and unlock duration are optional when using the JavaScript console. The unencrypted key will be held in memory until the unlock duration expires. If the unlock duration defaults to 300 seconds. An explicit duration of zero seconds unlocks the key until geth exits.

The account can be used with eth_sign and eth_sendTransaction while it is unlocked.

#### Parameters

- Account Address

- Passphrase

- Duration

```json
// Request
curl -X POST --data '{"jsonrpc":"2.0","method":"personal_unlockAccount","params":["0x0f54f47bf9b8e317b214ccd6a7c3e38b893cd7f0", "secret passphrase", 30],"id":1}' -H "Content-Type: application/json" http://localhost:8545

// Result
{"jsonrpc":"2.0","id":1,"result":true}
```

### `personal_sendTransaction`

Validate the given passphrase and submit transaction.

The transaction is the same argument as for eth_sendTransaction and contains the from address. If the passphrase can be used to decrypt the private key belonging to tx.from the transaction is verified, signed and send onto the network.

:::warning
The account is not unlocked globally in the node and cannot be used in other RPC calls.
:::

#### Parameters

- Object containing:

    from: DATA, 20 Bytes - The address the transaction is send from.

    to: DATA, 20 Bytes - (optional when creating new contract) The address the transaction is directed to.

    value: QUANTITY - value sent with this transaction

- Passphrase

```json
// Request
curl -X POST --data '{"jsonrpc":"2.0","method":"personal_sendTransaction","params":[{"from":"0x3b7252d007059ffc82d16d022da3cbf9992d2f70","to":"0xddd64b4712f7c8f1ace3c145c950339eddaf221d", "value":"0x16345785d8a0000"}, "passphrase"],"id":1}' -H "Content-Type: application/json" http://localhost:8545

// Result
{"jsonrpc":"2.0","id":1,"result":"0xd2a31ec1b89615c8d1f4d08fe4e4182efa4a9c0d5758ace6676f485ea60e154c"}
```

### `personal_sign`

The sign method calculates an Ethereum specific signature with: sign(keccack256("\x19Ethereum Signed Message:\n" + len(message) + message))).

#### Parameters

- Message

- Account Address

- Password

```json
// Request
curl -X POST --data '{"jsonrpc":"2.0","method":"personal_sign","params":["0xdeadbeaf", "0x3b7252d007059ffc82d16d022da3cbf9992d2f70", "password"],"id":1}' -H "Content-Type: application/json" http://localhost:8545

// Result
{"jsonrpc":"2.0","id":1,"result":"0xf9ff74c86aefeb5f6019d77280bbb44fb695b4d45cfe97e6eed7acd62905f4a85034d5c68ed25a2e7a8eeb9baf1b8401e4f865d92ec48c1763bf649e354d900b1c"}
```

### `personal_ecRecover`

ecRecover returns the address associated with the private key that was used to calculate the signature in personal_sign.

#### Parameters

- Message

- Signature returned from personal_sign

```json
// Request
curl -X POST --data '{"jsonrpc":"2.0","method":"personal_ecRecover","params":["0xdeadbeaf", "0xf9ff74c86aefeb5f6019d77280bbb44fb695b4d45cfe97e6eed7acd62905f4a85034d5c68ed25a2e7a8eeb9baf1b8401e4f865d92ec48c1763bf649e354d900b1c"],"id":1}' -H "Content-Type: application/json" http://localhost:8545

// Result
{"jsonrpc":"2.0","id":1,"result":"0x3b7252d007059ffc82d16d022da3cbf9992d2f70"}
```

## Miner Methods

### `miner_setGasPrice`

Sets the minimal gas price used to accept transactions. Any transaction below this limit is excluded from the validator block proposal process.

This method requires a `node` restart after being called because it changes the configuration file.

Make sure your `ethermintd start` call is not using the flag `minimum-gas-prices` because this value will be used instead of the one set on the configuration file.

#### Parameters

- Hex Gas Price

```json
// Request
curl -X POST --data '{"jsonrpc":"2.0","method":"miner_setGasPrice","params":["0x0"],"id":1}' -H "Content-Type: application/json" http://localhost:8545

// Result
{"jsonrpc":"2.0","id":1,"result":true}
```

### `miner_setEtherbase`

Sets the etherbase. It changes the wallet where the validator rewards will be deposited.

#### Parameters

- Account Address

```json
// Request
curl -X POST --data '{"jsonrpc":"2.0","method":"miner_setEtherbase","params":["0x3b7252d007059ffc82d16d022da3cbf9992d2f70"],"id":1}' -H "Content-Type: application/json" http://localhost:8545

// Result
{"jsonrpc":"2.0","id":1,"result":true}
```
