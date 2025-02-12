// SPDX-License-Identifier: MIT
pragma solidity ^0.8.28;

import { IBTC } from "./interfaces/IBTC.sol";

/// @title MezoTransfers
/// @notice Handles various transfer scenarios for Mezo native token - BTC.
contract MezoTransfers {
    // BTC ERC-20 token address on Mezo
    address private constant precompile = 0x7b7C000000000000000000000000000000000000;
    uint256 public balanceTracker;

    /// @notice Transfers native BTC and then BTC ERC-20 Token from the contract
    ///         which was previously funded.
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
    ///         which was previously funded.
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

    /// @notice Receive native then transfer all received native amount as
    ///         native BTC.
    function receiveSendNative(address payable recipient) external payable {
        require(msg.value > 0, "Must send some native tokens");

        // Transfer native BTC
        (bool sent, ) = recipient.call{value: msg.value}("");
        require(sent, "Native transfer failed");
    }

    /// @notice Receive native then transfer all received native amount as
    ///         BTC ERC-20 Token.
    function receiveSendBTCERC20(address payable recipient) external payable {
        require(msg.value > 0, "Must send some native tokens");
        
        // Transfer ERC-20 BTC
        bool success = IBTC(precompile).transfer(recipient, msg.value);
        require(success, "BTC ERC-20 transfer failed");
    }

    /// @notice Receive native then transfer half as native BTC and half as
    ///         BTC ERC-20 Token. All the transfers should be funded from the 
    ///         received native amount.
    function receiveSendNativeThenBTCERC20(address payable recipient) external payable {
        require(msg.value > 0, "Must send some native tokens");

        uint256 halfAmount = msg.value / 2;

        // Transfer native BTC
        (bool sent, ) = recipient.call{value: halfAmount}("");
        require(sent, "Native transfer failed");

        // Transfer ERC-20 BTC
        bool success = IBTC(precompile).transfer(recipient, halfAmount);
        require(success, "BTC ERC-20 transfer failed");
    }

    /// @notice Receive native then transfer half as BTC ERC-20 Token and half as
    ///         native BTC. All the transfers should be funded from the
    ///         received native amount.
    function receiveSendBTCERC20ThenNative(address payable recipient) external payable {
        require(msg.value > 0, "Must send some native tokens");

        uint256 halfAmount = msg.value / 2;

        // Transfer ERC-20 BTC
        bool success = IBTC(precompile).transfer(recipient, halfAmount);
        require(success, "BTC ERC-20 transfer failed");

        // Transfer native BTC
        (bool sent, ) = recipient.call{value: halfAmount}("");
        require(sent, "Native transfer failed");
    }

    /// @notice Transfers BTC ERC-20 Token and native BTC multiple times.
    ///         All the transfers should be funded from the received native amount.
    function multipleBTCERC20AndNative(address payable recipient) external payable {
        require(msg.value > 0, "Must send some native tokens");

        uint256 quarterAmount = msg.value / 4;

        // Transfer 1/4 of the token amount as ERC-20 BTC
        bool success1 = IBTC(precompile).transfer(recipient, quarterAmount);
        require(success1, "BTC ERC-20 transfer failed");

        // Transfer 1/4 of the native BTC
        (bool sent1, ) = recipient.call{value: quarterAmount}("");
        require(sent1, "Native transfer failed");

        // Transfer 1/4 of the token amount as ERC-20 BTC
        bool success2 = IBTC(precompile).transfer(recipient, quarterAmount);
        require(success2, "BTC ERC-20 transfer failed");

        // Transfer 1/4 of the native BTC
        (bool sent2, ) = recipient.call{value: quarterAmount}("");
        require(sent2, "Native transfer failed");
    }

    /// @notice Transfers with revert
    function transferWithRevert(address payable recipient) external payable {
        require(msg.value > 0, "Must send some native tokens");

        uint256 halfAmount = msg.value / 2;

        // Transfer ERC-20 BTC
        bool success = IBTC(precompile).transfer(recipient, halfAmount);
        require(success, "BTC ERC-20 transfer failed");

        // Transfer native BTC
        (bool sent, ) = recipient.call{value: halfAmount}("");
        require(sent, "Native transfer failed");

        revert("revert after transfers");
    }

    /// @notice Transfers native BTC then BTC ERC-20 Token within the same transaction.
    ///         All funds for native BTC transfer should come from native tokens received.
    ///         All the funds for BTC ERC20 transfer should be approved by the sender and
    ///         this function should faciliate the transfer.
    function transferNativeThenBTCERC20(address payable recipient, uint256 tokenAmount) external payable {
        require(msg.value > 0, "Must send some native tokens");

        // Transfer native BTC
        (bool sent, ) = recipient.call{value: msg.value}("");
        require(sent, "Native transfer failed");

        // Transfer BTC ERC-20
        bool success = IBTC(precompile).transferFrom(msg.sender, recipient, tokenAmount);
        require(success, "BTC ERC-20 transfer failed");
    }

    /// @notice Transfers BTC ERC-20 Token and then native BTC within the same transaction.
    ///         All funds for native BTC transfer should come from native tokens received.
    ///         All the funds for BTC ERC20 transfer should be approved by the sender and
    ///         this function should faciliate the transfer.
    function transferBTCERC20ThenNative(address payable recipient, uint256 tokenAmount) external payable {
        require(msg.value > 0, "Must send some native tokens");

        // Transfer BTC ERC-20
        bool success = IBTC(precompile).transferFrom(msg.sender, recipient, tokenAmount);
        require(success, "BTC ERC-20 transfer failed");

        // Transfer native BTC
        (bool sent, ) = recipient.call{value: msg.value}("");
        require(sent, "Native transfer failed");
    }

    /// @notice Update the storage variable, transfer ERC20 and then reset the
    ///         storage variable to its original value. This contract should be
    ///         funded with some BTC ERC-20 tokens before calling this function.
    function storageStateTransition(address recipient) external {
        uint256 balance = (address(this)).balance;
        require(balance > 0, "Contract has no BTC ERC-20 balance");

        balanceTracker = 42; // This is an arbitrary value for testing purposes only.

        // Transfer ERC-20.
        bool success = IBTC(precompile).transfer(recipient, balance);
        require(success, "BTC ERC-20 transfer failed");

        balanceTracker = 0; // Reset the storage variable to its original value.
    }
}