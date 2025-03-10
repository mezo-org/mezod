package maintenance

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/holiman/uint256"
	"github.com/mezo-org/mezod/precompile"
	"github.com/mezo-org/mezod/x/evm/statedb"
)

// SetPrecompileByteCodeMethodName is the name of the setPrecompileByteCode method.
// It matches the name of the method in the contract ABI.
const SetPrecompileByteCodeMethodName = "setPrecompileByteCode"

// SetFeeChainSplitterAddressMethodName is the name of the setFeeChainSplitterAddress method.
// It matches the name of the method in the contract ABI.
const SetFeeChainSplitterAddressMethodName = "setFeeChainSplitterAddress"

// setPrecompileByteCodeMethod is the implementation of the setPrecompileByteCode
// method that updates a precompile associated byte code stored in the statedb
//
// The method has the following input arguments:
// - precompile: The precompile contract address
// - code: The new byte code
//
// The method returns true if the update was executed.
type setPrecompileByteCodeMethod struct {
	poaKeeper PoaKeeper
	evmKeeper EvmKeeper
}

func newSetPrecompileByteCodeMethod(
	poaKeeper PoaKeeper,
	evmKeeper EvmKeeper,
) *setPrecompileByteCodeMethod {
	return &setPrecompileByteCodeMethod{
		poaKeeper: poaKeeper,
		evmKeeper: evmKeeper,
	}
}

func (m *setPrecompileByteCodeMethod) MethodName() string {
	return SetPrecompileByteCodeMethodName
}

func (m *setPrecompileByteCodeMethod) MethodType() precompile.MethodType {
	return precompile.Write
}

func (m *setPrecompileByteCodeMethod) RequiredGas(_ []byte) (uint64, bool) {
	// Fallback to the default gas calculation.
	return 0, false
}

func (m *setPrecompileByteCodeMethod) Payable() bool {
	return false
}

func (m *setPrecompileByteCodeMethod) Run(
	context *precompile.RunContext,
	inputs precompile.MethodInputs,
) (precompile.MethodOutputs, error) {
	if err := precompile.ValidateMethodInputsCount(inputs, 2); err != nil {
		return nil, err
	}

	precompileAddress, ok := inputs[0].(common.Address)
	if !ok {
		return nil, fmt.Errorf("value argument must be a valid address")
	}

	precompileBytecode, ok := inputs[1].([]byte)
	if !ok {
		return nil, fmt.Errorf("argument must be a valid byte slice")
	}

	// this method is restricted to the validator pool owner
	err := m.poaKeeper.CheckOwner(
		context.SdkCtx(),
		precompile.TypesConverter.Address.ToSDK(context.MsgSender()),
	)
	if err != nil {
		return nil, err
	}

	// make sure we are working with a precompile address
	isCustomPrecompile := m.evmKeeper.IsCustomPrecompile(precompileAddress)
	if !isCustomPrecompile {
		return nil, fmt.Errorf("address is not a precompile")
	}

	nonce := uint64(0)
	balance := uint256.Int{0}
	account := m.evmKeeper.GetAccount(context.SdkCtx(), precompileAddress)
	if account != nil {
		// is an existing account
		nonce = account.Nonce
		balance = *account.Balance
	}

	// set new code/codeHash
	newCodeHash := crypto.Keccak256Hash(precompileBytecode)
	m.evmKeeper.SetCode(context.SdkCtx(), newCodeHash[:], precompileBytecode)

	// update/set account
	err = m.evmKeeper.SetAccount(context.SdkCtx(), precompileAddress, statedb.Account{
		Nonce:    nonce,
		Balance:  &balance,
		CodeHash: newCodeHash[:],
	})
	if err != nil {
		return nil, err
	}

	// clear old code/codeHash
	if account != nil && len(account.CodeHash) > 0 {
		m.evmKeeper.SetCode(context.SdkCtx(), account.CodeHash, []byte{})
	}

	return precompile.MethodOutputs{true}, nil
}

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
