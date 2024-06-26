<!--
order: 1
-->

# State

## Params

Params is a module-wide configuration structure that stores system parameters
and defines the overall functioning of the staking module.

- Params: `Paramsspace("poa") -> amino(params)`

```go
type Params struct {
    MaxValidators   uint16         // Maximum number of validators
    Quorum          uint16	   // The percentage of validator approvals to reach to vote a decision (new validator or kick)
}
```

## Validator

Validators objects should be primarily stored and accessed by the
`OperatorAddr`, an SDK validator address for the operator of the validator. `ValidatorByConsAddr` is maintained per validator object in order to fulfill
the required lookups for slashing and validator-set updates.

- Validators: `0x21 | OperatorAddr -> amino(validator)`
- ValidatorsByConsAddr: `0x22 | ConsAddr -> OperatorAddr`
- ValidatorStates: `0x23 | OperatorAddr -> ValidatorState`

`Validators` is the primary index - it ensures that each operator can have only one
associated validator, where the public key of that validator can change in the
future.
`ValidatorByConsAddr` is an additional index that enables lookups for future uses (like automatic kick for misbehaving).
`ValidatorStates` holds the state of a validator. The validator can have 3 states: joining, joined or leaving. This state allows the End Blocker to know how to update the Tendermint Core validator state.

Each validator's state is stored in a `Validator` struct:

```go
type Validator struct {
    OperatorAddress         sdk.ValAddress  // address of the validator's operator; bech encoded in JSON
    ConsPubKey              crypto.PubKey   // the consensus public key of the validator; bech encoded in JSON
    Description             Description     // description terms for the validator
}

type Description struct {
    Moniker          string // name
    Identity         string // optional identity signature (ex. UPort or Keybase)
    Website          string // optional website link
    SecurityContact  string // optional email for security contact
    Details          string // optional details
}

const (
	ValidatorStateJoining uint16 = iota // The validator is joining the validator set, it is not yet present in Tendermint validator set
	ValidatorStateJoined  uint16 = iota // The validator is already present in Tendermind validator set
	ValidatorStateLeaving uint16 = iota // The validator is leaving the validator set, it will leave Tendermint validator set at the end of the block
)
```

## Application

The application pool tracks all the current applications. An operator can only be one validator, therefore the application can be accessed by the operator address.

- ApplicationPool: `0x24 | OperatorAddr -> amino(vote)`
- CandidateByConsAddr: `0x25 | ConsAddr -> OperatorAddr`

An application is stored in a `Vote` structure to track the current state of the vote like the current number of approvals. The subject field represents the potential new validator.

`CandidateByConsAddr` is an additional index to ensure there is no two applications with the same consensus public key

```go
type Vote struct {
	Subject   Validator        // The information of the potential new validator
	Approvals uint64           // The current number of approvals of the application
	Total     uint64           // The current number of total vote (approval+rejection)
	Voters    []sdk.AccAddress // The identity of validators who voted so far
}
```

## KickProposal

The kick proposal pool tracks all the current propositions to kick a validator. An operator can only be one validator, therefore the application can be accessed by the operator address.

- KickProposalPool: `0x26 | OperatorAddr -> amino(vote)`

An application is stored in a `Vote` structure to track the current state of the vote like the current number of approvals. The subject field represents the validator to be eventually kicked.
