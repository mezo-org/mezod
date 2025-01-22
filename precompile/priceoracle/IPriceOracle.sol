// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

/// @title  IPriceOracle
/// @notice Interface for the Price Oracle precompile.
/// @dev This is actually the ChainlinkAggregatorV3Interface interface.
interface IPriceOracle {
    /**
    * @return Decimal places of the precision used to represent the price.
    *         E.g. if decimals is 18, the price is represented with the 10^18 precision.
    */
    function decimals() external view returns (uint8);

    /**
    * @notice Returns the price from the last update tick (round) of the oracle.
    * @return roundId The round ID.
    * @return answer The price represented in the 10^decimals precision.
    * @return startedAt Start timestamp of the update tick.
    * @return updatedAt End timestamp of the update tick.
    * @return answeredInRound Deprecated - Previously used when answers could
    *         take multiple rounds to be computed.
    */
    function latestRoundData()
        external
        view
        returns (
            uint80 roundId,
            int256 answer,
            uint256 startedAt,
            uint256 updatedAt,
            uint80 answeredInRound
        );
}
