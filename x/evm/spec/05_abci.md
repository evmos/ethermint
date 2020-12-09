<!--
order: 5
-->

# ABCI

## InitGenesis

`InitGenesis` initializes the EVM module genesis state by setting the `GenesisState` fields to the
store. In particular it sets the parameters, configuration, accounts and transaction logs.

The function also performs the invariant that the EVM balance  from the `GenesisAccount` matches the
balance amount from the `EthAccount` as defined on the `auth` module.

## ExportGenesis

The `ExportGenesis` ABCI function exports the genesis state of the EVM module. In particular, it
retrieves all the accounts with their bytecode, balance and storage, the transaction logs, and the
EVM parameters and chain configuration.

## BeginBlock

The EVM module `BeginBlock` logic is executed prior to handling the state transitions from the
transactions. The main objective of this function is to:

* Set the block header hash mappings to the module state (`hash -> height` and `height -> hash`).
  This workaround is due to the fact that until the `v0.34.0` Tendermint version it wasn't possible
  to query and subscribe to a block by hash.

* Reset bloom filter and block transaction count. These variables, which are fields of the EVM
  `Keeper`, are updated on every EVM transaction.

## EndBlock

The EVM module `EndBlock` logic occurs after executing all the state transitions from the
transactions. The main objective of this function is to:

* Update the accounts. This operation retrieves the current account and balance values for each
  state object and updates the account represented on the stateObject with the given values. This is
  done since the account might have been updated by transactions other than the ones defined by the
  `x/evm` module, such as bank send or IBC transfers.
* Commit dirty state objects and delete empty ones from the store. This operation writes the
  contract code to the key value store in the case of contracts and updates the account's balance,
  which is set to the the bank module's `Keeper`.
* Clear account cache. This clears cache of state objects to handle account changes outside of the
  EVM.
* Store the block bloom to state. This is due for Web3 compatibility as the Ethereum headers contain
  this type as a  field. The Ethermint RPC uses this query to construct an Ethereum Header from a
  Tendermint Header.
