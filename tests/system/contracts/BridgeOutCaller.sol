// SPDX-License-Identifier: MIT
pragma solidity ^0.8.28;

import {IAssetsBridge} from "./interfaces/IAssetsBridge.sol";
import {IERC20} from "./interfaces/solidity/IERC20.sol";

contract BridgeOutCaller {
    address private constant BRIDGE =
        0x7B7C000000000000000000000000000000000012;
    address public immutable owner;

    constructor() {
        owner = msg.sender;
    }

    function execute(
        address token,
        uint256 amount,
        bytes calldata recipient
    ) external {
        bool ok = IERC20(token).transferFrom(
            msg.sender,
            address(this),
            amount
        );
        require(ok, "transferFrom failed");

        ok = IERC20(token).approve(BRIDGE, amount);
        require(ok, "approve failed");

        ok = IAssetsBridge(BRIDGE).bridgeOut(token, amount, 0, recipient);
        require(ok, "bridgeOut failed");

        try IERC20(token).transfer(owner, 1) returns (bool transferOk) {
            require(!transferOk, "unexpected transfer success");
        } catch {
            return;
        }
    }
}
