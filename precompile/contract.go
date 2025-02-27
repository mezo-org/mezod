package precompile

import (
	"bytes"
	"embed"
	"fmt"
	"math/big"

	store "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/tracing"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/holiman/uint256"
	"github.com/mezo-org/mezod/x/evm/statedb"
)

var _ vm.PrecompiledContract = &Contract{}

// Contract represents a precompiled contract that can be executed by the EVM.
type Contract struct {
	Abi      abi.ABI
	address  common.Address
	bytecode string
	methods  map[string]Method
}

type StateDBJournal struct {
	entries []journalEntry
}

type journalEntry struct {
	Address       common.Address
	Amount        *uint256.Int
	TracingReason tracing.BalanceChangeReason
	isSub         bool
}

func (c *StateDBJournal) AddBalance(
	address common.Address,
	amount *uint256.Int,
	tracingReason tracing.BalanceChangeReason,
) {
	c.entries = append(
		c.entries,
		journalEntry{
			Address:       address,
			Amount:        amount,
			TracingReason: tracingReason,
			isSub:         false,
		},
	)
}

func (c *StateDBJournal) SubBalance(
	address common.Address,
	amount *uint256.Int,
	tracingReason tracing.BalanceChangeReason,
) {
	c.entries = append(
		c.entries,
		journalEntry{
			Address:       address,
			Amount:        amount,
			TracingReason: tracingReason,
			isSub:         true,
		},
	)
}

// NewContract creates a new precompiled contract with the given ABI and address.
func NewContract(abi abi.ABI, address common.Address, bytecode string) *Contract {
	return &Contract{
		Abi:      abi,
		address:  address,
		bytecode: bytecode,
		methods:  make(map[string]Method),
	}
}

// RegisterMethods registers the given methods in the contract. If a method with
// the same name already exists, it will be overwritten. This function does not
// check whether the registered method name exists in the ABI - if not, the
// method will not be available for callers.
func (c *Contract) RegisterMethods(methods ...Method) *Contract {
	for _, method := range methods {
		c.methods[method.MethodName()] = method
	}

	return c
}

// Address returns the EVM address of the contract.
func (c *Contract) Address() common.Address {
	return c.address
}

// RequiredGas returns the amount of gas required to execute the contract call
// with the given input. If the target method does not determine the required
// gas by itself, the required gas is calculated based on the default algorithm
// that implies a flat cost plus additional cost based on the input size.
func (c *Contract) RequiredGas(input []byte) uint64 {
	methodID, methodInputArgs, err := c.parseCallInput(input)
	if err != nil {
		// Panic is unacceptable here, return 0 instead.
		return 0
	}

	method, _, err := c.methodByID(methodID)
	if err != nil {
		// Panic is unacceptable here, return 0 instead.
		return 0
	}

	requiredGas, ok := method.RequiredGas(methodInputArgs)
	if !ok {
		// Fall back to default required gas if method does not determine
		// the required gas by itself.
		requiredGas = DefaultRequiredGas(
			store.KVGasConfig(),
			method.MethodType(),
			methodInputArgs,
		)
	}

	return requiredGas
}

