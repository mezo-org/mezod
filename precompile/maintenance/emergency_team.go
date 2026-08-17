package maintenance

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/mezo-org/mezod/precompile"
	"github.com/mezo-org/mezod/x/evm/statedb"
)

// SetEmergencyTeamMethodName is the name of the setEmergencyTeam method. It
// matches the name of the method in the contract ABI.
const SetEmergencyTeamMethodName = "setEmergencyTeam"

// EmergencyTeamSetEventName is the name of the EmergencyTeamSet event. It
// matches the name of the event in the contract ABI.
const EmergencyTeamSetEventName = "EmergencyTeamSet"

// setEmergencyTeamMethod grants or revokes the emergency team role.
type setEmergencyTeamMethod struct {
	poaKeeper PoaKeeper
}

func newSetEmergencyTeamMethod(poaKeeper PoaKeeper) *setEmergencyTeamMethod {
	return &setEmergencyTeamMethod{poaKeeper: poaKeeper}
}

func (m *setEmergencyTeamMethod) MethodName() string {
	return SetEmergencyTeamMethodName
}

func (m *setEmergencyTeamMethod) MethodType() precompile.MethodType {
	return precompile.Write
}

func (m *setEmergencyTeamMethod) RequiredGas(_ []byte) (uint64, bool) {
	return 0, false
}

func (m *setEmergencyTeamMethod) Payable() bool {
	return false
}

func (m *setEmergencyTeamMethod) Run(
	context *precompile.RunContext,
	inputs precompile.MethodInputs,
) (precompile.MethodOutputs, []statedb.StateChange, error) {
	if err := precompile.ValidateMethodInputsCount(inputs, 1); err != nil {
		return nil, nil, err
	}

	team, ok := inputs[0].(common.Address)
	if !ok {
		return nil, nil, fmt.Errorf("team argument must be an address")
	}

	previous := m.poaKeeper.GetEmergencyTeam(context.SdkCtx())

	// This method is restricted to the PoA owner.
	err := m.poaKeeper.SetEmergencyTeam(
		context.SdkCtx(),
		precompile.TypesConverter.Address.ToSDK(context.MsgSender()),
		precompile.TypesConverter.Address.ToSDK(team),
	)
	if err != nil {
		return nil, nil, err
	}

	err = context.EventEmitter().Emit(
		NewEmergencyTeamSetEvent(
			precompile.TypesConverter.Address.FromSDK(previous),
			team,
		),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to emit EmergencyTeamSet event: [%w]", err)
	}

	return precompile.MethodOutputs{true}, nil, nil
}

// GetEmergencyTeamMethodName is the name of the getEmergencyTeam method. It
// matches the name of the method in the contract ABI.
const GetEmergencyTeamMethodName = "getEmergencyTeam"

// getEmergencyTeamMethod returns the current emergency team address.
type getEmergencyTeamMethod struct {
	poaKeeper PoaKeeper
}

func newGetEmergencyTeamMethod(poaKeeper PoaKeeper) *getEmergencyTeamMethod {
	return &getEmergencyTeamMethod{poaKeeper: poaKeeper}
}

func (m *getEmergencyTeamMethod) MethodName() string {
	return GetEmergencyTeamMethodName
}

func (m *getEmergencyTeamMethod) MethodType() precompile.MethodType {
	return precompile.Read
}

func (m *getEmergencyTeamMethod) RequiredGas(_ []byte) (uint64, bool) {
	return 0, false
}

func (m *getEmergencyTeamMethod) Payable() bool {
	return false
}

func (m *getEmergencyTeamMethod) Run(
	context *precompile.RunContext,
	inputs precompile.MethodInputs,
) (precompile.MethodOutputs, []statedb.StateChange, error) {
	if err := precompile.ValidateMethodInputsCount(inputs, 0); err != nil {
		return nil, nil, err
	}

	team := m.poaKeeper.GetEmergencyTeam(context.SdkCtx())

	return precompile.MethodOutputs{
		precompile.TypesConverter.Address.FromSDK(team),
	}, nil, nil
}

// EmergencyTeamSetEvent is emitted when the emergency team role is granted,
// rotated, or revoked. The zero address means the role is not granted.
type EmergencyTeamSetEvent struct {
	previous common.Address
	current  common.Address
}

func NewEmergencyTeamSetEvent(previous, current common.Address) *EmergencyTeamSetEvent {
	return &EmergencyTeamSetEvent{
		previous: previous,
		current:  current,
	}
}

func (e *EmergencyTeamSetEvent) EventName() string {
	return EmergencyTeamSetEventName
}

func (e *EmergencyTeamSetEvent) Arguments() []*precompile.EventArgument {
	return []*precompile.EventArgument{
		{Indexed: true, Value: e.previous},
		{Indexed: true, Value: e.current},
	}
}
