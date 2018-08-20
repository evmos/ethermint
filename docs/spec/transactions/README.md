# Transactions

> NOTE: The specification documented below is still highly active in development
and subject to change.

## Routing

Ethermint needs to parse and handle transactions routed for both the EVM and for
the Cosmos hub. We attempt to achieve this by mimicking [Geth's](https://github.com/ethereum/go-ethereum) `Transaction` structure and utilizing
the `Payload` as the potential encoding of a Cosmos-routed transaction. What
designates this encoding, and ultimately routing, is the `Transaction.Recipient`
address -- if this address matches some global unique predefined and configured
address, we regard it as a transaction meant for Cosmos, otherwise, the transaction
is a pure Ethereum transaction and will be executed in the EVM.

For Cosmos routed transactions, the `Transaction.Payload` will contain an
embedded encoded type: `EmbeddedTx`. This structure is analogous to the Cosmos
SDK `sdk.StdTx`. If a client wishes to send an `EmbeddedTx`, it must first encode
the embedded transaction, and then encode the embedding `Transaction`.

__Note__: The `Transaction` and `EmbeddedTx` types utilize the [Amino](https://github.com/tendermint/go-amino) object serialization protocol and as such,
the `Transaction` is not an exact replica of what will be found in Ethereum. Our
goal is to utilize Geth as a library, at least as much as possible, so it should
be expected that these types and the operations you may perform on them will keep
in line with Ethereum.

Being that Ethermint implements the ABCI application interface, as transactions
are sent they are passed through a series of handlers. Once such handler, `runTx`,
is responsible for invoking the `TxDecoder` which performs the business logic of
properly deserializing raw transaction bytes into either an Ethereum transaction
or a Cosmos transaction.

## Transactions & Messages

The SDK distinguishes between transactions (`sdk.Tx`) and messages (`sdk.Msg`).
A `sdk.Tx` is a list of `sdk.Msg`s wrapped with authentication and fee data. Users
can create messages containing arbitrary information by implementing the `sdk.Msg`
interface.

In Ethermint, the `Transaction` type implements the Cosmos SDK `sdk.Tx` interface.
It addition, it implements the Cosmos SDK `sdk.Msg` interface for the sole purpose
of being to perform basic validation checks in the `BaseApp`. It, however, has
no distinction between transactions and messages.

The `EmbeddedTx`, being analogous to the Cosmos SDK `sdk.StdTx`, implements the
Cosmos SDK `sdk.Tx` interface.

## Signatures

Ethermint supports [EIP-155](https://github.com/ethereum/EIPs/blob/master/EIPS/eip-155.md)
signatures. However, just as in Cosmos, Ethermint will support multiple signers.
A client is expected to sign the `Transaction` just as in Ethereum, however, the
`EmbeddedTx` contains a canonical signature structure that itself contains the
signature and other information such as an account's sequence number. The sequence
number is expected to increment every time a message is signed by a given account.
This prevents "replay attacks", where the same message could be executed over and
over again.

An `EmbeddedTx` list of signatures must much the unique list of addresses returned by
each message's `GetSigners` call. In addition, the address of first signer of the
`EmbeddedTx` is responsible for paying the fees and must also match the address of
the signer of the embedding `Transaction`. As such, there will be one duplicate
signature.

## Gas & Fees

TODO
