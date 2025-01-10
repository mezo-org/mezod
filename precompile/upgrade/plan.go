package upgrade

import (
	"fmt"
	"strconv"

	upgradetypes "cosmossdk.io/x/upgrade/types"
	"github.com/mezo-org/mezod/precompile"
)

// PlanMethodName is the name of the plan method. It matches the name
// of the method in the contract ABI.
const PlanMethodName = "plan"

// PlanMethod is the implementation of the plan method that returns
// the current upgrade plan
type PlanMethod struct {
	upgradeKeeper UpgradeKeeper
}

func newPlanMethod(upgradeKeeper UpgradeKeeper) *PlanMethod {
	return &PlanMethod{
		upgradeKeeper: upgradeKeeper,
	}
}

func (m *PlanMethod) MethodName() string {
	return PlanMethodName
}

func (m *PlanMethod) MethodType() precompile.MethodType {
	return precompile.Read
}

func (m *PlanMethod) RequiredGas(_ []byte) (uint64, bool) {
	// Fallback to the default gas calculation.
	return 0, false
}

func (m *PlanMethod) Payable() bool {
	return false
}

func (m *PlanMethod) Run(context *precompile.RunContext, inputs precompile.MethodInputs) (precompile.MethodOutputs, error) {
	if err := precompile.ValidateMethodInputsCount(inputs, 0); err != nil {
		return nil, err
	}

	plan, err := m.upgradeKeeper.GetUpgradePlan(context.SdkCtx())
	if err != nil {
		return nil, err
	}

	return precompile.MethodOutputs{plan.Name, strconv.FormatInt(plan.Height, 10), plan.Info}, nil
}

// SubmitPlanMethodName is the name of the submitPlan method. It matches the name
// of the method in the contract ABI.
const SubmitPlanMethodName = "submitPlan"

type SubmitPlanMethod struct {
	upgradeKeeper UpgradeKeeper
	poaKeeper     PoaKeeper
}

func newSubmitPlanMethod(upgradeKeeper UpgradeKeeper, poaKeeper PoaKeeper) *SubmitPlanMethod {
	return &SubmitPlanMethod{
		upgradeKeeper: upgradeKeeper,
		poaKeeper:     poaKeeper,
	}
}

func (m *SubmitPlanMethod) MethodName() string {
	return SubmitPlanMethodName
}

func (m *SubmitPlanMethod) MethodType() precompile.MethodType {
	return precompile.Write
}

func (m *SubmitPlanMethod) RequiredGas(_ []byte) (uint64, bool) {
	// Fallback to the default gas calculation.
	return 0, false
}

func (m *SubmitPlanMethod) Payable() bool {
	return false
}

func (m *SubmitPlanMethod) Run(context *precompile.RunContext, inputs precompile.MethodInputs) (precompile.MethodOutputs, error) {
	if err := precompile.ValidateMethodInputsCount(inputs, 3); err != nil {
		return nil, err
	}

	err := m.poaKeeper.CheckOwner(
		context.SdkCtx(),
		precompile.TypesConverter.Address.ToSDK(context.MsgSender()),
	)
	if err != nil {
		return nil, err
	}

	height, err := strconv.Atoi(inputs[1].(string))
	if err != nil {
		return nil, err
	}

	plan := upgradetypes.Plan{
		Name:   inputs[0].(string),
		Height: int64(height),
		Info:   inputs[2].(string),
	}

	err = m.upgradeKeeper.ScheduleUpgrade(context.SdkCtx(), plan)
	if err != nil {
		return nil, err
	}

	// emit event
	err = context.EventEmitter().Emit(
		NewPlanSubmittedEvent(plan.Name),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to emit PlanSubmitted event: [%w]", err)
	}

	return precompile.MethodOutputs{true}, nil
}

// PlanSubmittedName is the name of the PlanSubmitted event. It matches the name
// of the event in the contract ABI.
const PlanSubmittedEventName = "PlanSubmitted"

// PlanSubmittedEvent is the implementation of the PlanSubmitted event that contains
// the following arguments:
// - name: is the name of the submitted upgrade plan
type PlanSubmittedEvent struct {
	name string
}

func NewPlanSubmittedEvent(name string) *PlanSubmittedEvent {
	return &PlanSubmittedEvent{
		name: name,
	}
}

func (e *PlanSubmittedEvent) EventName() string {
	return PlanSubmittedEventName
}

func (e *PlanSubmittedEvent) Arguments() []*precompile.EventArgument {
	return []*precompile.EventArgument{
		{
			Indexed: false,
			Value:   e.name,
		},
	}
}

// CancelPlanMethodName is the name of the cancelPlan method. It matches the name
// of the method in the contract ABI.
const CancelPlanMethodName = "cancelPlan"

type CancelPlanMethod struct {
	upgradeKeeper UpgradeKeeper
	poaKeeper     PoaKeeper
}

func newCancelPlanMethod(upgradeKeeper UpgradeKeeper, poaKeeper PoaKeeper) *CancelPlanMethod {
	return &CancelPlanMethod{
		upgradeKeeper: upgradeKeeper,
		poaKeeper:     poaKeeper,
	}
}

func (m *CancelPlanMethod) MethodName() string {
	return CancelPlanMethodName
}

func (m *CancelPlanMethod) MethodType() precompile.MethodType {
	return precompile.Write
}

func (m *CancelPlanMethod) RequiredGas(_ []byte) (uint64, bool) {
	// Fallback to the default gas calculation.
	return 0, false
}

func (m *CancelPlanMethod) Payable() bool {
	return false
}

func (m *CancelPlanMethod) Run(context *precompile.RunContext, inputs precompile.MethodInputs) (precompile.MethodOutputs, error) {
	if err := precompile.ValidateMethodInputsCount(inputs, 0); err != nil {
		return nil, err
	}
	err := m.poaKeeper.CheckOwner(
		context.SdkCtx(),
		precompile.TypesConverter.Address.ToSDK(context.MsgSender()),
	)
	if err != nil {
		return nil, err
	}

	plan, err := m.upgradeKeeper.GetUpgradePlan(context.SdkCtx())
	if err != nil {
		return nil, err
	}

	err = m.upgradeKeeper.ClearUpgradePlan(context.SdkCtx())
	if err != nil {
		return nil, err
	}

	// emit event
	err = context.EventEmitter().Emit(
		NewPlanCanceledEvent(plan.Name),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to emit PlanCanceled event: [%w]", err)
	}

	return precompile.MethodOutputs{true}, nil
}

// PlanCanceledName is the name of the PlanCanceled event. It matches the name
// of the event in the contract ABI.
const PlanCanceledEventName = "PlanCanceled"

// PlanCanceledEvent is the implementation of the PlanCanceled event that contains
// the following arguments:
// - name: is the name of the canceled upgrade plan
type PlanCanceledEvent struct {
	name string
}

func NewPlanCanceledEvent(name string) *PlanCanceledEvent {
	return &PlanCanceledEvent{
		name: name,
	}
}

func (e *PlanCanceledEvent) EventName() string {
	return PlanCanceledEventName
}

func (e *PlanCanceledEvent) Arguments() []*precompile.EventArgument {
	return []*precompile.EventArgument{
		{
			Indexed: false,
			Value:   e.name,
		},
	}
}
