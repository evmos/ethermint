<!--
order: 1
-->

# Run a Node

Run a local node and start the REST and JSON-RPC clients {synopsis}

Clone and build Ethermint:

```bash
git clone <https://github.com/ChainSafe/ethermint>
cd ethermint
make install
```

Run the local testnet node with faucet enabled:

::: warning
The script below will remove any pre-existing binaries installed
:::

```bash
./init.sh
```

In another terminal window or tab, run the Ethereum JSON-RPC server as well as the SDK REST server:

```bash
emintcli rest-server --laddr "tcp://localhost:8545" --unlock-key mykey --chain-id 8
```

## Key Management

To run a node with the same key every time:
replace `emintcli keys add $KEY` in `./init.sh` with:

```bash
echo "your mnemonic here" | emintcli keys add ethermintkey --recover
```

::: tip
Ethermint currently only supports 24 word mnemonics.
:::

You can generate a new key/mnemonic with

```bash
emintcli keys add <mykey>
```

To export your ethermint key as an ethereum private key (for use with Metamask for example):

```bash
emintcli keys unsafe-export-eth-key <mykey>
```

## Requesting tokens though the testnet faucet

Once the ethermint daemon is up and running, you can request tokens to your address using the `faucet` module:

```bash
# query your initial balance
emintcli q bank balances $(emintcli keys show <mykey> -a)  

# send a tx to request tokens to your account address
emintcli tx faucet request 100photon --from <mykey>

# query your balance after the request
emintcli q bank balances $(emintcli keys show <mykey> -a)
```

You can also check to total amount funded by the faucet and the total supply of the chain via:

```bash
# total amount funded by the faucet
emintcli q faucet funded

# total supply
emintcli q supply total
```

## Next {hide}

Learn about Ethermint [accounts](./../basic/accounts.md) {hide}
