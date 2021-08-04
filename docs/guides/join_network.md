<!--
order: 5
-->

# Joining a Testnet

This document outlines the steps to join an existing testnet

## Steps

1. Install the Ethermint binary ethermintd

    ```bash
    go install https://github.com/tharsis/ethermint
    ```

2. Create an Ethermint account

    ```bash
    ethermintd keys add <keyname>
    ```

3. Copy genesis file

    Follow this [link](https://gist.github.com/araskachoi/43f86f3edff23729b817e8b0bb86295a) and copy it over to the directory ~/.ethermintd/config/genesis.json

4. Add peers

    Edit the file located in ~/.ethermintd/config/config.toml and edit line 350 (persistent_peers) to the following

    ```toml
    "05aa6587f07a0c6a9a8213f0138c4a76d476418a@18.204.206.179:26656,13d4a1c16d1f427988b7c499b6d150726aaf3aa0@3.86.104.251:26656,a00db749fa51e485c8376276d29d599258052f3e@54.210.246.165:26656"
    ```

5. Validate genesis and start the Ethermint network

    ```bash
    ethermintd validate-genesis

    ethermintd start --pruning=nothing --rpc.unsafe --log_level "main:info,state:info,mempool:info" --trace
    ```
