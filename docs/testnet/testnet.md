<!--
order: 3
-->

# Testnet

Learn how to deploy a local testnet or connect to an existing public one {synopsis}

## Pre-requisite Readings

- [Install Ethermint](./installation.md) {prereq}

### Supported OS

We officially support macOS, Windows and Linux only. Other platforms may work but there is no
guarantee. We will extend our support to other platforms after we have stabilized our current
architecture.

### Minimum Requirements

To run testnet nodes, you will need a machine with the following minimum requirements:

- 4-core, x86_64 architecture processor;
- 16 GB RAM;
- 1 TB of storage space.

## Multi-node, Public, Manual Testnet

If you are looking to connect to a persistent public testnet. You will need to manually configure your node.

### Genesis and Seeds

#### Copy the Genesis File

::: tip
If you want to start a network from scratch, you will need to start the [genesis procedure](#genesis-procedure) by creating a `genesis.json` and submit + collect the genesis transactions from the [validators](./validator-setup.md).
:::

If you want to connect to an existing testnet, fetch the testnet's `genesis.json` file and copy it into the `ethermintd`'s config directory (i.e `$HOME/.ethermintd/config/genesis.json`).

Then verify the correctness of the genesis configuration file:

```bash
ethermintd validate-genesis
```

#### Add Seed Nodes

Your node needs to know how to find peers. You'll need to add healthy seed nodes to `$HOME/.ethermintd/config/config.toml`. If those seeds aren't working, you can find more seeds and persistent peers on an existing explorer.

For more information on seeds and peers, you can the Tendermint [P2P documentation](https://docs.tendermint.com/master/spec/p2p/peer.html).

#### Start testnet

The final step is to [start the nodes](./run_node.md#start-node). Once enough voting power (+2/3) from the genesis validators is up-and-running, the testnet will start producing blocks.

## Next {hide}

Learn about how to setup a [validator](./validator-setup.md) node on Ethermint {hide}
