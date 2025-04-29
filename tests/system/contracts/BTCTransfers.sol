// SPDX-License-Identifier: MIT
pragma solidity ^0.8.28;

import {IBTC} from "./interfaces/IBTC.sol";

interface ITestbed {
    function transferWithRevert(address to, uint256 value) external returns (bool);
}

/// @title BTCTransfers
/// @notice Handles various transfer scenarios for Mezo native token - BTC.
contract BTCTransfers {
    // BTC ERC-20 token address on Mezo
    address private constant precompile = 0x7b7C000000000000000000000000000000000000;
    address private constant testbedPrecompile = 0x7b7c100000000000000000000000000000000000;
    uint256 public balanceTracker = 1;

    /// @notice Transfers native BTC and then ERC-20 Token from the contract
    ///         which was previously funded.
    function nativeThenERC20(address recipient) external {
        uint256 balance = IBTC(precompile).balanceOf(address(this));
        require(balance > 0, "No balance to transfer");

        uint256 halfBalance = balance / 2;

        // Transfer native BTC
        (bool sent,) = recipient.call{value: halfBalance}("");
        require(sent, "Transfer using call failed");

        // Transfer ERC-20
        bool success = IBTC(precompile).transfer(recipient, halfBalance);
        require(success, "Transfer using transfer failed");
    }

    /// @notice Transfers ERC-20 Token and then native BTC from the contract
    ///         which was previously funded.
    function erc20ThenNative(address recipient) external {
        uint256 balance = IBTC(precompile).balanceOf(address(this));
        require(balance > 0, "No balance to transfer");

        uint256 halfBalance = balance / 2;

        // Transfer ERC-20
        bool success = IBTC(precompile).transfer(recipient, halfBalance);
        require(success, "Transfer using transfer failed");

        // Transfer native BTC
        (bool sent,) = recipient.call{value: halfBalance}("");
        require(sent, "Transfer using call failed");
    }

    /// @notice Receive native then transfer all received native amount as
    ///         native BTC.
    function receiveSendNative(address payable recipient) external payable {
        require(msg.value > 0, "Must send some native tokens");

        // Transfer native BTC
        (bool sent,) = recipient.call{value: msg.value}("");
        require(sent, "Native transfer failed");
    }

    /// @notice Receive native then transfer all received native amount as
    ///         ERC-20 Token.
    function receiveSendERC20(address payable recipient) external payable {
        require(msg.value > 0, "Must send some native tokens");

        // Transfer ERC-20
        bool success = IBTC(precompile).transfer(recipient, msg.value);
        require(success, "ERC-20 transfer failed");
    }

    /// @notice Receive native then transfer half as native BTC and half as
    ///         ERC-20 Token. All the transfers should be funded from the
    ///         received native amount.
    function receiveSendNativeThenERC20(address payable recipient) external payable {
        require(msg.value > 0, "Must send some native tokens");

        uint256 halfAmount = msg.value / 2;

        // Transfer native BTC
        (bool sent,) = recipient.call{value: halfAmount}("");
        require(sent, "Native transfer failed");

        // Transfer ERC-20
        bool success = IBTC(precompile).transfer(recipient, halfAmount);
        require(success, "ERC-20 transfer failed");
    }

    /// @notice Receive native then transfer half as ERC-20 Token and half as
    ///         native BTC. All the transfers should be funded from the
    ///         received native amount.
    function receiveSendERC20ThenNative(address payable recipient) external payable {
        require(msg.value > 0, "Must send some native tokens");

        uint256 halfAmount = msg.value / 2;

        // Transfer ERC-20
        bool success = IBTC(precompile).transfer(recipient, halfAmount);
        require(success, "ERC-20 transfer failed");

        // Transfer native BTC
        (bool sent,) = recipient.call{value: halfAmount}("");
        require(sent, "Native transfer failed");
    }

    /// @notice Transfers ERC-20 Token and native BTC multiple times.
    ///         All the transfers should be funded from the received native amount.
    function receiveSendMultiple(address payable recipient) external payable {
        require(msg.value > 0, "Must send some native tokens");

        uint256 quarterAmount = msg.value / 4;

        // Transfer 1/4 of the token amount as ERC-20
        bool success1 = IBTC(precompile).transfer(recipient, quarterAmount);
        require(success1, "ERC-20 transfer failed");

        // Transfer 1/4 of the native BTC
        (bool sent1,) = recipient.call{value: quarterAmount}("");
        require(sent1, "Native transfer failed");

        // Transfer 1/4 of the token amount as ERC-20
        bool success2 = IBTC(precompile).transfer(recipient, quarterAmount);
        require(success2, "ERC-20 transfer failed");

        // Transfer 1/4 of the native BTC
        (bool sent2,) = recipient.call{value: quarterAmount}("");
        require(sent2, "Native transfer failed");
    }

    /// @notice Transfers with revert
    function receiveSendRevert(address payable recipient) external payable {
        require(msg.value > 0, "Must send some native tokens");

        uint256 halfAmount = msg.value / 2;

        // Transfer ERC-20
        bool success = IBTC(precompile).transfer(recipient, halfAmount);
        require(success, "ERC-20 transfer failed");

        // Transfer native BTC
        (bool sent,) = recipient.call{value: halfAmount}("");
        require(sent, "Native transfer failed");

        revert("revert after transfers");
    }

    /// @notice Transfers native BTC then ERC-20 Token within the same transaction.
    ///         All funds for native BTC transfer should come from native tokens received.
    ///         All the funds for ERC20 transfer should be approved by the sender and
    ///         this function should faciliate the transfer.
    function receiveSendNativeThenPullERC20(address payable recipient, uint256 tokenAmount) external payable {
        require(msg.value > 0, "Must send some native tokens");

        // Transfer native BTC
        (bool sent,) = recipient.call{value: msg.value}("");
        require(sent, "Native transfer failed");

        // Transfer ERC-20
        bool success = IBTC(precompile).transferFrom(msg.sender, recipient, tokenAmount);
        require(success, "ERC-20 transfer failed");
    }

    /// @notice Transfers ERC-20 Token and then native BTC within the same transaction.
    ///         All funds for native BTC transfer should come from native tokens received.
    ///         All the funds for ERC20 transfer should be approved by the sender and
    ///         this function should faciliate the transfer.
    function receivePullERC20ThenNative(address payable recipient, uint256 tokenAmount) external payable {
        require(msg.value > 0, "Must send some native tokens");

        // Transfer ERC-20
        bool success = IBTC(precompile).transferFrom(msg.sender, recipient, tokenAmount);
        require(success, "ERC-20 transfer failed");

        // Transfer native BTC
        (bool sent,) = recipient.call{value: msg.value}("");
        require(sent, "Native transfer failed");
    }

    /// @notice Update the storage variable, pull ERC20 from sender and
    ///         then reset the storage variable to its original value.
    function stateChangeThenPullERC20(uint256 amount, bool resetStateChange) external {
        balanceTracker = 42; // This is an arbitrary value for testing purposes only.

        // Transfer ERC-20.
        IBTC(precompile).transferFrom(msg.sender, address(this), amount);

        if (resetStateChange) {
            balanceTracker = 1; // Reset the storage variable to its original value.
        }
    }

    /// @notice Calls multiple precompile in order to reach the maximum
    ///         precompile calls allowed per transaction (10).
    ///         we'll do 1 balanceOf initial call, 9 transfer, all in the
    ///         limits, and breach the limits with a final transfer call.
    function erc20RevertsWhenExceedMaxPrecompileCalls(address recipient) external {
        // this accounts for the first precompile call
        uint256 balance = IBTC(precompile).balanceOf(address(this));
        require(balance > 0, "No balance to transfer");

        uint256 tenthBalance = balance / 10;

        // these takes up to the limit
        bool success = IBTC(precompile).transfer(recipient, tenthBalance);
        require(success, "Transfer using transfer failed");
        success = IBTC(precompile).transfer(recipient, tenthBalance);
        require(success, "Transfer using transfer failed");
        success = IBTC(precompile).transfer(recipient, tenthBalance);
        require(success, "Transfer using transfer failed");
        success = IBTC(precompile).transfer(recipient, tenthBalance);
        require(success, "Transfer using transfer failed");
        success = IBTC(precompile).transfer(recipient, tenthBalance);
        require(success, "Transfer using transfer failed");
        success = IBTC(precompile).transfer(recipient, tenthBalance);
        require(success, "Transfer using transfer failed");
        success = IBTC(precompile).transfer(recipient, tenthBalance);
        require(success, "Transfer using transfer failed");
        success = IBTC(precompile).transfer(recipient, tenthBalance);
        require(success, "Transfer using transfer failed");
        success = IBTC(precompile).transfer(recipient, tenthBalance);
        require(success, "Transfer using transfer failed");
        // this will revert
        IBTC(precompile).transfer(recipient, tenthBalance);
    }

    /// @notice Doing a single precompile call that transfers funds and reverts.
    function revertingInPrecompile(address recipient) external {
        uint256 balance = address(this).balance;
        require(balance > 0, "No balance to transfer");

        // Transfer with revert now.
        ITestbed(testbedPrecompile).transferWithRevert(recipient, balance);
    }

    /// @notice Transfers ERC-20 Token from the contract
    ///         then call a precompile which reverts
    ///         revert fully and all state as per pre-call of the method
    function erc20ThenRevertingInPrecompile(address recipient) external {
        uint256 balance = IBTC(precompile).balanceOf(address(this));
        require(balance > 0, "No balance to transfer");

        uint256 halfBalance = balance / 2;

        // Transfer ERC-20
        bool success = IBTC(precompile).transfer(recipient, halfBalance);
        require(success, "Transfer using transfer failed");

        // Transfer with revert now.
        ITestbed(testbedPrecompile).transferWithRevert(recipient, halfBalance);
    }

    /// @notice Transfers  ERC-20 Token from the contract
    ///         then call a second function which will move funds and
    ///         revert in turn
    function erc20ThenRevertingExternalCall(address recipient) external {
        uint256 balance = IBTC(precompile).balanceOf(address(this));
        require(balance > 0, "No balance to transfer");

        uint256 halfBalance = balance / 2;

        // Transfer ERC-20
        bool success = IBTC(precompile).transfer(recipient, halfBalance);
        require(success, "Transfer using transfer failed");

        // create the contract
        RevertingTransfer revContract = new RevertingTransfer();

        // Transfer ERC-20
        success = IBTC(precompile).transfer(address(revContract), halfBalance);
        require(success, "Transfer using transfer failed");

        // call it with a try catch
        try revContract.transferThenRevert(recipient) {
            // nothing to do
        } catch Error(string memory reason) {
            require(keccak256(bytes("some unexpected error")) == keccak256(bytes(reason)));
        }
    }

    /// @notice Transfers  ERC-20 Token from the contract
    ///         then call a second function which will move funds and
    ///         revert in turn
    function revertingExternalCallThenERC20Transfer(address recipient) external {
        uint256 balance = IBTC(precompile).balanceOf(address(this));
        require(balance > 0, "No balance to transfer");

        uint256 halfBalance = balance / 2;

        // create the contract
        RevertingTransfer revContract = new RevertingTransfer();

        // Transfer ERC-20
        bool success = IBTC(precompile).transfer(address(revContract), halfBalance);
        require(success, "Transfer using transfer failed");

        // call it with a try catch
        try revContract.transferThenRevert(recipient) {
            // nothing to do
        } catch Error(string memory reason) {
            require(keccak256(bytes("some unexpected error")) == keccak256(bytes(reason)));
        }

        // Transfer ERC-20
        success = IBTC(precompile).transfer(recipient, halfBalance);
        require(success, "Transfer using transfer failed");
    }

    /// @notice Transfers Send ERC20 from the contract to the recipient
    ///         then call a second function which will move funds and
    ///         revert inside a precompile, final balance is == to initial transfer
    function erc20ThenRevertingExternalCallInPrecompile(address recipient) external {
        uint256 balance = IBTC(precompile).balanceOf(address(this));
        require(balance > 0, "No balance to transfer");

        uint256 halfBalance = balance / 2;

        // Transfer ERC-20
        bool success = IBTC(precompile).transfer(recipient, halfBalance);
        require(success, "Transfer using transfer failed");

        // create the contract
        RevertingTransfer revContract = new RevertingTransfer();

        // Transfer ERC-20
        success = IBTC(precompile).transfer(address(revContract), halfBalance);
        require(success, "Transfer using transfer failed");

        // call it with a try catch
        try revContract.transferWithPrecompileRevert(recipient) {
            // nothing to do
        } catch {}
    }

    /// @notice Transfers Send ERC20 from the contract to the recipient
    ///         then call a second function which will move funds and
    ///         revert inside a precompile, final balance is == to initial transfer
    function erc20ThenRevertingExternalCallWithMultiplePrecompile(address recipient) external {
        uint256 balance = IBTC(precompile).balanceOf(address(this));
        require(balance > 0, "No balance to transfer");

        uint256 halfBalance = balance / 2;

        // Transfer ERC-20
        bool success = IBTC(precompile).transfer(recipient, halfBalance);
        require(success, "Transfer using transfer failed");

        // create the contract
        RevertingTransfer revContract = new RevertingTransfer();

        // Transfer ERC-20
        success = IBTC(precompile).transfer(address(revContract), halfBalance);
        require(success, "Transfer using transfer failed");

        // call it with a try catch
        try revContract.multipleTransferWithPrecompileRevert(recipient) {
            // nothing to do
        } catch {}
    }

    /// @notice Transfers Send ERC20 from the contract to the recipient
    ///         then call a second function which will move funds and
    ///         revert inside a precompile, final balance is == to initial transfer
    function revertingExternalCallInPrecompileThenERC20(address recipient) external {
        uint256 balance = IBTC(precompile).balanceOf(address(this));
        require(balance > 0, "No balance to transfer");

        uint256 halfBalance = balance / 2;

        // create the contract
        RevertingTransfer revContract = new RevertingTransfer();

        // Transfer ERC-20
        bool success = IBTC(precompile).transfer(address(revContract), halfBalance);
        require(success, "Transfer using transfer failed");

        // call it with a try catch
        try revContract.transferWithPrecompileRevert(recipient) {
            // nothing to do
        } catch {}

        // Transfer ERC-20
        success = IBTC(precompile).transfer(recipient, halfBalance);
        require(success, "Transfer using transfer failed");
    }
}

