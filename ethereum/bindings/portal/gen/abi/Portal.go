// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package abi

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
)

// PortalDepositInfo is an auto generated low-level Go binding around an user-defined struct.
type PortalDepositInfo struct {
	Balance  *big.Int
	UnlockAt uint32
}

// PortalSupportedToken is an auto generated low-level Go binding around an user-defined struct.
type PortalSupportedToken struct {
	Token        common.Address
	TokenAbility uint8
}

// PortalMetaData contains all meta data concerning the Portal contract.
var PortalMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"target\",\"type\":\"address\"}],\"name\":\"AddressEmptyCode\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"}],\"name\":\"AddressInsufficientBalance\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint32\",\"name\":\"unlockAt\",\"type\":\"uint32\"}],\"name\":\"DepositLocked\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"DepositNotFound\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"FailedInnerCall\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"IncorrectAmount\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"depositor\",\"type\":\"address\"}],\"name\":\"IncorrectDepositor\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"lockPeriod\",\"type\":\"uint256\"}],\"name\":\"IncorrectLockPeriod\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"enumPortal.TokenAbility\",\"name\":\"ability\",\"type\":\"uint8\"}],\"name\":\"IncorrectTokenAbility\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"token\",\"type\":\"address\"}],\"name\":\"IncorrectTokenAddress\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"depositBalance\",\"type\":\"uint256\"}],\"name\":\"InsufficientDepositAmount\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"token\",\"type\":\"address\"},{\"internalType\":\"enumPortal.TokenAbility\",\"name\":\"tokenAbility\",\"type\":\"uint8\"}],\"name\":\"InsufficientTokenAbility\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"InvalidInitialization\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint32\",\"name\":\"lockPeriod\",\"type\":\"uint32\"}],\"name\":\"LockPeriodOutOfRange\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint32\",\"name\":\"lockPeriod\",\"type\":\"uint32\"},{\"internalType\":\"uint32\",\"name\":\"newUnlockAt\",\"type\":\"uint32\"},{\"internalType\":\"uint32\",\"name\":\"existingUnlockAt\",\"type\":\"uint32\"}],\"name\":\"LockPeriodTooShort\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"NotInitializing\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"}],\"name\":\"OwnableInvalidOwner\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"}],\"name\":\"OwnableUnauthorizedAccount\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"token\",\"type\":\"address\"}],\"name\":\"SafeERC20FailedOperation\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"token\",\"type\":\"address\"},{\"internalType\":\"enumPortal.TokenAbility\",\"name\":\"tokenAbility\",\"type\":\"uint8\"}],\"name\":\"TokenAlreadySupported\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"token\",\"type\":\"address\"}],\"name\":\"TokenNotSupported\",\"type\":\"error\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"depositor\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"token\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"depositId\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"Deposited\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint64\",\"name\":\"version\",\"type\":\"uint64\"}],\"name\":\"Initialized\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"depositor\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"token\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"depositId\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint32\",\"name\":\"unlockAt\",\"type\":\"uint32\"},{\"indexed\":false,\"internalType\":\"uint32\",\"name\":\"lockPeriod\",\"type\":\"uint32\"}],\"name\":\"Locked\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint32\",\"name\":\"maxLockPeriod\",\"type\":\"uint32\"}],\"name\":\"MaxLockPeriodUpdated\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint32\",\"name\":\"minLockPeriod\",\"type\":\"uint32\"}],\"name\":\"MinLockPeriodUpdated\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"previousOwner\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"OwnershipTransferStarted\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"previousOwner\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"OwnershipTransferred\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"token\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"enumPortal.TokenAbility\",\"name\":\"tokenAbility\",\"type\":\"uint8\"}],\"name\":\"SupportedTokenAdded\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"depositor\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"token\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"depositId\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"Withdrawn\",\"type\":\"event\"},{\"inputs\":[],\"name\":\"acceptOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"components\":[{\"internalType\":\"address\",\"name\":\"token\",\"type\":\"address\"},{\"internalType\":\"enumPortal.TokenAbility\",\"name\":\"tokenAbility\",\"type\":\"uint8\"}],\"internalType\":\"structPortal.SupportedToken\",\"name\":\"supportedToken\",\"type\":\"tuple\"}],\"name\":\"addSupportedToken\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"token\",\"type\":\"address\"},{\"internalType\":\"uint96\",\"name\":\"amount\",\"type\":\"uint96\"},{\"internalType\":\"uint32\",\"name\":\"lockPeriod\",\"type\":\"uint32\"}],\"name\":\"deposit\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"depositCount\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"depositOwner\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"token\",\"type\":\"address\"},{\"internalType\":\"uint96\",\"name\":\"amount\",\"type\":\"uint96\"},{\"internalType\":\"uint32\",\"name\":\"lockPeriod\",\"type\":\"uint32\"}],\"name\":\"depositFor\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"deposits\",\"outputs\":[{\"internalType\":\"uint96\",\"name\":\"balance\",\"type\":\"uint96\"},{\"internalType\":\"uint32\",\"name\":\"unlockAt\",\"type\":\"uint32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"depositor\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"token\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"depositId\",\"type\":\"uint256\"}],\"name\":\"getDeposit\",\"outputs\":[{\"components\":[{\"internalType\":\"uint96\",\"name\":\"balance\",\"type\":\"uint96\"},{\"internalType\":\"uint32\",\"name\":\"unlockAt\",\"type\":\"uint32\"}],\"internalType\":\"structPortal.DepositInfo\",\"name\":\"\",\"type\":\"tuple\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"components\":[{\"internalType\":\"address\",\"name\":\"token\",\"type\":\"address\"},{\"internalType\":\"enumPortal.TokenAbility\",\"name\":\"tokenAbility\",\"type\":\"uint8\"}],\"internalType\":\"structPortal.SupportedToken[]\",\"name\":\"supportedTokens\",\"type\":\"tuple[]\"}],\"name\":\"initialize\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"token\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"depositId\",\"type\":\"uint256\"},{\"internalType\":\"uint32\",\"name\":\"lockPeriod\",\"type\":\"uint32\"}],\"name\":\"lock\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"maxLockPeriod\",\"outputs\":[{\"internalType\":\"uint32\",\"name\":\"\",\"type\":\"uint32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"minLockPeriod\",\"outputs\":[{\"internalType\":\"uint32\",\"name\":\"\",\"type\":\"uint32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"owner\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"pendingOwner\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"from\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"token\",\"type\":\"address\"},{\"internalType\":\"bytes\",\"name\":\"data\",\"type\":\"bytes\"}],\"name\":\"receiveApproval\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"renounceOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint32\",\"name\":\"_maxLockPeriod\",\"type\":\"uint32\"}],\"name\":\"setMaxLockPeriod\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint32\",\"name\":\"_minLockPeriod\",\"type\":\"uint32\"}],\"name\":\"setMinLockPeriod\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"name\":\"tokenAbility\",\"outputs\":[{\"internalType\":\"enumPortal.TokenAbility\",\"name\":\"\",\"type\":\"uint8\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"transferOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"token\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"depositId\",\"type\":\"uint256\"},{\"internalType\":\"uint96\",\"name\":\"amount\",\"type\":\"uint96\"}],\"name\":\"withdraw\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]",
}

// PortalABI is the input ABI used to generate the binding from.
// Deprecated: Use PortalMetaData.ABI instead.
var PortalABI = PortalMetaData.ABI

