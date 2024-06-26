package poa_test

import (
	"testing"

	"github.com/evmos/evmos/v12/x/poa"
	"github.com/evmos/evmos/v12/x/poa/types"
)

func TestHandleMsgSubmitApplication(t *testing.T) {
	// Test with maxValidator=15, quorum=66
	ctx, poaKeeper := poa.MockContext()
	handler := poa.NewHandler(poaKeeper)
	validator, _ := poa.MockValidator()
	poaKeeper.SetParams(ctx, types.DefaultParams())

	// The application is submitted correctly
	msg := types.NewMsgSubmitApplication(validator)
	_, err := handler(ctx, msg)
	if err != nil {
		t.Errorf("MsgSubmitApplication should submit an application, got error %v", err)
	}
	_, found := poaKeeper.GetApplication(ctx, validator.GetOperator())
	if !found {
		t.Errorf("MsgSubmitApplication should submit an application, the application has not been found")
	}
	_, found = poaKeeper.GetApplicationByConsAddr(ctx, validator.GetConsAddr())
	if !found {
		t.Errorf("MsgSubmitApplication should submit an application, the application has not been found by cons addr")
	}

	// A new application with the same validator cannot be created
	_, err = handler(ctx, msg)
	if err.Error() != types.ErrAlreadyApplying.Error() {
		t.Errorf("MsgSubmitApplication with duplicate, error should be %v, got %v", types.ErrAlreadyApplying.Error(), err.Error())
	}

	// Test with quorum=0
	ctx, poaKeeper = poa.MockContext()
	handler = poa.NewHandler(poaKeeper)
	validator, _ = poa.MockValidator()
	poaKeeper.SetParams(ctx, types.NewParams(15, 0))

	// The validator should be directly appended if the quorum is 0
	msg = types.NewMsgSubmitApplication(validator)
	_, err = handler(ctx, msg)
	if err != nil {
		t.Errorf("MsgSubmitApplication with quorum 0 should append validator, got error %v", err)
	}
	_, found = poaKeeper.GetValidator(ctx, validator.GetOperator())
	if !found {
		t.Errorf("MsgSubmitApplication with quorum 0 should append validator, the validator has not been found")
	}
	_, found = poaKeeper.GetValidatorByConsAddr(ctx, validator.GetConsAddr())
	if !found {
		t.Errorf("MsgSubmitApplication with quorum 0 should append validator, the validator has not been found by cons addr")
	}
	foundState, found := poaKeeper.GetValidatorState(ctx, validator.GetOperator())
	if !found {
		t.Errorf("MsgSubmitApplication with quorum 0 should append validator, the validator state has not been found")
	}
	if foundState != types.ValidatorStateJoining {
		t.Errorf("MsgSubmitApplication with quorum 0, the validator should have the state joining, if it is appended")
	}

	// A new application cannot be created if the validator already exist
	_, err = handler(ctx, msg)
	if err.Error() != types.ErrAlreadyValidator.Error() {
		t.Errorf("MsgSubmitApplication with duplicate, error should be %v, got %v", types.ErrAlreadyValidator.Error(), err.Error())
	}

	// Test max validators condition
	poaKeeper.SetParams(ctx, types.NewParams(1, 0))
	_, err = handler(ctx, msg)
	if err.Error() != types.ErrMaxValidatorsReached.Error() {
		t.Errorf("MsgSubmitApplication with max validators reached, error should be %v, got %v", types.ErrMaxValidatorsReached.Error(), err.Error())
	}
}

