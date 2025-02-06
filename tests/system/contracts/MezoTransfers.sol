// SPDX-License-Identifier: MIT
pragma solidity ^0.8.28;

import { IBTC } from "./interfaces/IBTC.sol";

/// @title MezoTransfers
/// @notice Handles various transfer scenarios for Mezo native token - BTC.
contract MezoTransfers {
    // BTC ERC-20 token address on Mezo
    address private constant precompile = 0x7b7C000000000000000000000000000000000000;

    event NativeTransfer(address indexed sender, address indexed recipient, uint256 amount);
    event BTCERC20Transfer(address indexed sender, address indexed recipient, address indexed token, uint256 amount);

    /// @notice Transfers native BTC and then BTC ERC-20 Token within the same transaction
    function nativeThenBTCERC20(address payable recipient, uint256 tokenAmount) external payable {
        require(msg.value > 0, "Must send some native tokens");

        // Transfer native BTC
        (bool sent, ) = recipient.call{value: msg.value}("");
        require(sent, "Native transfer failed");
        emit NativeTransfer(msg.sender, recipient, msg.value);

        // Transfer ERC-20 BTC
        bool success = IBTC(precompile).transferFrom(msg.sender, recipient, tokenAmount);
        require(success, "BTC ERC-20 transfer failed");
        emit BTCERC20Transfer(msg.sender, recipient, precompile, tokenAmount);
    }

    // @notice Transfers BTC ERC-20 Token and then native BTC within the same transaction
    function btcERC20ThenNative(address payable recipient, uint256 tokenAmount) external payable {
        require(msg.value > 0, "Must send some native tokens");

        // Transfer ERC-20 BTC
        bool success = IBTC(precompile).transferFrom(msg.sender, recipient, tokenAmount);
        require(success, "BTC ERC-20 transfer failed");
        emit BTCERC20Transfer(msg.sender, recipient, precompile, tokenAmount);

        // Transfer native BTC
        (bool sent, ) = recipient.call{value: msg.value}("");
        require(sent, "Native transfer failed");
        emit NativeTransfer(msg.sender, recipient, msg.value);
    }
}