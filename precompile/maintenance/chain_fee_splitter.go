package maintenance

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/mezo-org/mezod/precompile"
	"github.com/mezo-org/mezod/utils"
)

// SetChainFeeSplitterAddressMethodName is the name of the setChainFeeSplitterAddress method.
// It matches the name of the method in the contract ABI.
const SetChainFeeSplitterAddressMethodName = "setChainFeeSplitterAddress"

// Define a structure for the new method
type setChainFeeSplitterAddressMethod struct {
	poaKeeper  PoaKeeper
	evmKeeper  EvmKeeper
	bankKeeper BankKeeper
}

// Function to create a new instance of the method
func newSetChainFeeSplitterAddressMethod(poaKeeper PoaKeeper, evmKeeper EvmKeeper, bankKeeper BankKeeper) *setChainFeeSplitterAddressMethod {
	return &setChainFeeSplitterAddressMethod{
		poaKeeper:  poaKeeper,
		evmKeeper:  evmKeeper,
		bankKeeper: bankKeeper,
	}
}

// Implementing the MethodName function
func (m *setChainFeeSplitterAddressMethod) MethodName() string {
	return "setChainFeeSplitterAddress"
}

// Implementing the MethodType function
func (m *setChainFeeSplitterAddressMethod) MethodType() precompile.MethodType {
	return precompile.Write
}

// Fallback to the default gas calculation.
func (m *setChainFeeSplitterAddressMethod) RequiredGas(_ []byte) (uint64, bool) {
	return 0, false
}

// This method is not payable
func (m *setChainFeeSplitterAddressMethod) Payable() bool {
	return false
}

// Implementing the Run function to handle logic
func (m *setChainFeeSplitterAddressMethod) Run(
	context *precompile.RunContext,
	inputs precompile.MethodInputs,
) (precompile.MethodOutputs, error) {
	// Validate inputs count
	if err := precompile.ValidateMethodInputsCount(inputs, 1); err != nil {
		return nil, err
	}

	// Validate and extract the chain fee splitter address
	chainFeeSplitterAddress, ok := inputs[0].(common.Address)
	if !ok {
		return nil, fmt.Errorf("value argument must be a valid address")
	}

	// This method is restricted to the validator pool owner
	err := m.poaKeeper.CheckOwner(
		context.SdkCtx(),
		precompile.TypesConverter.Address.ToSDK(context.MsgSender()),
	)
	if err != nil {
		return nil, err
	}

	chainFeeSplitterBech32, err := utils.GetBech32AccountFromMezoAddress(
		precompile.TypesConverter.Address.ToSDK(chainFeeSplitterAddress),
	)
	if err != nil {
		return nil, err
	}

	// Check if the given address is on the blocked list of addresses
	blockedAddrs := m.bankKeeper.GetBlockedAddresses()
	if blocked, exists := blockedAddrs[chainFeeSplitterBech32]; exists && blocked {
		return nil, fmt.Errorf("address is on the blocked list")
	}

	params := m.evmKeeper.GetParams(context.SdkCtx())
	params.ChainFeeSplitterAddress = chainFeeSplitterAddress.Hex()
	err = m.evmKeeper.SetParams(context.SdkCtx(), params)
	if err != nil {
		return nil, err
	}

	return precompile.MethodOutputs{true}, nil
}

// GetChainFeeSplitterAddressMethodName is the name of the getChainFeeSplitterAddress method.
// It matches the name of the method in the contract ABI.
const GetChainFeeSplitterAddressMethodName = "getChainFeeSplitterAddress"

// getChainFeeSplitterAddressMethod is the implementation of the getChainFeeSplitterAddress
// method that gets the chain fee splitter address.
// The method returns the chain fee splitter address.
type getChainFeeSplitterAddressMethod struct {
	evmKeeper EvmKeeper
}

func newGetChainFeeSplitterAddressMethod(
	evmKeeper EvmKeeper,
) *getChainFeeSplitterAddressMethod {
	return &getChainFeeSplitterAddressMethod{
		evmKeeper: evmKeeper,
	}
}

func (m *getChainFeeSplitterAddressMethod) MethodName() string {
	return GetChainFeeSplitterAddressMethodName
}

func (m *getChainFeeSplitterAddressMethod) MethodType() precompile.MethodType {
	return precompile.Read
}

func (m *getChainFeeSplitterAddressMethod) RequiredGas(_ []byte) (uint64, bool) {
	// Fallback to the default gas calculation.
	return 0, false
}

func (m *getChainFeeSplitterAddressMethod) Payable() bool {
	return false
}

func (m *getChainFeeSplitterAddressMethod) Run(
	context *precompile.RunContext,
	inputs precompile.MethodInputs,
) (precompile.MethodOutputs, error) {
	if err := precompile.ValidateMethodInputsCount(inputs, 0); err != nil {
		return nil, err
	}

	params := m.evmKeeper.GetParams(context.SdkCtx())

	return precompile.MethodOutputs{common.HexToAddress(params.ChainFeeSplitterAddress)}, nil
}
