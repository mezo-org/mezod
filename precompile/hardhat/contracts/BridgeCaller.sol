// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

import { IBridgeCaller, AssetsLocked } from "../interfaces/IBridgeCaller.sol";

contract BridgeCaller is IBridgeCaller {
    address private constant precompile  = 0x7B7C000000000000000000000000000000000012;

    function bridge(AssetsLocked[] memory events) external returns (bool) {
        return IBridgeCaller(precompile).bridge(events);
    }
}
