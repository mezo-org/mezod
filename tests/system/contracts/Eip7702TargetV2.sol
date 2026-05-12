// SPDX-License-Identifier: MIT
pragma solidity ^0.8.28;

import {Eip7702TargetV1} from "./Eip7702TargetV1.sol";

// V2 exists solely to provide a second delegation target at a distinct
// deployment address with a distinguishable surface. `tickV2` is the
// behavioural discriminator used by the rotation test: V1 doesn't expose
// it, so calling it on the EOA proves the delegated code actually swapped.
contract Eip7702TargetV2 is Eip7702TargetV1 {
    function tickV2() external {}
}
