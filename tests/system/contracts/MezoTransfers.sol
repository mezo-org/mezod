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

    /// @notice Transfers native BTC and then BTC ERC-20 Token from the contract
    function nativeThenBTCERC20(address recipient) external {
        uint256 balance = IBTC(precompile).balanceOf(address(this));
        require(balance > 0, "No balance to transfer");

        uint256 halfBalance = balance / 2;

        // Transfer native BTC
        (bool sent, ) = recipient.call{value: halfBalance}("");
        require(sent, "Transfer using call failed");

        // Transfer ERC-20 BTC
        bool success = IBTC(precompile).transfer(recipient, halfBalance);
        require(success, "Transfer using transfer failed");
    }

    /// @notice Transfers BTC ERC-20 Token and then native BTC from the contract
    function btcERC20ThenNative(address recipient) external {
        uint256 balance = IBTC(precompile).balanceOf(address(this));
        require(balance > 0, "No balance to transfer");

        uint256 halfBalance = balance / 2;

        // Transfer ERC-20 BTC
        bool success = IBTC(precompile).transfer(recipient, halfBalance);
        require(success, "Transfer using transfer failed");

        // Transfer native BTC
        (bool sent, ) = recipient.call{value: halfBalance}("");
        require(sent, "Transfer using call failed");
    }

    /// @notice Transfers native BTC
    function receiveSendNative(address payable recipient) external payable {
        require(msg.value > 0, "Must send some native tokens");

        // Transfer native BTC
        (bool sent, ) = recipient.call{value: msg.value}("");
        require(sent, "Native transfer failed");
        emit NativeTransfer(msg.sender, recipient, msg.value);
    }

    /// @notice Receive native but transfer only BTC ERC-20 Token
    function receiveSendBTCERC20(address payable recipient, uint256 tokenAmount) external payable {
        require(msg.value > 0, "Must send some native tokens");
        
        // Transfer ERC-20 BTC
        bool success = IBTC(precompile).transferFrom(msg.sender, recipient, tokenAmount);
        require(success, "BTC ERC-20 transfer failed");
        emit BTCERC20Transfer(msg.sender, recipient, precompile, tokenAmount);
    }

    /// @notice Transfers native BTC and then BTC ERC-20 Token within the same transaction
    function receiveAndSendNativeThenBTCERC20(address payable recipient, uint256 tokenAmount) external payable {
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
    function receiveAndSendBtcERC20ThenNative(address payable recipient, uint256 tokenAmount) external payable {
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

    // @notice Transfers BTC ERC-20 Token and then native BTC within the same transaction
    function multipleBTCERC20AndNative(address payable recipient, uint256 tokenAmount) external payable {
        require(msg.value > 0, "Must send some native tokens");

        uint256 halfTokenAmount = tokenAmount / 2;
        uint256 halfNativeAmount = msg.value / 2;

        // Transfer 1/2 of the token amount as ERC-20 BTC
        bool success1 = IBTC(precompile).transferFrom(msg.sender, recipient, halfTokenAmount);
        require(success1, "BTC ERC-20 transfer failed");
        emit BTCERC20Transfer(msg.sender, recipient, precompile, halfTokenAmount);

        // Transfer 1/2 of the native BTC
        (bool sent1, ) = recipient.call{value: halfNativeAmount}("");
        require(sent1, "Native transfer failed");
        emit NativeTransfer(msg.sender, recipient, halfNativeAmount);

        // Transfer 1/2 of the token amount as ERC-20 BTC
        bool success2 = IBTC(precompile).transferFrom(msg.sender, recipient, halfTokenAmount);
        require(success2, "BTC ERC-20 transfer failed");
        emit BTCERC20Transfer(msg.sender, recipient, precompile, halfTokenAmount);

        // Transfer 1/2 of the native BTC
        (bool sent2, ) = recipient.call{value: halfNativeAmount}("");
        require(sent2, "Native transfer failed");
        emit NativeTransfer(msg.sender, recipient, halfNativeAmount);
    }
}