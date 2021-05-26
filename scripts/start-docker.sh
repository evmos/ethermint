#!/bin/bash

KEY="mykey"
CHAINID="ethermint-123"
MONIKER="mymoniker"
DATA_DIR=$(mktemp -d -t ethermint-datadir.XXXXX)

echo "create and add new keys"
./ethermintd keys add $KEY --home $DATA_DIR --no-backup --chain-id $CHAINID --algo "eth_secp256k1" --keyring-backend test
echo "init Ethermint with moniker=$MONIKER and chain-id=$CHAINID"
./ethermintd init $MONIKER --chain-id $CHAINID --home $DATA_DIR
echo "prepare genesis: Allocate genesis accounts"
./ethermintd add-genesis-account \
"$(./ethermintd keys show $KEY -a --home $DATA_DIR --keyring-backend test)" 1000000000000000000aphoton,1000000000000000000stake \
--home $DATA_DIR --keyring-backend test
echo "prepare genesis: Sign genesis transaction"
./ethermintd gentx $KEY 1000000000000000000stake --keyring-backend test --home $DATA_DIR --keyring-backend test --chain-id $CHAINID
echo "prepare genesis: Collect genesis tx"
./ethermintd collect-gentxs --home $DATA_DIR
echo "prepare genesis: Run validate-genesis to ensure everything worked and that the genesis file is setup correctly"
./ethermintd validate-genesis --home $DATA_DIR

echo "starting ethermint node $i in background ..."
./ethermintd start --pruning=nothing --rpc.unsafe \
--keyring-backend test --home $DATA_DIR \
>$DATA_DIR/node.log 2>&1 & disown

echo "started ethermint node"
tail -f /dev/null