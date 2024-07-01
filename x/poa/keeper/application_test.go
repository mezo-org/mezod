package keeper

import (
	"testing"

	"github.com/evmos/evmos/v12/x/poa/types"
	"github.com/google/go-cmp/cmp"
)

func TestSubmitApplication(t *testing.T) {
	// Test with maxValidator=15, quorum=66
	ctx, poaKeeper := mockContext()
	validator, _ := mockValidator()
	poaKeeper.setParams(ctx, types.DefaultParams())

	// The application is submitted correctly
	err := poaKeeper.SubmitApplication(ctx, validator)
	if err != nil {
		t.Errorf("SubmitApplication should submit an application, got error %v", err)
	}
	_, found := poaKeeper.GetApplication(ctx, validator.GetOperator())
	if !found {
		t.Errorf("SubmitApplication should submit an application, the application has not been found")
	}
	_, found = poaKeeper.GetApplicationByConsAddr(ctx, validator.GetConsAddr())
	if !found {
		t.Errorf("SubmitApplication should submit an application, the application has not been found by cons addr")
	}

	// A new application with the same validator cannot be created
	err = poaKeeper.SubmitApplication(ctx, validator)
	if err.Error() != types.ErrAlreadyApplying.Error() {
		t.Errorf("SubmitApplication with duplicate, error should be %v, got %v", types.ErrAlreadyApplying.Error(), err.Error())
	}

	// Test with quorum=0
	ctx, poaKeeper = mockContext()
	validator, _ = mockValidator()
	poaKeeper.setParams(ctx, types.NewParams(15, 0))

	// The validator should be directly appended if the quorum is 0
	err = poaKeeper.SubmitApplication(ctx, validator)
	if err != nil {
		t.Errorf("SubmitApplication with quorum 0 should append validator, got error %v", err)
	}
	_, found = poaKeeper.GetValidator(ctx, validator.GetOperator())
	if !found {
		t.Errorf("SubmitApplication with quorum 0 should append validator, the validator has not been found")
	}
	_, found = poaKeeper.GetValidatorByConsAddr(ctx, validator.GetConsAddr())
	if !found {
		t.Errorf("SubmitApplication with quorum 0 should append validator, the validator has not been found by cons addr")
	}
	foundState, found := poaKeeper.GetValidatorState(ctx, validator.GetOperator())
	if !found {
		t.Errorf("SubmitApplication with quorum 0 should append validator, the validator state has not been found")
	}
	if foundState != types.ValidatorStateJoining {
		t.Errorf("SubmitApplication with quorum 0, the validator should have the state joining, if it is appended")
	}

	// A new application cannot be created if the validator already exist
	err = poaKeeper.SubmitApplication(ctx, validator)
	if err.Error() != types.ErrAlreadyValidator.Error() {
		t.Errorf("SubmitApplication with duplicate, error should be %v, got %v", types.ErrAlreadyValidator.Error(), err.Error())
	}

	// Test max validators condition
	poaKeeper.setParams(ctx, types.NewParams(1, 0))
	err = poaKeeper.SubmitApplication(ctx, validator)
	if err.Error() != types.ErrMaxValidatorsReached.Error() {
		t.Errorf("SubmitApplication with max validators reached, error should be %v, got %v", types.ErrMaxValidatorsReached.Error(), err.Error())
	}
}

