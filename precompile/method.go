package precompile

import (
	"fmt"
	store "github.com/cosmos/cosmos-sdk/store/types"
)

// methodIDByteLength is the length of the method ID in bytes.
const methodIDByteLength = 4

// MethodType represents the type of the method.
type MethodType int

const (
	// Read represents a read method that does not modify the state.
	Read MethodType = iota
	// Write represents a write method that modifies the state
	// (i.e. an EVM transaction).
	Write
)

// MethodInputs is a convenience type representing an array of method inputs.
type MethodInputs []interface{}

// MethodOutputs is a convenience type representing an array of method outputs.
type MethodOutputs []interface{}

// Method represents a precompiled contract method.
type Method interface {
	// MethodName returns the name of the method.
	MethodName() string
	// MethodType returns the type of the method.
	MethodType() MethodType
	// RequiredGas returns the amount of gas required to execute the method.
	// The method may not implement its own gas calculation and return false
	// in the second return value to use the default gas calculation algorithm
	// provided by the precompiled contract.
	RequiredGas(methodInputArgs []byte) (uint64, bool)
	// Payable returns true if the method is payable and can accept a native
	// transfer of tokens. This is applicable only to method whose MethodType
	// is Write.
	Payable() bool
	// Run executes the method with the given inputs and returns the outputs.
	// The method is executed in the context of the given RunContext.
	Run(
		context *RunContext,
		inputs MethodInputs,
	) (MethodOutputs, error)
}

// DefaultRequiredGas calculates the default amount of gas required to execute
// a contract call with the given method type and input arguments. The default
// gas calculation algorithm implies a flat cost plus additional cost
// based on the input size: cost = costFlat + (costPerByte * len(methodInputArgs))
// The costFlat and costPerByte values are taken from the provided gas config.
func DefaultRequiredGas(
	gasConfig store.GasConfig,
	methodType MethodType,
	methodInputArgs []byte,
) uint64 {
	methodInputArgsByteLength := uint64(len(methodInputArgs))

	costFlat := store.Gas(0)
	costPerByte := store.Gas(0)

	switch methodType {
	case Read:
		costFlat, costPerByte = gasConfig.ReadCostFlat, gasConfig.ReadCostPerByte
	case Write:
		costFlat, costPerByte = gasConfig.WriteCostFlat, gasConfig.WriteCostPerByte
	}

	return costFlat + (costPerByte * methodInputArgsByteLength)
}

// ValidateMethodInputsCount validates the count of the given method inputs
// against the expected value. If the counts don't match, an error is returned.
func ValidateMethodInputsCount(inputs MethodInputs, expectedCount int) error {
	if actualCount := len(inputs); expectedCount != actualCount {
		return fmt.Errorf(
			"wrong inputs count for method; expected [%v], actual [%v]",
			expectedCount,
			actualCount,
		)
	}

	return nil
}