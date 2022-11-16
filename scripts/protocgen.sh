#!/usr/bin/env bash

set -eo pipefail

# get protoc executions
go get github.com/regen-network/cosmos-proto/protoc-gen-gocosmos 2>/dev/null

# get cosmos sdk
go get github.com/cosmos/cosmos-sdk 2>/dev/null

echo "Generating gogo proto code"
proto_dirs=$(find ./proto -path -prune -o -name '*.proto' -print0 | xargs -0 -n1 dirname | sort | uniq)
for dir in $proto_dirs; do
  for file in $(find "${dir}" -maxdepth 1 -name '*.proto'); do
    echo "File $file"
    if grep go_package $file &>/dev/null; then
      echo " >> generating"
      buf generate --template proto/buf.gen.gogo.yaml $file
    fi
  done
done

# TODO: command to generate docs using protoc-gen-doc was deleted here

# move proto files to the right places
cp -r github.com/evmos/ethermint/* ./
rm -rf github.com
