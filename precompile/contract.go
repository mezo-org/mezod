package precompile

import (
	"fmt"
	store "github.com/cosmos/cosmos-sdk/store/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/evmos/evmos/v12/x/evm/statedb"
	"math/big"
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
	methodID, methodInputArgs, err := c.parseCallInput(input)
	if err != nil {
		// Panic is unacceptable here, return 0 instead.
		return 0
	}

	method, _, err := c.methodByID(methodID)
	if err != nil {
		// Panic is unacceptable here, return 0 instead.
		return 0
	}

	requiredGas, ok := method.RequiredGas(methodInputArgs)
	if !ok {
		// Fall back to default required gas if method does not determine
		// the required gas by itself.
		requiredGas = DefaultRequiredGas(
			store.KVGasConfig(),
			method.MethodType(),
			methodInputArgs,
		)
	}

	return requiredGas
}

func (c *Contract) Run(
	evm *vm.EVM,
	contract *vm.Contract,
	readonly bool,
) (methodOutputArgs []byte, runErr error) {
	stateDB, ok := evm.StateDB.(*statedb.StateDB)
	if !ok {
		return nil, fmt.Errorf("cannot get state DB from EVM")
	}

	sdkCtx := stateDB.GetContext()

	// Capture the initial values of gas config to restore them after execution.
	kvGasConfig, transientKVGasConfig := sdkCtx.KVGasConfig(), sdkCtx.TransientKVGasConfig()
	// Use a zero gas config for Cosmos SDK operations to avoid extra costs
	// apart the RequiredGas already consumed on the EVM level.
	zeroGasConfig := store.GasConfig{}
	sdkCtx = sdkCtx.
		WithKVGasConfig(zeroGasConfig).
		WithTransientKVGasConfig(zeroGasConfig)
	// Set a deferred function to restore the initial gas config values
	// after the method execution.
	defer func() {
		sdkCtx = sdkCtx.
			WithKVGasConfig(kvGasConfig).
			WithTransientKVGasConfig(transientKVGasConfig)
	}()

	methodID, methodInputArgs, err := c.parseCallInput(contract.Input)
	if err != nil {
		return nil, fmt.Errorf("failed to parse contract input: [%w]", err)
	}

	method, methodABI, err := c.methodByID(methodID)
	if err != nil {
		return nil, fmt.Errorf("failed to get method by ID: [%w]", err)
	}

	if readonly && method.MethodType() == Write {
		return nil, fmt.Errorf("cannot call write method in read-only mode")
	}

	// Commit any draft changes to the EVM state DB before running the method.
	if err := stateDB.Commit(); err != nil {
		return nil, err
	}

	methodInputs, err := methodABI.Inputs.Unpack(methodInputArgs)
	if err != nil {
		return nil, fmt.Errorf("failed to unpack method input args: [%w]", err)
	}

	runCtx := &RunContext{
		evm: evm,
		contract: contract,
	}

	methodOutputs, err := method.Run(runCtx, methodInputs)
	if err != nil {
		return nil, fmt.Errorf("method errored out: [%w]", err)
	}

	methodOutputArgs, err = methodABI.Outputs.Pack(methodOutputs...)
	if err != nil {
		return nil, fmt.Errorf("failed to pack method output args: [%w]", err)
	}

	return methodOutputArgs, nil
}

func (c *Contract) parseCallInput(input []byte) ([]byte, []byte, error) {
	// We expect a proper method call. In any other case, prevent the bounds
	// out of range panic and return an error.
	if len(input) < methodIDByteLength {
		return nil, nil, fmt.Errorf("input is shorter than method ID length")
	}

	methodID := input[:methodIDByteLength]
	methodInputArgs := input[methodIDByteLength:]

	return methodID, methodInputArgs, nil
}

func (c *Contract) methodByID(methodID []byte) (Method, *abi.Method, error) {
	methodABI, err := c.abi.MethodById(methodID)
	if err != nil {
		return nil, nil, fmt.Errorf("method not found in ABI: [%w]", err)
	}

	// If the method is not a regular function, return an error because it's
	// either a constructor, a fallback, or a receive function.
	if abiType := methodABI.Type; abiType != abi.Function {
		return nil, nil, fmt.Errorf(
			"unexpected method type: [%v]",
			abiType,
		)
	}

	method, ok := c.methods[methodABI.Name]
	if !ok {
		return nil, nil, fmt.Errorf("method not found in precompile")
	}

	return method, methodABI, nil
}

type RunContext struct {
	evm *vm.EVM
	contract *vm.Contract
}

func (rc *RunContext) MsgSender() common.Address {
	return rc.contract.Caller()
}

func (rc *RunContext) TxOrigin() common.Address {
	return rc.evm.Origin
}

func (rc *RunContext) MsgValue() *big.Int {
	if value := rc.contract.Value(); value != nil {
		return value
	}

	return big.NewInt(0)
}