package precompile

import (
	"fmt"
	store "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
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
		requiredGas = DefaultRequiredGas(c.kvGasConfig(), method.MethodType(), methodInputArgs)
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

	gasStart := sdkCtx.GasMeter().GasConsumed()

	// The gas meter panics in case of an out-of-gas error. Recover from the panic
	// and handle the error gracefully by returning an EVM-specific out-of-gas error.
	defer func() {
		if r := recover(); r != nil {
			switch r.(type) {
			case sdk.ErrorOutOfGas:
				// Calculate and use gas used by the EVM before the panic.
				gasUsed := sdkCtx.GasMeter().GasConsumed() - gasStart
				_ = contract.UseGas(gasUsed)
				// Return an EVM-specific out-of-gas error.
				runErr = vm.ErrOutOfGas
				// Reset the gas config in the shared SDK context to the
				// zero value. This is an opposite action to the one that is
				// done upon gas meter recreation, just before method execution.
				sdkCtx = sdkCtx.
					WithKVGasConfig(store.GasConfig{}).
					WithTransientKVGasConfig(store.GasConfig{})
			default:
				panic(r)
			}
		}
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

	// Recreate the gas meter with contract gas limit applied.
	// Set gas config for persistent and transient KV store explicitly.
	// This is necessary as the context is shared between modules that may
	// apply different gas configs on their own.
	sdkCtx = sdkCtx.
		WithGasMeter(store.NewGasMeter(contract.Gas)).
		WithKVGasConfig(c.kvGasConfig()).
		WithTransientKVGasConfig(c.transientGasConfig())
	// As the gas meter was recreated, consume the gas that was already
	// used by the EVM.
	sdkCtx.GasMeter().ConsumeGas(gasStart, "consume gas already used by EVM")

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

	gasUsed := sdkCtx.GasMeter().GasConsumed() - gasStart

	if !contract.UseGas(gasUsed) {
		return nil, vm.ErrOutOfGas
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

func (c *Contract) kvGasConfig() store.GasConfig {
	return store.KVGasConfig()
}

func (c *Contract) transientGasConfig() store.GasConfig {
	return store.TransientGasConfig()
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