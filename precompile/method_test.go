package precompile

import (
	"testing"

	store "cosmossdk.io/store/types"
)

func TestDefaultRequiredGas(t *testing.T) {
	gasConfig := store.GasConfig{
		ReadCostFlat:     10,
		ReadCostPerByte:  2,
		WriteCostFlat:    20,
		WriteCostPerByte: 4,
	}

	tests := map[string]struct {
		methodType      MethodType
		methodInputArgs []byte
		expectedGas     uint64
	}{
		"read method": {
			methodType:      Read,
			methodInputArgs: []byte{0x1, 0x2},
			// readCostFlat + (readCostPerByte * len(methodInputArgs)) = 10 + 2 * 2 = 14
			expectedGas: 14,
		},
		"write method": {
			methodType:      Write,
			methodInputArgs: []byte{0x1, 0x2},
			// writeCostFlat + (writeCostPerByte * len(methodInputArgs)) = 20 + 4 * 2 = 28
			expectedGas: 28,
		},
		"empty input": {
			methodType:      Write,
			methodInputArgs: []byte{},
			// writeCostFlat = 20
			expectedGas: 20,
		},
	}

	for testName, test := range tests {
		t.Run(testName, func(t *testing.T) {
			gas := DefaultRequiredGas(gasConfig, test.methodType, test.methodInputArgs)

			if test.expectedGas != gas {
				t.Errorf(
					"unexpected gas\nexpected: %d\nactual:   %d",
					test.expectedGas,
					gas,
				)
			}
		})
	}
}

type mockMethod struct {
	methodName  string
	methodType  MethodType
	requiredGas uint64
	payable     bool

	run func(
		context *RunContext,
		inputs MethodInputs,
	) (MethodOutputs, error)
}

func (mm *mockMethod) MethodName() string {
	return mm.methodName
}

func (mm *mockMethod) MethodType() MethodType {
	return mm.methodType
}

func (mm *mockMethod) RequiredGas(_ []byte) (uint64, bool) {
	if mm.requiredGas == 0 {
		return 0, false
	}

	return mm.requiredGas, true
}

func (mm *mockMethod) Payable() bool {
	return mm.payable
}

func (mm *mockMethod) Run(
	context *RunContext,
	inputs MethodInputs,
) (MethodOutputs, error) {
	return mm.run(context, inputs)
}
