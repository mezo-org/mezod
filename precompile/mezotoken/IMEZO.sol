// SPDX-License-Identifier: MIT

pragma solidity ^0.8.20;

import "../erc20/IERC20WithPermit.sol";

/// @title  IMEZO
/// @notice Interface for the MEZO token.
interface IMEZO is IERC20WithPermit {
    /// @notice Emitted when a new minter is set.
    /// @param minter The address of the new minter.
    event MinterSet(address indexed minter);

    /// @notice Emitted when tokens are minted.
    /// @param to The address receiving the minted tokens.
    /// @param amount The amount of tokens minted.
    event Minted(address indexed to, uint256 amount);

    /// @notice Sets the minter address (only callable by POA owner).
    /// @param minter The address of the new minter.
    function setMinter(address minter) external returns (bool);

    /// @notice Gets the current minter address.
    /// @return The address of the current minter.
    function getMinter() external view returns (address);

    /// @notice Mints tokens to a recipient (only callable by minter).
    /// @param to The address to mint tokens to.
    /// @param amount The amount of tokens to mint.
    function mint(address to, uint256 amount) external returns (bool);
}
