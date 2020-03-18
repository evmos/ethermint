#!/bin/bash
rm -rf ~/.emintcli
rm -rf ~/.emintd

make install

emintcli config keyring-backend test
emintcli keys add mykey

# Set moniker and chain-id for Ethermint (Moniker can be anything, chain-id must be an integer)
emintd init mymoniker --chain-id 8

# Set up config for CLI
emintcli config chain-id 8
emintcli config output json
emintcli config indent true
emintcli config trust-node true

# Allocate genesis accounts (cosmos formatted addresses)
emintd add-genesis-account $(emintcli keys show mykey -a) 1000000000000000000photon,1000000000000000000stake

# Sign genesis transaction
emintd gentx --name mykey --keyring-backend test

# Collect genesis tx
emintd collect-gentxs

# Run this to ensure everything worked and that the genesis file is setup correctly
emintd validate-genesis

# Start the node (remove the --pruning=nothing flag if historical queries are not needed)
emintd start --pruning=nothing