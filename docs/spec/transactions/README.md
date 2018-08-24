# Transactions

> NOTE: The specification documented below is still highly active in development
and subject to change.

## Routing

Ethermint needs to parse and handle transactions routed for both the EVM and for
the Cosmos hub. We attempt to achieve this by mimicking [Geth's](https://github.com/ethereum/go-ethereum) `Transaction` structure to handle
Ethereum transactions and utilizing the SDK's `auth.StdTx` for Cosmos
transactions. Both of these structures are registered with an [Amino](https://github.com/tendermint/go-amino) codec, so the `TxDecoder` that in invoked
during the `BaseApp#runTx`, will be able to decode raw transaction bytes into the
appropriate transaction type which will then be passed onto handlers downstream.

__Note__: Our goal is to utilize Geth as a library, at least as much as possible,
so it should be expected that these types and the operations you may perform on
them will keep in line with Ethereum (e.g. signature algorithms and gas/fees).

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
signers for `auth.StdTx` Cosmos routed transactions. Signatures over the
`Transaction` type are identical to Ethereum. However, the `auth.StdTx` contains
a canonical signature structure that contains the signature itself and other
information such as an account's sequence number. This, in addition to the chainID,
helps prevent "replay attacks", where the same message could be executed over and
over again.

An `auth.StdTx` list of signatures must much the unique list of addresses returned
by each message's `GetSigners` call. In addition, the address of first signer of
the `auth.StdTx` is responsible for paying the fees.

## Gas & Fees

TODO
