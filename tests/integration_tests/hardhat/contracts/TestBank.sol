// SPDX-License-Identifier: MIT
pragma solidity >0.6.6;

contract TestBank {
    address constant bankContract = 0x0000000000000000000000000000000000000064;
    function nativeMint(uint amount) public {
        (bool result, bytes memory _data) = bankContract.call(abi.encodeWithSignature(
            "mint(address,uint256)", msg.sender, amount
        ));
        require(result, "native call");
    }
    function nativeBalanceOf(address addr) public returns (uint) {
        (bool result, bytes memory data) = bankContract.call(abi.encodeWithSignature(
            "balanceOf(address,address)", address(this), addr
        ));
        require(result, "native call");
        return abi.decode(data, (uint));
    }
    function nativeMintRevert(uint amount) public {
        nativeMint(amount);
        revert("test");
    }
}
