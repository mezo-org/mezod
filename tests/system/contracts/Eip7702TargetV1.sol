// SPDX-License-Identifier: MIT
pragma solidity ^0.8.28;

contract Eip7702TargetV1 {
    event Touched(address indexed sender, address indexed self, uint256 key, uint256 val);

    function setSlot(uint256 k, uint256 v) external {
        assembly {
            sstore(k, v)
        }
        emit Touched(msg.sender, address(this), k, v);
    }

    // Writes `count` consecutive SSTOREs starting at baseK. Used by
    // tests that need an EVM gas footprint above mezod's floor
    // (gasLimit * MinGasMultiplier = 0.5) so refund deltas are
    // observable; each SSTORE to a fresh slot costs 22100.
    function setSlotN(uint256 baseK, uint256 v, uint256 count) external {
        for (uint256 i = 0; i < count; i++) {
            uint256 slot = baseK + i;
            assembly {
                sstore(slot, v)
            }
        }
    }

    function readSlot(uint256 k) external view returns (uint256 v) {
        assembly {
            v := sload(k)
        }
    }

    function whoAmI() external view returns (address sender, address self) {
        return (msg.sender, address(this));
    }

    function tick() external {}

    // Empty calldata or unknown selector hitting the delegated EOA must
    // not revert; the tests that pin auth-loop side effects rely on the
    // outer Call landing with status 0x1 even when the tx body is empty.
    fallback() external payable {}
    receive() external payable {}
}
