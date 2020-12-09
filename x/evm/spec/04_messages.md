<!--
order: 4
-->

# Messages

## MsgEthereumTx

An EVM state transition can be achieved by using the `MsgEthereumTx`. This message encapsulates an
Ethereum transaction as an SDK message and contains the necessary transaction data fields.

One remark about the `MsgEthereumTx` is that it implements both the [`sdk.Msg`](https://github.com/cosmos/cosmos-sdk/blob/v0.39.2/types/tx_msg.go#L7-L29) and [`sdk.Tx`](https://github.com/cosmos/cosmos-sdk/blob/v0.39.2/types/tx_msg.go#L33-L41)
interfaces (generally SDK messages only implement the former, while the latter is a group of
messages bundled together). The reason of this, is because the `MsgEthereumTx` must not be included in a [`auth.StdTx`](https://github.com/cosmos/cosmos-sdk/blob/v0.39.2/x/auth/types/stdtx.go#L23-L30) (SDK's standard transaction type) as it performs gas and fee checks using the Ethereum logic from Geth instead of the Cosmos SDK checks done on the `auth` module `AnteHandler`.

+++ https://github.com/cosmos/ethermint/blob/v0.3.1/x/evm/types/msg.go#L117-L124

+++ https://github.com/cosmos/ethermint/blob/v0.3.1/x/evm/types/tx_data.go#L12-L29

This message validation is expected to fail if:

- `Data.Price` (i.e gas price) is ≤ 0.
- `Data.Amount` is negative

The transaction execution is expected to fail if:

- Any of the custom `AnteHandler` Ethereum decorators checks fail:
  - Minimum gas amount requirements for transaction
  - Tx sender account doesn't exist or hasn't enough balance for fees
  - Account sequence doesn't match the transaction `Data.AccountNonce`
  - Message signature verification fails
- EVM contract creation (i.e `evm.Create`) fails, or `evm.Call` fails
