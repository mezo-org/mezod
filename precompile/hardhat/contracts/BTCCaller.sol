// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

import { IBTC } from "../interfaces/IBTC.sol";

contract BTCCaller is IBTC {
    address private constant precompile = 0x7b7C000000000000000000000000000000000000;

    function name() external view returns (string memory) {
        return IBTC(precompile).name();
    }

    function symbol() external view returns (string memory) {
        return IBTC(precompile).symbol();
    }

    function decimals() external view returns (uint8) {
        return IBTC(precompile).decimals();
    }

    function totalSupply() external view returns (uint256) {
        return IBTC(precompile).totalSupply();
    }

    function balanceOf(address account) external view returns (uint256) {
        return IBTC(precompile).balanceOf(account);
    }

    function transfer(address to, uint256 value) external returns (bool) {
        return IBTC(precompile).transfer(to, value);
    }

    function allowance(address owner, address spender) external view returns (uint256) {
        return IBTC(precompile).allowance(owner, spender);
    }

    function approve(address spender, uint256 value) external returns (bool) {
        return IBTC(precompile).approve(spender, value);
    }

    function transferFrom(address from, address to, uint256 value) external returns (bool) {
        return IBTC(precompile).transferFrom(from, to, value);
    }

    function permit(
        address owner,
        address spender,
        uint256 amount,
        uint256 deadline,
        uint8 v,
        bytes32 r,
        bytes32 s
    ) external returns (bool) {
        return IBTC(precompile).permit(owner, spender, amount, deadline, v, r, s);
    }

    function DOMAIN_SEPARATOR() external view returns (bytes32) {
        return IBTC(precompile).DOMAIN_SEPARATOR();
    }

    // Deprecated as it is not compatible with EIP-2612.
    // Should be removed in the future.
    function nonce(address owner) external view returns (uint256) {
        return IBTC(precompile).nonce(owner);
    }

    function nonces(address owner) external view returns (uint256) {
        return IBTC(precompile).nonces(owner);
    }

    function PERMIT_TYPEHASH() external pure returns (bytes32) {
        return IBTC(precompile).PERMIT_TYPEHASH();
    }
}