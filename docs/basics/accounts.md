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
|-----------|---------|---------|---------|
| Ethermint | `ethm`  | `ethm`  |         |

There are 3 main types of HRP for the `Addresses`/`PubKeys` available by default on Ethermint:

- Addresses and Keys for **accounts**, which identify users (e.g. the sender of a `message`). They are derived using the **`eth_secp256k1`** curve.
- Addresses and Keys for **validator operators**, which identify the operators of validators. They are derived using the **`eth_secp256k1`** curve.
- Addresses and Keys for **consensus nodes**, which identify the validator nodes participating in consensus. They are derived using the **`ed25519`** curve.

|                    | Address bech32 Prefix | Pubkey bech32 Prefix | Curve           | Address byte length | Pubkey byte length |
|--------------------|-----------------------|----------------------|-----------------|---------------------|--------------------|
| Accounts           | `ethm`                | `ethmpub`            | `eth_secp256k1` | `20`                | `33` (compressed)  |
| Validator Operator | `ethmvaloper`         | `ethmvaloperpub`     | `eth_secp256k1` | `20`                | `33` (compressed)  |
| Consensus Nodes    | `ethmvalcons`         | `ethmvalconspub`     | `ed25519`       | `20`                | `32`               |

## Address formats for clients

`EthAccount` can be represented in both [Bech32](https://en.bitcoin.it/wiki/Bech32) (`ethm1...`) and hex (`0x...`) formats for Ethereum's Web3 tooling compatibility.

The Bech32 format is the default format for Cosmos-SDK queries and transactions through CLI and REST
clients. The hex format on the other hand, is the Ethereum `common.Address` representation of a
Cosmos `sdk.AccAddress`.

- **Address (Bech32)**: `ethm1j800cll9vq7l4rxfke2u74mjgkdlzrr0r5mu97`
- **Address ([EIP55](https://eips.ethereum.org/EIPS/eip-55) Hex)**: `0x91defC7fE5603DFA8CC9B655cF5772459BF10c6f`
- **Compressed Public Key**: `{"@type":"/ethermint.crypto.v1.ethsecp256k1.PubKey","key":"Aq9WtHGKtvX523b2ptvimGVfp3hZ1GDxVdINYWBM9+Gy"}`

### Address conversion

The `ethermintd debug addr <address>` can be used to convert an address between hex and bech32 formats. For example:

:::: tabs
::: tab Bech32

```bash
ethermintd debug addr ethm10jmp6sgh4cc6zt3e8gw05wavvejgr5pwtu750w
  Address bytes: [124 182 29 65 23 174 49 161 46 57 58 28 250 59 172 102 100 129 208 46]
  Address (hex): 7CB61D4117AE31A12E393A1CFA3BAC666481D02E
  Address (EIP-55): 0x7cB61D4117AE31a12E393a1Cfa3BaC666481D02E
  Bech32 Acc: ethm10jmp6sgh4cc6zt3e8gw05wavvejgr5pwtu750w
  Bech32 Val: ethmvaloper10jmp6sgh4cc6zt3e8gw05wavvejgr5pwyv5chn
```

:::
::: tab Hex

```bash
ethermintd debug addr 0x7cB61D4117AE31a12E393a1Cfa3BaC666481D02E
  Address bytes: [124 182 29 65 23 174 49 161 46 57 58 28 250 59 172 102 100 129 208 46]
  Address (hex): 7CB61D4117AE31A12E393A1CFA3BAC666481D02E
  Address (EIP-55): 0x7cB61D4117AE31a12E393a1Cfa3BaC666481D02E
  Bech32 Acc: ethm10jmp6sgh4cc6zt3e8gw05wavvejgr5pwtu750w
  Bech32 Val: ethmvaloper10jmp6sgh4cc6zt3e8gw05wavvejgr5pwyv5chn
```

:::
::::

### Key output

::: tip
The Cosmos SDK Keyring output (i.e `ethermintd keys`) only supports addresses and public keys in Bech32 format.
:::

We can use the `keys show` command of `ethermintd` with the flag `--bech <type> (acc|val|cons)` to
obtain the addresses and keys as mentioned above,

:::: tabs
::: tab Account

```bash
ethermintd keys show mykey --bech acc
- name: mykey
  type: local
  address: ethm1qsklxwt77qrxur494uvw07zjynu03dq9alwh37
  pubkey: '{"@type":"/ethermint.crypto.v1.ethsecp256k1.PubKey","key":"A8nbJ3eW9oAb2RNZoS8L71jFMfjk6zVa1UISYgKK9HPm"}'
  mnemonic: ""
```

:::
::: tab Validator

```bash
ethermintd keys show test --bech val
- name: mykey
  type: local
  address: ethmvaloper1qsklxwt77qrxur494uvw07zjynu03dq9rdsrlq
  pubkey: '{"@type":"/ethermint.crypto.v1.ethsecp256k1.PubKey","key":"A8nbJ3eW9oAb2RNZoS8L71jFMfjk6zVa1UISYgKK9HPm"}'
  mnemonic: ""
```

:::
::: tab Consensus

```bash
ethermintd keys show test --bech cons
- name: mykey
  type: local
  address: ethmvalcons1qsklxwt77qrxur494uvw07zjynu03dq9h7rlnp
  pubkey: '{"@type":"/ethermint.crypto.v1.ethsecp256k1.PubKey","key":"A8nbJ3eW9oAb2RNZoS8L71jFMfjk6zVa1UISYgKK9HPm"}'
  mnemonic: ""
```

:::
::::

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
curl -X GET "http://localhost:10337/cosmos/auth/v1beta1/accounts/ethm14au322k9munkmx5wrchz9q30juf5wjgz2cfqku" -H "accept: application/json"
```

### JSON-RPC

To retrieve the Ethereum hex address using Web3, use the JSON-RPC [`eth_accounts`](./../api/json-rpc/endpoints.md#eth-accounts) or [`personal_listAccounts`](./../api/json-rpc/endpoints#personal-listAccounts.md) endpoints:

```bash
# query against a local node
curl -X POST --data '{"jsonrpc":"2.0","method":"eth_accounts","params":[],"id":1}' -H "Content-Type: application/json" http://localhost:8545

curl -X POST --data '{"jsonrpc":"2.0","method":"personal_listAccounts","params":[],"id":1}' -H "Content-Type: application/json" http://localhost:8545
```
