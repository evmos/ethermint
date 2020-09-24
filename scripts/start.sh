#!/bin/sh
ethermintd --home /ethermint/node$ID/ethermintd/ start > ethermintd.log &
sleep 5
ethermintcli rest-server --laddr "tcp://localhost:8545" --chain-id "ethermint-7305661614933169792" --trace > ethermintcli.log &
tail -f /dev/null