<!--
order: 1
-->

# Running the Server

Learn how to run and setup the JSON-RPC server on Ethermint. {synopsis}

## Enable Server

To enable RPC server use the following flag (set to true by default).

```bash
ethermintd start --json-rpc.enable
```

## Defining Namespaces

`Eth`,`Net` and `Web3` [namespaces](./namespaces) are enabled by default. In order to enable other namespaces use flag `--json-rpc.api`.

```bash
ethermintd start --json-rpc.api eth,txpool,personal,net,debug,web3,miner
```

## Set a Gas Cap

`eth_call` and `eth_estimateGas` define a global gas cap over rpc for DoS protection. You can override the default gas cap value of 25,000,000 by passing a custom value when starting the node:

```bash
# set gas cap to 85M
ethermintd start --json-rpc.gas-cap 85000000000

# set gas cap to infinite (=0)
ethermintd start --json-rpc.gas-cap 0
```

## CORS

If accessing the RPC from a browser, CORS will need to be enabled with the appropriate domain set. Otherwise, JavaScript calls are limit by the same-origin policy and requests will fail:

```bash
ethermintd start --json-rpc.enable-unsafe-cors
```
