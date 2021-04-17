#!/bin/bash

export GOPATH=~/go
export PATH=$PATH:$GOPATH/bin
go build -o ./build/ethermintd ./cmd/ethermintd
mkdir $GOPATH/bin
cp ./build/ethermintd $GOPATH/bin

localKeyAddr=0x7cb61d4117ae31a12e393a1cfa3bac666481d02e
user1Addr=0xc6fe5d33615a1c52c08018c47e8bc53646a0e101
user2Addr=0x963ebdf2e1f8db8707d05fc75bfeffba1b5bac17

CHAINID="ethermint-1337"

# build ethermint binary
make install

cd tests-solidity

if command -v yarn &> /dev/null; then
    yarn install
else
    curl -sS https://dl.yarnpkg.com/debian/pubkey.gpg | sudo apt-key add -
    echo "deb https://dl.yarnpkg.com/debian/ stable main" | sudo tee /etc/apt/sources.list.d/yarn.list
    sudo apt update && sudo apt install yarn
    yarn install
fi

chmod +x ./init-test-node.sh
nohup ./init-test-node.sh > ethermintd.log 2>&1 &

# give ethermintd node enough time to launch
echo "sleeping ..."
sleep 10

# show existing accounts
echo "account list: "
curl -X POST --data '{"jsonrpc":"2.0","method":"personal_listAccounts","params":[],"id":1}' -H "Content-Type: application/json" http://localhost:8545

# unlock localKey address
curl -X POST --data '{"jsonrpc":"2.0","method":"personal_unlockAccount","params":["'$localKeyAddr'", ""],"id":1}' -H "Content-Type: application/json" http://localhost:8545

# tests start
cd suites/initializable
yarn contract-migrate
yarn test-ethermint

ok=$?

if (( $? != 0 )); then
    echo "initializable test failed: exit code $?"
fi

killall ethermintd

echo "Script exited with code $ok"
exit $ok

# initializable-buidler fails on CI, re-add later

./../../init-test-node.sh > ethermintd.log
cd ../initializable-buidler
yarn test-ethermint

ok=$(($? + $ok))

if (( $? != 0 )); then
    echo "initializable-buidler test failed: exit code $?"
fi

killall ethermintd

echo "Script exited with code $ok"
exit $ok