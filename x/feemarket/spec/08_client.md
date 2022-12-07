<!--
order: 8 -->

# Client

## CLI

A user can query and interact with the `feemarket` module using the CLI.

### Queries

The `query` commands allow users to query `feemarket` state.

```go
entangled query feemarket --help
```

#### Base Fee

The `base-fee` command allows users to query the block base fee by height.

```
entangled query feemarket base-fee [flags]
```

Example:

```
entangled query feemarket base-fee ...
```

Example Output:

```
base_fee: "512908936"
```

#### Block Gas

The `block-gas` command allows users to query the block gas by height.

```
entangled query feemarket block-gas [flags]
```

Example:

```
entangled query feemarket block-gas ...
```

Example Output:

```
gas: "21000"
```

#### Params

The `params` command allows users to query the module params.

```
entangled query params subspace [subspace] [key] [flags]
```

Example:

```
entangled query params subspace feemarket ElasticityMultiplier ...
```

Example Output:

```
key: ElasticityMultiplier
subspace: feemarket
value: "2"
```

## gRPC

### Queries

| Verb   | Method                                               | Description                                                                |
| ------ | ---------------------------------------------------- | -------------------------------------------------------------------------- |
| `gRPC`  | `ethermint.feemarket.v1.Query/Params`               | Get the module params                                                      |
| `gRPC`  | `ethermint.feemarket.v1.Query/BaseFee`              | Get the block base fee                                                     |
| `gRPC`  | `ethermint.feemarket.v1.Query/BlockGas`             | Get the block gas used                                                     |
| `GET`  | `/feemarket/evm/v1/params`                           | Get the module params                                                      |
| `GET`  | `/feemarket/evm/v1/base_fee`                         | Get the block base fee                                                     |
| `GET`  | `/feemarket/evm/v1/block_gas`                        | Get the block gas used                                                     |