func TestVoteApplication(t *testing.T) {
	ctx, poaKeeper := mockContext()
	voter1, _ := mockValidator()
	voter2, _ := mockValidator()
	candidate1, _ := mockValidator()
	candidate2, _ := mockValidator()
	nothing, _ := mockValidator()
	poaKeeper.setParams(ctx, types.NewParams(15, 100)) // Set quorum to 100%

	// Add voter to validator set
	poaKeeper.appendValidator(ctx, voter1)
	poaKeeper.appendValidator(ctx, voter2)

	// Add candidate to application pool
	poaKeeper.appendApplication(ctx, candidate1)
	poaKeeper.appendApplication(ctx, candidate2)

	// Cannot vote if candidate is not in application pool
	err := poaKeeper.VoteApplication(ctx, voter1.GetOperator(), nothing.GetOperator(), true)
	if err.Error() != types.ErrNoApplicationFound.Error() {
		t.Errorf("VoteApplication should fail with %v, got %v", types.ErrNoApplicationFound, err)
	}

	// Cannot vote if the voter is not in validator set
	err = poaKeeper.VoteApplication(ctx, nothing.GetOperator(), candidate1.GetOperator(), true)
	if err.Error() != types.ErrVoterNotValidator.Error() {
		t.Errorf("VoteApplication should fail with %v, got %v", types.ErrVoterNotValidator, err)
	}

	// Can vote an application
	err = poaKeeper.VoteApplication(ctx, voter1.GetOperator(), candidate1.GetOperator(), true)
	if err != nil {
		t.Errorf("VoteApplication should vote on an application, got error %v", err)
	}
	application, found := poaKeeper.GetApplication(ctx, candidate1.GetOperator())
	if !found {
		t.Errorf("VoteApplication with 1/2 approve should not remove the application")
	}
	_, found = poaKeeper.GetValidator(ctx, candidate1.GetOperator())
	if found {
		t.Errorf("VoteApplication with 1/2 approve should not append the candidate to the validator set")
	}
	if application.GetTotal() != 1 {
		t.Errorf("VoteApplication with approve should add one vote to the application")
	}
	if application.GetApprovals() != 1 {
		t.Errorf("VoteApplication with approve should add one approve to the application")
	}

	// Second approve should append the candidate to the validator pool
	err = poaKeeper.VoteApplication(ctx, voter2.GetOperator(), candidate1.GetOperator(), true)
	if err != nil {
		t.Errorf("VoteApplication 2 should vote on an application, got error %v", err)
	}
	_, found = poaKeeper.GetApplication(ctx, candidate1.GetOperator())
	if found {
		t.Errorf("VoteApplication with 2/2 approve should remove the application")
	}
	_, found = poaKeeper.GetValidator(ctx, candidate1.GetOperator())
	if !found {
		t.Errorf("VoteApplication with 2/2 approve should append the candidate to the validator set")
	}

	// Quorum 100%: one reject is sufficient to reject the validator application
	err = poaKeeper.VoteApplication(ctx, voter1.GetOperator(), candidate2.GetOperator(), false)
	if err != nil {
		t.Errorf("VoteApplication 3 should vote on an application, got error %v", err)
	}
	_, found = poaKeeper.GetApplication(ctx, candidate2.GetOperator())
	if found {
		t.Errorf("VoteApplication with 1 reject should reject the application")
	}
	_, found = poaKeeper.GetValidator(ctx, candidate2.GetOperator())
	if found {
		t.Errorf("VoteApplication application rejected should not append the candidate to the validator set")
	}

	// Reapply and set quorum to 1%
	poaKeeper.appendApplication(ctx, candidate2)
	poaKeeper.setParams(ctx, types.NewParams(15, 1))

	// One reject should update the vote but not reject totally the application
	err = poaKeeper.VoteApplication(ctx, voter1.GetOperator(), candidate2.GetOperator(), false)
	if err != nil {
		t.Errorf("VoteApplication 4 should vote on an application, got error %v", err)
	}
	application, found = poaKeeper.GetApplication(ctx, candidate2.GetOperator())
	if !found {
		t.Errorf("VoteApplication with 1/3 reject should not remove the application")
	}
	_, found = poaKeeper.GetValidator(ctx, candidate2.GetOperator())
	if found {
		t.Errorf("VoteApplication with 1/3 reject should not append the candidate to the validator set")
	}
	if application.GetTotal() != 1 {
		t.Errorf("VoteApplication with reject should add one vote to the application")
	}
	if application.GetApprovals() != 0 {
		t.Errorf("VoteApplication with reject should not add one approve to the application")
	}

	// Cannot vote if validator set is full
	poaKeeper.setParams(ctx, types.NewParams(3, 1))

	err = poaKeeper.VoteApplication(ctx, voter2.GetOperator(), candidate2.GetOperator(), false)
	if err.Error() != types.ErrMaxValidatorsReached.Error() {
		t.Errorf("VoteApplication should fail with %v, got %v", types.ErrMaxValidatorsReached, err)
	}
}

