// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

contract McopyCheck {
    /// @notice Copies arbitrary bytes using the Cancun `MCOPY` opcode.
    /// @dev If `MCOPY` is unsupported or behaves incorrectly, output will not
    ///      match input.
    function copy(bytes memory input) external pure returns (bytes memory out) {
        // Allocate a new result buffer with the same length as the input.
        out = new bytes(input.length);
        assembly {
            // Read input length.
            let len := mload(input)
            // Point to input bytes payload (skip the length word).
            let src := add(input, 0x20)
            // Point to output bytes payload (skip the length word).
            let dst := add(out, 0x20)
            // EIP-5656: copy `len` bytes from source data to destination data.
            mcopy(dst, src, len)
        }
    }

    /// @notice Copies overlapping memory ranges where destination is after source.
    /// @dev EIP-5656 requires memmove-like behavior (as if using an intermediate
    /// buffer), so overlap must not corrupt source bytes.
    function overlapCopyForward() external pure returns (bytes memory out) {
        out = hex"0102030405060708";
        assembly {
            let ptr := add(out, 0x20)
            // Copy 6 bytes from [0..5] to [2..7].
            mcopy(add(ptr, 2), ptr, 6)
        }
    }

    /// @notice Copies overlapping memory ranges where destination is before source.
    function overlapCopyBackward() external pure returns (bytes memory out) {
        out = hex"0102030405060708";
        assembly {
            let ptr := add(out, 0x20)
            // Copy 6 bytes from [2..7] to [0..5].
            mcopy(ptr, add(ptr, 2), 6)
        }
    }

    /// @notice Performs a zero-length overlapping copy.
    /// @dev Zero-length MCOPY should be a no-op even with overlapping ranges.
    function zeroLengthOverlapCopy() external pure returns (bytes memory out) {
        out = hex"0102030405060708";
        assembly {
            let ptr := add(out, 0x20)
            mcopy(add(ptr, 2), ptr, 0)
        }
    }
}
