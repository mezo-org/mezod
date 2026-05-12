// SPDX-License-Identifier: MIT
pragma solidity ^0.8.28;

import {Eip7702TargetV1} from "./Eip7702TargetV1.sol";

// V2 exists solely to provide a second delegation target at a distinct
// deployment address with a distinguishable surface. `tickV2` is the
// behavioural discriminator used by the rotation test: it returns a
// constant V1 cannot produce. V1 inherits a no-op `fallback`, so the
// mere fact that an unknown selector "succeeds" against V1 proves
// nothing — only a successful ABI-decode of the returned `2` does.
contract Eip7702TargetV2 is Eip7702TargetV1 {
    function tickV2() external pure returns (uint256) {
        return 2;
    }
}
