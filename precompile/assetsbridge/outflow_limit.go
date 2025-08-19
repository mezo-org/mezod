package assetsbridge

import (
	"errors"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/mezo-org/mezod/precompile"
)

const (
	SetOutflowLimitMethodName = "setOutflowLimit"
	GetOutflowLimitMethodName = "getOutflowLimit"
)

type SetOutflowLimitMethod struct {
	poaKeeper    PoaKeeper
	bridgeKeeper BridgeKeeper
}

func newSetOutflowLimitMethod(
	poaKeeper PoaKeeper,
	bridgeKeeper BridgeKeeper,
) *SetOutflowLimitMethod {
	return &SetOutflowLimitMethod{
		poaKeeper:    poaKeeper,
		bridgeKeeper: bridgeKeeper,
	}
}

func (m *SetOutflowLimitMethod) MethodName() string {
	return SetOutflowLimitMethodName
}

func (m *SetOutflowLimitMethod) MethodType() precompile.MethodType {
	return precompile.Write
}

func (m *SetOutflowLimitMethod) RequiredGas(_ []byte) (uint64, bool) {
	return 0, false
}

func (m *SetOutflowLimitMethod) Payable() bool {
	return false
}

func (m *SetOutflowLimitMethod) Run(
	context *precompile.RunContext,
	rawInputs precompile.MethodInputs,
) (precompile.MethodOutputs, error) {
	if err := precompile.ValidateMethodInputsCount(rawInputs, 2); err != nil {
		return nil, err
	}

	token, ok := rawInputs[0].(common.Address)
	if !ok {
		return nil, fmt.Errorf("invalid token address: %v", rawInputs[0])
	}

	limit, ok := rawInputs[1].(*big.Int)
	if !ok {
		return nil, fmt.Errorf("invalid limit: %v", rawInputs[1])
	}

	if err := m.poaKeeper.CheckOwner(
		context.SdkCtx(),
		precompile.TypesConverter.Address.ToSDK(context.MsgSender()),
	); err != nil {
		return nil, err
	}

	if limit.Sign() < 0 {
		return nil, errors.New("limit must be non-negative")
	}

	sdkLimit, err := precompile.TypesConverter.BigInt.ToSDK(limit)
	if err != nil {
		return nil, fmt.Errorf("failed to convert limit: [%w]", err)
	}

	m.bridgeKeeper.SetOutflowLimit(
		context.SdkCtx(),
		token.Bytes(),
		sdkLimit,
	)

	return precompile.MethodOutputs{true}, nil
}

type GetOutflowLimitMethod struct {
	bridgeKeeper BridgeKeeper
}

func newGetOutflowLimitMethod(
	bridgeKeeper BridgeKeeper,
) *GetOutflowLimitMethod {
	return &GetOutflowLimitMethod{
		bridgeKeeper: bridgeKeeper,
	}
}

func (m *GetOutflowLimitMethod) MethodName() string {
	return GetOutflowLimitMethodName
}

func (m *GetOutflowLimitMethod) MethodType() precompile.MethodType {
	return precompile.Read
}

func (m *GetOutflowLimitMethod) RequiredGas(_ []byte) (uint64, bool) {
	return 0, false
}

func (m *GetOutflowLimitMethod) Payable() bool {
	return false
}

func (m *GetOutflowLimitMethod) Run(
	context *precompile.RunContext,
	rawInputs precompile.MethodInputs,
) (precompile.MethodOutputs, error) {
	if err := precompile.ValidateMethodInputsCount(rawInputs, 1); err != nil {
		return nil, err
	}

	token, ok := rawInputs[0].(common.Address)
	if !ok {
		return nil, fmt.Errorf("invalid token address: %v", rawInputs[0])
	}

	limit := m.bridgeKeeper.GetOutflowLimit(
		context.SdkCtx(),
		token.Bytes(),
	)

	return precompile.MethodOutputs{limit.BigInt()}, nil
}
