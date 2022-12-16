#!/bin/bash

set -eu

ENV_FILE='./tests/escan/env.sh'
if [[ ! -f "$ENV_FILE" ]]; then
  echo 'Wrong working directory, must be run from root dir of this project'
  exit 1
fi

source "$ENV_FILE"

# Start the node (remove the --pruning=nothing flag if historical queries are not needed)
ethermintd start --pruning=nothing --rpc.unsafe --keyring-backend test --log_level error --json-rpc.api eth,txpool,personal,net,debug,web3 --api.enable --home "$HOME_DIR"
