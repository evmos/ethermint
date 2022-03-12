<!--
order: 1
-->

# Concepts

## Base fee

The base fee is a global base fee defined at the consensus level. It is adjusted for each block based on the total gas used in the previous block and gas target (block gas limit divided by elasticity multiplier) 

- it increases when blocks are above the gas target
- it decreases when blocks are below the gas target

Unlike the Cosmos SDK local `minimal-gas-prices`, this value is stored as a module parameter which provides a reliable value for validators to agree upon.

## Tip

In EIP-1559, the `tip` is a value that can be added to the `baseFee` in order to incentive transaction prioritization.

The transaction fee in Ethereum is calculated using the following the formula :

`transaction fee = (baseFee + tip) * gas units (limit)`

In Cosmos SDK there is no notion of prioritization, thus the tip for an EIP-1559 transaction in Ethermint should be zero (`MaxPriorityFeePerGas` JSON-RPC endpoint returns `0`)



## EIP-1559

A transaction pricing mechanism introduced in Ethereum that includes fixed-per-block network fee that is burned and dynamically expands/contracts block sizes to deal with transient congestion.

Transactions specify a maximum fee per gas they are willing to pay total (aka: max fee), which covers both the priority fee and the block's network fee per gas (aka: base fee)

Reference: [EIP1559](https://eips.ethereum.org/EIPS/eip-1559)

