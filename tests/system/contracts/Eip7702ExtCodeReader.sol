// SPDX-License-Identifier: MIT
pragma solidity ^0.8.28;

contract Eip7702ExtCodeReader {
    function sizeOf(address a) external view returns (uint256 s) {
        assembly {
            s := extcodesize(a)
        }
    }

    function hashOf(address a) external view returns (bytes32 h) {
        assembly {
            h := extcodehash(a)
        }
    }

    function copyOf(address a) external view returns (bytes memory out) {
        uint256 s;
        assembly {
            s := extcodesize(a)
        }
        out = new bytes(s);
        assembly {
            extcodecopy(a, add(out, 0x20), 0, s)
        }
    }
}
