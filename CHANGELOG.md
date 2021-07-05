<!--
Guiding Principles:

Changelogs are for humans, not machines.
There should be an entry for every single version.
The same types of changes should be grouped.
Versions and sections should be linkable.
The latest version comes first.
The release date of each version is displayed.
Mention whether you follow Semantic Versioning.

Usage:

Change log entries are to be added to the Unreleased section under the
appropriate stanza (see below). Each entry should ideally include a tag and
the Github issue reference in the following format:

* (<tag>) \#<issue-number> message

The issue numbers will later be link-ified during the release process so you do
not have to worry about including a link manually, but you can if you wish.

Types of changes (Stanzas):

"Features" for new features.
"Improvements" for changes in existing functionality.
"Deprecated" for soon-to-be removed features.
"Bug Fixes" for any bug fixes.
"Client Breaking" for breaking CLI commands and REST routes used by end-users.
"API Breaking" for breaking exported APIs used by developers building on SDK.
"State Machine Breaking" for any changes that result in a different AppState given same genesisState and txList.

Ref: https://keepachangelog.com/en/1.0.0/
-->

# Changelog

## Unreleased

### State Machine Breaking

* (evm) [tharsis#84](https://github.com/tharsis/ethermint/pull/84) Remove `journal`, `CommitStateDB` and `stateObjects`.
* (rpc, evm) [tharsis#81](https://github.com/tharsis/ethermint/pull/81) Remove tx `Receipt` from store and replace it with fields obtained from the Tendermint RPC client.
* (evm) [tharsis#72](https://github.com/tharsis/ethermint/issues/72) Update `AccessList` to use `TransientStore` instead of map.
* (evm) [tharsis#68](https://github.com/tharsis/ethermint/issues/68) Replace block hash storage map to use staking `HistoricalInfo`.

### API Breaking

* (proto, evm) [tharsis#207](https://github.com/tharsis/ethermint/issues/207) Replace `big.Int` in favor of `sdk.Int` for `TxData` fields
* (proto, evm) [tharsis#81](https://github.com/tharsis/ethermint/pull/81) gRPC Query and Tx service changes:
  * The `TxReceipt`, `TxReceiptsByBlockHeight` endpoints have been removed from the Query service.
  * The `ContractAddress`, `Bloom` have been removed from the `MsgEthereumTxResponse` and the
    response now contains the ethereum-formatted `Hash` in hex format.
* (eth) [\#845](https://github.com/cosmos/ethermint/pull/845) The `eth` namespace must be included in the list of API's as default to run the rpc server without error.
* (evm) [#202](https://github.com/tharsis/ethermint/pull/202) Web3 api `SendTransaction`/`SendRawTransaction` returns ethereum compatible transaction hash, and query api `GetTransaction*` also accept that.

### Improvements

* (rpc) [tharsis#181](https://github.com/tharsis/ethermint/pull/181) Use evm denomination for params on tx fee.
* (deps) [tharsis#165](https://github.com/tharsis/ethermint/pull/165) Bump Cosmos SDK and Tendermint versions to [v0.42.6](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.42.6) and [v0.34.11](https://github.com/tendermint/tendermint/releases/tag/v0.34.11), respectively.
* (evm) [tharsis#66](https://github.com/tharsis/ethermint/issues/66) Support legacy transaction types for signing.
* (evm) [tharsis#24](https://github.com/tharsis/ethermint/pull/24) Implement metrics for `MsgEthereumTx`, state transitions, `BeginBlock` and `EndBlock`.
* (rpc)  [#124](https://github.com/tharsis/ethermint/issues/124) Implement `txpool_content`, `txpool_inspect` and `txpool_status` RPC methods
* (rpc) [tharsis#112](https://github.com/tharsis/ethermint/pull/153) Fix `eth_coinbase` to return the ethereum address of the validator
* (rpc) [tharsis#176](https://github.com/tharsis/ethermint/issues/176) Support fetching pending nonce

### Bug Fixes

* (rpc) [tharsis#81](https://github.com/tharsis/ethermint/pull/81) Fix transaction hashing and decoding on `eth_sendTransaction`.
* (rpc) [tharsis#45](https://github.com/tharsis/ethermint/pull/45) Use `EmptyUncleHash` and `EmptyRootHash` for empty ethereum `Header` fields.

## [v0.4.1] - 2021-03-01

### API Breaking

* (faucet) [\#678](https://github.com/cosmos/ethermint/pull/678) Faucet module has been removed in favor of client libraries such as [`@cosmjs/faucet`](https://github.com/cosmos/cosmjs/tree/master/packages/faucet).
* (evm) [\#670](https://github.com/cosmos/ethermint/pull/670) Migrate types to the ones defined by the protobuf messages, which are required for the stargate release.

### Bug Fixes

* (evm) [\#799](https://github.com/cosmos/ethermint/issues/799) Fix wrong precision in calculation of gas fee.
* (evm) [\#760](https://github.com/cosmos/ethermint/issues/760) Fix Failed to call function EstimateGas.
* (evm) [\#767](https://github.com/cosmos/ethermint/issues/767) Fix error of timeout when using Truffle to deploy contract.
* (evm) [\#751](https://github.com/cosmos/ethermint/issues/751) Fix misused method to calculate block hash in evm related function.
* (evm) [\#721](https://github.com/cosmos/ethermint/issues/721) Fix mismatch block hash in rpc response when use eht.getBlock.
* (evm) [\#730](https://github.com/cosmos/ethermint/issues/730) Fix 'EIP2028' not open when Istanbul version has been enabled. 
* (evm) [\#749](https://github.com/cosmos/ethermint/issues/749) Fix panic in `AnteHandler` when gas price larger than 100000
* (evm) [\#747](https://github.com/cosmos/ethermint/issues/747) Fix format errors in String() of QueryETHLogs
* (evm) [\#742](https://github.com/cosmos/ethermint/issues/742) Add parameter check for evm query func. 
* (evm) [\#687](https://github.com/cosmos/ethermint/issues/687) Fix nonce check to explicitly check for the correct nonce, rather than a simple 'greater than' comparison. 
* (api) [\#687](https://github.com/cosmos/ethermint/issues/687) Returns error for a transaction with an incorrect nonce. 
* (evm) [\#674](https://github.com/cosmos/ethermint/issues/674) Reset all cache after account data has been committed in `EndBlock` to make sure every node state consistent.
* (evm) [\#672](https://github.com/cosmos/ethermint/issues/672) Fix panic of `wrong Block.Header.AppHash` when restart a node with snapshot.
* (evm) [\#775](https://github.com/cosmos/ethermint/issues/775) MisUse of headHash as blockHash when create EVM context.

### Features
* (api) [\#821](https://github.com/cosmos/ethermint/pull/821) Individually enable the api modules. Will be implemented in the latest version of ethermint with the upcoming stargate upgrade.

### Features
* (api) [\#825](https://github.com/cosmos/ethermint/pull/825) Individually enable the api modules. Will be implemented in the latest version of ethermint with the upcoming stargate upgrade.

## [v0.4.0] - 2020-12-15

### API Breaking

* (evm) [\#661](https://github.com/cosmos/ethermint/pull/661) `Balance` field has been removed from the evm module's `GenesisState`.

### Features

* (rpc) [\#571](https://github.com/cosmos/ethermint/pull/571) Add pending queries to JSON-RPC calls. This allows for the querying of pending transactions and other relevant information that pertains to the pending state:
  * `eth_getBalance`
  * `eth_getTransactionCount`
  * `eth_getBlockTransactionCountByNumber`
  * `eth_getBlockByNumber`
  * `eth_getTransactionByHash`
  * `eth_getTransactionByBlockNumberAndIndex`
  * `eth_sendTransaction` - the nonce will automatically update to its pending nonce (when none is explicitly provided)

### Improvements

* (evm) [\#661](https://github.com/cosmos/ethermint/pull/661) Add invariant check for account balance and account nonce.
* (deps) [\#654](https://github.com/cosmos/ethermint/pull/654) Bump go-ethereum version to [v1.9.25](https://github.com/ethereum/go-ethereum/releases/tag/v1.9.25)
* (evm) [\#627](https://github.com/cosmos/ethermint/issues/627) Add extra EIPs parameter to apply custom EVM jump tables.

### Bug Fixes

* (evm) [\#661](https://github.com/cosmos/ethermint/pull/661) Set nonce to the EVM account on genesis initialization.
* (rpc) [\#648](https://github.com/cosmos/ethermint/issues/648) Fix block cumulative gas used value.
* (evm) [\#621](https://github.com/cosmos/ethermint/issues/621) EVM `GenesisAccount` fields now share the same format as the auth module `Account`.
* (evm) [\#618](https://github.com/cosmos/ethermint/issues/618) Add missing EVM `Context` `GetHash` field that retrieves a the header hash from a given block height.
* (app) [\#617](https://github.com/cosmos/ethermint/issues/617) Fix genesis export functionality.
* (rpc) [\#574](https://github.com/cosmos/ethermint/issues/574) Fix outdated version from `eth_protocolVersion`.

## [v0.3.1] - 2020-11-24

### Improvements

* (deps) [\#615](https://github.com/cosmos/ethermint/pull/615) Bump Cosmos SDK version to [v0.39.2](https://github.com/cosmos/cosmos-sdk/tag/v0.39.2)
* (deps) [\#610](https://github.com/cosmos/ethermint/pull/610) Update Go dependency to 1.15+.
* (evm) [#603](https://github.com/cosmos/ethermint/pull/603) Add state transition params that enable or disable the EVM `Call` and `Create` operations.
* (deps) [\#602](https://github.com/cosmos/ethermint/pull/602) Bump tendermint version to [v0.33.9](https://github.com/tendermint/tendermint/releases/tag/v0.33.9)

### Bug Fixes

* (rpc) [\#613](https://github.com/cosmos/ethermint/issues/613) Fix potential deadlock caused if the keyring `List` returned an error.

## [v0.3.0] - 2020-11-16

### API Breaking

* (crypto) [\#559](https://github.com/cosmos/ethermint/pull/559) Refactored crypto package in preparation for the SDK's Stargate release:
  * `crypto.PubKeySecp256k1` and `crypto.PrivKeySecp256k1` are now `ethsecp256k1.PubKey` and `ethsecp256k1.PrivKey`, respectively
  * Moved SDK `SigningAlgo` implementation for Ethermint's Secp256k1 key to `crypto/hd` package.
* (rpc) [\#588](https://github.com/cosmos/ethermint/pull/588) The `rpc` package has been refactored to account for the separation of each
corresponding Ethereum API namespace:
  * `rpc/namespaces/eth`: `eth` namespace. Exposes the `PublicEthereumAPI` and the `PublicFilterAPI`.
  * `rpc/namespaces/personal`: `personal` namespace. Exposes the `PrivateAccountAPI`.
  * `rpc/namespaces/net`: `net` namespace. Exposes the `PublicNetAPI`.
  * `rpc/namespaces/web3`: `web3` namespace. Exposes the `PublicWeb3API`.
* (evm) [\#588](https://github.com/cosmos/ethermint/pull/588) The EVM transaction CLI has been removed in favor of the JSON-RPC.

### Improvements

* (deps) [\#594](https://github.com/cosmos/ethermint/pull/594) Bump go-ethereum version to [v1.9.24](https://github.com/ethereum/go-ethereum/releases/tag/v1.9.24)

### Bug Fixes

* (ante) [\#597](https://github.com/cosmos/ethermint/pull/597) Fix incorrect fee check on `AnteHandler`.
* (evm) [\#583](https://github.com/cosmos/ethermint/pull/583) Fixes incorrect resetting of tx count and block bloom during `BeginBlock`, as well as gas consumption.
* (crypto) [\#577](https://github.com/cosmos/ethermint/pull/577) Fix `BIP44HDPath` that did not prepend `m/` to the path. This now uses the `DefaultBaseDerivationPath` variable from go-ethereum to ensure addresses are consistent.

## [v0.2.1] - 2020-09-30

### Features

* (rpc) [\#552](https://github.com/cosmos/ethermint/pull/552) Implement Eth Personal namespace `personal_importRawKey`.

### Bug fixes

* (keys) [\#554](https://github.com/cosmos/ethermint/pull/554) Fix private key derivation.
* (app/ante) [\#550](https://github.com/cosmos/ethermint/pull/550) Update ante handler nonce verification to accept any nonce greater than or equal to the expected nonce to allow to successive transactions.

## [v0.2.0] - 2020-09-24

### State Machine Breaking

* (app) [\#540](https://github.com/cosmos/ethermint/issues/540) Chain identifier's format has been changed to match the Cosmos `chainID` [standard](https://github.com/ChainAgnostic/CAIPs/blob/master/CAIPs/caip-5.md), which is required for IBC. The epoch number of the ID is used as the EVM `chainID`.

### API Breaking

* (types) [\#503](https://github.com/cosmos/ethermint/pull/503) The `types.DenomDefault` constant for `"aphoton"` has been renamed to `types.AttoPhoton`.

### Improvements

* (types) [\#504](https://github.com/cosmos/ethermint/pull/504) Unmarshal a JSON `EthAccount` using an Ethereum hex address in addition to Bech32.
* (types) [\#503](https://github.com/cosmos/ethermint/pull/503) Add `--coin-denom` flag to testnet command that sets the given coin denomination to SDK and Ethermint parameters.
* (types) [\#502](https://github.com/cosmos/ethermint/pull/502) `EthAccount` now also exposes the Ethereum hex address in `string` format to clients.
* (types) [\#494](https://github.com/cosmos/ethermint/pull/494) Update `EthAccount` public key JSON type to `string`.
* (app) [\#471](https://github.com/cosmos/ethermint/pull/471) Add `x/upgrade` module for managing software updates.
* (`x/evm`) [\#458](https://github.com/cosmos/ethermint/pull/458) Define parameter for token denomination used for the EVM module.
* (`x/evm`) [\#443](https://github.com/cosmos/ethermint/issues/443) Support custom Ethereum `ChainConfig` params.
* (types) [\#434](https://github.com/cosmos/ethermint/issues/434) Update default denomination to Atto Photon (`aphoton`).
* (types) [\#515](https://github.com/cosmos/ethermint/pull/515) Update minimum gas price to be 1.

### Bug Fixes

* (ante) [\#525](https://github.com/cosmos/ethermint/pull/525) Add message validation decorator to `AnteHandler` for `MsgEthereumTx`.
* (types) [\#507](https://github.com/cosmos/ethermint/pull/507) Fix hardcoded `aphoton` on `EthAccount` balance getter and setter.
* (types) [\#501](https://github.com/cosmos/ethermint/pull/501) Fix bech32 encoding error by using the compressed ethereum secp256k1 public key.
* (`x/evm`) [\#496](https://github.com/cosmos/ethermint/pull/496) Fix bugs on `journal.revert` and `CommitStateDB.Copy`.
* (types) [\#480](https://github.com/cosmos/ethermint/pull/480) Update [BIP44](https://github.com/bitcoin/bips/blob/master/bip-0044.mediawiki) coin type to `60` to satisfy [EIP84](https://github.com/ethereum/EIPs/issues/84).
* (types) [\#513](https://github.com/cosmos/ethermint/pull/513) Fix simulated transaction bug that was causing a consensus error by unintentionally affecting the state.

## [v0.1.0] - 2020-08-23

### Improvements

* (sdk) [\#386](https://github.com/cosmos/ethermint/pull/386) Bump Cosmos SDK version to [v0.39.1](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.39.1)
* (`x/evm`) [\#181](https://github.com/cosmos/ethermint/issues/181) Updated EVM module to the recommended module structure.
* (app) [\#188](https://github.com/cosmos/ethermint/issues/186)  Misc cleanup:
  * (`x/evm`) Rename `EthereumTxMsg` --> `MsgEthereumTx` and `EmintMsg` --> `MsgEthermint` for consistency with SDK standards
  * Updated integration and unit tests to use `EthermintApp` as testing suite
  * Use expected `Keeper` interface for `AccountKeeper`
  * Replaced `count` type in keeper with `int`
  * Add SDK events for transactions
* [\#236](https://github.com/cosmos/ethermint/pull/236) Changes from upgrade:
  * (`app/ante`) Moved `AnteHandler` implementation to `app/ante`
  * (keys) Marked `ExportEthKeyCommand` as **UNSAFE**
  * (`x/evm`) Moved `BeginBlock` and `EndBlock` to `x/evm/abci.go`
* (`x/evm`) [\#255](https://github.com/cosmos/ethermint/pull/255) Add missing `GenesisState` fields and support `ExportGenesis` functionality.
* [\#272](https://github.com/cosmos/ethermint/pull/272) Add `Logger` for evm module.
* [\#317](https://github.com/cosmos/ethermint/pull/317) `GenesisAccount` validation.
* (`x/evm`) [\#319](https://github.com/cosmos/ethermint/pull/319) Various evm improvements:
  * Add transaction `[]*ethtypes.Logs` to evm's `GenesisState` to persist logs after an upgrade.
  * Remove evm `CodeKey` and `BlockKey`in favor of a prefix `Store`.
  * Set `BlockBloom` during `EndBlock` instead of `BeginBlock`.
  * `Commit` state object and `Finalize` storage after `InitGenesis` setup.
* (rpc) [\#325](https://github.com/cosmos/ethermint/pull/325) `eth_coinbase` JSON-RPC query now returns the node's validator address.

### Features

* (build) [\#378](https://github.com/cosmos/ethermint/pull/378) Create multi-node, local, automated testnet setup with `make localnet-start`.
* (rpc) [\#330](https://github.com/cosmos/ethermint/issues/330) Implement `PublicFilterAPI`'s `EventSystem` which subscribes to Tendermint events upon `Filter` creation.
* (rpc) [\#231](https://github.com/cosmos/ethermint/issues/231) Implement `NewBlockFilter` in rpc/filters.go which instantiates a polling block filter
  * Polls for new blocks via `BlockNumber` rpc call; if block number changes, it requests the new block via `GetBlockByNumber` rpc call and adds it to its internal list of blocks
  * Update `uninstallFilter` and `getFilterChanges` accordingly
  * `uninstallFilter` stops the polling goroutine
  * `getFilterChanges` returns the filter's internal list of block hashes and resets it
* (rpc) [\#54](https://github.com/cosmos/ethermint/issues/54), [\#55](https://github.com/cosmos/ethermint/issues/55)
  Implement `eth_getFilterLogs` and `eth_getLogs`:
  * For a given filter, look through each block for transactions. If there are transactions in the block, get the logs from it, and filter using the filterLogs method
  * `eth_getLogs` and `eth_getFilterChanges` for log filters use the same underlying method as `eth_getFilterLogs`
  * update `HandleMsgEthereumTx` to store logs using the ethereum hash
* (app) [\#187](https://github.com/cosmos/ethermint/issues/187) Add support for simulations.

### Bug Fixes

* (evm) [\#767](https://github.com/cosmos/ethermint/issues/767) Fix error of timeout when using Truffle to deploy contract.
* (evm) [\#751](https://github.com/cosmos/ethermint/issues/751) Fix misused method to calculate block hash in evm related function.
* (evm) [\#721](https://github.com/cosmos/ethermint/issues/721) Fix mismatch block hash in rpc response when use eth.getBlock.
* (evm) [\#730](https://github.com/cosmos/ethermint/issues/730) Fix 'EIP2028' not open when Istanbul version has been enabled.
* (app) [\#749](https://github.com/cosmos/ethermint/issues/749) Fix panic in `AnteHandler` when gas price larger than 100000
* (rpc) [\#305](https://github.com/cosmos/ethermint/issues/305) Update `eth_getTransactionCount` to check for account existence before getting sequence and return 0 as the nonce if it doesn't exist.
* (`x/evm`) [\#319](https://github.com/cosmos/ethermint/pull/319) Fix `SetBlockHash` that was setting the incorrect height during `BeginBlock`.
* (`x/evm`) [\#176](https://github.com/cosmos/ethermint/issues/176) Updated Web3 transaction hash from using RLP hash. Now all transaction hashes exposed are amino hashes:
  * Removes `Hash()` (RLP) function from `MsgEthereumTx` to avoid confusion or misuse in future.
