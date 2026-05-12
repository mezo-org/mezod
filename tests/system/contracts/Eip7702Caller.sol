// SPDX-License-Identifier: MIT
pragma solidity ^0.8.28;

// Eip7702Caller invokes a target via CALL / STATICCALL / DELEGATECALL so
// tests can pin geth's one-level delegation resolution rule from inside
// an EVM call frame (i.e. not just from a top-level EOA-originated tx).
contract Eip7702Caller {
    event Result(bool ok, bytes ret);

    function callInto(address target, bytes calldata payload)
        external
        returns (bool ok, bytes memory ret)
    {
        (ok, ret) = target.call(payload);
        emit Result(ok, ret);
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
