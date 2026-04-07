// SPDX-License-Identifier: MIT
pragma solidity ^0.8.28;

import {IAssetsBridge} from "./interfaces/IAssetsBridge.sol";

/// @title TripartyController
/// @notice Test controller for triparty BTC minting through the bridge.
contract TripartyController {
    address private constant bridgePrecompile =
        0x7B7C000000000000000000000000000000000012;

    struct CallbackRecord {
        uint256 requestId;
        address recipient;
        uint256 amount;
        bytes callbackData;
    }

    CallbackRecord[] public callbacks;
    bool public revertOnCallback;

    function requestMint(
        address recipient,
        uint256 amount,
        bytes calldata callbackData
    ) external returns (uint256) {
        return
            IAssetsBridge(bridgePrecompile).bridgeTriparty(
                recipient,
                amount,
                callbackData
            );
    }

    function onTripartyBridgeCompleted(
        uint256 requestId,
        address recipient,
        uint256 amount,
        bytes calldata callbackData
    ) external {
        if (revertOnCallback) {
            revert("callback reverted");
        }
        callbacks.push(
            CallbackRecord({
                requestId: requestId,
                recipient: recipient,
                amount: amount,
                callbackData: callbackData
            })
        );
    }

    function setRevertOnCallback(bool _revert) external {
        revertOnCallback = _revert;
    }

    function getCallbackCount() external view returns (uint256) {
        return callbacks.length;
    }

    function getCallback(
        uint256 index
    ) external view returns (uint256, address, uint256, bytes memory) {
        CallbackRecord storage cb = callbacks[index];
        return (cb.requestId, cb.recipient, cb.amount, cb.callbackData);
    }

    function batchRequestMint(
        address[] calldata recipients,
        uint256[] calldata amounts
    ) external {
        for (uint i = 0; i < recipients.length; i++) {
            IAssetsBridge(bridgePrecompile).bridgeTriparty(
                recipients[i],
                amounts[i],
                ""
            );
        }
    }
}
