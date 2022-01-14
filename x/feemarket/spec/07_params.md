<!--
order: 7 -->

# Parameters

The `x/feemarket` module contains the following parameters:

| Key                           | Type   | Default Values     |  Description |
| ----------------------------- | ------ | ----------- |------------- |
| NoBaseFee                     | bool   | false       | control the base fee adjustment |
| BaseFeeChangeDenominator      | uint32 | 8           | bounds the amount the base fee that can change between blocks |
| ElasticityMultiplier          | uint32 | 2           | bounds the maximum gas limit an EIP-1559 block may have |
| InitialBaseFee                | uint32 | 1000000000  | initial base fee for EIP-1559 blocks |
| EnableHeight                  | uint32 | 0           | height which enable fee adjustment |