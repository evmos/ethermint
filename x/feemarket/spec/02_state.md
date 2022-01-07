<!--
order: 2
-->

# State

## Block Gas Used

The total gas used by current block is stored in the KVStore at `EndBlock`.

It is initialized to `block_gas` defined in the genesis.

## Base Fee

### NoBaseFee

We introduce two parameters : `NoBaseFee`and `EnableHeight`

In case `NoBaseFee = true` or `height < EnableHeight`, the base fee value will be equal to `base_fee` defined in the genesis.

Those parameters allow us to introduce a static base fee or activate the base fee at a later stage.

### Calculation

The base fee is initialized at `EnableHeight` to the `InitialBaseFee` value defined in the genesis file.

The base fee is after adjusted according to the total gas used in the previous block.

```golang
parent_gas_target = parent_gas_limit / ELASTICITY_MULTIPLIER

if EnableHeight == block.number
    base_fee = INITIAL_BASE_FEE
else if parent_gas_used == parent_gas_target:
    base_fee = parent_base_fee
else if parent_gas_used > parent_gas_target:
    gas_used_delta = parent_gas_used - parent_gas_target
    base_fee_delta = max(parent_base_fee * gas_used_delta / parent_gas_target / BASE_FEE_MAX_CHANGE_DENOMINATOR, 1)
    base_fee = parent_base_fee + base_fee_delta
else:
    gas_used_delta = parent_gas_target - parent_gas_used
    base_fee_delta = parent_base_fee * gas_used_delta / parent_gas_target / BASE_FEE_MAX_CHANGE_DENOMINATOR
    base_fee = parent_base_fee - base_fee_delta

```