package types

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
)

type CallType int

const (
	// RPC call type is used on requests to eth_estimateGas rpc API endpoint
	RPC CallType = iota + 1
	// Internal call type is used in case of smart contract methods calls
	Internal
)

// ContractCall is an interface that represents a contract call.
type ContractCall interface {
	// From returns the address of the caller.
	From() common.Address
	// To returns the address of the contract being called.
	To() *common.Address
	// Data returns the data of the call.
	Data() []byte
	// GasLimit returns the gas limit for the call. If 0, the caller
	// should use the block's max gas limit as the default.
	GasLimit() uint64
}

// ERC20MintCall represents a mint(address,uint256) call for an ERC20 contract.
type ERC20MintCall struct {
	from, to common.Address
	data     []byte
}

// NewERC20MintCall creates a new ERC20MintCall.
func NewERC20MintCall(from, to, recipient []byte, amount *big.Int) (*ERC20MintCall, error) {
	addressType, err := abi.NewType("address", "", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create address type: %w", err)
	}

	uint256Type, err := abi.NewType("uint256", "", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create uint256 type: %w", err)
	}

	methodAbi := abi.Method{
		Name: "mint",
		ID:   []byte{0x40, 0xc1, 0x0f, 0x19}, // 0x40c10f19 is the function selector for mint(address,uint256)
		Type: abi.Function,
		Inputs: []abi.Argument{
			{Name: "account", Type: addressType},
			{Name: "amount", Type: uint256Type},
		},
		Outputs: []abi.Argument{},
	}
	contractAbi := abi.ABI{
		Methods: map[string]abi.Method{
			"mint": methodAbi,
		},
	}

	data, err := contractAbi.Pack("mint", common.BytesToAddress(recipient), amount)
	if err != nil {
		return nil, fmt.Errorf("failed to pack mint data: %w", err)
	}

	return &ERC20MintCall{
		from: common.BytesToAddress(from),
		to:   common.BytesToAddress(to),
		data: data,
	}, nil
}

func (c *ERC20MintCall) From() common.Address {
	return c.from
}

func (c *ERC20MintCall) To() *common.Address {
	return &c.to
}

func (c *ERC20MintCall) Data() []byte {
	return c.data
}

func (c *ERC20MintCall) GasLimit() uint64 {
	return 0
}

// ERC20BurnFromCall represents a burnFrom(address,uin256) call for an ERC20 contract.
type ERC20BurnFromCall struct {
	from, to common.Address
	data     []byte
}

// NewERC20BurnFromCall creates a new ERC20BurnFromCall.
func NewERC20BurnFromCall(from, to, address []byte, value *big.Int) (*ERC20BurnFromCall, error) {
	addressType, err := abi.NewType("address", "", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create address type: %w", err)
	}

	uint256Type, err := abi.NewType("uint256", "", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create uint256 type: %w", err)
	}

	methodAbi := abi.Method{
		Name: "burnFrom",
		ID:   []byte{0x79, 0xcc, 0x67, 0x90}, // 0x79cc6790 is the function selector for burnFrom(address,uint256)
		Type: abi.Function,
		Inputs: []abi.Argument{
			{Name: "account", Type: addressType},
			{Name: "value", Type: uint256Type},
		},
		Outputs: []abi.Argument{},
	}
	contractAbi := abi.ABI{
		Methods: map[string]abi.Method{
			"burnFrom": methodAbi,
		},
	}

	data, err := contractAbi.Pack("burnFrom", common.BytesToAddress(address), value)
	if err != nil {
		return nil, fmt.Errorf("failed to pack burnFrom data: %w", err)
	}

	return &ERC20BurnFromCall{
		from: common.BytesToAddress(from),
		to:   common.BytesToAddress(to),
		data: data,
	}, nil
}

func (c *ERC20BurnFromCall) From() common.Address {
	return c.from
}

func (c *ERC20BurnFromCall) To() *common.Address {
	return &c.to
}

func (c *ERC20BurnFromCall) Data() []byte {
	return c.data
}

func (c *ERC20BurnFromCall) GasLimit() uint64 {
	return 0
}

// TripartyCallbackGasLimit is the gas limit for the triparty callback call.
const TripartyCallbackGasLimit = uint64(1_000_000)

// TripartyCallbackCall represents an onTripartyBridgeCompleted(uint256,address,uint256,bytes)
// call issued by the PreBlocker to the controller after a triparty mint.
type TripartyCallbackCall struct {
	from, to common.Address
	data     []byte
}

// NewTripartyCallbackCall creates a new TripartyCallbackCall.
func NewTripartyCallbackCall(
	from, to []byte,
	requestID *big.Int,
	recipient []byte,
	amount *big.Int,
	callbackData []byte,
) (*TripartyCallbackCall, error) {
	addressType, err := abi.NewType("address", "", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create address type: %w", err)
	}

	uint256Type, err := abi.NewType("uint256", "", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create uint256 type: %w", err)
	}

	bytesType, err := abi.NewType("bytes", "", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create bytes type: %w", err)
	}

	// 0x2c144c16 is the function selector for
	// onTripartyBridgeCompleted(uint256,address,uint256,bytes)
	methodAbi := abi.Method{
		Name: "onTripartyBridgeCompleted",
		ID:   []byte{0x2c, 0x14, 0x4c, 0x16},
		Type: abi.Function,
		Inputs: []abi.Argument{
			{Name: "requestId", Type: uint256Type},
			{Name: "recipient", Type: addressType},
			{Name: "amount", Type: uint256Type},
			{Name: "callbackData", Type: bytesType},
		},
		Outputs: []abi.Argument{},
	}
	contractAbi := abi.ABI{
		Methods: map[string]abi.Method{
			"onTripartyBridgeCompleted": methodAbi,
		},
	}

	if callbackData == nil {
		callbackData = []byte{}
	}

	data, err := contractAbi.Pack(
		"onTripartyBridgeCompleted",
		requestID,
		common.BytesToAddress(recipient),
		amount,
		callbackData,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to pack triparty callback data: %w", err)
	}

	return &TripartyCallbackCall{
		from: common.BytesToAddress(from),
		to:   common.BytesToAddress(to),
		data: data,
	}, nil
}

func (c *TripartyCallbackCall) From() common.Address {
	return c.from
}

func (c *TripartyCallbackCall) To() *common.Address {
	return &c.to
}

func (c *TripartyCallbackCall) Data() []byte {
	return c.data
}

func (c *TripartyCallbackCall) GasLimit() uint64 {
	return TripartyCallbackGasLimit
}
