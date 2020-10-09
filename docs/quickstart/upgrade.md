<!--
order: 5
-->

# Upgrade Node

Learn how to upgrade your full node to the latest software version {synopsis}

## Software Upgrade

These instructions are for full nodes that have ran on previous versions of and would like to upgrade to the latest testnet.

First, stop your instance of `ethermintd`. Next, upgrade the software:

```bash
cd ethermint
git fetch --all && git checkout <new_version>
make install
```

::: tip
If you have issues at this step, please check that you have the latest stable version of GO installed.
:::

You will need to ensure that the version installed matches the one needed for th testnet. Check the Ethermint [release page](https://github.com/cosmos/ethermint/releases) for details on each release.

## Upgrade Genesis File

:::warning
If the new version you are upgrading to has breaking changes, you will have to restart your chain. If it is **not** breaking, you can skip to [Restart](#restart-node).
:::

To upgrade the genesis file, you can either fetch it from a trusted source or export it locally using the `ethermintd export` command.

### Fetch from a Trusted Source

If you are joining an existing testnet, you can fetch the genesis from the appropriate testnet source/repository where the genesis file is hosted.

Save the new genesis as `new_genesis.json`. Then, replace the old `genesis.json` with `new_genesis.json`.

```bash
cd $HOME/.ethermintd/config
cp -f genesis.json new_genesis.json
mv new_genesis.json genesis.json
```

Finally, go to the [reset data](./run_node.md#reset-data) section.

### Export State to a new Genesis locally

Ethermint can dump the entire application state to a JSON file. This, besides upgrades, can be
useful for manual analysis of the state at a given height.

Export state with:

```bash
ethermintd export > new_genesis.json
```

You can also export state from a particular height (at the end of processing the block of that height):

```bash
ethermintd export --height [height] > new_genesis.json
```

If you plan to start a new network for 0 height (i.e genesis) from the exported state, export with the `--for-zero-height` flag:

```bash
ethermintd export --height [height] --for-zero-height > new_genesis.json
```

Then, replace the old `genesis.json` with `new_genesis.json`.

```bash
cp -f genesis.json new_genesis.json
mv new_genesis.json genesis.json
```

At this point, you might want to run a script to update the exported genesis into a genesis state that is compatible with your new version.

You can use the `migrate` command to migrate from a given version to the next one (eg: `v0.X.X` to `v1.X.X`):

```bash
ethermintd migrate [target-version] [/path/to/genesis.json] --chain-id=<new_chain_id> --genesis-time=<yyyy-mm-ddThh:mm:ssZ>
```

## Restart Node

To restart your node once the new genesis has been updated, use the `start` command:

```bash
ethermintd start
```

## Next {hide}

Learn about how to setup a [validator](./validator-setup.md) node on Ethermint {hide}
