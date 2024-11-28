package maintenance

import (
	"fmt"

	"github.com/mezo-org/mezod/precompile"
)

// SetSupportNonEIP155TxsMethodName is the name of the setSupportNonEIP155Txs method.
// It matches the name of the method in the contract ABI.
const SetSupportNonEIP155TxsMethodName = "setSupportNonEIP155Txs"

// setSupportNonEIP155TxsMethod is the implementation of the setSupportNonEIP155Txs
// method that enables/disables support for non-EIP155 txs.
//
// The method has the following input arguments:
// - value: The new value of the flag.
//
// The method returns true if the update was executed.
type setSupportNonEIP155TxsMethod struct {
	poaKeeper PoaKeeper
	evmKeeper EvmKeeper
}

func newSetSupportNonEIP155TxsMethod(
	poaKeeper PoaKeeper,
	evmKeeper EvmKeeper,
) *setSupportNonEIP155TxsMethod {
	return &setSupportNonEIP155TxsMethod{
		poaKeeper: poaKeeper,
		evmKeeper: evmKeeper,
	}
}

func (m *setSupportNonEIP155TxsMethod) MethodName() string {
	return SetSupportNonEIP155TxsMethodName
}

func (m *setSupportNonEIP155TxsMethod) MethodType() precompile.MethodType {
	return precompile.Write
}

func (m *setSupportNonEIP155TxsMethod) RequiredGas(_ []byte) (uint64, bool) {
	// Fallback to the default gas calculation.
	return 0, false
}

func (m *setSupportNonEIP155TxsMethod) Payable() bool {
	return false
}

func (m *setSupportNonEIP155TxsMethod) Run(
	context *precompile.RunContext,
	inputs precompile.MethodInputs,
) (precompile.MethodOutputs, error) {
	if err := precompile.ValidateMethodInputsCount(inputs, 1); err != nil {
		return nil, err
	}

	value, ok := inputs[0].(bool)
	if !ok {
		return nil, fmt.Errorf("value argument must be bool")
	}

	err := m.poaKeeper.CheckOwner(
		context.SdkCtx(),
		precompile.TypesConverter.Address.ToSDK(context.MsgSender()),
	)
	if err != nil {
		return nil, err
	}

	params := m.evmKeeper.GetParams(context.SdkCtx())
	params.AllowUnprotectedTxs = value
	err = m.evmKeeper.SetParams(context.SdkCtx(), params)
	if err != nil {
		return nil, err
	}

	return precompile.MethodOutputs{true}, nil
}

// GetSupportNonEIP155TxsMethodName is the name of the getSupportNonEIP155Txs method.
// It matches the name of the method in the contract ABI.
const GetSupportNonEIP155TxsMethodName = "getSupportNonEIP155Txs"

// getSupportNonEIP155TxsMethod is the implementation of the getSupportNonEIP155Txs
// method that gets the status of support for non-EIP155 txs.
//
// The method returns true if the support is enabled and false otherwise.
type getSupportNonEIP155TxsMethod struct {
	evmKeeper EvmKeeper
}

func newGetSupportNonEIP155TxsMethod(
	evmKeeper EvmKeeper,
) *getSupportNonEIP155TxsMethod {
	return &getSupportNonEIP155TxsMethod{
		evmKeeper: evmKeeper,
	}
}

func (m *getSupportNonEIP155TxsMethod) MethodName() string {
	return GetSupportNonEIP155TxsMethodName
}

func (m *getSupportNonEIP155TxsMethod) MethodType() precompile.MethodType {
	return precompile.Read
}

func (m *getSupportNonEIP155TxsMethod) RequiredGas(_ []byte) (uint64, bool) {
	// Fallback to the default gas calculation.
	return 0, false
}

func (m *getSupportNonEIP155TxsMethod) Payable() bool {
	return false
}

func (m *getSupportNonEIP155TxsMethod) Run(
	context *precompile.RunContext,
	inputs precompile.MethodInputs,
) (precompile.MethodOutputs, error) {
	if err := precompile.ValidateMethodInputsCount(inputs, 0); err != nil {
		return nil, err
	}

	params := m.evmKeeper.GetParams(context.SdkCtx())

	return precompile.MethodOutputs{params.AllowUnprotectedTxs}, nil
}
