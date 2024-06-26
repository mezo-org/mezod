package types_test

import (
	"testing"

	"github.com/evmos/evmos/v12/x/poa"
	"github.com/evmos/evmos/v12/x/poa/types"
)

func TestAddVote(t *testing.T) {
	validator, _ := poa.MockValidator()
	account1 := poa.MockValAddress()
	account2 := poa.MockValAddress()
	vote := types.NewVote(validator)

	if vote.GetTotal() != 0 {
		t.Errorf("Vote should contain no vote when created")
	}

	alreadyVoted := vote.AddVote(account1, true)
	if alreadyVoted != false {
		t.Errorf("AddVote should return false if the voter hasn't voted yet")
	}
	if vote.GetTotal() != 1 {
		t.Errorf("AddVote should increase the number of votes in the vote")
	}
	if vote.GetApprovals() != 1 {
		t.Errorf("AddVote with approval should increase the number of approvals in the vote")
	}

	alreadyVoted = vote.AddVote(account2, false)
	if alreadyVoted != false {
		t.Errorf("AddVote should return false if the voter hasn't voted yet")
	}
	if vote.GetTotal() != 2 {
		t.Errorf("AddVote should increase the number of votes in the vote")
	}
	if vote.GetApprovals() != 1 {
		t.Errorf("AddVote with reject should not increase the number of approvals in the vote")
	}

	alreadyVoted = vote.AddVote(account1, true)
	if alreadyVoted != true {
		t.Errorf("AddVote should return true if the voter has already voted")
	}
	if vote.GetTotal() != 2 {
		t.Errorf("AddVote should not increase the number of votes if the voter has already voted")
	}

	alreadyVoted = vote.AddVote(account2, true)
	if alreadyVoted != true {
		t.Errorf("AddVote should return true if the voter has already voted")
	}
	if vote.GetTotal() != 2 {
		t.Errorf("AddVote should not increase the number of votes if the voter has already voted")
	}
}

func TestCheckQuorum(t *testing.T) {
	validator, _ := poa.MockValidator()
	account1 := poa.MockValAddress()
	account2 := poa.MockValAddress()
	account3 := poa.MockValAddress()
	account4 := poa.MockValAddress()
	account5 := poa.MockValAddress()
	vote1 := types.NewVote(validator)
	vote2 := types.NewVote(validator)
	vote3 := types.NewVote(validator)

	// Quorum should be a percentage
	_, _, err := vote1.CheckQuorum(100, 101)
	if err == nil {
		t.Errorf("CheckQuorum should return an error if quorum is not a percentage")
	}

	// Should always be approved if the quorum is 0
	reached, approved, err := vote1.CheckQuorum(100, 0)
	if reached == false || approved == false || err != nil {
		t.Errorf("CheckQuorum should return approval if quorum is 0, %v, %v, %v", reached, approved, err)
	}

	// Quorum of 100 means all of the voters must approve the vote
	reached, approved, err = vote1.CheckQuorum(2, 100)
	if reached == true || approved == true || err != nil {
		t.Errorf("100 percents: Quorum should not be reached with 0/2 vote, %v, %v, %v", reached, approved, err)
	}
	vote1.AddVote(account1, true)
	reached, approved, err = vote1.CheckQuorum(2, 100)
	if reached == true || approved == true || err != nil {
		t.Errorf("100 percents: Quorum should not be reached with 1/2 votes, %v, %v, %v", reached, approved, err)
	}
	vote1.AddVote(account2, true)
	reached, approved, err = vote1.CheckQuorum(2, 100)
	if reached == false || approved == false || err != nil {
		t.Errorf("100 percents: Quorum should be reached with 2/2 votes, %v, %v, %v", reached, approved, err)
	}

	// Quorum of 50 means more than half of the voters must approve the vote
	vote2.AddVote(account1, true)
	vote2.AddVote(account2, true)
	reached, approved, err = vote2.CheckQuorum(5, 50)
	if reached == true || approved == true || err != nil {
		t.Errorf("50 percents: Quorum should not be reached with 2/5 votes, %v, %v, %v", reached, approved, err)
	}
	vote2.AddVote(account3, false)
	vote2.AddVote(account4, false)
	reached, approved, err = vote2.CheckQuorum(5, 50)
	if reached == true || approved == true || err != nil {
		t.Errorf("50 percents: Quorum should not be reached with 2/5 approvals, %v, %v, %v", reached, approved, err)
	}
	vote2.AddVote(account5, true)
	reached, approved, err = vote2.CheckQuorum(5, 50)
	if reached == false || approved == false || err != nil {
		t.Errorf("50 percents: Quorum should be reached with 3/5 approvals, %v, %v, %v", reached, approved, err)
	}

	// Quorum is reached and vote rejected if the required number of approval cannot be reached
	vote3.AddVote(account1, false)
	vote3.AddVote(account2, false)
	reached, approved, err = vote3.CheckQuorum(6, 66)
	if reached == true || approved == true || err != nil {
		t.Errorf("Vote3 quorum should not be reached with 2 votes, %v, %v, %v", reached, approved, err)
	}
	// With 3 rejections, the approval cannot be reached anymore
	vote3.AddVote(account3, false)
	reached, approved, err = vote3.CheckQuorum(6, 66)
	if reached == false || approved == true || err != nil {
		t.Errorf("Vote3 should have a reached quorum but not approved (rejected), %v, %v, %v", reached, approved, err)
	}
}
