package assetsbridge

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/mezo-org/mezod/precompile"
)

// GetSourceBTCTokenMethodName is the name of the getSourceBTCToken method.
// It matches the name of the method in the contract ABI.
//
//nolint:gosec
const GetSourceBTCTokenMethodName = "getSourceBTCToken"

// GetSourceBTCTokenMethod is the implementation of the getSourceBTCToken method.
type GetSourceBTCTokenMethod struct {
	bridgeKeeper BridgeKeeper
}

func newGetSourceBTCTokenMethod(bridgeKeeper BridgeKeeper) *GetSourceBTCTokenMethod {
	return &GetSourceBTCTokenMethod{
		bridgeKeeper: bridgeKeeper,
	}
}

func (m *GetSourceBTCTokenMethod) MethodName() string {
	return GetSourceBTCTokenMethodName
}

func (m *GetSourceBTCTokenMethod) MethodType() precompile.MethodType {
	return precompile.Read
}

func (m *GetSourceBTCTokenMethod) RequiredGas(_ []byte) (uint64, bool) {
	// Fallback to the default gas calculation.
	return 0, false
}

func (m *GetSourceBTCTokenMethod) Payable() bool {
	return false
}

func (m *GetSourceBTCTokenMethod) Run(
	context *precompile.RunContext,
	inputs precompile.MethodInputs,
) (precompile.MethodOutputs, error) {
	if err := precompile.ValidateMethodInputsCount(inputs, 0); err != nil {
		return nil, err
	}

	sourceBTCToken := m.bridgeKeeper.GetSourceBTCToken(context.SdkCtx())

	return precompile.MethodOutputs{common.HexToAddress(sourceBTCToken)}, nil
}
