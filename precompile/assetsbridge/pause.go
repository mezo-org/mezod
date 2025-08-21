package assetsbridge

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/mezo-org/mezod/precompile"
)

const (
	SetPauserMethodName      = "setPauser"
	GetPauserMethodName      = "getPauser"
	PauseBridgeOutMethodName = "pauseBridgeOut"
)

type SetPauserMethod struct {
	poaKeeper    PoaKeeper
	bridgeKeeper BridgeKeeper
}

func newSetPauserMethod(
	poaKeeper PoaKeeper,
	bridgeKeeper BridgeKeeper,
) *SetPauserMethod {
	return &SetPauserMethod{
		poaKeeper:    poaKeeper,
		bridgeKeeper: bridgeKeeper,
	}
}

func (m *SetPauserMethod) MethodName() string {
	return SetPauserMethodName
}

func (m *SetPauserMethod) MethodType() precompile.MethodType {
	return precompile.Write
}

func (m *SetPauserMethod) RequiredGas(_ []byte) (uint64, bool) {
	return 0, false
}

func (m *SetPauserMethod) Payable() bool {
	return false
}

func (m *SetPauserMethod) Run(
	context *precompile.RunContext,
	rawInputs precompile.MethodInputs,
) (precompile.MethodOutputs, error) {
	if err := precompile.ValidateMethodInputsCount(rawInputs, 1); err != nil {
		return nil, err
	}

	pauser, ok := rawInputs[0].(common.Address)
	if !ok {
		return nil, fmt.Errorf("invalid pauser address: %v", rawInputs[0])
	}

	if err := m.poaKeeper.CheckOwner(
		context.SdkCtx(),
		precompile.TypesConverter.Address.ToSDK(context.MsgSender()),
	); err != nil {
		return nil, err
	}

	m.bridgeKeeper.SetPauser(
		context.SdkCtx(),
		precompile.TypesConverter.Address.ToSDK(pauser),
	)

	return precompile.MethodOutputs{true}, nil
}

type GetPauserMethod struct {
	bridgeKeeper BridgeKeeper
}

func newGetPauserMethod(bridgeKeeper BridgeKeeper) *GetPauserMethod {
	return &GetPauserMethod{
		bridgeKeeper: bridgeKeeper,
	}
}

func (m *GetPauserMethod) MethodName() string {
	return GetPauserMethodName
}

func (m *GetPauserMethod) MethodType() precompile.MethodType {
	return precompile.Read
}

func (m *GetPauserMethod) RequiredGas(_ []byte) (uint64, bool) {
	return 0, false
}

func (m *GetPauserMethod) Payable() bool {
	return false
}

func (m *GetPauserMethod) Run(
	context *precompile.RunContext,
	rawInputs precompile.MethodInputs,
) (precompile.MethodOutputs, error) {
	if err := precompile.ValidateMethodInputsCount(rawInputs, 0); err != nil {
		return nil, err
	}

	pauser := m.bridgeKeeper.GetPauser(context.SdkCtx())

	return precompile.MethodOutputs{precompile.TypesConverter.Address.FromSDK(pauser)}, nil
}

type PauseBridgeOutMethod struct {
	bridgeKeeper BridgeKeeper
}

func newPauseBridgeOutMethod(bridgeKeeper BridgeKeeper) *PauseBridgeOutMethod {
	return &PauseBridgeOutMethod{
		bridgeKeeper: bridgeKeeper,
	}
}

func (m *PauseBridgeOutMethod) MethodName() string {
	return PauseBridgeOutMethodName
}

func (m *PauseBridgeOutMethod) MethodType() precompile.MethodType {
	return precompile.Write
}

func (m *PauseBridgeOutMethod) RequiredGas(_ []byte) (uint64, bool) {
	return 0, false
}

func (m *PauseBridgeOutMethod) Payable() bool {
	return false
}

func (m *PauseBridgeOutMethod) Run(
	context *precompile.RunContext,
	rawInputs precompile.MethodInputs,
) (precompile.MethodOutputs, error) {
	if err := precompile.ValidateMethodInputsCount(rawInputs, 0); err != nil {
		return nil, err
	}

	err := m.bridgeKeeper.PauseBridgeOut(
		context.SdkCtx(),
		precompile.TypesConverter.Address.ToSDK(context.MsgSender()),
	)
	if err != nil {
		return nil, err
	}

	return precompile.MethodOutputs{true}, nil
}
