// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

/// @title  IUpgrade
/// @notice Interface for the Upgrade precompile
interface IUpgrade {
    /** 
     * @notice Emitted when an upgrade plan is canceled
     * @param name The name of the canceled upgrade plan
     * @param height The block height of the canceled upgrade plan
     */ 
    event PlanCanceled(string name, int64 height);

    /** 
     * @notice Emitted when a new upgrade plan is submitted
     * @param name The name of the new upgrade plan
     * @param height The block height of the new upgrade plan
     */ 
    event PlanSubmitted(string name, int64 height);

    /** 
     * @notice Returns the latest upgrade plan
     * @return name The name of the upgrade plan
     * @return height The block height to activate the upgrade
     * @return info The upgrade information
     */
    function plan() external view returns (string memory name, int64 height, string memory info);

    /** 
     * @notice Returns `true` after updating the contracts upgrade plan
     * @param name The name of the upgrade plan
     * @param height The block height to activate the upgrade
     * @param info The upgrade information
     * @dev Must be called by contract owner
     */
    function submitPlan(string calldata name, int64 height, string calldata info) external returns (bool);

    /**
     * @notice Returns `true` after cancelling the contracts upgrade plan
     * @dev Must be called by contract owner
     */
    function cancelPlan() external returns (bool);
}