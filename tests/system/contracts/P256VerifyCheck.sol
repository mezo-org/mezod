// SPDX-License-Identifier: MIT
pragma solidity ^0.8.28;

/// @notice Surface helper for the P256VERIFY precompile at 0x0100 added by
/// EIP-7951 (secp256r1 signature verification). Mirrors the static-call
/// shim used by Bls12381Check so the test file can drive the precompile
/// directly from JS.
contract P256VerifyCheck {
    function staticCall(address target, bytes calldata payload)
        external
        view
        returns (bool ok, bytes memory out)
    {
        (ok, out) = target.staticcall(payload);
    }
}
