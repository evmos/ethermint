<!--
order: 2
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

+++ https://github.com/tharsis/ethermint/blob/main/types/account.pb.go#L28-L33

## Addresses and Public Keys

[BIP-0173](https://github.com/satoshilabs/slips/blob/master/slip-0173.md) defines a new format for segregated witness output addresses that contains a human-readable part that identifies the Bech32 usage. Ethermint uses the following HRP (human readable prefix) as the base HRP:

| Network   | Mainnet | Testnet | Regtest |
| --------- | ------- | ------- | ------- |
| Ethermint | `eth`   | `eth`   |         |

There are 3 main types of HRP for the `Addresses`/`PubKeys` available by default on Ethermint:

- Addresses and Keys for **accounts**, which identify users (e.g. the sender of a `message`). They are derived using the **`eth_secp256k1`** curve.
- Addresses and Keys for **validator operators**, which identify the operators of validators. They are derived using the **`eth_secp256k1`** curve.
- Addresses and Keys for **consensus nodes**, which identify the validator nodes participating in consensus. They are derived using the **`ed25519`** curve.

|                    | Address bech32 Prefix | Pubkey bech32 Prefix | Curve           | Address byte length | Pubkey byte length |
|--------------------|-----------------------|----------------------|-----------------|---------------------|--------------------|
| Accounts           | `eth`                 | `ethpub`             | `eth_secp256k1` | `20`                | `33` (compressed)  |
| Validator Operator | `ethvaloper`          | `ethvaloperpub`      | `eth_secp256k1` | `20`                | `33` (compressed)  |
| Consensus Nodes    | `ethvalcons`          | `ethvalconspub`      | `ed25519`       | `20`                | `32`               |

## Address formats for clients

`EthAccount` can be represented in both [Bech32](https://en.bitcoin.it/wiki/Bech32) (`eth1...`) and hex (`0x...`) formats for Ethereum's Web3 tooling compatibility.

The Bech32 format is the default format for Cosmos-SDK queries and transactions through CLI and REST
clients. The hex format on the other hand, is the Ethereum `common.Address` representation of a
Cosmos `sdk.AccAddress`.

- **Address (Bech32)**: `eth14au322k9munkmx5wrchz9q30juf5wjgz2cfqku`
- **Address ([EIP55](https://eips.ethereum.org/EIPS/eip-55) Hex)**: `0xAF79152AC5dF276D9A8e1E2E22822f9713474902`
- **Compressed Public Key**: `{"@type":"/ethermint.crypto.v1beta1.ethsecp256k1.PubKey","key":"ApNNebT58zlZxO2yjHiRTJ7a7ufjIzeq5HhLrbmtg9Y/"}`

### Address conversion

The `ethermintd debug addr <address>` can be used to convert an address between hex and bech32 formats. For example:

```bash
ethermintd debug addr eth12uqc42yj77hk64cdr3vsnpkfs6k0pllln7rudt
    Address: [87 1 138 168 146 247 175 109 87 13 28 89 9 134 201 134 172 240 255 255]
    Address (hex): 57018AA892F7AF6D570D1C590986C986ACF0FFFF
    Bech32 Acc: eth12uqc42yj77hk64cdr3vsnpkfs6k0pllln7rudt
    Bech32 Val: ethvaloper12uqc42yj77hk64cdr3vsnpkfs6k0pllldvagr4

ethermintd debug addr 57018AA892F7af6D570D1c590986c986aCf0fFff
    Address: [87 1 138 168 146 247 175 109 87 13 28 89 9 134 201 134 172 240 255 255]
    Address (hex): 57018AA892F7AF6D570D1C590986C986ACF0FFFF
    Bech32 Acc: eth12uqc42yj77hk64cdr3vsnpkfs6k0pllln7rudt
    Bech32 Val: ethvaloper12uqc42yj77hk64cdr3vsnpkfs6k0pllldvagr4
```

::: tip
Add the `0x` prefix to the returned hex address above to represent the Ethereum hex address format. For example:
`Address (hex): 57018AA892F7AF6D570D1C590986C986ACF0FFFF` implies that `0x57018AA892F7AF6D570D1C590986C986ACF0FFFF` is the Ethereum hex address.
:::

### Key output

::: tip
The Cosmos SDK Keyring output (i.e `ethermintd keys`) only supports addresses and public keys in Bech32 format.
:::

We can use the `keys show` command of `ethermintd` with the flag `--bech <type> (acc|val|cons)` to
obtain the addresses and keys as mentioned above,

```bash
ethermintd keys show mykey --bech acc
- name: mykey
  type: local
  address: eth1qsklxwt77qrxur494uvw07zjynu03dq9alwh37
  pubkey: '{"@type":"/ethermint.crypto.v1.ethsecp256k1.PubKey","key":"A8nbJ3eW9oAb2RNZoS8L71jFMfjk6zVa1UISYgKK9HPm"}'
  mnemonic: ""

ethermintd keys show test --bech val
- name: mykey
  type: local
  address: ethvaloper1qsklxwt77qrxur494uvw07zjynu03dq9rdsrlq
  pubkey: '{"@type":"/ethermint.crypto.v1.ethsecp256k1.PubKey","key":"A8nbJ3eW9oAb2RNZoS8L71jFMfjk6zVa1UISYgKK9HPm"}'
  mnemonic: ""

ethermintd keys show test --bech cons
- name: mykey
  type: local
  address: ethvalcons1qsklxwt77qrxur494uvw07zjynu03dq9h7rlnp
  pubkey: '{"@type":"/ethermint.crypto.v1.ethsecp256k1.PubKey","key":"A8nbJ3eW9oAb2RNZoS8L71jFMfjk6zVa1UISYgKK9HPm"}'
  mnemonic: ""
```

## Querying an Account

You can query an account address using the CLI, gRPC or

### Command Line Interface

```bash
# NOTE: the --output (-o) flag will define the output format in JSON or YAML (text)
ethermintd q auth account $(ethermintd keys show <MYKEY> -a) -o text
|
  '@type': /ethermint.types.v1beta1.EthAccount
  base_account:
    account_number: "3"
    address: inj14au322k9munkmx5wrchz9q30juf5wjgz2cfqku
    pub_key: null
    sequence: "0"
  code_hash: xdJGAYb3IzySfn2y3McDwOUAtlPKgic7e/rYBF2FpHA=
```

### Cosmos gRPC and REST

``` bash
# GET /cosmos/auth/v1beta1/accounts/{address}
curl -X GET "http://localhost:10337/cosmos/auth/v1beta1/accounts/eth14au322k9munkmx5wrchz9q30juf5wjgz2cfqku" -H "accept: application/json"
```

### JSON-RPC

To retrieve the Ethereum hex address using Web3, use the JSON-RPC [`eth_accounts`](./../api/json-rpc/endpoints#eth-accounts) or [`personal_listAccounts`](./../api/json-rpc/endpoints#personal-listAccounts) endpoints:

```bash
# query against a local node
curl -X POST --data '{"jsonrpc":"2.0","method":"eth_accounts","params":[],"id":1}' -H "Content-Type: application/json" http://localhost:8545

curl -X POST --data '{"jsonrpc":"2.0","method":"personal_listAccounts","params":[],"id":1}' -H "Content-Type: application/json" http://localhost:8545
```

## Next {hide}

Learn about Ethermint [transactions](./transactions.md) {hide}
