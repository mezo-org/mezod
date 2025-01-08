package maintenance

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/mezo-org/mezod/precompile"
	"github.com/mezo-org/mezod/x/evm/statedb"
)

// SetCustomPrecompileByteCodeMethodName is the name of the setCustomPrecompileByteCode method.
// It matches the name of the method in the contract ABI.
const SetCustomPrecompileByteCodeMethodName = "setCustomPrecompileByteCode"

// setCustomPrecompileByteCodeMethod is the implementation of the setCustomPrecompileByteCode
// method that enables/disables support for non-EIP155 txs.
//
// The method has the following input arguments:
// - value: The new value of the flag.
//
// The method returns true if the update was executed.
type setCustomPrecompileByteCodeMethod struct {
	poaKeeper PoaKeeper
	evmKeeper EvmKeeper
}

func newSetCustomPrecompileByteCodeMethod(
	poaKeeper PoaKeeper,
	evmKeeper EvmKeeper,
) *setCustomPrecompileByteCodeMethod {
	return &setCustomPrecompileByteCodeMethod{
		poaKeeper: poaKeeper,
		evmKeeper: evmKeeper,
	}
}

func (m *setCustomPrecompileByteCodeMethod) MethodName() string {
	return SetCustomPrecompileByteCodeMethodName
}

func (m *setCustomPrecompileByteCodeMethod) MethodType() precompile.MethodType {
	return precompile.Write
}

func (m *setCustomPrecompileByteCodeMethod) RequiredGas(_ []byte) (uint64, bool) {
	// Fallback to the default gas calculation.
	return 0, false
}

func (m *setCustomPrecompileByteCodeMethod) Payable() bool {
	return false
}

func (m *setCustomPrecompileByteCodeMethod) Run(
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
