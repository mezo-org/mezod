package keeper_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/evmos/evmos/v12/x/poa"
	"github.com/evmos/evmos/v12/x/poa/types"
)

func TestGetKickProposal(t *testing.T) {
	ctx, poaKeeper := poa.MockContext()
	validator1, _ := poa.MockValidator()
	validator2, _ := poa.MockValidator()
	kickProposal := types.NewVote(validator1)

	poaKeeper.SetKickProposal(ctx, kickProposal)

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

	poaKeeper.AppendKickProposal(ctx, validator)

	_, foundKickProposal := poaKeeper.GetKickProposal(ctx, validator.GetOperator())

	if !foundKickProposal {
		t.Errorf("AppendKickProposal should append the kick proposal. Found val: %v", foundKickProposal)
	}
}

func TestRemoveKickProposal(t *testing.T) {
	ctx, poaKeeper := poa.MockContext()
	validator, _ := poa.MockValidator()

	// Append and remove kick proposal
	poaKeeper.AppendKickProposal(ctx, validator)
	poaKeeper.RemoveKickProposal(ctx, validator.GetOperator())

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

	poaKeeper.SetKickProposal(ctx, kickProposal1)
	poaKeeper.SetKickProposal(ctx, kickProposal2)

	retrievedKickProposals := poaKeeper.GetAllKickProposals(ctx)
	if len(retrievedKickProposals) != 2 {
		t.Errorf("GetAllKickProposals should find %v kick proposal, found %v", 2, len(retrievedKickProposals))
	}
}
