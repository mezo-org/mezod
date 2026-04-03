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
	GetTripartyBlockDelayMethodName       = "getTripartyBlockDelay"
	SetTripartyLimitsMethodName           = "setTripartyLimits"
	GetTripartyLimitsMethodName           = "getTripartyLimits"
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
	if err := precompile.ValidateMethodInputsCount(rawInputs, 3); err != nil {
		return nil, nil, err
	}

	recipient, ok := rawInputs[0].(common.Address)
	if !ok {
		return nil, nil, fmt.Errorf("invalid recipient address: %v", rawInputs[0])
	}

	amount, ok := rawInputs[1].(*big.Int)
	if !ok {
		return nil, nil, fmt.Errorf("invalid amount: %v", rawInputs[1])
	}

	if amount == nil {
		return nil, nil, fmt.Errorf("invalid amount: nil")
	}

	callbackData, ok := rawInputs[2].([]byte)
	if !ok {
		return nil, nil, fmt.Errorf("invalid callbackData: %v", rawInputs[2])
	}

	sdkAmount, err := precompile.TypesConverter.BigInt.ToSDK(amount)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to convert amount: [%w]", err)
	}

	requestID, err := m.bridgeKeeper.CreateTripartyBridgeRequest(
		context.SdkCtx(),
		recipient.Hex(),
		sdkAmount,
		callbackData,
		context.MsgSender().Hex(),
	)
	if err != nil {
		return nil, nil, err
	}

	requestIDBigInt := precompile.TypesConverter.BigInt.FromSDK(requestID)

	err = context.EventEmitter().Emit(
		NewTripartyBridgeRequestedEvent(requestIDBigInt, recipient, amount, context.MsgSender()),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to emit TripartyBridgeRequested event: [%w]", err)
	}

	return precompile.MethodOutputs{requestIDBigInt}, nil, nil
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

	if !delay.IsInt64() {
		return nil, nil, fmt.Errorf("delay exceeds maximum value")
	}

	if err := m.poaKeeper.CheckOwner(
		context.SdkCtx(),
		precompile.TypesConverter.Address.ToSDK(context.MsgSender()),
	); err != nil {
		return nil, nil, err
	}

	m.bridgeKeeper.SetTripartyBlockDelay(
		context.SdkCtx(),
		delay.Int64(),
	)

	err := context.EventEmitter().Emit(
		NewTripartyBlockDelaySetEvent(delay),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to emit TripartyBlockDelaySet event: [%w]", err)
	}

	return precompile.MethodOutputs{true}, nil, nil
}

// --- getTripartyBlockDelay ---

type GetTripartyBlockDelayMethod struct {
	bridgeKeeper BridgeKeeper
}

func newGetTripartyBlockDelayMethod(
	bridgeKeeper BridgeKeeper,
) *GetTripartyBlockDelayMethod {
	return &GetTripartyBlockDelayMethod{bridgeKeeper: bridgeKeeper}
}

func (m *GetTripartyBlockDelayMethod) MethodName() string {
	return GetTripartyBlockDelayMethodName
}

func (m *GetTripartyBlockDelayMethod) MethodType() precompile.MethodType {
	return precompile.Read
}

func (m *GetTripartyBlockDelayMethod) RequiredGas(_ []byte) (uint64, bool) {
	return 0, false
}

func (m *GetTripartyBlockDelayMethod) Payable() bool {
	return false
}

func (m *GetTripartyBlockDelayMethod) Run(
	context *precompile.RunContext,
	rawInputs precompile.MethodInputs,
) (precompile.MethodOutputs, []statedb.StateChange, error) {
	if err := precompile.ValidateMethodInputsCount(rawInputs, 0); err != nil {
		return nil, nil, err
	}

	delay := m.bridgeKeeper.GetTripartyBlockDelay(context.SdkCtx())

	return precompile.MethodOutputs{new(big.Int).SetInt64(delay)}, nil, nil
}

// --- Events ---

const (
	TripartyBridgeRequestedEventName   = "TripartyBridgeRequested"
	TripartyControllerAllowedEventName = "TripartyControllerAllowed"
	TripartyPausedEventName            = "TripartyPaused"
	TripartyBlockDelaySetEventName     = "TripartyBlockDelaySet"
	TripartyLimitsSetEventName         = "TripartyLimitsSet"
)

// TripartyBridgeRequestedEvent is emitted when a triparty bridge request is made.
type TripartyBridgeRequestedEvent struct {
	requestID  *big.Int
	recipient  common.Address
	amount     *big.Int
	controller common.Address
}

