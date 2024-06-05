package precompile

import store "github.com/cosmos/cosmos-sdk/store/types"

const methodIDByteLength = 4

type MethodType int

const (
	Read MethodType = iota
	Write
)

type Method interface {
	MethodName() string
	MethodType() MethodType
	RequiredGas(methodArgs []byte) (uint64, bool)
	Run() // TODO: Specify this method.
}

func DefaultRequiredGas(
	gasConfig store.GasConfig,
	methodType MethodType,
	methodArgs []byte,
) uint64 {
	methodArgsByteLength := uint64(len(methodArgs))

	costFlat := store.Gas(0)
	costPerByte := store.Gas(0)

	switch methodType {
	case Read:
		costFlat, costPerByte = gasConfig.ReadCostFlat, gasConfig.ReadCostPerByte
	case Write:
		costFlat, costPerByte = gasConfig.WriteCostFlat, gasConfig.WriteCostPerByte
	}

	return costFlat + (costPerByte * methodArgsByteLength)
}