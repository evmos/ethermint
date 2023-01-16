// SPDX-License-Identifier: MIT
pragma solidity > 0.8.0;

contract Inner {
    event TestEvent(uint256);
    function test() public returns (uint256) {
        emit TestEvent(42);
        return 42;
    }
}

// An contract that do lots of message calls
contract TestMessageCall {
    Inner _inner;
    constructor() {
        _inner = new Inner();
    }

    function test(uint iterations) public returns (uint256) {
        uint256 n = 0;
        for (uint i=0; i < iterations; i++) {
            n += _inner.test();
        }
        return n;
    }
}
