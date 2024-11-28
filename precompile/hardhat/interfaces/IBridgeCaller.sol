// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

/// @title AssetsLocked
/// @notice Represents bridged assets.
struct AssetsLocked {
    uint256 sequenceNumber;
    address recipient;
    uint256 tbtcAmount;
}

/// @title  IBridgeCaller
/// @notice Interface for the Bridge precompile
interface IBridgeCaller {
    /**
     * @notice Helper function used to enable bridged assets obeservability.
     */
    function bridge(AssetsLocked[] memory events) external returns (bool);
}
