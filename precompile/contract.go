package precompile

import (
	"fmt"
	store "github.com/cosmos/cosmos-sdk/store/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"
)

var _ vm.PrecompiledContract = &Contract{}

type Contract struct {
	abi abi.ABI
	address common.Address
	methods map[string]Method
}

func NewContract(abi abi.ABI, address common.Address) *Contract {
	return &Contract{
		abi:     abi,
		address: address,
		methods: make(map[string]Method),
	}
}

func (c *Contract) RegisterMethods(methods ...Method) *Contract {
	for _, method := range methods {
		c.methods[method.MethodName()] = method
	}

	return c
}

func (c *Contract) Address() common.Address {
	return c.address
}

func (c *Contract) RequiredGas(input []byte) uint64 {
	method, methodArgs, err := c.parseInput(input)
	if err != nil {
		// Panic is unacceptable here, return 0 instead.
		return 0
	}

	requiredGas, ok := method.RequiredGas(methodArgs)
	if !ok {
		// Fall back to default required gas if method does not determine
		// the required gas by itself.
		requiredGas = DefaultRequiredGas(c.kvGasConfig(), method.MethodType(), methodArgs)
	}

	return requiredGas
}

func (c *Contract) Run(
	evm *vm.EVM,
	contract *vm.Contract,
	readonly bool,
) ([]byte, error) {
	// TODO: Implement this method.
	panic("implement me")
}

func (c *Contract) parseInput(input []byte) (Method, []byte, error) {
	// Prevent the bounds out of range panic.
	if len(input) < methodIDByteLength {
		return nil, nil, fmt.Errorf("input is shorter than method ID length")
	}

	methodID := input[:methodIDByteLength]
	methodArgs := input[methodIDByteLength:]

	methodABI, err := c.abi.MethodById(methodID)
	if err != nil {
		return nil, nil, fmt.Errorf("method not found in ABI: [%w]", err)
	}

	method, ok := c.methods[methodABI.Name]
	if !ok {
		return nil, nil, fmt.Errorf("method not found in precompile")
	}

	return method, methodArgs, nil
}

func (c *Contract) kvGasConfig() store.GasConfig {
	return store.KVGasConfig()
}

func (c *Contract) transientGasConfig() store.GasConfig {
	return store.TransientGasConfig()
}