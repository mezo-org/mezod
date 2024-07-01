package keeper

import (
	"testing"

	"github.com/evmos/evmos/v12/x/poa"
	"github.com/evmos/evmos/v12/x/poa/types"
	"github.com/google/go-cmp/cmp"
)

func TestProposeKick(t *testing.T) {
	// Test with maxValidator=15, quorum=66
	ctx, poaKeeper := poa.MockContext()
	validator1, _ := poa.MockValidator()
	validator2, _ := poa.MockValidator()
	nothing, _ := poa.MockValidator()
	poaKeeper.SetParams(ctx, poaKeeper.authority, types.DefaultParams())

	// Add validators to validator set
	poaKeeper.appendValidator(ctx, validator1)
	poaKeeper.appendValidator(ctx, validator2)

	// Cannot propose to kick oneself
	err := poaKeeper.ProposeKick(ctx, validator1.GetOperator(), validator1.GetOperator())
	if err.Error() != types.ErrProposerIsCandidate.Error() {
		t.Errorf("ProposeKick with same address, error should be %v, got %v", types.ErrProposerIsCandidate.Error(), err.Error())
	}

	// The kick proposal is created correctly
	err = poaKeeper.ProposeKick(ctx, validator1.GetOperator(), validator2.GetOperator())
	if err != nil {
		t.Errorf("ProposeKick should create a kick proposal, got error %v", err)
	}
	_, found := poaKeeper.GetKickProposal(ctx, validator1.GetOperator())
	if !found {
		t.Errorf("ProposeKick should create a kick proposal, the kick proposal has not been found")
	}

	// A new application with the same validator cannot be created
	err = poaKeeper.ProposeKick(ctx, validator1.GetOperator(), validator2.GetOperator())
	if err.Error() != types.ErrAlreadyInKickProposal.Error() {
		t.Errorf("ProposeKick with duplicate, error should be %v, got %v", types.ErrAlreadyInKickProposal.Error(), err.Error())
	}

	// A non validator cannot create a kick proposal
	err = poaKeeper.ProposeKick(ctx, validator2.GetOperator(), nothing.GetOperator())
	if err.Error() != types.ErrProposerNotValidator.Error() {
		t.Errorf("ProposeKick sent by a non validator, error should be %v, got %v", types.ErrProposerNotValidator.Error(), err.Error())
	}

	// A non validator cannot be proposed to be kicked
	err = poaKeeper.ProposeKick(ctx, nothing.GetOperator(), validator2.GetOperator())
	if err.Error() != types.ErrNotValidator.Error() {
		t.Errorf("ProposeKick propose a non validator, error should be %v, got %v", types.ErrNotValidator.Error(), err.Error())
	}

	// Test with quorum=0
	ctx, poaKeeper = poa.MockContext()
	validator1, _ = poa.MockValidator()
	validator2, _ = poa.MockValidator()
	poaKeeper.SetParams(ctx, poaKeeper.authority, types.NewParams(15, 0))

	// Add validators to validator set
	poaKeeper.appendValidator(ctx, validator1)
	poaKeeper.appendValidator(ctx, validator2)

	// The validator should be directly appended if the quorum is 0
	err = poaKeeper.ProposeKick(ctx, validator1.GetOperator(), validator2.GetOperator())
	if err != nil {
		t.Errorf("ProposeKick with quorum 0 should kick validator, got error %v", err)
	}
	// Check state is leaving
	foundState, found := poaKeeper.GetValidatorState(ctx, validator1.GetOperator())
	if !found {
		t.Errorf("ProposeKick with quorum 0 should not directly remove the validator from the validator set")
	}
	if foundState != types.ValidatorStateLeaving {
		t.Errorf("ProposeKick with quorum 0, the validator state should be leaving")
	}
}

