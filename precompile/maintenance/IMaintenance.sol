// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

/// @title  IMaintenance
/// @notice Interface for the Maintenance precompile
interface IMaintenance {
    /**
     * @notice Emitted when the Emergency Team role is granted, rotated, or revoked.
     * @param previous The Emergency Team before the call. 0x0 if the role was not granted.
     * @param current The Emergency Team after the call. 0x0 if the role is revoked.
     */
    event EmergencyTeamSet(address indexed previous, address indexed current);

    /**
     * @notice Emitted when the bridge lockdown is enabled or disabled.
     * @param bridgeIn True if bridging in is stopped.
     * @param bridgeOut True if bridging out is stopped.
     */
    event BridgeLockdownSet(bool bridgeIn, bool bridgeOut);

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

    /**
     * @notice Updates the byte code associated with a precompile
     * @param precompile The precompile contract address
     * @param code The new byte code to use
     * @dev Must be called by contract owner.
     */
    function setPrecompileByteCode(address precompile, bytes calldata code) external returns (bool);

    /**
     * @notice Sets the chain fee splitter address
     * @param chainFeeSplitterAddress The new chain fee splitter address
     * @dev Must be called by contract owner.
     */
    function setChainFeeSplitterAddress(address chainFeeSplitterAddress) external returns (bool);

    /**
     * @notice Gets the chain fee splitter address.
     * @return The chain fee splitter address.
     */
    function getChainFeeSplitterAddress() external view returns (address);

    /**
     * @notice Sets the minimum gas price of 1 gas unit.
     * @param minGasPrice The new minimum gas price denominated in abtc (1e18 precision).
     * @dev Must be called by contract owner.
     */
    function setMinGasPrice(uint256 minGasPrice) external returns (bool);

    /**
     * @notice Gets the minimum gas price of 1 gas unit.
     * @return The minimum gas price of 1 gas unit denominated in abtc (1e18 precision).
     */
    function getMinGasPrice() external view returns (uint256);

    /**
     * @notice Sets the maximum number of precompile calls allowed per transaction execution.
     * @param value The new maximum precompile calls per execution.
     * @dev Must be called by contract owner.
     */
    function setMaxPrecompilesCallsPerExecution(uint32 value) external returns (bool);

    /**
     * @notice Gets the maximum number of precompile calls allowed per transaction execution.
     * @return The current maximum precompile calls per execution.
     */
    function getMaxPrecompilesCallsPerExecution() external view returns (uint32);

    /**
     * @notice Enables or disables the SELFDESTRUCT opcode.
     * @param disabled True to disable SELFDESTRUCT, false to enable it.
     * @dev When disabled, SELFDESTRUCT fails the current call frame with an
     *      invalid-opcode error.
     * @dev Must be called by contract owner.
     */
    function setSelfDestructDisabled(bool disabled) external returns (bool);

    /**
     * @notice Checks whether the SELFDESTRUCT opcode is disabled.
     * @return True if SELFDESTRUCT is disabled. False otherwise.
     */
    function getSelfDestructDisabled() external view returns (bool);

    /**
     * @notice Grants the Emergency Team role to the given address.
     * @param team The address that takes the role. Pass 0x0 to revoke the role.
     * @dev The grant is a single step and can be repeated to rotate the role.
     * @dev Must be called by contract owner.
     */
    function setEmergencyTeam(address team) external returns (bool);

    /**
     * @notice Gets the current Emergency Team.
     * @return team The Emergency Team. 0x0 if the role is not granted.
     */
    function getEmergencyTeam() external view returns (address team);

    /**
     * @notice Enables or disables the bridge lockdown, per direction.
     * @param bridgeIn True to stop bridging in, false to allow it.
     * @param bridgeOut True to stop bridging out, false to allow it.
     * @dev Pass false for both directions to disable the lockdown. The lockdown
     *      does not change the bridge outflow limits.
     * @dev bridgeIn currently stops the triparty inbound path only. The
     *      validator-observed path has no pause switch yet.
     * @dev Must be called by the Emergency Team or the contract owner.
     */
    function setBridgeLockdown(bool bridgeIn, bool bridgeOut) external returns (bool);

    /**
     * @notice Gets the bridge lockdown state, per direction.
     * @return bridgeIn True if bridging in is stopped.
     * @return bridgeOut True if bridging out is stopped.
     */
    function getBridgeLockdown() external view returns (bool bridgeIn, bool bridgeOut);
}
