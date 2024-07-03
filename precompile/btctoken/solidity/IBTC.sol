// SPDX-License-Identifier: MIT

pragma solidity ^0.8.20;

import "./IERC20.sol";
import "./IERC20Metadata.sol";
import "./IApproveAndCall.sol";
import "./IERC20WithPermit.sol";

/// @title  IBTC
/// @notice Interface for the BTC token.
interface IBTC is IERC20, IERC20WithPermit, IERC20Metadata, IApproveAndCall {
}