func TestVoteKickProposal(t *testing.T) {
	ctx, poaKeeper := poa.MockContext()
	voter1, _ := poa.MockValidator()
	voter2, _ := poa.MockValidator()
	validator1, _ := poa.MockValidator()
	poaKeeper.SetParams(ctx, poaKeeper.authority, types.NewParams(15, 100)) // Set quorum to 100%

	// Add voter to validator set
	poaKeeper.appendValidator(ctx, voter1)
	poaKeeper.appendValidator(ctx, voter2)
	poaKeeper.appendValidator(ctx, validator1)

	// Cannot vote if no kick proposal
	err := poaKeeper.VoteKickProposal(ctx, voter1.GetOperator(), validator1.GetOperator(), true)
	if err.Error() != types.ErrNoKickProposalFound.Error() {
		t.Errorf("VoteKickProposal with no kick proposal, error should be %v, got %v", types.ErrNoKickProposalFound.Error(), err.Error())
	}

	// Create a kick proposal
	poaKeeper.appendKickProposal(ctx, validator1)

	// Cannot vote to kick oneself
	err = poaKeeper.VoteKickProposal(ctx, validator1.GetOperator(), validator1.GetOperator(), true)
	if err.Error() != types.ErrVoterIsCandidate.Error() {
		t.Errorf("VoteKickProposal with same address, error should be %v, got %v", types.ErrVoterIsCandidate.Error(), err.Error())
	}

	// Can vote a kick proposal
	err = poaKeeper.VoteKickProposal(ctx, voter1.GetOperator(), validator1.GetOperator(), true)
	if err != nil {
		t.Errorf("VoteKickProposal should vote on a kick proposal, got error %v", err)
	}
	_, found := poaKeeper.GetValidator(ctx, validator1.GetOperator())
	if !found {
		t.Errorf("VoteKickProposal with 1/2 approve should not remove the validator")
	}
	kickProposal, found := poaKeeper.GetKickProposal(ctx, validator1.GetOperator())
	if !found {
		t.Errorf("VoteKickProposal with 1/2 approve should not remove the kick proposal")
	}
	if kickProposal.GetTotal() != 1 {
		t.Errorf("VoteKickProposal with approve should add one vote to the kick proposal")
	}
	if kickProposal.GetApprovals() != 1 {
		t.Errorf("VoteKickProposal with approve should add one approve to the kick proposal")
	}

	// Second approve should set the set of the validator to leaving
	err = poaKeeper.VoteKickProposal(ctx, voter2.GetOperator(), validator1.GetOperator(), true)
	if err != nil {
		t.Errorf("VoteKickProposal 2 should vote on a kick proposal, got error %v", err)
	}
	_, found = poaKeeper.GetKickProposal(ctx, validator1.GetOperator())
	if found {
		t.Errorf("VoteKickProposal with 2/2 approve should remove the kick proposal")
	}
	validatorState, found := poaKeeper.GetValidatorState(ctx, validator1.GetOperator())
	if !found {
		t.Errorf("VoteKickProposal with 2/2 approve should not directly remove the validator from the validator set")
	}
	if validatorState != types.ValidatorStateLeaving {
		t.Errorf("VoteKickProposal with 2/2 approve should set the state of the validator to leaving")
	}

	// Quorum 100%: one reject is sufficient to reject the kick proposal
	poaKeeper.appendKickProposal(ctx, voter2)
	err = poaKeeper.VoteKickProposal(ctx, voter1.GetOperator(), voter2.GetOperator(), false)
	if err != nil {
		t.Errorf("VoteKickProposal 3 should vote on a kick proposal, got error %v", err)
	}
	_, found = poaKeeper.GetKickProposal(ctx, voter2.GetOperator())
	if found {
		t.Errorf("VoteKickProposal with 1 reject should reject the kick proposal")
	}
	validatorState, found = poaKeeper.GetValidatorState(ctx, voter2.GetOperator())
	if !found {
		t.Errorf("VoteKickProposal kick proposal rejected should not remove the validator")
	}
	if validatorState == types.ValidatorStateLeaving {
		t.Errorf("VoteKickProposal kick proposal rejected should not set the state of the validator to leaving")
	}

	// Reapply and set quorum to 1%
	poaKeeper.appendKickProposal(ctx, voter2)
	poaKeeper.SetParams(ctx, poaKeeper.authority, types.NewParams(15, 1))

	// One reject should update the vote but not reject totally the kick proposal
	err = poaKeeper.VoteKickProposal(ctx, voter1.GetOperator(), voter2.GetOperator(), false)
	if err != nil {
		t.Errorf("VoteKickProposal 4 should vote on a kick proposal, got error %v", err)
	}
	kickProposal, found = poaKeeper.GetKickProposal(ctx, voter2.GetOperator())
	if !found {
		t.Errorf("VoteKickProposal with 1/2 reject should not remove the kick proposal")
	}
	_, found = poaKeeper.GetValidatorState(ctx, voter2.GetOperator())
	if !found {
		t.Errorf("VoteKickProposal with 1/2 rejec should not remove the validator")
	}
	if validatorState == types.ValidatorStateLeaving {
		t.Errorf("VoteKickProposal with 1/2 rejec should not set the state of the validator to leaving")
	}
	if kickProposal.GetTotal() != 1 {
		t.Errorf("VoteKickProposal with 1/2 reject should add one vote to the kick proposal")
	}
	if kickProposal.GetApprovals() != 0 {
		t.Errorf("VoteKickProposal with 1/2 reject should not add one approve to the kick proposal")
	}
}

