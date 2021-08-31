# ADR 002: EVM Hooks

## Changelog

- 2021-08-11: first draft

## Status

PROPOSED

## Abstract

The current ADR proposes a hook interface to EVM module, with the goal of extending the tx processing logic externally,
specifically to support EVM contract calling native modules through logs.

## Context

Currently there are no way for EVM smart contracts to call cosmos native modules, one way to do this is by emitting
specific logs from contract, and recognize those logs in tx processing code and convert them to native module calls.

To do this in an extensible way, we can add a post tx processing hook into the evm module, which allows third party to
add custom logic to process transaction logs.

## Decision

This ADR propose to add an `EvmHooks` interface and a method to register hooks in the `EvmKeeper`:

```golang
type EvmHooks interface {
	  PostTxProcessing(ctx sdk.Context, txHash ethcmn.Hash, logs []*ethtypes.Log) error
}

func (k *EvmKeeper) SetHooks(eh types.EvmHooks) *Keeper;
```

- `PostTxProcessing` is only called after EVM transaction finished succesfully, it's executed in the same cache context
  as the EVM transaction, if it returns an error, the whole EVM transaction is reverted, if the hook implementor don't
  want to revert the tx, he can always return a nil instead.

  Basically the contract sends an asynchronous message to native modules, there's no way for it to catch and recover the error.

There are no default hooks implemented in evm module, so the proposal is totally backward compatible, only opens extra
extensibility for certain use cases.

### Use Case: Call Native Module

To support contract calling native module with this proposal, one can define a log signature, and emits the specific log
from smart contract, native logic registers a `PostTxProcessing` hook which recognize the log and do the native module
call.

For example, to support smart contract to transfer native tokens, one can define and emit a `__CosmosNativeBankSend` log
signature in smart contract like this:

```solidity
event __CosmosNativeBankSend(address recipient, uint256 amount, string denom);

function withdraw_to_native_token(amount uint256) public {
    _balances[msg.sender] -= amount;
    // send native tokens from contract address to msg.sender.
    emit __CosmosNativeBankSend(msg.sender, amount, "native_denom");
}
```

And the application registers a `BankSendHook` to `EvmKeeper`, it recognize the log and convert it to a call to the bank
module's `SendCoinsFromAccountToAccount` method:

```golang
var (
	// BankSendEvent represent the signature of
	// `event __CosmosNativeBankSend(address recipient, uint256 amount, string denom)`
	BankSendEvent abi.Event
)

func init() {
	addressType, _ := abi.NewType("address", "", nil)
	uint256Type, _ := abi.NewType("uint256", "", nil)
	stringType, _ := abi.NewType("string", "", nil)
	BankSendEvent = abi.NewEvent(
		"__CosmosNativeBankSend",
		"__CosmosNativeBankSend",
		false,
		abi.Arguments{abi.Argument{
			Name:    "recipient",
			Type:    addressType,
			Indexed: false,
		}, abi.Argument{
			Name:    "amount",
			Type:    uint256Type,
			Indexed: false,
		}, abi.Argument{
			Name:    "denom",
			Type:    stringType,
			Indexed: false,
		}},
	)
}

type BankSendHook struct {
	bankKeeper bankkeeper.Keeper
}

func NewBankSendHook(bankKeeper bankkeeper.Keeper) *BankSendHook {
	return &BankSendHook{
		bankKeeper: bankKeeper,
	}
}

func (h BankSendHook) PostTxProcessing(ctx sdk.Context, txHash ethcmn.Hash, logs []*ethtypes.Log) error {
	for _, log := range logs {
		if len(log.Topics) > 0 && log.Topics[0] == BankSendEvent.ID {
			unpacked, err := BankSendEvent.Inputs.Unpack(log.Data)
			if err != nil {
				// ignore unrecognizable log
				continue
			}
			contract := sdk.AccAddress(log.Address.Bytes())
			recipient := sdk.AccAddress(unpacked[0].(ethcmn.Address).Bytes())
			coins := sdk.NewCoins(sdk.NewCoin(unpacked[2].(string), sdk.NewIntFromBigInt(unpacked[1].(*big.Int))))
			err = h.bankKeeper.SendCoins(ctx, contract, recipient, coins)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
```

Register the hook in `app.go`:

```golang
evmKeeper.SetHooks(NewBankSendHook(bankKeeper));
```

## Consequences

### Backwards Compatibility

The proposed ADR is backward compatible.

### Positive

- Improve extensibility of evm module

### Negative

None
