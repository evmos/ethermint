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

## [Unreleased]

### Improvements

* (sdk) [\#171](https://github.com/ChainSafe/ethermint/issues/177) Bump Cosmos SDK version to [v0.38.1](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.38.1) [@fedekunze](https://github.com/fedekunze):
  * Add `x/evidence` module to ethermint app
  * Bump Go requirement to 1.14+
* (`x/evm`) [\#181](https://github.com/ChainSafe/ethermint/issues/181) Updated EVM module to the recommended module structure. [@fedekunze](https://github.com/fedekunze)
* (app) [\#188](https://github.com/ChainSafe/ethermint/issues/186)  Misc cleanup [@fedekunze](https://github.com/fedekunze):
  * (`x/evm`) Rename `EthereumTxMsg` --> `MsgEthereumTx` and `EmintMsg` --> `MsgEthermint` for consistency with SDK standards
  * Updated integration and unit tests to use `EthermintApp` as testing suite
  * Use expected keeper interface for `AccountKeeper`
  * Replaced `count` type in keeper with `int`
  * Add SDK events for transactions
* [\#236](https://github.com/ChainSafe/ethermint/pull/236) Changes from upgrade [@fedekunze](https://github.com/fedekunze)
  * (`app/ante`) Moved `AnteHandler` implementation to `app/ante`
  * (keys) Marked `ExportEthKeyCommand` as **UNSAFE**
  * (`x/evm`) Moved `BeginBlock` and `EndBlock` to `x/evm/abci.go`

### Features

* (rpc) [\#231](https://github.com/ChainSafe/ethermint/issues/231) Implement NewBlockFilter in rpc/filters.go which instantiates a polling block filter
  * Polls for new blocks via BlockNumber rpc call; if block number changes, it requests the new block via GetBlockByNumber rpc call and adds it to its internal list of blocks
  * Update uninstallFilter and getFilterChanges accordingly
  * uninstallFilter stops the polling goroutine
  * getFilterChanges returns the filter's internal list of block hashes and resets it

* (rpc) [\#54](https://github.com/ChainSafe/ethermint/issues/54) [\#55](https://github.com/ChainSafe/ethermint/issues/55)
  Implement eth_getFilterLogs and eth_getLogs
  * for a given filter, look through each block for transactions. If there are transactions in the block, get the logs from it, and filter using the filterLogs method
  * eth_getLogs and eth_getFilterChanges for log filters use the same underlying method as eth_getFilterLogs
  * update HandleMsgEthereumTx to store logs using the ethereum hash

### Bug Fixes

* (x/evm) [\#176](https://github.com/ChainSafe/ethermint/issues/176) Updated Web3 transaction hash from using RLP hash. Now all transaction hashes exposed are amino hashes.
  * Removes `Hash()` (RLP) function from `MsgEthereumTx` to avoid confusion or misuse in future.
