// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

contract TestTransfer {
    address constant recursiveContract = 0x0000000000000000000000000000000000000066;
    event Result(bool indexed result);
    struct Params {
        address receiver;
        uint256 amount;
    }
    constructor() payable {}

    function nativeTransfer(Params memory params) public {
        (bool result, ) = recursiveContract.call(abi.encodeWithSignature(
            "nativeTransfer(address,uint256)", params.receiver, params.amount
        ));
        require(result, "native transfer");
    }

    function recursiveTransfer(Params[] memory params) public {
        nativeTransfer(params[0]);
        nativeTransfer(params[1]);

        Params[] memory _params = new Params[](params.length - 2);
        for (uint i = 2; i < params.length; i++) {
            _params[i - 2] = params[i];
        }
        (bool result, ) = address(this).call(abi.encodeWithSignature(
            "recursiveTransfer((address,uint256)[])", _params
        ));
        emit Result(result);
    }
}