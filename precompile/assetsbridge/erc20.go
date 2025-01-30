package assetsbridge

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/mezo-org/mezod/precompile"
	bridgetypes "github.com/mezo-org/mezod/x/bridge/types"
)

// CreateERC20TokenMappingMethodName is the name of the createERC20TokenMapping method.
// It matches the name of the method in the contract ABI.
//
//nolint:gosec
const CreateERC20TokenMappingMethodName = "createERC20TokenMapping"

// CreateERC20TokenMappingMethod is the implementation of the createERC20TokenMapping method.
type CreateERC20TokenMappingMethod struct {
	poaKeeper    PoaKeeper
	bridgeKeeper BridgeKeeper
}

func newCreateERC20TokenMappingMethod(
	poaKeeper PoaKeeper,
	bridgeKeeper BridgeKeeper,
) *CreateERC20TokenMappingMethod {
	return &CreateERC20TokenMappingMethod{
		poaKeeper:    poaKeeper,
		bridgeKeeper: bridgeKeeper,
	}
}

func (m *CreateERC20TokenMappingMethod) MethodName() string {
	return CreateERC20TokenMappingMethodName
}

func (m *CreateERC20TokenMappingMethod) MethodType() precompile.MethodType {
	return precompile.Write
}

func (m *CreateERC20TokenMappingMethod) RequiredGas(_ []byte) (uint64, bool) {
	// Fallback to the default gas calculation.
	return 0, false
}

func (m *CreateERC20TokenMappingMethod) Payable() bool {
	return false
}

func (m *CreateERC20TokenMappingMethod) Run(
	context *precompile.RunContext,
	inputs precompile.MethodInputs,
) (precompile.MethodOutputs, error) {
	if err := precompile.ValidateMethodInputsCount(inputs, 2); err != nil {
		return nil, err
	}

	sourceToken, ok := inputs[0].(common.Address)
	if !ok {
		return nil, fmt.Errorf("source token must be common.Address")
	}

	mezoToken, ok := inputs[1].(common.Address)
	if !ok {
		return nil, fmt.Errorf("mezo token must be common.Address")
	}

	err := m.poaKeeper.CheckOwner(
		context.SdkCtx(),
		precompile.TypesConverter.Address.ToSDK(context.MsgSender()),
	)
	if err != nil {
		return nil, err
	}

	err = m.bridgeKeeper.CreateERC20TokenMapping(
		context.SdkCtx(),
		&bridgetypes.ERC20TokenMapping{
			SourceToken: sourceToken.Hex(),
			MezoToken:   mezoToken.Hex(),
		},
	)
	if err != nil {
		return nil, err
	}

	err = context.EventEmitter().Emit(
		NewERC20TokenMappingCreatedEvent(sourceToken, mezoToken),
	)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to emit ERC20TokenMappingCreated event: [%w]",
			err,
		)
	}

	return precompile.MethodOutputs{true}, nil
}

// ERC20TokenMappingCreatedEventName is the name of the ERC20TokenMappingCreated event.
// It matches the name of the event in the contract ABI.
//
//nolint:gosec
const ERC20TokenMappingCreatedEventName = "ERC20TokenMappingCreated"

// ERC20TokenMappingCreatedEvent is the implementation of the ERC20TokenMappingCreated
// event that contains the following arguments:
// - sourceToken (indexed): the address of the ERC20 token on the source chain,
// - mezoToken (indexed): The address of the ERC20 token on the Mezo chain.
type ERC20TokenMappingCreatedEvent struct {
	sourceToken, mezoToken common.Address
}

func NewERC20TokenMappingCreatedEvent(
	sourceToken, mezoToken common.Address,
) *ERC20TokenMappingCreatedEvent {
	return &ERC20TokenMappingCreatedEvent{
		sourceToken: sourceToken,
		mezoToken:   mezoToken,
	}
}

func (e *ERC20TokenMappingCreatedEvent) EventName() string {
	return ERC20TokenMappingCreatedEventName
}

func (e *ERC20TokenMappingCreatedEvent) Arguments() []*precompile.EventArgument {
	return []*precompile.EventArgument{
		{
			Indexed: true,
			Value:   e.sourceToken,
		},
		{
			Indexed: true,
			Value:   e.mezoToken,
		},
	}
}

// DeleteERC20TokenMappingMethodName is the name of the deleteERC20TokenMapping method.
// It matches the name of the method in the contract ABI.
//
//nolint:gosec
const DeleteERC20TokenMappingMethodName = "deleteERC20TokenMapping"

// DeleteERC20TokenMappingMethod is the implementation of the deleteERC20TokenMapping method.
type DeleteERC20TokenMappingMethod struct {
	poaKeeper    PoaKeeper
	bridgeKeeper BridgeKeeper
}

