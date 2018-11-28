# Transactions

> NOTE: The specification documented below is still highly active in development
and subject to change.

## Routing

Ethermint needs to parse and handle transactions routed for both the EVM and for
the Cosmos hub. We attempt to achieve this by mimicking
[Geth's](https://github.com/ethereum/go-ethereum) `Transaction` structure and
treat it as a unique Cosmos SDK message type. An Ethereum transaction is a single
[`sdk.Msg`](https://godoc.org/github.com/cosmos/cosmos-sdk/types#Msg) contained
in an [`auth.StdTx`](https://godoc.org/github.com/cosmos/cosmos-sdk/x/auth#StdTx).
All relevant Ethereum transaction information is contained in this message. This
includes the signature, gas, payload, etc.

Being that Ethermint implements the Tendermint ABCI application interface, as
transactions are consumed, they are passed through a series of handlers. Once such
handler, the `AnteHandler`, is responsible for performing preliminary message
execution business logic such as fee payment, signature verification, etc. This is
particular to Cosmos SDK routed transactions. Ethereum routed transactions will
bypass this as the EVM handles the same business logic.

Ethereum routed transactions coming from a web3 source are expected to be RLP
encoded, however all internal interaction between Ethermint and Tendermint will
utilize Amino encoding.

__Note__: Our goal is to utilize Geth/Turbo-Geth as a library, at least as much
as possible, so it should be expected that these types and the operations you may
perform on them will keep in line with Ethereum (e.g. signature algorithms and
gas/fees). In addition, we aim to have existing tooling and frameworks in the
Ethereum ecosystem have 100% compatibility with creating transactions in Ethermint.

## Signatures

Ethermint supports [EIP-155](https://github.com/ethereum/EIPs/blob/master/EIPS/eip-155.md)
signatures. A `Transaction` is expected to have a single signature for Ethereum
routed transactions. However, just as in Cosmos, Ethermint will support multiple
signers for non-Ethereum transactions. Signatures over the
`Transaction` type are identical to Ethereum and the signatures will not be duplicated
in the embedding [`auth.StdTx`](https://godoc.org/github.com/cosmos/cosmos-sdk/x/auth#StdTx).

## Gas & Fees

TODO
