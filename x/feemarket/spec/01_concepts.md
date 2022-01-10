<!--
order: 1
-->

# Concepts

## Base fee

The base fee is a global base fee defined at the consensus level. It is adjusted for each block based on the total gas used in the previous block and gas target (block gas limit divided by elasticity multiplier) 

- it increases when blocks are above the gas target
- it decreases when blocks are below the gas target

Unlike the Cosmos SDK local `minimal-gas-prices`, this value is persisted in the KVStore which provides a reliable value for validators to agree upon.

## Tip

To be consistent with EIP-1559, the `tip` is a local value that each node can define and be added to the `baseFee`.

The transaction fee is calculated using the following the formula :

`transaction fee = (baseFee + tip) * gas units (limit)` 

## EIP-1559

A transaction pricing mechanism introduced in Ethereum that includes fixed-per-block network fee that is burned and dynamically expands/contracts block sizes to deal with transient congestion.

Transactions specify a maximum fee per gas they are willing to pay total (aka: max fee), which covers both the priority fee and the block's network fee per gas (aka: base fee)

Reference:
https://github.com/ethereum/EIPs/blob/master/EIPS/eip-1559.md

