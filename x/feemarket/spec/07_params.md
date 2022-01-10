<!--
order: 7 -->

# Parameters

The `x/feemarket` module contains the following parameters:

| Key                           | Type   | Example     |  Description |
| ----------------------------- | ------ | ----------- |------------- |
| NoBaseFee                     | bool   | false       | control the base fee adjustment |
| BaseFeeChangeDenominator      | uint32 | 8           | parameter in the fee adjustment|
| ElasticityMultiplier          | uint32 | 2           | parameter in the fee adjustment|
| InitialBaseFee                | uint32 | 1000000000  | initial base fee when starting the adjustment |
| EnableHeight                  | uint32 | 0           | control the base fee adjustment|