package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/evmos/evmos/v12/x/poa/types"
)

// Get a kick proposal
func (k Keeper) GetKickProposal(ctx sdk.Context, addr sdk.ValAddress) (kickProposal types.Vote, found bool) {
	store := ctx.KVStore(k.storeKey)

	// Search the value
	value := store.Get(types.GetKickProposalKey(addr))
	if value == nil {
		return kickProposal, false
	}

	// Return the value
	kickProposal = types.MustUnmarshalVote(k.cdc, value)
	return kickProposal, true
}

// Set kick proposal details
func (k Keeper) SetKickProposal(ctx sdk.Context, kickProposal types.Vote) {
	store := ctx.KVStore(k.storeKey)
	bz := types.MustMarshalVote(k.cdc, kickProposal)
	store.Set(types.GetKickProposalKey(kickProposal.GetSubject().GetOperator()), bz)
}

// Append a new kick proposal with a new vote
func (k Keeper) AppendKickProposal(ctx sdk.Context, candidate types.Validator) {
	kickProposalNewVote := types.NewVote(candidate)
	k.SetKickProposal(ctx, kickProposalNewVote)
}

// Remove the kick proposal
func (k Keeper) RemoveKickProposal(ctx sdk.Context, address sdk.ValAddress) {
	_, found := k.GetKickProposal(ctx, address)
	if !found {
		return
	}

	// Delete the kick proposal record
	store := ctx.KVStore(k.storeKey)
	store.Delete(types.GetKickProposalKey(address))
}

// Get the set of all kick proposals
func (k Keeper) GetAllKickProposals(ctx sdk.Context) (kickProposals []types.Vote) {
	store := ctx.KVStore(k.storeKey)

	iterator := sdk.KVStorePrefixIterator(store, types.KickProposalPoolKey)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		kickProposal := types.MustUnmarshalVote(k.cdc, iterator.Value())
		kickProposals = append(kickProposals, kickProposal)
	}

	return kickProposals
}
