package types

import (
	"math"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func NewVote(subject Validator) Vote {
	return Vote{
		Subject:   subject,
		Approvals: 0,
		Total:     0,
		Voters:    [][]byte{},
	}
}

// Add a vote
func (v *Vote) AddVote(voter sdk.ValAddress, approve bool) (alreadyVoted bool) {
	// Verify if the voter already voted
	for _, currentVoter := range v.Voters {
		if voter.Equals(sdk.ValAddress(currentVoter)) {
			// The voter already voted
			return true
		}
	}

	// Append the voter in the voters list
	v.Voters = append(v.Voters, voter)

	// Update vote status
	v.Total += 1
	if approve {
		v.Approvals += 1
	}

	return false
}

// Check if the quorum has been reached
// voterPoolSize is the total number of possible voters in the vote
// Quorum is the percentage of voters to reach to approve or reject the vote
// Reached true -> the quorum has been reached
// Approved true -> the vote has been approved, otherwise it has been rejected
func (v Vote) CheckQuorum(voterPoolSize uint64, quorum uint64) (reached bool, approved bool, err error) {
	// Check parameters
	if quorum > 100 {
		return false, false, ErrInvalidQuorumValue
	}
	if voterPoolSize < v.Total {
		return false, false, ErrInvalidVoterPoolSize
	}

	// Get the necessary number of approval to approve the vote
	necessaryApproval := uint64(math.Ceil(float64(voterPoolSize*quorum) / 100.0))

	// Check if the vote is approved
	if v.Approvals >= necessaryApproval {
		return true, true, nil
	}

	// Get the number of remaining voters in the pool
	remainingVoters := voterPoolSize - v.Total

	// Check if the vote can still be approved in the future
	if (v.Approvals + remainingVoters) >= necessaryApproval {
		// The vote can still be approved, therefore the quorum has not been reached
		return false, false, nil
	} else {
		// The vote can't be approved anymore, therefore the quorum has been reached to reject the proposition
		return true, false, nil
	}
}

// Vote encoding functions
func MustMarshalVote(cdc codec.BinaryCodec, v Vote) []byte {
	return cdc.MustMarshal(&v)
}

func MustUnmarshalVote(cdc codec.BinaryCodec, value []byte) Vote {
	vote, err := UnmarshalVote(cdc, value)
	if err != nil {
		panic(err)
	}

	return vote
}

func UnmarshalVote(cdc codec.BinaryCodec, value []byte) (v Vote, err error) {
	err = cdc.Unmarshal(value, &v)
	return v, err
}
