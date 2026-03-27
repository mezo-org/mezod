package assetsbridge

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/mezo-org/mezod/precompile"
	"github.com/mezo-org/mezod/x/evm/statedb"
	evmtypes "github.com/mezo-org/mezod/x/evm/types"
)

const (
	BridgeTripartyMethodName              = "bridgeTriparty"
	AllowTripartyControllerMethodName     = "allowTripartyController"
	IsAllowedTripartyControllerMethodName = "isAllowedTripartyController"
	PauseTripartyMethodName               = "pauseTriparty"
	SetTripartyBlockDelayMethodName       = "setTripartyBlockDelay"
)

// --- bridgeTriparty ---

type BridgeTripartyMethod struct {
	bridgeKeeper BridgeKeeper
}

func newBridgeTripartyMethod(bridgeKeeper BridgeKeeper) *BridgeTripartyMethod {
	return &BridgeTripartyMethod{bridgeKeeper: bridgeKeeper}
}

func (m *BridgeTripartyMethod) MethodName() string {
	return BridgeTripartyMethodName
}

func (m *BridgeTripartyMethod) MethodType() precompile.MethodType {
	return precompile.Write
}

func (m *BridgeTripartyMethod) RequiredGas(_ []byte) (uint64, bool) {
	return 0, false
}

func (m *BridgeTripartyMethod) Payable() bool {
	return false
}

func (m *BridgeTripartyMethod) Run(
	context *precompile.RunContext,
	rawInputs precompile.MethodInputs,
) (precompile.MethodOutputs, []statedb.StateChange, error) {
	if err := precompile.ValidateMethodInputsCount(rawInputs, 2); err != nil {
		return nil, nil, err
	}

	recipient, ok := rawInputs[0].(common.Address)
	if !ok {
		return nil, nil, fmt.Errorf("invalid recipient address: %v", rawInputs[0])
	}

	if recipient == (common.Address{}) {
		return nil, nil, fmt.Errorf("recipient address must not be the zero address")
	}

	amount, ok := rawInputs[1].(*big.Int)
	if !ok {
		return nil, nil, fmt.Errorf("invalid amount: %v", rawInputs[1])
	}

	if amount == nil || amount.Sign() <= 0 {
		return nil, nil, fmt.Errorf("amount must be positive")
	}

	sdkCtx := context.SdkCtx()

	if m.bridgeKeeper.IsTripartyPaused(sdkCtx) {
		return nil, nil, fmt.Errorf("triparty bridging is paused")
	}

	sender := precompile.TypesConverter.Address.ToSDK(context.MsgSender())
	if !m.bridgeKeeper.IsAllowedTripartyController(sdkCtx, sender) {
		return nil, nil, fmt.Errorf("caller is not an allowed triparty controller")
	}

	err := context.EventEmitter().Emit(
		NewTripartyBridgeRequestedEvent(recipient, amount, context.MsgSender()),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to emit TripartyBridgeRequested event: [%w]", err)
	}

	return precompile.MethodOutputs{true}, nil, nil
}

// --- allowTripartyController ---

type AllowTripartyControllerMethod struct {
	poaKeeper    PoaKeeper
	bridgeKeeper BridgeKeeper
}

func newAllowTripartyControllerMethod(
	poaKeeper PoaKeeper,
	bridgeKeeper BridgeKeeper,
) *AllowTripartyControllerMethod {
	return &AllowTripartyControllerMethod{
		poaKeeper:    poaKeeper,
		bridgeKeeper: bridgeKeeper,
	}
}

func (m *AllowTripartyControllerMethod) MethodName() string {
	return AllowTripartyControllerMethodName
}

func (m *AllowTripartyControllerMethod) MethodType() precompile.MethodType {
	return precompile.Write
}

func (m *AllowTripartyControllerMethod) RequiredGas(_ []byte) (uint64, bool) {
	return 0, false
}

func (m *AllowTripartyControllerMethod) Payable() bool {
	return false
}

func (m *AllowTripartyControllerMethod) Run(
	context *precompile.RunContext,
	rawInputs precompile.MethodInputs,
) (precompile.MethodOutputs, []statedb.StateChange, error) {
	if err := precompile.ValidateMethodInputsCount(rawInputs, 2); err != nil {
		return nil, nil, err
	}

	controller, ok := rawInputs[0].(common.Address)
	if !ok {
		return nil, nil, fmt.Errorf("invalid controller address: %v", rawInputs[0])
	}

	if controller == (common.Address{}) {
		return nil, nil, fmt.Errorf("controller address must not be the zero address")
	}

	isAllowed, ok := rawInputs[1].(bool)
	if !ok {
		return nil, nil, fmt.Errorf("invalid isAllowed value: %v", rawInputs[1])
	}

	if err := m.poaKeeper.CheckOwner(
		context.SdkCtx(),
		precompile.TypesConverter.Address.ToSDK(context.MsgSender()),
	); err != nil {
		return nil, nil, err
	}

	m.bridgeKeeper.AllowTripartyController(
		context.SdkCtx(),
		precompile.TypesConverter.Address.ToSDK(controller),
		isAllowed,
	)

	err := context.EventEmitter().Emit(
		NewTripartyControllerAllowedEvent(controller, isAllowed),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to emit TripartyControllerAllowed event: [%w]", err)
	}

	return precompile.MethodOutputs{true}, nil, nil
}

