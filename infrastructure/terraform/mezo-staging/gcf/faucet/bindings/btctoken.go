// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package btctoken

import (
	"errors"
	"math/big"
	"strings"

	ethereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"
)

// Reference imports to suppress errors if they are not otherwise used.
var (
	_ = errors.New
	_ = big.NewInt
	_ = strings.NewReader
	_ = ethereum.NotFound
	_ = bind.Bind
	_ = common.Big1
	_ = types.BloomLookup
	_ = event.NewSubscription
	_ = abi.ConvertType
)

// BtctokenMetaData contains all meta data concerning the Btctoken contract.
var BtctokenMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"spender\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"Approval\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"from\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"Transfer\",\"type\":\"event\"},{\"inputs\":[],\"name\":\"DOMAIN_SEPARATOR\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"PERMIT_TYPEHASH\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"name\":\"allowance\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"spender\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"approve\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"name\":\"balanceOf\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"decimals\",\"outputs\":[{\"internalType\":\"uint8\",\"name\":\"\",\"type\":\"uint8\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"name\",\"outputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"name\":\"nonce\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"spender\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"deadline\",\"type\":\"uint256\"},{\"internalType\":\"uint8\",\"name\":\"v\",\"type\":\"uint8\"},{\"internalType\":\"bytes32\",\"name\":\"r\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"s\",\"type\":\"bytes32\"}],\"name\":\"permit\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"symbol\",\"outputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"totalSupply\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"recipient\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"transfer\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"sender\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"recipient\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"transferFrom\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]",
}

// BtctokenABI is the input ABI used to generate the binding from.
// Deprecated: Use BtctokenMetaData.ABI instead.
var BtctokenABI = BtctokenMetaData.ABI

// Btctoken is an auto generated Go binding around an Ethereum contract.
type Btctoken struct {
	BtctokenCaller     // Read-only binding to the contract
	BtctokenTransactor // Write-only binding to the contract
	BtctokenFilterer   // Log filterer for contract events
}

// BtctokenCaller is an auto generated read-only Go binding around an Ethereum contract.
type BtctokenCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// BtctokenTransactor is an auto generated write-only Go binding around an Ethereum contract.
type BtctokenTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// BtctokenFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type BtctokenFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// BtctokenSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type BtctokenSession struct {
	Contract     *Btctoken         // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// BtctokenCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type BtctokenCallerSession struct {
	Contract *BtctokenCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts   // Call options to use throughout this session
}

// BtctokenTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type BtctokenTransactorSession struct {
	Contract     *BtctokenTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts   // Transaction auth options to use throughout this session
}

// BtctokenRaw is an auto generated low-level Go binding around an Ethereum contract.
type BtctokenRaw struct {
	Contract *Btctoken // Generic contract binding to access the raw methods on
}

// BtctokenCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type BtctokenCallerRaw struct {
	Contract *BtctokenCaller // Generic read-only contract binding to access the raw methods on
}

// BtctokenTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type BtctokenTransactorRaw struct {
	Contract *BtctokenTransactor // Generic write-only contract binding to access the raw methods on
}

