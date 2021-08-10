<!--
order: 2
-->

# `ethermintd`

`ethermintd` is the all-in-one command-line interface. It supports wallet management, queries and transaction operations {synopsis}

## Pre-requisite Readings

- [Installation](./installation.md) {prereq}

## Build and Configuration

### Using `ethermintd`

After you have obtained the latest `ethermintd` binary, run:

```bash
ethermintd [command]
```

Check the version you are running using

```bash
ethermintd version
```

There is also a `-h`, `--help` command available

```bash
ethermintd -h
```

::: tip
You can also enable auto-completion with the `ethermintd completion` command. For example, at the start of a bash session, run `. <(ethermintd completion)`, and all `ethermintd` subcommands will be auto-completed.
:::

### Config and data directory

By default, your config and data are stored in the folder located at the `~/.ethermintd` directory.

:::warning
Make sure you have backed up your wallet storage after creating the wallet or else your funds may be inaccessible in case of accident forever.
:::

To specify the `ethermintd` config and data storage directory; you can update it using the global flag `--home <directory>`

### Client configuration

We can view the default client config setting by using `ethermintd config` command:

```bash
ethermintd config
{
 "chain-id": "",
 "keyring-backend": "os",
 "output": "text",
 "node": "tcp://localhost:26657",
 "broadcast-mode": "sync"
}
```

We can make changes to the default settings upon our choices, so it allows users to set the configuration beforehand all at once, so it would be ready with the same config afterward.

For example, the chain identifier can be changed to `ethermint-777` from a blank name by using:

```bash
ethermintd config "chain-id" ethermint-777
ethermintd config
{
 "chain-id": "ethermint-777",
 "keyring-backend": "os",
 "output": "text",
 "node": "tcp://localhost:26657",
 "broadcast-mode": "sync"
}
```

Other values can be changed in the same way.

Alternatively, we can directly make the changes to the config values in one place at client.toml. It is under the path of `.ethermint/config/client.toml` in the folder where we installed ethermint:

```toml
############################################################################
### Client Configuration ###

############################################################################

# The network chain ID

chain-id = "ethermint-777"

# The keyring's backend, where the keys are stored (os|file|kwallet|pass|test|memory)

keyring-backend = "os"

# CLI output format (text|json)

output = "number"

# <host>:<port> to Tendermint RPC interface for this chain

node = "tcp://localhost:26657"

# Transaction broadcasting mode (sync|async|block)

broadcast-mode = "sync"
```

After the necessary changes are made in the `client.toml`, then save. For example, if we directly change the chain-id from `ethermint-0` to `etherminttest-1`, and output to number, it would change instantly as shown below.

```bash
ethermintd config
{
 "chain-id": "etherminttest-1",
 "keyring-backend": "os",
 "output": "number",
 "node": "tcp://localhost:26657",
 "broadcast-mode": "sync"
}
```

### Options

A list of commonly used flags of `ethermintd` is listed below:

| Option              | Description                   | Type         | Default Value   |
| ------------------- | ----------------------------- | ------------ | --------------- |
| `--chain-id`        | Full Chain ID                 | String       | ---             |
| `--home`            | Directory for config and data | string       | `~/.ethermintd` |
| `--keyring-backend` | Select keyring's backend      | os/file/test | os              |
| `--output`          | Output format                 | string       | "text"          |

## Command list

A list of commonly used `ethermintd` commands. You can obtain the full list by using the `ethermintd -h` command.

| Command         | Description              | Subcommands (example)                                        |
| --------------- | ------------------------ | ------------------------------------------------------------ |
| `keys`          | Keys management          | `list`, `show`, `add`, `add  --recover`, `delete`                |
| `tx`            | Transactions subcommands | `bank send`, `ibc-transfer transfer`, `distribution withdraw-all-rewards` |
| `query`         | Query subcommands        | `bank balance`, `staking validators`, `gov proposals`                          |
| `tendermint`    | Tendermint subcommands   | `show-address`, `show-node-id`, `version`                                |
| `config` | Client configuration     |                                                              |
| `init`  | Initialize full node     |                                                              |
| `start` | Run full node            |                                                              |
|     `version`            |     Ethermint version                     |                                                              |
