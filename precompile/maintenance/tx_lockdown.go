package maintenance

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/mezo-org/mezod/precompile"
	"github.com/mezo-org/mezod/x/evm/statedb"
	evmtypes "github.com/mezo-org/mezod/x/evm/types"
)

// SetTxLockdownMethodName is the name of the setTxLockdown method. It matches
// the name of the method in the contract ABI.
const SetTxLockdownMethodName = "setTxLockdown"

// TxLockdownSetEventName is the name of the TxLockdownSet event. It matches
// the name of the event in the contract ABI.
const TxLockdownSetEventName = "TxLockdownSet"

// setTxLockdownMethod enables or disables the transaction lockdown. Disabling
// the lockdown also clears the extra allowlists.
type setTxLockdownMethod struct {
	poaKeeper PoaKeeper
	evmKeeper EvmKeeper
}

func newSetTxLockdownMethod(
	poaKeeper PoaKeeper,
	evmKeeper EvmKeeper,
) *setTxLockdownMethod {
	return &setTxLockdownMethod{
		poaKeeper: poaKeeper,
		evmKeeper: evmKeeper,
	}
}

func (m *setTxLockdownMethod) MethodName() string {
	return SetTxLockdownMethodName
}

func (m *setTxLockdownMethod) MethodType() precompile.MethodType {
	return precompile.Write
}

func (m *setTxLockdownMethod) RequiredGas(_ []byte) (uint64, bool) {
	return 0, false
}

func (m *setTxLockdownMethod) Payable() bool {
	return false
}

func (m *setTxLockdownMethod) Run(
	context *precompile.RunContext,
	inputs precompile.MethodInputs,
) (precompile.MethodOutputs, []statedb.StateChange, error) {
	if err := precompile.ValidateMethodInputsCount(inputs, 1); err != nil {
		return nil, nil, err
	}

	enabled, ok := inputs[0].(bool)
	if !ok {
		return nil, nil, fmt.Errorf("enabled argument must be a boolean")
	}

	// This method is restricted to the PoA owner and the emergency team.
	err := m.poaKeeper.CheckOwnerOrEmergencyTeam(
		context.SdkCtx(),
		precompile.TypesConverter.Address.ToSDK(context.MsgSender()),
	)
	if err != nil {
		return nil, nil, err
	}

	params := m.evmKeeper.GetParams(context.SdkCtx())
	params.TxLockdownEnabled = enabled

	// Disabling the lockdown clears both extra allowlists. The next lockdown
	// starts from a known state.
	if !enabled {
		params.TxLockdownSenders = nil
		params.TxLockdownTargets = nil
	}

	if err := m.evmKeeper.SetParams(context.SdkCtx(), params); err != nil {
		return nil, nil, err
	}

	err = context.EventEmitter().Emit(
		NewTxLockdownSetEvent(enabled),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to emit TxLockdownSet event: [%w]", err)
	}

	return precompile.MethodOutputs{true}, nil, nil
}

// GetTxLockdownMethodName is the name of the getTxLockdown method. It matches
// the name of the method in the contract ABI.
const GetTxLockdownMethodName = "getTxLockdown"

// getTxLockdownMethod reports the transaction lockdown state.
type getTxLockdownMethod struct {
	evmKeeper EvmKeeper
}

func newGetTxLockdownMethod(evmKeeper EvmKeeper) *getTxLockdownMethod {
	return &getTxLockdownMethod{evmKeeper: evmKeeper}
}

func (m *getTxLockdownMethod) MethodName() string {
	return GetTxLockdownMethodName
}

func (m *getTxLockdownMethod) MethodType() precompile.MethodType {
	return precompile.Read
}

func (m *getTxLockdownMethod) RequiredGas(_ []byte) (uint64, bool) {
	return 0, false
}

func (m *getTxLockdownMethod) Payable() bool {
	return false
}

func (m *getTxLockdownMethod) Run(
	context *precompile.RunContext,
	inputs precompile.MethodInputs,
) (precompile.MethodOutputs, []statedb.StateChange, error) {
	if err := precompile.ValidateMethodInputsCount(inputs, 0); err != nil {
		return nil, nil, err
	}

	params := m.evmKeeper.GetParams(context.SdkCtx())

	return precompile.MethodOutputs{params.TxLockdownEnabled}, nil, nil
}

// SetTxLockdownAllowlistMethodName is the name of the setTxLockdownAllowlist
// method. It matches the name of the method in the contract ABI.
const SetTxLockdownAllowlistMethodName = "setTxLockdownAllowlist"

// TxLockdownAllowlistSetEventName is the name of the TxLockdownAllowlistSet
// event. It matches the name of the event in the contract ABI.
const TxLockdownAllowlistSetEventName = "TxLockdownAllowlistSet"

// setTxLockdownAllowlistMethod replaces the extra allowlists evaluated while
// the transaction lockdown is active.
type setTxLockdownAllowlistMethod struct {
	poaKeeper PoaKeeper
	evmKeeper EvmKeeper
}

func newSetTxLockdownAllowlistMethod(
	poaKeeper PoaKeeper,
	evmKeeper EvmKeeper,
) *setTxLockdownAllowlistMethod {
	return &setTxLockdownAllowlistMethod{
		poaKeeper: poaKeeper,
		evmKeeper: evmKeeper,
	}
}

func (m *setTxLockdownAllowlistMethod) MethodName() string {
	return SetTxLockdownAllowlistMethodName
}

func (m *setTxLockdownAllowlistMethod) MethodType() precompile.MethodType {
	return precompile.Write
}

func (m *setTxLockdownAllowlistMethod) RequiredGas(_ []byte) (uint64, bool) {
	return 0, false
}

