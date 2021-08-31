# ADR 002: EVM Hooks

## Changelog

- 2021-08-11: first draft

## Status

PROPOSED

## Abstract

The current ADR proposes a hook interface to EVM module, with the goal of extending the tx processing logic externally,
specifically to support EVM contract calling native modules through logs.

## Context

<!-- > This section describes the forces at play, including technological, political, social, and project local. These forces are probably in tension, and should be called out as such. The language in this section is value-neutral. It is simply describing facts. It should clearly explain the problem and motivation that the proposal aims to resolve. -->

Currently there are no way for EVM smart contracts to call cosmos native modules, one way to do this is by emitting
specific logs from contract, and recognize those logs in tx processing code and convert them to native module calls.

To do this in an extensible way, we can add a post tx processing hook into the evm module, which allows third party to
add custom logic to process transaction logs.

## Decision

<!-- > This section describes our response to these forces. It is stated in full sentences, with active voice. "We will ..." -->

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

  The errors returned by hooks are translated into a vm error `failed to process native logs`, the detailed error
  message is stored in the return value.

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

<!-- > This section describes the resulting context, after applying the decision. All consequences should be listed here, not just the "positive" ones. A particular decision may have positive, negative, and neutral consequences, but all of them affect the team and project in the future. -->

### Backwards Compatibility

<!-- All ADRs that introduce backwards incompatibilities must include a section describing these incompatibilities and their severity. The ADR must explain how the author proposes to deal with these incompatibilities. ADR submissions without a sufficient backwards compatibility treatise may be rejected outright. -->

The proposed ADR is backward compatible.

### Positive

- Improve extensibility of evm module

### Negative

None

### Neutral

None

## Further Discussions

<!-- While an ADR is in the DRAFT or PROPOSED stage, this section should contain a summary of issues to be solved in future iterations (usually referencing comments from a pull-request discussion).
Later, this section can optionally list ideas or improvements the author or reviewers found during the analysis of this ADR. -->

## Test Cases [optional]

<!-- Test cases for an implementation are mandatory for ADRs that are affecting consensus changes. Other ADRs can choose to include links to test cases if applicable. -->

## References

<!-- - {reference link} -->
