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

    function bridgeOut(address token, uint256 amount, uint8 chain, bytes calldata recipient) external returns (bool) {
        return IAssetsBridge(precompile).bridgeOut(token, amount, chain, recipient);
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

    function getCurrentSequenceTip() external view returns (uint256) {
        return IAssetsBridge(precompile).getCurrentSequenceTip();
    }

    function getSourceBTCToken() external view returns (address) {
        return IAssetsBridge(precompile).getSourceBTCToken();
    }

    function setOutflowLimit(address token, uint256 limit) external returns (bool) {
        return IAssetsBridge(precompile).setOutflowLimit(token, limit);
    }

    function getOutflowLimit(address token) external view returns (uint256) {
        return IAssetsBridge(precompile).getOutflowLimit(token);
    }

    function getOutflowCapacity(address token) external view returns (uint256 capacity, uint256 resetHeight) {
        return IAssetsBridge(precompile).getOutflowCapacity(token);
    }

    function setPauser(address pauser) external returns (bool) {
        return IAssetsBridge(precompile).setPauser(pauser);
    }

    function getPauser() external view returns (address) {
        return IAssetsBridge(precompile).getPauser();
    }

    function pauseBridgeOut() external returns (bool) {
        return IAssetsBridge(precompile).pauseBridgeOut();
    }

    function setMinBridgeOutAmount(address mezoToken, uint256 minAmount) external returns (bool) {
        return IAssetsBridge(precompile).setMinBridgeOutAmount(mezoToken, minAmount);
    }

    function getMinBridgeOutAmount(address mezoToken) external view returns (uint256) {
        return IAssetsBridge(precompile).getMinBridgeOutAmount(mezoToken);
    }

    function setMinBridgeOutAmountForBitcoinChain(uint256 minAmount) external returns (bool) {
        return IAssetsBridge(precompile).setMinBridgeOutAmountForBitcoinChain(minAmount);
    }

    function getMinBridgeOutAmountForBitcoinChain() external view returns (uint256) {
        return IAssetsBridge(precompile).getMinBridgeOutAmountForBitcoinChain();
    }

    function bridgeTriparty(address recipient, uint256 amount, bytes calldata callbackData) external returns (uint256 requestId) {
        return IAssetsBridge(precompile).bridgeTriparty(recipient, amount, callbackData);
    }

    function allowTripartyController(address controller, bool isAllowed) external returns (bool) {
        return IAssetsBridge(precompile).allowTripartyController(controller, isAllowed);
    }

    function isAllowedTripartyController(address controller) external view returns (bool) {
        return IAssetsBridge(precompile).isAllowedTripartyController(controller);
    }

    function pauseTriparty(bool isPaused) external returns (bool) {
        return IAssetsBridge(precompile).pauseTriparty(isPaused);
    }

    function setTripartyBlockDelay(uint256 delay) external returns (bool) {
        return IAssetsBridge(precompile).setTripartyBlockDelay(delay);
    }

    function getTripartyBlockDelay() external view returns (uint256 delay) {
        return IAssetsBridge(precompile).getTripartyBlockDelay();
    }

    function setTripartyLimits(uint256 perRequestLimit, uint256 windowLimit) external returns (bool) {
        return IAssetsBridge(precompile).setTripartyLimits(perRequestLimit, windowLimit);
    }

    function getTripartyLimits() external view returns (uint256 perRequestLimit, uint256 windowLimit) {
        return IAssetsBridge(precompile).getTripartyLimits();
    }
}

