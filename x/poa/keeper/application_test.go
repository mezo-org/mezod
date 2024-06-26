package keeper_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/evmos/evmos/v12/x/poa"
	"github.com/evmos/evmos/v12/x/poa/types"
)

func TestGetApplication(t *testing.T) {
	ctx, poaKeeper := poa.MockContext()
	validator1, _ := poa.MockValidator()
	validator2, _ := poa.MockValidator()
	application := types.NewVote(validator1)

	poaKeeper.SetApplication(ctx, application)

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
	ctx, poaKeeper := poa.MockContext()
	validator1, _ := poa.MockValidator()
	validator2, _ := poa.MockValidator()
	application := types.NewVote(validator1)
	application2 := types.NewVote(validator2)

	poaKeeper.SetApplication(ctx, application)
	poaKeeper.SetApplicationByConsAddr(ctx, application)

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
	poaKeeper.SetApplicationByConsAddr(ctx, application2)
	_, found = poaKeeper.GetApplication(ctx, validator2.GetOperator())
	if found {
		t.Errorf("GetApplicationByConsAddr should not find application if it has not been set with SetApplication")
	}
}

func TestAppendApplication(t *testing.T) {
	ctx, poaKeeper := poa.MockContext()
	validator, _ := poa.MockValidator()

	poaKeeper.AppendApplication(ctx, validator)

	_, foundApplication := poaKeeper.GetApplication(ctx, validator.GetOperator())
	_, foundConsAddr := poaKeeper.GetApplicationByConsAddr(ctx, validator.GetConsAddr())

	if !foundApplication || !foundConsAddr {
		t.Errorf("AppendValidator should append the application. Found val: %v, found consAddr: %v", foundApplication, foundConsAddr)
	}
}

func TestRemoveApplication(t *testing.T) {
	ctx, poaKeeper := poa.MockContext()
	validator, _ := poa.MockValidator()

	// Append  and remove application
	poaKeeper.AppendApplication(ctx, validator)
	poaKeeper.RemoveApplication(ctx, validator.GetOperator())

	// Should not find a removed validator
	_, foundApplication := poaKeeper.GetApplication(ctx, validator.GetOperator())
	_, foundConsAddr := poaKeeper.GetApplicationByConsAddr(ctx, validator.GetConsAddr())

	if foundApplication || foundConsAddr {
		t.Errorf("RemoveApplication should remove application record. Found val: %v, found consAddr: %v", foundApplication, foundConsAddr)
	}
}

func TestGetAllApplications(t *testing.T) {
	ctx, poaKeeper := poa.MockContext()
	validator1, _ := poa.MockValidator()
	validator2, _ := poa.MockValidator()
	application1 := types.NewVote(validator1)
	application2 := types.NewVote(validator2)

	poaKeeper.SetApplication(ctx, application1)
	poaKeeper.SetApplication(ctx, application2)

	retrievedApplications := poaKeeper.GetAllApplications(ctx)
	if len(retrievedApplications) != 2 {
		t.Errorf("GetAllApplications should find %v applications, found %v", 2, len(retrievedApplications))
	}
}
