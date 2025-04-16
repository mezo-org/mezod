// SPDX-License-Identifier: MITx

pragma solidity ^0.8.20;

/// @title  ITestbed
/// @notice Interface for the ITestbed contract.
interface ITestbed {
    /**
     * @dev Emitted when `value` tokens are moved from one account (`from`) to
     * another (`to`).
     *
     * Note that `value` may be zero.
     */
    event Transfer(address indexed from, address indexed to, uint256 value);

    /**
     * @dev Moves a `value` amount of tokens from the caller's account to `to`,
     * then call revert (doing so reverting the transfer that just occured.
     *
     * Returns always false.
     */
    function transferWithRevert(address to, uint256 value) external returns (bool);
}
