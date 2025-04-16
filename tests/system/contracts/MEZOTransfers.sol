// SPDX-License-Identifier: MIT
pragma solidity ^0.8.28;

import {IMEZO} from "./interfaces/IMEZO.sol";
import {IBTC} from "./interfaces/IBTC.sol";

/// @title MEZOTransfers
/// @notice Handles various transfer scenarios for the MEZO token.
contract MEZOTransfers {
    // MEZO token address on Mezo.
    address private constant mezoPrecompile = 0x7B7c000000000000000000000000000000000001;
    // BTC token address on Mezo.
    address private constant btcPrecompile = 0x7b7C000000000000000000000000000000000000;

    /// @notice Transfers half of the contract's MEZO balance to the recipient.
    function transferSomeMEZO(address recipient) external {
        uint256 balance = IMEZO(mezoPrecompile).balanceOf(address(this));
        require(balance > 0, "No balance to transfer");

        uint256 halfBalance = balance / 2;

        bool success = IMEZO(mezoPrecompile).transfer(recipient, halfBalance);
        require(success, "Transfer failed");
    }

    /// @notice Transfers the entire contract's MEZO balance to the recipient.
    function transferAllMEZO(address recipient) external {
        uint256 balance = IMEZO(mezoPrecompile).balanceOf(address(this));
        require(balance > 0, "No balance to transfer");

        bool success = IMEZO(mezoPrecompile).transfer(recipient, balance);
        require(success, "Transfer failed");
    }

    /// @notice Pulls the given amount of MEZO from the sender to the contract.
    function pullMEZO(address sender, uint256 amount) external {
        bool success = IMEZO(mezoPrecompile).transferFrom(sender, address(this), amount);
        require(success, "Transfer failed");
    }

    /// @notice Pulls the given amount of MEZO from the sender to the given recipient.
    function pullMEZOToRecipient(address sender, address recipient, uint256 amount) external {
        bool success = IMEZO(mezoPrecompile).transferFrom(sender, recipient, amount);
        require(success, "Transfer failed");
    }

    /// @notice Receives native BTC (msg.value) and transfers the contract's 
    ///         MEZO balance to the recipient.
    function receiveNativeThenTransferMEZO(address recipient) external payable {
        require(msg.value > 0, "Must send some native BTC");

        uint256 mezoBalance = IMEZO(mezoPrecompile).balanceOf(address(this));
        require(mezoBalance > 0, "No MEZO balance to transfer");

        bool success = IMEZO(mezoPrecompile).transfer(recipient, mezoBalance);
        require(success, "Transfer failed");
    }

    /// @notice Receives native BTC (msg.value) and pulls the given amount of 
    ///         MEZO from the sender to the contract.
    function receiveNativeThenPullMEZO(address sender, uint256 amount) external payable {
        require(msg.value > 0, "Must send some native BTC");

        bool success = IMEZO(mezoPrecompile).transferFrom(sender, address(this), amount);
        require(success, "Transfer failed");
    }

    /// @notice Sends the entire contract's BTC balance natively then transfers 
    ///         the contract's MEZO balance to the recipient.
    function sendNativeThenTransferMEZO(address recipient) external {
        uint256 btcBalance = address(this).balance;
        require(btcBalance > 0, "No BTC balance to send");

        (bool sent,) = recipient.call{value: btcBalance}("");
        require(sent, "Native transfer failed");

        uint256 mezoBalance = IMEZO(mezoPrecompile).balanceOf(address(this));
        require(mezoBalance > 0, "No MEZO balance to transfer");

        bool success = IMEZO(mezoPrecompile).transfer(recipient, mezoBalance);
        require(success, "Transfer failed");
    }

    /// @notice Transfers the contract's MEZO balance then sends the entire contract's 
    ///         BTC balance natively to the recipient.
    function transferMEZOThenSendNative(address recipient) external {
        uint256 mezoBalance = IMEZO(mezoPrecompile).balanceOf(address(this));
        require(mezoBalance > 0, "No MEZO balance to transfer");

        bool success = IMEZO(mezoPrecompile).transfer(recipient, mezoBalance);
        require(success, "Transfer failed");

        uint256 btcBalance = address(this).balance;
        require(btcBalance > 0, "No BTC balance to send");

        (bool sent,) = recipient.call{value: btcBalance}("");
        require(sent, "Native transfer failed");
    }

    /// @notice Sends the contract's BTC balance natively to the recipient then pulls 
    ///         the given MEZO amount from the given sender.
    function sendNativeThenPullMEZO(address btcRecipient, address mezoSender, uint256 mezoAmount) external {
        uint256 btcBalance = address(this).balance;
        require(btcBalance > 0, "No BTC balance to send");

        (bool sent,) = btcRecipient.call{value: btcBalance}("");
        require(sent, "Native transfer failed");

        bool success = IMEZO(mezoPrecompile).transferFrom(mezoSender, address(this), mezoAmount);
        require(success, "Transfer failed");
    }

    /// @notice Pulls the given amount of MEZO from the sender then sends the entire 
    ///         contract's BTC balance natively to the recipient.
    function pullMEZOThenSendNative(address mezoSender, uint256 mezoAmount, address btcRecipient) external {
        bool success = IMEZO(mezoPrecompile).transferFrom(mezoSender, address(this), mezoAmount);
        require(success, "Transfer failed");

        uint256 btcBalance = address(this).balance;
        require(btcBalance > 0, "No BTC balance to send");

        (bool sent,) = btcRecipient.call{value: btcBalance}("");
        require(sent, "Native transfer failed");
    }

    /// @notice Transfers the contract's BTC balance (using ERC20 precompile) then transfers 
    ///         the contract's MEZO balance to the recipient.
    function transferBTCThenTransferMEZO(address recipient) external {
        uint256 btcBalance = IBTC(btcPrecompile).balanceOf(address(this));
        require(btcBalance > 0, "No BTC balance to transfer");

        bool success1 = IBTC(btcPrecompile).transfer(recipient, btcBalance);
        require(success1, "Transfer failed");

        uint256 mezoBalance = IMEZO(mezoPrecompile).balanceOf(address(this));
        require(mezoBalance > 0, "No MEZO balance to transfer");

        bool success2 = IMEZO(mezoPrecompile).transfer(recipient, mezoBalance);
        require(success2, "Transfer failed");
    }

    /// @notice Transfers the contract's MEZO balance then transfers the contract's BTC 
    ///         balance (using ERC20 precompile) to the recipient.
    function transferMEZOThenTransferBTC(address recipient) external {
        uint256 mezoBalance = IMEZO(mezoPrecompile).balanceOf(address(this));
        require(mezoBalance > 0, "No MEZO balance to transfer");

        bool success1 = IMEZO(mezoPrecompile).transfer(recipient, mezoBalance);
        require(success1, "Transfer failed");

        uint256 btcBalance = IBTC(btcPrecompile).balanceOf(address(this));
        require(btcBalance > 0, "No BTC balance to transfer");

        bool success2 = IBTC(btcPrecompile).transfer(recipient, btcBalance);
        require(success2, "Transfer failed");
    }

    /// @notice Pulls the given amount of BTC from the sender then pulls the given amount of 
    ///         MEZO from the sender.
    function pullBTCThenPullMEZO(address sender, uint256 btcAmount, uint256 mezoAmount) external {
        bool success1 = IBTC(btcPrecompile).transferFrom(sender, address(this), btcAmount);
        require(success1, "Transfer failed");

        bool success2 = IMEZO(mezoPrecompile).transferFrom(sender, address(this), mezoAmount);
        require(success2, "Transfer failed");
    }
    
    /// @notice Pulls the given amount of MEZO from the sender then pulls the given amount of 
    ///         BTC from the sender.
    function pullMEZOThenPullBTC(address sender, uint256 mezoAmount, uint256 btcAmount) external {
        bool success1 = IMEZO(mezoPrecompile).transferFrom(sender, address(this), mezoAmount);
        require(success1, "Transfer failed");

        bool success2 = IBTC(btcPrecompile).transferFrom(sender, address(this), btcAmount);
        require(success2, "Transfer failed");
    }

    /// @notice Sends half of the contract's BTC balance natively then transfers another half
    ///         using the BTC ERC20 precompile then transfers the contract's MEZO balance 
    ///         to the recipient.
    function sendNativeThenTransferBTCThenTransferMEZO(address recipient) external {
        uint256 btcBalance = address(this).balance;
        require(btcBalance > 0, "No BTC balance to send");

        uint256 halfBalance = btcBalance / 2;

        (bool sent,) = recipient.call{value: halfBalance}("");
        require(sent, "Native transfer failed");    

        bool success1 = IBTC(btcPrecompile).transfer(recipient, halfBalance);
        require(success1, "Transfer failed");

        uint256 mezoBalance = IMEZO(mezoPrecompile).balanceOf(address(this));
        require(mezoBalance > 0, "No MEZO balance to transfer");

        bool success2 = IMEZO(mezoPrecompile).transfer(recipient, mezoBalance);
        require(success2, "Transfer failed");
    }
}
