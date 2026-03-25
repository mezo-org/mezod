// SPDX-License-Identifier: MIT
pragma solidity ^0.8.28;

interface ITransientStorageReader {
    function load(uint256 key) external view returns (uint256 loaded);
}

contract TransientStorageReader {
    /// @notice Stores the last value this contract observed during a test path.
    uint256 public lastLoaded;

    function writeAndLoad(uint256 key, uint256 value)
        external
        returns (uint256 loaded)
    {
        assembly {
            tstore(key, value)
            loaded := tload(key)
        }
    }

    function storeLoaded(uint256 key) external returns (uint256 loaded) {
        assembly {
            loaded := tload(key)
        }
        lastLoaded = loaded;
    }

    function load(uint256 key) external view returns (uint256 loaded) {
        assembly {
            loaded := tload(key)
        }
    }
}

contract TransientStorageCheck {
    /// @notice Stores the last value observed during a state-changing test path.
    /// @dev We use this because transient storage is cleared after each tx.
    uint256 public lastLoaded;

    /// @notice Stores a transient value and reads it back in the same call.
    /// @dev lastLoaded captures the result so tests can inspect it later.
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
    /// @dev lastLoaded captures the result across same-contract call frames.
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

    /// @notice Stores transient value, then asks another contract to read it.
    /// @dev lastLoaded captures that another contract sees its own storage.
    function setAndOtherContractLoad(
        address other,
        uint256 key,
        uint256 value
    ) external returns (uint256 loaded) {
        assembly {
            tstore(key, value)
        }
        loaded = ITransientStorageReader(other).load(key);
        lastLoaded = loaded;
    }

    /// @notice Writes via DELEGATECALL, then reads from this contract and other.
    /// @dev EIP-1153 assigns transient storage to the caller under DELEGATECALL.
    function delegateCallSetAndLoad(
        address target,
        uint256 key,
        uint256 value
    ) external returns (uint256 loaded) {
        (bool success, ) = target.delegatecall(
            abi.encodeWithSelector(
                TransientStorageReader.writeAndLoad.selector,
                key,
                value
            )
        );
        require(success, "DELEGATECALL failed");

        loaded = this.load(key);
        lastLoaded = loaded;
        TransientStorageReader(target).storeLoaded(key);
    }

    /// @notice Writes via CALLCODE, then reads from this contract and other.
    /// @dev EIP-1153 assigns transient storage to the caller under CALLCODE too.
    function callCodeSetAndLoad(
        address target,
        uint256 key,
        uint256 value
    ) external returns (uint256 loaded) {
        bytes memory payload = abi.encodeWithSelector(
            TransientStorageReader.writeAndLoad.selector,
            key,
            value
        );
        bool success;
        assembly {
            success := callcode(
                gas(),
                target,
                0,
                add(payload, 0x20),
                mload(payload),
                0,
                0
            )
        }
        require(success, "CALLCODE failed");

        loaded = this.load(key);
        lastLoaded = loaded;
        TransientStorageReader(target).storeLoaded(key);
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
