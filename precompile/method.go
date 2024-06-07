package precompile

import store "github.com/cosmos/cosmos-sdk/store/types"

const methodIDByteLength = 4

type MethodType int

const (
	Read MethodType = iota
	Write
)

type MethodInputs []interface{}

type MethodOutputs []interface{}

type Method interface {
	MethodName() string
	MethodType() MethodType
	RequiredGas(methodInputArgs []byte) (uint64, bool)
	Payable() bool
	Run(
		context *RunContext,
		inputs MethodInputs,
	) (MethodOutputs, error)
}

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