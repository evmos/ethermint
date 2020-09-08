#!/bin/bash

KEY="mykey"
TESTKEY="test"
CHAINID=123321
MONIKER="localtestnet"

# stop and remove existing daemon and client data and process(es)
rm -rf $PWD/.ethermint*
pkill -f "ethermint*"

type "ethermintd" 2> /dev/null || make build-ethermint
type "ethermintcli" 2> /dev/null || make build-ethermint

$PWD/build/ethermintcli config keyring-backend test

# Set up config for CLI
$PWD/build/ethermintcli config chain-id $CHAINID
$PWD/build/ethermintcli config output json
$PWD/build/ethermintcli config indent true
$PWD/build/ethermintcli config trust-node true

# if $KEY exists it should be deleted
$PWD/build/ethermintcli keys add $KEY

# Set moniker and chain-id for Ethermint (Moniker can be anything, chain-id must be an integer)
$PWD/build/ethermintd init $MONIKER --chain-id $CHAINID

# Change parameter token denominations to aphoton
cat $HOME/.ethermintd/config/genesis.json | jq '.app_state["staking"]["params"]["bond_denom"]="aphoton"' > $HOME/.ethermintd/config/tmp_genesis.json && mv $HOME/.ethermintd/config/tmp_genesis.json $HOME/.ethermintd/config/genesis.json
cat $HOME/.ethermintd/config/genesis.json | jq '.app_state["crisis"]["constant_fee"]["denom"]="aphoton"' > $HOME/.ethermintd/config/tmp_genesis.json && mv $HOME/.ethermintd/config/tmp_genesis.json $HOME/.ethermintd/config/genesis.json
cat $HOME/.ethermintd/config/genesis.json | jq '.app_state["gov"]["deposit_params"]["min_deposit"][0]["denom"]="aphoton"' > $HOME/.ethermintd/config/tmp_genesis.json && mv $HOME/.ethermintd/config/tmp_genesis.json $HOME/.ethermintd/config/genesis.json
cat $HOME/.ethermintd/config/genesis.json | jq '.app_state["mint"]["params"]["mint_denom"]="aphoton"' > $HOME/.ethermintd/config/tmp_genesis.json && mv $HOME/.ethermintd/config/tmp_genesis.json $HOME/.ethermintd/config/genesis.json

# Enable faucet
cat $HOME/.ethermintd/config/genesis.json | jq '.app_state["faucet"]["enable_faucet"]=true' >  $HOME/.ethermintd/config/tmp_genesis.json && mv $HOME/.ethermintd/config/tmp_genesis.json $HOME/.ethermintd/config/genesis.json

# Allocate genesis accounts (cosmos formatted addresses)
$PWD/build/ethermintd add-genesis-account "$("$PWD"/build/ethermintcli keys show "$KEY$i" -a)" 100000000000000000000aphoton

# Sign genesis transaction
$PWD/build/ethermintd gentx --name $KEY --amount=1000000000000000000aphoton --keyring-backend test

# Collect genesis tx
$PWD/build/ethermintd collect-gentxs

# Run this to ensure everything worked and that the genesis file is setup correctly
$PWD/build/ethermintd validate-genesis

# Start the node (remove the --pruning=nothing flag if historical queries are not needed) in background and log to file
$PWD/build/ethermintd start --pruning=nothing --rpc.unsafe --log_level "main:info,state:info,mempool:info" --trace > ethermintd.log &

sleep 1

# Start the rest server with unlocked faucet key in background and log to file
$PWD/build/ethermintcli rest-server --laddr "tcp://localhost:8545" --unlock-key $KEY --chain-id $CHAINID --trace > ethermintcli.log &

solcjs --abi $PWD/tests-solidity/suites/basic/contracts/Counter.sol --bin -o $PWD/tests-solidity/suites/basic/counter
mv $PWD/tests-solidity/suites/basic/counter/*.abi $PWD/tests-solidity/suites/basic/counter/counter_sol.abi
mv $PWD/tests-solidity/suites/basic/counter/*.bin $PWD/tests-solidity/suites/basic/counter/counter_sol.bin

ACCT=$(curl --fail --silent -X POST --data '{"jsonrpc":"2.0","method":"eth_accounts","params":[],"id":1}' -H "Content-Type: application/json" http://localhost:8545 | grep -o '\0x[^"]*' 2>&1)

echo $ACCT

curl -X POST --data '{"jsonrpc":"2.0","method":"personal_unlockAccount","params":["'$ACCT'", ""],"id":1}' -H "Content-Type: application/json" http://localhost:8545

PRIVKEY="$("$PWD"/build/ethermintcli keys unsafe-export-eth-key $KEY)"

echo $PRIVKEY

## need to get the private key from the account in order to check this functionality.
cd tests-solidity/suites/basic/ && go get && go run main.go $ACCT
