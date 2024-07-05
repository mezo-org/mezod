package types

import (
	"fmt"
)

// Default parameter namespace
const (
	// Default max number of validators
	DefaultMaxValidators uint32 = 15
	// Default quorum percentage
	DefaultQuorum uint32 = 66
)

// ParamsKey store key for params
var ParamsKey = []byte("Params")

// NewParams creates a new Params object
func NewParams(maxValidators uint32, quorum uint32) Params {
	return Params{
		MaxValidators: maxValidators,
		Quorum:        quorum,
	}
}

// String implements the stringer interface for Params
func (p Params) String() string {
	return fmt.Sprintf("Max validators: %d, quorum: %d percents", p.MaxValidators, p.Quorum)
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
	v, ok := i.(uint32)
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
	v, ok := i.(uint32)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v > 100 {
		return fmt.Errorf("quorum must be a percentage: %d", v)
	}

	return nil
}
