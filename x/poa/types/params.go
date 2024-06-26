package types

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/x/params"
)

// Default parameter namespace
const (
	// Default max number of validators
	DefaultMaxValidators uint16 = 15
	// Default quorum percentage
	DefaultQuorum uint16 = 66
)

// Parameter store keys
var (
	KeyMaxValidators = []byte("MaxValidators")
	KeyQuorum        = []byte("Quorum")
)

// ParamKeyTable for poa module
func ParamKeyTable() params.KeyTable {
	return params.NewKeyTable().RegisterParamSet(&Params{})
}

// Params - used for initializing default parameter for poa at genesis
type Params struct {
	MaxValidators uint16 `json:"max_validators"`
	Quorum        uint16 `json:"quorum"`
}

// NewParams creates a new Params object
func NewParams(maxValidators uint16, quorum uint16) Params {
	return Params{
		MaxValidators: maxValidators,
		Quorum:        quorum,
	}
}

// String implements the stringer interface for Params
func (p Params) String() string {
	return fmt.Sprintf("Max validators: %d, quorum: %d percents", p.MaxValidators, p.Quorum)
}

// ParamSetPairs - Implements params.ParamSet
func (p *Params) ParamSetPairs() params.ParamSetPairs {
	return params.ParamSetPairs{
		params.NewParamSetPair(KeyMaxValidators, &p.MaxValidators, validateMaxValidators),
		params.NewParamSetPair(KeyQuorum, &p.Quorum, validateQuorum),
	}
}

// DefaultParams defines the parameters for this module
func DefaultParams() Params {
	return NewParams(DefaultMaxValidators, DefaultQuorum)
}

// Validate a set of params
func (p Params) Validate() error {
	if err := validateMaxValidators(p.MaxValidators); err != nil {
		return err
	}
	if err := validateQuorum(p.Quorum); err != nil {
		return err
	}
	return nil
}

// Validate maxValidators param
func validateMaxValidators(i interface{}) error {
	v, ok := i.(uint16)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v == 0 {
		return fmt.Errorf("max validators must be positive: %d", v)
	}

	return nil
}

// Quorum must be a percentage
func validateQuorum(i interface{}) error {
	v, ok := i.(uint16)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v > 100 {
		return fmt.Errorf("quorum must be a percentage: %d", v)
	}

	return nil
}
