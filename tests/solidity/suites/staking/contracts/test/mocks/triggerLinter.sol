pragma solidity ^0.5.17;


/**
This contract has a bunch of functions that are non-conseuqential. 
Changing them will trigger the linter. This is just a way to test the linter which triggers on change of .sol files.
 */

contract triggerLinter
{
    uint256 public x;

    function setX(uint256 _x) public {
        x = _x;
    }
}
