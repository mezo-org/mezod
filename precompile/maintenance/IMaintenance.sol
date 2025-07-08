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
}