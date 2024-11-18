// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

import { IValidatorPool, Description, Privilege } from "../interfaces/IValidatorPool.sol";

contract ValidatorPoolCaller is IValidatorPool {
    address private constant precompile = 0x7B7C000000000000000000000000000000000011;

    function acceptOwnership() external returns (bool) {
        return IValidatorPool(precompile).acceptOwnership();
    }

    function application(address operator) external view returns (bytes32 consPubKey, Description memory description) {
        return IValidatorPool(precompile).application(operator);
    }

    function applications() external view returns (address[] memory) {
        return IValidatorPool(precompile).applications();
    }

    function approveApplication(address operator) external returns (bool) {
        return IValidatorPool(precompile).approveApplication(operator);
    }

    function candidateOwner() external view returns (address) {
        return IValidatorPool(precompile).candidateOwner();
    }

    function cleanupApplications() external returns (bool) {
        return IValidatorPool(precompile).cleanupApplications();
    }

    function kick(address operator) external returns (bool) {
        return IValidatorPool(precompile).kick(operator);
    }

    function leave() external returns (bool) {
        return IValidatorPool(precompile).leave();
    }

    function owner() external view returns (address) {
        return IValidatorPool(precompile).owner();
    }

    function submitApplication(
        bytes32 consPubKey,
        Description calldata description
    ) external returns (bool) {
        return IValidatorPool(precompile).submitApplication(consPubKey, description);
    }

    function transferOwnership(address newOwner) external returns (bool) {
        return IValidatorPool(precompile).transferOwnership(newOwner);
    }

    function validator(address operator) external view returns (bytes32 consPubKey, Description memory description) {
        return IValidatorPool(precompile).validator(operator);
    }

    function validators() external view returns (address[] memory) {
        return IValidatorPool(precompile).validators();
    }

    function addPrivilege(
        address[] calldata operators,
        uint8 privilegeId
    ) external returns (bool) {
        return IValidatorPool(precompile).addPrivilege(operators, privilegeId);
    }

    function removePrivilege(
        address[] calldata operators,
        uint8 privilegeId
    ) external returns (bool) {
        return IValidatorPool(precompile).removePrivilege(operators, privilegeId);
    }

    function validatorsByPrivilege(
        uint8 privilegeId
    ) external view returns (address[] memory operators) {
        return IValidatorPool(precompile).validatorsByPrivilege(privilegeId);
    }

    function privileges() external view returns (Privilege[] memory) {
        return IValidatorPool(precompile).privileges();
    }
}