func TestHandleMsgProposeKick(t *testing.T) {
	// Test with maxValidator=15, quorum=66
	ctx, poaKeeper := poa.MockContext()
	handler := poa.NewHandler(poaKeeper)
	validator1, _ := poa.MockValidator()
	validator2, _ := poa.MockValidator()
	nothing, _ := poa.MockValidator()
	poaKeeper.SetParams(ctx, types.DefaultParams())

	// Add validators to validator set
	poaKeeper.AppendValidator(ctx, validator1)
	poaKeeper.AppendValidator(ctx, validator2)

	// Cannot propose to kick oneself
	msg := types.NewMsgProposeKick(validator1.GetOperator(), validator1.GetOperator())
	_, err := handler(ctx, msg)
	if err.Error() != types.ErrProposerIsCandidate.Error() {
		t.Errorf("MsgProposeKick with same address, error should be %v, got %v", types.ErrProposerIsCandidate.Error(), err.Error())
	}

	// The kick proposal is created correctly
	msg = types.NewMsgProposeKick(validator1.GetOperator(), validator2.GetOperator())
	_, err = handler(ctx, msg)
	if err != nil {
		t.Errorf("MsgProposeKick should create a kick proposal, got error %v", err)
	}
	_, found := poaKeeper.GetKickProposal(ctx, validator1.GetOperator())
	if !found {
		t.Errorf("MsgProposeKick should create a kick proposal, the kick proposal has not been found")
	}

	// A new application with the same validator cannot be created
	_, err = handler(ctx, msg)
	if err.Error() != types.ErrAlreadyInKickProposal.Error() {
		t.Errorf("MsgProposeKick with duplicate, error should be %v, got %v", types.ErrAlreadyInKickProposal.Error(), err.Error())
	}

	// A non validator cannot create a kick proposal
	msg = types.NewMsgProposeKick(validator2.GetOperator(), nothing.GetOperator())
	_, err = handler(ctx, msg)
	if err.Error() != types.ErrProposerNotValidator.Error() {
		t.Errorf("MsgProposeKick sent by a non validator, error should be %v, got %v", types.ErrProposerNotValidator.Error(), err.Error())
	}

	// A non validator cannot be proposed to be kicked
	msg = types.NewMsgProposeKick(nothing.GetOperator(), validator2.GetOperator())
	_, err = handler(ctx, msg)
	if err.Error() != types.ErrNotValidator.Error() {
		t.Errorf("MsgProposeKick propose a non validator, error should be %v, got %v", types.ErrNotValidator.Error(), err.Error())
	}

	// Test with quorum=0
	ctx, poaKeeper = poa.MockContext()
	handler = poa.NewHandler(poaKeeper)
	validator1, _ = poa.MockValidator()
	validator2, _ = poa.MockValidator()
	poaKeeper.SetParams(ctx, types.NewParams(15, 0))

	// Add validators to validator set
	poaKeeper.AppendValidator(ctx, validator1)
	poaKeeper.AppendValidator(ctx, validator2)

	// The validator should be directly appended if the quorum is 0
	msg = types.NewMsgProposeKick(validator1.GetOperator(), validator2.GetOperator())
	_, err = handler(ctx, msg)
	if err != nil {
		t.Errorf("MsgProposeKick with quorum 0 should kick validator, got error %v", err)
	}
	// Check state is leaving
	foundState, found := poaKeeper.GetValidatorState(ctx, validator1.GetOperator())
	if !found {
		t.Errorf("MsgProposeKick with quorum 0 should not directly remove the validator from the validator set")
	}
	if foundState != types.ValidatorStateLeaving {
		t.Errorf("MsgProposeKick with quorum 0, the validator state should be leaving")
	}
}