func TestGetApplication(t *testing.T) {
	ctx, poaKeeper := mockContext()
	validator1, _ := mockValidator()
	validator2, _ := mockValidator()
	application := types.NewVote(validator1)

	poaKeeper.setApplication(ctx, application)

	// Should find the correct application
	retrievedApplication, found := poaKeeper.GetApplication(ctx, validator1.GetOperator())
	if !found {
		t.Errorf("GetApplication should find application if it has been set")
	}

	if !cmp.Equal(application.GetSubject(), retrievedApplication.GetSubject()) {
		t.Errorf("GetApplication should find %v, found %v", application.GetSubject(), retrievedApplication.GetSubject())
	}
	if application.GetTotal() != retrievedApplication.GetTotal() {
		t.Errorf("GetApplication should find %v votes, found %v", application.GetTotal(), retrievedApplication.GetTotal())
	}
	if application.GetApprovals() != retrievedApplication.GetApprovals() {
		t.Errorf("GetApplication should find %v approvals, found %v", application.GetApprovals(), retrievedApplication.GetApprovals())
	}

	// Should not find a unset application
	_, found = poaKeeper.GetApplication(ctx, validator2.GetOperator())
	if found {
		t.Errorf("GetApplication should not find application if it has not been set")
	}
}

func TestGetApplicationByConsAddr(t *testing.T) {
	ctx, poaKeeper := mockContext()
	validator1, _ := mockValidator()
	validator2, _ := mockValidator()
	application := types.NewVote(validator1)
	application2 := types.NewVote(validator2)

	poaKeeper.setApplication(ctx, application)
	poaKeeper.setApplicationByConsAddr(ctx, application)

	// Should find the correct application
	retrievedApplication, found := poaKeeper.GetApplicationByConsAddr(ctx, application.GetSubject().GetConsAddr())
	if !found {
		t.Errorf("GetApplicationByConsAddr should find application if it has been set")
	}

	if !cmp.Equal(application.GetSubject(), retrievedApplication.GetSubject()) {
		t.Errorf("GetApplicationByConsAddr should find %v, found %v", application.GetSubject(), retrievedApplication.GetSubject())
	}
	if application.GetTotal() != retrievedApplication.GetTotal() {
		t.Errorf("GetApplicationByConsAddr should find %v votes, found %v", application.GetTotal(), retrievedApplication.GetTotal())
	}
	if application.GetApprovals() != retrievedApplication.GetApprovals() {
		t.Errorf("GetApplicationByConsAddr should find %v approvals, found %v", application.GetApprovals(), retrievedApplication.GetApprovals())
	}

	// Should not find a unset application
	_, found = poaKeeper.GetApplication(ctx, validator2.GetOperator())
	if found {
		t.Errorf("GetApplicationByConsAddr should not find application if it has not been set")
	}

	// Should not find the application if we call SetApplicationByConsAddr without SetApplication
	poaKeeper.setApplicationByConsAddr(ctx, application2)
	_, found = poaKeeper.GetApplication(ctx, validator2.GetOperator())
	if found {
		t.Errorf("GetApplicationByConsAddr should not find application if it has not been set with SetApplication")
	}
}

func TestAppendApplication(t *testing.T) {
	ctx, poaKeeper := mockContext()
	validator, _ := mockValidator()

	poaKeeper.appendApplication(ctx, validator)

	_, foundApplication := poaKeeper.GetApplication(ctx, validator.GetOperator())
	_, foundConsAddr := poaKeeper.GetApplicationByConsAddr(ctx, validator.GetConsAddr())

	if !foundApplication || !foundConsAddr {
		t.Errorf("AppendValidator should append the application. Found val: %v, found consAddr: %v", foundApplication, foundConsAddr)
	}
}

func TestRemoveApplication(t *testing.T) {
	ctx, poaKeeper := mockContext()
	validator, _ := mockValidator()

	// Append  and remove application
	poaKeeper.appendApplication(ctx, validator)
	poaKeeper.removeApplication(ctx, validator.GetOperator())

	// Should not find a removed validator
	_, foundApplication := poaKeeper.GetApplication(ctx, validator.GetOperator())
	_, foundConsAddr := poaKeeper.GetApplicationByConsAddr(ctx, validator.GetConsAddr())

	if foundApplication || foundConsAddr {
		t.Errorf("RemoveApplication should remove application record. Found val: %v, found consAddr: %v", foundApplication, foundConsAddr)
	}
}

func TestGetAllApplications(t *testing.T) {
	ctx, poaKeeper := mockContext()
	validator1, _ := mockValidator()
	validator2, _ := mockValidator()
	application1 := types.NewVote(validator1)
	application2 := types.NewVote(validator2)

	poaKeeper.setApplication(ctx, application1)
	poaKeeper.setApplication(ctx, application2)

	retrievedApplications := poaKeeper.GetAllApplications(ctx)
	if len(retrievedApplications) != 2 {
		t.Errorf("GetAllApplications should find %v applications, found %v", 2, len(retrievedApplications))
	}
}
