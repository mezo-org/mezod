// SPDX-License-Identifier: MIT
pragma solidity ^0.8.28;

// Eip7702Caller invokes a target via CALL / STATICCALL / DELEGATECALL so
// tests can pin geth's one-level delegation resolution rule from inside
// an EVM call frame (i.e. not just from a top-level EOA-originated tx).
contract Eip7702Caller {
    event Result(bool ok, bytes ret);
    event CallGas(uint256 gasFirst, uint256 gasSecond);

    function callInto(address target, bytes calldata payload)
        external
        returns (bool ok, bytes memory ret)
    {
        (ok, ret) = target.call(payload);
        emit Result(ok, ret);
    }

    // Two sequential CALLs to the same target with the same payload,
    // measured via gasleft(). Both calls happen inside the same tx so the
    // EIP-2929 access list survives between them: the first call sees the
    // target (and, if delegated, the delegate) as cold; the second sees
    // them as warm. Emitted as CallGas so the harness can decode it from
    // the receipt.
    function callTwiceMeasured(address target, bytes calldata payload)
        external
        returns (uint256 gasFirst, uint256 gasSecond)
    {
        uint256 g0 = gasleft();
        (bool ok1, ) = target.call(payload);
        uint256 g1 = gasleft();
        (bool ok2, ) = target.call(payload);
        uint256 g2 = gasleft();
        require(ok1 && ok2, "call failed");
        gasFirst = g0 - g1;
        gasSecond = g1 - g2;
        emit CallGas(gasFirst, gasSecond);
    }

    function staticCallInto(address target, bytes calldata payload)
        external
        view
        returns (bool ok, bytes memory ret)
    {
        (ok, ret) = target.staticcall(payload);
    }

    function delegateCallInto(address target, bytes calldata payload)
        external
        returns (bool ok, bytes memory ret)
    {
        (ok, ret) = target.delegatecall(payload);
        emit Result(ok, ret);
    }

    function readSlot(uint256 k) external view returns (uint256 v) {
        assembly {
            v := sload(k)
        }
    }
}