func TestHandleMsgVoteApplication(t *testing.T) {
	ctx, poaKeeper := poa.MockContext()
	handler := poa.NewHandler(poaKeeper)
	voter1, _ := poa.MockValidator()
	voter2, _ := poa.MockValidator()
	candidate1, _ := poa.MockValidator()
	candidate2, _ := poa.MockValidator()
	nothing, _ := poa.MockValidator()
	poaKeeper.SetParams(ctx, types.NewParams(15, 100)) // Set quorum to 100%

	// Add voter to validator set
	poaKeeper.AppendValidator(ctx, voter1)
	poaKeeper.AppendValidator(ctx, voter2)

	// Add candidate to application pool
	poaKeeper.AppendApplication(ctx, candidate1)
	poaKeeper.AppendApplication(ctx, candidate2)

	// Cannot vote if candidate is not in application pool
	msg := types.NewMsgVote(types.VoteTypeApplication, voter1.GetOperator(), nothing.GetOperator(), true)
	_, err := handler(ctx, msg)
	if err.Error() != types.ErrNoApplicationFound.Error() {
		t.Errorf("MsgVoteApplication should fail with %v, got %v", types.ErrNoApplicationFound, err)
	}

	// Cannot vote if the voter is not in validator set
	msg = types.NewMsgVote(types.VoteTypeApplication, nothing.GetOperator(), candidate1.GetOperator(), true)
	_, err = handler(ctx, msg)
	if err.Error() != types.ErrVoterNotValidator.Error() {
		t.Errorf("MsgVoteApplication should fail with %v, got %v", types.ErrVoterNotValidator, err)
	}

	// Can vote an application
	msg = types.NewMsgVote(types.VoteTypeApplication, voter1.GetOperator(), candidate1.GetOperator(), true)
	_, err = handler(ctx, msg)
	if err != nil {
		t.Errorf("MsgVoteApplication should vote on an application, got error %v", err)
	}
	application, found := poaKeeper.GetApplication(ctx, candidate1.GetOperator())
	if !found {
		t.Errorf("MsgVoteApplication with 1/2 approve should not remove the application")
	}
	_, found = poaKeeper.GetValidator(ctx, candidate1.GetOperator())
	if found {
		t.Errorf("MsgVoteApplication with 1/2 approve should not append the candidate to the validator set")
	}
	if application.GetTotal() != 1 {
		t.Errorf("MsgVoteApplication with approve should add one vote to the application")
	}
	if application.GetApprovals() != 1 {
		t.Errorf("MsgVoteApplication with approve should add one approve to the application")
	}

	// Second approve should append the candidate to the validator pool
	msg = types.NewMsgVote(types.VoteTypeApplication, voter2.GetOperator(), candidate1.GetOperator(), true)
	_, err = handler(ctx, msg)
	if err != nil {
		t.Errorf("MsgVoteApplication 2 should vote on an application, got error %v", err)
	}
	_, found = poaKeeper.GetApplication(ctx, candidate1.GetOperator())
	if found {
		t.Errorf("MsgVoteApplication with 2/2 approve should remove the application")
	}
	_, found = poaKeeper.GetValidator(ctx, candidate1.GetOperator())
	if !found {
		t.Errorf("MsgVoteApplication with 2/2 approve should append the candidate to the validator set")
	}

	// Quorum 100%: one reject is sufficient to reject the validator application
	msg = types.NewMsgVote(types.VoteTypeApplication, voter1.GetOperator(), candidate2.GetOperator(), false)
	_, err = handler(ctx, msg)
	if err != nil {
		t.Errorf("MsgVoteApplication 3 should vote on an application, got error %v", err)
	}
	_, found = poaKeeper.GetApplication(ctx, candidate2.GetOperator())
	if found {
		t.Errorf("MsgVoteApplication with 1 reject should reject the application")
	}
	_, found = poaKeeper.GetValidator(ctx, candidate2.GetOperator())
	if found {
		t.Errorf("MsgVoteApplication application rejected should not append the candidate to the validator set")
	}

	// Reapply and set quorum to 1%
	poaKeeper.AppendApplication(ctx, candidate2)
	poaKeeper.SetParams(ctx, types.NewParams(15, 1))

	// One reject should update the vote but not reject totally the application
	msg = types.NewMsgVote(types.VoteTypeApplication, voter1.GetOperator(), candidate2.GetOperator(), false)
	_, err = handler(ctx, msg)
	if err != nil {
		t.Errorf("MsgVoteApplication 4 should vote on an application, got error %v", err)
	}
	application, found = poaKeeper.GetApplication(ctx, candidate2.GetOperator())
	if !found {
		t.Errorf("MsgVoteApplication with 1/3 reject should not remove the application")
	}
	_, found = poaKeeper.GetValidator(ctx, candidate2.GetOperator())
	if found {
		t.Errorf("MsgVoteApplication with 1/3 reject should not append the candidate to the validator set")
	}
	if application.GetTotal() != 1 {
		t.Errorf("MsgVoteApplication with reject should add one vote to the application")
	}
	if application.GetApprovals() != 0 {
		t.Errorf("MsgVoteApplication with reject should not add one approve to the application")
	}

	// Cannot vote if validator set is full
	poaKeeper.SetParams(ctx, types.NewParams(3, 1))

	msg = types.NewMsgVote(types.VoteTypeApplication, voter2.GetOperator(), candidate2.GetOperator(), false)
	_, err = handler(ctx, msg)
	if err.Error() != types.ErrMaxValidatorsReached.Error() {
		t.Errorf("MsgVoteApplication should fail with %v, got %v", types.ErrMaxValidatorsReached, err)
	}
}

