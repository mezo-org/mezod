package maintenance

import (
	"fmt"

	upgradetypes "cosmossdk.io/x/upgrade/types"
	"github.com/mezo-org/mezod/precompile"
	"github.com/mezo-org/mezod/x/evm/statedb"
)

// SetChainLockdownMethodName is the name of the setChainLockdown method. It
// matches the name of the method in the contract ABI.
const SetChainLockdownMethodName = "setChainLockdown"

// ChainLockdownSetEventName is the name of the ChainLockdownSet event. It
// matches the name of the event in the contract ABI.
const ChainLockdownSetEventName = "ChainLockdownSet"

// setChainLockdownMethod halts the chain. It schedules an upgrade plan under
// the given name, at the next block height. The name must have no registered
// handler, so the x/upgrade PreBlocker panics on every validator. The chain
// cannot recover on chain.
type setChainLockdownMethod struct {
	poaKeeper     PoaKeeper
	upgradeKeeper UpgradeKeeper
}

func newSetChainLockdownMethod(
	poaKeeper PoaKeeper,
	upgradeKeeper UpgradeKeeper,
) *setChainLockdownMethod {
	return &setChainLockdownMethod{
		poaKeeper:     poaKeeper,
		upgradeKeeper: upgradeKeeper,
	}
}

func (m *setChainLockdownMethod) MethodName() string {
	return SetChainLockdownMethodName
}

func (m *setChainLockdownMethod) MethodType() precompile.MethodType {
	return precompile.Write
}

func (m *setChainLockdownMethod) RequiredGas(_ []byte) (uint64, bool) {
	return 0, false
}

func (m *setChainLockdownMethod) Payable() bool {
	return false
}

func (m *setChainLockdownMethod) Run(
	context *precompile.RunContext,
	inputs precompile.MethodInputs,
) (precompile.MethodOutputs, []statedb.StateChange, error) {
	if err := precompile.ValidateMethodInputsCount(inputs, 1); err != nil {
		return nil, nil, err
	}

	planName, ok := inputs[0].(string)
	if !ok {
		return nil, nil, fmt.Errorf("planName argument must be a string")
	}

	// This method is restricted to the PoA owner and the emergency team.
	err := m.poaKeeper.CheckOwnerOrEmergencyTeam(
		context.SdkCtx(),
		precompile.TypesConverter.Address.ToSDK(context.MsgSender()),
	)
	if err != nil {
		return nil, nil, err
	}

	if planName == "" {
		return nil, nil, fmt.Errorf("plan name cannot be empty")
	}

	// A plan whose name has a registered handler upgrades the chain instead
	// of halting it.
	if m.upgradeKeeper.HasHandler(planName) {
		return nil, nil, fmt.Errorf(
			"the binary has an upgrade handler for plan name %q; the chain would not halt",
			planName,
		)
	}

	doneHeight, err := m.upgradeKeeper.GetDoneHeight(context.SdkCtx(), planName)
	if err != nil {
		return nil, nil, err
	}
	if doneHeight != 0 {
		return nil, nil, fmt.Errorf(
			"an upgrade with plan name %q was already completed",
			planName,
		)
	}

	// No handler is registered for the plan name. The x/upgrade PreBlocker
	// panics at this height and every validator stops.
	plan := upgradetypes.Plan{
		Name:   planName,
		Height: context.SdkCtx().BlockHeight() + 1,
	}

	err = m.upgradeKeeper.ScheduleUpgrade(context.SdkCtx(), plan)
	if err != nil {
		return nil, nil, err
	}

	err = context.EventEmitter().Emit(
		NewChainLockdownSetEvent(planName),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to emit ChainLockdownSet event: [%w]", err)
	}

	return precompile.MethodOutputs{true}, nil, nil
}

// ChainLockdownSetEvent is emitted when the chain lockdown is set. It carries
// the name of the scheduled upgrade plan that halts the chain.
type ChainLockdownSetEvent struct {
	name string
}

func NewChainLockdownSetEvent(name string) *ChainLockdownSetEvent {
	return &ChainLockdownSetEvent{
		name: name,
	}
}

func (e *ChainLockdownSetEvent) EventName() string {
	return ChainLockdownSetEventName
}

func (e *ChainLockdownSetEvent) Arguments() []*precompile.EventArgument {
	return []*precompile.EventArgument{
		{Indexed: false, Value: e.name},
	}
}
