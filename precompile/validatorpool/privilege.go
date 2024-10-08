package validatorpool

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/mezo-org/mezod/precompile"
	bridgetypes "github.com/mezo-org/mezod/x/bridge/types"
)

// privileges is a map of privilege id to privilege name. Captures all
// privileges supported by the validator pool.
var privileges = map[uint8]string{
	1: bridgetypes.ValidatorPrivilege,
}

// AddPrivilegeMethodName is the name of the addPrivilege method. It matches the
// name of the method in the contract ABI.
const AddPrivilegeMethodName = "addPrivilege"

// AddPrivilegeMethod is the implementation of the addPrivilege method that
// adds the given privilege to the specified operators.
//
// The method has the following input arguments:
// - operators: list of operator addresses to add the privilege to,
// - privilegeId: the privilege to add.
type AddPrivilegeMethod struct {
	keeper PoaKeeper
}

func newAddPrivilegeMethod(pk PoaKeeper) *AddPrivilegeMethod {
	return &AddPrivilegeMethod{
		keeper: pk,
	}
}

func (apm *AddPrivilegeMethod) MethodName() string {
	return AddPrivilegeMethodName
}

func (apm *AddPrivilegeMethod) MethodType() precompile.MethodType {
	return precompile.Write
}

func (apm *AddPrivilegeMethod) RequiredGas(_ []byte) (
	uint64,
	bool,
) {
	// Fallback to the default gas calculation.
	return 0, false
}

func (apm *AddPrivilegeMethod) Payable() bool {
	return false
}

func (apm *AddPrivilegeMethod) Run(
	context *precompile.RunContext,
	inputs precompile.MethodInputs,
) (precompile.MethodOutputs, error) {
	if err := precompile.ValidateMethodInputsCount(inputs, 2); err != nil {
		return nil, err
	}

	operators, ok := inputs[0].([]common.Address)
	if !ok {
		return nil, fmt.Errorf("operators argument must be of type []common.Address")
	}

	//nolint:revive,stylecheck
	privilegeId, ok := inputs[1].(uint8)
	if !ok {
		return nil, fmt.Errorf("privilegeId argument must be of type uint8")
	}

	privilege, ok := privileges[privilegeId]
	if !ok {
		return nil, fmt.Errorf("unknown privilege id")
	}

	operatorsSdk := make([]types.ValAddress, len(operators))
	for i, operator := range operators {
		operatorsSdk[i] = types.ValAddress(precompile.TypesConverter.Address.ToSDK(operator))
	}

	err := apm.keeper.AddPrivilege(
		context.SdkCtx(),
		precompile.TypesConverter.Address.ToSDK(context.MsgSender()),
		operatorsSdk,
		privilege,
	)
	if err != nil {
		return nil, err
	}

	for _, operator := range operators {
		err = context.EventEmitter().Emit(
			NewPrivilegeAddedEvent(
				operator,
				privilegeId,
			),
		)
		if err != nil {
			return nil, fmt.Errorf("failed to emit PrivilegeAdded event: [%w]", err)
		}
	}

	return precompile.MethodOutputs{true}, nil
}

// RemovePrivilegeMethodName is the name of the removePrivilege method. It
// matches the name of the method in the contract ABI.
const RemovePrivilegeMethodName = "removePrivilege"

// RemovePrivilegeMethod is the implementation of the removePrivilege method that
// removes the given privilege from the specified operators.
//
// The method has the following input arguments:
// - operators: list of operator addresses to remove the privilege from,
// - privilegeId: the privilege to remove.
type RemovePrivilegeMethod struct {
	keeper PoaKeeper
}

func newRemovePrivilegeMethod(pk PoaKeeper) *RemovePrivilegeMethod {
	return &RemovePrivilegeMethod{
		keeper: pk,
	}
}

func (rpm *RemovePrivilegeMethod) MethodName() string {
	return RemovePrivilegeMethodName
}

func (rpm *RemovePrivilegeMethod) MethodType() precompile.MethodType {
	return precompile.Write
}

func (rpm *RemovePrivilegeMethod) RequiredGas(_ []byte) (
	uint64,
	bool,
) {
	// Fallback to the default gas calculation.
	return 0, false
}

