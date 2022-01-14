<!--
order: 2
-->

# State

The x/feemarket module keeps state in two primary objects:



|                  | Description                    | Key            | Value               | Store     |
| -----------      | ------------------------------ | ---------------| ------------------- | --------- |
| BlockGasUsed     | gas used in the block          | `[]byte{1}`    | `[]byte{gas_used}`  | KV        |
| BaseFee          | block's base fee               | `[]byte{2}`    | `[]byte(baseFee)`   | KV        |
