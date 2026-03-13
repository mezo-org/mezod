// SPDX-License-Identifier: MIT
pragma solidity ^0.8.28;

contract Push0Check {
    /// @notice Returns zero from ordinary Solidity code.
    /// @dev Solidity should compile this to PUSH0 on Shanghai or later.
    function zero() external pure returns (uint256) {
        return 0; // make it use PUSH0 in Shanghai or later
    }

    /// @notice Returns zero from inline assembly.
    /// @dev This gives another execution path that uses PUSH0.
    function zeroFromAssembly() external pure returns (uint256 value) {
        assembly {
            value := 0
        }
    }
}
