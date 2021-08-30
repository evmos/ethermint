# ADR 002: EVM Hooks

## Changelog

- 2021-08-11: first draft

## Status

PROPOSED

## Abstract

The current ADR proposes a hook interface to EVM module, with the goal of extend the functionality externally,
specifically to support communication from EVM contract to native modules through logs.

## Context

There are some requirements to enable EVM contract to communicate with cosmos native modules, which could have multiple
use cases. We need some kind of hooks to support these extension requirements in a generic way.

## Decision

This ADR propose to add `EvmHooks` interface:

```golang
type EvmHooks interface {
	  PostTxProcessing(ctx sdk.Context, txHash ethcmn.Hash, logs []*ethtypes.Log) error
}
```

- `PostTxProcessing` is called after EVM transaction finished, executed with the same cache context as the EVM
  transaction execution, if `PostTxProcessing` returns an error, the whole EVM transaction is reverted.

  `PostTxProcessing` can be used to allow evm contract to call native module functionalities through logs, for example,
a `BankSendHook` could implement the hook to convert a specific log and convert it to a call to the bank module's
`SendCoinsFromAccountToAccount` method, so a contract could emit that specific log to transfer native tokens, like this:

  ```solidity
  function withdraw_to_native_token(amount uint256, eth_dest address) public {
      _balances[msg.sender] -= amount;
      // send native tokens from contract address to msg.sender.
      emit __CosmosNativeBankSend(msg.sender, amount, "native_denom");
  }
  ```

There are no hooks implemented in the default application, but other applications could implement custom hooks and
register them to the `EvmKeeper`, for example:

```golang
evmKeeper.SetHooks(NewHook());
```

