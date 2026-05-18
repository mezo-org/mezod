// SPDX-License-Identifier: MIT
pragma solidity ^0.8.28;

/// @notice Surface helper for the MODEXP precompile at 0x05. Used by the
/// Osaka system tests to pin EIP-7823 (input upper bound) and EIP-7883
/// (gas cost increase).
contract ModexpCheck {
    function staticCall(address target, bytes calldata payload)
        external
        view
        returns (bool ok, bytes memory out)
    {
        (ok, out) = target.staticcall(payload);
    }

    function staticCallWithGas(address target, bytes calldata payload)
        external
        view
        returns (bool ok, bytes memory out, uint256 gasUsed)
    {
        uint256 g = gasleft();
        (ok, out) = target.staticcall(payload);
        gasUsed = g - gasleft();
    }
}