func TestGetKickProposal(t *testing.T) {
	ctx, poaKeeper := poa.MockContext()
	validator1, _ := poa.MockValidator()
	validator2, _ := poa.MockValidator()
	kickProposal := types.NewVote(validator1)

	poaKeeper.setKickProposal(ctx, kickProposal)

	// Should find the correct kick proposal
	retrievedKickProposal, found := poaKeeper.GetKickProposal(ctx, validator1.GetOperator())
	if !found {
		t.Errorf("GetKickProposal should find kick proposal if it has been set")
	}

	if !cmp.Equal(kickProposal.GetSubject(), retrievedKickProposal.GetSubject()) {
		t.Errorf("GetKickProposal should find %v, found %v", kickProposal.GetSubject(), retrievedKickProposal.GetSubject())
	}
	if kickProposal.GetTotal() != retrievedKickProposal.GetTotal() {
		t.Errorf("GetKickProposal should find %v votes, found %v", kickProposal.GetTotal(), retrievedKickProposal.GetTotal())
	}
	if kickProposal.GetApprovals() != retrievedKickProposal.GetApprovals() {
		t.Errorf("GetKickProposal should find %v approvals, found %v", kickProposal.GetApprovals(), retrievedKickProposal.GetApprovals())
	}

	// Should not find a unset kick proposal
	_, found = poaKeeper.GetKickProposal(ctx, validator2.GetOperator())
	if found {
		t.Errorf("GetKickProposal should not find kick proposal if it has not been set")
	}
}

func TestAppendKickProposal(t *testing.T) {
	ctx, poaKeeper := poa.MockContext()
	validator, _ := poa.MockValidator()

	poaKeeper.appendKickProposal(ctx, validator)

	_, foundKickProposal := poaKeeper.GetKickProposal(ctx, validator.GetOperator())

	if !foundKickProposal {
		t.Errorf("AppendKickProposal should append the kick proposal. Found val: %v", foundKickProposal)
	}
}

func TestRemoveKickProposal(t *testing.T) {
	ctx, poaKeeper := poa.MockContext()
	validator, _ := poa.MockValidator()

	// Append and remove kick proposal
	poaKeeper.appendKickProposal(ctx, validator)
	poaKeeper.removeKickProposal(ctx, validator.GetOperator())

	// Should not find a removed validator
	_, found := poaKeeper.GetKickProposal(ctx, validator.GetOperator())

	if found {
		t.Errorf("RemoveKickProposal should remove kick proposal record")
	}
}

func TestGetAllKickProposals(t *testing.T) {
	ctx, poaKeeper := poa.MockContext()
	validator1, _ := poa.MockValidator()
	validator2, _ := poa.MockValidator()
	kickProposal1 := types.NewVote(validator1)
	kickProposal2 := types.NewVote(validator2)

	poaKeeper.setKickProposal(ctx, kickProposal1)
	poaKeeper.setKickProposal(ctx, kickProposal2)

	retrievedKickProposals := poaKeeper.GetAllKickProposals(ctx)
	if len(retrievedKickProposals) != 2 {
		t.Errorf("GetAllKickProposals should find %v kick proposal, found %v", 2, len(retrievedKickProposals))
	}
}