// Run executes a contract call with the given input. The call is executed in
// the context of the given EVM and contract instances. The right precompiled
// method is determined based on the call input and the contract's ABI.
// The execution can be performed in read-only mode, which means that only
// Read methods can be executed. The call can also transfer value but the
// target method must be of type Write and support payable calls. Output
// arguments of the called method are returned as a byte slice.
func (c *Contract) Run(
	evm *vm.EVM,
	contract *vm.Contract,
	readOnlyMode bool,
) (methodOutputArgs []byte, runErr error) {
	// The x/evm module provides its own state DB implementation, which
	// reflects EVM state changes in the Cosmos SDK store upon commit.
	// We can safely cast the EVM state DB to the *statedb.StateDB type
	// because we know this is the type EVM instance was created with.
	stateDB, ok := evm.StateDB.(*statedb.StateDB)
	if !ok {
		return nil, fmt.Errorf("cannot get state DB from EVM")
	}

	sdkCtx, ctxCheckpoint := stateDB.CacheContext()

	// Capture the initial values of gas config to restore them after execution.
	kvGasConfig, transientKVGasConfig := sdkCtx.KVGasConfig(), sdkCtx.TransientKVGasConfig()
	// Use a zero gas config for Cosmos SDK operations to avoid extra costs
	// apart the RequiredGas already consumed on the EVM level.
	zeroGasConfig := store.GasConfig{}
	sdkCtx = sdkCtx.
		WithKVGasConfig(zeroGasConfig).
		WithTransientKVGasConfig(zeroGasConfig)
	// Set a deferred function to restore the initial gas config values
	// after the method execution. This action is not strictly necessary
	// as gas config changes are applied to a copy of the original SDK
	// context and are not propagated back. However, making this cleanup
	// just in case may allow to avoid potential issues in the future,
	// in case the copied context is used in some other way.
	defer func() {
		sdkCtx = sdkCtx.
			WithKVGasConfig(kvGasConfig).
			WithTransientKVGasConfig(transientKVGasConfig)
	}()

	eventEmitter := NewEventEmitter(sdkCtx, c.Abi, c.address, stateDB)
	runCtx := NewRunContext(sdkCtx, evm, contract, eventEmitter)

	methodID, methodInputArgs, err := c.parseCallInput(contract.Input)
	if err != nil {
		return nil, fmt.Errorf("failed to parse contract input: [%w]", err)
	}

	method, methodABI, err := c.methodByID(methodID)
	if err != nil {
		return nil, fmt.Errorf("failed to get method by ID: [%w]", err)
	}

	// Execute some validation based on method type.
	switch method.MethodType() {
	case Read:
		// Read methods can be executed regardless of the mode but can
		// never accept value.
		if runCtx.IsMsgValue() {
			return nil, fmt.Errorf("read method cannot accept value")
		}
	case Write:
		// Write methods cannot be executed in read-only mode and can accept
		// value if the specific method supports it.
		if readOnlyMode {
			return nil, fmt.Errorf("write method cannot be executed in read-only mode")
		}
		if runCtx.IsMsgValue() && !method.Payable() {
			return nil, fmt.Errorf("non-payable write method cannot accept value")
		}
	default:
		panic("unexpected method type")
	}

	// Commit any draft changes to the EVM state DB before running the method.
	if err := stateDB.CommitCacheContext(); err != nil {
		return nil, err
	}

	methodInputs, err := methodABI.Inputs.Unpack(methodInputArgs)
	if err != nil {
		return nil, fmt.Errorf("failed to unpack method input args: [%w]", err)
	}

	methodOutputs, err := method.Run(runCtx, methodInputs)
	if err != nil {
		return nil, fmt.Errorf("method errored out: [%w]", err)
	}

	methodOutputArgs, err = methodABI.Outputs.Pack(methodOutputs...)
	if err != nil {
		return nil, fmt.Errorf("failed to pack method output args: [%w]", err)
	}

	// now if nothing failed, we executed the journal entries against the stateDB
	c.syncJournalEntries(runCtx.journal, c.Address(), ctxCheckpoint, stateDB)

	return methodOutputArgs, nil
}

func (c *Contract) syncJournalEntries(
	journal *StateDBJournal,
	address common.Address,
	cacheCtxCheckpoint *statedb.CachedCtxCheckpoint,
	stateDB *statedb.StateDB,
) {
	// fist save the checkpoint
	stateDB.RegisterCachedContextCheckpoint(address, cacheCtxCheckpoint)

	for _, v := range journal.entries {
		if v.isSub {
			stateDB.SubBalance(v.Address, v.Amount, v.TracingReason)
			continue
		}

		stateDB.AddBalance(v.Address, v.Amount, v.TracingReason)
	}

	journal.entries = nil
}

