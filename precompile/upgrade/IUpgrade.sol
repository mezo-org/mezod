// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

/// @title  IUpgrade
/// @notice Interface for the Upgrade precompile
interface IUpgrade {
    /** 
     * @notice Returns the latest upgrade plan
     * @return name The name of the upgrade plan
     * @return height The block height to activate the upgrade
     * @return info The upgrade information
     */
    function plan() external view returns (string memory name, string memory height, string memory info);

    /** 
     * @notice Returns `true` after updating the contracts upgrade plan
     * @param name The name of the upgrade plan
     * @param height The block height to activate the upgrade
     * @param info The upgrade information
     * @dev Must be called by contract owner
     */
    function submitPlan(string calldata name, string calldata height, string calldata info) external returns (bool);

    /** 
     * @notice Returns `true` after cancelling the contracts upgrade plan
     * @dev Must be called by contract owner
     */
    function cancelPlan() external returns (bool);
}