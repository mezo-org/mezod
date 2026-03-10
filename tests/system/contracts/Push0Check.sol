// SPDX-License-Identifier: MIT
pragma solidity ^0.8.28;

contract Push0Check {
    function zero() external pure returns (uint256) {
        return 0; // make it use PUSH0 in Shanghai or later
    }
}