// NewBtctoken creates a new instance of Btctoken, bound to a specific deployed contract.
func NewBtctoken(address common.Address, backend bind.ContractBackend) (*Btctoken, error) {
	contract, err := bindBtctoken(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &Btctoken{BtctokenCaller: BtctokenCaller{contract: contract}, BtctokenTransactor: BtctokenTransactor{contract: contract}, BtctokenFilterer: BtctokenFilterer{contract: contract}}, nil
}

// NewBtctokenCaller creates a new read-only instance of Btctoken, bound to a specific deployed contract.
func NewBtctokenCaller(address common.Address, caller bind.ContractCaller) (*BtctokenCaller, error) {
	contract, err := bindBtctoken(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &BtctokenCaller{contract: contract}, nil
}

// NewBtctokenTransactor creates a new write-only instance of Btctoken, bound to a specific deployed contract.
func NewBtctokenTransactor(address common.Address, transactor bind.ContractTransactor) (*BtctokenTransactor, error) {
	contract, err := bindBtctoken(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &BtctokenTransactor{contract: contract}, nil
}

// NewBtctokenFilterer creates a new log filterer instance of Btctoken, bound to a specific deployed contract.
func NewBtctokenFilterer(address common.Address, filterer bind.ContractFilterer) (*BtctokenFilterer, error) {
	contract, err := bindBtctoken(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &BtctokenFilterer{contract: contract}, nil
}

// bindBtctoken binds a generic wrapper to an already deployed contract.
func bindBtctoken(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := BtctokenMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Btctoken *BtctokenRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Btctoken.Contract.BtctokenCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Btctoken *BtctokenRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Btctoken.Contract.BtctokenTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Btctoken *BtctokenRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Btctoken.Contract.BtctokenTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Btctoken *BtctokenCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Btctoken.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Btctoken *BtctokenTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Btctoken.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Btctoken *BtctokenTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Btctoken.Contract.contract.Transact(opts, method, params...)
}

// DOMAINSEPARATOR is a free data retrieval call binding the contract method 0x3644e515.
//
// Solidity: function DOMAIN_SEPARATOR() view returns(bytes32)
func (_Btctoken *BtctokenCaller) DOMAINSEPARATOR(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _Btctoken.contract.Call(opts, &out, "DOMAIN_SEPARATOR")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// DOMAINSEPARATOR is a free data retrieval call binding the contract method 0x3644e515.
//
// Solidity: function DOMAIN_SEPARATOR() view returns(bytes32)
func (_Btctoken *BtctokenSession) DOMAINSEPARATOR() ([32]byte, error) {
	return _Btctoken.Contract.DOMAINSEPARATOR(&_Btctoken.CallOpts)
}

// DOMAINSEPARATOR is a free data retrieval call binding the contract method 0x3644e515.
//
// Solidity: function DOMAIN_SEPARATOR() view returns(bytes32)
func (_Btctoken *BtctokenCallerSession) DOMAINSEPARATOR() ([32]byte, error) {
	return _Btctoken.Contract.DOMAINSEPARATOR(&_Btctoken.CallOpts)
}

// PERMITTYPEHASH is a free data retrieval call binding the contract method 0x30adf81f.
//
// Solidity: function PERMIT_TYPEHASH() view returns(bytes32)
func (_Btctoken *BtctokenCaller) PERMITTYPEHASH(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _Btctoken.contract.Call(opts, &out, "PERMIT_TYPEHASH")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// PERMITTYPEHASH is a free data retrieval call binding the contract method 0x30adf81f.
//
// Solidity: function PERMIT_TYPEHASH() view returns(bytes32)
func (_Btctoken *BtctokenSession) PERMITTYPEHASH() ([32]byte, error) {
	return _Btctoken.Contract.PERMITTYPEHASH(&_Btctoken.CallOpts)
}

// PERMITTYPEHASH is a free data retrieval call binding the contract method 0x30adf81f.
//
// Solidity: function PERMIT_TYPEHASH() view returns(bytes32)
func (_Btctoken *BtctokenCallerSession) PERMITTYPEHASH() ([32]byte, error) {
	return _Btctoken.Contract.PERMITTYPEHASH(&_Btctoken.CallOpts)
}

// Allowance is a free data retrieval call binding the contract method 0xdd62ed3e.
//
// Solidity: function allowance(address , address ) view returns(uint256)
func (_Btctoken *BtctokenCaller) Allowance(opts *bind.CallOpts, arg0 common.Address, arg1 common.Address) (*big.Int, error) {
	var out []interface{}
	err := _Btctoken.contract.Call(opts, &out, "allowance", arg0, arg1)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// Allowance is a free data retrieval call binding the contract method 0xdd62ed3e.
//
// Solidity: function allowance(address , address ) view returns(uint256)
func (_Btctoken *BtctokenSession) Allowance(arg0 common.Address, arg1 common.Address) (*big.Int, error) {
	return _Btctoken.Contract.Allowance(&_Btctoken.CallOpts, arg0, arg1)
}

// Allowance is a free data retrieval call binding the contract method 0xdd62ed3e.
//
// Solidity: function allowance(address , address ) view returns(uint256)
func (_Btctoken *BtctokenCallerSession) Allowance(arg0 common.Address, arg1 common.Address) (*big.Int, error) {
	return _Btctoken.Contract.Allowance(&_Btctoken.CallOpts, arg0, arg1)
}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf(address ) view returns(uint256)
func (_Btctoken *BtctokenCaller) BalanceOf(opts *bind.CallOpts, arg0 common.Address) (*big.Int, error) {
	var out []interface{}
	err := _Btctoken.contract.Call(opts, &out, "balanceOf", arg0)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf(address ) view returns(uint256)
func (_Btctoken *BtctokenSession) BalanceOf(arg0 common.Address) (*big.Int, error) {
	return _Btctoken.Contract.BalanceOf(&_Btctoken.CallOpts, arg0)
}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf(address ) view returns(uint256)
func (_Btctoken *BtctokenCallerSession) BalanceOf(arg0 common.Address) (*big.Int, error) {
	return _Btctoken.Contract.BalanceOf(&_Btctoken.CallOpts, arg0)
}

// Decimals is a free data retrieval call binding the contract method 0x313ce567.
//
// Solidity: function decimals() view returns(uint8)
func (_Btctoken *BtctokenCaller) Decimals(opts *bind.CallOpts) (uint8, error) {
	var out []interface{}
	err := _Btctoken.contract.Call(opts, &out, "decimals")

	if err != nil {
		return *new(uint8), err
	}

	out0 := *abi.ConvertType(out[0], new(uint8)).(*uint8)

	return out0, err

}

// Decimals is a free data retrieval call binding the contract method 0x313ce567.
//
// Solidity: function decimals() view returns(uint8)
func (_Btctoken *BtctokenSession) Decimals() (uint8, error) {
	return _Btctoken.Contract.Decimals(&_Btctoken.CallOpts)
}

// Decimals is a free data retrieval call binding the contract method 0x313ce567.
//
// Solidity: function decimals() view returns(uint8)
func (_Btctoken *BtctokenCallerSession) Decimals() (uint8, error) {
	return _Btctoken.Contract.Decimals(&_Btctoken.CallOpts)
}

// Name is a free data retrieval call binding the contract method 0x06fdde03.
//
// Solidity: function name() view returns(string)
func (_Btctoken *BtctokenCaller) Name(opts *bind.CallOpts) (string, error) {
	var out []interface{}
	err := _Btctoken.contract.Call(opts, &out, "name")

	if err != nil {
		return *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)

	return out0, err

}

// Name is a free data retrieval call binding the contract method 0x06fdde03.
//
// Solidity: function name() view returns(string)
func (_Btctoken *BtctokenSession) Name() (string, error) {
	return _Btctoken.Contract.Name(&_Btctoken.CallOpts)
}

// Name is a free data retrieval call binding the contract method 0x06fdde03.
//
// Solidity: function name() view returns(string)
func (_Btctoken *BtctokenCallerSession) Name() (string, error) {
	return _Btctoken.Contract.Name(&_Btctoken.CallOpts)
}

// Nonce is a free data retrieval call binding the contract method 0x70ae92d2.
//
// Solidity: function nonce(address ) view returns(uint256)
func (_Btctoken *BtctokenCaller) Nonce(opts *bind.CallOpts, arg0 common.Address) (*big.Int, error) {
	var out []interface{}
	err := _Btctoken.contract.Call(opts, &out, "nonce", arg0)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// Nonce is a free data retrieval call binding the contract method 0x70ae92d2.
//
// Solidity: function nonce(address ) view returns(uint256)
func (_Btctoken *BtctokenSession) Nonce(arg0 common.Address) (*big.Int, error) {
	return _Btctoken.Contract.Nonce(&_Btctoken.CallOpts, arg0)
}

// Nonce is a free data retrieval call binding the contract method 0x70ae92d2.
//
// Solidity: function nonce(address ) view returns(uint256)
func (_Btctoken *BtctokenCallerSession) Nonce(arg0 common.Address) (*big.Int, error) {
	return _Btctoken.Contract.Nonce(&_Btctoken.CallOpts, arg0)
}

// Symbol is a free data retrieval call binding the contract method 0x95d89b41.
//
// Solidity: function symbol() view returns(string)
func (_Btctoken *BtctokenCaller) Symbol(opts *bind.CallOpts) (string, error) {
	var out []interface{}
	err := _Btctoken.contract.Call(opts, &out, "symbol")

	if err != nil {
		return *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)

	return out0, err

}

// Symbol is a free data retrieval call binding the contract method 0x95d89b41.
//
// Solidity: function symbol() view returns(string)
func (_Btctoken *BtctokenSession) Symbol() (string, error) {
	return _Btctoken.Contract.Symbol(&_Btctoken.CallOpts)
}

// Symbol is a free data retrieval call binding the contract method 0x95d89b41.
//
// Solidity: function symbol() view returns(string)
func (_Btctoken *BtctokenCallerSession) Symbol() (string, error) {
	return _Btctoken.Contract.Symbol(&_Btctoken.CallOpts)
}

// TotalSupply is a free data retrieval call binding the contract method 0x18160ddd.
//
// Solidity: function totalSupply() view returns(uint256)
func (_Btctoken *BtctokenCaller) TotalSupply(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _Btctoken.contract.Call(opts, &out, "totalSupply")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// TotalSupply is a free data retrieval call binding the contract method 0x18160ddd.
//
// Solidity: function totalSupply() view returns(uint256)
func (_Btctoken *BtctokenSession) TotalSupply() (*big.Int, error) {
	return _Btctoken.Contract.TotalSupply(&_Btctoken.CallOpts)
}

// TotalSupply is a free data retrieval call binding the contract method 0x18160ddd.
//
// Solidity: function totalSupply() view returns(uint256)
func (_Btctoken *BtctokenCallerSession) TotalSupply() (*big.Int, error) {
	return _Btctoken.Contract.TotalSupply(&_Btctoken.CallOpts)
}

// Approve is a paid mutator transaction binding the contract method 0x095ea7b3.
//
// Solidity: function approve(address spender, uint256 amount) returns(bool)
func (_Btctoken *BtctokenTransactor) Approve(opts *bind.TransactOpts, spender common.Address, amount *big.Int) (*types.Transaction, error) {
	return _Btctoken.contract.Transact(opts, "approve", spender, amount)
}

// Approve is a paid mutator transaction binding the contract method 0x095ea7b3.
//
// Solidity: function approve(address spender, uint256 amount) returns(bool)
func (_Btctoken *BtctokenSession) Approve(spender common.Address, amount *big.Int) (*types.Transaction, error) {
	return _Btctoken.Contract.Approve(&_Btctoken.TransactOpts, spender, amount)
}

// Approve is a paid mutator transaction binding the contract method 0x095ea7b3.
//
// Solidity: function approve(address spender, uint256 amount) returns(bool)
func (_Btctoken *BtctokenTransactorSession) Approve(spender common.Address, amount *big.Int) (*types.Transaction, error) {
	return _Btctoken.Contract.Approve(&_Btctoken.TransactOpts, spender, amount)
}

// Permit is a paid mutator transaction binding the contract method 0xd505accf.
//
// Solidity: function permit(address owner, address spender, uint256 amount, uint256 deadline, uint8 v, bytes32 r, bytes32 s) returns(bool)
func (_Btctoken *BtctokenTransactor) Permit(opts *bind.TransactOpts, owner common.Address, spender common.Address, amount *big.Int, deadline *big.Int, v uint8, r [32]byte, s [32]byte) (*types.Transaction, error) {
	return _Btctoken.contract.Transact(opts, "permit", owner, spender, amount, deadline, v, r, s)
}

// Permit is a paid mutator transaction binding the contract method 0xd505accf.
//
// Solidity: function permit(address owner, address spender, uint256 amount, uint256 deadline, uint8 v, bytes32 r, bytes32 s) returns(bool)
func (_Btctoken *BtctokenSession) Permit(owner common.Address, spender common.Address, amount *big.Int, deadline *big.Int, v uint8, r [32]byte, s [32]byte) (*types.Transaction, error) {
	return _Btctoken.Contract.Permit(&_Btctoken.TransactOpts, owner, spender, amount, deadline, v, r, s)
}

// Permit is a paid mutator transaction binding the contract method 0xd505accf.
//
// Solidity: function permit(address owner, address spender, uint256 amount, uint256 deadline, uint8 v, bytes32 r, bytes32 s) returns(bool)
func (_Btctoken *BtctokenTransactorSession) Permit(owner common.Address, spender common.Address, amount *big.Int, deadline *big.Int, v uint8, r [32]byte, s [32]byte) (*types.Transaction, error) {
	return _Btctoken.Contract.Permit(&_Btctoken.TransactOpts, owner, spender, amount, deadline, v, r, s)
}

// Transfer is a paid mutator transaction binding the contract method 0xa9059cbb.
//
// Solidity: function transfer(address recipient, uint256 amount) returns(bool)
func (_Btctoken *BtctokenTransactor) Transfer(opts *bind.TransactOpts, recipient common.Address, amount *big.Int) (*types.Transaction, error) {
	return _Btctoken.contract.Transact(opts, "transfer", recipient, amount)
}

// Transfer is a paid mutator transaction binding the contract method 0xa9059cbb.
//
// Solidity: function transfer(address recipient, uint256 amount) returns(bool)
func (_Btctoken *BtctokenSession) Transfer(recipient common.Address, amount *big.Int) (*types.Transaction, error) {
	return _Btctoken.Contract.Transfer(&_Btctoken.TransactOpts, recipient, amount)
}

// Transfer is a paid mutator transaction binding the contract method 0xa9059cbb.
//
// Solidity: function transfer(address recipient, uint256 amount) returns(bool)
func (_Btctoken *BtctokenTransactorSession) Transfer(recipient common.Address, amount *big.Int) (*types.Transaction, error) {
	return _Btctoken.Contract.Transfer(&_Btctoken.TransactOpts, recipient, amount)
}

// TransferFrom is a paid mutator transaction binding the contract method 0x23b872dd.
//
// Solidity: function transferFrom(address sender, address recipient, uint256 amount) returns(bool)
func (_Btctoken *BtctokenTransactor) TransferFrom(opts *bind.TransactOpts, sender common.Address, recipient common.Address, amount *big.Int) (*types.Transaction, error) {
	return _Btctoken.contract.Transact(opts, "transferFrom", sender, recipient, amount)
}

// TransferFrom is a paid mutator transaction binding the contract method 0x23b872dd.
//
// Solidity: function transferFrom(address sender, address recipient, uint256 amount) returns(bool)
func (_Btctoken *BtctokenSession) TransferFrom(sender common.Address, recipient common.Address, amount *big.Int) (*types.Transaction, error) {
	return _Btctoken.Contract.TransferFrom(&_Btctoken.TransactOpts, sender, recipient, amount)
}

// TransferFrom is a paid mutator transaction binding the contract method 0x23b872dd.
//
// Solidity: function transferFrom(address sender, address recipient, uint256 amount) returns(bool)
func (_Btctoken *BtctokenTransactorSession) TransferFrom(sender common.Address, recipient common.Address, amount *big.Int) (*types.Transaction, error) {
	return _Btctoken.Contract.TransferFrom(&_Btctoken.TransactOpts, sender, recipient, amount)
}

// BtctokenApprovalIterator is returned from FilterApproval and is used to iterate over the raw logs and unpacked data for Approval events raised by the Btctoken contract.
type BtctokenApprovalIterator struct {
	Event *BtctokenApproval // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *BtctokenApprovalIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BtctokenApproval)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(BtctokenApproval)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *BtctokenApprovalIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BtctokenApprovalIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BtctokenApproval represents a Approval event raised by the Btctoken contract.
type BtctokenApproval struct {
	Owner   common.Address
	Spender common.Address
	Value   *big.Int
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterApproval is a free log retrieval operation binding the contract event 0x8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925.
//
// Solidity: event Approval(address indexed owner, address indexed spender, uint256 value)
func (_Btctoken *BtctokenFilterer) FilterApproval(opts *bind.FilterOpts, owner []common.Address, spender []common.Address) (*BtctokenApprovalIterator, error) {

	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}
	var spenderRule []interface{}
	for _, spenderItem := range spender {
		spenderRule = append(spenderRule, spenderItem)
	}

	logs, sub, err := _Btctoken.contract.FilterLogs(opts, "Approval", ownerRule, spenderRule)
	if err != nil {
		return nil, err
	}
	return &BtctokenApprovalIterator{contract: _Btctoken.contract, event: "Approval", logs: logs, sub: sub}, nil
}

// WatchApproval is a free log subscription operation binding the contract event 0x8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925.
//
// Solidity: event Approval(address indexed owner, address indexed spender, uint256 value)
func (_Btctoken *BtctokenFilterer) WatchApproval(opts *bind.WatchOpts, sink chan<- *BtctokenApproval, owner []common.Address, spender []common.Address) (event.Subscription, error) {

	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}
	var spenderRule []interface{}
	for _, spenderItem := range spender {
		spenderRule = append(spenderRule, spenderItem)
	}

	logs, sub, err := _Btctoken.contract.WatchLogs(opts, "Approval", ownerRule, spenderRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BtctokenApproval)
				if err := _Btctoken.contract.UnpackLog(event, "Approval", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseApproval is a log parse operation binding the contract event 0x8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925.
//
// Solidity: event Approval(address indexed owner, address indexed spender, uint256 value)
func (_Btctoken *BtctokenFilterer) ParseApproval(log types.Log) (*BtctokenApproval, error) {
	event := new(BtctokenApproval)
	if err := _Btctoken.contract.UnpackLog(event, "Approval", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BtctokenTransferIterator is returned from FilterTransfer and is used to iterate over the raw logs and unpacked data for Transfer events raised by the Btctoken contract.
type BtctokenTransferIterator struct {
	Event *BtctokenTransfer // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *BtctokenTransferIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BtctokenTransfer)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(BtctokenTransfer)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *BtctokenTransferIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BtctokenTransferIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BtctokenTransfer represents a Transfer event raised by the Btctoken contract.
type BtctokenTransfer struct {
	From  common.Address
	To    common.Address
	Value *big.Int
	Raw   types.Log // Blockchain specific contextual infos
}

// FilterTransfer is a free log retrieval operation binding the contract event 0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef.
//
// Solidity: event Transfer(address indexed from, address indexed to, uint256 value)
func (_Btctoken *BtctokenFilterer) FilterTransfer(opts *bind.FilterOpts, from []common.Address, to []common.Address) (*BtctokenTransferIterator, error) {

	var fromRule []interface{}
	for _, fromItem := range from {
		fromRule = append(fromRule, fromItem)
	}
	var toRule []interface{}
	for _, toItem := range to {
		toRule = append(toRule, toItem)
	}

	logs, sub, err := _Btctoken.contract.FilterLogs(opts, "Transfer", fromRule, toRule)
	if err != nil {
		return nil, err
	}
	return &BtctokenTransferIterator{contract: _Btctoken.contract, event: "Transfer", logs: logs, sub: sub}, nil
}

// WatchTransfer is a free log subscription operation binding the contract event 0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef.
//
// Solidity: event Transfer(address indexed from, address indexed to, uint256 value)
func (_Btctoken *BtctokenFilterer) WatchTransfer(opts *bind.WatchOpts, sink chan<- *BtctokenTransfer, from []common.Address, to []common.Address) (event.Subscription, error) {

	var fromRule []interface{}
	for _, fromItem := range from {
		fromRule = append(fromRule, fromItem)
	}
	var toRule []interface{}
	for _, toItem := range to {
		toRule = append(toRule, toItem)
	}

	logs, sub, err := _Btctoken.contract.WatchLogs(opts, "Transfer", fromRule, toRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BtctokenTransfer)
				if err := _Btctoken.contract.UnpackLog(event, "Transfer", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseTransfer is a log parse operation binding the contract event 0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef.
//
// Solidity: event Transfer(address indexed from, address indexed to, uint256 value)
func (_Btctoken *BtctokenFilterer) ParseTransfer(log types.Log) (*BtctokenTransfer, error) {
	event := new(BtctokenTransfer)
	if err := _Btctoken.contract.UnpackLog(event, "Transfer", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
