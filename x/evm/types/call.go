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

// ERC20BalanceOfCall represents a balanceOf(address) call for an ERC20 contract.
type ERC20BalanceOfCall struct {
	from, to common.Address
	data     []byte
}

// NewERC20BalanceOfCall creates a new ERC20BalanceOfCall.
func NewERC20BalanceOfCall(from, to, address []byte) (*ERC20BalanceOfCall, error) {
	addressType, err := abi.NewType("address", "", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create address type: %w", err)
	}

	uint256Type, err := abi.NewType("uint256", "", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create uint256 type: %w", err)
	}

	methodAbi := abi.Method{
		Name: "balanceOf",
		ID:   []byte{0x70, 0xa0, 0x82, 0x31}, // 0x70a08231 is the function selector for balanceOf(address)
		Type: abi.Function,
		Inputs: []abi.Argument{
			{Name: "account", Type: addressType},
		},
		Outputs: []abi.Argument{
			{Name: "balance", Type: uint256Type},
		},
	}
	contractAbi := abi.ABI{
		Methods: map[string]abi.Method{
			"balanceOf": methodAbi,
		},
	}

	data, err := contractAbi.Pack("balanceOf", common.BytesToAddress(address))
	if err != nil {
		return nil, fmt.Errorf("failed to pack balanceOf data: %w", err)
	}

	return &ERC20BalanceOfCall{
		from: common.BytesToAddress(from),
		to:   common.BytesToAddress(to),
		data: data,
	}, nil
}

func (c *ERC20BalanceOfCall) From() common.Address {
	return c.from
}

func (c *ERC20BalanceOfCall) To() *common.Address {
	return &c.to
}

func (c *ERC20BalanceOfCall) Data() []byte {
	return c.data
}

// ERC20AllowanceCall represents a allowance(address,address) call for an ERC20 contract.
type ERC20AllowanceCall struct {
	from, to common.Address
	data     []byte
}

// NewERC20AllowanceCall creates a new ERC20AllowanceCall.
func NewERC20AllowanceCall(from, to, owner, spender []byte) (*ERC20AllowanceCall, error) {
	addressType, err := abi.NewType("address", "", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create address type: %w", err)
	}

	uint256Type, err := abi.NewType("uint256", "", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create uint256 type: %w", err)
	}

	methodAbi := abi.Method{
		Name: "allowance",
		ID:   []byte{0xdd, 0x62, 0xed, 0x3e}, // 0xdd62ed3e is the function selector for allowance(address,address)
		Type: abi.Function,
		Inputs: []abi.Argument{
			{Name: "owner", Type: addressType},
			{Name: "spender", Type: addressType},
		},
		Outputs: []abi.Argument{
			{Name: "balance", Type: uint256Type},
		},
	}
	contractAbi := abi.ABI{
		Methods: map[string]abi.Method{
			"allowance": methodAbi,
		},
	}

	data, err := contractAbi.Pack("allowance", common.BytesToAddress(owner), common.BytesToAddress(spender))
	if err != nil {
		return nil, fmt.Errorf("failed to pack allowance data: %w", err)
	}

	return &ERC20AllowanceCall{
		from: common.BytesToAddress(from),
		to:   common.BytesToAddress(to),
		data: data,
	}, nil
}

func (c *ERC20AllowanceCall) From() common.Address {
	return c.from
}

func (c *ERC20AllowanceCall) To() *common.Address {
	return &c.to
}

func (c *ERC20AllowanceCall) Data() []byte {
	return c.data
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
