// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

/// @title AssetsLocked
/// @notice Represents bridged assets.
struct AssetsLocked {
    uint256 sequenceNumber;
    address recipient;
    uint256 tbtcAmount;
}

/// @title  IAssetsBridge
/// @notice Interface for the Assets Bridge precompile
interface IAssetsBridge {
    /**
     * @notice Helper function used to enable bridged assets observability.
     */
    function bridge(AssetsLocked[] memory events) external returns (bool);
}
