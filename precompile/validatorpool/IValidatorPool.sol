// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

/// @title Description
/// @notice Description struct for a validator's `description` argument
struct Description {
    /// @notice moniker is the validator's name.
    string moniker;
    /// @notice identity is the optional identity signature (ex. UPort or Keybase).
    string identity;
    /// @notice website is the optional website link.
    string website;
    /// @notice securityContact is the optional security contact information.
    string securityContact;
    /// @notice details is the optional details about the validator. 
    string details;
}

/// @title Privilege
/// @notice Description of a validator's privilege.
struct Privilege {
    /// @notice id is the unique identifier of the privilege.
    uint8 id;
    /// @notice name is the name of the privilege.
    string name;
}

/// @title  IValidatorPool
/// @notice Interface for the ValidatorPool precompile
interface IValidatorPool {
    /** 
     * @notice Emitted when a validator application is approved by the contract owner
     * @param operator The validator applications operator address
     */ 
    event ApplicationApproved(address indexed operator);

    /** 
     * @notice Emitted when pending applications are cleaned by the contract owner
     */ 
    event ApplicationsCleaned();

    /** 
     * @notice Emitted when a new validator application is successfully submitted
     * @param operator The validators operator address
     * @param consPubKey The validators consensus public key
     * @param description The validators description (moniker, identity, website, securityContact & details)
     */ 
    event ApplicationSubmitted(address indexed operator, bytes32 indexed consPubKey, Description description);

    /** 
     * @notice Emitted when the contract owner starts the transferOwnership flow
     * @param previousOwner The current owner (soon to be previous)
     * @param newOwner The intended new owner (current candidateOwner)
     */ 
    event OwnershipTransferStarted(address indexed previousOwner, address indexed newOwner);
    
    /** 
     * @notice Emitted when the candidateOwner accepts ownership, completing the transferOwnership flow
     * @param previousOwner The previous owner
     * @param newOwner The new owner (previously candidateOwner)
     */ 
    event OwnershipTransferred(address indexed previousOwner, address indexed newOwner);
    
    /** 
     * @notice Emitted when a validator application is approved by the contract owner
     * @param operator The validator applications operator address
     */ 
    event ValidatorJoined(address indexed operator);

    /** 
     * @notice Emitted when the owner kicks a validator from the pool
     * @param operator The operator address of the validator being kicked
     */ 
    event ValidatorKicked(address indexed operator);

    /** 
     * @notice Emitted when a validator voluntarily leaves the pool
     * @param operator The operator address of the validator that left the pool
     */ 
    event ValidatorLeft(address indexed operator);

    /**
     * @notice Emitted when a privilege is added to a validator.
     * @param operator The operator address of the validator.
     * @param privilegeId The privilege added.
     */
    event PrivilegeAdded(address indexed operator, uint8 indexed privilegeId);

    /**
     * @notice Emitted when a privilege is removed from a validator.
     * @param operator The operator address of the validator.
     * @param privilegeId The privilege removed.
     */
    event PrivilegeRemoved(address indexed operator, uint8 indexed privilegeId);

    /** 
     * @notice Returns `true` after updating the contracts owner
     * @dev Must be called by contract candidateOwner
     */
    function acceptOwnership() external returns (bool);

    /**
     * @notice Returns validator information for a specificed application
     * @param operator The operator address of the target application/validator
     * @return consPubKey The validators consensus public key
     * @return description The validators description (moniker, identity, website, securityContact & details)
     */
    function application(address operator) external view returns (bytes32 consPubKey, Description calldata description);

    /** 
     * @notice Returns list of operator addresses with pending applications
     */ 
    function applications() external view returns (address[] calldata);

    /** 
     * @notice Returns `true` after successfully approving a validator application
     * @param operator The operator address of the target application
     * @dev Must be called by contract owner
     */
    function approveApplication(address operator) external returns (bool);
    
    /** 
     * @notice Returns the address of the current contract candidateOwner
     */
    function candidateOwner() external view returns (address);

    /** 
     * @notice Returns `true` after successfully removing all pending applications
     * @dev Must be called by contract owner
     */
    function cleanupApplications() external returns (bool);

    /** 
     * @notice Returns `true` after successfully removing a validator with operator address from the pool
     * @param operator The operator address of the target validator
     * @dev Must be called by contract owner
     */
    function kick(address operator) external returns (bool);

    /** 
     * @notice Returns `true` after removing a validator with operator address equal to `msg.sender`
     */
    function leave() external returns (bool);

    /** 
     * @notice Returns the address of the current contract owner
     */
    function owner() external view returns (address);

    /** 
     * @notice Returns `true` if validator application is successfully submitted
     * @param consPubKey The validators consensus public key
     * @param description The validators description (moniker, identity, website, securityContact & details)
     */
    function submitApplication(
        bytes32 consPubKey,
        Description calldata description
    ) external returns (bool);

    /** 
     * @notice Returns `true` after updating the contracts candidateOwner
     * @param newOwner The address to transfer ownership to
     * @dev Must be called by contract owner
     */
    function transferOwnership(address newOwner) external returns (bool);

    /** 
     * @notice Returns validator information for a specificed validator
     * @param operator The operator address of the target validator
     * @return consPubKey The validators consensus public key
     * @return description The validators description (moniker, identity, website, securityContact & details)
     */
    function validator(address operator) external view returns (bytes32 consPubKey, Description calldata description);

    /** 
     * @notice Returns list of operator addresses of current validators
     */ 
    function validators() external view returns (address[] calldata);

    /**
     * @notice Adds the given privilege to the specified operators.
     * @param operators List of operator addresses to add the privilege to.
     * @param privilegeId The privilege to add.
     * @dev Must be called by contract owner.
     */
    function addPrivilege(
        address[] calldata operators,
        uint8 privilegeId
    ) external returns (bool);

    /**
     * @notice Removes the given privilege from the specified operators.
     * @param operators List of operator addresses to remove the privilege from.
     * @param privilegeId The privilege to add.
     * @dev Must be called by contract owner.
     */
    function removePrivilege(
        address[] calldata operators,
        uint8 privilegeId
    ) external returns (bool);

    /**
     * @notice Returns a list of validators with the specified privilege.
     * @param privilegeId The privilege to filter by.
     * @return operators List of operator addresses of validators with the specified privilege.
     */
    function validatorsByPrivilege(
        uint8 privilegeId
    ) external view returns (address[] memory operators);

    /**
     * @notice Returns all privileges available for validators.
     * @return List of all privileges available for validators.
     */
    function privileges() external view returns (Privilege[] memory);
}