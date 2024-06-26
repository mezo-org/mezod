<!--
order: 2
-->

# Messages

In this section, we describe the processing of the messages and the corresponding updates to the state.

## MsgSubmitApplication

A new application to become a validator is made using the `MsgSubmitApplication` message.

```go
type MsgSubmitApplication struct {
    Candidate Validator
}
```

This message is expected to fail if:

- another application with the operator address is already registered
- another application with the pubkey is already registered
- another validator with the operator address is already registered
- another validator with the pubkey is already registered
- the validator set is full
- the description fields are too large

This message creates and stores a new `Vote` object in the application pool.

## MsgProposeKick

A new kick proposal to kick a validator is made using the `MsgProposeKick` message.

```go
type MsgProposeKick struct {
    ValidatorAddr   sdk.ValAddress
    ProposerAddr    sdfk.ValAddres
}
```

This message is expected to fail if:

- the proposer address is not in the validator set
- the validator address is not in the validator set
- the validator address is already in the kick proposal pool

This message creates and stores a new `Vote` object in the kick proposal pool.

## MsgVote

MsgVote is a single message that can be used to:
- Approve the application of a candidate to become a validator
- Reject the application of a candidate to become a validator
- Approve a kick proposal to remove a validator
- Reject a kick proposal to remove a validator

```go
type MsgVote struct {
	VoteType      uint16         `json:"type"`
	VoterAddr     sdk.ValAddress `json:"voter"`
	CandidateAddr sdk.ValAddress `json:"candidate"`
	Approve       bool           `json:"approve"`
}

const (
	VoteTypeApplication  uint16 = iota
	VoteTypeKickProposal uint16 = iota
)
```

This message is expected to fail if:

- the voter has already voted
- the voter is not a validator
- the candidate address is not in the application pool in case of an application
- the candidate address is not in the kick proposal pool in case of a kick proposal

In case of an application, this message updates the vote status of the application. If the approval quorum is reached, the candidate is appended into the validator set.
In case of an kick proposal, this message updates the vote status of the kick proposal. If the approval quorum is reached, the candidate is removed from the validator set.

## MsgLeaveValidatorSet

A current validator arbitrarily leaves the validator set using the MsgLeaveValidatorSet message.

```go
type MsgLeaveValidatorSet struct {
    ValidatorAddr   sdk.ValAddress
}
```

This message is expected to fail if:

- the validator is the only validator of the set
- the validator address is not in the validator set

The message removes the validator from the validator set.
