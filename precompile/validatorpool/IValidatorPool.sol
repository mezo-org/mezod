// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

struct Description {
    string moniker;
    string identity;
    string website;
    string securityContract;
    string details;
}

/// @title  IValidatorPool
/// @notice Interface for the Validator Pool
interface IValidatorPool {
    function getApplications() external view returns (address[] calldata);
    function getApplication(address operator) external view returns (address, bytes32, Description calldata);
    function submitApplication(bytes32 consPubKey, address operator, Description calldata description) external returns (bool);
    function approveApplication(address operator) external returns (bool);
    function getValidators() external view returns (address[] calldata);
    function getValidator(address operator) external view returns (address, bytes32, Description calldata);
    function leave() external returns (bool);
    function kick(address operator) external returns (bool);
    function owner() external view returns (address);
    function candidateOwner() external view returns (address);
    function transferOwnership(address newOwner) external returns (bool);
    function acceptOwnership() external returns (bool);
}