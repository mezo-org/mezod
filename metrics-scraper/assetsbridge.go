// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package metricsscraper

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

// AssetsLocked is an auto generated low-level Go binding around an user-defined struct.
type AssetsLocked struct {
	SequenceNumber *big.Int
	Recipient      common.Address
	Amount         *big.Int
	Token          common.Address
}

// ERC20TokenMapping is an auto generated low-level Go binding around an user-defined struct.
type ERC20TokenMapping struct {
	SourceToken common.Address
	MezoToken   common.Address
}

// AssetsBridgeMetaData contains all meta data concerning the AssetsBridge contract.
var AssetsBridgeMetaData = &bind.MetaData{
	ABI: "[{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"unlockSequenceNumber\",\"type\":\"uint256\"},{\"indexed\":true,\"internalType\":\"bytes\",\"name\":\"recipient\",\"type\":\"bytes\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"token\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"address\",\"name\":\"sender\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint8\",\"name\":\"chain\",\"type\":\"uint8\"}],\"name\":\"AssetsUnlocked\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"sourceToken\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"mezoToken\",\"type\":\"address\"}],\"name\":\"ERC20TokenMappingCreated\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"sourceToken\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"mezoToken\",\"type\":\"address\"}],\"name\":\"ERC20TokenMappingDeleted\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"mezoToken\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"minAmount\",\"type\":\"uint256\"}],\"name\":\"MinBridgeOutAmountSet\",\"type\":\"event\"},{\"inputs\":[{\"components\":[{\"internalType\":\"uint256\",\"name\":\"sequenceNumber\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"recipient\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"token\",\"type\":\"address\"}],\"internalType\":\"structAssetsLocked[]\",\"name\":\"events\",\"type\":\"tuple[]\"}],\"name\":\"bridge\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"token\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"internalType\":\"uint8\",\"name\":\"chain\",\"type\":\"uint8\"},{\"internalType\":\"bytes\",\"name\":\"recipient\",\"type\":\"bytes\"}],\"name\":\"bridgeOut\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"sourceToken\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"mezoToken\",\"type\":\"address\"}],\"name\":\"createERC20TokenMapping\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"sourceToken\",\"type\":\"address\"}],\"name\":\"deleteERC20TokenMapping\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getCurrentSequenceTip\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"sourceToken\",\"type\":\"address\"}],\"name\":\"getERC20TokenMapping\",\"outputs\":[{\"components\":[{\"internalType\":\"address\",\"name\":\"sourceToken\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"mezoToken\",\"type\":\"address\"}],\"internalType\":\"structERC20TokenMapping\",\"name\":\"\",\"type\":\"tuple\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getERC20TokensMappings\",\"outputs\":[{\"components\":[{\"internalType\":\"address\",\"name\":\"sourceToken\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"mezoToken\",\"type\":\"address\"}],\"internalType\":\"structERC20TokenMapping[]\",\"name\":\"\",\"type\":\"tuple[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getMaxERC20TokensMappings\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"mezoToken\",\"type\":\"address\"}],\"name\":\"getMinBridgeOutAmount\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"token\",\"type\":\"address\"}],\"name\":\"getOutflowCapacity\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"capacity\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"resetHeight\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"token\",\"type\":\"address\"}],\"name\":\"getOutflowLimit\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getPauser\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getSourceBTCToken\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"pauseBridgeOut\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"mezoToken\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"minAmount\",\"type\":\"uint256\"}],\"name\":\"setMinBridgeOutAmount\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"token\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"limit\",\"type\":\"uint256\"}],\"name\":\"setOutflowLimit\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"pauser\",\"type\":\"address\"}],\"name\":\"setPauser\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]",
}

// AssetsBridgeABI is the input ABI used to generate the binding from.
// Deprecated: Use AssetsBridgeMetaData.ABI instead.
var AssetsBridgeABI = AssetsBridgeMetaData.ABI

// AssetsBridge is an auto generated Go binding around an Ethereum contract.
type AssetsBridge struct {
	AssetsBridgeCaller     // Read-only binding to the contract
	AssetsBridgeTransactor // Write-only binding to the contract
	AssetsBridgeFilterer   // Log filterer for contract events
}

// AssetsBridgeCaller is an auto generated read-only Go binding around an Ethereum contract.
type AssetsBridgeCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// AssetsBridgeTransactor is an auto generated write-only Go binding around an Ethereum contract.
type AssetsBridgeTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// AssetsBridgeFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type AssetsBridgeFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// AssetsBridgeSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type AssetsBridgeSession struct {
	Contract     *AssetsBridge     // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// AssetsBridgeCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type AssetsBridgeCallerSession struct {
	Contract *AssetsBridgeCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts       // Call options to use throughout this session
}

// AssetsBridgeTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type AssetsBridgeTransactorSession struct {
	Contract     *AssetsBridgeTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts       // Transaction auth options to use throughout this session
}

// AssetsBridgeRaw is an auto generated low-level Go binding around an Ethereum contract.
type AssetsBridgeRaw struct {
	Contract *AssetsBridge // Generic contract binding to access the raw methods on
}

// AssetsBridgeCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type AssetsBridgeCallerRaw struct {
	Contract *AssetsBridgeCaller // Generic read-only contract binding to access the raw methods on
}

// AssetsBridgeTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type AssetsBridgeTransactorRaw struct {
	Contract *AssetsBridgeTransactor // Generic write-only contract binding to access the raw methods on
}

