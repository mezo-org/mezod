// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

import { IAssetsBridge, AssetsLocked, ERC20TokenMapping } from "../interfaces/IAssetsBridge.sol";

contract AssetsBridgeCaller is IAssetsBridge {
    address private constant precompile = 0x7B7C000000000000000000000000000000000012;

    function bridge(AssetsLocked[] memory events) external returns (bool) {
        return IAssetsBridge(precompile).bridge(events);
    }

    function createERC20TokenMapping(address sourceToken, address mezoToken) external returns (bool) {
        return IAssetsBridge(precompile).createERC20TokenMapping(sourceToken, mezoToken);
    }

    function deleteERC20TokenMapping(address sourceToken) external returns (bool) {
        return IAssetsBridge(precompile).deleteERC20TokenMapping(sourceToken);
    }

    function getERC20TokenMapping(address sourceToken) external view returns (ERC20TokenMapping memory) {
        return IAssetsBridge(precompile).getERC20TokenMapping(sourceToken);
    }

    function getERC20TokensMappings() external view returns (ERC20TokenMapping[] memory) {
        return IAssetsBridge(precompile).getERC20TokensMappings();
    }

    function getMaxERC20TokensMappings() external view returns (uint256) {
        return IAssetsBridge(precompile).getMaxERC20TokensMappings();
    }

    function getSourceBTCToken() external view returns (address) {
        return IAssetsBridge(precompile).getSourceBTCToken();
    }
}
