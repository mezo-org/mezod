package precompile

import (
	"fmt"
	"math/big"
	"reflect"
	"testing"

	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/holiman/uint256"
	"github.com/mezo-org/mezod/x/evm/statedb"
)

func TestContract_Address(t *testing.T) {
	address := common.HexToAddress("0x1")

	contract := NewContract(abi.ABI{}, address, "")

	if actualAddress := contract.Address(); address != actualAddress {
		t.Errorf(
			"unexpected address\n expected: %v\n actual:   %v",
			address,
			actualAddress,
		)
	}
}

func TestContract_RequiredGas(t *testing.T) {
	tests := map[string]struct {
		method   Method
		expected uint64
	}{
		"method implements its own gas calculation": {
			method: &mockMethod{methodName: "testMethod", requiredGas: 10},
			// Gas cost taken directly from the mock method.
			expected: 10,
		},
		"write method does not implement its own gas calculation": {
			method: &mockMethod{methodName: "testMethod", methodType: Write},
			// writeCostFlat + (writeCostPerByte * methodInputArgsByteLength) = 2000 + 30 * 2 = 2060
			// Flat and per-byte costs are taken from the default gas config: store.KVGasConfig()
			expected: 2060,
		},
		"read method does not implement its own gas calculation": {
			method: &mockMethod{methodName: "testMethod", methodType: Read},
			// readCostFlat + (readCostPerByte * methodInputArgsByteLength) = 1000 + 3 * 2 = 1006
			// Flat and per-byte costs are taken from the default gas config: store.KVGasConfig()
			expected: 1006,
		},
	}

	for testName, test := range tests {
		t.Run(testName, func(t *testing.T) {
			// Construct a mock ABI to make function lookup possible.
			contractAbi := abi.ABI{
				Methods: map[string]abi.Method{
					"testMethod": {
						Name: "testMethod",
						ID:   []byte{0x1, 0x2, 0x3, 0x4},
						Type: abi.Function,
					},
				},
			}

			contract := NewContract(contractAbi, common.HexToAddress("0x1"), "")

			contract.RegisterMethods(test.method)

			// Construct an input whose first 4 bytes correspond to the method ID
			// and the subsequent 2 bytes are the method input arguments.
			input := []byte{0x1, 0x2, 0x3, 0x4, 0x5, 0x6}

			if actual := contract.RequiredGas(input); test.expected != actual {
				t.Errorf(
					"unexpected gas\n expected: %v\n actual:   %v",
					test.expected,
					actual,
				)
			}
		})
	}
}

func TestContract_RegisterMethods(t *testing.T) {
	contract := NewContract(abi.ABI{}, common.HexToAddress("0x1"), "")

	method1 := &mockMethod{methodName: "testMethod1", methodType: Write}
	method2Write := &mockMethod{methodName: "testMethod2", methodType: Write}
	method2Read := &mockMethod{methodName: "testMethod2", methodType: Read}

	// Register methods. As method2 is registered twice, the second
	// registration should overwrite the first one.
	contract.RegisterMethods(method1, method2Write, method2Read)

	expectedMethods := map[string]Method{
		"testMethod1": method1,
		"testMethod2": method2Read,
	}
	if !reflect.DeepEqual(expectedMethods, contract.methods) {
		t.Errorf(
			"unexpected methods\n expected: %v\n actual:   %v",
			expectedMethods,
			contract.methods,
		)
	}
}

