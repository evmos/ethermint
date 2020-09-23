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

### API Breaking

* (types) [\#503](https://github.com/ChainSafe/ethermint/pull/503) The `types.DenomDefault` constant for `"aphoton"` has been renamed to `types.AttoPhoton`.

### Improvements

* (types) [\#504](https://github.com/ChainSafe/ethermint/pull/504) Unmarshal a JSON `EthAccount` using an Ethereum hex address in addition to Bech32.
* (types) [\#503](https://github.com/ChainSafe/ethermint/pull/503) Add `--coin-denom` flag to testnet command that sets the given coin denomination to SDK and Ethermint parameters.
* (types) [\#502](https://github.com/ChainSafe/ethermint/pull/502) `EthAccount` now also exposes the Ethereum hex address in `string` format to clients.
* (types) [\#494](https://github.com/ChainSafe/ethermint/pull/494) Update `EthAccount` public key JSON type to `string`.
* (app) [\#471](https://github.com/ChainSafe/ethermint/pull/471) Add `x/upgrade` module for managing software updates.
* (`x/evm`) [\#458](https://github.com/ChainSafe/ethermint/pull/458) Define parameter for token denomination used for the EVM module.
* (`x/evm`) [\#443](https://github.com/ChainSafe/ethermint/issues/443) Support custom Ethereum `ChainConfig` params.
* (types) [\#434](https://github.com/ChainSafe/ethermint/issues/434) Update default denomination to Atto Photon (`aphoton`).
* (types) [\#515](https://github.com/ChainSafe/ethermint/pull/515) Update minimum gas price to be 1.

### Bug Fixes

* (ante) [\#525](https://github.com/ChainSafe/ethermint/pull/525) Add message validation decorator to `AnteHandler` for `MsgEthereumTx`.
* (types) [\#507](https://github.com/ChainSafe/ethermint/pull/507) Fix hardcoded `aphoton` on `EthAccount` balance getter and setter.
* (`x/evm`) [\#496](https://github.com/ChainSafe/ethermint/pull/496) Fix bugs on `journal.revert` and `CommitStateDB.Copy`.
* (types) [\#480](https://github.com/ChainSafe/ethermint/pull/480) Update [BIP44](https://github.com/bitcoin/bips/blob/master/bip-0044.mediawiki) coin type to `60` to satisfy [EIP84](https://github.com/ethereum/EIPs/issues/84).
* (types) [\#513](https://github.com/ChainSafe/ethermint/pull/513) Fix simulated transaction bug that was causing a consensus error by unintentionally affecting the state.

## [v0.1.0] - 2020-08-23

### Improvements

* (sdk) [\#386](https://github.com/ChainSafe/ethermint/pull/386) Bump Cosmos SDK version to [v0.39.1](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.39.1)
* (`x/evm`) [\#181](https://github.com/ChainSafe/ethermint/issues/181) Updated EVM module to the recommended module structure.
* (app) [\#188](https://github.com/ChainSafe/ethermint/issues/186)  Misc cleanup:
  * (`x/evm`) Rename `EthereumTxMsg` --> `MsgEthereumTx` and `EmintMsg` --> `MsgEthermint` for consistency with SDK standards
  * Updated integration and unit tests to use `EthermintApp` as testing suite
  * Use expected `Keeper` interface for `AccountKeeper`
  * Replaced `count` type in keeper with `int`
  * Add SDK events for transactions
* [\#236](https://github.com/ChainSafe/ethermint/pull/236) Changes from upgrade:
  * (`app/ante`) Moved `AnteHandler` implementation to `app/ante`
  * (keys) Marked `ExportEthKeyCommand` as **UNSAFE**
  * (`x/evm`) Moved `BeginBlock` and `EndBlock` to `x/evm/abci.go`
* (`x/evm`) [\#255](https://github.com/ChainSafe/ethermint/pull/255) Add missing `GenesisState` fields and support `ExportGenesis` functionality.
* [\#272](https://github.com/ChainSafe/ethermint/pull/272) Add `Logger` for evm module.
* [\#317](https://github.com/ChainSafe/ethermint/pull/317) `GenesisAccount` validation.
* (`x/evm`) [\#319](https://github.com/ChainSafe/ethermint/pull/319) Various evm improvements:
  * Add transaction `[]*ethtypes.Logs` to evm's `GenesisState` to persist logs after an upgrade.
  * Remove evm `CodeKey` and `BlockKey`in favor of a prefix `Store`.
  * Set `BlockBloom` during `EndBlock` instead of `BeginBlock`.
  * `Commit` state object and `Finalize` storage after `InitGenesis` setup.
* (rpc) [\#325](https://github.com/ChainSafe/ethermint/pull/325) `eth_coinbase` JSON-RPC query now returns the node's validator address.

### Features

* (build) [\#378](https://github.com/ChainSafe/ethermint/pull/378) Create multi-node, local, automated testnet setup with `make localnet-start`.
* (rpc) [\#330](https://github.com/ChainSafe/ethermint/issues/330) Implement `PublicFilterAPI`'s `EventSystem` which subscribes to Tendermint events upon `Filter` creation.
* (rpc) [\#231](https://github.com/ChainSafe/ethermint/issues/231) Implement `NewBlockFilter` in rpc/filters.go which instantiates a polling block filter
  * Polls for new blocks via `BlockNumber` rpc call; if block number changes, it requests the new block via `GetBlockByNumber` rpc call and adds it to its internal list of blocks
  * Update `uninstallFilter` and `getFilterChanges` accordingly
  * `uninstallFilter` stops the polling goroutine
  * `getFilterChanges` returns the filter's internal list of block hashes and resets it
* (rpc) [\#54](https://github.com/ChainSafe/ethermint/issues/54), [\#55](https://github.com/ChainSafe/ethermint/issues/55)
  Implement `eth_getFilterLogs` and `eth_getLogs`:
  * For a given filter, look through each block for transactions. If there are transactions in the block, get the logs from it, and filter using the filterLogs method
  * `eth_getLogs` and `eth_getFilterChanges` for log filters use the same underlying method as `eth_getFilterLogs`
  * update `HandleMsgEthereumTx` to store logs using the ethereum hash
* (app) [\#187](https://github.com/ChainSafe/ethermint/issues/187) Add support for simulations.

### Bug Fixes

* (rpc) [\#305](https://github.com/ChainSafe/ethermint/issues/305) Update `eth_getTransactionCount` to check for account existence before getting sequence and return 0 as the nonce if it doesn't exist.
* (`x/evm`) [\#319](https://github.com/ChainSafe/ethermint/pull/319) Fix `SetBlockHash` that was setting the incorrect height during `BeginBlock`.
* (`x/evm`) [\#176](https://github.com/ChainSafe/ethermint/issues/176) Updated Web3 transaction hash from using RLP hash. Now all transaction hashes exposed are amino hashes:
  * Removes `Hash()` (RLP) function from `MsgEthereumTx` to avoid confusion or misuse in future.
