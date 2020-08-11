#!/bin/bash

# "stable" mode tests assume data is static
# "live" mode tests assume data dynamic

SCRIPT=$(basename ${BASH_SOURCE[0]})
TEST=""
QTD=1
SLEEP_TIMEOUT=5
TEST_QTD=1

#PORT AND RPC_PORT 3 initial digits, to be concat with a suffix later when node is initialized
RPC_PORT="854"
IP_ADDR="0.0.0.0"
MODE="rpc"

KEY="mykey"
CHAINID=8
MONIKER="mymoniker"

## default port prefixes for ethermintd
NODE_P2P_PORT="2660"
NODE_PORT="2663"
NODE_RPC_PORT="2666"

usage() {
  echo "Usage: $SCRIPT"
  echo "Optional command line arguments"
  echo "-t <string>  -- Test to run. eg: rpc"
  echo "-q <number>  -- Quantity of nodes to run. eg: 3"
  echo "-z <number>  -- Quantity of nodes to run tests against eg: 3"
  echo "-s <number>  -- Sleep between operations in secs. eg: 5"
  exit 1
}

while getopts "h?t:q:z:s:" args; do
case $args in
    h|\?)
      usage;
      exit;;
    t ) TEST=${OPTARG};;
    q ) QTD=${OPTARG};;
    z ) TEST_QTD=${OPTARG};;
    s ) SLEEP_TIMEOUT=${OPTARG};;
  esac
done

set -euxo pipefail

DATA_DIR=$(mktemp -d -t ethermint-datadir.XXXXX)

if [[ ! "$DATA_DIR" ]]; then
  echo "Could not create $DATA_DIR"
  exit 1
fi

DATA_CLI_DIR=$(mktemp -d -t ethermint-cli-datadir.XXXXX)

if [[ ! "$DATA_CLI_DIR" ]]; then
  echo "Could not create $DATA_CLI_DIR"
  exit 1
fi

# Compile ethermint
echo "compiling ethermint"
make build-ethermint

# PID array declaration
arr=()

# PID arraycli declaration
arrcli=()

init_func() {
  echo "create and add new keys"
  "$PWD"/build/ethermintcli config keyring-backend test --home "$DATA_CLI_DIR$i"
  "$PWD"/build/ethermintcli keys add $KEY"$i" --home "$DATA_CLI_DIR$i" --no-backup --chain-id $CHAINID --algo "eth_secp256k1"
  echo "init Ethermint with moniker=$MONIKER and chain-id=$CHAINID"
  "$PWD"/build/ethermintd init $MONIKER --chain-id $CHAINID --home "$DATA_DIR$i"
  echo "init ethermintcli with chain-id=$CHAINID and config it trust-node true"
  "$PWD"/build/ethermintcli config chain-id $CHAINID --home "$DATA_CLI_DIR$i"
  "$PWD"/build/ethermintcli config output json --home "$DATA_CLI_DIR$i"
  "$PWD"/build/ethermintcli config indent true --home "$DATA_CLI_DIR$i"
  "$PWD"/build/ethermintcli config trust-node true --home "$DATA_CLI_DIR$i"
  echo "prepare genesis: Allocate genesis accounts"
  "$PWD"/build/ethermintd add-genesis-account \
  "$("$PWD"/build/ethermintcli keys show "$KEY$i" -a --home "$DATA_CLI_DIR$i" )" 1000000000000000000photon,1000000000000000000stake \
  --home "$DATA_DIR$i" --home-client "$DATA_CLI_DIR$i"
  echo "prepare genesis: Sign genesis transaction"
  "$PWD"/build/ethermintd gentx --name $KEY"$i" --keyring-backend test --home "$DATA_DIR$i" --home-client "$DATA_CLI_DIR$i"
  echo "prepare genesis: Collect genesis tx"
  "$PWD"/build/ethermintd collect-gentxs --home "$DATA_DIR$i"
  echo "prepare genesis: Run validate-genesis to ensure everything worked and that the genesis file is setup correctly"
  "$PWD"/build/ethermintd validate-genesis --home "$DATA_DIR$i"
}

start_func() {
  echo "starting ethermint node $i in background ..."
  "$PWD"/build/ethermintd start --pruning=nothing --rpc.unsafe --log_level "main:info,state:info,mempool:info" \
  --p2p.laddr tcp://$IP_ADDR:$NODE_P2P_PORT"$i" --address tcp://$IP_ADDR:$NODE_PORT"$i" --rpc.laddr tcp://$IP_ADDR:$NODE_RPC_PORT"$i" \
  --home "$DATA_DIR$i" \
  >"$DATA_DIR"/node"$i".log 2>&1 & disown

  ETHERMINT_PID=$!
  echo "started ethermint node, pid=$ETHERMINT_PID"
  # add PID to array
  arr+=("$ETHERMINT_PID")
}

start_cli_func() {
  echo "starting ethermint node $i in background ..."
  "$PWD"/build/ethermintcli rest-server --unlock-key $KEY"$i" --chain-id $CHAINID --trace \
  --laddr "tcp://localhost:$RPC_PORT$i" --node tcp://$IP_ADDR:$NODE_RPC_PORT"$i" \
  --home "$DATA_CLI_DIR$i" --read-timeout 30 --write-timeout 30 \
  >"$DATA_CLI_DIR"/cli"$i".log 2>&1 & disown

  ETHERMINT_CLI_PID=$!
  echo "started ethermintcli node, pid=$ETHERMINT_CLI_PID"
  # add PID to array
  arrcli+=("$ETHERMINT_CLI_PID")
}

# Run node with static blockchain database
# For loop N times
for i in $(seq 1 "$QTD"); do
  init_func "$i"
  start_func "$i"
  sleep 1
  start_cli_func "$i"
  echo "sleeping $SLEEP_TIMEOUT seconds for startup"
  sleep "$SLEEP_TIMEOUT"
  echo "done sleeping"
done

echo "sleeping $SLEEP_TIMEOUT seconds before running tests ... "
sleep "$SLEEP_TIMEOUT"
echo "done sleeping"

set +e

if [[ -z $TEST || $TEST == "rpc" ]]; then

  for i in $(seq 1 "$TEST_QTD"); do
    HOST_RPC=http://$IP_ADDR:$RPC_PORT"$i"
    echo "going to test ethermint node $HOST_RPC ..."
    MODE=$MODE HOST=$HOST_RPC go test ./tests/... -timeout=300s -v -short

    RPC_FAIL=$?
  done

fi

stop_func() {
  ETHERMINT_PID=$i
  echo "shutting down node, pid=$ETHERMINT_PID ..."

  # Shutdown ethermint node
  kill -9 "$ETHERMINT_PID"
  wait "$ETHERMINT_PID"
}


for i in "${arrcli[@]}"; do
  stop_func "$i"
done

for i in "${arr[@]}"; do
  stop_func "$i"
done

if [[ (-z $TEST || $TEST == "rpc") && $RPC_FAIL -ne 0 ]]; then
  exit $RPC_FAIL
else
  exit 0
fi
