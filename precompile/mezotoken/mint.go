package mezotoken

import (
	"fmt"
	"math/big"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	"github.com/ethereum/go-ethereum/common"
	"github.com/mezo-org/mezod/precompile"
	evmkeeper "github.com/mezo-org/mezod/x/evm/keeper"
	evmtypes "github.com/mezo-org/mezod/x/evm/types"
)

const SetMinterMethodName = "setMinter"

// SetMinterMethod implements the setMinter method for the MEZO token precompile.
type SetMinterMethod struct {
	evmKeeper evmkeeper.Keeper
	poaKeeper PoaKeeper
}

// NewSetMinterMethod creates a new SetMinterMethod instance.
func NewSetMinterMethod(evmKeeper evmkeeper.Keeper, poaKeeper PoaKeeper) *SetMinterMethod {
	return &SetMinterMethod{
		evmKeeper: evmKeeper,
		poaKeeper: poaKeeper,
	}
}

// MethodName returns the name of the method.
func (m *SetMinterMethod) MethodName() string {
	return SetMinterMethodName
}

// MethodType returns the method type.
func (m *SetMinterMethod) MethodType() precompile.MethodType {
	return precompile.Write
}

// RequiredGas returns the gas required for the method.
func (m *SetMinterMethod) RequiredGas(_ []byte) (uint64, bool) {
	return 0, false
}

// Payable returns whether the method is payable.
func (m *SetMinterMethod) Payable() bool {
	return false
}

// Run executes the setMinter method.
func (m *SetMinterMethod) Run(
	context *precompile.RunContext,
	inputs precompile.MethodInputs,
) (precompile.MethodOutputs, error) {
	// Validate input count
	if err := precompile.ValidateMethodInputsCount(inputs, 1); err != nil {
		return nil, err
	}

	// Extract minter address
	minter, ok := inputs[0].(common.Address)
	if !ok {
		return nil, fmt.Errorf("minter argument must be common.Address")
	}

	// Check if sender is the POA owner
	sender := precompile.TypesConverter.Address.ToSDK(context.MsgSender())
	err := m.poaKeeper.CheckOwner(context.SdkCtx(), sender)
	if err != nil {
		return nil, err
	}

	// Update params with the new minter
	params := m.evmKeeper.GetParams(context.SdkCtx())
	params.MezoMinterAddress = minter.Hex()
	err = m.evmKeeper.SetParams(context.SdkCtx(), params)
	if err != nil {
		return nil, err
	}

	// Emit event
	err = context.EventEmitter().Emit(
		NewMinterSetEvent(minter),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to emit MinterSet event: [%w]", err)
	}

	return precompile.MethodOutputs{true}, nil
}

const MinterSetEventName = "MinterSet"

// MinterSetEvent represents the MinterSet event.
type MinterSetEvent struct {
	minter common.Address
}

// NewMinterSetEvent creates a new MinterSetEvent instance.
func NewMinterSetEvent(minter common.Address) *MinterSetEvent {
	return &MinterSetEvent{
		minter: minter,
	}
}

// EventName returns the event name.
func (e *MinterSetEvent) EventName() string {
	return MinterSetEventName
}

// Arguments returns the event arguments.
func (e *MinterSetEvent) Arguments() []*precompile.EventArgument {
	return []*precompile.EventArgument{
		{
			Indexed: true,
			Value:   e.minter,
		},
	}
}

const GetMinterMethodName = "getMinter"

// GetMinterMethod implements the getMinter method for the MEZO token precompile.
type GetMinterMethod struct {
	evmKeeper evmkeeper.Keeper
}

// NewGetMinterMethod creates a new GetMinterMethod instance.
func NewGetMinterMethod(evmKeeper evmkeeper.Keeper) *GetMinterMethod {
	return &GetMinterMethod{
		evmKeeper: evmKeeper,
	}
}

// MethodName returns the name of the method.
func (m *GetMinterMethod) MethodName() string {
	return GetMinterMethodName
}

// MethodType returns the method type.
func (m *GetMinterMethod) MethodType() precompile.MethodType {
	return precompile.Read
}

// RequiredGas returns the gas required for the method.
func (m *GetMinterMethod) RequiredGas(_ []byte) (uint64, bool) {
	return 0, false
}

// Payable returns whether the method is payable.
func (m *GetMinterMethod) Payable() bool {
	return false
}

