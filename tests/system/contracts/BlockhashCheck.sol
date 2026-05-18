// SPDX-License-Identifier: MIT
pragma solidity ^0.8.28;

/// @notice Surface helper to drive the BLOCKHASH opcode from JS, used by
/// the EIP-2935 Mezo-divergence test. Mezo resolves BLOCKHASH through
/// x/poa's historical info store rather than the EIP-2935 history-storage
/// system contract, so the test exercises the opcode and asserts the
/// outcome against eth_getBlockByNumber.
contract BlockhashCheck {
    function blockHashOf(uint256 n) external view returns (bytes32) {
        return blockhash(n);
    }
}
