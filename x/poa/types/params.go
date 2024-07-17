package types

import (
	"fmt"
)

// Default parameter namespace
const (
	// DefaultMaxValidators is the default maximum number of validators.
	DefaultMaxValidators uint32 = 150
	// DefaultHistoricalEntries is the default number of historical entries
	// to persist in store.
	DefaultHistoricalEntries uint32 = 10000
)

// ParamsKey store key for params
var ParamsKey = []byte("Params")

// NewParams creates a new Params object
func NewParams(maxValidators uint32) Params {
	return Params{
		MaxValidators: maxValidators,
	}
}

// String implements the stringer interface for Params
func (p Params) String() string {
	return fmt.Sprintf("max validators: %d", p.MaxValidators)
}

// DefaultParams defines the parameters for this module
func DefaultParams() Params {
	return NewParams(DefaultMaxValidators)
}

// Validate a set of params
func (p Params) Validate() error {
	if err := validateMaxValidators(p.MaxValidators); err != nil {
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