func newDeleteERC20TokenMappingMethod(
	poaKeeper PoaKeeper,
	bridgeKeeper BridgeKeeper,
) *DeleteERC20TokenMappingMethod {
	return &DeleteERC20TokenMappingMethod{
		poaKeeper:    poaKeeper,
		bridgeKeeper: bridgeKeeper,
	}
}

func (m *DeleteERC20TokenMappingMethod) MethodName() string {
	return DeleteERC20TokenMappingMethodName
}

func (m *DeleteERC20TokenMappingMethod) MethodType() precompile.MethodType {
	return precompile.Write
}

func (m *DeleteERC20TokenMappingMethod) RequiredGas(_ []byte) (uint64, bool) {
	// Fallback to the default gas calculation.
	return 0, false
}

func (m *DeleteERC20TokenMappingMethod) Payable() bool {
	return false
}

func (m *DeleteERC20TokenMappingMethod) Run(
	context *precompile.RunContext,
	inputs precompile.MethodInputs,
) (precompile.MethodOutputs, error) {
	if err := precompile.ValidateMethodInputsCount(inputs, 1); err != nil {
		return nil, err
	}

	sourceToken, ok := inputs[0].(common.Address)
	if !ok {
		return nil, fmt.Errorf("source token must be common.Address")
	}

	err := m.poaKeeper.CheckOwner(
		context.SdkCtx(),
		precompile.TypesConverter.Address.ToSDK(context.MsgSender()),
	)
	if err != nil {
		return nil, err
	}

	// Fetch the mapping before deleting to feed the event.
	// Ignore mapping existence flag as it is checked in DeleteERC20TokenMapping.
	mapping, _ := m.bridgeKeeper.GetERC20TokenMapping(
		context.SdkCtx(),
		sourceToken.Hex(),
	)

	err = m.bridgeKeeper.DeleteERC20TokenMapping(
		context.SdkCtx(),
		sourceToken.Hex(),
	)
	if err != nil {
		return nil, err
	}

	err = context.EventEmitter().Emit(
		NewERC20TokenMappingDeletedEvent(
			sourceToken,
			common.HexToAddress(mapping.MezoToken),
		),
	)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to emit ERC20TokenMappingDeleted event: [%w]",
			err,
		)
	}

	return precompile.MethodOutputs{true}, nil
}

// ERC20TokenMappingDeletedEventName is the name of the ERC20TokenMappingDeleted event.
// It matches the name of the event in the contract ABI.
//
//nolint:gosec
const ERC20TokenMappingDeletedEventName = "ERC20TokenMappingDeleted"

// ERC20TokenMappingDeletedEvent is the implementation of the ERC20TokenMappingDeleted
// event that contains the following arguments:
// - sourceToken (indexed): the address of the ERC20 token on the source chain,
// - mezoToken (indexed): The address of the ERC20 token on the Mezo chain.
type ERC20TokenMappingDeletedEvent struct {
	sourceToken, mezoToken common.Address
}

func NewERC20TokenMappingDeletedEvent(
	sourceToken, mezoToken common.Address,
) *ERC20TokenMappingDeletedEvent {
	return &ERC20TokenMappingDeletedEvent{
		sourceToken: sourceToken,
		mezoToken:   mezoToken,
	}
}

func (e *ERC20TokenMappingDeletedEvent) EventName() string {
	return ERC20TokenMappingDeletedEventName
}

func (e *ERC20TokenMappingDeletedEvent) Arguments() []*precompile.EventArgument {
	return []*precompile.EventArgument{
		{
			Indexed: true,
			Value:   e.sourceToken,
		},
		{
			Indexed: true,
			Value:   e.mezoToken,
		},
	}
}

// GetERC20TokenMappingMethodName is the name of the getERC20TokenMapping method.
// It matches the name of the method in the contract ABI.
//
//nolint:gosec
const GetERC20TokenMappingMethodName = "getERC20TokenMapping"

// GetERC20TokenMappingMethod is the implementation of the getERC20TokenMapping method.
type GetERC20TokenMappingMethod struct {
	bridgeKeeper BridgeKeeper
}

func newGetERC20TokenMappingMethod(bridgeKeeper BridgeKeeper) *GetERC20TokenMappingMethod {
	return &GetERC20TokenMappingMethod{
		bridgeKeeper: bridgeKeeper,
	}
}

func (m *GetERC20TokenMappingMethod) MethodName() string {
	return GetERC20TokenMappingMethodName
}

func (m *GetERC20TokenMappingMethod) MethodType() precompile.MethodType {
	return precompile.Read
}

func (m *GetERC20TokenMappingMethod) RequiredGas(_ []byte) (uint64, bool) {
	// Fallback to the default gas calculation.
	return 0, false
}

func (m *GetERC20TokenMappingMethod) Payable() bool {
	return false
}

