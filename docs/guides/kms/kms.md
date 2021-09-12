<!--
order: 1
-->

# Tendermint KMS

[Tendermint KMS](https://github.com/iqlusioninc/tmkms) is a key management service that allows separating key management from Tendermint nodes. In addition it provides other advantages such as:

- Improved security and risk management policies
- Unified API and support for various HSM (hardware security modules)
- Double signing protection (software or hardware based)

It is recommended that the KMS service runs in a separate physical hosts.

## Building

Detailed build instructions can be found [here](https://github.com/iqlusioninc/tmkms#installation).

::: tip
When compiling the KMS, ensure you have enabled the applicable features:
:::

| Backend                 | Recommended Command line          |
|-------------------------|-----------------------------------|
| YubiHSM                 | `cargo build --features yubihsm`  |
| Ledger + Tendermint App | `cargo build --features ledgertm` |

## Configuration

A KMS can be configured using the following HSMs:

### Using a YubiHSM
  
Detailed information on how to setup a KMS with YubiHSM2 can be found [here](https://github.com/iqlusioninc/tmkms/blob/master/README.yubihsm.md)