// Portal is an auto generated Go binding around an Ethereum contract.
type Portal struct {
	PortalCaller     // Read-only binding to the contract
	PortalTransactor // Write-only binding to the contract
	PortalFilterer   // Log filterer for contract events
}

// PortalCaller is an auto generated read-only Go binding around an Ethereum contract.
type PortalCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// PortalTransactor is an auto generated write-only Go binding around an Ethereum contract.
type PortalTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// PortalFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type PortalFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// PortalSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type PortalSession struct {
	Contract     *Portal           // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// PortalCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type PortalCallerSession struct {
	Contract *PortalCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts // Call options to use throughout this session
}

// PortalTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type PortalTransactorSession struct {
	Contract     *PortalTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// PortalRaw is an auto generated low-level Go binding around an Ethereum contract.
type PortalRaw struct {
	Contract *Portal // Generic contract binding to access the raw methods on
}

// PortalCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type PortalCallerRaw struct {
	Contract *PortalCaller // Generic read-only contract binding to access the raw methods on
}

// PortalTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type PortalTransactorRaw struct {
	Contract *PortalTransactor // Generic write-only contract binding to access the raw methods on
}

// NewPortal creates a new instance of Portal, bound to a specific deployed contract.
func NewPortal(address common.Address, backend bind.ContractBackend) (*Portal, error) {
	contract, err := bindPortal(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &Portal{PortalCaller: PortalCaller{contract: contract}, PortalTransactor: PortalTransactor{contract: contract}, PortalFilterer: PortalFilterer{contract: contract}}, nil
}

// NewPortalCaller creates a new read-only instance of Portal, bound to a specific deployed contract.
func NewPortalCaller(address common.Address, caller bind.ContractCaller) (*PortalCaller, error) {
	contract, err := bindPortal(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &PortalCaller{contract: contract}, nil
}

// NewPortalTransactor creates a new write-only instance of Portal, bound to a specific deployed contract.
func NewPortalTransactor(address common.Address, transactor bind.ContractTransactor) (*PortalTransactor, error) {
	contract, err := bindPortal(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &PortalTransactor{contract: contract}, nil
}

// NewPortalFilterer creates a new log filterer instance of Portal, bound to a specific deployed contract.
func NewPortalFilterer(address common.Address, filterer bind.ContractFilterer) (*PortalFilterer, error) {
	contract, err := bindPortal(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &PortalFilterer{contract: contract}, nil
}

// bindPortal binds a generic wrapper to an already deployed contract.
func bindPortal(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(PortalABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Portal *PortalRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Portal.Contract.PortalCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Portal *PortalRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Portal.Contract.PortalTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Portal *PortalRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Portal.Contract.PortalTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Portal *PortalCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Portal.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Portal *PortalTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Portal.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Portal *PortalTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Portal.Contract.contract.Transact(opts, method, params...)
}

// DepositCount is a free data retrieval call binding the contract method 0x2dfdf0b5.
//
// Solidity: function depositCount() view returns(uint256)
func (_Portal *PortalCaller) DepositCount(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _Portal.contract.Call(opts, &out, "depositCount")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// DepositCount is a free data retrieval call binding the contract method 0x2dfdf0b5.
//
// Solidity: function depositCount() view returns(uint256)
func (_Portal *PortalSession) DepositCount() (*big.Int, error) {
	return _Portal.Contract.DepositCount(&_Portal.CallOpts)
}

// DepositCount is a free data retrieval call binding the contract method 0x2dfdf0b5.
//
// Solidity: function depositCount() view returns(uint256)
func (_Portal *PortalCallerSession) DepositCount() (*big.Int, error) {
	return _Portal.Contract.DepositCount(&_Portal.CallOpts)
}

// Deposits is a free data retrieval call binding the contract method 0x5d93a3fc.
//
// Solidity: function deposits(address , address , uint256 ) view returns(uint96 balance, uint32 unlockAt)
func (_Portal *PortalCaller) Deposits(opts *bind.CallOpts, arg0 common.Address, arg1 common.Address, arg2 *big.Int) (struct {
	Balance  *big.Int
	UnlockAt uint32
}, error) {
	var out []interface{}
	err := _Portal.contract.Call(opts, &out, "deposits", arg0, arg1, arg2)

	outstruct := new(struct {
		Balance  *big.Int
		UnlockAt uint32
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.Balance = *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)
	outstruct.UnlockAt = *abi.ConvertType(out[1], new(uint32)).(*uint32)

	return *outstruct, err

}

// Deposits is a free data retrieval call binding the contract method 0x5d93a3fc.
//
// Solidity: function deposits(address , address , uint256 ) view returns(uint96 balance, uint32 unlockAt)
func (_Portal *PortalSession) Deposits(arg0 common.Address, arg1 common.Address, arg2 *big.Int) (struct {
	Balance  *big.Int
	UnlockAt uint32
}, error) {
	return _Portal.Contract.Deposits(&_Portal.CallOpts, arg0, arg1, arg2)
}

// Deposits is a free data retrieval call binding the contract method 0x5d93a3fc.
//
// Solidity: function deposits(address , address , uint256 ) view returns(uint96 balance, uint32 unlockAt)
func (_Portal *PortalCallerSession) Deposits(arg0 common.Address, arg1 common.Address, arg2 *big.Int) (struct {
	Balance  *big.Int
	UnlockAt uint32
}, error) {
	return _Portal.Contract.Deposits(&_Portal.CallOpts, arg0, arg1, arg2)
}

// GetDeposit is a free data retrieval call binding the contract method 0x563170e3.
//
// Solidity: function getDeposit(address depositor, address token, uint256 depositId) view returns((uint96,uint32))
func (_Portal *PortalCaller) GetDeposit(opts *bind.CallOpts, depositor common.Address, token common.Address, depositId *big.Int) (PortalDepositInfo, error) {
	var out []interface{}
	err := _Portal.contract.Call(opts, &out, "getDeposit", depositor, token, depositId)

	if err != nil {
		return *new(PortalDepositInfo), err
	}

	out0 := *abi.ConvertType(out[0], new(PortalDepositInfo)).(*PortalDepositInfo)

	return out0, err

}

// GetDeposit is a free data retrieval call binding the contract method 0x563170e3.
//
// Solidity: function getDeposit(address depositor, address token, uint256 depositId) view returns((uint96,uint32))
func (_Portal *PortalSession) GetDeposit(depositor common.Address, token common.Address, depositId *big.Int) (PortalDepositInfo, error) {
	return _Portal.Contract.GetDeposit(&_Portal.CallOpts, depositor, token, depositId)
}

// GetDeposit is a free data retrieval call binding the contract method 0x563170e3.
//
// Solidity: function getDeposit(address depositor, address token, uint256 depositId) view returns((uint96,uint32))
func (_Portal *PortalCallerSession) GetDeposit(depositor common.Address, token common.Address, depositId *big.Int) (PortalDepositInfo, error) {
	return _Portal.Contract.GetDeposit(&_Portal.CallOpts, depositor, token, depositId)
}

// MaxLockPeriod is a free data retrieval call binding the contract method 0x4b1d29b4.
//
// Solidity: function maxLockPeriod() view returns(uint32)
func (_Portal *PortalCaller) MaxLockPeriod(opts *bind.CallOpts) (uint32, error) {
	var out []interface{}
	err := _Portal.contract.Call(opts, &out, "maxLockPeriod")

	if err != nil {
		return *new(uint32), err
	}

	out0 := *abi.ConvertType(out[0], new(uint32)).(*uint32)

	return out0, err

}

// MaxLockPeriod is a free data retrieval call binding the contract method 0x4b1d29b4.
//
// Solidity: function maxLockPeriod() view returns(uint32)
func (_Portal *PortalSession) MaxLockPeriod() (uint32, error) {
	return _Portal.Contract.MaxLockPeriod(&_Portal.CallOpts)
}

// MaxLockPeriod is a free data retrieval call binding the contract method 0x4b1d29b4.
//
// Solidity: function maxLockPeriod() view returns(uint32)
func (_Portal *PortalCallerSession) MaxLockPeriod() (uint32, error) {
	return _Portal.Contract.MaxLockPeriod(&_Portal.CallOpts)
}

// MinLockPeriod is a free data retrieval call binding the contract method 0x73ae54b5.
//
// Solidity: function minLockPeriod() view returns(uint32)
func (_Portal *PortalCaller) MinLockPeriod(opts *bind.CallOpts) (uint32, error) {
	var out []interface{}
	err := _Portal.contract.Call(opts, &out, "minLockPeriod")

	if err != nil {
		return *new(uint32), err
	}

	out0 := *abi.ConvertType(out[0], new(uint32)).(*uint32)

	return out0, err

}

// MinLockPeriod is a free data retrieval call binding the contract method 0x73ae54b5.
//
// Solidity: function minLockPeriod() view returns(uint32)
func (_Portal *PortalSession) MinLockPeriod() (uint32, error) {
	return _Portal.Contract.MinLockPeriod(&_Portal.CallOpts)
}

// MinLockPeriod is a free data retrieval call binding the contract method 0x73ae54b5.
//
// Solidity: function minLockPeriod() view returns(uint32)
func (_Portal *PortalCallerSession) MinLockPeriod() (uint32, error) {
	return _Portal.Contract.MinLockPeriod(&_Portal.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_Portal *PortalCaller) Owner(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _Portal.contract.Call(opts, &out, "owner")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_Portal *PortalSession) Owner() (common.Address, error) {
	return _Portal.Contract.Owner(&_Portal.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_Portal *PortalCallerSession) Owner() (common.Address, error) {
	return _Portal.Contract.Owner(&_Portal.CallOpts)
}

// PendingOwner is a free data retrieval call binding the contract method 0xe30c3978.
//
// Solidity: function pendingOwner() view returns(address)
func (_Portal *PortalCaller) PendingOwner(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _Portal.contract.Call(opts, &out, "pendingOwner")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// PendingOwner is a free data retrieval call binding the contract method 0xe30c3978.
//
// Solidity: function pendingOwner() view returns(address)
func (_Portal *PortalSession) PendingOwner() (common.Address, error) {
	return _Portal.Contract.PendingOwner(&_Portal.CallOpts)
}

// PendingOwner is a free data retrieval call binding the contract method 0xe30c3978.
//
// Solidity: function pendingOwner() view returns(address)
func (_Portal *PortalCallerSession) PendingOwner() (common.Address, error) {
	return _Portal.Contract.PendingOwner(&_Portal.CallOpts)
}

// TokenAbility is a free data retrieval call binding the contract method 0x6572c5dd.
//
// Solidity: function tokenAbility(address ) view returns(uint8)
func (_Portal *PortalCaller) TokenAbility(opts *bind.CallOpts, arg0 common.Address) (uint8, error) {
	var out []interface{}
	err := _Portal.contract.Call(opts, &out, "tokenAbility", arg0)

	if err != nil {
		return *new(uint8), err
	}

	out0 := *abi.ConvertType(out[0], new(uint8)).(*uint8)

	return out0, err

}

// TokenAbility is a free data retrieval call binding the contract method 0x6572c5dd.
//
// Solidity: function tokenAbility(address ) view returns(uint8)
func (_Portal *PortalSession) TokenAbility(arg0 common.Address) (uint8, error) {
	return _Portal.Contract.TokenAbility(&_Portal.CallOpts, arg0)
}

// TokenAbility is a free data retrieval call binding the contract method 0x6572c5dd.
//
// Solidity: function tokenAbility(address ) view returns(uint8)
func (_Portal *PortalCallerSession) TokenAbility(arg0 common.Address) (uint8, error) {
	return _Portal.Contract.TokenAbility(&_Portal.CallOpts, arg0)
}

// AcceptOwnership is a paid mutator transaction binding the contract method 0x79ba5097.
//
// Solidity: function acceptOwnership() returns()
func (_Portal *PortalTransactor) AcceptOwnership(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Portal.contract.Transact(opts, "acceptOwnership")
}

// AcceptOwnership is a paid mutator transaction binding the contract method 0x79ba5097.
//
// Solidity: function acceptOwnership() returns()
func (_Portal *PortalSession) AcceptOwnership() (*types.Transaction, error) {
	return _Portal.Contract.AcceptOwnership(&_Portal.TransactOpts)
}

// AcceptOwnership is a paid mutator transaction binding the contract method 0x79ba5097.
//
// Solidity: function acceptOwnership() returns()
func (_Portal *PortalTransactorSession) AcceptOwnership() (*types.Transaction, error) {
	return _Portal.Contract.AcceptOwnership(&_Portal.TransactOpts)
}

// AddSupportedToken is a paid mutator transaction binding the contract method 0x0908c7dc.
//
// Solidity: function addSupportedToken((address,uint8) supportedToken) returns()
func (_Portal *PortalTransactor) AddSupportedToken(opts *bind.TransactOpts, supportedToken PortalSupportedToken) (*types.Transaction, error) {
	return _Portal.contract.Transact(opts, "addSupportedToken", supportedToken)
}

// AddSupportedToken is a paid mutator transaction binding the contract method 0x0908c7dc.
//
// Solidity: function addSupportedToken((address,uint8) supportedToken) returns()
func (_Portal *PortalSession) AddSupportedToken(supportedToken PortalSupportedToken) (*types.Transaction, error) {
	return _Portal.Contract.AddSupportedToken(&_Portal.TransactOpts, supportedToken)
}

// AddSupportedToken is a paid mutator transaction binding the contract method 0x0908c7dc.
//
// Solidity: function addSupportedToken((address,uint8) supportedToken) returns()
func (_Portal *PortalTransactorSession) AddSupportedToken(supportedToken PortalSupportedToken) (*types.Transaction, error) {
	return _Portal.Contract.AddSupportedToken(&_Portal.TransactOpts, supportedToken)
}

// Deposit is a paid mutator transaction binding the contract method 0x31645d4e.
//
// Solidity: function deposit(address token, uint96 amount, uint32 lockPeriod) returns()
func (_Portal *PortalTransactor) Deposit(opts *bind.TransactOpts, token common.Address, amount *big.Int, lockPeriod uint32) (*types.Transaction, error) {
	return _Portal.contract.Transact(opts, "deposit", token, amount, lockPeriod)
}

// Deposit is a paid mutator transaction binding the contract method 0x31645d4e.
//
// Solidity: function deposit(address token, uint96 amount, uint32 lockPeriod) returns()
func (_Portal *PortalSession) Deposit(token common.Address, amount *big.Int, lockPeriod uint32) (*types.Transaction, error) {
	return _Portal.Contract.Deposit(&_Portal.TransactOpts, token, amount, lockPeriod)
}

// Deposit is a paid mutator transaction binding the contract method 0x31645d4e.
//
// Solidity: function deposit(address token, uint96 amount, uint32 lockPeriod) returns()
func (_Portal *PortalTransactorSession) Deposit(token common.Address, amount *big.Int, lockPeriod uint32) (*types.Transaction, error) {
	return _Portal.Contract.Deposit(&_Portal.TransactOpts, token, amount, lockPeriod)
}

// DepositFor is a paid mutator transaction binding the contract method 0xdfb6c2d2.
//
// Solidity: function depositFor(address depositOwner, address token, uint96 amount, uint32 lockPeriod) returns()
func (_Portal *PortalTransactor) DepositFor(opts *bind.TransactOpts, depositOwner common.Address, token common.Address, amount *big.Int, lockPeriod uint32) (*types.Transaction, error) {
	return _Portal.contract.Transact(opts, "depositFor", depositOwner, token, amount, lockPeriod)
}

// DepositFor is a paid mutator transaction binding the contract method 0xdfb6c2d2.
//
// Solidity: function depositFor(address depositOwner, address token, uint96 amount, uint32 lockPeriod) returns()
func (_Portal *PortalSession) DepositFor(depositOwner common.Address, token common.Address, amount *big.Int, lockPeriod uint32) (*types.Transaction, error) {
	return _Portal.Contract.DepositFor(&_Portal.TransactOpts, depositOwner, token, amount, lockPeriod)
}

// DepositFor is a paid mutator transaction binding the contract method 0xdfb6c2d2.
//
// Solidity: function depositFor(address depositOwner, address token, uint96 amount, uint32 lockPeriod) returns()
func (_Portal *PortalTransactorSession) DepositFor(depositOwner common.Address, token common.Address, amount *big.Int, lockPeriod uint32) (*types.Transaction, error) {
	return _Portal.Contract.DepositFor(&_Portal.TransactOpts, depositOwner, token, amount, lockPeriod)
}

// Initialize is a paid mutator transaction binding the contract method 0x2c4b24ae.
//
// Solidity: function initialize((address,uint8)[] supportedTokens) returns()
func (_Portal *PortalTransactor) Initialize(opts *bind.TransactOpts, supportedTokens []PortalSupportedToken) (*types.Transaction, error) {
	return _Portal.contract.Transact(opts, "initialize", supportedTokens)
}

// Initialize is a paid mutator transaction binding the contract method 0x2c4b24ae.
//
// Solidity: function initialize((address,uint8)[] supportedTokens) returns()
func (_Portal *PortalSession) Initialize(supportedTokens []PortalSupportedToken) (*types.Transaction, error) {
	return _Portal.Contract.Initialize(&_Portal.TransactOpts, supportedTokens)
}

// Initialize is a paid mutator transaction binding the contract method 0x2c4b24ae.
//
// Solidity: function initialize((address,uint8)[] supportedTokens) returns()
func (_Portal *PortalTransactorSession) Initialize(supportedTokens []PortalSupportedToken) (*types.Transaction, error) {
	return _Portal.Contract.Initialize(&_Portal.TransactOpts, supportedTokens)
}

// Lock is a paid mutator transaction binding the contract method 0xf5e8d327.
//
// Solidity: function lock(address token, uint256 depositId, uint32 lockPeriod) returns()
func (_Portal *PortalTransactor) Lock(opts *bind.TransactOpts, token common.Address, depositId *big.Int, lockPeriod uint32) (*types.Transaction, error) {
	return _Portal.contract.Transact(opts, "lock", token, depositId, lockPeriod)
}

// Lock is a paid mutator transaction binding the contract method 0xf5e8d327.
//
// Solidity: function lock(address token, uint256 depositId, uint32 lockPeriod) returns()
func (_Portal *PortalSession) Lock(token common.Address, depositId *big.Int, lockPeriod uint32) (*types.Transaction, error) {
	return _Portal.Contract.Lock(&_Portal.TransactOpts, token, depositId, lockPeriod)
}

// Lock is a paid mutator transaction binding the contract method 0xf5e8d327.
//
// Solidity: function lock(address token, uint256 depositId, uint32 lockPeriod) returns()
func (_Portal *PortalTransactorSession) Lock(token common.Address, depositId *big.Int, lockPeriod uint32) (*types.Transaction, error) {
	return _Portal.Contract.Lock(&_Portal.TransactOpts, token, depositId, lockPeriod)
}

// ReceiveApproval is a paid mutator transaction binding the contract method 0x8f4ffcb1.
//
// Solidity: function receiveApproval(address from, uint256 amount, address token, bytes data) returns()
func (_Portal *PortalTransactor) ReceiveApproval(opts *bind.TransactOpts, from common.Address, amount *big.Int, token common.Address, data []byte) (*types.Transaction, error) {
	return _Portal.contract.Transact(opts, "receiveApproval", from, amount, token, data)
}

// ReceiveApproval is a paid mutator transaction binding the contract method 0x8f4ffcb1.
//
// Solidity: function receiveApproval(address from, uint256 amount, address token, bytes data) returns()
func (_Portal *PortalSession) ReceiveApproval(from common.Address, amount *big.Int, token common.Address, data []byte) (*types.Transaction, error) {
	return _Portal.Contract.ReceiveApproval(&_Portal.TransactOpts, from, amount, token, data)
}

// ReceiveApproval is a paid mutator transaction binding the contract method 0x8f4ffcb1.
//
// Solidity: function receiveApproval(address from, uint256 amount, address token, bytes data) returns()
func (_Portal *PortalTransactorSession) ReceiveApproval(from common.Address, amount *big.Int, token common.Address, data []byte) (*types.Transaction, error) {
	return _Portal.Contract.ReceiveApproval(&_Portal.TransactOpts, from, amount, token, data)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_Portal *PortalTransactor) RenounceOwnership(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Portal.contract.Transact(opts, "renounceOwnership")
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_Portal *PortalSession) RenounceOwnership() (*types.Transaction, error) {
	return _Portal.Contract.RenounceOwnership(&_Portal.TransactOpts)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_Portal *PortalTransactorSession) RenounceOwnership() (*types.Transaction, error) {
	return _Portal.Contract.RenounceOwnership(&_Portal.TransactOpts)
}

// SetMaxLockPeriod is a paid mutator transaction binding the contract method 0xf64a6c90.
//
// Solidity: function setMaxLockPeriod(uint32 _maxLockPeriod) returns()
func (_Portal *PortalTransactor) SetMaxLockPeriod(opts *bind.TransactOpts, _maxLockPeriod uint32) (*types.Transaction, error) {
	return _Portal.contract.Transact(opts, "setMaxLockPeriod", _maxLockPeriod)
}

// SetMaxLockPeriod is a paid mutator transaction binding the contract method 0xf64a6c90.
//
// Solidity: function setMaxLockPeriod(uint32 _maxLockPeriod) returns()
func (_Portal *PortalSession) SetMaxLockPeriod(_maxLockPeriod uint32) (*types.Transaction, error) {
	return _Portal.Contract.SetMaxLockPeriod(&_Portal.TransactOpts, _maxLockPeriod)
}

// SetMaxLockPeriod is a paid mutator transaction binding the contract method 0xf64a6c90.
//
// Solidity: function setMaxLockPeriod(uint32 _maxLockPeriod) returns()
func (_Portal *PortalTransactorSession) SetMaxLockPeriod(_maxLockPeriod uint32) (*types.Transaction, error) {
	return _Portal.Contract.SetMaxLockPeriod(&_Portal.TransactOpts, _maxLockPeriod)
}

// SetMinLockPeriod is a paid mutator transaction binding the contract method 0x92673f55.
//
// Solidity: function setMinLockPeriod(uint32 _minLockPeriod) returns()
func (_Portal *PortalTransactor) SetMinLockPeriod(opts *bind.TransactOpts, _minLockPeriod uint32) (*types.Transaction, error) {
	return _Portal.contract.Transact(opts, "setMinLockPeriod", _minLockPeriod)
}

// SetMinLockPeriod is a paid mutator transaction binding the contract method 0x92673f55.
//
// Solidity: function setMinLockPeriod(uint32 _minLockPeriod) returns()
func (_Portal *PortalSession) SetMinLockPeriod(_minLockPeriod uint32) (*types.Transaction, error) {
	return _Portal.Contract.SetMinLockPeriod(&_Portal.TransactOpts, _minLockPeriod)
}

// SetMinLockPeriod is a paid mutator transaction binding the contract method 0x92673f55.
//
// Solidity: function setMinLockPeriod(uint32 _minLockPeriod) returns()
func (_Portal *PortalTransactorSession) SetMinLockPeriod(_minLockPeriod uint32) (*types.Transaction, error) {
	return _Portal.Contract.SetMinLockPeriod(&_Portal.TransactOpts, _minLockPeriod)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_Portal *PortalTransactor) TransferOwnership(opts *bind.TransactOpts, newOwner common.Address) (*types.Transaction, error) {
	return _Portal.contract.Transact(opts, "transferOwnership", newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_Portal *PortalSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _Portal.Contract.TransferOwnership(&_Portal.TransactOpts, newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_Portal *PortalTransactorSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _Portal.Contract.TransferOwnership(&_Portal.TransactOpts, newOwner)
}

// Withdraw is a paid mutator transaction binding the contract method 0x6292ab35.
//
// Solidity: function withdraw(address token, uint256 depositId, uint96 amount) returns()
func (_Portal *PortalTransactor) Withdraw(opts *bind.TransactOpts, token common.Address, depositId *big.Int, amount *big.Int) (*types.Transaction, error) {
	return _Portal.contract.Transact(opts, "withdraw", token, depositId, amount)
}

// Withdraw is a paid mutator transaction binding the contract method 0x6292ab35.
//
// Solidity: function withdraw(address token, uint256 depositId, uint96 amount) returns()
func (_Portal *PortalSession) Withdraw(token common.Address, depositId *big.Int, amount *big.Int) (*types.Transaction, error) {
	return _Portal.Contract.Withdraw(&_Portal.TransactOpts, token, depositId, amount)
}

// Withdraw is a paid mutator transaction binding the contract method 0x6292ab35.
//
// Solidity: function withdraw(address token, uint256 depositId, uint96 amount) returns()
func (_Portal *PortalTransactorSession) Withdraw(token common.Address, depositId *big.Int, amount *big.Int) (*types.Transaction, error) {
	return _Portal.Contract.Withdraw(&_Portal.TransactOpts, token, depositId, amount)
}

// PortalDepositedIterator is returned from FilterDeposited and is used to iterate over the raw logs and unpacked data for Deposited events raised by the Portal contract.
type PortalDepositedIterator struct {
	Event *PortalDeposited // Event containing the contract specifics and raw log

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
func (it *PortalDepositedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(PortalDeposited)
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
		it.Event = new(PortalDeposited)
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
func (it *PortalDepositedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *PortalDepositedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// PortalDeposited represents a Deposited event raised by the Portal contract.
type PortalDeposited struct {
	Depositor common.Address
	Token     common.Address
	DepositId *big.Int
	Amount    *big.Int
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterDeposited is a free log retrieval operation binding the contract event 0xf5681f9d0db1b911ac18ee83d515a1cf1051853a9eae418316a2fdf7dea427c5.
//
// Solidity: event Deposited(address indexed depositor, address indexed token, uint256 indexed depositId, uint256 amount)
func (_Portal *PortalFilterer) FilterDeposited(opts *bind.FilterOpts, depositor []common.Address, token []common.Address, depositId []*big.Int) (*PortalDepositedIterator, error) {

	var depositorRule []interface{}
	for _, depositorItem := range depositor {
		depositorRule = append(depositorRule, depositorItem)
	}
	var tokenRule []interface{}
	for _, tokenItem := range token {
		tokenRule = append(tokenRule, tokenItem)
	}
	var depositIdRule []interface{}
	for _, depositIdItem := range depositId {
		depositIdRule = append(depositIdRule, depositIdItem)
	}

	logs, sub, err := _Portal.contract.FilterLogs(opts, "Deposited", depositorRule, tokenRule, depositIdRule)
	if err != nil {
		return nil, err
	}
	return &PortalDepositedIterator{contract: _Portal.contract, event: "Deposited", logs: logs, sub: sub}, nil
}

// WatchDeposited is a free log subscription operation binding the contract event 0xf5681f9d0db1b911ac18ee83d515a1cf1051853a9eae418316a2fdf7dea427c5.
//
// Solidity: event Deposited(address indexed depositor, address indexed token, uint256 indexed depositId, uint256 amount)
func (_Portal *PortalFilterer) WatchDeposited(opts *bind.WatchOpts, sink chan<- *PortalDeposited, depositor []common.Address, token []common.Address, depositId []*big.Int) (event.Subscription, error) {

	var depositorRule []interface{}
	for _, depositorItem := range depositor {
		depositorRule = append(depositorRule, depositorItem)
	}
	var tokenRule []interface{}
	for _, tokenItem := range token {
		tokenRule = append(tokenRule, tokenItem)
	}
	var depositIdRule []interface{}
	for _, depositIdItem := range depositId {
		depositIdRule = append(depositIdRule, depositIdItem)
	}

	logs, sub, err := _Portal.contract.WatchLogs(opts, "Deposited", depositorRule, tokenRule, depositIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(PortalDeposited)
				if err := _Portal.contract.UnpackLog(event, "Deposited", log); err != nil {
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

// ParseDeposited is a log parse operation binding the contract event 0xf5681f9d0db1b911ac18ee83d515a1cf1051853a9eae418316a2fdf7dea427c5.
//
// Solidity: event Deposited(address indexed depositor, address indexed token, uint256 indexed depositId, uint256 amount)
func (_Portal *PortalFilterer) ParseDeposited(log types.Log) (*PortalDeposited, error) {
	event := new(PortalDeposited)
	if err := _Portal.contract.UnpackLog(event, "Deposited", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// PortalInitializedIterator is returned from FilterInitialized and is used to iterate over the raw logs and unpacked data for Initialized events raised by the Portal contract.
type PortalInitializedIterator struct {
	Event *PortalInitialized // Event containing the contract specifics and raw log

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
func (it *PortalInitializedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(PortalInitialized)
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
		it.Event = new(PortalInitialized)
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
func (it *PortalInitializedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *PortalInitializedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// PortalInitialized represents a Initialized event raised by the Portal contract.
type PortalInitialized struct {
	Version uint64
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterInitialized is a free log retrieval operation binding the contract event 0xc7f505b2f371ae2175ee4913f4499e1f2633a7b5936321eed1cdaeb6115181d2.
//
// Solidity: event Initialized(uint64 version)
func (_Portal *PortalFilterer) FilterInitialized(opts *bind.FilterOpts) (*PortalInitializedIterator, error) {

	logs, sub, err := _Portal.contract.FilterLogs(opts, "Initialized")
	if err != nil {
		return nil, err
	}
	return &PortalInitializedIterator{contract: _Portal.contract, event: "Initialized", logs: logs, sub: sub}, nil
}

// WatchInitialized is a free log subscription operation binding the contract event 0xc7f505b2f371ae2175ee4913f4499e1f2633a7b5936321eed1cdaeb6115181d2.
//
// Solidity: event Initialized(uint64 version)
func (_Portal *PortalFilterer) WatchInitialized(opts *bind.WatchOpts, sink chan<- *PortalInitialized) (event.Subscription, error) {

	logs, sub, err := _Portal.contract.WatchLogs(opts, "Initialized")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(PortalInitialized)
				if err := _Portal.contract.UnpackLog(event, "Initialized", log); err != nil {
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

// ParseInitialized is a log parse operation binding the contract event 0xc7f505b2f371ae2175ee4913f4499e1f2633a7b5936321eed1cdaeb6115181d2.
//
// Solidity: event Initialized(uint64 version)
func (_Portal *PortalFilterer) ParseInitialized(log types.Log) (*PortalInitialized, error) {
	event := new(PortalInitialized)
	if err := _Portal.contract.UnpackLog(event, "Initialized", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// PortalLockedIterator is returned from FilterLocked and is used to iterate over the raw logs and unpacked data for Locked events raised by the Portal contract.
type PortalLockedIterator struct {
	Event *PortalLocked // Event containing the contract specifics and raw log

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
func (it *PortalLockedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(PortalLocked)
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
		it.Event = new(PortalLocked)
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
func (it *PortalLockedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *PortalLockedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// PortalLocked represents a Locked event raised by the Portal contract.
type PortalLocked struct {
	Depositor  common.Address
	Token      common.Address
	DepositId  *big.Int
	UnlockAt   uint32
	LockPeriod uint32
	Raw        types.Log // Blockchain specific contextual infos
}

// FilterLocked is a free log retrieval operation binding the contract event 0x8b65b80ac62fde507cb8196bad6c93c114c2babc6ac846aae39ed6943ad36c49.
//
// Solidity: event Locked(address indexed depositor, address indexed token, uint256 indexed depositId, uint32 unlockAt, uint32 lockPeriod)
func (_Portal *PortalFilterer) FilterLocked(opts *bind.FilterOpts, depositor []common.Address, token []common.Address, depositId []*big.Int) (*PortalLockedIterator, error) {

	var depositorRule []interface{}
	for _, depositorItem := range depositor {
		depositorRule = append(depositorRule, depositorItem)
	}
	var tokenRule []interface{}
	for _, tokenItem := range token {
		tokenRule = append(tokenRule, tokenItem)
	}
	var depositIdRule []interface{}
	for _, depositIdItem := range depositId {
		depositIdRule = append(depositIdRule, depositIdItem)
	}

	logs, sub, err := _Portal.contract.FilterLogs(opts, "Locked", depositorRule, tokenRule, depositIdRule)
	if err != nil {
		return nil, err
	}
	return &PortalLockedIterator{contract: _Portal.contract, event: "Locked", logs: logs, sub: sub}, nil
}

// WatchLocked is a free log subscription operation binding the contract event 0x8b65b80ac62fde507cb8196bad6c93c114c2babc6ac846aae39ed6943ad36c49.
//
// Solidity: event Locked(address indexed depositor, address indexed token, uint256 indexed depositId, uint32 unlockAt, uint32 lockPeriod)
func (_Portal *PortalFilterer) WatchLocked(opts *bind.WatchOpts, sink chan<- *PortalLocked, depositor []common.Address, token []common.Address, depositId []*big.Int) (event.Subscription, error) {

	var depositorRule []interface{}
	for _, depositorItem := range depositor {
		depositorRule = append(depositorRule, depositorItem)
	}
	var tokenRule []interface{}
	for _, tokenItem := range token {
		tokenRule = append(tokenRule, tokenItem)
	}
	var depositIdRule []interface{}
	for _, depositIdItem := range depositId {
		depositIdRule = append(depositIdRule, depositIdItem)
	}

	logs, sub, err := _Portal.contract.WatchLogs(opts, "Locked", depositorRule, tokenRule, depositIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(PortalLocked)
				if err := _Portal.contract.UnpackLog(event, "Locked", log); err != nil {
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

// ParseLocked is a log parse operation binding the contract event 0x8b65b80ac62fde507cb8196bad6c93c114c2babc6ac846aae39ed6943ad36c49.
//
// Solidity: event Locked(address indexed depositor, address indexed token, uint256 indexed depositId, uint32 unlockAt, uint32 lockPeriod)
func (_Portal *PortalFilterer) ParseLocked(log types.Log) (*PortalLocked, error) {
	event := new(PortalLocked)
	if err := _Portal.contract.UnpackLog(event, "Locked", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// PortalMaxLockPeriodUpdatedIterator is returned from FilterMaxLockPeriodUpdated and is used to iterate over the raw logs and unpacked data for MaxLockPeriodUpdated events raised by the Portal contract.
type PortalMaxLockPeriodUpdatedIterator struct {
	Event *PortalMaxLockPeriodUpdated // Event containing the contract specifics and raw log

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
func (it *PortalMaxLockPeriodUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(PortalMaxLockPeriodUpdated)
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
		it.Event = new(PortalMaxLockPeriodUpdated)
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
func (it *PortalMaxLockPeriodUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *PortalMaxLockPeriodUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// PortalMaxLockPeriodUpdated represents a MaxLockPeriodUpdated event raised by the Portal contract.
type PortalMaxLockPeriodUpdated struct {
	MaxLockPeriod uint32
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterMaxLockPeriodUpdated is a free log retrieval operation binding the contract event 0xe02644567ab9266166c374f84f05396b070729fc139339e70d0237bb37e59dc5.
//
// Solidity: event MaxLockPeriodUpdated(uint32 maxLockPeriod)
func (_Portal *PortalFilterer) FilterMaxLockPeriodUpdated(opts *bind.FilterOpts) (*PortalMaxLockPeriodUpdatedIterator, error) {

	logs, sub, err := _Portal.contract.FilterLogs(opts, "MaxLockPeriodUpdated")
	if err != nil {
		return nil, err
	}
	return &PortalMaxLockPeriodUpdatedIterator{contract: _Portal.contract, event: "MaxLockPeriodUpdated", logs: logs, sub: sub}, nil
}

// WatchMaxLockPeriodUpdated is a free log subscription operation binding the contract event 0xe02644567ab9266166c374f84f05396b070729fc139339e70d0237bb37e59dc5.
//
// Solidity: event MaxLockPeriodUpdated(uint32 maxLockPeriod)
func (_Portal *PortalFilterer) WatchMaxLockPeriodUpdated(opts *bind.WatchOpts, sink chan<- *PortalMaxLockPeriodUpdated) (event.Subscription, error) {

	logs, sub, err := _Portal.contract.WatchLogs(opts, "MaxLockPeriodUpdated")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(PortalMaxLockPeriodUpdated)
				if err := _Portal.contract.UnpackLog(event, "MaxLockPeriodUpdated", log); err != nil {
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

// ParseMaxLockPeriodUpdated is a log parse operation binding the contract event 0xe02644567ab9266166c374f84f05396b070729fc139339e70d0237bb37e59dc5.
//
// Solidity: event MaxLockPeriodUpdated(uint32 maxLockPeriod)
func (_Portal *PortalFilterer) ParseMaxLockPeriodUpdated(log types.Log) (*PortalMaxLockPeriodUpdated, error) {
	event := new(PortalMaxLockPeriodUpdated)
	if err := _Portal.contract.UnpackLog(event, "MaxLockPeriodUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// PortalMinLockPeriodUpdatedIterator is returned from FilterMinLockPeriodUpdated and is used to iterate over the raw logs and unpacked data for MinLockPeriodUpdated events raised by the Portal contract.
type PortalMinLockPeriodUpdatedIterator struct {
	Event *PortalMinLockPeriodUpdated // Event containing the contract specifics and raw log

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
func (it *PortalMinLockPeriodUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(PortalMinLockPeriodUpdated)
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
		it.Event = new(PortalMinLockPeriodUpdated)
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
func (it *PortalMinLockPeriodUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *PortalMinLockPeriodUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// PortalMinLockPeriodUpdated represents a MinLockPeriodUpdated event raised by the Portal contract.
type PortalMinLockPeriodUpdated struct {
	MinLockPeriod uint32
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterMinLockPeriodUpdated is a free log retrieval operation binding the contract event 0x4c35d0e4acd88f9d47ba71b6a74a890a34499d0af9d7536e5b46c2b190ea18be.
//
// Solidity: event MinLockPeriodUpdated(uint32 minLockPeriod)
func (_Portal *PortalFilterer) FilterMinLockPeriodUpdated(opts *bind.FilterOpts) (*PortalMinLockPeriodUpdatedIterator, error) {

	logs, sub, err := _Portal.contract.FilterLogs(opts, "MinLockPeriodUpdated")
	if err != nil {
		return nil, err
	}
	return &PortalMinLockPeriodUpdatedIterator{contract: _Portal.contract, event: "MinLockPeriodUpdated", logs: logs, sub: sub}, nil
}

// WatchMinLockPeriodUpdated is a free log subscription operation binding the contract event 0x4c35d0e4acd88f9d47ba71b6a74a890a34499d0af9d7536e5b46c2b190ea18be.
//
// Solidity: event MinLockPeriodUpdated(uint32 minLockPeriod)
func (_Portal *PortalFilterer) WatchMinLockPeriodUpdated(opts *bind.WatchOpts, sink chan<- *PortalMinLockPeriodUpdated) (event.Subscription, error) {

	logs, sub, err := _Portal.contract.WatchLogs(opts, "MinLockPeriodUpdated")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(PortalMinLockPeriodUpdated)
				if err := _Portal.contract.UnpackLog(event, "MinLockPeriodUpdated", log); err != nil {
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

// ParseMinLockPeriodUpdated is a log parse operation binding the contract event 0x4c35d0e4acd88f9d47ba71b6a74a890a34499d0af9d7536e5b46c2b190ea18be.
//
// Solidity: event MinLockPeriodUpdated(uint32 minLockPeriod)
func (_Portal *PortalFilterer) ParseMinLockPeriodUpdated(log types.Log) (*PortalMinLockPeriodUpdated, error) {
	event := new(PortalMinLockPeriodUpdated)
	if err := _Portal.contract.UnpackLog(event, "MinLockPeriodUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// PortalOwnershipTransferStartedIterator is returned from FilterOwnershipTransferStarted and is used to iterate over the raw logs and unpacked data for OwnershipTransferStarted events raised by the Portal contract.
type PortalOwnershipTransferStartedIterator struct {
	Event *PortalOwnershipTransferStarted // Event containing the contract specifics and raw log

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
func (it *PortalOwnershipTransferStartedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(PortalOwnershipTransferStarted)
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
		it.Event = new(PortalOwnershipTransferStarted)
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
func (it *PortalOwnershipTransferStartedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *PortalOwnershipTransferStartedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// PortalOwnershipTransferStarted represents a OwnershipTransferStarted event raised by the Portal contract.
type PortalOwnershipTransferStarted struct {
	PreviousOwner common.Address
	NewOwner      common.Address
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterOwnershipTransferStarted is a free log retrieval operation binding the contract event 0x38d16b8cac22d99fc7c124b9cd0de2d3fa1faef420bfe791d8c362d765e22700.
//
// Solidity: event OwnershipTransferStarted(address indexed previousOwner, address indexed newOwner)
func (_Portal *PortalFilterer) FilterOwnershipTransferStarted(opts *bind.FilterOpts, previousOwner []common.Address, newOwner []common.Address) (*PortalOwnershipTransferStartedIterator, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _Portal.contract.FilterLogs(opts, "OwnershipTransferStarted", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return &PortalOwnershipTransferStartedIterator{contract: _Portal.contract, event: "OwnershipTransferStarted", logs: logs, sub: sub}, nil
}

// WatchOwnershipTransferStarted is a free log subscription operation binding the contract event 0x38d16b8cac22d99fc7c124b9cd0de2d3fa1faef420bfe791d8c362d765e22700.
//
// Solidity: event OwnershipTransferStarted(address indexed previousOwner, address indexed newOwner)
func (_Portal *PortalFilterer) WatchOwnershipTransferStarted(opts *bind.WatchOpts, sink chan<- *PortalOwnershipTransferStarted, previousOwner []common.Address, newOwner []common.Address) (event.Subscription, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _Portal.contract.WatchLogs(opts, "OwnershipTransferStarted", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(PortalOwnershipTransferStarted)
				if err := _Portal.contract.UnpackLog(event, "OwnershipTransferStarted", log); err != nil {
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

// ParseOwnershipTransferStarted is a log parse operation binding the contract event 0x38d16b8cac22d99fc7c124b9cd0de2d3fa1faef420bfe791d8c362d765e22700.
//
// Solidity: event OwnershipTransferStarted(address indexed previousOwner, address indexed newOwner)
func (_Portal *PortalFilterer) ParseOwnershipTransferStarted(log types.Log) (*PortalOwnershipTransferStarted, error) {
	event := new(PortalOwnershipTransferStarted)
	if err := _Portal.contract.UnpackLog(event, "OwnershipTransferStarted", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// PortalOwnershipTransferredIterator is returned from FilterOwnershipTransferred and is used to iterate over the raw logs and unpacked data for OwnershipTransferred events raised by the Portal contract.
type PortalOwnershipTransferredIterator struct {
	Event *PortalOwnershipTransferred // Event containing the contract specifics and raw log

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
func (it *PortalOwnershipTransferredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(PortalOwnershipTransferred)
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
		it.Event = new(PortalOwnershipTransferred)
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
func (it *PortalOwnershipTransferredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *PortalOwnershipTransferredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// PortalOwnershipTransferred represents a OwnershipTransferred event raised by the Portal contract.
type PortalOwnershipTransferred struct {
	PreviousOwner common.Address
	NewOwner      common.Address
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterOwnershipTransferred is a free log retrieval operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_Portal *PortalFilterer) FilterOwnershipTransferred(opts *bind.FilterOpts, previousOwner []common.Address, newOwner []common.Address) (*PortalOwnershipTransferredIterator, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _Portal.contract.FilterLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return &PortalOwnershipTransferredIterator{contract: _Portal.contract, event: "OwnershipTransferred", logs: logs, sub: sub}, nil
}

// WatchOwnershipTransferred is a free log subscription operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_Portal *PortalFilterer) WatchOwnershipTransferred(opts *bind.WatchOpts, sink chan<- *PortalOwnershipTransferred, previousOwner []common.Address, newOwner []common.Address) (event.Subscription, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _Portal.contract.WatchLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(PortalOwnershipTransferred)
				if err := _Portal.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
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

// ParseOwnershipTransferred is a log parse operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_Portal *PortalFilterer) ParseOwnershipTransferred(log types.Log) (*PortalOwnershipTransferred, error) {
	event := new(PortalOwnershipTransferred)
	if err := _Portal.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// PortalSupportedTokenAddedIterator is returned from FilterSupportedTokenAdded and is used to iterate over the raw logs and unpacked data for SupportedTokenAdded events raised by the Portal contract.
type PortalSupportedTokenAddedIterator struct {
	Event *PortalSupportedTokenAdded // Event containing the contract specifics and raw log

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
func (it *PortalSupportedTokenAddedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(PortalSupportedTokenAdded)
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
		it.Event = new(PortalSupportedTokenAdded)
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
func (it *PortalSupportedTokenAddedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *PortalSupportedTokenAddedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// PortalSupportedTokenAdded represents a SupportedTokenAdded event raised by the Portal contract.
type PortalSupportedTokenAdded struct {
	Token        common.Address
	TokenAbility uint8
	Raw          types.Log // Blockchain specific contextual infos
}

// FilterSupportedTokenAdded is a free log retrieval operation binding the contract event 0x29dd5553eda23e846a442697aeea6662a9699d1a79bb82afba4ba8994898b92c.
//
// Solidity: event SupportedTokenAdded(address indexed token, uint8 tokenAbility)
func (_Portal *PortalFilterer) FilterSupportedTokenAdded(opts *bind.FilterOpts, token []common.Address) (*PortalSupportedTokenAddedIterator, error) {

	var tokenRule []interface{}
	for _, tokenItem := range token {
		tokenRule = append(tokenRule, tokenItem)
	}

	logs, sub, err := _Portal.contract.FilterLogs(opts, "SupportedTokenAdded", tokenRule)
	if err != nil {
		return nil, err
	}
	return &PortalSupportedTokenAddedIterator{contract: _Portal.contract, event: "SupportedTokenAdded", logs: logs, sub: sub}, nil
}

// WatchSupportedTokenAdded is a free log subscription operation binding the contract event 0x29dd5553eda23e846a442697aeea6662a9699d1a79bb82afba4ba8994898b92c.
//
// Solidity: event SupportedTokenAdded(address indexed token, uint8 tokenAbility)
func (_Portal *PortalFilterer) WatchSupportedTokenAdded(opts *bind.WatchOpts, sink chan<- *PortalSupportedTokenAdded, token []common.Address) (event.Subscription, error) {

	var tokenRule []interface{}
	for _, tokenItem := range token {
		tokenRule = append(tokenRule, tokenItem)
	}

	logs, sub, err := _Portal.contract.WatchLogs(opts, "SupportedTokenAdded", tokenRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(PortalSupportedTokenAdded)
				if err := _Portal.contract.UnpackLog(event, "SupportedTokenAdded", log); err != nil {
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

// ParseSupportedTokenAdded is a log parse operation binding the contract event 0x29dd5553eda23e846a442697aeea6662a9699d1a79bb82afba4ba8994898b92c.
//
// Solidity: event SupportedTokenAdded(address indexed token, uint8 tokenAbility)
func (_Portal *PortalFilterer) ParseSupportedTokenAdded(log types.Log) (*PortalSupportedTokenAdded, error) {
	event := new(PortalSupportedTokenAdded)
	if err := _Portal.contract.UnpackLog(event, "SupportedTokenAdded", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// PortalWithdrawnIterator is returned from FilterWithdrawn and is used to iterate over the raw logs and unpacked data for Withdrawn events raised by the Portal contract.
type PortalWithdrawnIterator struct {
	Event *PortalWithdrawn // Event containing the contract specifics and raw log

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
func (it *PortalWithdrawnIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(PortalWithdrawn)
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
		it.Event = new(PortalWithdrawn)
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
func (it *PortalWithdrawnIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *PortalWithdrawnIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// PortalWithdrawn represents a Withdrawn event raised by the Portal contract.
type PortalWithdrawn struct {
	Depositor common.Address
	Token     common.Address
	DepositId *big.Int
	Amount    *big.Int
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterWithdrawn is a free log retrieval operation binding the contract event 0x91fb9d98b786c57d74c099ccd2beca1739e9f6a81fb49001ca465c4b7591bbe2.
//
// Solidity: event Withdrawn(address indexed depositor, address indexed token, uint256 indexed depositId, uint256 amount)
func (_Portal *PortalFilterer) FilterWithdrawn(opts *bind.FilterOpts, depositor []common.Address, token []common.Address, depositId []*big.Int) (*PortalWithdrawnIterator, error) {

	var depositorRule []interface{}
	for _, depositorItem := range depositor {
		depositorRule = append(depositorRule, depositorItem)
	}
	var tokenRule []interface{}
	for _, tokenItem := range token {
		tokenRule = append(tokenRule, tokenItem)
	}
	var depositIdRule []interface{}
	for _, depositIdItem := range depositId {
		depositIdRule = append(depositIdRule, depositIdItem)
	}

	logs, sub, err := _Portal.contract.FilterLogs(opts, "Withdrawn", depositorRule, tokenRule, depositIdRule)
	if err != nil {
		return nil, err
	}
	return &PortalWithdrawnIterator{contract: _Portal.contract, event: "Withdrawn", logs: logs, sub: sub}, nil
}

// WatchWithdrawn is a free log subscription operation binding the contract event 0x91fb9d98b786c57d74c099ccd2beca1739e9f6a81fb49001ca465c4b7591bbe2.
//
// Solidity: event Withdrawn(address indexed depositor, address indexed token, uint256 indexed depositId, uint256 amount)
func (_Portal *PortalFilterer) WatchWithdrawn(opts *bind.WatchOpts, sink chan<- *PortalWithdrawn, depositor []common.Address, token []common.Address, depositId []*big.Int) (event.Subscription, error) {

	var depositorRule []interface{}
	for _, depositorItem := range depositor {
		depositorRule = append(depositorRule, depositorItem)
	}
	var tokenRule []interface{}
	for _, tokenItem := range token {
		tokenRule = append(tokenRule, tokenItem)
	}
	var depositIdRule []interface{}
	for _, depositIdItem := range depositId {
		depositIdRule = append(depositIdRule, depositIdItem)
	}

	logs, sub, err := _Portal.contract.WatchLogs(opts, "Withdrawn", depositorRule, tokenRule, depositIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(PortalWithdrawn)
				if err := _Portal.contract.UnpackLog(event, "Withdrawn", log); err != nil {
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

// ParseWithdrawn is a log parse operation binding the contract event 0x91fb9d98b786c57d74c099ccd2beca1739e9f6a81fb49001ca465c4b7591bbe2.
//
// Solidity: event Withdrawn(address indexed depositor, address indexed token, uint256 indexed depositId, uint256 amount)
func (_Portal *PortalFilterer) ParseWithdrawn(log types.Log) (*PortalWithdrawn, error) {
	event := new(PortalWithdrawn)
	if err := _Portal.contract.UnpackLog(event, "Withdrawn", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
