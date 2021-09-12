<!--
order: 2
-->

# Faucet

Check how to obtain testnet tokens from the Ethermint faucet website {synopsis}

## Requesting tokens

You can request tokens for the testnet by using the Ethermint [faucet](https://ethermint.io/faucet).
Simply fill in your address on the input field in bech32 (`ethm1...`) or hex (`0x...`) format.

::: warning
If you use your bech32 address, make sure you input the [account address](./../basics/accounts#addresses-and-public-keys) (`ethm1...`) and **NOT** the validator operator address (`ethmvaloper1...`)
:::

<!-- TODO: update with final website design -->
![faucet site](./img/faucet_web_page.png)

## Rate limits

To prevent the faucet account from draining the available funds, the Ethermint testnet faucet
imposes a maximum number of request for a period of time. By default the faucet service accepts 1
request per day per address. All addresses **must** be authenticated using
[Auth0](https://auth0.com/) before requesting tokens.

<!-- TODO: add screenshots of authentication window -->

## Amount

For each request, the faucet transfers $1 ~ * ~ 10^{18}$ aphotons (i.e 1 Photon) to the given address.
