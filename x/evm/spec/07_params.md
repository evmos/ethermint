<!--
order: 7
-->

# Parameters

The evm module contains the following parameters:

| Key            | Type   | Default Value |
|----------------|--------|---------------|
| `EVMDenom`     | string | `"aphoton"`   |
| `EnableCreate` | bool   | `true`        |
| `EnableCall`   | bool   | `true`        |
| `ExtraEIPs`    | []int  | TBD           |

## EVM denom

The evm denomination parameter defines the token denomination used on the EVM state transitions and
gas consumption for EVM messages.

The EVM Denom is used on the following cases:

* `AnteHandler`: for calculating sufficient balance to pay for gas cost or transaction fees.
* `journal`: to revert certain state executions (`balanceChange` and `suicideChange`).
* `stateObject`: to track the `evm_denom` balance of the object account.
* `CommitStateDB`: to update account balance from an existing state object.

For example, on Ethereum, the `evm_denom` would be `ETH`. In the case of Ethermint, the default denomination is the [atto photon](./../../../docs/basics/photon.md). In terms of precision, the `PHOTON` and `ETH` share the same value, _i.e_ `1 PHOTON = 10^18 atto photon` and `1 ETH = 10^18 wei`.

::: danger
SDK applications that want to import the EVM module as a dependency will need to set their own `evm_denom` (i.e not `"aphoton"`).
:::

## Enable Create

The enable create parameter toggles state transitions that use the `vm.Create` function. When the
parameter is disabled, it will prevent all contract creation functionality.

## Enable Transfer

The enable transfer toggles state transitions that use the `vm.Call` function. When the parameter is
disabled, it will prevent transfers between accounts and executing a smart contract call.

## Extra EIPs

The extra EIPs parameter defines the set of activateable Ethereum Improvement Proposals ([EIPs](https://ethereum.org/en/eips/))
on the Ethereum VM `Config` that apply custom jump tables.
The supported activateable EIPS are:

* [EIP 1344](https://eips.ethereum.org/EIPS/eip-1344)
* [EIP 1884](https://eips.ethereum.org/EIPS/eip-1884)
* [EIP 2200](https://eips.ethereum.org/EIPS/eip-2200)
* [EIP 2315](https://eips.ethereum.org/EIPS/eip-2315)
* [EIP 2929](https://eips.ethereum.org/EIPS/eip-2929)
