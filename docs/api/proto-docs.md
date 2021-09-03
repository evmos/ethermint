<!-- This file is auto-generated. Please do not modify it yourself. -->
# Protobuf Documentation
<a name="top"></a>

## Table of Contents

- [ethermint/crypto/v1/ethsecp256k1/keys.proto](#ethermint/crypto/v1/ethsecp256k1/keys.proto)
    - [PrivKey](#ethermint.crypto.v1.ethsecp256k1.PrivKey)
    - [PubKey](#ethermint.crypto.v1.ethsecp256k1.PubKey)
  
- [ethermint/evm/v1/evm.proto](#ethermint/evm/v1/evm.proto)
    - [AccessTuple](#ethermint.evm.v1.AccessTuple)
    - [ChainConfig](#ethermint.evm.v1.ChainConfig)
    - [Log](#ethermint.evm.v1.Log)
    - [Params](#ethermint.evm.v1.Params)
    - [State](#ethermint.evm.v1.State)
    - [TransactionLogs](#ethermint.evm.v1.TransactionLogs)
    - [TxResult](#ethermint.evm.v1.TxResult)
  
- [ethermint/evm/v1/genesis.proto](#ethermint/evm/v1/genesis.proto)
    - [GenesisAccount](#ethermint.evm.v1.GenesisAccount)
    - [GenesisState](#ethermint.evm.v1.GenesisState)
  
- [ethermint/evm/v1/tx.proto](#ethermint/evm/v1/tx.proto)
    - [AccessListTx](#ethermint.evm.v1.AccessListTx)
    - [DynamicFeeTx](#ethermint.evm.v1.DynamicFeeTx)
    - [ExtensionOptionsEthereumTx](#ethermint.evm.v1.ExtensionOptionsEthereumTx)
    - [LegacyTx](#ethermint.evm.v1.LegacyTx)
    - [MsgEthereumTx](#ethermint.evm.v1.MsgEthereumTx)
    - [MsgEthereumTxResponse](#ethermint.evm.v1.MsgEthereumTxResponse)
  
    - [Msg](#ethermint.evm.v1.Msg)
  
- [ethermint/evm/v1/query.proto](#ethermint/evm/v1/query.proto)
    - [EstimateGasResponse](#ethermint.evm.v1.EstimateGasResponse)
    - [EthCallRequest](#ethermint.evm.v1.EthCallRequest)
    - [QueryAccountRequest](#ethermint.evm.v1.QueryAccountRequest)
    - [QueryAccountResponse](#ethermint.evm.v1.QueryAccountResponse)
    - [QueryBalanceRequest](#ethermint.evm.v1.QueryBalanceRequest)
    - [QueryBalanceResponse](#ethermint.evm.v1.QueryBalanceResponse)
    - [QueryBaseFeeRequest](#ethermint.evm.v1.QueryBaseFeeRequest)
    - [QueryBaseFeeResponse](#ethermint.evm.v1.QueryBaseFeeResponse)
    - [QueryBlockBloomRequest](#ethermint.evm.v1.QueryBlockBloomRequest)
    - [QueryBlockBloomResponse](#ethermint.evm.v1.QueryBlockBloomResponse)
    - [QueryBlockLogsRequest](#ethermint.evm.v1.QueryBlockLogsRequest)
    - [QueryBlockLogsResponse](#ethermint.evm.v1.QueryBlockLogsResponse)
    - [QueryCodeRequest](#ethermint.evm.v1.QueryCodeRequest)
    - [QueryCodeResponse](#ethermint.evm.v1.QueryCodeResponse)
    - [QueryCosmosAccountRequest](#ethermint.evm.v1.QueryCosmosAccountRequest)
    - [QueryCosmosAccountResponse](#ethermint.evm.v1.QueryCosmosAccountResponse)
    - [QueryParamsRequest](#ethermint.evm.v1.QueryParamsRequest)
    - [QueryParamsResponse](#ethermint.evm.v1.QueryParamsResponse)
    - [QueryStaticCallResponse](#ethermint.evm.v1.QueryStaticCallResponse)
    - [QueryStorageRequest](#ethermint.evm.v1.QueryStorageRequest)
    - [QueryStorageResponse](#ethermint.evm.v1.QueryStorageResponse)
    - [QueryTxLogsRequest](#ethermint.evm.v1.QueryTxLogsRequest)
    - [QueryTxLogsResponse](#ethermint.evm.v1.QueryTxLogsResponse)
    - [QueryValidatorAccountRequest](#ethermint.evm.v1.QueryValidatorAccountRequest)
    - [QueryValidatorAccountResponse](#ethermint.evm.v1.QueryValidatorAccountResponse)
  
    - [Query](#ethermint.evm.v1.Query)
  
- [ethermint/feemarket/v1/feemarket.proto](#ethermint/feemarket/v1/feemarket.proto)
    - [Params](#ethermint.feemarket.v1.Params)
  
- [ethermint/feemarket/v1/genesis.proto](#ethermint/feemarket/v1/genesis.proto)
    - [GenesisState](#ethermint.feemarket.v1.GenesisState)
  
- [ethermint/feemarket/v1/query.proto](#ethermint/feemarket/v1/query.proto)
    - [QueryBaseFeeRequest](#ethermint.feemarket.v1.QueryBaseFeeRequest)
    - [QueryBaseFeeResponse](#ethermint.feemarket.v1.QueryBaseFeeResponse)
    - [QueryBlockGasRequest](#ethermint.feemarket.v1.QueryBlockGasRequest)
    - [QueryBlockGasResponse](#ethermint.feemarket.v1.QueryBlockGasResponse)
    - [QueryParamsRequest](#ethermint.feemarket.v1.QueryParamsRequest)
    - [QueryParamsResponse](#ethermint.feemarket.v1.QueryParamsResponse)
  
    - [Query](#ethermint.feemarket.v1.Query)
  
- [ethermint/types/v1/account.proto](#ethermint/types/v1/account.proto)
    - [EthAccount](#ethermint.types.v1.EthAccount)
  
- [ethermint/types/v1/web3.proto](#ethermint/types/v1/web3.proto)
    - [ExtensionOptionsWeb3Tx](#ethermint.types.v1.ExtensionOptionsWeb3Tx)
  
- [Scalar Value Types](#scalar-value-types)



<a name="ethermint/crypto/v1/ethsecp256k1/keys.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## ethermint/crypto/v1/ethsecp256k1/keys.proto



<a name="ethermint.crypto.v1.ethsecp256k1.PrivKey"></a>

### PrivKey
PrivKey defines a type alias for an ecdsa.PrivateKey that implements
Tendermint's PrivateKey interface.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `key` | [bytes](#bytes) |  |  |






<a name="ethermint.crypto.v1.ethsecp256k1.PubKey"></a>

### PubKey
PubKey defines a type alias for an ecdsa.PublicKey that implements
Tendermint's PubKey interface. It represents the 33-byte compressed public
key format.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `key` | [bytes](#bytes) |  |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="ethermint/evm/v1/evm.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## ethermint/evm/v1/evm.proto



<a name="ethermint.evm.v1.AccessTuple"></a>

### AccessTuple
AccessTuple is the element type of an access list.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `address` | [string](#string) |  | hex formatted ethereum address |
| `storage_keys` | [string](#string) | repeated | hex formatted hashes of the storage keys |






<a name="ethermint.evm.v1.ChainConfig"></a>

### ChainConfig
ChainConfig defines the Ethereum ChainConfig parameters using *sdk.Int values
instead of *big.Int.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `homestead_block` | [string](#string) |  | Homestead switch block (nil no fork, 0 = already homestead) |
| `dao_fork_block` | [string](#string) |  | TheDAO hard-fork switch block (nil no fork) |
| `dao_fork_support` | [bool](#bool) |  | Whether the nodes supports or opposes the DAO hard-fork |
| `eip150_block` | [string](#string) |  | EIP150 implements the Gas price changes (https://github.com/ethereum/EIPs/issues/150) EIP150 HF block (nil no fork) |
| `eip150_hash` | [string](#string) |  | EIP150 HF hash (needed for header only clients as only gas pricing changed) |
| `eip155_block` | [string](#string) |  | EIP155Block HF block |
| `eip158_block` | [string](#string) |  | EIP158 HF block |
| `byzantium_block` | [string](#string) |  | Byzantium switch block (nil no fork, 0 = already on byzantium) |
| `constantinople_block` | [string](#string) |  | Constantinople switch block (nil no fork, 0 = already activated) |
| `petersburg_block` | [string](#string) |  | Petersburg switch block (nil same as Constantinople) |
| `istanbul_block` | [string](#string) |  | Istanbul switch block (nil no fork, 0 = already on istanbul) |
| `muir_glacier_block` | [string](#string) |  | Eip-2384 (bomb delay) switch block (nil no fork, 0 = already activated) |
| `berlin_block` | [string](#string) |  | Berlin switch block (nil = no fork, 0 = already on berlin) |
| `catalyst_block` | [string](#string) |  | Catalyst switch block (nil = no fork, 0 = already on catalyst) |
| `london_block` | [string](#string) |  | London switch block (nil = no fork, 0 = already on london) |






<a name="ethermint.evm.v1.Log"></a>

### Log
Log represents an protobuf compatible Ethereum Log that defines a contract
log event. These events are generated by the LOG opcode and stored/indexed by
the node.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `address` | [string](#string) |  | address of the contract that generated the event |
| `topics` | [string](#string) | repeated | list of topics provided by the contract. |
| `data` | [bytes](#bytes) |  | supplied by the contract, usually ABI-encoded |
| `block_number` | [uint64](#uint64) |  | block in which the transaction was included |
| `tx_hash` | [string](#string) |  | hash of the transaction |
| `tx_index` | [uint64](#uint64) |  | index of the transaction in the block |
| `block_hash` | [string](#string) |  | hash of the block in which the transaction was included |
| `index` | [uint64](#uint64) |  | index of the log in the block |
| `removed` | [bool](#bool) |  | The Removed field is true if this log was reverted due to a chain reorganisation. You must pay attention to this field if you receive logs through a filter query. |






<a name="ethermint.evm.v1.Params"></a>

### Params
Params defines the EVM module parameters


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `evm_denom` | [string](#string) |  | evm denom represents the token denomination used to run the EVM state transitions. |
| `enable_create` | [bool](#bool) |  | enable create toggles state transitions that use the vm.Create function |
| `enable_call` | [bool](#bool) |  | enable call toggles state transitions that use the vm.Call function |
| `extra_eips` | [int64](#int64) | repeated | extra eips defines the additional EIPs for the vm.Config |
| `chain_config` | [ChainConfig](#ethermint.evm.v1.ChainConfig) |  | chain config defines the EVM chain configuration parameters |






<a name="ethermint.evm.v1.State"></a>

### State
State represents a single Storage key value pair item.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `key` | [string](#string) |  |  |
| `value` | [string](#string) |  |  |






<a name="ethermint.evm.v1.TransactionLogs"></a>

### TransactionLogs
TransactionLogs define the logs generated from a transaction execution
with a given hash. It it used for import/export data as transactions are not
persisted on blockchain state after an upgrade.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `hash` | [string](#string) |  |  |
| `logs` | [Log](#ethermint.evm.v1.Log) | repeated |  |






<a name="ethermint.evm.v1.TxResult"></a>

### TxResult
TxResult stores results of Tx execution.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `contract_address` | [string](#string) |  | contract_address contains the ethereum address of the created contract (if any). If the state transition is an evm.Call, the contract address will be empty. |
| `bloom` | [bytes](#bytes) |  | bloom represents the bloom filter bytes |
| `tx_logs` | [TransactionLogs](#ethermint.evm.v1.TransactionLogs) |  | tx_logs contains the transaction hash and the proto-compatible ethereum logs. |
| `ret` | [bytes](#bytes) |  | ret defines the bytes from the execution. |
| `reverted` | [bool](#bool) |  | reverted flag is set to true when the call has been reverted |
| `gas_used` | [uint64](#uint64) |  | gas_used notes the amount of gas consumed while execution |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="ethermint/evm/v1/genesis.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## ethermint/evm/v1/genesis.proto



<a name="ethermint.evm.v1.GenesisAccount"></a>

### GenesisAccount
GenesisAccount defines an account to be initialized in the genesis state.
Its main difference between with Geth's GenesisAccount is that it uses a
custom storage type and that it doesn't contain the private key field.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `address` | [string](#string) |  | address defines an ethereum hex formated address of an account |
| `code` | [string](#string) |  | code defines the hex bytes of the account code. |
| `storage` | [State](#ethermint.evm.v1.State) | repeated | storage defines the set of state key values for the account. |






<a name="ethermint.evm.v1.GenesisState"></a>

### GenesisState
GenesisState defines the evm module's genesis state.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `accounts` | [GenesisAccount](#ethermint.evm.v1.GenesisAccount) | repeated | accounts is an array containing the ethereum genesis accounts. |
| `params` | [Params](#ethermint.evm.v1.Params) |  | params defines all the paramaters of the module. |
| `txs_logs` | [TransactionLogs](#ethermint.evm.v1.TransactionLogs) | repeated |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="ethermint/evm/v1/tx.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## ethermint/evm/v1/tx.proto



<a name="ethermint.evm.v1.AccessListTx"></a>

### AccessListTx
AccessListTx is the data of EIP-2930 access list transactions.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `chain_id` | [string](#string) |  | destination EVM chain ID |
| `nonce` | [uint64](#uint64) |  | nonce corresponds to the account nonce (transaction sequence). |
| `gas_price` | [string](#string) |  | gas price defines the value for each gas unit |
| `gas` | [uint64](#uint64) |  | gas defines the gas limit defined for the transaction. |
| `to` | [string](#string) |  | hex formatted address of the recipient |
| `value` | [string](#string) |  | value defines the unsigned integer value of the transaction amount. |
| `data` | [bytes](#bytes) |  | input defines the data payload bytes of the transaction. |
| `accesses` | [AccessTuple](#ethermint.evm.v1.AccessTuple) | repeated |  |
| `v` | [bytes](#bytes) |  | v defines the signature value |
| `r` | [bytes](#bytes) |  | r defines the signature value |
| `s` | [bytes](#bytes) |  | s define the signature value |






<a name="ethermint.evm.v1.DynamicFeeTx"></a>

### DynamicFeeTx
DynamicFeeTx is the data of EIP-1559 dinamic fee transactions.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `chain_id` | [string](#string) |  | destination EVM chain ID |
| `nonce` | [uint64](#uint64) |  | nonce corresponds to the account nonce (transaction sequence). |
| `gas_tip_cap` | [string](#string) |  | gas tip cap defines the max value for the gas tip |
| `gas_fee_cap` | [string](#string) |  | gas fee cap defines the max value for the gas fee |
| `gas` | [uint64](#uint64) |  | gas defines the gas limit defined for the transaction. |
| `to` | [string](#string) |  | hex formatted address of the recipient |
| `value` | [string](#string) |  | value defines the the transaction amount. |
| `data` | [bytes](#bytes) |  | input defines the data payload bytes of the transaction. |
| `accesses` | [AccessTuple](#ethermint.evm.v1.AccessTuple) | repeated |  |
| `v` | [bytes](#bytes) |  | v defines the signature value |
| `r` | [bytes](#bytes) |  | r defines the signature value |
| `s` | [bytes](#bytes) |  | s define the signature value |






<a name="ethermint.evm.v1.ExtensionOptionsEthereumTx"></a>

### ExtensionOptionsEthereumTx







<a name="ethermint.evm.v1.LegacyTx"></a>

### LegacyTx
LegacyTx is the transaction data of regular Ethereum transactions.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `nonce` | [uint64](#uint64) |  | nonce corresponds to the account nonce (transaction sequence). |
| `gas_price` | [string](#string) |  | gas price defines the value for each gas unit |
| `gas` | [uint64](#uint64) |  | gas defines the gas limit defined for the transaction. |
| `to` | [string](#string) |  | hex formatted address of the recipient |
| `value` | [string](#string) |  | value defines the unsigned integer value of the transaction amount. |
| `data` | [bytes](#bytes) |  | input defines the data payload bytes of the transaction. |
| `v` | [bytes](#bytes) |  | v defines the signature value |
| `r` | [bytes](#bytes) |  | r defines the signature value |
| `s` | [bytes](#bytes) |  | s define the signature value |






<a name="ethermint.evm.v1.MsgEthereumTx"></a>

### MsgEthereumTx
MsgEthereumTx encapsulates an Ethereum transaction as an SDK message.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `data` | [google.protobuf.Any](#google.protobuf.Any) |  | inner transaction data

caches |
| `size` | [double](#double) |  | encoded storage size of the transaction |
| `hash` | [string](#string) |  | transaction hash in hex format |
| `from` | [string](#string) |  | ethereum signer address in hex format. This address value is checked against the address derived from the signature (V, R, S) using the secp256k1 elliptic curve |






<a name="ethermint.evm.v1.MsgEthereumTxResponse"></a>

### MsgEthereumTxResponse
MsgEthereumTxResponse defines the Msg/EthereumTx response type.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `hash` | [string](#string) |  | ethereum transaction hash in hex format. This hash differs from the Tendermint sha256 hash of the transaction bytes. See https://github.com/tendermint/tendermint/issues/6539 for reference |
| `logs` | [Log](#ethermint.evm.v1.Log) | repeated | logs contains the transaction hash and the proto-compatible ethereum logs. |
| `ret` | [bytes](#bytes) |  | returned data from evm function (result or data supplied with revert opcode) |
| `vm_error` | [string](#string) |  | vm error is the error returned by vm execution |
| `gas_used` | [uint64](#uint64) |  | gas consumed by the transaction |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->


<a name="ethermint.evm.v1.Msg"></a>

### Msg
Msg defines the evm Msg service.

| Method Name | Request Type | Response Type | Description | HTTP Verb | Endpoint |
| ----------- | ------------ | ------------- | ------------| ------- | -------- |
| `EthereumTx` | [MsgEthereumTx](#ethermint.evm.v1.MsgEthereumTx) | [MsgEthereumTxResponse](#ethermint.evm.v1.MsgEthereumTxResponse) | EthereumTx defines a method submitting Ethereum transactions. | |

 <!-- end services -->



<a name="ethermint/evm/v1/query.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## ethermint/evm/v1/query.proto



<a name="ethermint.evm.v1.EstimateGasResponse"></a>

### EstimateGasResponse
EstimateGasResponse defines EstimateGas response


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `gas` | [uint64](#uint64) |  | the estimated gas |






<a name="ethermint.evm.v1.EthCallRequest"></a>

### EthCallRequest
EthCallRequest defines EthCall request


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `args` | [bytes](#bytes) |  | same json format as the json rpc api. |
| `gas_cap` | [uint64](#uint64) |  | the default gas cap to be used |






<a name="ethermint.evm.v1.QueryAccountRequest"></a>

### QueryAccountRequest
QueryAccountRequest is the request type for the Query/Account RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `address` | [string](#string) |  | address is the ethereum hex address to query the account for. |






<a name="ethermint.evm.v1.QueryAccountResponse"></a>

### QueryAccountResponse
QueryAccountResponse is the response type for the Query/Account RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `balance` | [string](#string) |  | balance is the balance of the EVM denomination. |
| `code_hash` | [string](#string) |  | code hash is the hex-formatted code bytes from the EOA. |
| `nonce` | [uint64](#uint64) |  | nonce is the account's sequence number. |






<a name="ethermint.evm.v1.QueryBalanceRequest"></a>

### QueryBalanceRequest
QueryBalanceRequest is the request type for the Query/Balance RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `address` | [string](#string) |  | address is the ethereum hex address to query the balance for. |






<a name="ethermint.evm.v1.QueryBalanceResponse"></a>

### QueryBalanceResponse
QueryBalanceResponse is the response type for the Query/Balance RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `balance` | [string](#string) |  | balance is the balance of the EVM denomination. |






<a name="ethermint.evm.v1.QueryBaseFeeRequest"></a>

### QueryBaseFeeRequest
QueryBaseFeeRequest defines the request type for querying the EIP1559 base
fee.






<a name="ethermint.evm.v1.QueryBaseFeeResponse"></a>

### QueryBaseFeeResponse
BaseFeeResponse returns the EIP1559 base fee.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `base_fee` | [string](#string) |  |  |






<a name="ethermint.evm.v1.QueryBlockBloomRequest"></a>

### QueryBlockBloomRequest
QueryBlockBloomRequest is the request type for the Query/BlockBloom RPC
method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `height` | [int64](#int64) |  | height of the block which we want to query the bloom filter. Tendermint always replace the query request header by the current context header, height cannot be extracted from there, so we need to explicitly pass it in parameter. |






<a name="ethermint.evm.v1.QueryBlockBloomResponse"></a>

### QueryBlockBloomResponse
QueryBlockBloomResponse is the response type for the Query/BlockBloom RPC
method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `bloom` | [bytes](#bytes) |  | bloom represents bloom filter for the given block hash. |






<a name="ethermint.evm.v1.QueryBlockLogsRequest"></a>

### QueryBlockLogsRequest
QueryBlockLogsRequest is the request type for the Query/BlockLogs RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `hash` | [string](#string) |  | hash is the block hash to query the logs for. |
| `pagination` | [cosmos.base.query.v1beta1.PageRequest](#cosmos.base.query.v1beta1.PageRequest) |  | pagination defines an optional pagination for the request. |






<a name="ethermint.evm.v1.QueryBlockLogsResponse"></a>

### QueryBlockLogsResponse
QueryTxLogs is the response type for the Query/BlockLogs RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `tx_logs` | [TransactionLogs](#ethermint.evm.v1.TransactionLogs) | repeated | logs represents the ethereum logs generated at the given block hash. |
| `pagination` | [cosmos.base.query.v1beta1.PageResponse](#cosmos.base.query.v1beta1.PageResponse) |  | pagination defines the pagination in the response. |






<a name="ethermint.evm.v1.QueryCodeRequest"></a>

### QueryCodeRequest
QueryCodeRequest is the request type for the Query/Code RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `address` | [string](#string) |  | address is the ethereum hex address to query the code for. |






<a name="ethermint.evm.v1.QueryCodeResponse"></a>

### QueryCodeResponse
QueryCodeResponse is the response type for the Query/Code RPC
method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `code` | [bytes](#bytes) |  | code represents the code bytes from an ethereum address. |






<a name="ethermint.evm.v1.QueryCosmosAccountRequest"></a>

### QueryCosmosAccountRequest
QueryCosmosAccountRequest is the request type for the Query/CosmosAccount RPC
method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `address` | [string](#string) |  | address is the ethereum hex address to query the account for. |






<a name="ethermint.evm.v1.QueryCosmosAccountResponse"></a>

### QueryCosmosAccountResponse
QueryCosmosAccountResponse is the response type for the Query/CosmosAccount
RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `cosmos_address` | [string](#string) |  | cosmos_address is the cosmos address of the account. |
| `sequence` | [uint64](#uint64) |  | sequence is the account's sequence number. |
| `account_number` | [uint64](#uint64) |  | account_number is the account numbert |






<a name="ethermint.evm.v1.QueryParamsRequest"></a>

### QueryParamsRequest
QueryParamsRequest defines the request type for querying x/evm parameters.






<a name="ethermint.evm.v1.QueryParamsResponse"></a>

### QueryParamsResponse
QueryParamsResponse defines the response type for querying x/evm parameters.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `params` | [Params](#ethermint.evm.v1.Params) |  | params define the evm module parameters. |






<a name="ethermint.evm.v1.QueryStaticCallResponse"></a>

### QueryStaticCallResponse
QueryStaticCallRequest defines static call response


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `data` | [bytes](#bytes) |  |  |






<a name="ethermint.evm.v1.QueryStorageRequest"></a>

### QueryStorageRequest
QueryStorageRequest is the request type for the Query/Storage RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `address` | [string](#string) |  | address is the ethereum hex address to query the storage state for. |
| `key` | [string](#string) |  | key defines the key of the storage state |






<a name="ethermint.evm.v1.QueryStorageResponse"></a>

### QueryStorageResponse
QueryStorageResponse is the response type for the Query/Storage RPC
method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `value` | [string](#string) |  | key defines the storage state value hash associated with the given key. |






<a name="ethermint.evm.v1.QueryTxLogsRequest"></a>

### QueryTxLogsRequest
QueryTxLogsRequest is the request type for the Query/TxLogs RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `hash` | [string](#string) |  | hash is the ethereum transaction hex hash to query the logs for. |






<a name="ethermint.evm.v1.QueryTxLogsResponse"></a>

### QueryTxLogsResponse
QueryTxLogs is the response type for the Query/TxLogs RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `logs` | [Log](#ethermint.evm.v1.Log) | repeated | logs represents the ethereum logs generated from the given transaction. |






<a name="ethermint.evm.v1.QueryValidatorAccountRequest"></a>

### QueryValidatorAccountRequest
QueryValidatorAccountRequest is the request type for the
Query/ValidatorAccount RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `cons_address` | [string](#string) |  | cons_address is the validator cons address to query the account for. |






<a name="ethermint.evm.v1.QueryValidatorAccountResponse"></a>

### QueryValidatorAccountResponse
QueryValidatorAccountResponse is the response type for the
Query/ValidatorAccount RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `account_address` | [string](#string) |  | account_address is the cosmos address of the account in bech32 format. |
| `sequence` | [uint64](#uint64) |  | sequence is the account's sequence number. |
| `account_number` | [uint64](#uint64) |  | account_number is the account number |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->


<a name="ethermint.evm.v1.Query"></a>

### Query
Query defines the gRPC querier service.

| Method Name | Request Type | Response Type | Description | HTTP Verb | Endpoint |
| ----------- | ------------ | ------------- | ------------| ------- | -------- |
| `Account` | [QueryAccountRequest](#ethermint.evm.v1.QueryAccountRequest) | [QueryAccountResponse](#ethermint.evm.v1.QueryAccountResponse) | Account queries an Ethereum account. | GET|/ethermint/evm/v1/account/{address}|
| `CosmosAccount` | [QueryCosmosAccountRequest](#ethermint.evm.v1.QueryCosmosAccountRequest) | [QueryCosmosAccountResponse](#ethermint.evm.v1.QueryCosmosAccountResponse) | CosmosAccount queries an Ethereum account's Cosmos Address. | GET|/ethermint/evm/v1/cosmos_account/{address}|
| `ValidatorAccount` | [QueryValidatorAccountRequest](#ethermint.evm.v1.QueryValidatorAccountRequest) | [QueryValidatorAccountResponse](#ethermint.evm.v1.QueryValidatorAccountResponse) | ValidatorAccount queries an Ethereum account's from a validator consensus Address. | GET|/ethermint/evm/v1/validator_account/{cons_address}|
| `Balance` | [QueryBalanceRequest](#ethermint.evm.v1.QueryBalanceRequest) | [QueryBalanceResponse](#ethermint.evm.v1.QueryBalanceResponse) | Balance queries the balance of a the EVM denomination for a single EthAccount. | GET|/ethermint/evm/v1/balances/{address}|
| `Storage` | [QueryStorageRequest](#ethermint.evm.v1.QueryStorageRequest) | [QueryStorageResponse](#ethermint.evm.v1.QueryStorageResponse) | Storage queries the balance of all coins for a single account. | GET|/ethermint/evm/v1/storage/{address}/{key}|
| `Code` | [QueryCodeRequest](#ethermint.evm.v1.QueryCodeRequest) | [QueryCodeResponse](#ethermint.evm.v1.QueryCodeResponse) | Code queries the balance of all coins for a single account. | GET|/ethermint/evm/v1/codes/{address}|
| `TxLogs` | [QueryTxLogsRequest](#ethermint.evm.v1.QueryTxLogsRequest) | [QueryTxLogsResponse](#ethermint.evm.v1.QueryTxLogsResponse) | TxLogs queries ethereum logs from a transaction. | GET|/ethermint/evm/v1/tx_logs/{hash}|
| `BlockLogs` | [QueryBlockLogsRequest](#ethermint.evm.v1.QueryBlockLogsRequest) | [QueryBlockLogsResponse](#ethermint.evm.v1.QueryBlockLogsResponse) | BlockLogs queries all the ethereum logs for a given block hash. | GET|/ethermint/evm/v1/block_logs/{hash}|
| `BlockBloom` | [QueryBlockBloomRequest](#ethermint.evm.v1.QueryBlockBloomRequest) | [QueryBlockBloomResponse](#ethermint.evm.v1.QueryBlockBloomResponse) | BlockBloom queries the block bloom filter bytes at a given height. | GET|/ethermint/evm/v1/block_bloom|
| `Params` | [QueryParamsRequest](#ethermint.evm.v1.QueryParamsRequest) | [QueryParamsResponse](#ethermint.evm.v1.QueryParamsResponse) | Params queries the parameters of x/evm module. | GET|/ethermint/evm/v1/params|
| `EthCall` | [EthCallRequest](#ethermint.evm.v1.EthCallRequest) | [MsgEthereumTxResponse](#ethermint.evm.v1.MsgEthereumTxResponse) | EthCall implements the `eth_call` rpc api | GET|/ethermint/evm/v1/eth_call|
| `EstimateGas` | [EthCallRequest](#ethermint.evm.v1.EthCallRequest) | [EstimateGasResponse](#ethermint.evm.v1.EstimateGasResponse) | EstimateGas implements the `eth_estimateGas` rpc api | GET|/ethermint/evm/v1/estimate_gas|

 <!-- end services -->



<a name="ethermint/feemarket/v1/feemarket.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## ethermint/feemarket/v1/feemarket.proto



<a name="ethermint.feemarket.v1.Params"></a>

### Params
Params defines the EVM module parameters


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `no_base_fee` | [bool](#bool) |  | no base fee forces the EIP-1559 base fee to 0 (needed for 0 price calls) |
| `base_fee_change_denominator` | [uint32](#uint32) |  | base fee change denominator bounds the amount the base fee can change between blocks. |
| `elasticity_multiplier` | [uint32](#uint32) |  | elasticity multiplier bounds the maximum gas limit an EIP-1559 block may have. |
| `initial_base_fee` | [int64](#int64) |  | initial base fee for EIP-1559 blocks. |
| `enable_height` | [int64](#int64) |  | height at which the base fee calculation is enabled. |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="ethermint/feemarket/v1/genesis.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## ethermint/feemarket/v1/genesis.proto



<a name="ethermint.feemarket.v1.GenesisState"></a>

### GenesisState
GenesisState defines the feemarket module's genesis state.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `params` | [Params](#ethermint.feemarket.v1.Params) |  | params defines all the paramaters of the module. |
| `base_fee` | [string](#string) |  | base fee is the exported value from previous software version. Zero by default. |
| `block_gas` | [uint64](#uint64) |  | block gas is the amount of gas used on the last block before the upgrade. Zero by default. |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="ethermint/feemarket/v1/query.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## ethermint/feemarket/v1/query.proto



<a name="ethermint.feemarket.v1.QueryBaseFeeRequest"></a>

### QueryBaseFeeRequest
QueryBaseFeeRequest defines the request type for querying the EIP1559 base
fee.






<a name="ethermint.feemarket.v1.QueryBaseFeeResponse"></a>

### QueryBaseFeeResponse
BaseFeeResponse returns the EIP1559 base fee.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `base_fee` | [string](#string) |  |  |






<a name="ethermint.feemarket.v1.QueryBlockGasRequest"></a>

### QueryBlockGasRequest
QueryBlockGasRequest defines the request type for querying the EIP1559 base
fee.






<a name="ethermint.feemarket.v1.QueryBlockGasResponse"></a>

### QueryBlockGasResponse
QueryBlockGasResponse returns block gas used for a given height.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `gas` | [int64](#int64) |  |  |






<a name="ethermint.feemarket.v1.QueryParamsRequest"></a>

### QueryParamsRequest
QueryParamsRequest defines the request type for querying x/evm parameters.






<a name="ethermint.feemarket.v1.QueryParamsResponse"></a>

### QueryParamsResponse
QueryParamsResponse defines the response type for querying x/evm parameters.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `params` | [Params](#ethermint.feemarket.v1.Params) |  | params define the evm module parameters. |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->


<a name="ethermint.feemarket.v1.Query"></a>

### Query
Query defines the gRPC querier service.

| Method Name | Request Type | Response Type | Description | HTTP Verb | Endpoint |
| ----------- | ------------ | ------------- | ------------| ------- | -------- |
| `Params` | [QueryParamsRequest](#ethermint.feemarket.v1.QueryParamsRequest) | [QueryParamsResponse](#ethermint.feemarket.v1.QueryParamsResponse) | Params queries the parameters of x/feemarket module. | GET|/feemarket/evm/v1/params|
| `BaseFee` | [QueryBaseFeeRequest](#ethermint.feemarket.v1.QueryBaseFeeRequest) | [QueryBaseFeeResponse](#ethermint.feemarket.v1.QueryBaseFeeResponse) | BaseFee queries the base fee of the parent block of the current block. | GET|/feemarket/evm/v1/base_fee|
| `BlockGas` | [QueryBlockGasRequest](#ethermint.feemarket.v1.QueryBlockGasRequest) | [QueryBlockGasResponse](#ethermint.feemarket.v1.QueryBlockGasResponse) | BlockGas queries the gas used at a given block height | GET|/feemarket/evm/v1/block_gas|

 <!-- end services -->



<a name="ethermint/types/v1/account.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## ethermint/types/v1/account.proto



<a name="ethermint.types.v1.EthAccount"></a>

### EthAccount
EthAccount implements the authtypes.AccountI interface and embeds an
authtypes.BaseAccount type. It is compatible with the auth AccountKeeper.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `base_account` | [cosmos.auth.v1beta1.BaseAccount](#cosmos.auth.v1beta1.BaseAccount) |  |  |
| `code_hash` | [string](#string) |  |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="ethermint/types/v1/web3.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## ethermint/types/v1/web3.proto



<a name="ethermint.types.v1.ExtensionOptionsWeb3Tx"></a>

### ExtensionOptionsWeb3Tx



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `typed_data_chain_id` | [uint64](#uint64) |  | typed data chain id used only in EIP712 Domain and should match Ethereum network ID in a Web3 provider (e.g. Metamask). |
| `fee_payer` | [string](#string) |  | fee payer is an account address for the fee payer. It will be validated during EIP712 signature checking. |
| `fee_payer_sig` | [bytes](#bytes) |  | fee payer sig is a signature data from the fee paying account, allows to perform fee delegation when using EIP712 Domain. |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



## Scalar Value Types

| .proto Type | Notes | C++ | Java | Python | Go | C# | PHP | Ruby |
| ----------- | ----- | --- | ---- | ------ | -- | -- | --- | ---- |
| <a name="double" /> double |  | double | double | float | float64 | double | float | Float |
| <a name="float" /> float |  | float | float | float | float32 | float | float | Float |
| <a name="int32" /> int32 | Uses variable-length encoding. Inefficient for encoding negative numbers – if your field is likely to have negative values, use sint32 instead. | int32 | int | int | int32 | int | integer | Bignum or Fixnum (as required) |
| <a name="int64" /> int64 | Uses variable-length encoding. Inefficient for encoding negative numbers – if your field is likely to have negative values, use sint64 instead. | int64 | long | int/long | int64 | long | integer/string | Bignum |
| <a name="uint32" /> uint32 | Uses variable-length encoding. | uint32 | int | int/long | uint32 | uint | integer | Bignum or Fixnum (as required) |
| <a name="uint64" /> uint64 | Uses variable-length encoding. | uint64 | long | int/long | uint64 | ulong | integer/string | Bignum or Fixnum (as required) |
| <a name="sint32" /> sint32 | Uses variable-length encoding. Signed int value. These more efficiently encode negative numbers than regular int32s. | int32 | int | int | int32 | int | integer | Bignum or Fixnum (as required) |
| <a name="sint64" /> sint64 | Uses variable-length encoding. Signed int value. These more efficiently encode negative numbers than regular int64s. | int64 | long | int/long | int64 | long | integer/string | Bignum |
| <a name="fixed32" /> fixed32 | Always four bytes. More efficient than uint32 if values are often greater than 2^28. | uint32 | int | int | uint32 | uint | integer | Bignum or Fixnum (as required) |
| <a name="fixed64" /> fixed64 | Always eight bytes. More efficient than uint64 if values are often greater than 2^56. | uint64 | long | int/long | uint64 | ulong | integer/string | Bignum |
| <a name="sfixed32" /> sfixed32 | Always four bytes. | int32 | int | int | int32 | int | integer | Bignum or Fixnum (as required) |
| <a name="sfixed64" /> sfixed64 | Always eight bytes. | int64 | long | int/long | int64 | long | integer/string | Bignum |
| <a name="bool" /> bool |  | bool | boolean | boolean | bool | bool | boolean | TrueClass/FalseClass |
| <a name="string" /> string | A string must always contain UTF-8 encoded or 7-bit ASCII text. | string | String | str/unicode | string | string | string | String (UTF-8) |
| <a name="bytes" /> bytes | May contain any arbitrary sequence of bytes. | string | ByteString | str | []byte | ByteString | string | String (ASCII-8BIT) |