contract RevertingTransfer {
    // BTC ERC-20 token address on Mezo
    address private constant precompile = 0x7b7C000000000000000000000000000000000000;
    address private constant testbedPrecompile = 0x7b7c100000000000000000000000000000000000;

    function transferThenRevert(address recipient) external {
        uint256 balance = IBTC(precompile).balanceOf(address(this));
        require(balance > 0, "No balance to transfer");

        // Transfer all remaining ERC-20
        bool success = IBTC(precompile).transfer(recipient, balance);
        require(success, "Transfer using transfer failed");

        revert("some unexpected error");
    }

    function transferWithPrecompileRevert(address recipient) external {
        uint256 balance = IBTC(precompile).balanceOf(address(this));
        require(balance > 0, "No balance to transfer");

        // Transfer all remaining ERC-20
        ITestbed(testbedPrecompile).transferWithRevert(recipient, balance);
    }

    function multipleTransferWithPrecompileRevert(address recipient) external {
        uint256 balance = IBTC(precompile).balanceOf(address(this));
        require(balance > 0, "No balance to transfer");

        uint256 fourthBalance = balance / 4;

        // Transfer some ERC-20
        bool success = IBTC(precompile).transfer(recipient, fourthBalance);
        require(success, "Transfer using transfer failed");

        // Transfer some ERC-20
        success = IBTC(precompile).transfer(recipient, fourthBalance);
        require(success, "Transfer using transfer failed");

        // Transfer some ERC-20
        success = IBTC(precompile).transfer(recipient, fourthBalance);
        require(success, "Transfer using transfer failed");

        // Transfer all remaining ERC-20 with panic
        ITestbed(testbedPrecompile).transferWithRevert(recipient, fourthBalance);
    }
}
