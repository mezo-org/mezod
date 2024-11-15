package keeper

import (
	"context"
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/mezo-org/mezod/x/poa/types"
)

// ValidatorByConsAddr is a compatibility method used by Connect. It wraps the PoaKeeper Validator type in a
// ValidatorCompat struct which provides compatibility with the x/staking methods that Connect expects.
func (k Keeper) ValidatorByConsAddr(ctx context.Context, addr sdk.ConsAddress) (stakingtypes.ValidatorI, error) {
	val, found := k.GetValidatorByConsAddr(sdk.UnwrapSDKContext(ctx), addr)
	if !found {
		return nil, types.ErrNoValidatorFound
	}
	return types.ValidatorCompat{Validator: val}, nil
}

// TotalBondedTokens is used by Connect to retrieve the total "stake" from which to calculate the stake-weighted
// median. Each active (including leaving) validator in the PoaKeeper will return 1 for its bonded tokens and is
// effectively/weighted
// equally.
func (k Keeper) TotalBondedTokens(ctx context.Context) (math.Int, error) {
	// Each Validator has 1 Token, Total tokens is the number of active + leaving vals
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	var total int64
	for _, val := range k.GetAllValidators(sdk.UnwrapSDKContext(sdkCtx)) {
		valState, found := k.GetValidatorState(sdkCtx, val.GetOperator())
		if found && (valState == types.ValidatorStateActive || valState == types.ValidatorStateLeaving) {
			total += 1
		}
	}
	return math.NewInt(total), nil
}
