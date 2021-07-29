#!/bin/sh

# prepare sloc v0.5.17 in PATH
solc --combined-json bin,abi --allow-paths . ./tests/solidity/suites/staking/contracts/test/mocks/StandardTokenMock.sol \
    | jq ".contracts.\"./tests/solidity/suites/staking/contracts/test/mocks/StandardTokenMock.sol:StandardTokenMock\"" \
    > x/evm/keeper/ERC20Contract.json