// Run executes the getMinter method.
func (m *GetMinterMethod) Run(
	context *precompile.RunContext,
	inputs precompile.MethodInputs,
) (precompile.MethodOutputs, error) {
	// Validate input count (should be 0 for getMinter)
	if err := precompile.ValidateMethodInputsCount(inputs, 0); err != nil {
		return nil, err
	}

	// Get params and return the minter address
	params := m.evmKeeper.GetParams(context.SdkCtx())
	minterAddress := common.HexToAddress(params.MezoMinterAddress)

	return precompile.MethodOutputs{minterAddress}, nil
}

const MintMethodName = "mint"

// MintMethod implements the mint method for the MEZO token precompile.
type MintMethod struct {
	bankKeeper bankkeeper.Keeper
	evmKeeper  evmkeeper.Keeper
	denom      string
}

// NewMintMethod creates a new MintMethod instance.
func NewMintMethod(
	bankKeeper bankkeeper.Keeper,
	evmKeeper evmkeeper.Keeper,
	denom string,
) *MintMethod {
	return &MintMethod{
		bankKeeper: bankKeeper,
		evmKeeper:  evmKeeper,
		denom:      denom,
	}
}

// MethodName returns the name of the method.
func (m *MintMethod) MethodName() string {
	return MintMethodName
}

// MethodType returns the method type.
func (m *MintMethod) MethodType() precompile.MethodType {
	return precompile.Write
}

// RequiredGas returns the gas required for the method.
func (m *MintMethod) RequiredGas(_ []byte) (uint64, bool) {
	return 0, false
}

// Payable returns whether the method is payable.
func (m *MintMethod) Payable() bool {
	return false
}

// Run executes the mint method.
func (m *MintMethod) Run(
	context *precompile.RunContext,
	inputs precompile.MethodInputs,
) (precompile.MethodOutputs, error) {
	// Validate input count
	if err := precompile.ValidateMethodInputsCount(inputs, 2); err != nil {
		return nil, err
	}

	// Extract to address
	to, ok := inputs[0].(common.Address)
	if !ok {
		return nil, fmt.Errorf("to argument must be common.Address")
	}

	// Extract amount
	amount, ok := inputs[1].(*big.Int)
	if !ok {
		return nil, fmt.Errorf("amount argument must be *big.Int")
	}

	// Validate to address is not zero
	if to == (common.Address{}) {
		return nil, fmt.Errorf("cannot mint to zero address")
	}

	// Validate amount is positive
	if amount == nil || amount.Sign() <= 0 {
		return nil, fmt.Errorf("amount must be positive")
	}

	// Check if sender is the minter
	sender := context.MsgSender()
	params := m.evmKeeper.GetParams(context.SdkCtx())
	minterAddress := common.HexToAddress(params.MezoMinterAddress)

	if minterAddress == (common.Address{}) {
		return nil, fmt.Errorf("minter not set")
	}

	if sender != minterAddress {
		return nil, fmt.Errorf("sender is not the minter")
	}

	// Mint tokens to the module account
	coins := sdk.NewCoins(sdk.NewCoin(m.denom, sdkmath.NewIntFromBigInt(amount)))
	err := m.bankKeeper.MintCoins(context.SdkCtx(), evmtypes.ModuleName, coins)
	if err != nil {
		return nil, fmt.Errorf("failed to mint coins: [%w]", err)
	}

	// Transfer tokens from module account to recipient
	toSDK := precompile.TypesConverter.Address.ToSDK(to)
	err = m.bankKeeper.SendCoinsFromModuleToAccount(context.SdkCtx(), evmtypes.ModuleName, toSDK, coins)
	if err != nil {
		return nil, fmt.Errorf("failed to send coins: [%w]", err)
	}

	// Emit event
	err = context.EventEmitter().Emit(
		NewMintedEvent(to, amount),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to emit Minted event: [%w]", err)
	}

	return precompile.MethodOutputs{true}, nil
}

const MintedEventName = "Minted"

// MintedEvent represents the Minted event.
type MintedEvent struct {
	to     common.Address
	amount *big.Int
}

// NewMintedEvent creates a new MintedEvent instance.
func NewMintedEvent(to common.Address, amount *big.Int) *MintedEvent {
	return &MintedEvent{
		to:     to,
		amount: amount,
	}
}

// EventName returns the event name.
func (e *MintedEvent) EventName() string {
	return MintedEventName
}

// Arguments returns the event arguments.
func (e *MintedEvent) Arguments() []*precompile.EventArgument {
	return []*precompile.EventArgument{
		{
			Indexed: true,
			Value:   e.to,
		},
		{
			Indexed: false,
			Value:   e.amount,
		},
	}
}