func (m *setTxLockdownAllowlistMethod) Payable() bool {
	return false
}

func (m *setTxLockdownAllowlistMethod) Run(
	context *precompile.RunContext,
	inputs precompile.MethodInputs,
) (precompile.MethodOutputs, []statedb.StateChange, error) {
	if err := precompile.ValidateMethodInputsCount(inputs, 2); err != nil {
		return nil, nil, err
	}

	senders, ok := inputs[0].([]common.Address)
	if !ok {
		return nil, nil, fmt.Errorf("senders argument must be of type []common.Address")
	}

	targets, ok := inputs[1].([]common.Address)
	if !ok {
		return nil, nil, fmt.Errorf("targets argument must be of type []common.Address")
	}

	// This method is restricted to the PoA owner and the emergency team.
	err := m.poaKeeper.CheckOwnerOrEmergencyTeam(
		context.SdkCtx(),
		precompile.TypesConverter.Address.ToSDK(context.MsgSender()),
	)
	if err != nil {
		return nil, nil, err
	}

	params := m.evmKeeper.GetParams(context.SdkCtx())
	// The call replaces both lists, it does not append to them.
	params.TxLockdownSenders = txLockdownHexAddresses(senders)
	params.TxLockdownTargets = txLockdownHexAddresses(targets)

	// SetParams validates the parameters, so an invalid or oversized list
	// reverts the call.
	if err := m.evmKeeper.SetParams(context.SdkCtx(), params); err != nil {
		return nil, nil, err
	}

	err = context.EventEmitter().Emit(
		NewTxLockdownAllowlistSetEvent(senders, targets),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to emit TxLockdownAllowlistSet event: [%w]", err)
	}

	return precompile.MethodOutputs{true}, nil, nil
}

// txLockdownHexAddresses converts EVM addresses to their normalized
// hex-encoded form used in the EVM parameters.
func txLockdownHexAddresses(addresses []common.Address) []string {
	hexAddresses := make([]string, len(addresses))
	for i, address := range addresses {
		hexAddresses[i] = evmtypes.BytesToHexAddress(address.Bytes())
	}

	return hexAddresses
}

// GetTxLockdownAllowlistMethodName is the name of the getTxLockdownAllowlist
// method. It matches the name of the method in the contract ABI.
const GetTxLockdownAllowlistMethodName = "getTxLockdownAllowlist"

// getTxLockdownAllowlistMethod reports the extra allowlists evaluated while the
// transaction lockdown is active.
type getTxLockdownAllowlistMethod struct {
	evmKeeper EvmKeeper
}

func newGetTxLockdownAllowlistMethod(evmKeeper EvmKeeper) *getTxLockdownAllowlistMethod {
	return &getTxLockdownAllowlistMethod{evmKeeper: evmKeeper}
}

func (m *getTxLockdownAllowlistMethod) MethodName() string {
	return GetTxLockdownAllowlistMethodName
}

func (m *getTxLockdownAllowlistMethod) MethodType() precompile.MethodType {
	return precompile.Read
}

func (m *getTxLockdownAllowlistMethod) RequiredGas(_ []byte) (uint64, bool) {
	return 0, false
}

func (m *getTxLockdownAllowlistMethod) Payable() bool {
	return false
}

func (m *getTxLockdownAllowlistMethod) Run(
	context *precompile.RunContext,
	inputs precompile.MethodInputs,
) (precompile.MethodOutputs, []statedb.StateChange, error) {
	if err := precompile.ValidateMethodInputsCount(inputs, 0); err != nil {
		return nil, nil, err
	}

	params := m.evmKeeper.GetParams(context.SdkCtx())

	return precompile.MethodOutputs{
		txLockdownAddresses(params.TxLockdownSenders),
		txLockdownAddresses(params.TxLockdownTargets),
	}, nil, nil
}

// txLockdownAddresses converts hex-encoded addresses from the EVM parameters
// to EVM addresses.
func txLockdownAddresses(hexAddresses []string) []common.Address {
	addresses := make([]common.Address, len(hexAddresses))
	for i, hexAddress := range hexAddresses {
		addresses[i] = common.BytesToAddress(
			evmtypes.HexAddressToBytes(hexAddress),
		)
	}

	return addresses
}

// TxLockdownSetEvent is emitted when the transaction lockdown is enabled or
// disabled.
type TxLockdownSetEvent struct {
	enabled bool
}

func NewTxLockdownSetEvent(enabled bool) *TxLockdownSetEvent {
	return &TxLockdownSetEvent{enabled: enabled}
}

func (e *TxLockdownSetEvent) EventName() string {
	return TxLockdownSetEventName
}

func (e *TxLockdownSetEvent) Arguments() []*precompile.EventArgument {
	return []*precompile.EventArgument{
		{Indexed: false, Value: e.enabled},
	}
}

// TxLockdownAllowlistSetEvent is emitted when the extra allowlists of the
// transaction lockdown are replaced.
type TxLockdownAllowlistSetEvent struct {
	senders []common.Address
	targets []common.Address
}

func NewTxLockdownAllowlistSetEvent(
	senders []common.Address,
	targets []common.Address,
) *TxLockdownAllowlistSetEvent {
	return &TxLockdownAllowlistSetEvent{
		senders: senders,
		targets: targets,
	}
}

func (e *TxLockdownAllowlistSetEvent) EventName() string {
	return TxLockdownAllowlistSetEventName
}

func (e *TxLockdownAllowlistSetEvent) Arguments() []*precompile.EventArgument {
	return []*precompile.EventArgument{
		{Indexed: false, Value: e.senders},
		{Indexed: false, Value: e.targets},
	}
}