func TestHandleMsgVoteKickProposal(t *testing.T) {
	ctx, poaKeeper := poa.MockContext()
	handler := poa.NewHandler(poaKeeper)
	voter1, _ := poa.MockValidator()
	voter2, _ := poa.MockValidator()
	validator1, _ := poa.MockValidator()
	poaKeeper.SetParams(ctx, types.NewParams(15, 100)) // Set quorum to 100%

	// Add voter to validator set
	poaKeeper.AppendValidator(ctx, voter1)
	poaKeeper.AppendValidator(ctx, voter2)
	poaKeeper.AppendValidator(ctx, validator1)

	// Cannot vote if no kick proposal
	msg := types.NewMsgVote(types.VoteTypeKickProposal, voter1.GetOperator(), validator1.GetOperator(), true)
	_, err := handler(ctx, msg)
	if err.Error() != types.ErrNoKickProposalFound.Error() {
		t.Errorf("MsgVoteKickProposal with no kick proposal, error should be %v, got %v", types.ErrNoKickProposalFound.Error(), err.Error())
	}

	// Create a kick proposal
	poaKeeper.AppendKickProposal(ctx, validator1)

	// Cannot vote to kick oneself
	msg = types.NewMsgVote(types.VoteTypeKickProposal, validator1.GetOperator(), validator1.GetOperator(), true)
	_, err = handler(ctx, msg)
	if err.Error() != types.ErrVoterIsCandidate.Error() {
		t.Errorf("MsgVoteKickProposal with same address, error should be %v, got %v", types.ErrVoterIsCandidate.Error(), err.Error())
	}

	// Can vote a kick proposal
	msg = types.NewMsgVote(types.VoteTypeKickProposal, voter1.GetOperator(), validator1.GetOperator(), true)
	_, err = handler(ctx, msg)
	if err != nil {
		t.Errorf("MsgVoteKickProposal should vote on a kick proposal, got error %v", err)
	}
	_, found := poaKeeper.GetValidator(ctx, validator1.GetOperator())
	if !found {
		t.Errorf("MsgVoteKickProposal with 1/2 approve should not remove the validator")
	}
	kickProposal, found := poaKeeper.GetKickProposal(ctx, validator1.GetOperator())
	if !found {
		t.Errorf("MsgVoteKickProposal with 1/2 approve should not remove the kick proposal")
	}
	if kickProposal.GetTotal() != 1 {
		t.Errorf("MsgVoteKickProposal with approve should add one vote to the kick proposal")
	}
	if kickProposal.GetApprovals() != 1 {
		t.Errorf("MsgVoteKickProposal with approve should add one approve to the kick proposal")
	}

	// Second approve should set the set of the validator to leaving
	msg = types.NewMsgVote(types.VoteTypeKickProposal, voter2.GetOperator(), validator1.GetOperator(), true)
	_, err = handler(ctx, msg)
	if err != nil {
		t.Errorf("MsgVoteKickProposal 2 should vote on a kick proposal, got error %v", err)
	}
	_, found = poaKeeper.GetKickProposal(ctx, validator1.GetOperator())
	if found {
		t.Errorf("MsgVoteKickProposal with 2/2 approve should remove the kick proposal")
	}
	validatorState, found := poaKeeper.GetValidatorState(ctx, validator1.GetOperator())
	if !found {
		t.Errorf("MsgVoteKickProposal with 2/2 approve should not directly remove the validator from the validator set")
	}
	if validatorState != types.ValidatorStateLeaving {
		t.Errorf("MsgVoteKickProposal with 2/2 approve should set the state of the validator to leaving")
	}

	// Quorum 100%: one reject is sufficient to reject the kick proposal
	poaKeeper.AppendKickProposal(ctx, voter2)
	msg = types.NewMsgVote(types.VoteTypeKickProposal, voter1.GetOperator(), voter2.GetOperator(), false)
	_, err = handler(ctx, msg)
	if err != nil {
		t.Errorf("MsgVoteKickProposal 3 should vote on a kick proposal, got error %v", err)
	}
	_, found = poaKeeper.GetKickProposal(ctx, voter2.GetOperator())
	if found {
		t.Errorf("MsgVoteKickProposal with 1 reject should reject the kick proposal")
	}
	validatorState, found = poaKeeper.GetValidatorState(ctx, voter2.GetOperator())
	if !found {
		t.Errorf("MsgVoteKickProposal kick proposal rejected should not remove the validator")
	}
	if validatorState == types.ValidatorStateLeaving {
		t.Errorf("MsgVoteKickProposal kick proposal rejected should not set the state of the validator to leaving")
	}

	// Reapply and set quorum to 1%
	poaKeeper.AppendKickProposal(ctx, voter2)
	poaKeeper.SetParams(ctx, types.NewParams(15, 1))

	// One reject should update the vote but not reject totally the kick proposal
	msg = types.NewMsgVote(types.VoteTypeKickProposal, voter1.GetOperator(), voter2.GetOperator(), false)
	_, err = handler(ctx, msg)
	if err != nil {
		t.Errorf("MsgVoteKickProposal 4 should vote on a kick proposal, got error %v", err)
	}
	kickProposal, found = poaKeeper.GetKickProposal(ctx, voter2.GetOperator())
	if !found {
		t.Errorf("MsgVoteKickProposal with 1/2 reject should not remove the kick proposal")
	}
	_, found = poaKeeper.GetValidatorState(ctx, voter2.GetOperator())
	if !found {
		t.Errorf("MsgVoteKickProposal with 1/2 rejec should not remove the validator")
	}
	if validatorState == types.ValidatorStateLeaving {
		t.Errorf("MsgVoteKickProposal with 1/2 rejec should not set the state of the validator to leaving")
	}
	if kickProposal.GetTotal() != 1 {
		t.Errorf("MsgVoteKickProposal with 1/2 reject should add one vote to the kick proposal")
	}
	if kickProposal.GetApprovals() != 0 {
		t.Errorf("MsgVoteKickProposal with 1/2 reject should not add one approve to the kick proposal")
	}
}

