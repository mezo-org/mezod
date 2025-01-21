// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

import { IPriceOracle } from "../interfaces/IPriceOracle.sol";

contract PriceOracleCaller is IPriceOracle {
    address private constant precompile  = 0x7b7c000000000000000000000000000000000015;

    function decimals() external view returns (uint8) {
        return IPriceOracle(precompile).decimals();
    }

    function latestRoundData()
        external
        view
        returns (
            uint80 roundId,
            int256 answer,
            uint256 startedAt,
            uint256 updatedAt,
            uint80 answeredInRound
        ) {
        return IPriceOracle(precompile).latestRoundData();
    }
}
