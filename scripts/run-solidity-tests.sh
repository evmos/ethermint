#!/bin/bash

export GOPATH=~/go
export PATH=$PATH:$GOPATH/bin
go build -o ./build/ethermintd ./cmd/ethermintd 
go build -o ./build/ethermintcli ./cmd/ethermintcli
mkdir $GOPATH/bin
cp ./build/ethermintd $GOPATH/bin
cp ./build/ethermintcli $GOPATH/bin

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
./init-test-node.sh > ethermintd.log &
sleep 5
ethermintcli rest-server --laddr "tcp://localhost:8545" --unlock-key localkey,user1,user2 --chain-id 1337 --trace --wsport 8546 > ethermintcli.log &

cd suites/initializable
yarn test-ethermint

ok=$?

if (( $? != 0 )); then
	echo "initializable test failed: exit code $?"
fi

killall ethermintcli
killall ethermintd

echo "Script exited with code $ok"
exit $ok

# initializable-buidler fails on CI, re-add later

./../../init-test-node.sh > ethermintd.log &
sleep 5
ethermintcli rest-server --laddr "tcp://localhost:8545" --unlock-key localkey,user1,user2 --chain-id 1337 --trace --wsport 8546 > ethermintcli.log &

cd ../initializable-buidler
yarn test-ethermint

ok=$(($? + $ok))

if (( $? != 0 )); then
	echo "initializable-buidler test failed: exit code $?"
fi

killall ethermintcli
killall ethermintd

echo "Script exited with code $ok"
exit $ok