func testHandleMsgLeaveValidatorSet(t *testing.T) {
	ctx, poaKeeper := poa.MockContext()
	handler := poa.NewHandler(poaKeeper)
	validator1, _ := poa.MockValidator()
	validator2, _ := poa.MockValidator()
	poaKeeper.SetParams(ctx, types.DefaultParams())

	poaKeeper.AppendValidator(ctx, validator1)

	// Can't leave the validator set if only one validator
	msg := types.NewMsgLeaveValidatorSet(validator1.GetOperator())
	_, err := handler(ctx, msg)
	if err.Error() != types.ErrOnlyOneValidator.Error() {
		t.Errorf("MsgLeaveValidatorSet with one validator, error should be %v, got %v", types.ErrOnlyOneValidator.Error(), err.Error())
	}

	// Can't leave the validator set if not validator
	msg = types.NewMsgLeaveValidatorSet(validator2.GetOperator())
	_, err = handler(ctx, msg)
	if err.Error() != types.ErrNotValidator.Error() {
		t.Errorf("MsgLeaveValidatorSet when not validator, error should be %v, got %v", types.ErrNotValidator.Error(), err.Error())
	}

	poaKeeper.AppendValidator(ctx, validator2)
	poaKeeper.AppendKickProposal(ctx, validator1)

	// Can leave the validator set
	msg = types.NewMsgLeaveValidatorSet(validator1.GetOperator())
	_, err = handler(ctx, msg)
	if err != nil {
		t.Errorf("MsgLeaveValidatorSet should leave the validator set, got error %v", err)
	}
	_, found := poaKeeper.GetKickProposal(ctx, validator1.GetOperator())
	if found {
		t.Errorf("MsgLeaveValidatorSet should remove existing kick proposal")
	}
	validatorState, found := poaKeeper.GetValidatorState(ctx, validator1.GetOperator())
	if !found {
		t.Errorf("MsgLeaveValidatorSet should not directly remove the validator")
	}
	if validatorState != types.ValidatorStateLeaving {
		t.Errorf("MsgLeaveValidatorSet should set the state of the validator to leaving")
	}
}
