// SPDX-License-Identifier: MIT
pragma solidity 0.8.10;

import "@openzeppelin/contracts/token/ERC20/ERC20.sol";

contract TestERC20A is ERC20 {
    constructor() public ERC20("TestERC20", "Test") {
        _mint(msg.sender, 100000000000000000000000000);
    }

    // 0x9ffb86a5
    function revertWithMsg() public pure {
        revert("Function has been reverted");
    }

    // 0x3246485d
    function revertWithoutMsg() public pure {
        revert();
    }
}
