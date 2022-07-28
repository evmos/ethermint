# ADR 003: Stateful Precompiled Contracts

## Changelog

- 2022-06-16: first draft

## Status

DRAFT

## Abstract

Support stateful precompiled contracts to improve inter-operabilities between EVM smart contracts and native functionalities.

## Context

We need inter-operabilities to allow EVM smart contracts to access native functionalities, like manage native tokens through bank module, send/receive IBC messages through IBC modules.

The EVM hooks solution is not ideal, because the message processing is asynchronously, so the caller can't get the return value.

Precompiled contract can provide a much better interface for smart contract developers. But the default precompiled contract implementation in go-ethereum is stateless, we need to do some patches to support stateful ones properly.

## Decision

### Interface Changes

Change the `PrecompiledContract` interface like this (need to patch `go-ethereum`):

```
 type PrecompiledContract interface {
  	 RequiredGas(input []byte) uint64
-  	 Run(input []byte) ([]byte, error)
+  	 Run(evm *vm.EVM, input []byte, caller common.Address, value *big.Int, readonly bool) ([]byte, error)
 }
```

There are extra parameters passed to the precompiled contract:

- `caller`: aka. `msg.sender`.
- `value`: aka. `msg.value`.
- `readonly`: it's set to `true` for both `staticcall` and `delegatecall`, in these call types, the callee contract is not supposed to modify states. A stateful contract normally should just fail if it's `true`.

### Snapshot and Revert

To implement a stateful precompiled contract, one should be aware of the semantics of `StateDB` itself, basically it keeps all the state writes in memory and maintains a list of journal logs for the write operations, and it supports snapshot and revert by undoing the journal logs backward to a certain point in history. The dirty states are written into cosmos-sdk storage when commit at the end of the tx execution.

The precompiled contract must not write to cosmos-sdk storage directly, because the side effects can't be reverted by `StateDB` when exception happens. You should always maintain the dirty states in memory, and append a journal entry to `StateDB` for each modification which can undo it when called.

When reading from cosmos-sdk storage, you are actually reading the committed states, you need to read the in memory caches for the dirty states, for example the accounts and EVM contract storage are cached in `StateDB` itself, and different precompiled contracts may cache different native states.

It's also tricky if not impossible to let two precompiled contracts to write to the same piece of native states, because their in-memory states would be in conflicts.

### Example

```golang
// ExtState manage in memory dirty states which are committed together with StateDB
type ExtState interface {
	Commit(sdk.Context) error
}

// ExtStateDB expose `AppendJournalEntry` api on top of `vm.StateDB` interface
type ExtStateDB interface {
	AppendJournalEntry(statedb.JournalEntry)
}

// BankContract expose native token functionalities to EVM smart contract
type BankContract struct {
	ctx        sdk.Context
	bankKeeper types.BankKeeper
	balances   map[common.Address]map[common.Address]*Balance
}

func (bc *BankContract) RequiredGas(input []byte) uint64 {
	// TODO estimate required gas
	return 0
}

func (bc *BankContract) Run(evm *vm.EVM, input []byte, caller common.Address, value *big.Int, readonly bool) ([]byte, error) {
	stateDB, ok := evm.StateDB.(ExtStateDB)
	if !ok {
		return nil, errors.New("not run in ethermint")
	}

	// parse input
	methodID := input[:4]
	if bytes.Equal(methodID, MintMethod.ID) {
		if readonly {
			return nil, errors.New("the method is not readonly")
		}
		args, err := MintMethod.Inputs.Unpack(input[4:])
		if err != nil {
			return nil, errors.New("fail to unpack input arguments")
		}
		recipient := args[0].(common.Address)
		amount := args[1].(*big.Int)
		if amount.Sign() <= 0 {
			return nil, errors.New("invalid amount")
		}

		if _, ok := bc.balances[caller]; !ok {
			bc.balances[caller] = make(map[common.Address]*Balance)
		}
		balances := bc.balances[caller]
		if balance, ok := balances[recipient]; ok {
			balance.DirtyAmount = new(big.Int).Add(balance.DirtyAmount, amount)
		} else {
			// query original amount
			addr := sdk.AccAddress(recipient.Bytes())
			originAmount := bc.bankKeeper.GetBalance(bc.ctx, addr, EVMDenom(caller)).Amount.BigInt()
			dirtyAmount := new(big.Int).Add(originAmount, amount)
			balances[recipient] = &Balance{
				OriginAmount: originAmount,
				DirtyAmount:  dirtyAmount,
			}
		}
		stateDB.AppendJournalEntry(bankMintChange{bc: bc, caller: caller, recipient: recipient, amount: amount})
  } else if bytes.Equal(methodID, BalanceOfMethod.ID) {
		args, err := BalanceOfMethod.Inputs.Unpack(input[4:])
		if err != nil {
			return nil, errors.New("fail to unpack input arguments")
		}
		token := args[0].(common.Address)
		addr := args[1].(common.Address)
		if balances, ok := bc.balances[token]; ok {
			if balance, ok := balances[addr]; ok {
				return BalanceOfMethod.Outputs.Pack(balance.DirtyAmount)
			}
		}
		// query from storage
		amount := bc.bankKeeper.GetBalance(bc.ctx, sdk.AccAddress(addr.Bytes()), EVMDenom(token)).Amount.BigInt()
		return BalanceOfMethod.Outputs.Pack(amount)
	} else {
		return nil, errors.New("unknown method")
	}
  return nil, nil
}

func (bc *BankContract) Commit(ctx sdk.Context) error {
  // Write the dirty balances through bc.bankKeeper
}

type bankMintChange struct {
	bc        *BankContract
	caller    common.Address
	recipient common.Address
	amount    *big.Int
}

func (ch bankMintChange) Revert(*statedb.StateDB) {
	balance := ch.bc.balances[ch.caller][ch.recipient]
	balance.DirtyAmount = new(big.Int).Sub(balance.DirtyAmount, ch.amount)
}

func (ch bankMintChange) Dirtied() *common.Address {
	return nil
}
```

## Consequences

> This section describes the resulting context, after applying the decision. All consequences should be listed here, not just the "positive" ones. A particular decision may have positive, negative, and neutral consequences, but all of them affect the team and project in the future.

### Backwards Compatibility

- State machine breaking
- Need to patch `go-ethereum`

### Positive

- Better interface for interaction between EVM contract and native functionalities.

### Negative

- Need to patch `go-ethereum`.

### Neutral

- Precompiled contract implementation need to be careful with the in memory dirty states to maintain the invariants.

## Further Discussions

## Test Cases

- Check the state is persisted after tx committed.
- Check exception revert works.
- Check multiple precompiled contracts don't intervene each other.
- Check static call and delegate call don't mutate states.

## References

- [ADR-001 EVM Hooks](adr-002-evm-hooks.md)
