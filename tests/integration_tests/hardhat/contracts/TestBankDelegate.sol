// SPDX-License-Identifier: MIT
pragma solidity >0.6.6;

contract TestBankDelegate {
    function nativeMint(address _contract, uint256 amount) public {
        (bool result, bytes memory _data) = _contract.delegatecall(abi.encodeWithSignature(
            "nativeMint(uint256)", amount
        ));
        require(result, "native call");
    }
    function nativeBalanceOf(address _contract, address addr) public returns (uint256) {
        (bool result, bytes memory data) = _contract.delegatecall(abi.encodeWithSignature(
            "nativeBalanceOf(address)", addr
        ));
        require(result, "native call");
        return abi.decode(data, (uint256));
    }
    function nativeMintRevert(address _contract, uint256 amount) public {
        nativeMint(_contract, amount);
        revert("test");
    }
}