// --- isAllowedTripartyController ---

type IsAllowedTripartyControllerMethod struct {
	bridgeKeeper BridgeKeeper
}

func newIsAllowedTripartyControllerMethod(
	bridgeKeeper BridgeKeeper,
) *IsAllowedTripartyControllerMethod {
	return &IsAllowedTripartyControllerMethod{bridgeKeeper: bridgeKeeper}
}

func (m *IsAllowedTripartyControllerMethod) MethodName() string {
	return IsAllowedTripartyControllerMethodName
}

func (m *IsAllowedTripartyControllerMethod) MethodType() precompile.MethodType {
	return precompile.Read
}

func (m *IsAllowedTripartyControllerMethod) RequiredGas(_ []byte) (uint64, bool) {
	return 0, false
}

func (m *IsAllowedTripartyControllerMethod) Payable() bool {
	return false
}

func (m *IsAllowedTripartyControllerMethod) Run(
	context *precompile.RunContext,
	rawInputs precompile.MethodInputs,
) (precompile.MethodOutputs, []statedb.StateChange, error) {
	if err := precompile.ValidateMethodInputsCount(rawInputs, 1); err != nil {
		return nil, nil, err
	}

	controller, ok := rawInputs[0].(common.Address)
	if !ok {
		return nil, nil, fmt.Errorf("invalid controller address: %v", rawInputs[0])
	}

	isAllowed := m.bridgeKeeper.IsAllowedTripartyController(
		context.SdkCtx(),
		precompile.TypesConverter.Address.ToSDK(controller),
	)

	return precompile.MethodOutputs{isAllowed}, nil, nil
}

// --- pauseTriparty ---

type PauseTripartyMethod struct {
	bridgeKeeper BridgeKeeper
}

func newPauseTripartyMethod(bridgeKeeper BridgeKeeper) *PauseTripartyMethod {
	return &PauseTripartyMethod{bridgeKeeper: bridgeKeeper}
}

func (m *PauseTripartyMethod) MethodName() string {
	return PauseTripartyMethodName
}

func (m *PauseTripartyMethod) MethodType() precompile.MethodType {
	return precompile.Write
}

func (m *PauseTripartyMethod) RequiredGas(_ []byte) (uint64, bool) {
	return 0, false
}

func (m *PauseTripartyMethod) Payable() bool {
	return false
}

func (m *PauseTripartyMethod) Run(
	context *precompile.RunContext,
	rawInputs precompile.MethodInputs,
) (precompile.MethodOutputs, []statedb.StateChange, error) {
	if err := precompile.ValidateMethodInputsCount(rawInputs, 1); err != nil {
		return nil, nil, err
	}

	isPaused, ok := rawInputs[0].(bool)
	if !ok {
		return nil, nil, fmt.Errorf("invalid isPaused value: %v", rawInputs[0])
	}

	sdkCtx := context.SdkCtx()

	pauser := m.bridgeKeeper.GetPauser(sdkCtx)
	if evmtypes.IsZeroHexAddress(evmtypes.BytesToHexAddress(pauser)) {
		return nil, nil, fmt.Errorf("no pauser is set")
	}

	sender := precompile.TypesConverter.Address.ToSDK(context.MsgSender())
	if !pauser.Equals(sender) {
		return nil, nil, fmt.Errorf("caller is not the pauser")
	}

	m.bridgeKeeper.SetTripartyPaused(sdkCtx, isPaused)

	err := context.EventEmitter().Emit(
		NewTripartyPausedEvent(isPaused),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to emit TripartyPaused event: [%w]", err)
	}

	return precompile.MethodOutputs{true}, nil, nil
}

// --- setTripartyBlockDelay ---

type SetTripartyBlockDelayMethod struct {
	poaKeeper    PoaKeeper
	bridgeKeeper BridgeKeeper
}

func newSetTripartyBlockDelayMethod(
	poaKeeper PoaKeeper,
	bridgeKeeper BridgeKeeper,
) *SetTripartyBlockDelayMethod {
	return &SetTripartyBlockDelayMethod{
		poaKeeper:    poaKeeper,
		bridgeKeeper: bridgeKeeper,
	}
}

func (m *SetTripartyBlockDelayMethod) MethodName() string {
	return SetTripartyBlockDelayMethodName
}

func (m *SetTripartyBlockDelayMethod) MethodType() precompile.MethodType {
	return precompile.Write
}

