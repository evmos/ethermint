<!--
order: 3
-->

# Testnet

Learn how to deploy a local testnet or connect to an existing one {synopsis}

## Pre-requisite Readings

- [Run Node](./run_node.md) {prereq}

## Genesis and Seeds

### Copy the Genesis File

<!-- TODO: link to genesis procedure -->
::: tip
If you want to start a network from scratch, you will need to start the genesis procedure.
:::

If you want to connect to an existing testnet, fetch the testnet's `genesis.json` file and copy it into the `emintd`'s config directory (i.e `$HOME/.emintd/config/genesis.json`).

Then verify the correctness of the genesis configuration file:

```bash
emintd validate-genesis
```

### Add Seed Nodes

Your node needs to know how to find peers. You'll need to add healthy seed nodes to `$HOME/.emintd/config/config.toml`. If those seeds aren't working, you can find more seeds and persistent peers on an existing explorer.

For more information on seeds and peers, you can the Tendermint [P2P documentation](https://docs.tendermint.com/master/spec/p2p/peer.html).

### Start testnet

The final step is to [start the nodes](./run_node.md#start-node). Once enough voting power (+2/3) from the genesis validators is up-and-running, the testnet will start producing blocks.

## Testnet faucet

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

Learn about how to setup a [validator](./validator-setup.md) node on Ethermint {hide}
