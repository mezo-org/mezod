package maintenance

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/mezo-org/mezod/precompile"
)

// SetFeeChainSplitterAddressMethodName is the name of the setFeeChainSplitterAddress method.
// It matches the name of the method in the contract ABI.
const SetFeeChainSplitterAddressMethodName = "setFeeChainSplitterAddress"

// Define a structure for the new method
type setFeeChainSplitterAddressMethod struct {
	poaKeeper PoaKeeper
	evmKeeper EvmKeeper
}

// Function to create a new instance of the method
func newSetFeeChainSplitterAddressMethod(poaKeeper PoaKeeper, evmKeeper EvmKeeper) *setFeeChainSplitterAddressMethod {
	return &setFeeChainSplitterAddressMethod{
		poaKeeper: poaKeeper,
		evmKeeper: evmKeeper,
	}
}

// Implementing the MethodName function
func (m *setFeeChainSplitterAddressMethod) MethodName() string {
	return "setFeeChainSplitterAddress"
}

// Implementing the MethodType function
func (m *setFeeChainSplitterAddressMethod) MethodType() precompile.MethodType {
	return precompile.Write
}

// Fallback to the default gas calculation.
func (m *setFeeChainSplitterAddressMethod) RequiredGas(_ []byte) (uint64, bool) {
	return 0, false
}

// This method is not payable
func (m *setFeeChainSplitterAddressMethod) Payable() bool {
	return false
}

// Implementing the Run function to handle logic
func (m *setFeeChainSplitterAddressMethod) Run(
	context *precompile.RunContext,
	inputs precompile.MethodInputs,
) (precompile.MethodOutputs, error) {
	// Validate inputs count
	if err := precompile.ValidateMethodInputsCount(inputs, 1); err != nil {
		return nil, err
	}

	// Validate and extract the fee chain splitter address
	feeChainSplitterAddress, ok := inputs[0].(common.Address)
	if !ok {
		return nil, fmt.Errorf("value argument must be a valid address")
	}

	// This method assumes some restriction logic which can be defined
	err := m.poaKeeper.CheckOwner(
		context.SdkCtx(),
		precompile.TypesConverter.Address.ToSDK(context.MsgSender()),
	)
	if err != nil {
		return nil, err
	}

	params := m.evmKeeper.GetParams(context.SdkCtx())
	params.ChainFeeSplitterAddress = feeChainSplitterAddress.String()
	err = m.evmKeeper.SetParams(context.SdkCtx(), params)
	if err != nil {
		return nil, err
	}

	return precompile.MethodOutputs{true}, nil
}

// GetFeeChainSplitterAddressMethodName is the name of the getFeeChainSplitterAddress method.
// It matches the name of the method in the contract ABI.
const GetFeeChainSplitterAddressMethodName = "getFeeChainSplitterAddress"

// getFeeChainSplitterAddressMethod is the implementation of the getFeeChainSplitterAddress
// method that gets the fee chain splitter address.
// The method returns the fee chain splitter address.
type getFeeChainSplitterAddressMethod struct {
	evmKeeper EvmKeeper
}

func newGetFeeChainSplitterAddressMethod(
	evmKeeper EvmKeeper,
) *getFeeChainSplitterAddressMethod {
	return &getFeeChainSplitterAddressMethod{
		evmKeeper: evmKeeper,
	}
}

func (m *getFeeChainSplitterAddressMethod) MethodName() string {
	return GetFeeChainSplitterAddressMethodName
}

func (m *getFeeChainSplitterAddressMethod) MethodType() precompile.MethodType {
	return precompile.Read
}

func (m *getFeeChainSplitterAddressMethod) RequiredGas(_ []byte) (uint64, bool) {
	// Fallback to the default gas calculation.
	return 0, false
}

func (m *getFeeChainSplitterAddressMethod) Payable() bool {
	return false
}

func (m *getFeeChainSplitterAddressMethod) Run(
	context *precompile.RunContext,
	inputs precompile.MethodInputs,
) (precompile.MethodOutputs, error) {
	if err := precompile.ValidateMethodInputsCount(inputs, 0); err != nil {
		return nil, err
	}

	params := m.evmKeeper.GetParams(context.SdkCtx())

	return precompile.MethodOutputs{params.ChainFeeSplitterAddress}, nil
}
