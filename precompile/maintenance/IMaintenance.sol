// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

/// @title  IMaintenance
/// @notice Interface for the Maintenance precompile
interface IMaintenance {
    /**
     * @notice Enables/disables support for the non-EIP155 txs without replay protection.
     * @param value The new value of the flag.
     * @dev Must be called by contract owner.
     */
    function setSupportNonEIP155Txs(bool value) external returns (bool);

    /**
     * @notice Checks status of support for the non-EIP155 txs without replay protection.
     * @return True if non-EIP155 txs are supported. False otherwise.
     */
    function getSupportNonEIP155Txs() external view returns (bool);
}