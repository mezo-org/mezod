// SPDX-License-Identifier: MIT
pragma solidity ^0.8.28;

contract Forwarder {
    function run(address to) external payable returns (bool ok) {
        (ok, ) = to.call{value: msg.value}("");
    }
}
