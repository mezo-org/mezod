// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

contract TransientStorageCheck {
    uint256 public lastLoaded;

    /// @notice Stores a transient value and reads it back in the same call.
    /// @dev This proves TSTORE + TLOAD work within one transaction.
    function setAndLoad(uint256 key, uint256 value)
        external
        returns (uint256 loaded)
    {
        assembly {
            tstore(key, value)
            loaded := tload(key)
        }
        lastLoaded = loaded;
    }

    /// @notice Reads a transient value for a key.
    /// @dev If called in a new transaction, the value should be zero because
    /// transient storage is cleared after each transaction.
    function load(uint256 key) external view returns (uint256 loaded) {
        assembly {
            loaded := tload(key)
        }
    }

    /// @notice Stores transient value, then reads it via an external call.
    /// @dev This checks that transient storage is shared across call frames of
    /// the same contract inside one transaction.
    function setAndExternalLoad(uint256 key, uint256 value)
        external
        returns (uint256 loaded)
    {
        assembly {
            tstore(key, value)
        }
        loaded = this.load(key);
        lastLoaded = loaded;
    }

    /// @notice Performs a STATICCALL to setAndLoad.
    /// @dev EIP-1153 requires TSTORE to fail in static context.
    function staticCallSetAndLoad(uint256 key, uint256 value)
        external
        view
        returns (bool success, bytes memory returndata)
    {
        (success, returndata) = address(this).staticcall(
            abi.encodeWithSelector(this.setAndLoad.selector, key, value)
        );
    }
}
