// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

contract Inner {
    event TestEvent0(uint256);
    event TestEvent1(uint256);

    function test() public returns (uint256) {
        emit TestEvent0(42);
        emit TestEvent1(42);
        return 42;
    }
}

// An contract that do lots of message calls
contract TestMessageCall {
    Inner _inner;

    constructor() public {
        _inner = new Inner();
    }

    function test(uint iterations) public returns (uint256) {
        uint256 n = 0;
        for (uint i = 0; i < iterations; i++) {
            n += _inner.test();
        }
        return n;
    }

    function inner() public view returns (address) {
        return address(_inner);
    }
}
