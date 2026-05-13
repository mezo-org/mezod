// SPDX-License-Identifier: MIT
pragma solidity ^0.8.28;

contract Bls12381Check {
    function staticCall(address target, bytes calldata payload)
        external
        view
        returns (bool ok, bytes memory out)
    {
        (ok, out) = target.staticcall(payload);
    }
}