// parseCallInput extracts the method ID and input arguments from the given
// input byte slice. The method ID is expected to be the first 4 bytes of the
// input. If the input is shorter than 4 bytes, an error is returned.
func (c *Contract) parseCallInput(input []byte) ([]byte, []byte, error) {
	// We expect a proper method call. In any other case, prevent the bounds
	// out of range panic and return an error.
	if len(input) < methodIDByteLength {
		return nil, nil, fmt.Errorf("input is shorter than method ID length")
	}

	methodID := input[:methodIDByteLength]
	methodInputArgs := input[methodIDByteLength:]

	return methodID, methodInputArgs, nil
}

// methodByID returns the precompiled method and the corresponding ABI method
// based on the given method ID. If the method is not found in the ABI or is
// not registered in the precompiled contract, an error is returned.
func (c *Contract) methodByID(methodID []byte) (Method, *abi.Method, error) {
	methodABI, err := c.Abi.MethodById(methodID)
	if err != nil {
		return nil, nil, fmt.Errorf("method not found in ABI: [%w]", err)
	}

	// If the method is not a regular function, return an error because it's
	// either a constructor, a fallback, or a receive function.
	if abiType := methodABI.Type; abiType != abi.Function {
		return nil, nil, fmt.Errorf(
			"unexpected method type: [%v]",
			abiType,
		)
	}

	method, ok := c.methods[methodABI.Name]
	if !ok {
		return nil, nil, fmt.Errorf("method not found in precompile")
	}

	return method, methodABI, nil
}

// Bytecode returns the EVM bytecode of the contract.
func (c *Contract) Bytecode() string {
	return c.bytecode
}

// RunContext represents the context in which a precompiled contract method is
// executed. It provides access to the EVM, the contract, and the event emitter.
type RunContext struct {
	sdkCtx       sdk.Context
	evm          *vm.EVM
	contract     *vm.Contract
	eventEmitter *EventEmitter
	journal      *StateDBJournal
}

// NewRunContext creates a new run context with the given EVM, contract, and
// event emitter instances.
func NewRunContext(
	sdkCtx sdk.Context,
	evm *vm.EVM,
	contract *vm.Contract,
	eventEmitter *EventEmitter,
) *RunContext {
	return &RunContext{
		sdkCtx:       sdkCtx,
		evm:          evm,
		contract:     contract,
		eventEmitter: eventEmitter,
		journal:      &StateDBJournal{},
	}
}

func (rc *RunContext) Journal() *StateDBJournal {
	return rc.journal
}

// SdkCtx returns the Cosmos SDK context associated with the run context.
func (rc *RunContext) SdkCtx() sdk.Context {
	return rc.sdkCtx
}

// MsgSender returns the address of the message sender. This corresponds to the
// msg.sender in Solidity.
func (rc *RunContext) MsgSender() common.Address {
	return rc.contract.Caller()
}

// TxOrigin returns the address of the transaction originator. This corresponds
// to the tx.origin in Solidity.
func (rc *RunContext) TxOrigin() common.Address {
	return rc.evm.Origin
}

// MsgValue returns the value sent with the message. This corresponds to the
// msg.value in Solidity.
func (rc *RunContext) MsgValue() *big.Int {
	if value := rc.contract.Value(); value != nil {
		return value.ToBig()
	}

	return big.NewInt(0)
}

// IsMsgValue returns true if the message value is greater than zero.
func (rc *RunContext) IsMsgValue() bool {
	return rc.MsgValue().Sign() > 0
}

// EventEmitter returns the event emitter instance associated with the run context.
// The event emitter can be used to emit EVM events from the precompiled contract.
func (rc *RunContext) EventEmitter() *EventEmitter {
	return rc.eventEmitter
}

// LoadAbiFile loads the ABI file from the given file system and path. The ABI file
// is expected to be in JSON format. If the file cannot be loaded or parsed, an
// error is returned.
func LoadAbiFile(fs embed.FS, path string) (abi.ABI, error) {
	abiBytes, err := fs.ReadFile(path)
	if err != nil {
		return abi.ABI{}, fmt.Errorf("cannot load ABI file [%w]", err)
	}

	abiJSON, err := abi.JSON(bytes.NewReader(abiBytes))
	if err != nil {
		return abi.ABI{}, fmt.Errorf("cannot parse ABI file [%w]", err)
	}

	return abiJSON, nil
}