func NewTripartyBridgeRequestedEvent(
	requestID *big.Int,
	recipient common.Address,
	amount *big.Int,
	controller common.Address,
) *TripartyBridgeRequestedEvent {
	return &TripartyBridgeRequestedEvent{
		requestID:  requestID,
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
		{Indexed: true, Value: e.requestID},
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

// --- setTripartyLimits ---

type SetTripartyLimitsMethod struct {
	poaKeeper    PoaKeeper
	bridgeKeeper BridgeKeeper
}

func newSetTripartyLimitsMethod(
	poaKeeper PoaKeeper,
	bridgeKeeper BridgeKeeper,
) *SetTripartyLimitsMethod {
	return &SetTripartyLimitsMethod{
		poaKeeper:    poaKeeper,
		bridgeKeeper: bridgeKeeper,
	}
}

func (m *SetTripartyLimitsMethod) MethodName() string {
	return SetTripartyLimitsMethodName
}

func (m *SetTripartyLimitsMethod) MethodType() precompile.MethodType {
	return precompile.Write
}

func (m *SetTripartyLimitsMethod) RequiredGas(_ []byte) (uint64, bool) {
	return 0, false
}

func (m *SetTripartyLimitsMethod) Payable() bool {
	return false
}

func (m *SetTripartyLimitsMethod) Run(
	context *precompile.RunContext,
	rawInputs precompile.MethodInputs,
) (precompile.MethodOutputs, []statedb.StateChange, error) {
	if err := precompile.ValidateMethodInputsCount(rawInputs, 2); err != nil {
		return nil, nil, err
	}

	perRequestLimit, ok := rawInputs[0].(*big.Int)
	if !ok {
		return nil, nil, fmt.Errorf("invalid perRequestLimit value: %v", rawInputs[0])
	}

	if perRequestLimit == nil || perRequestLimit.Sign() < 0 {
		return nil, nil, fmt.Errorf("perRequestLimit must be non-negative")
	}

	windowLimit, ok := rawInputs[1].(*big.Int)
	if !ok {
		return nil, nil, fmt.Errorf("invalid windowLimit value: %v", rawInputs[1])
	}

	if windowLimit == nil || windowLimit.Sign() < 0 {
		return nil, nil, fmt.Errorf("windowLimit must be non-negative")
	}

	if err := m.poaKeeper.CheckOwner(
		context.SdkCtx(),
		precompile.TypesConverter.Address.ToSDK(context.MsgSender()),
	); err != nil {
		return nil, nil, err
	}

	sdkPerRequestLimit, err := precompile.TypesConverter.BigInt.ToSDK(perRequestLimit)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to convert perRequestLimit: [%w]", err)
	}

	sdkWindowLimit, err := precompile.TypesConverter.BigInt.ToSDK(windowLimit)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to convert windowLimit: [%w]", err)
	}

	m.bridgeKeeper.SetTripartyPerRequestLimit(context.SdkCtx(), sdkPerRequestLimit)
	m.bridgeKeeper.SetTripartyWindowLimit(context.SdkCtx(), sdkWindowLimit)

	err = context.EventEmitter().Emit(
		NewTripartyLimitsSetEvent(perRequestLimit, windowLimit),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to emit TripartyLimitsSet event: [%w]", err)
	}

	return precompile.MethodOutputs{true}, nil, nil
}

// --- getTripartyLimits ---

type GetTripartyLimitsMethod struct {
	bridgeKeeper BridgeKeeper
}

func newGetTripartyLimitsMethod(
	bridgeKeeper BridgeKeeper,
) *GetTripartyLimitsMethod {
	return &GetTripartyLimitsMethod{bridgeKeeper: bridgeKeeper}
}

func (m *GetTripartyLimitsMethod) MethodName() string {
	return GetTripartyLimitsMethodName
}

func (m *GetTripartyLimitsMethod) MethodType() precompile.MethodType {
	return precompile.Read
}

func (m *GetTripartyLimitsMethod) RequiredGas(_ []byte) (uint64, bool) {
	return 0, false
}

func (m *GetTripartyLimitsMethod) Payable() bool {
	return false
}

func (m *GetTripartyLimitsMethod) Run(
	context *precompile.RunContext,
	rawInputs precompile.MethodInputs,
) (precompile.MethodOutputs, []statedb.StateChange, error) {
	if err := precompile.ValidateMethodInputsCount(rawInputs, 0); err != nil {
		return nil, nil, err
	}

	sdkCtx := context.SdkCtx()

	perRequestLimit := m.bridgeKeeper.GetTripartyPerRequestLimit(sdkCtx)
	windowLimit := m.bridgeKeeper.GetTripartyWindowLimit(sdkCtx)

	return precompile.MethodOutputs{
		precompile.TypesConverter.BigInt.FromSDK(perRequestLimit),
		precompile.TypesConverter.BigInt.FromSDK(windowLimit),
	}, nil, nil
}

// TripartyLimitsSetEvent is emitted when the triparty limits are updated.
type TripartyLimitsSetEvent struct {
	perRequestLimit *big.Int
	windowLimit     *big.Int
}

func NewTripartyLimitsSetEvent(
	perRequestLimit *big.Int,
	windowLimit *big.Int,
) *TripartyLimitsSetEvent {
	return &TripartyLimitsSetEvent{
		perRequestLimit: perRequestLimit,
		windowLimit:     windowLimit,
	}
}

func (e *TripartyLimitsSetEvent) EventName() string {
	return TripartyLimitsSetEventName
}

func (e *TripartyLimitsSetEvent) Arguments() []*precompile.EventArgument {
	return []*precompile.EventArgument{
		{Indexed: false, Value: e.perRequestLimit},
		{Indexed: false, Value: e.windowLimit},
	}
}
