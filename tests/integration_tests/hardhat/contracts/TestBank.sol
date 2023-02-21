// SPDX-License-Identifier: MIT
pragma solidity >0.6.6;
import "@openzeppelin/contracts/token/ERC20/ERC20.sol";

contract TestBank is ERC20 {
    address constant bankContract = 0x0000000000000000000000000000000000000064;
    constructor() public ERC20("Bitcoin MAX", "MAX") {
		_mint(msg.sender, 100000000000000000000000000);
	}
    function moveToNative(uint256 amount) public {
        _burn(msg.sender, amount);
        (bool result, ) = bankContract.call(abi.encodeWithSignature(
            "mint(address,uint256)", msg.sender, amount
        ));
        require(result, "native call");
    }
    function moveFromNative(uint256 amount) public {
        (bool result, ) = bankContract.call(abi.encodeWithSignature(
            "burn(address,uint256)", msg.sender, amount
        ));
        require(result, "native call");
        _mint(msg.sender, amount);
    }
    function nativeBalanceOf(address addr) public returns (uint256) {
        (bool result, bytes memory data) = bankContract.call(abi.encodeWithSignature(
            "balanceOf(address,address)", address(this), addr
        ));
        require(result, "native call");
        return abi.decode(data, (uint256));
    }
    function moveToNativeRevert(uint256 amount) public {
        moveToNative(amount);
        revert("test");
    }
    function nativeTransfer(address recipient, uint256 amount) public {
        _transfer(msg.sender, recipient, amount);
        (bool result, ) = bankContract.call(abi.encodeWithSignature(
            "transfer(address,address,uint256)", msg.sender, recipient, amount
        ));
        require(result, "native transfer");
    }
}