func TestContract_Run(t *testing.T) {
	tests := map[string]struct {
		method         Method
		methodInputs   MethodInputs
		value          *uint256.Int
		readOnlyMode   bool
		expectedOutput []byte
		expectedError  error
	}{
		"happy path": {
			method: &mockMethod{
				methodName: "testMethod",
				methodType: Write,
				payable:    true,
				run: func(
					_ *RunContext,
					inputs MethodInputs,
				) (MethodOutputs, error) {
					if len(inputs) != 2 {
						return nil, fmt.Errorf("unexpected number of inputs")
					}

					inArg1, ok := inputs[0].(*big.Int)
					if !ok {
						return nil, fmt.Errorf("unexpected type of input 1")
					}

					inArg2, ok := inputs[1].([]byte)
					if !ok {
						return nil, fmt.Errorf("unexpected type of input 2")
					}

					outArg2 := new(big.Int).Add(inArg1, new(big.Int).SetBytes(inArg2))

					return []interface{}{outArg2}, nil
				},
			},
			methodInputs: []interface{}{
				big.NewInt(10),
				[]byte{0x7B}, // 123
			},
			value: uint256.NewInt(1000),
			// Sum of inputs (10 + 123 = 133) as hex, left-padded to 32 bytes.
			expectedOutput: append(make([]byte, 31), 0x85),
		},
		"read method with value": {
			method: &mockMethod{
				methodName: "testMethod",
				methodType: Read,
			},
			methodInputs: []interface{}{
				big.NewInt(10),
				[]byte{0x7B}, // 123
			},
			value:         uint256.NewInt(1000),
			expectedError: fmt.Errorf("read method cannot accept value"),
		},
		"write method with read-only mode": {
			method: &mockMethod{
				methodName: "testMethod",
				methodType: Write,
			},
			methodInputs: []interface{}{
				big.NewInt(10),
				[]byte{0x7B}, // 123
			},
			readOnlyMode:  true,
			expectedError: fmt.Errorf("write method cannot be executed in read-only mode"),
		},
		"non-payable write method with value": {
			method: &mockMethod{
				methodName: "testMethod",
				methodType: Write,
				payable:    false,
			},
			methodInputs: []interface{}{
				big.NewInt(10),
				[]byte{0x7B}, // 123
			},
			value:         uint256.NewInt(1000),
			expectedError: fmt.Errorf("non-payable write method cannot accept value"),
		},
	}

	for testName, test := range tests {
		t.Run(testName, func(t *testing.T) {
			uint256Type, err := abi.NewType("uint256", "uint256", nil)
			if err != nil {
				t.Fatal(err)
			}
			bytesType, err := abi.NewType("bytes", "bytes", nil)
			if err != nil {
				t.Fatal(err)
			}

			// Construct a mock ABI to make function lookup possible.
			methodAbi := abi.Method{
				Name: "testMethod",
				ID:   []byte{0x1, 0x2, 0x3, 0x4},
				Type: abi.Function,
				Inputs: []abi.Argument{
					{Name: "inArg1", Type: uint256Type},
					{Name: "inArg2", Type: bytesType},
				},
				Outputs: []abi.Argument{
					{Name: "outArg1", Type: uint256Type},
				},
			}
			contractAbi := abi.ABI{
				Methods: map[string]abi.Method{
					"testMethod": methodAbi,
				},
			}

			contract := NewContract(contractAbi, common.HexToAddress("0x1"), "")

			contract.RegisterMethods(test.method)

			methodInputArgs, err := methodAbi.Inputs.Pack(test.methodInputs...)
			if err != nil {
				t.Fatal(err)
			}

			// sdkCtx := sdk.Context{}
			sdkCtx := testutil.DefaultContext(
				storetypes.NewKVStoreKey(t.Name()+"_TestCacheContext"),
				storetypes.NewTransientStoreKey("transient_"+t.Name()))

			evm := &vm.EVM{
				StateDB: statedb.New(sdkCtx, statedb.NewMockKeeper(), statedb.TxConfig{}),
			}

			vmContract := vm.NewContract(&Contract{}, nil, test.value, 0)
			// Construct an input whose first 4 bytes correspond to the method ID
			// and the subsequent bytes are the method input arguments.
			vmContract.Input = append([]byte{0x1, 0x2, 0x3, 0x4}, methodInputArgs...)

			output, err := contract.Run(evm, vmContract, test.readOnlyMode)

			if !reflect.DeepEqual(test.expectedError, err) {
				t.Errorf(
					"unexpected error\n expected: %v\n actual:   %v",
					test.expectedError,
					err,
				)
			}

			if !reflect.DeepEqual(test.expectedOutput, output) {
				t.Errorf(
					"unexpected output\n expected: %v\n actual:   %v",
					test.expectedOutput,
					output,
				)
			}
		})
	}
}

func TestRunContext_MsgSender(t *testing.T) {
	caller := common.HexToAddress("0x1")
	contract := &vm.Contract{CallerAddress: caller}

	runContext := NewRunContext(sdk.Context{}, nil, contract, nil)

	if actualCaller := runContext.MsgSender(); caller != actualCaller {
		t.Errorf(
			"unexpected caller\n expected: %v\n actual:   %v",
			caller,
			actualCaller,
		)
	}
}

func TestRunContext_TxOrigin(t *testing.T) {
	origin := common.HexToAddress("0x1")
	evm := &vm.EVM{TxContext: vm.TxContext{Origin: origin}}

	runContext := NewRunContext(sdk.Context{}, evm, nil, nil)

	if actualOrigin := runContext.TxOrigin(); origin != actualOrigin {
		t.Errorf(
			"unexpected origin\n expected: %v\n actual:   %v",
			origin,
			actualOrigin,
		)
	}
}

func TestRunContext_MsgValue(t *testing.T) {
	value := uint256.NewInt(10)
	contract := vm.NewContract(&Contract{}, nil, value, 0)

	runContext := NewRunContext(sdk.Context{}, nil, contract, nil)

	if actualValue := runContext.MsgValue(); value.ToBig().Cmp(actualValue) != 0 {
		t.Errorf(
			"unexpected value\n expected: %v\n actual:   %v",
			value,
			actualValue,
		)
	}
}

func TestRunContext_IsMsgValue(t *testing.T) {
	tests := []struct {
		name     string
		value    *uint256.Int
		expected bool
	}{
		{
			name:     "value is greater than zero",
			value:    uint256.NewInt(10),
			expected: true,
		},
		{
			name:     "value is zero",
			value:    uint256.NewInt(0),
			expected: false,
		},
		{
			name:     "value is nil",
			value:    nil,
			expected: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			contract := vm.NewContract(&Contract{}, nil, test.value, 0)

			runContext := NewRunContext(sdk.Context{}, nil, contract, nil)

			if actual := runContext.IsMsgValue(); test.expected != actual {
				t.Errorf(
					"unexpected result\n expected: %v\n actual:   %v",
					test.expected,
					actual,
				)
			}
		})
	}
}

func TestRunContext_EventEmitter(t *testing.T) {
	eventEmitter := &EventEmitter{}

	runContext := NewRunContext(sdk.Context{}, nil, nil, eventEmitter)

	if actualEventEmitter := runContext.EventEmitter(); !reflect.DeepEqual(
		eventEmitter,
		actualEventEmitter,
	) {
		t.Errorf(
			"unexpected event emiiter\n expected: %v\n actual:   %v",
			eventEmitter,
			actualEventEmitter,
		)
	}
}
