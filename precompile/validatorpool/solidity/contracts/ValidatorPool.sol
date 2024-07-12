// SPDX-License-Identifier: UNLICENSED
pragma solidity ^0.8.24;

// https://docs.openzeppelin.com/contracts/5.x/api/access#Ownable
import "@openzeppelin/contracts/access/Ownable.sol";
// https://docs.openzeppelin.com/contracts/5.x/api/access#Ownable2Step
import "@openzeppelin/contracts/access/Ownable2Step.sol";

contract ValidatorPool is Ownable2Step {
    // validator application states    
    enum State { none, pending, approved }
    // validator struct
    struct Validator {
        address consPubKey;
        State state;
    }
    mapping(address operator => Validator validator) public validators;
    uint public slots; // open validator slots
    
    constructor(address initialOwner, uint maxValidators) Ownable(initialOwner) {
        slots = maxValidators;
    }

    // events 
    event ApplicationSubmitted(address operator, address consensusPubKey);
    event ApplicationApproved(address candidate);
    event ValidatorJoined(address validator);
    event ValidatorKicked(address validator);
    event ValidatorLeft(address validator);

    // errors
    error InvalidOperator(address operator, string message);
    error ApplicationExists(address operator, string message);
    error NoValidatorSlots(string message);
    error InvalidState(address operator, string message);

    // Apply to become a validator
    // 1. A candidate issues a submitApplication EVM transaction to the ValidatorPool precompiled contract. The candidate passes their consensus public key and operator address as input. There is no way to submit applications on behalf of someone else - the candidate must do it directly.
    // 2. The ValidatorPool precompile puts the application in the pending applications queue and emits a ApplicationSubmitted event.
    function submitApplication(address consensusPubKey, address operator) public {
        // There is no way to submit applications on behalf of someone else - the candidate must do it directly.
        if (msg.sender != operator) {
            revert InvalidOperator(operator, "Operator address does not match sender");
        }
        // Application is automatically rejected if candidate is already a validator
        // Application is automatically rejected if candidate already submitted another application
        if (validators[operator].consPubKey != address(0)) {    
            revert ApplicationExists(operator, "Operator has an existing application");
        }
        // Application is automatically rejected if the validator pool max cap is reached
        if (slots == 0) {
            revert NoValidatorSlots("Validator pool is full");
        }

        // add to validators with pending state     
        validators[operator].consPubKey = consensusPubKey;
        validators[operator].state = State.pending;

        emit ApplicationSubmitted(operator, consensusPubKey);
    }

    // Approve validator application
    // 1. The owner reviews pending application of a candidate and decides about its fate.
    // 2. The owner issues an approveApplication EVM transaction to the ValidatorPool precompiled contract to approve a pending application. No explicit action is needed for a rejected application - the approver can just ignore it.
    // 3. The ValidatorPool promotes the candidate to a validator and gives it a voting power equal to all other validators existing in the network. The ApplicationApproved and ValidatorJoined events are emitted as result.
    function approveApplication(address operator) public onlyOwner {
        // Candidate has a pending application
        if (validators[operator].state != State.pending) {
            revert InvalidState(operator, "Operator does not have a pending application");
        }
        // Make sure there are remaining validator slots
        if (slots == 0) {
            revert NoValidatorSlots("Validator pool is full");
        }
        // Set validator state to approved
        validators[operator].state = State.approved;
        slots--;
        
        emit ApplicationApproved(operator);
        emit ValidatorJoined(operator);
    }

    // Kick from validator pool
    // 1. The owner issues a kick EVM transaction to the ValidatorPool precompiled contract in order to kick a malfunctioning validator from the validator pool.
    // 2. The ValidatorPool removes the validator from the active validators pool. Voting power for remaining validators is adjusted to maintain the even distribution invariant. A ValidatorKicked event is emitted as result.
    function kick(address operator) public onlyOwner {
        // Validator must actually be a validator
        if (validators[operator].state != State.approved) {
            revert InvalidState(operator, "Operator is not an approved validator");
        }

        // reset operators state
        validators[operator].consPubKey = address(0);
        validators[operator].state = State.none;
        slots++;
        emit ValidatorKicked(operator);
    }

    // Leave validator pool
    // 1. A validator issues a leave EVM transaction to the ValidatorPool precompiled contract in order to voluntarily leave the validator pool.
    // 2. The ValidatorPool removes the validator from the active validators pool. Voting power for remaining validators is adjusted to maintain the even distribution invariant. A ValidatorLeft event is emitted as result.
    function leave() public {
        // Validator must actually be a validator
        if (validators[msg.sender].state != State.approved) {
            revert InvalidState(msg.sender, "Operator is not an approved validator");
        }
        validators[msg.sender].consPubKey = address(0);
        validators[msg.sender].state = State.none;
        slots++;
        emit ValidatorLeft(msg.sender);
    }
}