package maintenance

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/mezo-org/mezod/precompile"
	"github.com/mezo-org/mezod/x/evm/statedb"
)

// SetPrecompileByteCodeMethodName is the name of the setPrecompileByteCode method.
// It matches the name of the method in the contract ABI.
const SetPrecompileByteCodeMethodName = "setPrecompileByteCode"

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

	account := m.evmKeeper.GetAccount(context.SdkCtx(), precompileAddress)
	prevCodeHash := account.CodeHash

	newCodeHash := crypto.Keccak256Hash(precompileBytecode)
	m.evmKeeper.SetCode(context.SdkCtx(), newCodeHash[:], precompileBytecode)

	err = m.evmKeeper.SetAccount(context.SdkCtx(), precompileAddress, statedb.Account{
		Nonce:    account.Nonce,
		Balance:  account.Balance,
		CodeHash: newCodeHash[:],
	})
	if err != nil {
		return nil, err
	}

	m.evmKeeper.SetCode(context.SdkCtx(), prevCodeHash, []byte{})

	return precompile.MethodOutputs{true}, nil
}