func (rpm *RemovePrivilegeMethod) Payable() bool {
	return false
}

func (rpm *RemovePrivilegeMethod) Run(
	context *precompile.RunContext,
	inputs precompile.MethodInputs,
) (precompile.MethodOutputs, error) {
	if err := precompile.ValidateMethodInputsCount(inputs, 2); err != nil {
		return nil, err
	}

	operators, ok := inputs[0].([]common.Address)
	if !ok {
		return nil, fmt.Errorf("operators argument must be of type []common.Address")
	}

	//nolint:revive,stylecheck
	privilegeId, ok := inputs[1].(uint8)
	if !ok {
		return nil, fmt.Errorf("privilegeId argument must be of type uint8")
	}

	privilege, ok := privileges[privilegeId]
	if !ok {
		return nil, fmt.Errorf("unknown privilege id")
	}

	operatorsSdk := make([]types.ValAddress, len(operators))
	for i, operator := range operators {
		operatorsSdk[i] = types.ValAddress(precompile.TypesConverter.Address.ToSDK(operator))
	}

	err := rpm.keeper.RemovePrivilege(
		context.SdkCtx(),
		precompile.TypesConverter.Address.ToSDK(context.MsgSender()),
		operatorsSdk,
		privilege,
	)
	if err != nil {
		return nil, err
	}

	for _, operator := range operators {
		err = context.EventEmitter().Emit(
			NewPrivilegeRemovedEvent(
				operator,
				privilegeId,
			),
		)
		if err != nil {
			return nil, fmt.Errorf("failed to emit PrivilegeRemoved event: [%w]", err)
		}
	}

	return precompile.MethodOutputs{true}, nil
}

// PrivilegeAddedEventName is the name of the PrivilegeAdded event.
// It matches the name of the event in the contract ABI.
const PrivilegeAddedEventName = "PrivilegeAdded"

// PrivilegeAddedEvent is the implementation of the PrivilegeAdded event that
// contains the following arguments:
// - operator (indexed): is the operator address of the validator,
// - privilegeId (indexed): is the privilege added.
type PrivilegeAddedEvent struct {
	operator common.Address
	//nolint:revive,stylecheck
	privilegeId uint8
}

//nolint:revive,stylecheck
func NewPrivilegeAddedEvent(operator common.Address, privilegeId uint8) *PrivilegeAddedEvent {
	return &PrivilegeAddedEvent{
		operator:    operator,
		privilegeId: privilegeId,
	}
}

func (pae *PrivilegeAddedEvent) EventName() string {
	return PrivilegeAddedEventName
}

func (pae *PrivilegeAddedEvent) Arguments() []*precompile.EventArgument {
	return []*precompile.EventArgument{
		{
			Indexed: true,
			Value:   pae.operator,
		},
		{
			Indexed: true,
			Value:   pae.privilegeId,
		},
	}
}

// PrivilegeRemovedEventName is the name of the PrivilegeRemoved event.
// It matches the name of the event in the contract ABI.
const PrivilegeRemovedEventName = "PrivilegeRemoved"

// PrivilegeRemovedEvent is the implementation of the PrivilegeRemoved event that
// contains the following arguments:
// - operator (indexed): is the operator address of the validator,
// - privilegeId (indexed): is the privilege removed.
type PrivilegeRemovedEvent struct {
	operator common.Address
	//nolint:revive,stylecheck
	privilegeId uint8
}

//nolint:revive,stylecheck
func NewPrivilegeRemovedEvent(operator common.Address, privilegeId uint8) *PrivilegeRemovedEvent {
	return &PrivilegeRemovedEvent{
		operator:    operator,
		privilegeId: privilegeId,
	}
}

func (pre *PrivilegeRemovedEvent) EventName() string {
	return PrivilegeRemovedEventName
}

func (pre *PrivilegeRemovedEvent) Arguments() []*precompile.EventArgument {
	return []*precompile.EventArgument{
		{
			Indexed: true,
			Value:   pre.operator,
		},
		{
			Indexed: true,
			Value:   pre.privilegeId,
		},
	}
}
