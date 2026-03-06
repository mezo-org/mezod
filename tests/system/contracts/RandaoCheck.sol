// SPDX-License-Identifier: MIT
pragma solidity ^0.8.28;

contract RandaoCheck {
    function values() external view returns (uint256 difficulty, uint256 prevrandao) {
        return (block.difficulty, block.prevrandao);
    }
}
