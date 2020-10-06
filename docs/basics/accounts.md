<!--
order: 1
-->

# Accounts

This document describes the in-built accounts system of Ethermint. {synopsis}

## Pre-requisite Readings

- [Cosmos SDK Accounts](https://docs.cosmos.network/master/basics/accounts.html) {prereq}
- [Ethereum Accounts](https://ethereum.org/en/whitepaper/#ethereum-accounts) {prereq}

## Ethermint Accounts

Ethermint defines its own custom `Account` type that uses Ethereum's ECDSA secp256k1 curve for keys. This
satisfies the [EIP84](https://github.com/ethereum/EIPs/issues/84) for full [BIP44](https://github.com/bitcoin/bips/blob/master/bip-0044.mediawiki) paths.
The root HD path for Ethermint-based accounts is `m/44'/60'/0'/0`.

+++ https://github.com/ChainSafe/ethermint/blob/v0.1.0/types/account.go#L31-L36

## Addresses and Public Keys

There are 3 main types of `Addresses`/`PubKeys` available by default on Ethermint:

- Addresses and Keys for **accounts**, which identify users (e.g. the sender of a `message`). They are derived using the **`eth_secp256k1`** curve.
- Addresses and Keys for **validator operators**, which identify the operators of validators. They are derived using the **`eth_secp256k1`** curve.
- Addresses and Keys for **consensus nodes**, which identify the validator nodes participating in consensus. They are derived using the **`ed25519`** curve.

|                    | Address bech32 Prefix | Pubkey bech32 Prefix | Curve           | Address byte length | Pubkey byte length |
|--------------------|-----------------------|----------------------|-----------------|---------------------|--------------------|
| Accounts           | `eth`                 | `ethpub`             | `eth_secp256k1` | `20`                | `33` (compressed)  |
| Validator Operator | `ethvaloper`          | `ethvaloperpub`      | `eth_secp256k1` | `20`                | `33` (compressed)  |
| Consensus Nodes    | `ethvalcons`          | `ethvalconspub`      | `ed25519`       | `20`                | `32`               |

## Address formats for clients

`EthAccount`s can be represented in both [Bech32](https://en.bitcoin.it/wiki/Bech32) and hex format for Ethereum's Web3 tooling compatibility.

The Bech32 format is the default format for Cosmos-SDK queries and transactions through CLI and REST
clients. The hex format on the other hand, is the Ethereum `common.Address` representation of a
Cosmos `sdk.AccAddress`.

- Address (Bech32): `eth1crwhac03z2pcgu88jfnqnwu66xlthlz2rhljah`
- Address ([EIP55](https://eips.ethereum.org/EIPS/eip-55) Hex): `0xc0dd7ee1f112838470e7926609bb9ad1bebbfc4a`
- Compressed Public Key (Bech32): `ethpub1pfqnmk6pqnwwuw0h9hj58t2hyzwvqc3truhhp5tl5hfucezcfy2rs8470nkyzju2vmk645fzmw2wveaqcqek767kwa0es9rmxe9nmmjq84cpny3fvj6tpg`

You can query an account address using the Cosmos CLI or REST clients:

```bash
# NOTE: the --output (-o) flag will define the output format in JSON or YAML (text)
ethermintcli q auth account $(ethermintcli keys show <MYKEY> -a) -o text
|
  address: eth1f8rqrfwut7ngkxwth0gt99h0lxnxsp09ngvzwl
  eth_address: 0x49c601A5DC5FA68b19CBbbd0b296eFF9a66805e5
  coins:
  - denom: aphoton
    amount: "1000000000000000000"
  - denom: stake
    amount: "999999999900000000"
  public_key: ethpub1pfqnmkepqw45vpsn6dzvm7k22zrghx0nfewjdfacy7wyycv5evfk57kyhwr8cqj5r4x
  account_number: 0
  sequence: 1
  code_hash: c5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470
```

``` bash
# GET /auth/accounts/{address}
curl -X GET "<NODE_IP>/auth/accounts/eth1f8rqrfwut7ngkxwth0gt99h0lxnxsp09ngvzwl" -H "accept: application/json"
```

::: tip
The Cosmos SDK Keyring output (i.e `ethermintcli keys`) only supports addresses and public keys in Bech32 format.
:::

To retrieve the Ethereum hex address using Web3, use the JSON-RPC `eth_accounts` endpoint:

```bash
# query against a local node
curl -X POST --data '{"jsonrpc":"2.0","method":"eth_accounts","params":[],"id":1}' -H "Content-Type: application/json" http://localhost:26664
```

## Next {hide}

Learn about Ethermint [transactions](./transactions.md) {hide}
