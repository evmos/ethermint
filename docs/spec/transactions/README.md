# Transactions

> NOTE: The specification documented below is still highly active in development
and subject to change.

## Routing

Ethermint needs to parse and handle transactions routed for both the EVM and for
the Cosmos hub. We attempt to achieve this by mimicking [Geth's](https://github.com/ethereum/go-ethereum) `Transaction` structure and utilizing
the `Payload` as the potential encoding of a Cosmos-routed transaction. What
designates this encoding, and ultimately routing, is the `Recipient` address --
if this address matches some global unique predefined and configured address,
we regard it as a transaction meant for Cosmos, otherwise, the transaction is a
pure Ethereum transaction and will be executed in the EVM.

For Cosmos routed transactions, the `Transaction.Payload` will contain an [Amino](https://github.com/tendermint/go-amino) encoded embedded transaction that must
implement the `sdk.Tx` interface. Note, the embedding (outer) `Transaction` is
still RLP encoded in order to preserve compatibility with existing tooling. In
addition, at launch, Ethermint will only support the `auth.StdTx` embedded Cosmos
transaction type.

Being that Ethermint implements the Tendermint ABCI application interface, as
transactions are consumed, they are passed through a series of handlers. Once such
handler, `runTx`, is responsible for invoking the `TxDecoder` which performs the
business logic of properly deserializing raw transaction bytes into either an
Ethereum transaction or a Cosmos transaction.

__Note__: Our goal is to utilize Geth as a library, at least as much as possible,
so it should be expected that these types and the operations you may perform on
them will keep in line with Ethereum (e.g. signature algorithms and gas/fees).
In addition, we aim to have existing tooling and frameworks in the Ethereum
ecosystem have 100% compatibility with creating transactions in Ethermint.

## Transactions & Messages

The SDK distinguishes between transactions (`sdk.Tx`) and messages (`sdk.Msg`).
A `sdk.Tx` is a list of `sdk.Msg` wrapped with authentication and fee data. Users
can create messages containing arbitrary information by implementing the `sdk.Msg`
interface.

In Ethermint, the `Transaction` type implements the Cosmos SDK `sdk.Tx` interface.
It addition, it implements the Cosmos SDK `sdk.Msg` interface for the sole purpose
of being to perform basic validation checks in the `BaseApp`. It, however, has
no distinction between transactions and messages.

## Signatures

Ethermint supports [EIP-155](https://github.com/ethereum/EIPs/blob/master/EIPS/eip-155.md)
signatures. A `Transaction` is expected to have a single signature for Ethereum
routed transactions. However, just as in Cosmos, Ethermint will support multiple
signers for embedded Cosmos routed transactions. Signatures over the
`Transaction` type are identical to Ethereum. However, the embedded transaction contains
a canonical signature structure that contains the signature itself and other
information such as an account's sequence number. This, in addition to the chainID,
helps prevent "replay attacks", where the same message could be executed over and
over again.

An embedded transaction's list of signatures must much the unique list of addresses
returned by each message's `GetSigners` call. In addition, the address of first
signer of the embedded transaction is responsible for paying the fees.

## Gas & Fees

TODO
