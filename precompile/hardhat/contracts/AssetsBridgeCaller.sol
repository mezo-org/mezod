// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

import { IAssetsBridge, AssetsLocked } from "../interfaces/IAssetsBridge.sol";

contract AssetsBridgeCaller is IAssetsBridge {
    address private constant precompile  = 0x7B7C000000000000000000000000000000000012;

    function bridge(AssetsLocked[] memory events) external returns (bool) {
        return IAssetsBridge(precompile).bridge(events);
    }
}
