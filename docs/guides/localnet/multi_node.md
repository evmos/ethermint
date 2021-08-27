<!--
order: 2
-->

# Multi Node

## Pre-requisite Readings

- [Install Starport](https://docs.starport.network/#install-starport)  {prereq}
- [Install Docker](https://docs.docker.com/engine/installation/)  {prereq}
- [Install docker-compose](https://docs.docker.com/compose/install/)  {prereq}

## Automated Localnet with Starport

Once you have installed `starport`, just run the localnet by using

```bash
starport chain serve --reset-once -v -c ./starport.yml
```

## Automated Localnet with Docker

### Build & Start

To build start a 4 node testnet run:

```bash
make localnet-start
```

This command creates a 4-node network using the `ethermintdnode` Docker image.
The ports for each node are found in this table:

| Node ID          | P2P Port | Tendermint RPC Port | REST/ Ethereum JSON-RPC Port | WebSocket Port |
|------------------|----------|---------------------|------------------------------|----------------|
| `ethermintnode0` | `26656`  | `26657`             | `8545`                       | `8546`         |
| `ethermintnode1` | `26659`  | `26660`             | `8547`                       | `8548`         |
| `ethermintnode2` | `26661`  | `26662`             | `8549`                       | `8550`         |
| `ethermintnode3` | `26663`  | `26664`             | `8551`                       | `8552`         |

To update the binary, just rebuild it and restart the nodes

```bash
make localnet-start
```

The command above  command will run containers in the background using Docker compose. You will see the network being created:

```bash
...
Creating network "ethermint_localnet" with driver "bridge"
Creating ethermintdnode0 ... done
Creating ethermintdnode2 ... done
Creating ethermintdnode1 ... done
Creating ethermintdnode3 ... done
```

### Stop Localnet

Once you are done, execute:

```bash
make localnet-stop
```

### Configuration

The `make localnet-start` creates files for a 4-node testnet in `./build` by
calling the `ethermintd testnet` command. This outputs a handful of files in the
`./build` directory:

```bash
tree -L 3 build/

build/
├── ethermintd
├── ethermintd
├── gentxs
│   ├── node0.json
│   ├── node1.json
│   ├── node2.json
│   └── node3.json
├── node0
│   ├── ethermintd
│   │   ├── key_seed.json
│   │   └── keyring-test-cosmos
│   └── ethermintd
│       ├── config
│       ├── data
│       └── ethermintd.log
├── node1
│   ├── ethermintd
│   │   ├── key_seed.json
│   │   └── keyring-test-cosmos
│   └── ethermintd
│       ├── config
│       ├── data
│       └── ethermintd.log
├── node2
│   ├── ethermintd
│   │   ├── key_seed.json
│   │   └── keyring-test-cosmos
│   └── ethermintd
│       ├── config
│       ├── data
│       └── ethermintd.log
└── node3
    ├── ethermintd
    │   ├── key_seed.json
    │   └── keyring-test-cosmos
    └── ethermintd
        ├── config
        ├── data
        └── ethermintd.log
```

Each `./build/nodeN` directory is mounted to the `/ethermintd` directory in each container.

### Logging

In order to see the logs of a particular node you can use the following command:

```bash
# node 0: daemon logs
docker exec ethermintdnode0 tail ethermintd.log

# node 0: REST & RPC logs
docker exec ethermintdnode0 tail ethermintd.log
```

The logs for the daemon will look like:

```bash
I[2020-07-29|17:33:52.452] starting ABCI with Tendermint                module=main
E[2020-07-29|17:33:53.394] Can't add peer's address to addrbook         module=p2p err="Cannot add non-routable address 272a247b837653cf068d39efd4c407ffbd9a0e6f@192.168.10.5:26656"
E[2020-07-29|17:33:53.394] Can't add peer's address to addrbook         module=p2p err="Cannot add non-routable address 3e05d3637b7ebf4fc0948bbef01b54d670aa810a@192.168.10.4:26656"
E[2020-07-29|17:33:53.394] Can't add peer's address to addrbook         module=p2p err="Cannot add non-routable address 689f8606ede0b26ad5b79ae244c14cc67ab4efe7@192.168.10.3:26656"
I[2020-07-29|17:33:58.828] Executed block                               module=state height=88 validTxs=0 invalidTxs=0
I[2020-07-29|17:33:58.830] Committed state                              module=state height=88 txs=0 appHash=90CC5FA53CF8B5EC49653A14DA20888AD81C92FCF646F04D501453FD89FCC791
I[2020-07-29|17:34:04.032] Executed block                               module=state height=89 validTxs=0 invalidTxs=0
I[2020-07-29|17:34:04.034] Committed state                              module=state height=89 txs=0 appHash=0B54C4DB1A0DACB1EEDCD662B221C048C826D309FD2A2F31FF26BAE8D2D7D8D7
I[2020-07-29|17:34:09.381] Executed block                               module=state height=90 validTxs=0 invalidTxs=0
I[2020-07-29|17:34:09.383] Committed state                              module=state height=90 txs=0 appHash=75FD1EE834F0669D5E717C812F36B21D5F20B3CCBB45E8B8D415CB9C4513DE51
I[2020-07-29|17:34:14.700] Executed block                               module=state height=91 validTxs=0 invalidTxs=0
```

::: tip
You can disregard the `Can't add peer's address to addrbook` warning. As long as the blocks are
being produced and the app hashes are the same for each node, there should not be any issues.
:::

Whereas the logs for the REST & RPC server would look like:

```bash
I[2020-07-30|09:39:17.488] Starting application REST service (chain-id: "7305661614933169792")... module=rest-server
I[2020-07-30|09:39:17.488] Starting RPC HTTP server on 127.0.0.1:8545   module=rest-server
...
```

#### Follow Logs

You can also watch logs as they are produced via Docker with the `--follow` (`-f`) flag, for
example:

```bash
docker logs -f ethermintdnode0
```

### Interact with the Localnet

#### Ethereum JSON-RPC & Websocket Ports

To interact with the testnet via WebSockets or RPC/API, you will send your request to the corresponding ports:

| EVM JSON-RPC | Eth Websocket |
|--------------|---------------|
| `8545`       | `8546`        |

You can send a curl command such as:

```bash
curl -X POST --data '{"jsonrpc":"2.0","method":"eth_accounts","params":[],"id":1}' -H "Content-Type: application/json" 192.162.10.1:8545
```

::: tip
The IP address will be the public IP of the docker container.
:::

Additional instructions on how to interact with the WebSocket can be found on the [events documentation](./events.md#ethereum-websocket).

### Keys & Accounts

To interact with `ethermintd` and start querying state or creating txs, you use the
`ethermintd` directory of any given node as your `home`, for example:

```bash
ethermintd keys list --home ./build/node0/ethermintd
```

Now that accounts exists, you may create new accounts and send those accounts
funds!

::: tip
**Note**: Each node's seed is located at `./build/nodeN/ethermintd/key_seed.json` and can be restored to the CLI using the `ethermintd keys add --restore` command
:::

### Special Binaries

If you have multiple binaries with different names, you can specify which one to run with the BINARY environment variable. The path of the binary is relative to the attached volume. For example:

```bash
# Run with custom binary
BINARY=ethermint make localnet-start
```
