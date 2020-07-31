#!/bin/sh
emintd --home /ethermint/node$ID/emintd/ start > emintd.log &
sleep 5
emintcli rest-server --laddr "tcp://localhost:8545" --chain-id 7305661614933169792 --trace > emintcli.log &
tail -f /dev/null