func (m *SetTripartyBlockDelayMethod) RequiredGas(_ []byte) (uint64, bool) {
	return 0, false
}

func (m *SetTripartyBlockDelayMethod) Payable() bool {
	return false
}

func (m *SetTripartyBlockDelayMethod) Run(
	context *precompile.RunContext,
	rawInputs precompile.MethodInputs,
) (precompile.MethodOutputs, []statedb.StateChange, error) {
	if err := precompile.ValidateMethodInputsCount(rawInputs, 1); err != nil {
		return nil, nil, err
	}

	delay, ok := rawInputs[0].(*big.Int)
	if !ok {
		return nil, nil, fmt.Errorf("invalid delay value: %v", rawInputs[0])
	}

	if delay == nil || delay.Sign() <= 0 {
		return nil, nil, fmt.Errorf("delay must be at least 1")
	}

	if !delay.IsUint64() {
		return nil, nil, fmt.Errorf("delay exceeds maximum uint64 value")
	}

	if err := m.poaKeeper.CheckOwner(
		context.SdkCtx(),
		precompile.TypesConverter.Address.ToSDK(context.MsgSender()),
	); err != nil {
		return nil, nil, err
	}

	m.bridgeKeeper.SetTripartyBlockDelay(
		context.SdkCtx(),
		delay.Uint64(),
	)

	err := context.EventEmitter().Emit(
		NewTripartyBlockDelaySetEvent(delay),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to emit TripartyBlockDelaySet event: [%w]", err)
	}

	return precompile.MethodOutputs{true}, nil, nil
}

// --- Events ---

const (
	TripartyBridgeRequestedEventName   = "TripartyBridgeRequested"
	TripartyControllerAllowedEventName = "TripartyControllerAllowed"
	TripartyPausedEventName            = "TripartyPaused"
	TripartyBlockDelaySetEventName     = "TripartyBlockDelaySet"
)

// TripartyBridgeRequestedEvent is emitted when a triparty bridge request is made.
type TripartyBridgeRequestedEvent struct {
	recipient  common.Address
	amount     *big.Int
	controller common.Address
}

func NewTripartyBridgeRequestedEvent(
	recipient common.Address,
	amount *big.Int,
	controller common.Address,
) *TripartyBridgeRequestedEvent {
	return &TripartyBridgeRequestedEvent{
		recipient:  recipient,
		amount:     amount,
		controller: controller,
	}
}

func (e *TripartyBridgeRequestedEvent) EventName() string {
	return TripartyBridgeRequestedEventName
}

func (e *TripartyBridgeRequestedEvent) Arguments() []*precompile.EventArgument {
	return []*precompile.EventArgument{
		{Indexed: true, Value: e.recipient},
		{Indexed: false, Value: e.amount},
		{Indexed: false, Value: e.controller},
	}
}

// TripartyControllerAllowedEvent is emitted when a triparty controller is allowed or disallowed.
type TripartyControllerAllowedEvent struct {
	controller common.Address
	isAllowed  bool
}

func NewTripartyControllerAllowedEvent(
	controller common.Address,
	isAllowed bool,
) *TripartyControllerAllowedEvent {
	return &TripartyControllerAllowedEvent{
		controller: controller,
		isAllowed:  isAllowed,
	}
}

func (e *TripartyControllerAllowedEvent) EventName() string {
	return TripartyControllerAllowedEventName
}

func (e *TripartyControllerAllowedEvent) Arguments() []*precompile.EventArgument {
	return []*precompile.EventArgument{
		{Indexed: true, Value: e.controller},
		{Indexed: false, Value: e.isAllowed},
	}
}

// TripartyPausedEvent is emitted when triparty bridging is paused or unpaused.
type TripartyPausedEvent struct {
	isPaused bool
}

func NewTripartyPausedEvent(isPaused bool) *TripartyPausedEvent {
	return &TripartyPausedEvent{isPaused: isPaused}
}

func (e *TripartyPausedEvent) EventName() string {
	return TripartyPausedEventName
}

func (e *TripartyPausedEvent) Arguments() []*precompile.EventArgument {
	return []*precompile.EventArgument{
		{Indexed: false, Value: e.isPaused},
	}
}

// TripartyBlockDelaySetEvent is emitted when the triparty block delay is updated.
type TripartyBlockDelaySetEvent struct {
	delay *big.Int
}

func NewTripartyBlockDelaySetEvent(delay *big.Int) *TripartyBlockDelaySetEvent {
	return &TripartyBlockDelaySetEvent{delay: delay}
}

func (e *TripartyBlockDelaySetEvent) EventName() string {
	return TripartyBlockDelaySetEventName
}

func (e *TripartyBlockDelaySetEvent) Arguments() []*precompile.EventArgument {
	return []*precompile.EventArgument{
		{Indexed: false, Value: e.delay},
	}
}
