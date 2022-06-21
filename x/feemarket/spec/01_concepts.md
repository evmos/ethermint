<!--
order: 1
-->

# Concepts

## EIP-1559

[EIP-1559](https://github.com/ethereum/EIPs/blob/master/EIPS/eip-1559.md) describes a pricing mechanism that was proposed on Ethereum to improve to calculation of transaction fees. It includes a fixed-per-block network fee that is burned and dynamically expands/contracts block sizes to deal with peaks of network congestion.

Before EIP-1559 the transaction fee is calculated with

```
fee = gasPrice * gasLimit
```

, where `gasPrice` is the price per gas and `gasLimit` describes the amount of gas required to perform the transaction. The more complex operations a transaction requires, the higher the gasLimit (See [Executing EVM bytecode](https://docs.evmos.org/modules/evm/01_concepts.html#executing-evm-bytecode)). To submit a transaction, the signer needs to specify the `gasPrice`.

With EIP-1559 enabled, the transaction fee is calculated with

```
fee = (baseFee + priorityTip) * gasLimit
```

, where `baseFee` is the fixed-per-block network fee per gas and `priorityTip` is an additional fee per gas that can be set optionally. Note, that both the base fee and the priority tip are a gas prices. To submit a transaction with EIP-1559, the signer needs to specify the `gasFeeCap` a maximum fee per gas they are willing to pay total and optionally the `priorityFee` , which covers both the priority fee and the block's network fee per gas (aka: base fee)

Reference: [EIP1559](https://eips.ethereum.org/EIPS/eip-1559)

## Base fee

The base fee per unit gas is a global gas price defined at the consensus level. It is adjusted for each block based on the total gas used in the previous block and gas target (`block gas limit / elasticity multiplier`)

- it increases when blocks are above the gas target
- it decreases when blocks are below the gas target

Unlike the Cosmos SDK local `minimal-gas-prices`, this value is stored as a module parameter which provides a reliable value for validators to agree upon.

## Tip

In EIP-1559, the `tip` is a value that can be added to the `baseFee` in order to incentive transaction prioritization.

The transaction fee in Ethereum is calculated using the following the formula :

`transaction fee = (baseFee + tip) * gas units (limit)`

In Cosmos SDK there is no notion of prioritization, thus the tip for an EIP-1559 transaction in Ethermint should be zero (`MaxPriorityFeePerGas` JSON-RPC endpoint returns `0`)

## Global Minimum Gas Price

The minimum gas price needed for transactions to be processed. It applies to both Cosmos and EVM transactions. Governance can change this `feemarket` module parameter value. If the effective gas price or the minimum gas price is lower than the global `MinGasPrice` (`min-gas-price (local) < MinGasPrice (global) OR EffectiveGasPrice < MinGasPrice`), then `MinGasPrice` is used as a lower bound. If transactions are rejected due to having a gas price lower than `MinGasPrice`, users need to resend the transactions with a gas price higher than `MinGasPrice`. In the case of EIP-1559 (dynamic fee transactions), users must increase the priority fee for their transactions to be valid.
