// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

import { IUpgrade } from "../interfaces/IUpgrade.sol";

contract UpgradeCaller is IUpgrade {
    address private constant precompile = 0x7b7c000000000000000000000000000000000014;

    function plan() external view returns (string memory name, int64 height, string memory info) {
        return IUpgrade(precompile).plan();
    }

    function submitPlan(string calldata name, int64 height, string calldata info) external returns (bool) {
        return IUpgrade(precompile).submitPlan(name, height, info);
    }

    function cancelPlan() external returns (bool) {
        return IUpgrade(precompile).cancelPlan();
    }
}