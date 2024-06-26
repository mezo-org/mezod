package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Query endpoints supported by the poa querier
const (
	QueryValidators    = "validators"
	QueryValidator     = "validator"
	QueryParams        = "params"
	QueryApplications  = "applications"
	QueryKickProposals = "kick-proposals"
)

// Defines the params for the following queries:
// - 'custom/poa/validator'
type QueryValidatorParams struct {
	ValidatorAddr sdk.ValAddress
}

func NewQueryValidatorParams(validatorAddr sdk.ValAddress) QueryValidatorParams {
	return QueryValidatorParams{
		ValidatorAddr: validatorAddr,
	}
}
