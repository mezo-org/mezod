package maintenance

import (
	"fmt"

	"github.com/mezo-org/mezod/precompile"
	"github.com/mezo-org/mezod/x/evm/statedb"
)

// SetBridgeLockdownMethodName is the name of the setBridgeLockdown method. It
// matches the name of the method in the contract ABI.
const SetBridgeLockdownMethodName = "setBridgeLockdown"

// BridgeLockdownSetEventName is the name of the BridgeLockdownSet event. It
// matches the name of the event in the contract ABI.
const BridgeLockdownSetEventName = "BridgeLockdownSet"

// setBridgeLockdownMethod enables or disables the bridge lockdown, per
// direction. Passing false for both directions disables the lockdown.
type setBridgeLockdownMethod struct {
	poaKeeper    PoaKeeper
	bridgeKeeper BridgeKeeper
}

func newSetBridgeLockdownMethod(
	poaKeeper PoaKeeper,
	bridgeKeeper BridgeKeeper,
) *setBridgeLockdownMethod {
	return &setBridgeLockdownMethod{
		poaKeeper:    poaKeeper,
		bridgeKeeper: bridgeKeeper,
	}
}

func (m *setBridgeLockdownMethod) MethodName() string {
	return SetBridgeLockdownMethodName
}

func (m *setBridgeLockdownMethod) MethodType() precompile.MethodType {
	return precompile.Write
}

func (m *setBridgeLockdownMethod) RequiredGas(_ []byte) (uint64, bool) {
	return 0, false
}

func (m *setBridgeLockdownMethod) Payable() bool {
	return false
}

func (m *setBridgeLockdownMethod) Run(
	context *precompile.RunContext,
	inputs precompile.MethodInputs,
) (precompile.MethodOutputs, []statedb.StateChange, error) {
	if err := precompile.ValidateMethodInputsCount(inputs, 2); err != nil {
		return nil, nil, err
	}

	bridgeIn, ok := inputs[0].(bool)
	if !ok {
		return nil, nil, fmt.Errorf("bridgeIn argument must be a boolean")
	}

	bridgeOut, ok := inputs[1].(bool)
	if !ok {
		return nil, nil, fmt.Errorf("bridgeOut argument must be a boolean")
	}

	// This method is restricted to the PoA owner and the emergency team.
	err := m.poaKeeper.CheckOwnerOrEmergencyTeam(
		context.SdkCtx(),
		precompile.TypesConverter.Address.ToSDK(context.MsgSender()),
	)
	if err != nil {
		return nil, nil, err
	}

	m.bridgeKeeper.SetBridgeInPaused(context.SdkCtx(), bridgeIn)
	m.bridgeKeeper.SetBridgeOutPaused(context.SdkCtx(), bridgeOut)

	err = context.EventEmitter().Emit(
		NewBridgeLockdownSetEvent(bridgeIn, bridgeOut),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to emit BridgeLockdownSet event: [%w]", err)
	}

	return precompile.MethodOutputs{true}, nil, nil
}

// GetBridgeLockdownMethodName is the name of the getBridgeLockdown method. It
// matches the name of the method in the contract ABI.
const GetBridgeLockdownMethodName = "getBridgeLockdown"

// getBridgeLockdownMethod reports the bridge lockdown state per direction.
type getBridgeLockdownMethod struct {
	bridgeKeeper BridgeKeeper
}

func newGetBridgeLockdownMethod(bridgeKeeper BridgeKeeper) *getBridgeLockdownMethod {
	return &getBridgeLockdownMethod{bridgeKeeper: bridgeKeeper}
}

func (m *getBridgeLockdownMethod) MethodName() string {
	return GetBridgeLockdownMethodName
}

func (m *getBridgeLockdownMethod) MethodType() precompile.MethodType {
	return precompile.Read
}

func (m *getBridgeLockdownMethod) RequiredGas(_ []byte) (uint64, bool) {
	return 0, false
}

func (m *getBridgeLockdownMethod) Payable() bool {
	return false
}

func (m *getBridgeLockdownMethod) Run(
	context *precompile.RunContext,
	inputs precompile.MethodInputs,
) (precompile.MethodOutputs, []statedb.StateChange, error) {
	if err := precompile.ValidateMethodInputsCount(inputs, 0); err != nil {
		return nil, nil, err
	}

	return precompile.MethodOutputs{
		m.bridgeKeeper.IsBridgeInPaused(context.SdkCtx()),
		m.bridgeKeeper.IsBridgeOutPaused(context.SdkCtx()),
	}, nil, nil
}

// BridgeLockdownSetEvent is emitted when the bridge lockdown is enabled or
// disabled.
type BridgeLockdownSetEvent struct {
	bridgeIn  bool
	bridgeOut bool
}

func NewBridgeLockdownSetEvent(bridgeIn, bridgeOut bool) *BridgeLockdownSetEvent {
	return &BridgeLockdownSetEvent{
		bridgeIn:  bridgeIn,
		bridgeOut: bridgeOut,
	}
}

func (e *BridgeLockdownSetEvent) EventName() string {
	return BridgeLockdownSetEventName
}

func (e *BridgeLockdownSetEvent) Arguments() []*precompile.EventArgument {
	return []*precompile.EventArgument{
		{Indexed: false, Value: e.bridgeIn},
		{Indexed: false, Value: e.bridgeOut},
	}
}
