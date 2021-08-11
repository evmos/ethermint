<!--
order: 1
-->

# Running the Server

Learn how to run and setup the JSON-RPC server on Ethermint. {synopsis}

## Enable Server

To enable RPC server use the following flag (set to true by default).

```bash
ethermintd start --evm-rpc.enable
```

## Defining Namespaces

`Eth`,`Net` and `Web3` [namespaces](./namespaces) are enabled by default. In order to enable other namespaces use flag `--evm-rpc.api`.

```bash
ethermintd start --evm-rpc.api eth,txpool,personal,net,debug,web3,miner
```

### CORS

If accessing the RPC from a browser, CORS will need to be enabled with the appropriate domain set. Otherwise, JavaScript calls are limit by the same-origin policy and requests will fail:

```bash
ethermintd start --evm-rpc.enable-unsafe-cors
```
