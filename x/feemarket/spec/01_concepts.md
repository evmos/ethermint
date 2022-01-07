<!--
order: 1
-->

# Concepts

## Base fee

The base fee is a global base fee defined at the consensus level. It is adjusted for each block based on the total gas used in the previous block.

Unlike the Cosmos SDK local `minimal-gas-prices`, this value is persisted in the KVStore which make a reliable value for validators to agree upon.

## Tip

To be consistent with EIP-1559, the `tip` is a local value that each node can define and be added to the `baseFee`.

The transaction fee is calculated using the following the formula :

`transaction fee = (baseFee + tip) * gas units (limit)` 

