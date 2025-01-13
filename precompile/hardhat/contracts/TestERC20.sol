// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

contract TestERC20 {
    address public minter;
    uint256 public totalSupply = 0;
    mapping(address => uint256) public balanceOf;

    event Minted(address indexed account, uint256 amount);

    constructor(address _minter) {
        minter = _minter;
    }

    function mint(address account, uint256 amount) external {
        require(msg.sender == minter, "TestERC20: only minter can mint");

        emit Minted(account, amount);

        totalSupply += amount;
        balanceOf[account] += amount;
    }
}