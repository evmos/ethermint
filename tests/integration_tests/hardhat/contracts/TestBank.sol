// SPDX-License-Identifier: MIT
pragma solidity >0.6.6;

contract TestBank {
    address constant bankContract = 0x0000000000000000000000000000000000000064;
    function nativeMint(uint256 amount) public {
        (bool result, ) = bankContract.call(abi.encodeWithSignature(
            "mint(address,uint256)", msg.sender, amount
        ));
        require(result, "native call");
    }
    function nativeBurn(uint256 amount) public {
        (bool result, ) = bankContract.call(abi.encodeWithSignature(
            "burn(address,uint256)", msg.sender, amount
        ));
        require(result, "native call");
    }
    function nativeBalanceOf(address addr) public returns (uint256) {
        (bool result, bytes memory data) = bankContract.call(abi.encodeWithSignature(
            "balanceOf(address,address)", address(this), addr
        ));
        require(result, "native call");
        return abi.decode(data, (uint256));
    }
    function nativeMintRevert(uint256 amount) public {
        nativeMint(amount);
        revert("test");
    }
    function nativeTransfer(address recipient, uint256 amount) public {
        (bool result, ) = bankContract.call(abi.encodeWithSignature(
            "transfer(address,address,uint256)", msg.sender, recipient, amount
        ));
        require(result, "native transfer");
    }
}