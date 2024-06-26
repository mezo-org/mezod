# PoA

Proof-of-Authority module for Cosmos SDK

This module implements a simple Proof of Authority system to determine the Tendermint validator set of the Cosmos SDK application through internal voting.

This simple module can be used in a Cosmos SDK application without the dependency of other modules.

An initial validator set is defined in the genesis file. Then, validators can be appended or kicked from the validator set through voting from the current validators. A quorum percentage is defined in the parameters of the module. This quorum defines the number of approvals required to vote decision. For example: if the quorum is 50% and the current validator set contains 10 validators. 5 validator approvals are required to accept a new candidate in the validator set. All validators in the system have equal voting power.

## Queries

The following queries are available to consult state of the validator set:

- `validator`      Query a validator
- `validators`     Query all validators
- `params`         Query the params
- `applications`   Query the applications to become validator
- `kick-proposals` Query the kick proposals to remove validator

They can be called with the command `<cli> query poa <query>`

## Transactions

The following transactions are available to interact with the validator set

- `apply`               Apply to become a new validator in the network
- `propose-kick`        Propose to kick a validator from the validator
- `vote-application`    Approve or reject the application to become a validator
- `vote-kick-proposal`  Approve or reject a kick proposal to remove a validator
- `leave-validator-set` Instantly leave the validator set

They can be called with the command `<cli> tx poa <tx>`

## Technical specifications

The specifications of this module can be found [here](./spec/README.md)

## Demonstration Example

An example of the interaction with the validator set can be found here with the SupplyChainX application https://github.com/ltacker/supplychainx/blob/master/DEMO.md