// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

import { IMEZO } from "../interfaces/IMEZO.sol";

contract MEZOCaller is IMEZO {
    address private constant precompile = 0x7B7c000000000000000000000000000000000001;

    function name() external view returns (string memory) {
        return IMEZO(precompile).name();
    }

    function symbol() external view returns (string memory) {
        return IMEZO(precompile).symbol();
    }

    function decimals() external view returns (uint8) {
        return IMEZO(precompile).decimals();
    }

    function totalSupply() external view returns (uint256) {
        return IMEZO(precompile).totalSupply();
    }

    function balanceOf(address account) external view returns (uint256) {
        return IMEZO(precompile).balanceOf(account);
    }

    function transfer(address to, uint256 value) external returns (bool) {
        return IMEZO(precompile).transfer(to, value);
    }

    function allowance(address owner, address spender) external view returns (uint256) {
        return IMEZO(precompile).allowance(owner, spender);
    }

    function approve(address spender, uint256 value) external returns (bool) {
        return IMEZO(precompile).approve(spender, value);
    }

    function transferFrom(address from, address to, uint256 value) external returns (bool) {
        return IMEZO(precompile).transferFrom(from, to, value);
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
        return IMEZO(precompile).permit(owner, spender, amount, deadline, v, r, s);
    }

    function DOMAIN_SEPARATOR() external view returns (bytes32) {
        return IMEZO(precompile).DOMAIN_SEPARATOR();
    }

    function nonce(address owner) external view returns (uint256) {
        return IMEZO(precompile).nonce(owner);
    }

    function PERMIT_TYPEHASH() external pure returns (bytes32) {
        return IMEZO(precompile).PERMIT_TYPEHASH();
    }
}