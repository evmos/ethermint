#!/bin/bash

CHAINID="ethermint-1337"
MONIKER="localtestnet"

VAL_KEY="localkey"
VAL_MNEMONIC="gesture inject test cycle original hollow east ridge hen combine junk child bacon zero hope comfort vacuum milk pitch cage oppose unhappy lunar seat"

USER1_KEY="user1"
USER1_MNEMONIC="copper push brief egg scan entry inform record adjust fossil boss egg comic alien upon aspect dry avoid interest fury window hint race symptom"

USER2_KEY="user2"
USER2_MNEMONIC="maximum display century economy unlock van census kite error heart snow filter midnight usage egg venture cash kick motor survey drastic edge muffin visual"

# remove existing daemon and client
rm -rf ~/.ethermint*

ethermintcli config keyring-backend test

# Set up config for CLI
ethermintcli config chain-id $CHAINID
ethermintcli config output json
ethermintcli config indent true
ethermintcli config trust-node true

# Import keys from mnemonics
echo $VAL_MNEMONIC | ethermintcli keys add $VAL_KEY --recover
echo $USER1_MNEMONIC | ethermintcli keys add $USER1_KEY --recover
echo $USER2_MNEMONIC | ethermintcli keys add $USER2_KEY --recover

# Set moniker and chain-id for Ethermint (Moniker can be anything, chain-id must be an integer)
ethermintd init $MONIKER --chain-id $CHAINID

# Allocate genesis accounts (cosmos formatted addresses)
ethermintd add-genesis-account $(ethermintcli keys show $VAL_KEY -a) 1000000000000000000000aphoton,10000000000000000stake
ethermintd add-genesis-account $(ethermintcli keys show $USER1_KEY -a) 1000000000000000000000aphoton,10000000000000000stake
ethermintd add-genesis-account $(ethermintcli keys show $USER2_KEY -a) 1000000000000000000000aphoton,10000000000000000stake

# Sign genesis transaction
ethermintd gentx --name $VAL_KEY --keyring-backend test

# Collect genesis tx
ethermintd collect-gentxs

# Enable faucet
cat  $HOME/.ethermintd/config/genesis.json | jq '.app_state["faucet"]["enable_faucet"]=true' >  $HOME/.ethermintd/config/tmp_genesis.json && mv $HOME/.ethermintd/config/tmp_genesis.json $HOME/.ethermintd/config/genesis.json

echo -e '\n\ntestnet faucet enabled'
echo -e 'to transfer tokens to your account address use:'
echo -e "ethermintcli tx faucet request 100aphoton --from $VAL_KEY\n"


# Run this to ensure everything worked and that the genesis file is setup correctly
ethermintd validate-genesis

# Command to run the rest server in a different terminal/window
echo -e '\nrun the following command in a different terminal/window to run the REST server and JSON-RPC:'
echo -e "ethermintcli rest-server --laddr \"tcp://localhost:8545\" --wsport 8546 --unlock-key $VAL_KEY,$USER1_KEY,$USER2_KEY --chain-id $CHAINID --trace\n"

# Start the node (remove the --pruning=nothing flag if historical queries are not needed)
ethermintd start --pruning=nothing --rpc.unsafe --log_level "main:info,state:info,mempool:info" --trace
