// SPDX-License-Identifier: MIT
pragma solidity >0.6.6;

contract TestBankCaller {
    uint256 state;

    function mint(address callee, uint amount) public {
        (bool success, bytes memory data) = callee.call(abi.encodeWithSignature(
            "moveToNativeRevert(uint256)", amount
        ));
        if (!success) {
            // ignore the error and move on
            state++;
        }
    }

    function getLastState() public view returns (uint256) {
        return state;
    }
}
