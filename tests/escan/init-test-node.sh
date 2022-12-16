#!/bin/bash

set -eu

ENV_FILE='./tests/escan/env.sh'
if [[ ! -f "$ENV_FILE" ]]; then
  echo 'Wrong working directory, must be run from root dir of this project'
  exit 1
fi

source "$ENV_FILE"

MONIKER="localtestnet"

# localKey address 0x7cb61d4117ae31a12e393a1cfa3bac666481d02e
VAL_KEY="localkey"
VAL_MNEMONIC="gesture inject test cycle original hollow east ridge hen combine junk child bacon zero hope comfort vacuum milk pitch cage oppose unhappy lunar seat"

# remove existing daemon and client
rm -rf "$HOME_DIR"

# Import keys from mnemonics
echo "$VAL_MNEMONIC"   | ethermintd keys add "$VAL_KEY"   --recover --keyring-backend test --algo "eth_secp256k1" --home "$HOME_DIR"
echo "$USER1_MNEMONIC" | ethermintd keys add "$USER1_KEY" --recover --keyring-backend test --algo "eth_secp256k1" --home "$HOME_DIR"
echo "$USER2_MNEMONIC" | ethermintd keys add "$USER2_KEY" --recover --keyring-backend test --algo "eth_secp256k1" --home "$HOME_DIR"
echo "$USER3_MNEMONIC" | ethermintd keys add "$USER3_KEY" --recover --keyring-backend test --algo "eth_secp256k1" --home "$HOME_DIR"
echo "$USER4_MNEMONIC" | ethermintd keys add "$USER4_KEY" --recover --keyring-backend test --algo "eth_secp256k1" --home "$HOME_DIR"

ethermintd init "$MONIKER" --chain-id "$CHAIN_ID" --home "$HOME_DIR"

# Set gas limit in genesis
cat "$HOME_DIR/config/genesis.json" | jq '.consensus_params["block"]["max_gas"]="10000000"' > "$HOME_DIR/config/genesis.json.tmp" && mv "$HOME_DIR/config/genesis.json.tmp" "$HOME_DIR/config/genesis.json"

# Reduce the block time to 1s
sed -i -e '/^timeout_commit =/ s/= .*/= "850ms"/' "$HOME_DIR/config/config.toml"

# Allocate genesis accounts (cosmos formatted addresses)
ethermintd add-genesis-account "$(ethermintd keys show "$VAL_KEY"   -a --keyring-backend test --home "$HOME_DIR")" 1000000000000000000000aphoton,1000000000000000000stake --keyring-backend test --home "$HOME_DIR"
ethermintd add-genesis-account "$(ethermintd keys show "$USER1_KEY" -a --keyring-backend test --home "$HOME_DIR")" 1000000000000000000000aphoton,1000000000000000000stake --keyring-backend test --home "$HOME_DIR"
ethermintd add-genesis-account "$(ethermintd keys show "$USER2_KEY" -a --keyring-backend test --home "$HOME_DIR")" 1000000000000000000000aphoton,1000000000000000000stake --keyring-backend test --home "$HOME_DIR"
ethermintd add-genesis-account "$(ethermintd keys show "$USER3_KEY" -a --keyring-backend test --home "$HOME_DIR")" 1000000000000000000000aphoton,1000000000000000000stake --keyring-backend test --home "$HOME_DIR"
ethermintd add-genesis-account "$(ethermintd keys show "$USER4_KEY" -a --keyring-backend test --home "$HOME_DIR")" 1000000000000000000000aphoton,1000000000000000000stake --keyring-backend test --home "$HOME_DIR"

# Sign genesis transaction
ethermintd gentx "$VAL_KEY" 1000000000000000000stake --amount=1000000000000000000000aphoton --chain-id "$CHAIN_ID" --keyring-backend test --home "$HOME_DIR"

# Collect genesis tx
ethermintd collect-gentxs --home "$HOME_DIR"

# Run this to ensure everything worked and that the genesis file is setup correctly
ethermintd validate-genesis --home "$HOME_DIR"

# Start the node (remove the --pruning=nothing flag if historical queries are not needed)
ethermintd start --pruning=nothing --rpc.unsafe --keyring-backend test --log_level error --json-rpc.api eth,txpool,personal,net,debug,web3,escan --api.enable --home "$HOME_DIR"
