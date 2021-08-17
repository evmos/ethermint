<!--
order: 3
-->

# Transactions

## Routing

Ethermint needs to parse and handle transactions routed for both the EVM and for the Cosmos hub. We
attempt to achieve this by mimicking [geth's](https://github.com/ethereum/go-ethereum) `Transaction`
structure and treat it as a unique Cosmos SDK message type. An Ethereum transaction is a single
[`sdk.Msg`](https://godoc.org/github.com/cosmos/cosmos-sdk/types#Msg). All relevant Ethereum
transaction information is contained in this message. This includes the signature, gas, payload,
amount, etc.

Being that Ethermint implements the Tendermint ABCI application interface, as transactions are
consumed, they are passed through a series of handlers. Once such handler, the `AnteHandler`, is
responsible for performing preliminary message execution business logic such as fee payment,
signature verification, etc. This is particular to Cosmos SDK routed transactions. Ethereum routed
transactions will bypass this as the EVM handles the same business logic.

All EVM transactions are [RLP](./../core/encoding.md#rlp) encoded using a custom tx encoder.

## Signers

The signature processing and verification in Ethereum is performed by the `Signer` interface. The
protocol supports different signer types based on the chain configuration params and the block number.

+++ https://github.com/ethereum/go-ethereum/blob/v1.10.3/core/types/transaction_signing.go#L145-L166

Ethermint supports all Ethereum `Signer`s up to the latest go-ethereum version (London, Berlin,
EIP155, Homestead and Frontier). The chain will generate the latest `Signer` type depending on the
`ChainConfig`.

## Next {hide}

Learn about how [gas](./gas.md) is used on Ethermint {hide}
