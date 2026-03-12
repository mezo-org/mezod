// SPDX-License-Identifier: MIT
pragma solidity ^0.8.28;

/// @notice Minimal factory to test EIP-3860 initcode size limits.
/// @dev It deploys raw initcode filled with STOP (`0x00`) bytes.
contract InitcodeLimitCheck {
    address public lastDeployed;

    /// @notice Tries to deploy initcode of an exact byte size.
    /// @dev Reverts when CREATE fails (for example, size > 49152 under EIP-3860).
    function deployWithInitcodeSize(uint256 size)
        external
        returns (address deployed)
    {
        bytes memory initcode = new bytes(size);
        assembly {
            deployed := create(0, add(initcode, 0x20), mload(initcode))
        }

        require(deployed != address(0), "CREATE failed");
        lastDeployed = deployed;
    }

    /// @notice Tries to deploy initcode of an exact byte size using CREATE2.
    /// @dev Reverts when CREATE2 fails (for example, size > 49152 under EIP-3860).
    function deployWithInitcodeSizeCreate2(uint256 size, bytes32 salt)
        external
        returns (address deployed)
    {
        bytes memory initcode = new bytes(size);
        assembly {
            deployed := create2(0, add(initcode, 0x20), mload(initcode), salt)
        }

        require(deployed != address(0), "CREATE2 failed");
        lastDeployed = deployed;
    }
}