// NewAssetsBridge creates a new instance of AssetsBridge, bound to a specific deployed contract.
func NewAssetsBridge(address common.Address, backend bind.ContractBackend) (*AssetsBridge, error) {
	contract, err := bindAssetsBridge(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &AssetsBridge{AssetsBridgeCaller: AssetsBridgeCaller{contract: contract}, AssetsBridgeTransactor: AssetsBridgeTransactor{contract: contract}, AssetsBridgeFilterer: AssetsBridgeFilterer{contract: contract}}, nil
}

// NewAssetsBridgeCaller creates a new read-only instance of AssetsBridge, bound to a specific deployed contract.
func NewAssetsBridgeCaller(address common.Address, caller bind.ContractCaller) (*AssetsBridgeCaller, error) {
	contract, err := bindAssetsBridge(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &AssetsBridgeCaller{contract: contract}, nil
}

// NewAssetsBridgeTransactor creates a new write-only instance of AssetsBridge, bound to a specific deployed contract.
func NewAssetsBridgeTransactor(address common.Address, transactor bind.ContractTransactor) (*AssetsBridgeTransactor, error) {
	contract, err := bindAssetsBridge(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &AssetsBridgeTransactor{contract: contract}, nil
}

// NewAssetsBridgeFilterer creates a new log filterer instance of AssetsBridge, bound to a specific deployed contract.
func NewAssetsBridgeFilterer(address common.Address, filterer bind.ContractFilterer) (*AssetsBridgeFilterer, error) {
	contract, err := bindAssetsBridge(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &AssetsBridgeFilterer{contract: contract}, nil
}

// bindAssetsBridge binds a generic wrapper to an already deployed contract.
func bindAssetsBridge(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(AssetsBridgeABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_AssetsBridge *AssetsBridgeRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _AssetsBridge.Contract.AssetsBridgeCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_AssetsBridge *AssetsBridgeRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _AssetsBridge.Contract.AssetsBridgeTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_AssetsBridge *AssetsBridgeRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _AssetsBridge.Contract.AssetsBridgeTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_AssetsBridge *AssetsBridgeCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _AssetsBridge.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_AssetsBridge *AssetsBridgeTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _AssetsBridge.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_AssetsBridge *AssetsBridgeTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _AssetsBridge.Contract.contract.Transact(opts, method, params...)
}

// GetCurrentSequenceTip is a free data retrieval call binding the contract method 0xc2eddfea.
//
// Solidity: function getCurrentSequenceTip() view returns(uint256)
func (_AssetsBridge *AssetsBridgeCaller) GetCurrentSequenceTip(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _AssetsBridge.contract.Call(opts, &out, "getCurrentSequenceTip")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetCurrentSequenceTip is a free data retrieval call binding the contract method 0xc2eddfea.
//
// Solidity: function getCurrentSequenceTip() view returns(uint256)
func (_AssetsBridge *AssetsBridgeSession) GetCurrentSequenceTip() (*big.Int, error) {
	return _AssetsBridge.Contract.GetCurrentSequenceTip(&_AssetsBridge.CallOpts)
}

// GetCurrentSequenceTip is a free data retrieval call binding the contract method 0xc2eddfea.
//
// Solidity: function getCurrentSequenceTip() view returns(uint256)
func (_AssetsBridge *AssetsBridgeCallerSession) GetCurrentSequenceTip() (*big.Int, error) {
	return _AssetsBridge.Contract.GetCurrentSequenceTip(&_AssetsBridge.CallOpts)
}

// GetERC20TokenMapping is a free data retrieval call binding the contract method 0x48a451c3.
//
// Solidity: function getERC20TokenMapping(address sourceToken) view returns((address,address))
func (_AssetsBridge *AssetsBridgeCaller) GetERC20TokenMapping(opts *bind.CallOpts, sourceToken common.Address) (ERC20TokenMapping, error) {
	var out []interface{}
	err := _AssetsBridge.contract.Call(opts, &out, "getERC20TokenMapping", sourceToken)

	if err != nil {
		return *new(ERC20TokenMapping), err
	}

	out0 := *abi.ConvertType(out[0], new(ERC20TokenMapping)).(*ERC20TokenMapping)

	return out0, err

}

// GetERC20TokenMapping is a free data retrieval call binding the contract method 0x48a451c3.
//
// Solidity: function getERC20TokenMapping(address sourceToken) view returns((address,address))
func (_AssetsBridge *AssetsBridgeSession) GetERC20TokenMapping(sourceToken common.Address) (ERC20TokenMapping, error) {
	return _AssetsBridge.Contract.GetERC20TokenMapping(&_AssetsBridge.CallOpts, sourceToken)
}

// GetERC20TokenMapping is a free data retrieval call binding the contract method 0x48a451c3.
//
// Solidity: function getERC20TokenMapping(address sourceToken) view returns((address,address))
func (_AssetsBridge *AssetsBridgeCallerSession) GetERC20TokenMapping(sourceToken common.Address) (ERC20TokenMapping, error) {
	return _AssetsBridge.Contract.GetERC20TokenMapping(&_AssetsBridge.CallOpts, sourceToken)
}

// GetERC20TokensMappings is a free data retrieval call binding the contract method 0x7d1f3add.
//
// Solidity: function getERC20TokensMappings() view returns((address,address)[])
func (_AssetsBridge *AssetsBridgeCaller) GetERC20TokensMappings(opts *bind.CallOpts) ([]ERC20TokenMapping, error) {
	var out []interface{}
	err := _AssetsBridge.contract.Call(opts, &out, "getERC20TokensMappings")

	if err != nil {
		return *new([]ERC20TokenMapping), err
	}

	out0 := *abi.ConvertType(out[0], new([]ERC20TokenMapping)).(*[]ERC20TokenMapping)

	return out0, err

}

// GetERC20TokensMappings is a free data retrieval call binding the contract method 0x7d1f3add.
//
// Solidity: function getERC20TokensMappings() view returns((address,address)[])
func (_AssetsBridge *AssetsBridgeSession) GetERC20TokensMappings() ([]ERC20TokenMapping, error) {
	return _AssetsBridge.Contract.GetERC20TokensMappings(&_AssetsBridge.CallOpts)
}

// GetERC20TokensMappings is a free data retrieval call binding the contract method 0x7d1f3add.
//
// Solidity: function getERC20TokensMappings() view returns((address,address)[])
func (_AssetsBridge *AssetsBridgeCallerSession) GetERC20TokensMappings() ([]ERC20TokenMapping, error) {
	return _AssetsBridge.Contract.GetERC20TokensMappings(&_AssetsBridge.CallOpts)
}

// GetMaxERC20TokensMappings is a free data retrieval call binding the contract method 0xb6b3b89f.
//
// Solidity: function getMaxERC20TokensMappings() view returns(uint256)
func (_AssetsBridge *AssetsBridgeCaller) GetMaxERC20TokensMappings(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _AssetsBridge.contract.Call(opts, &out, "getMaxERC20TokensMappings")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetMaxERC20TokensMappings is a free data retrieval call binding the contract method 0xb6b3b89f.
//
// Solidity: function getMaxERC20TokensMappings() view returns(uint256)
func (_AssetsBridge *AssetsBridgeSession) GetMaxERC20TokensMappings() (*big.Int, error) {
	return _AssetsBridge.Contract.GetMaxERC20TokensMappings(&_AssetsBridge.CallOpts)
}

// GetMaxERC20TokensMappings is a free data retrieval call binding the contract method 0xb6b3b89f.
//
// Solidity: function getMaxERC20TokensMappings() view returns(uint256)
func (_AssetsBridge *AssetsBridgeCallerSession) GetMaxERC20TokensMappings() (*big.Int, error) {
	return _AssetsBridge.Contract.GetMaxERC20TokensMappings(&_AssetsBridge.CallOpts)
}

// GetMinBridgeOutAmount is a free data retrieval call binding the contract method 0x419b3009.
//
// Solidity: function getMinBridgeOutAmount(address mezoToken) view returns(uint256)
func (_AssetsBridge *AssetsBridgeCaller) GetMinBridgeOutAmount(opts *bind.CallOpts, mezoToken common.Address) (*big.Int, error) {
	var out []interface{}
	err := _AssetsBridge.contract.Call(opts, &out, "getMinBridgeOutAmount", mezoToken)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetMinBridgeOutAmount is a free data retrieval call binding the contract method 0x419b3009.
//
// Solidity: function getMinBridgeOutAmount(address mezoToken) view returns(uint256)
func (_AssetsBridge *AssetsBridgeSession) GetMinBridgeOutAmount(mezoToken common.Address) (*big.Int, error) {
	return _AssetsBridge.Contract.GetMinBridgeOutAmount(&_AssetsBridge.CallOpts, mezoToken)
}

// GetMinBridgeOutAmount is a free data retrieval call binding the contract method 0x419b3009.
//
// Solidity: function getMinBridgeOutAmount(address mezoToken) view returns(uint256)
func (_AssetsBridge *AssetsBridgeCallerSession) GetMinBridgeOutAmount(mezoToken common.Address) (*big.Int, error) {
	return _AssetsBridge.Contract.GetMinBridgeOutAmount(&_AssetsBridge.CallOpts, mezoToken)
}

// GetOutflowCapacity is a free data retrieval call binding the contract method 0xd7fa1750.
//
// Solidity: function getOutflowCapacity(address token) view returns(uint256 capacity, uint256 resetHeight)
func (_AssetsBridge *AssetsBridgeCaller) GetOutflowCapacity(opts *bind.CallOpts, token common.Address) (struct {
	Capacity    *big.Int
	ResetHeight *big.Int
}, error) {
	var out []interface{}
	err := _AssetsBridge.contract.Call(opts, &out, "getOutflowCapacity", token)

	outstruct := new(struct {
		Capacity    *big.Int
		ResetHeight *big.Int
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.Capacity = *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)
	outstruct.ResetHeight = *abi.ConvertType(out[1], new(*big.Int)).(**big.Int)

	return *outstruct, err

}

// GetOutflowCapacity is a free data retrieval call binding the contract method 0xd7fa1750.
//
// Solidity: function getOutflowCapacity(address token) view returns(uint256 capacity, uint256 resetHeight)
func (_AssetsBridge *AssetsBridgeSession) GetOutflowCapacity(token common.Address) (struct {
	Capacity    *big.Int
	ResetHeight *big.Int
}, error) {
	return _AssetsBridge.Contract.GetOutflowCapacity(&_AssetsBridge.CallOpts, token)
}

// GetOutflowCapacity is a free data retrieval call binding the contract method 0xd7fa1750.
//
// Solidity: function getOutflowCapacity(address token) view returns(uint256 capacity, uint256 resetHeight)
func (_AssetsBridge *AssetsBridgeCallerSession) GetOutflowCapacity(token common.Address) (struct {
	Capacity    *big.Int
	ResetHeight *big.Int
}, error) {
	return _AssetsBridge.Contract.GetOutflowCapacity(&_AssetsBridge.CallOpts, token)
}

// GetOutflowLimit is a free data retrieval call binding the contract method 0x69475e7e.
//
// Solidity: function getOutflowLimit(address token) view returns(uint256)
func (_AssetsBridge *AssetsBridgeCaller) GetOutflowLimit(opts *bind.CallOpts, token common.Address) (*big.Int, error) {
	var out []interface{}
	err := _AssetsBridge.contract.Call(opts, &out, "getOutflowLimit", token)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetOutflowLimit is a free data retrieval call binding the contract method 0x69475e7e.
//
// Solidity: function getOutflowLimit(address token) view returns(uint256)
func (_AssetsBridge *AssetsBridgeSession) GetOutflowLimit(token common.Address) (*big.Int, error) {
	return _AssetsBridge.Contract.GetOutflowLimit(&_AssetsBridge.CallOpts, token)
}

// GetOutflowLimit is a free data retrieval call binding the contract method 0x69475e7e.
//
// Solidity: function getOutflowLimit(address token) view returns(uint256)
func (_AssetsBridge *AssetsBridgeCallerSession) GetOutflowLimit(token common.Address) (*big.Int, error) {
	return _AssetsBridge.Contract.GetOutflowLimit(&_AssetsBridge.CallOpts, token)
}

// GetPauser is a free data retrieval call binding the contract method 0x7008b548.
//
// Solidity: function getPauser() view returns(address)
func (_AssetsBridge *AssetsBridgeCaller) GetPauser(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _AssetsBridge.contract.Call(opts, &out, "getPauser")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// GetPauser is a free data retrieval call binding the contract method 0x7008b548.
//
// Solidity: function getPauser() view returns(address)
func (_AssetsBridge *AssetsBridgeSession) GetPauser() (common.Address, error) {
	return _AssetsBridge.Contract.GetPauser(&_AssetsBridge.CallOpts)
}

// GetPauser is a free data retrieval call binding the contract method 0x7008b548.
//
// Solidity: function getPauser() view returns(address)
func (_AssetsBridge *AssetsBridgeCallerSession) GetPauser() (common.Address, error) {
	return _AssetsBridge.Contract.GetPauser(&_AssetsBridge.CallOpts)
}

// GetSourceBTCToken is a free data retrieval call binding the contract method 0x3d236363.
//
// Solidity: function getSourceBTCToken() view returns(address)
func (_AssetsBridge *AssetsBridgeCaller) GetSourceBTCToken(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _AssetsBridge.contract.Call(opts, &out, "getSourceBTCToken")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// GetSourceBTCToken is a free data retrieval call binding the contract method 0x3d236363.
//
// Solidity: function getSourceBTCToken() view returns(address)
func (_AssetsBridge *AssetsBridgeSession) GetSourceBTCToken() (common.Address, error) {
	return _AssetsBridge.Contract.GetSourceBTCToken(&_AssetsBridge.CallOpts)
}

// GetSourceBTCToken is a free data retrieval call binding the contract method 0x3d236363.
//
// Solidity: function getSourceBTCToken() view returns(address)
func (_AssetsBridge *AssetsBridgeCallerSession) GetSourceBTCToken() (common.Address, error) {
	return _AssetsBridge.Contract.GetSourceBTCToken(&_AssetsBridge.CallOpts)
}

// Bridge is a paid mutator transaction binding the contract method 0xc0b3cc19.
//
// Solidity: function bridge((uint256,address,uint256,address)[] events) returns(bool)
func (_AssetsBridge *AssetsBridgeTransactor) Bridge(opts *bind.TransactOpts, events []AssetsLocked) (*types.Transaction, error) {
	return _AssetsBridge.contract.Transact(opts, "bridge", events)
}

// Bridge is a paid mutator transaction binding the contract method 0xc0b3cc19.
//
// Solidity: function bridge((uint256,address,uint256,address)[] events) returns(bool)
func (_AssetsBridge *AssetsBridgeSession) Bridge(events []AssetsLocked) (*types.Transaction, error) {
	return _AssetsBridge.Contract.Bridge(&_AssetsBridge.TransactOpts, events)
}

// Bridge is a paid mutator transaction binding the contract method 0xc0b3cc19.
//
// Solidity: function bridge((uint256,address,uint256,address)[] events) returns(bool)
func (_AssetsBridge *AssetsBridgeTransactorSession) Bridge(events []AssetsLocked) (*types.Transaction, error) {
	return _AssetsBridge.Contract.Bridge(&_AssetsBridge.TransactOpts, events)
}

// BridgeOut is a paid mutator transaction binding the contract method 0xff39fd04.
//
// Solidity: function bridgeOut(address token, uint256 amount, uint8 chain, bytes recipient) returns(bool)
func (_AssetsBridge *AssetsBridgeTransactor) BridgeOut(opts *bind.TransactOpts, token common.Address, amount *big.Int, chain uint8, recipient []byte) (*types.Transaction, error) {
	return _AssetsBridge.contract.Transact(opts, "bridgeOut", token, amount, chain, recipient)
}

// BridgeOut is a paid mutator transaction binding the contract method 0xff39fd04.
//
// Solidity: function bridgeOut(address token, uint256 amount, uint8 chain, bytes recipient) returns(bool)
func (_AssetsBridge *AssetsBridgeSession) BridgeOut(token common.Address, amount *big.Int, chain uint8, recipient []byte) (*types.Transaction, error) {
	return _AssetsBridge.Contract.BridgeOut(&_AssetsBridge.TransactOpts, token, amount, chain, recipient)
}

// BridgeOut is a paid mutator transaction binding the contract method 0xff39fd04.
//
// Solidity: function bridgeOut(address token, uint256 amount, uint8 chain, bytes recipient) returns(bool)
func (_AssetsBridge *AssetsBridgeTransactorSession) BridgeOut(token common.Address, amount *big.Int, chain uint8, recipient []byte) (*types.Transaction, error) {
	return _AssetsBridge.Contract.BridgeOut(&_AssetsBridge.TransactOpts, token, amount, chain, recipient)
}

// CreateERC20TokenMapping is a paid mutator transaction binding the contract method 0x3b5586eb.
//
// Solidity: function createERC20TokenMapping(address sourceToken, address mezoToken) returns(bool)
func (_AssetsBridge *AssetsBridgeTransactor) CreateERC20TokenMapping(opts *bind.TransactOpts, sourceToken common.Address, mezoToken common.Address) (*types.Transaction, error) {
	return _AssetsBridge.contract.Transact(opts, "createERC20TokenMapping", sourceToken, mezoToken)
}

// CreateERC20TokenMapping is a paid mutator transaction binding the contract method 0x3b5586eb.
//
// Solidity: function createERC20TokenMapping(address sourceToken, address mezoToken) returns(bool)
func (_AssetsBridge *AssetsBridgeSession) CreateERC20TokenMapping(sourceToken common.Address, mezoToken common.Address) (*types.Transaction, error) {
	return _AssetsBridge.Contract.CreateERC20TokenMapping(&_AssetsBridge.TransactOpts, sourceToken, mezoToken)
}

// CreateERC20TokenMapping is a paid mutator transaction binding the contract method 0x3b5586eb.
//
// Solidity: function createERC20TokenMapping(address sourceToken, address mezoToken) returns(bool)
func (_AssetsBridge *AssetsBridgeTransactorSession) CreateERC20TokenMapping(sourceToken common.Address, mezoToken common.Address) (*types.Transaction, error) {
	return _AssetsBridge.Contract.CreateERC20TokenMapping(&_AssetsBridge.TransactOpts, sourceToken, mezoToken)
}

// DeleteERC20TokenMapping is a paid mutator transaction binding the contract method 0x26c66d05.
//
// Solidity: function deleteERC20TokenMapping(address sourceToken) returns(bool)
func (_AssetsBridge *AssetsBridgeTransactor) DeleteERC20TokenMapping(opts *bind.TransactOpts, sourceToken common.Address) (*types.Transaction, error) {
	return _AssetsBridge.contract.Transact(opts, "deleteERC20TokenMapping", sourceToken)
}

// DeleteERC20TokenMapping is a paid mutator transaction binding the contract method 0x26c66d05.
//
// Solidity: function deleteERC20TokenMapping(address sourceToken) returns(bool)
func (_AssetsBridge *AssetsBridgeSession) DeleteERC20TokenMapping(sourceToken common.Address) (*types.Transaction, error) {
	return _AssetsBridge.Contract.DeleteERC20TokenMapping(&_AssetsBridge.TransactOpts, sourceToken)
}

// DeleteERC20TokenMapping is a paid mutator transaction binding the contract method 0x26c66d05.
//
// Solidity: function deleteERC20TokenMapping(address sourceToken) returns(bool)
func (_AssetsBridge *AssetsBridgeTransactorSession) DeleteERC20TokenMapping(sourceToken common.Address) (*types.Transaction, error) {
	return _AssetsBridge.Contract.DeleteERC20TokenMapping(&_AssetsBridge.TransactOpts, sourceToken)
}

// PauseBridgeOut is a paid mutator transaction binding the contract method 0x2f1d448f.
//
// Solidity: function pauseBridgeOut() returns(bool)
func (_AssetsBridge *AssetsBridgeTransactor) PauseBridgeOut(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _AssetsBridge.contract.Transact(opts, "pauseBridgeOut")
}

// PauseBridgeOut is a paid mutator transaction binding the contract method 0x2f1d448f.
//
// Solidity: function pauseBridgeOut() returns(bool)
func (_AssetsBridge *AssetsBridgeSession) PauseBridgeOut() (*types.Transaction, error) {
	return _AssetsBridge.Contract.PauseBridgeOut(&_AssetsBridge.TransactOpts)
}

// PauseBridgeOut is a paid mutator transaction binding the contract method 0x2f1d448f.
//
// Solidity: function pauseBridgeOut() returns(bool)
func (_AssetsBridge *AssetsBridgeTransactorSession) PauseBridgeOut() (*types.Transaction, error) {
	return _AssetsBridge.Contract.PauseBridgeOut(&_AssetsBridge.TransactOpts)
}

// SetMinBridgeOutAmount is a paid mutator transaction binding the contract method 0x9ea147ea.
//
// Solidity: function setMinBridgeOutAmount(address mezoToken, uint256 minAmount) returns(bool)
func (_AssetsBridge *AssetsBridgeTransactor) SetMinBridgeOutAmount(opts *bind.TransactOpts, mezoToken common.Address, minAmount *big.Int) (*types.Transaction, error) {
	return _AssetsBridge.contract.Transact(opts, "setMinBridgeOutAmount", mezoToken, minAmount)
}

// SetMinBridgeOutAmount is a paid mutator transaction binding the contract method 0x9ea147ea.
//
// Solidity: function setMinBridgeOutAmount(address mezoToken, uint256 minAmount) returns(bool)
func (_AssetsBridge *AssetsBridgeSession) SetMinBridgeOutAmount(mezoToken common.Address, minAmount *big.Int) (*types.Transaction, error) {
	return _AssetsBridge.Contract.SetMinBridgeOutAmount(&_AssetsBridge.TransactOpts, mezoToken, minAmount)
}

// SetMinBridgeOutAmount is a paid mutator transaction binding the contract method 0x9ea147ea.
//
// Solidity: function setMinBridgeOutAmount(address mezoToken, uint256 minAmount) returns(bool)
func (_AssetsBridge *AssetsBridgeTransactorSession) SetMinBridgeOutAmount(mezoToken common.Address, minAmount *big.Int) (*types.Transaction, error) {
	return _AssetsBridge.Contract.SetMinBridgeOutAmount(&_AssetsBridge.TransactOpts, mezoToken, minAmount)
}

// SetOutflowLimit is a paid mutator transaction binding the contract method 0x0f49bc7b.
//
// Solidity: function setOutflowLimit(address token, uint256 limit) returns(bool)
func (_AssetsBridge *AssetsBridgeTransactor) SetOutflowLimit(opts *bind.TransactOpts, token common.Address, limit *big.Int) (*types.Transaction, error) {
	return _AssetsBridge.contract.Transact(opts, "setOutflowLimit", token, limit)
}

// SetOutflowLimit is a paid mutator transaction binding the contract method 0x0f49bc7b.
//
// Solidity: function setOutflowLimit(address token, uint256 limit) returns(bool)
func (_AssetsBridge *AssetsBridgeSession) SetOutflowLimit(token common.Address, limit *big.Int) (*types.Transaction, error) {
	return _AssetsBridge.Contract.SetOutflowLimit(&_AssetsBridge.TransactOpts, token, limit)
}

// SetOutflowLimit is a paid mutator transaction binding the contract method 0x0f49bc7b.
//
// Solidity: function setOutflowLimit(address token, uint256 limit) returns(bool)
func (_AssetsBridge *AssetsBridgeTransactorSession) SetOutflowLimit(token common.Address, limit *big.Int) (*types.Transaction, error) {
	return _AssetsBridge.Contract.SetOutflowLimit(&_AssetsBridge.TransactOpts, token, limit)
}

// SetPauser is a paid mutator transaction binding the contract method 0x2d88af4a.
//
// Solidity: function setPauser(address pauser) returns(bool)
func (_AssetsBridge *AssetsBridgeTransactor) SetPauser(opts *bind.TransactOpts, pauser common.Address) (*types.Transaction, error) {
	return _AssetsBridge.contract.Transact(opts, "setPauser", pauser)
}

// SetPauser is a paid mutator transaction binding the contract method 0x2d88af4a.
//
// Solidity: function setPauser(address pauser) returns(bool)
func (_AssetsBridge *AssetsBridgeSession) SetPauser(pauser common.Address) (*types.Transaction, error) {
	return _AssetsBridge.Contract.SetPauser(&_AssetsBridge.TransactOpts, pauser)
}

// SetPauser is a paid mutator transaction binding the contract method 0x2d88af4a.
//
// Solidity: function setPauser(address pauser) returns(bool)
func (_AssetsBridge *AssetsBridgeTransactorSession) SetPauser(pauser common.Address) (*types.Transaction, error) {
	return _AssetsBridge.Contract.SetPauser(&_AssetsBridge.TransactOpts, pauser)
}

// AssetsBridgeAssetsUnlockedIterator is returned from FilterAssetsUnlocked and is used to iterate over the raw logs and unpacked data for AssetsUnlocked events raised by the AssetsBridge contract.
type AssetsBridgeAssetsUnlockedIterator struct {
	Event *AssetsBridgeAssetsUnlocked // Event containing the contract specifics and raw log

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
func (it *AssetsBridgeAssetsUnlockedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AssetsBridgeAssetsUnlocked)
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
		it.Event = new(AssetsBridgeAssetsUnlocked)
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
func (it *AssetsBridgeAssetsUnlockedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AssetsBridgeAssetsUnlockedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AssetsBridgeAssetsUnlocked represents a AssetsUnlocked event raised by the AssetsBridge contract.
type AssetsBridgeAssetsUnlocked struct {
	UnlockSequenceNumber *big.Int
	Recipient            common.Hash
	Token                common.Address
	Sender               common.Address
	Amount               *big.Int
	Chain                uint8
	Raw                  types.Log // Blockchain specific contextual infos
}

// FilterAssetsUnlocked is a free log retrieval operation binding the contract event 0x018e20d97c5c8ad507a796cdc5eadfee3d30de1fa9b13231094211f18134e7c8.
//
// Solidity: event AssetsUnlocked(uint256 indexed unlockSequenceNumber, bytes indexed recipient, address indexed token, address sender, uint256 amount, uint8 chain)
func (_AssetsBridge *AssetsBridgeFilterer) FilterAssetsUnlocked(opts *bind.FilterOpts, unlockSequenceNumber []*big.Int, recipient [][]byte, token []common.Address) (*AssetsBridgeAssetsUnlockedIterator, error) {

	var unlockSequenceNumberRule []interface{}
	for _, unlockSequenceNumberItem := range unlockSequenceNumber {
		unlockSequenceNumberRule = append(unlockSequenceNumberRule, unlockSequenceNumberItem)
	}
	var recipientRule []interface{}
	for _, recipientItem := range recipient {
		recipientRule = append(recipientRule, recipientItem)
	}
	var tokenRule []interface{}
	for _, tokenItem := range token {
		tokenRule = append(tokenRule, tokenItem)
	}

	logs, sub, err := _AssetsBridge.contract.FilterLogs(opts, "AssetsUnlocked", unlockSequenceNumberRule, recipientRule, tokenRule)
	if err != nil {
		return nil, err
	}
	return &AssetsBridgeAssetsUnlockedIterator{contract: _AssetsBridge.contract, event: "AssetsUnlocked", logs: logs, sub: sub}, nil
}

// WatchAssetsUnlocked is a free log subscription operation binding the contract event 0x018e20d97c5c8ad507a796cdc5eadfee3d30de1fa9b13231094211f18134e7c8.
//
// Solidity: event AssetsUnlocked(uint256 indexed unlockSequenceNumber, bytes indexed recipient, address indexed token, address sender, uint256 amount, uint8 chain)
func (_AssetsBridge *AssetsBridgeFilterer) WatchAssetsUnlocked(opts *bind.WatchOpts, sink chan<- *AssetsBridgeAssetsUnlocked, unlockSequenceNumber []*big.Int, recipient [][]byte, token []common.Address) (event.Subscription, error) {

	var unlockSequenceNumberRule []interface{}
	for _, unlockSequenceNumberItem := range unlockSequenceNumber {
		unlockSequenceNumberRule = append(unlockSequenceNumberRule, unlockSequenceNumberItem)
	}
	var recipientRule []interface{}
	for _, recipientItem := range recipient {
		recipientRule = append(recipientRule, recipientItem)
	}
	var tokenRule []interface{}
	for _, tokenItem := range token {
		tokenRule = append(tokenRule, tokenItem)
	}

	logs, sub, err := _AssetsBridge.contract.WatchLogs(opts, "AssetsUnlocked", unlockSequenceNumberRule, recipientRule, tokenRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AssetsBridgeAssetsUnlocked)
				if err := _AssetsBridge.contract.UnpackLog(event, "AssetsUnlocked", log); err != nil {
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

// ParseAssetsUnlocked is a log parse operation binding the contract event 0x018e20d97c5c8ad507a796cdc5eadfee3d30de1fa9b13231094211f18134e7c8.
//
// Solidity: event AssetsUnlocked(uint256 indexed unlockSequenceNumber, bytes indexed recipient, address indexed token, address sender, uint256 amount, uint8 chain)
func (_AssetsBridge *AssetsBridgeFilterer) ParseAssetsUnlocked(log types.Log) (*AssetsBridgeAssetsUnlocked, error) {
	event := new(AssetsBridgeAssetsUnlocked)
	if err := _AssetsBridge.contract.UnpackLog(event, "AssetsUnlocked", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// AssetsBridgeERC20TokenMappingCreatedIterator is returned from FilterERC20TokenMappingCreated and is used to iterate over the raw logs and unpacked data for ERC20TokenMappingCreated events raised by the AssetsBridge contract.
type AssetsBridgeERC20TokenMappingCreatedIterator struct {
	Event *AssetsBridgeERC20TokenMappingCreated // Event containing the contract specifics and raw log

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
func (it *AssetsBridgeERC20TokenMappingCreatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AssetsBridgeERC20TokenMappingCreated)
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
		it.Event = new(AssetsBridgeERC20TokenMappingCreated)
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
func (it *AssetsBridgeERC20TokenMappingCreatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AssetsBridgeERC20TokenMappingCreatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AssetsBridgeERC20TokenMappingCreated represents a ERC20TokenMappingCreated event raised by the AssetsBridge contract.
type AssetsBridgeERC20TokenMappingCreated struct {
	SourceToken common.Address
	MezoToken   common.Address
	Raw         types.Log // Blockchain specific contextual infos
}

// FilterERC20TokenMappingCreated is a free log retrieval operation binding the contract event 0x8fc8785d532ca970395cfbe93300ce1a8cea58d5c3bdf0bc8641db1b70a8e94a.
//
// Solidity: event ERC20TokenMappingCreated(address indexed sourceToken, address indexed mezoToken)
func (_AssetsBridge *AssetsBridgeFilterer) FilterERC20TokenMappingCreated(opts *bind.FilterOpts, sourceToken []common.Address, mezoToken []common.Address) (*AssetsBridgeERC20TokenMappingCreatedIterator, error) {

	var sourceTokenRule []interface{}
	for _, sourceTokenItem := range sourceToken {
		sourceTokenRule = append(sourceTokenRule, sourceTokenItem)
	}
	var mezoTokenRule []interface{}
	for _, mezoTokenItem := range mezoToken {
		mezoTokenRule = append(mezoTokenRule, mezoTokenItem)
	}

	logs, sub, err := _AssetsBridge.contract.FilterLogs(opts, "ERC20TokenMappingCreated", sourceTokenRule, mezoTokenRule)
	if err != nil {
		return nil, err
	}
	return &AssetsBridgeERC20TokenMappingCreatedIterator{contract: _AssetsBridge.contract, event: "ERC20TokenMappingCreated", logs: logs, sub: sub}, nil
}

// WatchERC20TokenMappingCreated is a free log subscription operation binding the contract event 0x8fc8785d532ca970395cfbe93300ce1a8cea58d5c3bdf0bc8641db1b70a8e94a.
//
// Solidity: event ERC20TokenMappingCreated(address indexed sourceToken, address indexed mezoToken)
func (_AssetsBridge *AssetsBridgeFilterer) WatchERC20TokenMappingCreated(opts *bind.WatchOpts, sink chan<- *AssetsBridgeERC20TokenMappingCreated, sourceToken []common.Address, mezoToken []common.Address) (event.Subscription, error) {

	var sourceTokenRule []interface{}
	for _, sourceTokenItem := range sourceToken {
		sourceTokenRule = append(sourceTokenRule, sourceTokenItem)
	}
	var mezoTokenRule []interface{}
	for _, mezoTokenItem := range mezoToken {
		mezoTokenRule = append(mezoTokenRule, mezoTokenItem)
	}

	logs, sub, err := _AssetsBridge.contract.WatchLogs(opts, "ERC20TokenMappingCreated", sourceTokenRule, mezoTokenRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AssetsBridgeERC20TokenMappingCreated)
				if err := _AssetsBridge.contract.UnpackLog(event, "ERC20TokenMappingCreated", log); err != nil {
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

// ParseERC20TokenMappingCreated is a log parse operation binding the contract event 0x8fc8785d532ca970395cfbe93300ce1a8cea58d5c3bdf0bc8641db1b70a8e94a.
//
// Solidity: event ERC20TokenMappingCreated(address indexed sourceToken, address indexed mezoToken)
func (_AssetsBridge *AssetsBridgeFilterer) ParseERC20TokenMappingCreated(log types.Log) (*AssetsBridgeERC20TokenMappingCreated, error) {
	event := new(AssetsBridgeERC20TokenMappingCreated)
	if err := _AssetsBridge.contract.UnpackLog(event, "ERC20TokenMappingCreated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// AssetsBridgeERC20TokenMappingDeletedIterator is returned from FilterERC20TokenMappingDeleted and is used to iterate over the raw logs and unpacked data for ERC20TokenMappingDeleted events raised by the AssetsBridge contract.
type AssetsBridgeERC20TokenMappingDeletedIterator struct {
	Event *AssetsBridgeERC20TokenMappingDeleted // Event containing the contract specifics and raw log

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
func (it *AssetsBridgeERC20TokenMappingDeletedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AssetsBridgeERC20TokenMappingDeleted)
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
		it.Event = new(AssetsBridgeERC20TokenMappingDeleted)
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
func (it *AssetsBridgeERC20TokenMappingDeletedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AssetsBridgeERC20TokenMappingDeletedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AssetsBridgeERC20TokenMappingDeleted represents a ERC20TokenMappingDeleted event raised by the AssetsBridge contract.
type AssetsBridgeERC20TokenMappingDeleted struct {
	SourceToken common.Address
	MezoToken   common.Address
	Raw         types.Log // Blockchain specific contextual infos
}

// FilterERC20TokenMappingDeleted is a free log retrieval operation binding the contract event 0x6d123955084b0eb2c6c6eded567352c16761c53a8f16d4f8e63e0d6daf3469f4.
//
// Solidity: event ERC20TokenMappingDeleted(address indexed sourceToken, address indexed mezoToken)
func (_AssetsBridge *AssetsBridgeFilterer) FilterERC20TokenMappingDeleted(opts *bind.FilterOpts, sourceToken []common.Address, mezoToken []common.Address) (*AssetsBridgeERC20TokenMappingDeletedIterator, error) {

	var sourceTokenRule []interface{}
	for _, sourceTokenItem := range sourceToken {
		sourceTokenRule = append(sourceTokenRule, sourceTokenItem)
	}
	var mezoTokenRule []interface{}
	for _, mezoTokenItem := range mezoToken {
		mezoTokenRule = append(mezoTokenRule, mezoTokenItem)
	}

	logs, sub, err := _AssetsBridge.contract.FilterLogs(opts, "ERC20TokenMappingDeleted", sourceTokenRule, mezoTokenRule)
	if err != nil {
		return nil, err
	}
	return &AssetsBridgeERC20TokenMappingDeletedIterator{contract: _AssetsBridge.contract, event: "ERC20TokenMappingDeleted", logs: logs, sub: sub}, nil
}

// WatchERC20TokenMappingDeleted is a free log subscription operation binding the contract event 0x6d123955084b0eb2c6c6eded567352c16761c53a8f16d4f8e63e0d6daf3469f4.
//
// Solidity: event ERC20TokenMappingDeleted(address indexed sourceToken, address indexed mezoToken)
func (_AssetsBridge *AssetsBridgeFilterer) WatchERC20TokenMappingDeleted(opts *bind.WatchOpts, sink chan<- *AssetsBridgeERC20TokenMappingDeleted, sourceToken []common.Address, mezoToken []common.Address) (event.Subscription, error) {

	var sourceTokenRule []interface{}
	for _, sourceTokenItem := range sourceToken {
		sourceTokenRule = append(sourceTokenRule, sourceTokenItem)
	}
	var mezoTokenRule []interface{}
	for _, mezoTokenItem := range mezoToken {
		mezoTokenRule = append(mezoTokenRule, mezoTokenItem)
	}

	logs, sub, err := _AssetsBridge.contract.WatchLogs(opts, "ERC20TokenMappingDeleted", sourceTokenRule, mezoTokenRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AssetsBridgeERC20TokenMappingDeleted)
				if err := _AssetsBridge.contract.UnpackLog(event, "ERC20TokenMappingDeleted", log); err != nil {
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

// ParseERC20TokenMappingDeleted is a log parse operation binding the contract event 0x6d123955084b0eb2c6c6eded567352c16761c53a8f16d4f8e63e0d6daf3469f4.
//
// Solidity: event ERC20TokenMappingDeleted(address indexed sourceToken, address indexed mezoToken)
func (_AssetsBridge *AssetsBridgeFilterer) ParseERC20TokenMappingDeleted(log types.Log) (*AssetsBridgeERC20TokenMappingDeleted, error) {
	event := new(AssetsBridgeERC20TokenMappingDeleted)
	if err := _AssetsBridge.contract.UnpackLog(event, "ERC20TokenMappingDeleted", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// AssetsBridgeMinBridgeOutAmountSetIterator is returned from FilterMinBridgeOutAmountSet and is used to iterate over the raw logs and unpacked data for MinBridgeOutAmountSet events raised by the AssetsBridge contract.
type AssetsBridgeMinBridgeOutAmountSetIterator struct {
	Event *AssetsBridgeMinBridgeOutAmountSet // Event containing the contract specifics and raw log

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
func (it *AssetsBridgeMinBridgeOutAmountSetIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AssetsBridgeMinBridgeOutAmountSet)
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
		it.Event = new(AssetsBridgeMinBridgeOutAmountSet)
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
func (it *AssetsBridgeMinBridgeOutAmountSetIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AssetsBridgeMinBridgeOutAmountSetIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AssetsBridgeMinBridgeOutAmountSet represents a MinBridgeOutAmountSet event raised by the AssetsBridge contract.
type AssetsBridgeMinBridgeOutAmountSet struct {
	MezoToken common.Address
	MinAmount *big.Int
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterMinBridgeOutAmountSet is a free log retrieval operation binding the contract event 0xefe7503330d9d3ecc28191a2acccb3c450562c8f261f1ad466d2215f4a4b35e3.
//
// Solidity: event MinBridgeOutAmountSet(address indexed mezoToken, uint256 minAmount)
func (_AssetsBridge *AssetsBridgeFilterer) FilterMinBridgeOutAmountSet(opts *bind.FilterOpts, mezoToken []common.Address) (*AssetsBridgeMinBridgeOutAmountSetIterator, error) {

	var mezoTokenRule []interface{}
	for _, mezoTokenItem := range mezoToken {
		mezoTokenRule = append(mezoTokenRule, mezoTokenItem)
	}

	logs, sub, err := _AssetsBridge.contract.FilterLogs(opts, "MinBridgeOutAmountSet", mezoTokenRule)
	if err != nil {
		return nil, err
	}
	return &AssetsBridgeMinBridgeOutAmountSetIterator{contract: _AssetsBridge.contract, event: "MinBridgeOutAmountSet", logs: logs, sub: sub}, nil
}

// WatchMinBridgeOutAmountSet is a free log subscription operation binding the contract event 0xefe7503330d9d3ecc28191a2acccb3c450562c8f261f1ad466d2215f4a4b35e3.
//
// Solidity: event MinBridgeOutAmountSet(address indexed mezoToken, uint256 minAmount)
func (_AssetsBridge *AssetsBridgeFilterer) WatchMinBridgeOutAmountSet(opts *bind.WatchOpts, sink chan<- *AssetsBridgeMinBridgeOutAmountSet, mezoToken []common.Address) (event.Subscription, error) {

	var mezoTokenRule []interface{}
	for _, mezoTokenItem := range mezoToken {
		mezoTokenRule = append(mezoTokenRule, mezoTokenItem)
	}

	logs, sub, err := _AssetsBridge.contract.WatchLogs(opts, "MinBridgeOutAmountSet", mezoTokenRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AssetsBridgeMinBridgeOutAmountSet)
				if err := _AssetsBridge.contract.UnpackLog(event, "MinBridgeOutAmountSet", log); err != nil {
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

// ParseMinBridgeOutAmountSet is a log parse operation binding the contract event 0xefe7503330d9d3ecc28191a2acccb3c450562c8f261f1ad466d2215f4a4b35e3.
//
// Solidity: event MinBridgeOutAmountSet(address indexed mezoToken, uint256 minAmount)
func (_AssetsBridge *AssetsBridgeFilterer) ParseMinBridgeOutAmountSet(log types.Log) (*AssetsBridgeMinBridgeOutAmountSet, error) {
	event := new(AssetsBridgeMinBridgeOutAmountSet)
	if err := _AssetsBridge.contract.UnpackLog(event, "MinBridgeOutAmountSet", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
