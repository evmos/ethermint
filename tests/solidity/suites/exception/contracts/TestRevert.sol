// SPDX-License-Identifier: MIT
pragma solidity >=0.8.0;

contract State {
    uint256 a = 0;
    function set(uint256 input) public { 
        a = input; 
        require(a < 10);
    }
    function force_set(uint256 input) public { 
        a = input; 
    }
    function query() public view returns(uint256) {
        return a;
    }
}

contract TestRevert {
    State state;
    constructor() {
        state = new State();
    }
    function try_set(uint256 input) public {
        try state.set(input) {
        } catch (bytes memory) {
        }
    }
    function set(uint256 input) public {
        state.force_set(input);
    }
    function query() public view returns(uint256) { 
        return state.query();
    }
}
