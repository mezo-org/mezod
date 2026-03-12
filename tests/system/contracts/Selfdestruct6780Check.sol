// SPDX-License-Identifier: MIT
pragma solidity ^0.8.28;

contract DestructibleContract6780 {
    constructor() payable {}

    receive() external payable {}

    /// @notice Returns a constant so we can verify code still exists.
    function ping() external pure returns (uint256) {
        return 1;
    }

    /// @notice Calls SELFDESTRUCT and sends remaining balance to beneficiary.
    function destroy(address payable beneficiary) external {
        selfdestruct(beneficiary);
    }
}

/// @notice Demonstrates EIP-6780 SELFDESTRUCT behavior.
/// @dev Post-Cancun:
/// - destroying an old contract (created in a prior tx) should keep its code,
/// - destroying a contract in the same tx as creation should remove its code.
contract Selfdestruct6780Check {
    address public lastDestructible;

    /// @notice Creates a destructible contract in one transaction.
    function createDestructible() external returns (address destructible) {
        destructible = address(new DestructibleContract6780());
        lastDestructible = destructible;
    }

    /// @notice Destroys an already-existing contract in a later transaction.
    /// @dev Under EIP-6780 this should not delete contract code.
    function destroyExisting(address destructible, address payable beneficiary)
        external
    {
        DestructibleContract6780(payable(destructible)).destroy(beneficiary);
    }

    /// @notice Creates and destroys a contract in the same transaction.
    /// @dev Under EIP-6780 this path should delete contract code.
    function createAndDestroySameTx(address payable beneficiary)
        external
        payable
        returns (address destructible)
    {
        DestructibleContract6780 created =
            new DestructibleContract6780{value: msg.value}();
        destructible = address(created);
        lastDestructible = destructible;
        created.destroy(beneficiary);
    }
}