func (m *GetERC20TokenMappingMethod) Run(
	context *precompile.RunContext,
	inputs precompile.MethodInputs,
) (precompile.MethodOutputs, error) {
	if err := precompile.ValidateMethodInputsCount(inputs, 1); err != nil {
		return nil, err
	}

	sourceToken, ok := inputs[0].(common.Address)
	if !ok {
		return nil, fmt.Errorf("source token must be common.Address")
	}

	type mappingDescriptor struct {
		SourceToken common.Address
		MezoToken   common.Address
	}

	mapping, exists := m.bridgeKeeper.GetERC20TokenMapping(
		context.SdkCtx(),
		sourceToken.Hex(),
	)
	if !exists {
		return precompile.MethodOutputs{
			mappingDescriptor{
				SourceToken: common.Address{},
				MezoToken:   common.Address{},
			},
		}, nil
	}

	return precompile.MethodOutputs{
		mappingDescriptor{
			SourceToken: common.HexToAddress(mapping.SourceToken),
			MezoToken:   common.HexToAddress(mapping.MezoToken),
		},
	}, nil
}

// GetERC20TokensMappingsMethodName is the name of the getERC20TokensMappings method.
// It matches the name of the method in the contract ABI.
//
//nolint:gosec
const GetERC20TokensMappingsMethodName = "getERC20TokensMappings"

// GetERC20TokensMappingsMethod is the implementation of the getERC20TokensMappings method.
type GetERC20TokensMappingsMethod struct {
	bridgeKeeper BridgeKeeper
}

func newGetERC20TokensMappingsMethod(bridgeKeeper BridgeKeeper) *GetERC20TokensMappingsMethod {
	return &GetERC20TokensMappingsMethod{
		bridgeKeeper: bridgeKeeper,
	}
}

func (m *GetERC20TokensMappingsMethod) MethodName() string {
	return GetERC20TokensMappingsMethodName
}

func (m *GetERC20TokensMappingsMethod) MethodType() precompile.MethodType {
	return precompile.Read
}

func (m *GetERC20TokensMappingsMethod) RequiredGas(_ []byte) (uint64, bool) {
	// Fallback to the default gas calculation.
	return 0, false
}

func (m *GetERC20TokensMappingsMethod) Payable() bool {
	return false
}

func (m *GetERC20TokensMappingsMethod) Run(
	context *precompile.RunContext,
	inputs precompile.MethodInputs,
) (precompile.MethodOutputs, error) {
	if err := precompile.ValidateMethodInputsCount(inputs, 0); err != nil {
		return nil, err
	}

	type mappingDescriptor struct {
		SourceToken common.Address
		MezoToken   common.Address
	}

	mappings := make([]mappingDescriptor, 0)

	for _, mapping := range m.bridgeKeeper.GetERC20TokensMappings(context.SdkCtx()) {
		mappings = append(mappings, mappingDescriptor{
			SourceToken: common.HexToAddress(mapping.SourceToken),
			MezoToken:   common.HexToAddress(mapping.MezoToken),
		})
	}

	return precompile.MethodOutputs{mappings}, nil
}

// GetMaxERC20TokensMappingsMethodName is the name of the getMaxERC20TokensMappings method.
// It matches the name of the method in the contract ABI.
//
//nolint:gosec
const GetMaxERC20TokensMappingsMethodName = "getMaxERC20TokensMappings"

// GetMaxERC20TokensMappingsMethod is the implementation of the getMaxERC20TokensMappings method.
type GetMaxERC20TokensMappingsMethod struct {
	bridgeKeeper BridgeKeeper
}

func newGetMaxERC20TokensMappingsMethod(bridgeKeeper BridgeKeeper) *GetMaxERC20TokensMappingsMethod {
	return &GetMaxERC20TokensMappingsMethod{
		bridgeKeeper: bridgeKeeper,
	}
}

func (m *GetMaxERC20TokensMappingsMethod) MethodName() string {
	return GetMaxERC20TokensMappingsMethodName
}

func (m *GetMaxERC20TokensMappingsMethod) MethodType() precompile.MethodType {
	return precompile.Read
}

func (m *GetMaxERC20TokensMappingsMethod) RequiredGas(_ []byte) (uint64, bool) {
	// Fallback to the default gas calculation.
	return 0, false
}

func (m *GetMaxERC20TokensMappingsMethod) Payable() bool {
	return false
}

func (m *GetMaxERC20TokensMappingsMethod) Run(
	context *precompile.RunContext,
	inputs precompile.MethodInputs,
) (precompile.MethodOutputs, error) {
	if err := precompile.ValidateMethodInputsCount(inputs, 0); err != nil {
		return nil, err
	}

	params := m.bridgeKeeper.GetParams(context.SdkCtx())

	return precompile.MethodOutputs{big.NewInt(int64(params.MaxErc20TokensMappings))}, nil
}
