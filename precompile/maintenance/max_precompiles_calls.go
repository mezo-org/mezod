package maintenance

import (
	"fmt"

	"github.com/mezo-org/mezod/precompile"
)

// SetMaxPrecompilesCallsPerExecutionMethodName is the name
// of the setMaxPrecompilesCallsPerExecution method.
// It matches the name of the method in the contract ABI.
const SetMaxPrecompilesCallsPerExecutionMethodName = "setMaxPrecompilesCallsPerExecution"

type setMaxPrecompilesCallsPerExecutionMethod struct {
	poaKeeper PoaKeeper
	evmKeeper EvmKeeper
}

func newSetMaxPrecompilesCallsPerExecutionMethod(
	poaKeeper PoaKeeper,
	evmKeeper EvmKeeper,
) *setMaxPrecompilesCallsPerExecutionMethod {
	return &setMaxPrecompilesCallsPerExecutionMethod{
		poaKeeper: poaKeeper,
		evmKeeper: evmKeeper,
	}
}

func (m *setMaxPrecompilesCallsPerExecutionMethod) MethodName() string {
	return SetMaxPrecompilesCallsPerExecutionMethodName
}

func (m *setMaxPrecompilesCallsPerExecutionMethod) MethodType() precompile.MethodType {
	return precompile.Write
}

func (m *setMaxPrecompilesCallsPerExecutionMethod) RequiredGas(_ []byte) (uint64, bool) {
	return 0, false
}

func (m *setMaxPrecompilesCallsPerExecutionMethod) Payable() bool {
	return false
}

func (m *setMaxPrecompilesCallsPerExecutionMethod) Run(
	context *precompile.RunContext,
	inputs precompile.MethodInputs,
) (precompile.MethodOutputs, error) {
	if err := precompile.ValidateMethodInputsCount(inputs, 1); err != nil {
		return nil, err
	}

	value, ok := inputs[0].(uint32)
	if !ok {
		return nil, fmt.Errorf("invalid max precompiles calls per execution type")
	}

	if value < 1 {
		return nil, fmt.Errorf("max precompiles calls per execution must be at least 1")
	}

	// This method is restricted to the PoA owner
	err := m.poaKeeper.CheckOwner(
		context.SdkCtx(),
		precompile.TypesConverter.Address.ToSDK(context.MsgSender()),
	)
	if err != nil {
		return nil, err
	}

	params := m.evmKeeper.GetParams(context.SdkCtx())
	params.MaxPrecompilesCallsPerExecution = value
	err = m.evmKeeper.SetParams(context.SdkCtx(), params)
	if err != nil {
		return nil, err
	}

	return precompile.MethodOutputs{true}, nil
}

// GetMaxPrecompilesCallsPerExecutionMethodName is the name
// of the getMaxPrecompilesCallsPerExecution method.
// It matches the name of the method in the contract ABI.
const GetMaxPrecompilesCallsPerExecutionMethodName = "getMaxPrecompilesCallsPerExecution"

type getMaxPrecompilesCallsPerExecutionMethod struct {
	evmKeeper EvmKeeper
}

func newGetMaxPrecompilesCallsPerExecutionMethod(
	evmKeeper EvmKeeper,
) *getMaxPrecompilesCallsPerExecutionMethod {
	return &getMaxPrecompilesCallsPerExecutionMethod{evmKeeper: evmKeeper}
}

func (m *getMaxPrecompilesCallsPerExecutionMethod) MethodName() string {
	return GetMaxPrecompilesCallsPerExecutionMethodName
}

func (m *getMaxPrecompilesCallsPerExecutionMethod) MethodType() precompile.MethodType {
	return precompile.Read
}

func (m *getMaxPrecompilesCallsPerExecutionMethod) RequiredGas(_ []byte) (uint64, bool) {
	return 0, false
}

func (m *getMaxPrecompilesCallsPerExecutionMethod) Payable() bool {
	return false
}

func (m *getMaxPrecompilesCallsPerExecutionMethod) Run(
	context *precompile.RunContext,
	inputs precompile.MethodInputs,
) (precompile.MethodOutputs, error) {
	if err := precompile.ValidateMethodInputsCount(inputs, 0); err != nil {
		return nil, err
	}

	params := m.evmKeeper.GetParams(context.SdkCtx())

	return precompile.MethodOutputs{params.MaxPrecompilesCallsPerExecution}, nil
}
