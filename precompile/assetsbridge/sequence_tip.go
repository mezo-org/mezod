package assetsbridge

import (
	"github.com/mezo-org/mezod/precompile"
	"github.com/mezo-org/mezod/x/evm/statedb"
)

// GetCurrentSequenceTipMethodName is the name of the getCurrentSequenceTip method.
// It matches the name of the method in the contract ABI.
//
//nolint:gosec
const GetCurrentSequenceTipMethodName = "getCurrentSequenceTip"

// GetCurrentSequenceTipMethod is the implementation of the getCurrentSequenceTip method.
type GetCurrentSequenceTipMethod struct {
	bridgeKeeper BridgeKeeper
}

func newGetCurrentSequenceTipMethod(bridgeKeeper BridgeKeeper) *GetCurrentSequenceTipMethod {
	return &GetCurrentSequenceTipMethod{
		bridgeKeeper: bridgeKeeper,
	}
}

func (m *GetCurrentSequenceTipMethod) MethodName() string {
	return GetCurrentSequenceTipMethodName
}

func (m *GetCurrentSequenceTipMethod) MethodType() precompile.MethodType {
	return precompile.Read
}

func (m *GetCurrentSequenceTipMethod) RequiredGas(_ []byte) (uint64, bool) {
	// Fallback to the default gas calculation.
	return 0, false
}

func (m *GetCurrentSequenceTipMethod) Payable() bool {
	return false
}

func (m *GetCurrentSequenceTipMethod) Run(
	context *precompile.RunContext,
	inputs precompile.MethodInputs,
) (precompile.MethodOutputs, []statedb.StateChange, error) {
	if err := precompile.ValidateMethodInputsCount(inputs, 0); err != nil {
		return nil, nil, err
	}

	return precompile.MethodOutputs{
		precompile.TypesConverter.BigInt.FromSDK(
			m.bridgeKeeper.GetAssetsLockedSequenceTip(context.SdkCtx()),
		),
	}, nil, nil
}
