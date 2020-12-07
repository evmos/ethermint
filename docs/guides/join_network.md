<!--
order: 5
-->


# Joining Chainsafe's Public Testnet

This document outlines the steps to join the public testnet hosted by [Chainsafe](https://chainsafe.io). 

## Steps:
1. Install the Ethermint binaries (ethermintd & ethermint cli):
```
git clone https://github.com/cosmos/ethermint
cd ethermint
make install
```

2. Create an Ethermint account:
```
ethermintcli keys add <keyname>
```

3. Copy genesis file:
Follow this [link](https://gist.github.com/araskachoi/43f86f3edff23729b817e8b0bb86295a) and copy it over to the directory ~/.ethermintd/config/genesis.json

4. Add peers:
Edit the file located in ~/.ethermintd/config/config.toml and edit line 350 (persistent_peers) to the following;
```
"aad1e55aa61b9142e57db756b398bb769ec4132a@54.210.246.165:26656,519613343d7d0773fb99da6fffca858d8df33cd4@18.204.206.179:26656,64ffe6743666b0e186638013117f164c5cf16d21@3.86.104.251:26656"
```

5. Validate genesis and start the Ethermint network:
```
ethermintd validate-genesis
```
```
ethermintd start --pruning=nothing --rpc.unsafe --log_level "main:info,state:info,mempool:info" --trace
```
(we recommend running the command in the background for convenience)

6. Start the RPC server:
```
ethermintcli rest-server --laddr "tcp://localhost:8545" --unlock-key $KEY --chain-id etherminttestnet-1 --trace
```
where `$KEY` is the key name that was used in step 2.
(we recommend running the command in the background for convenience)

7. Request funds from the faucet:
You will need to know the Ethereum hex address, and it can be found with the following command:

```
curl -X POST --data '{"jsonrpc":"2.0","method":"eth_accounts","params":[],"id":1}' -H "Content-Type: application/json" http://localhost:8545
```
Using the output of the above command, you will then send the command with your valid Ethereum address:
```
curl --header "Content-Type: application/json" --request POST --data '{"address":"0xYourEthereumHexAddress"}' 3.95.21.91:3000
```

## Public Testnet Node RPC Endpoints

Node0: `54.210.246.165:8545`
Node1: `3.86.104.251:8545`
Node2: `18.204.206.179:8545`

example:
```
curl -X POST --data '{"jsonrpc":"2.0","method":"eth_chainId","params":[],"id":1}' -H "Content-Type: application/json" 54.210.246.165:8545
```