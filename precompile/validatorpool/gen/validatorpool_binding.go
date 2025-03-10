// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package validatorpool

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

// Description is an auto generated low-level Go binding around an user-defined struct.
type Description struct {
	Moniker         string
	Identity        string
	Website         string
	SecurityContact string
	Details         string
}

// Privilege is an auto generated low-level Go binding around an user-defined struct.
type Privilege struct {
	Id   uint8
	Name string
}

// ValidatorpoolMetaData contains all meta data concerning the Validatorpool contract.
var ValidatorpoolMetaData = &bind.MetaData{
	ABI: "[{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"operator\",\"type\":\"address\"}],\"name\":\"ApplicationApproved\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[],\"name\":\"ApplicationsCleaned\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"operator\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"consPubKey\",\"type\":\"bytes32\"},{\"components\":[{\"internalType\":\"string\",\"name\":\"moniker\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"identity\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"website\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"securityContact\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"details\",\"type\":\"string\"}],\"internalType\":\"structDescription\",\"name\":\"description\",\"type\":\"tuple\"}],\"name\":\"ApplicationSubmitted\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"previousOwner\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"OwnershipTransferStarted\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"previousOwner\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"OwnershipTransferred\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"operator\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"uint8\",\"name\":\"privilegeId\",\"type\":\"uint8\"}],\"name\":\"PrivilegeAdded\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"operator\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"uint8\",\"name\":\"privilegeId\",\"type\":\"uint8\"}],\"name\":\"PrivilegeRemoved\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"operator\",\"type\":\"address\"}],\"name\":\"ValidatorJoined\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"operator\",\"type\":\"address\"}],\"name\":\"ValidatorKicked\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"operator\",\"type\":\"address\"}],\"name\":\"ValidatorLeft\",\"type\":\"event\"},{\"inputs\":[],\"name\":\"acceptOwnership\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address[]\",\"name\":\"operators\",\"type\":\"address[]\"},{\"internalType\":\"uint8\",\"name\":\"privilegeId\",\"type\":\"uint8\"}],\"name\":\"addPrivilege\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"operator\",\"type\":\"address\"}],\"name\":\"application\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"consPubKey\",\"type\":\"bytes32\"},{\"components\":[{\"internalType\":\"string\",\"name\":\"moniker\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"identity\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"website\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"securityContact\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"details\",\"type\":\"string\"}],\"internalType\":\"structDescription\",\"name\":\"description\",\"type\":\"tuple\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"applications\",\"outputs\":[{\"internalType\":\"address[]\",\"name\":\"\",\"type\":\"address[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"operator\",\"type\":\"address\"}],\"name\":\"approveApplication\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"candidateOwner\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"cleanupApplications\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"operator\",\"type\":\"address\"}],\"name\":\"kick\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"leave\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"owner\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"privileges\",\"outputs\":[{\"components\":[{\"internalType\":\"uint8\",\"name\":\"id\",\"type\":\"uint8\"},{\"internalType\":\"string\",\"name\":\"name\",\"type\":\"string\"}],\"internalType\":\"structPrivilege[]\",\"name\":\"\",\"type\":\"tuple[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address[]\",\"name\":\"operators\",\"type\":\"address[]\"},{\"internalType\":\"uint8\",\"name\":\"privilegeId\",\"type\":\"uint8\"}],\"name\":\"removePrivilege\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"consPubKey\",\"type\":\"bytes32\"},{\"components\":[{\"internalType\":\"string\",\"name\":\"moniker\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"identity\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"website\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"securityContact\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"details\",\"type\":\"string\"}],\"internalType\":\"structDescription\",\"name\":\"description\",\"type\":\"tuple\"}],\"name\":\"submitApplication\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"transferOwnership\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"operator\",\"type\":\"address\"}],\"name\":\"validator\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"consPubKey\",\"type\":\"bytes32\"},{\"components\":[{\"internalType\":\"string\",\"name\":\"moniker\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"identity\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"website\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"securityContact\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"details\",\"type\":\"string\"}],\"internalType\":\"structDescription\",\"name\":\"description\",\"type\":\"tuple\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"validators\",\"outputs\":[{\"internalType\":\"address[]\",\"name\":\"\",\"type\":\"address[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint8\",\"name\":\"privilegeId\",\"type\":\"uint8\"}],\"name\":\"validatorsByPrivilege\",\"outputs\":[{\"internalType\":\"address[]\",\"name\":\"operators\",\"type\":\"address[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"}]",
}

// ValidatorpoolABI is the input ABI used to generate the binding from.
// Deprecated: Use ValidatorpoolMetaData.ABI instead.
var ValidatorpoolABI = ValidatorpoolMetaData.ABI

// Validatorpool is an auto generated Go binding around an Ethereum contract.
type Validatorpool struct {
	ValidatorpoolCaller     // Read-only binding to the contract
	ValidatorpoolTransactor // Write-only binding to the contract
	ValidatorpoolFilterer   // Log filterer for contract events
}

// ValidatorpoolCaller is an auto generated read-only Go binding around an Ethereum contract.
type ValidatorpoolCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ValidatorpoolTransactor is an auto generated write-only Go binding around an Ethereum contract.
type ValidatorpoolTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ValidatorpoolFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type ValidatorpoolFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ValidatorpoolSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type ValidatorpoolSession struct {
	Contract     *Validatorpool    // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// ValidatorpoolCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type ValidatorpoolCallerSession struct {
	Contract *ValidatorpoolCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts        // Call options to use throughout this session
}

// ValidatorpoolTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type ValidatorpoolTransactorSession struct {
	Contract     *ValidatorpoolTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts        // Transaction auth options to use throughout this session
}

// ValidatorpoolRaw is an auto generated low-level Go binding around an Ethereum contract.
type ValidatorpoolRaw struct {
	Contract *Validatorpool // Generic contract binding to access the raw methods on
}

// ValidatorpoolCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type ValidatorpoolCallerRaw struct {
	Contract *ValidatorpoolCaller // Generic read-only contract binding to access the raw methods on
}

// ValidatorpoolTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type ValidatorpoolTransactorRaw struct {
	Contract *ValidatorpoolTransactor // Generic write-only contract binding to access the raw methods on
}

// NewValidatorpool creates a new instance of Validatorpool, bound to a specific deployed contract.
func NewValidatorpool(address common.Address, backend bind.ContractBackend) (*Validatorpool, error) {
	contract, err := bindValidatorpool(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &Validatorpool{ValidatorpoolCaller: ValidatorpoolCaller{contract: contract}, ValidatorpoolTransactor: ValidatorpoolTransactor{contract: contract}, ValidatorpoolFilterer: ValidatorpoolFilterer{contract: contract}}, nil
}

// NewValidatorpoolCaller creates a new read-only instance of Validatorpool, bound to a specific deployed contract.
func NewValidatorpoolCaller(address common.Address, caller bind.ContractCaller) (*ValidatorpoolCaller, error) {
	contract, err := bindValidatorpool(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &ValidatorpoolCaller{contract: contract}, nil
}

// NewValidatorpoolTransactor creates a new write-only instance of Validatorpool, bound to a specific deployed contract.
func NewValidatorpoolTransactor(address common.Address, transactor bind.ContractTransactor) (*ValidatorpoolTransactor, error) {
	contract, err := bindValidatorpool(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &ValidatorpoolTransactor{contract: contract}, nil
}

// NewValidatorpoolFilterer creates a new log filterer instance of Validatorpool, bound to a specific deployed contract.
func NewValidatorpoolFilterer(address common.Address, filterer bind.ContractFilterer) (*ValidatorpoolFilterer, error) {
	contract, err := bindValidatorpool(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &ValidatorpoolFilterer{contract: contract}, nil
}

// bindValidatorpool binds a generic wrapper to an already deployed contract.
func bindValidatorpool(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(ValidatorpoolABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Validatorpool *ValidatorpoolRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Validatorpool.Contract.ValidatorpoolCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Validatorpool *ValidatorpoolRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Validatorpool.Contract.ValidatorpoolTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Validatorpool *ValidatorpoolRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Validatorpool.Contract.ValidatorpoolTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Validatorpool *ValidatorpoolCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Validatorpool.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Validatorpool *ValidatorpoolTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Validatorpool.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Validatorpool *ValidatorpoolTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Validatorpool.Contract.contract.Transact(opts, method, params...)
}

// Application is a free data retrieval call binding the contract method 0x8bd2cc11.
//
// Solidity: function application(address operator) view returns(bytes32 consPubKey, (string,string,string,string,string) description)
func (_Validatorpool *ValidatorpoolCaller) Application(opts *bind.CallOpts, operator common.Address) (struct {
	ConsPubKey  [32]byte
	Description Description
}, error,
) {
	var out []interface{}
	err := _Validatorpool.contract.Call(opts, &out, "application", operator)

	outstruct := new(struct {
		ConsPubKey  [32]byte
		Description Description
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.ConsPubKey = *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)
	outstruct.Description = *abi.ConvertType(out[1], new(Description)).(*Description)

	return *outstruct, err
}

// Application is a free data retrieval call binding the contract method 0x8bd2cc11.
//
// Solidity: function application(address operator) view returns(bytes32 consPubKey, (string,string,string,string,string) description)
func (_Validatorpool *ValidatorpoolSession) Application(operator common.Address) (struct {
	ConsPubKey  [32]byte
	Description Description
}, error,
) {
	return _Validatorpool.Contract.Application(&_Validatorpool.CallOpts, operator)
}

// Application is a free data retrieval call binding the contract method 0x8bd2cc11.
//
// Solidity: function application(address operator) view returns(bytes32 consPubKey, (string,string,string,string,string) description)
func (_Validatorpool *ValidatorpoolCallerSession) Application(operator common.Address) (struct {
	ConsPubKey  [32]byte
	Description Description
}, error,
) {
	return _Validatorpool.Contract.Application(&_Validatorpool.CallOpts, operator)
}

// Applications is a free data retrieval call binding the contract method 0x7ce5e82f.
//
// Solidity: function applications() view returns(address[])
func (_Validatorpool *ValidatorpoolCaller) Applications(opts *bind.CallOpts) ([]common.Address, error) {
	var out []interface{}
	err := _Validatorpool.contract.Call(opts, &out, "applications")
	if err != nil {
		return *new([]common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new([]common.Address)).(*[]common.Address)

	return out0, err
}

// Applications is a free data retrieval call binding the contract method 0x7ce5e82f.
//
// Solidity: function applications() view returns(address[])
func (_Validatorpool *ValidatorpoolSession) Applications() ([]common.Address, error) {
	return _Validatorpool.Contract.Applications(&_Validatorpool.CallOpts)
}

// Applications is a free data retrieval call binding the contract method 0x7ce5e82f.
//
// Solidity: function applications() view returns(address[])
func (_Validatorpool *ValidatorpoolCallerSession) Applications() ([]common.Address, error) {
	return _Validatorpool.Contract.Applications(&_Validatorpool.CallOpts)
}

// CandidateOwner is a free data retrieval call binding the contract method 0xc105ea2b.
//
// Solidity: function candidateOwner() view returns(address)
func (_Validatorpool *ValidatorpoolCaller) CandidateOwner(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _Validatorpool.contract.Call(opts, &out, "candidateOwner")
	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err
}

// CandidateOwner is a free data retrieval call binding the contract method 0xc105ea2b.
//
// Solidity: function candidateOwner() view returns(address)
func (_Validatorpool *ValidatorpoolSession) CandidateOwner() (common.Address, error) {
	return _Validatorpool.Contract.CandidateOwner(&_Validatorpool.CallOpts)
}

// CandidateOwner is a free data retrieval call binding the contract method 0xc105ea2b.
//
// Solidity: function candidateOwner() view returns(address)
func (_Validatorpool *ValidatorpoolCallerSession) CandidateOwner() (common.Address, error) {
	return _Validatorpool.Contract.CandidateOwner(&_Validatorpool.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_Validatorpool *ValidatorpoolCaller) Owner(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _Validatorpool.contract.Call(opts, &out, "owner")
	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_Validatorpool *ValidatorpoolSession) Owner() (common.Address, error) {
	return _Validatorpool.Contract.Owner(&_Validatorpool.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_Validatorpool *ValidatorpoolCallerSession) Owner() (common.Address, error) {
	return _Validatorpool.Contract.Owner(&_Validatorpool.CallOpts)
}

// Privileges is a free data retrieval call binding the contract method 0x2dea97ba.
//
// Solidity: function privileges() view returns((uint8,string)[])
func (_Validatorpool *ValidatorpoolCaller) Privileges(opts *bind.CallOpts) ([]Privilege, error) {
	var out []interface{}
	err := _Validatorpool.contract.Call(opts, &out, "privileges")
	if err != nil {
		return *new([]Privilege), err
	}

	out0 := *abi.ConvertType(out[0], new([]Privilege)).(*[]Privilege)

	return out0, err
}

// Privileges is a free data retrieval call binding the contract method 0x2dea97ba.
//
// Solidity: function privileges() view returns((uint8,string)[])
func (_Validatorpool *ValidatorpoolSession) Privileges() ([]Privilege, error) {
	return _Validatorpool.Contract.Privileges(&_Validatorpool.CallOpts)
}

// Privileges is a free data retrieval call binding the contract method 0x2dea97ba.
//
// Solidity: function privileges() view returns((uint8,string)[])
func (_Validatorpool *ValidatorpoolCallerSession) Privileges() ([]Privilege, error) {
	return _Validatorpool.Contract.Privileges(&_Validatorpool.CallOpts)
}

// Validator is a free data retrieval call binding the contract method 0x223b3b7a.
//
// Solidity: function validator(address operator) view returns(bytes32 consPubKey, (string,string,string,string,string) description)
func (_Validatorpool *ValidatorpoolCaller) Validator(opts *bind.CallOpts, operator common.Address) (struct {
	ConsPubKey  [32]byte
	Description Description
}, error,
) {
	var out []interface{}
	err := _Validatorpool.contract.Call(opts, &out, "validator", operator)

	outstruct := new(struct {
		ConsPubKey  [32]byte
		Description Description
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.ConsPubKey = *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)
	outstruct.Description = *abi.ConvertType(out[1], new(Description)).(*Description)

	return *outstruct, err
}

// Validator is a free data retrieval call binding the contract method 0x223b3b7a.
//
// Solidity: function validator(address operator) view returns(bytes32 consPubKey, (string,string,string,string,string) description)
func (_Validatorpool *ValidatorpoolSession) Validator(operator common.Address) (struct {
	ConsPubKey  [32]byte
	Description Description
}, error,
) {
	return _Validatorpool.Contract.Validator(&_Validatorpool.CallOpts, operator)
}

// Validator is a free data retrieval call binding the contract method 0x223b3b7a.
//
// Solidity: function validator(address operator) view returns(bytes32 consPubKey, (string,string,string,string,string) description)
func (_Validatorpool *ValidatorpoolCallerSession) Validator(operator common.Address) (struct {
	ConsPubKey  [32]byte
	Description Description
}, error,
) {
	return _Validatorpool.Contract.Validator(&_Validatorpool.CallOpts, operator)
}

// Validators is a free data retrieval call binding the contract method 0xca1e7819.
//
// Solidity: function validators() view returns(address[])
func (_Validatorpool *ValidatorpoolCaller) Validators(opts *bind.CallOpts) ([]common.Address, error) {
	var out []interface{}
	err := _Validatorpool.contract.Call(opts, &out, "validators")
	if err != nil {
		return *new([]common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new([]common.Address)).(*[]common.Address)

	return out0, err
}

// Validators is a free data retrieval call binding the contract method 0xca1e7819.
//
// Solidity: function validators() view returns(address[])
func (_Validatorpool *ValidatorpoolSession) Validators() ([]common.Address, error) {
	return _Validatorpool.Contract.Validators(&_Validatorpool.CallOpts)
}

// Validators is a free data retrieval call binding the contract method 0xca1e7819.
//
// Solidity: function validators() view returns(address[])
func (_Validatorpool *ValidatorpoolCallerSession) Validators() ([]common.Address, error) {
	return _Validatorpool.Contract.Validators(&_Validatorpool.CallOpts)
}

// ValidatorsByPrivilege is a free data retrieval call binding the contract method 0x617177f5.
//
// Solidity: function validatorsByPrivilege(uint8 privilegeId) view returns(address[] operators)
func (_Validatorpool *ValidatorpoolCaller) ValidatorsByPrivilege(opts *bind.CallOpts, privilegeId uint8) ([]common.Address, error) {
	var out []interface{}
	err := _Validatorpool.contract.Call(opts, &out, "validatorsByPrivilege", privilegeId)
	if err != nil {
		return *new([]common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new([]common.Address)).(*[]common.Address)

	return out0, err
}

// ValidatorsByPrivilege is a free data retrieval call binding the contract method 0x617177f5.
//
// Solidity: function validatorsByPrivilege(uint8 privilegeId) view returns(address[] operators)
func (_Validatorpool *ValidatorpoolSession) ValidatorsByPrivilege(privilegeId uint8) ([]common.Address, error) {
	return _Validatorpool.Contract.ValidatorsByPrivilege(&_Validatorpool.CallOpts, privilegeId)
}

// ValidatorsByPrivilege is a free data retrieval call binding the contract method 0x617177f5.
//
// Solidity: function validatorsByPrivilege(uint8 privilegeId) view returns(address[] operators)
func (_Validatorpool *ValidatorpoolCallerSession) ValidatorsByPrivilege(privilegeId uint8) ([]common.Address, error) {
	return _Validatorpool.Contract.ValidatorsByPrivilege(&_Validatorpool.CallOpts, privilegeId)
}

// AcceptOwnership is a paid mutator transaction binding the contract method 0x79ba5097.
//
// Solidity: function acceptOwnership() returns(bool)
func (_Validatorpool *ValidatorpoolTransactor) AcceptOwnership(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Validatorpool.contract.Transact(opts, "acceptOwnership")
}

// AcceptOwnership is a paid mutator transaction binding the contract method 0x79ba5097.
//
// Solidity: function acceptOwnership() returns(bool)
func (_Validatorpool *ValidatorpoolSession) AcceptOwnership() (*types.Transaction, error) {
	return _Validatorpool.Contract.AcceptOwnership(&_Validatorpool.TransactOpts)
}

// AcceptOwnership is a paid mutator transaction binding the contract method 0x79ba5097.
//
// Solidity: function acceptOwnership() returns(bool)
func (_Validatorpool *ValidatorpoolTransactorSession) AcceptOwnership() (*types.Transaction, error) {
	return _Validatorpool.Contract.AcceptOwnership(&_Validatorpool.TransactOpts)
}

// AddPrivilege is a paid mutator transaction binding the contract method 0x644bb08b.
//
// Solidity: function addPrivilege(address[] operators, uint8 privilegeId) returns(bool)
func (_Validatorpool *ValidatorpoolTransactor) AddPrivilege(opts *bind.TransactOpts, operators []common.Address, privilegeId uint8) (*types.Transaction, error) {
	return _Validatorpool.contract.Transact(opts, "addPrivilege", operators, privilegeId)
}

// AddPrivilege is a paid mutator transaction binding the contract method 0x644bb08b.
//
// Solidity: function addPrivilege(address[] operators, uint8 privilegeId) returns(bool)
func (_Validatorpool *ValidatorpoolSession) AddPrivilege(operators []common.Address, privilegeId uint8) (*types.Transaction, error) {
	return _Validatorpool.Contract.AddPrivilege(&_Validatorpool.TransactOpts, operators, privilegeId)
}

// AddPrivilege is a paid mutator transaction binding the contract method 0x644bb08b.
//
// Solidity: function addPrivilege(address[] operators, uint8 privilegeId) returns(bool)
func (_Validatorpool *ValidatorpoolTransactorSession) AddPrivilege(operators []common.Address, privilegeId uint8) (*types.Transaction, error) {
	return _Validatorpool.Contract.AddPrivilege(&_Validatorpool.TransactOpts, operators, privilegeId)
}

// ApproveApplication is a paid mutator transaction binding the contract method 0xe3ae4d0a.
//
// Solidity: function approveApplication(address operator) returns(bool)
func (_Validatorpool *ValidatorpoolTransactor) ApproveApplication(opts *bind.TransactOpts, operator common.Address) (*types.Transaction, error) {
	return _Validatorpool.contract.Transact(opts, "approveApplication", operator)
}

// ApproveApplication is a paid mutator transaction binding the contract method 0xe3ae4d0a.
//
// Solidity: function approveApplication(address operator) returns(bool)
func (_Validatorpool *ValidatorpoolSession) ApproveApplication(operator common.Address) (*types.Transaction, error) {
	return _Validatorpool.Contract.ApproveApplication(&_Validatorpool.TransactOpts, operator)
}

// ApproveApplication is a paid mutator transaction binding the contract method 0xe3ae4d0a.
//
// Solidity: function approveApplication(address operator) returns(bool)
func (_Validatorpool *ValidatorpoolTransactorSession) ApproveApplication(operator common.Address) (*types.Transaction, error) {
	return _Validatorpool.Contract.ApproveApplication(&_Validatorpool.TransactOpts, operator)
}

// CleanupApplications is a paid mutator transaction binding the contract method 0x54023a4e.
//
// Solidity: function cleanupApplications() returns(bool)
func (_Validatorpool *ValidatorpoolTransactor) CleanupApplications(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Validatorpool.contract.Transact(opts, "cleanupApplications")
}

// CleanupApplications is a paid mutator transaction binding the contract method 0x54023a4e.
//
// Solidity: function cleanupApplications() returns(bool)
func (_Validatorpool *ValidatorpoolSession) CleanupApplications() (*types.Transaction, error) {
	return _Validatorpool.Contract.CleanupApplications(&_Validatorpool.TransactOpts)
}

// CleanupApplications is a paid mutator transaction binding the contract method 0x54023a4e.
//
// Solidity: function cleanupApplications() returns(bool)
func (_Validatorpool *ValidatorpoolTransactorSession) CleanupApplications() (*types.Transaction, error) {
	return _Validatorpool.Contract.CleanupApplications(&_Validatorpool.TransactOpts)
}

// Kick is a paid mutator transaction binding the contract method 0x96c55175.
//
// Solidity: function kick(address operator) returns(bool)
func (_Validatorpool *ValidatorpoolTransactor) Kick(opts *bind.TransactOpts, operator common.Address) (*types.Transaction, error) {
	return _Validatorpool.contract.Transact(opts, "kick", operator)
}

// Kick is a paid mutator transaction binding the contract method 0x96c55175.
//
// Solidity: function kick(address operator) returns(bool)
func (_Validatorpool *ValidatorpoolSession) Kick(operator common.Address) (*types.Transaction, error) {
	return _Validatorpool.Contract.Kick(&_Validatorpool.TransactOpts, operator)
}

// Kick is a paid mutator transaction binding the contract method 0x96c55175.
//
// Solidity: function kick(address operator) returns(bool)
func (_Validatorpool *ValidatorpoolTransactorSession) Kick(operator common.Address) (*types.Transaction, error) {
	return _Validatorpool.Contract.Kick(&_Validatorpool.TransactOpts, operator)
}

// Leave is a paid mutator transaction binding the contract method 0xd66d9e19.
//
// Solidity: function leave() returns(bool)
func (_Validatorpool *ValidatorpoolTransactor) Leave(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Validatorpool.contract.Transact(opts, "leave")
}

// Leave is a paid mutator transaction binding the contract method 0xd66d9e19.
//
// Solidity: function leave() returns(bool)
func (_Validatorpool *ValidatorpoolSession) Leave() (*types.Transaction, error) {
	return _Validatorpool.Contract.Leave(&_Validatorpool.TransactOpts)
}

// Leave is a paid mutator transaction binding the contract method 0xd66d9e19.
//
// Solidity: function leave() returns(bool)
func (_Validatorpool *ValidatorpoolTransactorSession) Leave() (*types.Transaction, error) {
	return _Validatorpool.Contract.Leave(&_Validatorpool.TransactOpts)
}

// RemovePrivilege is a paid mutator transaction binding the contract method 0x33576275.
//
// Solidity: function removePrivilege(address[] operators, uint8 privilegeId) returns(bool)
func (_Validatorpool *ValidatorpoolTransactor) RemovePrivilege(opts *bind.TransactOpts, operators []common.Address, privilegeId uint8) (*types.Transaction, error) {
	return _Validatorpool.contract.Transact(opts, "removePrivilege", operators, privilegeId)
}

// RemovePrivilege is a paid mutator transaction binding the contract method 0x33576275.
//
// Solidity: function removePrivilege(address[] operators, uint8 privilegeId) returns(bool)
func (_Validatorpool *ValidatorpoolSession) RemovePrivilege(operators []common.Address, privilegeId uint8) (*types.Transaction, error) {
	return _Validatorpool.Contract.RemovePrivilege(&_Validatorpool.TransactOpts, operators, privilegeId)
}

// RemovePrivilege is a paid mutator transaction binding the contract method 0x33576275.
//
// Solidity: function removePrivilege(address[] operators, uint8 privilegeId) returns(bool)
func (_Validatorpool *ValidatorpoolTransactorSession) RemovePrivilege(operators []common.Address, privilegeId uint8) (*types.Transaction, error) {
	return _Validatorpool.Contract.RemovePrivilege(&_Validatorpool.TransactOpts, operators, privilegeId)
}

// SubmitApplication is a paid mutator transaction binding the contract method 0x6a050f5f.
//
// Solidity: function submitApplication(bytes32 consPubKey, (string,string,string,string,string) description) returns(bool)
func (_Validatorpool *ValidatorpoolTransactor) SubmitApplication(opts *bind.TransactOpts, consPubKey [32]byte, description Description) (*types.Transaction, error) {
	return _Validatorpool.contract.Transact(opts, "submitApplication", consPubKey, description)
}

// SubmitApplication is a paid mutator transaction binding the contract method 0x6a050f5f.
//
// Solidity: function submitApplication(bytes32 consPubKey, (string,string,string,string,string) description) returns(bool)
func (_Validatorpool *ValidatorpoolSession) SubmitApplication(consPubKey [32]byte, description Description) (*types.Transaction, error) {
	return _Validatorpool.Contract.SubmitApplication(&_Validatorpool.TransactOpts, consPubKey, description)
}

// SubmitApplication is a paid mutator transaction binding the contract method 0x6a050f5f.
//
// Solidity: function submitApplication(bytes32 consPubKey, (string,string,string,string,string) description) returns(bool)
func (_Validatorpool *ValidatorpoolTransactorSession) SubmitApplication(consPubKey [32]byte, description Description) (*types.Transaction, error) {
	return _Validatorpool.Contract.SubmitApplication(&_Validatorpool.TransactOpts, consPubKey, description)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns(bool)
func (_Validatorpool *ValidatorpoolTransactor) TransferOwnership(opts *bind.TransactOpts, newOwner common.Address) (*types.Transaction, error) {
	return _Validatorpool.contract.Transact(opts, "transferOwnership", newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns(bool)
func (_Validatorpool *ValidatorpoolSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _Validatorpool.Contract.TransferOwnership(&_Validatorpool.TransactOpts, newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns(bool)
func (_Validatorpool *ValidatorpoolTransactorSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _Validatorpool.Contract.TransferOwnership(&_Validatorpool.TransactOpts, newOwner)
}

// ValidatorpoolApplicationApprovedIterator is returned from FilterApplicationApproved and is used to iterate over the raw logs and unpacked data for ApplicationApproved events raised by the Validatorpool contract.
type ValidatorpoolApplicationApprovedIterator struct {
	Event *ValidatorpoolApplicationApproved // Event containing the contract specifics and raw log

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
func (it *ValidatorpoolApplicationApprovedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ValidatorpoolApplicationApproved)
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
		it.Event = new(ValidatorpoolApplicationApproved)
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
func (it *ValidatorpoolApplicationApprovedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ValidatorpoolApplicationApprovedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ValidatorpoolApplicationApproved represents a ApplicationApproved event raised by the Validatorpool contract.
type ValidatorpoolApplicationApproved struct {
	Operator common.Address
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterApplicationApproved is a free log retrieval operation binding the contract event 0x6ca6150407f26e90367ff690c8b617cad626020aa12080384e3b31479c0442fb.
//
// Solidity: event ApplicationApproved(address indexed operator)
func (_Validatorpool *ValidatorpoolFilterer) FilterApplicationApproved(opts *bind.FilterOpts, operator []common.Address) (*ValidatorpoolApplicationApprovedIterator, error) {
	var operatorRule []interface{}
	for _, operatorItem := range operator {
		operatorRule = append(operatorRule, operatorItem)
	}

	logs, sub, err := _Validatorpool.contract.FilterLogs(opts, "ApplicationApproved", operatorRule)
	if err != nil {
		return nil, err
	}
	return &ValidatorpoolApplicationApprovedIterator{contract: _Validatorpool.contract, event: "ApplicationApproved", logs: logs, sub: sub}, nil
}

// WatchApplicationApproved is a free log subscription operation binding the contract event 0x6ca6150407f26e90367ff690c8b617cad626020aa12080384e3b31479c0442fb.
//
// Solidity: event ApplicationApproved(address indexed operator)
func (_Validatorpool *ValidatorpoolFilterer) WatchApplicationApproved(opts *bind.WatchOpts, sink chan<- *ValidatorpoolApplicationApproved, operator []common.Address) (event.Subscription, error) {
	var operatorRule []interface{}
	for _, operatorItem := range operator {
		operatorRule = append(operatorRule, operatorItem)
	}

	logs, sub, err := _Validatorpool.contract.WatchLogs(opts, "ApplicationApproved", operatorRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ValidatorpoolApplicationApproved)
				if err := _Validatorpool.contract.UnpackLog(event, "ApplicationApproved", log); err != nil {
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

// ParseApplicationApproved is a log parse operation binding the contract event 0x6ca6150407f26e90367ff690c8b617cad626020aa12080384e3b31479c0442fb.
//
// Solidity: event ApplicationApproved(address indexed operator)
func (_Validatorpool *ValidatorpoolFilterer) ParseApplicationApproved(log types.Log) (*ValidatorpoolApplicationApproved, error) {
	event := new(ValidatorpoolApplicationApproved)
	if err := _Validatorpool.contract.UnpackLog(event, "ApplicationApproved", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ValidatorpoolApplicationSubmittedIterator is returned from FilterApplicationSubmitted and is used to iterate over the raw logs and unpacked data for ApplicationSubmitted events raised by the Validatorpool contract.
type ValidatorpoolApplicationSubmittedIterator struct {
	Event *ValidatorpoolApplicationSubmitted // Event containing the contract specifics and raw log

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
func (it *ValidatorpoolApplicationSubmittedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ValidatorpoolApplicationSubmitted)
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
		it.Event = new(ValidatorpoolApplicationSubmitted)
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
func (it *ValidatorpoolApplicationSubmittedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ValidatorpoolApplicationSubmittedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ValidatorpoolApplicationSubmitted represents a ApplicationSubmitted event raised by the Validatorpool contract.
type ValidatorpoolApplicationSubmitted struct {
	Operator    common.Address
	ConsPubKey  [32]byte
	Description Description
	Raw         types.Log // Blockchain specific contextual infos
}

// FilterApplicationSubmitted is a free log retrieval operation binding the contract event 0x738bb6fe2828245f1bae61cfd83f725ef5bc836fea2b209921f8adbcfa42ec44.
//
// Solidity: event ApplicationSubmitted(address indexed operator, bytes32 indexed consPubKey, (string,string,string,string,string) description)
func (_Validatorpool *ValidatorpoolFilterer) FilterApplicationSubmitted(opts *bind.FilterOpts, operator []common.Address, consPubKey [][32]byte) (*ValidatorpoolApplicationSubmittedIterator, error) {
	var operatorRule []interface{}
	for _, operatorItem := range operator {
		operatorRule = append(operatorRule, operatorItem)
	}
	var consPubKeyRule []interface{}
	for _, consPubKeyItem := range consPubKey {
		consPubKeyRule = append(consPubKeyRule, consPubKeyItem)
	}

	logs, sub, err := _Validatorpool.contract.FilterLogs(opts, "ApplicationSubmitted", operatorRule, consPubKeyRule)
	if err != nil {
		return nil, err
	}
	return &ValidatorpoolApplicationSubmittedIterator{contract: _Validatorpool.contract, event: "ApplicationSubmitted", logs: logs, sub: sub}, nil
}

// WatchApplicationSubmitted is a free log subscription operation binding the contract event 0x738bb6fe2828245f1bae61cfd83f725ef5bc836fea2b209921f8adbcfa42ec44.
//
// Solidity: event ApplicationSubmitted(address indexed operator, bytes32 indexed consPubKey, (string,string,string,string,string) description)
func (_Validatorpool *ValidatorpoolFilterer) WatchApplicationSubmitted(opts *bind.WatchOpts, sink chan<- *ValidatorpoolApplicationSubmitted, operator []common.Address, consPubKey [][32]byte) (event.Subscription, error) {
	var operatorRule []interface{}
	for _, operatorItem := range operator {
		operatorRule = append(operatorRule, operatorItem)
	}
	var consPubKeyRule []interface{}
	for _, consPubKeyItem := range consPubKey {
		consPubKeyRule = append(consPubKeyRule, consPubKeyItem)
	}

	logs, sub, err := _Validatorpool.contract.WatchLogs(opts, "ApplicationSubmitted", operatorRule, consPubKeyRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ValidatorpoolApplicationSubmitted)
				if err := _Validatorpool.contract.UnpackLog(event, "ApplicationSubmitted", log); err != nil {
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

// ParseApplicationSubmitted is a log parse operation binding the contract event 0x738bb6fe2828245f1bae61cfd83f725ef5bc836fea2b209921f8adbcfa42ec44.
//
// Solidity: event ApplicationSubmitted(address indexed operator, bytes32 indexed consPubKey, (string,string,string,string,string) description)
func (_Validatorpool *ValidatorpoolFilterer) ParseApplicationSubmitted(log types.Log) (*ValidatorpoolApplicationSubmitted, error) {
	event := new(ValidatorpoolApplicationSubmitted)
	if err := _Validatorpool.contract.UnpackLog(event, "ApplicationSubmitted", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ValidatorpoolApplicationsCleanedIterator is returned from FilterApplicationsCleaned and is used to iterate over the raw logs and unpacked data for ApplicationsCleaned events raised by the Validatorpool contract.
type ValidatorpoolApplicationsCleanedIterator struct {
	Event *ValidatorpoolApplicationsCleaned // Event containing the contract specifics and raw log

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
func (it *ValidatorpoolApplicationsCleanedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ValidatorpoolApplicationsCleaned)
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
		it.Event = new(ValidatorpoolApplicationsCleaned)
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
func (it *ValidatorpoolApplicationsCleanedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ValidatorpoolApplicationsCleanedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ValidatorpoolApplicationsCleaned represents a ApplicationsCleaned event raised by the Validatorpool contract.
type ValidatorpoolApplicationsCleaned struct {
	Raw types.Log // Blockchain specific contextual infos
}

// FilterApplicationsCleaned is a free log retrieval operation binding the contract event 0xf89df90fd6304ff85707899b9493af5c13ef21ffded0e76ad9c900ca2a372a60.
//
// Solidity: event ApplicationsCleaned()
func (_Validatorpool *ValidatorpoolFilterer) FilterApplicationsCleaned(opts *bind.FilterOpts) (*ValidatorpoolApplicationsCleanedIterator, error) {
	logs, sub, err := _Validatorpool.contract.FilterLogs(opts, "ApplicationsCleaned")
	if err != nil {
		return nil, err
	}
	return &ValidatorpoolApplicationsCleanedIterator{contract: _Validatorpool.contract, event: "ApplicationsCleaned", logs: logs, sub: sub}, nil
}

// WatchApplicationsCleaned is a free log subscription operation binding the contract event 0xf89df90fd6304ff85707899b9493af5c13ef21ffded0e76ad9c900ca2a372a60.
//
// Solidity: event ApplicationsCleaned()
func (_Validatorpool *ValidatorpoolFilterer) WatchApplicationsCleaned(opts *bind.WatchOpts, sink chan<- *ValidatorpoolApplicationsCleaned) (event.Subscription, error) {
	logs, sub, err := _Validatorpool.contract.WatchLogs(opts, "ApplicationsCleaned")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ValidatorpoolApplicationsCleaned)
				if err := _Validatorpool.contract.UnpackLog(event, "ApplicationsCleaned", log); err != nil {
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

// ParseApplicationsCleaned is a log parse operation binding the contract event 0xf89df90fd6304ff85707899b9493af5c13ef21ffded0e76ad9c900ca2a372a60.
//
// Solidity: event ApplicationsCleaned()
func (_Validatorpool *ValidatorpoolFilterer) ParseApplicationsCleaned(log types.Log) (*ValidatorpoolApplicationsCleaned, error) {
	event := new(ValidatorpoolApplicationsCleaned)
	if err := _Validatorpool.contract.UnpackLog(event, "ApplicationsCleaned", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ValidatorpoolOwnershipTransferStartedIterator is returned from FilterOwnershipTransferStarted and is used to iterate over the raw logs and unpacked data for OwnershipTransferStarted events raised by the Validatorpool contract.
type ValidatorpoolOwnershipTransferStartedIterator struct {
	Event *ValidatorpoolOwnershipTransferStarted // Event containing the contract specifics and raw log

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
func (it *ValidatorpoolOwnershipTransferStartedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ValidatorpoolOwnershipTransferStarted)
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
		it.Event = new(ValidatorpoolOwnershipTransferStarted)
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
func (it *ValidatorpoolOwnershipTransferStartedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ValidatorpoolOwnershipTransferStartedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ValidatorpoolOwnershipTransferStarted represents a OwnershipTransferStarted event raised by the Validatorpool contract.
type ValidatorpoolOwnershipTransferStarted struct {
	PreviousOwner common.Address
	NewOwner      common.Address
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterOwnershipTransferStarted is a free log retrieval operation binding the contract event 0x38d16b8cac22d99fc7c124b9cd0de2d3fa1faef420bfe791d8c362d765e22700.
//
// Solidity: event OwnershipTransferStarted(address indexed previousOwner, address indexed newOwner)
func (_Validatorpool *ValidatorpoolFilterer) FilterOwnershipTransferStarted(opts *bind.FilterOpts, previousOwner []common.Address, newOwner []common.Address) (*ValidatorpoolOwnershipTransferStartedIterator, error) {
	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _Validatorpool.contract.FilterLogs(opts, "OwnershipTransferStarted", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return &ValidatorpoolOwnershipTransferStartedIterator{contract: _Validatorpool.contract, event: "OwnershipTransferStarted", logs: logs, sub: sub}, nil
}

// WatchOwnershipTransferStarted is a free log subscription operation binding the contract event 0x38d16b8cac22d99fc7c124b9cd0de2d3fa1faef420bfe791d8c362d765e22700.
//
// Solidity: event OwnershipTransferStarted(address indexed previousOwner, address indexed newOwner)
func (_Validatorpool *ValidatorpoolFilterer) WatchOwnershipTransferStarted(opts *bind.WatchOpts, sink chan<- *ValidatorpoolOwnershipTransferStarted, previousOwner []common.Address, newOwner []common.Address) (event.Subscription, error) {
	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _Validatorpool.contract.WatchLogs(opts, "OwnershipTransferStarted", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ValidatorpoolOwnershipTransferStarted)
				if err := _Validatorpool.contract.UnpackLog(event, "OwnershipTransferStarted", log); err != nil {
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
func (_Validatorpool *ValidatorpoolFilterer) ParseOwnershipTransferStarted(log types.Log) (*ValidatorpoolOwnershipTransferStarted, error) {
	event := new(ValidatorpoolOwnershipTransferStarted)
	if err := _Validatorpool.contract.UnpackLog(event, "OwnershipTransferStarted", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ValidatorpoolOwnershipTransferredIterator is returned from FilterOwnershipTransferred and is used to iterate over the raw logs and unpacked data for OwnershipTransferred events raised by the Validatorpool contract.
type ValidatorpoolOwnershipTransferredIterator struct {
	Event *ValidatorpoolOwnershipTransferred // Event containing the contract specifics and raw log

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
func (it *ValidatorpoolOwnershipTransferredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ValidatorpoolOwnershipTransferred)
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
		it.Event = new(ValidatorpoolOwnershipTransferred)
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
func (it *ValidatorpoolOwnershipTransferredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ValidatorpoolOwnershipTransferredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ValidatorpoolOwnershipTransferred represents a OwnershipTransferred event raised by the Validatorpool contract.
type ValidatorpoolOwnershipTransferred struct {
	PreviousOwner common.Address
	NewOwner      common.Address
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterOwnershipTransferred is a free log retrieval operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_Validatorpool *ValidatorpoolFilterer) FilterOwnershipTransferred(opts *bind.FilterOpts, previousOwner []common.Address, newOwner []common.Address) (*ValidatorpoolOwnershipTransferredIterator, error) {
	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _Validatorpool.contract.FilterLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return &ValidatorpoolOwnershipTransferredIterator{contract: _Validatorpool.contract, event: "OwnershipTransferred", logs: logs, sub: sub}, nil
}

// WatchOwnershipTransferred is a free log subscription operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_Validatorpool *ValidatorpoolFilterer) WatchOwnershipTransferred(opts *bind.WatchOpts, sink chan<- *ValidatorpoolOwnershipTransferred, previousOwner []common.Address, newOwner []common.Address) (event.Subscription, error) {
	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _Validatorpool.contract.WatchLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ValidatorpoolOwnershipTransferred)
				if err := _Validatorpool.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
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
func (_Validatorpool *ValidatorpoolFilterer) ParseOwnershipTransferred(log types.Log) (*ValidatorpoolOwnershipTransferred, error) {
	event := new(ValidatorpoolOwnershipTransferred)
	if err := _Validatorpool.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ValidatorpoolPrivilegeAddedIterator is returned from FilterPrivilegeAdded and is used to iterate over the raw logs and unpacked data for PrivilegeAdded events raised by the Validatorpool contract.
type ValidatorpoolPrivilegeAddedIterator struct {
	Event *ValidatorpoolPrivilegeAdded // Event containing the contract specifics and raw log

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
func (it *ValidatorpoolPrivilegeAddedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ValidatorpoolPrivilegeAdded)
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
		it.Event = new(ValidatorpoolPrivilegeAdded)
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
func (it *ValidatorpoolPrivilegeAddedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ValidatorpoolPrivilegeAddedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ValidatorpoolPrivilegeAdded represents a PrivilegeAdded event raised by the Validatorpool contract.
type ValidatorpoolPrivilegeAdded struct {
	Operator    common.Address
	PrivilegeId uint8
	Raw         types.Log // Blockchain specific contextual infos
}

// FilterPrivilegeAdded is a free log retrieval operation binding the contract event 0xbbc589ea14ee9e28889819229308f5768fdef26f4ce1ec0474711021fbe5ffda.
//
// Solidity: event PrivilegeAdded(address indexed operator, uint8 indexed privilegeId)
func (_Validatorpool *ValidatorpoolFilterer) FilterPrivilegeAdded(opts *bind.FilterOpts, operator []common.Address, privilegeId []uint8) (*ValidatorpoolPrivilegeAddedIterator, error) {
	var operatorRule []interface{}
	for _, operatorItem := range operator {
		operatorRule = append(operatorRule, operatorItem)
	}
	var privilegeIdRule []interface{}
	for _, privilegeIdItem := range privilegeId {
		privilegeIdRule = append(privilegeIdRule, privilegeIdItem)
	}

	logs, sub, err := _Validatorpool.contract.FilterLogs(opts, "PrivilegeAdded", operatorRule, privilegeIdRule)
	if err != nil {
		return nil, err
	}
	return &ValidatorpoolPrivilegeAddedIterator{contract: _Validatorpool.contract, event: "PrivilegeAdded", logs: logs, sub: sub}, nil
}

// WatchPrivilegeAdded is a free log subscription operation binding the contract event 0xbbc589ea14ee9e28889819229308f5768fdef26f4ce1ec0474711021fbe5ffda.
//
// Solidity: event PrivilegeAdded(address indexed operator, uint8 indexed privilegeId)
func (_Validatorpool *ValidatorpoolFilterer) WatchPrivilegeAdded(opts *bind.WatchOpts, sink chan<- *ValidatorpoolPrivilegeAdded, operator []common.Address, privilegeId []uint8) (event.Subscription, error) {
	var operatorRule []interface{}
	for _, operatorItem := range operator {
		operatorRule = append(operatorRule, operatorItem)
	}
	var privilegeIdRule []interface{}
	for _, privilegeIdItem := range privilegeId {
		privilegeIdRule = append(privilegeIdRule, privilegeIdItem)
	}

	logs, sub, err := _Validatorpool.contract.WatchLogs(opts, "PrivilegeAdded", operatorRule, privilegeIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ValidatorpoolPrivilegeAdded)
				if err := _Validatorpool.contract.UnpackLog(event, "PrivilegeAdded", log); err != nil {
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

// ParsePrivilegeAdded is a log parse operation binding the contract event 0xbbc589ea14ee9e28889819229308f5768fdef26f4ce1ec0474711021fbe5ffda.
//
// Solidity: event PrivilegeAdded(address indexed operator, uint8 indexed privilegeId)
func (_Validatorpool *ValidatorpoolFilterer) ParsePrivilegeAdded(log types.Log) (*ValidatorpoolPrivilegeAdded, error) {
	event := new(ValidatorpoolPrivilegeAdded)
	if err := _Validatorpool.contract.UnpackLog(event, "PrivilegeAdded", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ValidatorpoolPrivilegeRemovedIterator is returned from FilterPrivilegeRemoved and is used to iterate over the raw logs and unpacked data for PrivilegeRemoved events raised by the Validatorpool contract.
type ValidatorpoolPrivilegeRemovedIterator struct {
	Event *ValidatorpoolPrivilegeRemoved // Event containing the contract specifics and raw log

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
func (it *ValidatorpoolPrivilegeRemovedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ValidatorpoolPrivilegeRemoved)
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
		it.Event = new(ValidatorpoolPrivilegeRemoved)
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
func (it *ValidatorpoolPrivilegeRemovedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ValidatorpoolPrivilegeRemovedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ValidatorpoolPrivilegeRemoved represents a PrivilegeRemoved event raised by the Validatorpool contract.
type ValidatorpoolPrivilegeRemoved struct {
	Operator    common.Address
	PrivilegeId uint8
	Raw         types.Log // Blockchain specific contextual infos
}

// FilterPrivilegeRemoved is a free log retrieval operation binding the contract event 0xcb083e8b4c947cf054748f270b2d74abc1c2067aa154d9c6c1560851fdb210ef.
//
// Solidity: event PrivilegeRemoved(address indexed operator, uint8 indexed privilegeId)
func (_Validatorpool *ValidatorpoolFilterer) FilterPrivilegeRemoved(opts *bind.FilterOpts, operator []common.Address, privilegeId []uint8) (*ValidatorpoolPrivilegeRemovedIterator, error) {
	var operatorRule []interface{}
	for _, operatorItem := range operator {
		operatorRule = append(operatorRule, operatorItem)
	}
	var privilegeIdRule []interface{}
	for _, privilegeIdItem := range privilegeId {
		privilegeIdRule = append(privilegeIdRule, privilegeIdItem)
	}

	logs, sub, err := _Validatorpool.contract.FilterLogs(opts, "PrivilegeRemoved", operatorRule, privilegeIdRule)
	if err != nil {
		return nil, err
	}
	return &ValidatorpoolPrivilegeRemovedIterator{contract: _Validatorpool.contract, event: "PrivilegeRemoved", logs: logs, sub: sub}, nil
}

// WatchPrivilegeRemoved is a free log subscription operation binding the contract event 0xcb083e8b4c947cf054748f270b2d74abc1c2067aa154d9c6c1560851fdb210ef.
//
// Solidity: event PrivilegeRemoved(address indexed operator, uint8 indexed privilegeId)
func (_Validatorpool *ValidatorpoolFilterer) WatchPrivilegeRemoved(opts *bind.WatchOpts, sink chan<- *ValidatorpoolPrivilegeRemoved, operator []common.Address, privilegeId []uint8) (event.Subscription, error) {
	var operatorRule []interface{}
	for _, operatorItem := range operator {
		operatorRule = append(operatorRule, operatorItem)
	}
	var privilegeIdRule []interface{}
	for _, privilegeIdItem := range privilegeId {
		privilegeIdRule = append(privilegeIdRule, privilegeIdItem)
	}

	logs, sub, err := _Validatorpool.contract.WatchLogs(opts, "PrivilegeRemoved", operatorRule, privilegeIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ValidatorpoolPrivilegeRemoved)
				if err := _Validatorpool.contract.UnpackLog(event, "PrivilegeRemoved", log); err != nil {
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

// ParsePrivilegeRemoved is a log parse operation binding the contract event 0xcb083e8b4c947cf054748f270b2d74abc1c2067aa154d9c6c1560851fdb210ef.
//
// Solidity: event PrivilegeRemoved(address indexed operator, uint8 indexed privilegeId)
func (_Validatorpool *ValidatorpoolFilterer) ParsePrivilegeRemoved(log types.Log) (*ValidatorpoolPrivilegeRemoved, error) {
	event := new(ValidatorpoolPrivilegeRemoved)
	if err := _Validatorpool.contract.UnpackLog(event, "PrivilegeRemoved", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ValidatorpoolValidatorJoinedIterator is returned from FilterValidatorJoined and is used to iterate over the raw logs and unpacked data for ValidatorJoined events raised by the Validatorpool contract.
type ValidatorpoolValidatorJoinedIterator struct {
	Event *ValidatorpoolValidatorJoined // Event containing the contract specifics and raw log

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
func (it *ValidatorpoolValidatorJoinedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ValidatorpoolValidatorJoined)
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
		it.Event = new(ValidatorpoolValidatorJoined)
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
func (it *ValidatorpoolValidatorJoinedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ValidatorpoolValidatorJoinedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ValidatorpoolValidatorJoined represents a ValidatorJoined event raised by the Validatorpool contract.
type ValidatorpoolValidatorJoined struct {
	Operator common.Address
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterValidatorJoined is a free log retrieval operation binding the contract event 0xd5828184f48f65962d10eac907318df85953d4e3542a0f09b5932ee3fe398bdd.
//
// Solidity: event ValidatorJoined(address indexed operator)
func (_Validatorpool *ValidatorpoolFilterer) FilterValidatorJoined(opts *bind.FilterOpts, operator []common.Address) (*ValidatorpoolValidatorJoinedIterator, error) {
	var operatorRule []interface{}
	for _, operatorItem := range operator {
		operatorRule = append(operatorRule, operatorItem)
	}

	logs, sub, err := _Validatorpool.contract.FilterLogs(opts, "ValidatorJoined", operatorRule)
	if err != nil {
		return nil, err
	}
	return &ValidatorpoolValidatorJoinedIterator{contract: _Validatorpool.contract, event: "ValidatorJoined", logs: logs, sub: sub}, nil
}

// WatchValidatorJoined is a free log subscription operation binding the contract event 0xd5828184f48f65962d10eac907318df85953d4e3542a0f09b5932ee3fe398bdd.
//
// Solidity: event ValidatorJoined(address indexed operator)
func (_Validatorpool *ValidatorpoolFilterer) WatchValidatorJoined(opts *bind.WatchOpts, sink chan<- *ValidatorpoolValidatorJoined, operator []common.Address) (event.Subscription, error) {
	var operatorRule []interface{}
	for _, operatorItem := range operator {
		operatorRule = append(operatorRule, operatorItem)
	}

	logs, sub, err := _Validatorpool.contract.WatchLogs(opts, "ValidatorJoined", operatorRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ValidatorpoolValidatorJoined)
				if err := _Validatorpool.contract.UnpackLog(event, "ValidatorJoined", log); err != nil {
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

// ParseValidatorJoined is a log parse operation binding the contract event 0xd5828184f48f65962d10eac907318df85953d4e3542a0f09b5932ee3fe398bdd.
//
// Solidity: event ValidatorJoined(address indexed operator)
func (_Validatorpool *ValidatorpoolFilterer) ParseValidatorJoined(log types.Log) (*ValidatorpoolValidatorJoined, error) {
	event := new(ValidatorpoolValidatorJoined)
	if err := _Validatorpool.contract.UnpackLog(event, "ValidatorJoined", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ValidatorpoolValidatorKickedIterator is returned from FilterValidatorKicked and is used to iterate over the raw logs and unpacked data for ValidatorKicked events raised by the Validatorpool contract.
type ValidatorpoolValidatorKickedIterator struct {
	Event *ValidatorpoolValidatorKicked // Event containing the contract specifics and raw log

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
func (it *ValidatorpoolValidatorKickedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ValidatorpoolValidatorKicked)
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
		it.Event = new(ValidatorpoolValidatorKicked)
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
func (it *ValidatorpoolValidatorKickedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ValidatorpoolValidatorKickedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ValidatorpoolValidatorKicked represents a ValidatorKicked event raised by the Validatorpool contract.
type ValidatorpoolValidatorKicked struct {
	Operator common.Address
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterValidatorKicked is a free log retrieval operation binding the contract event 0xf7d421add37cde2bb92a3f14d758be4f15659c8fb7d9144631026e862dc0a587.
//
// Solidity: event ValidatorKicked(address indexed operator)
func (_Validatorpool *ValidatorpoolFilterer) FilterValidatorKicked(opts *bind.FilterOpts, operator []common.Address) (*ValidatorpoolValidatorKickedIterator, error) {
	var operatorRule []interface{}
	for _, operatorItem := range operator {
		operatorRule = append(operatorRule, operatorItem)
	}

	logs, sub, err := _Validatorpool.contract.FilterLogs(opts, "ValidatorKicked", operatorRule)
	if err != nil {
		return nil, err
	}
	return &ValidatorpoolValidatorKickedIterator{contract: _Validatorpool.contract, event: "ValidatorKicked", logs: logs, sub: sub}, nil
}

// WatchValidatorKicked is a free log subscription operation binding the contract event 0xf7d421add37cde2bb92a3f14d758be4f15659c8fb7d9144631026e862dc0a587.
//
// Solidity: event ValidatorKicked(address indexed operator)
func (_Validatorpool *ValidatorpoolFilterer) WatchValidatorKicked(opts *bind.WatchOpts, sink chan<- *ValidatorpoolValidatorKicked, operator []common.Address) (event.Subscription, error) {
	var operatorRule []interface{}
	for _, operatorItem := range operator {
		operatorRule = append(operatorRule, operatorItem)
	}

	logs, sub, err := _Validatorpool.contract.WatchLogs(opts, "ValidatorKicked", operatorRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ValidatorpoolValidatorKicked)
				if err := _Validatorpool.contract.UnpackLog(event, "ValidatorKicked", log); err != nil {
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

// ParseValidatorKicked is a log parse operation binding the contract event 0xf7d421add37cde2bb92a3f14d758be4f15659c8fb7d9144631026e862dc0a587.
//
// Solidity: event ValidatorKicked(address indexed operator)
func (_Validatorpool *ValidatorpoolFilterer) ParseValidatorKicked(log types.Log) (*ValidatorpoolValidatorKicked, error) {
	event := new(ValidatorpoolValidatorKicked)
	if err := _Validatorpool.contract.UnpackLog(event, "ValidatorKicked", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ValidatorpoolValidatorLeftIterator is returned from FilterValidatorLeft and is used to iterate over the raw logs and unpacked data for ValidatorLeft events raised by the Validatorpool contract.
type ValidatorpoolValidatorLeftIterator struct {
	Event *ValidatorpoolValidatorLeft // Event containing the contract specifics and raw log

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
func (it *ValidatorpoolValidatorLeftIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ValidatorpoolValidatorLeft)
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
		it.Event = new(ValidatorpoolValidatorLeft)
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
func (it *ValidatorpoolValidatorLeftIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ValidatorpoolValidatorLeftIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ValidatorpoolValidatorLeft represents a ValidatorLeft event raised by the Validatorpool contract.
type ValidatorpoolValidatorLeft struct {
	Operator common.Address
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterValidatorLeft is a free log retrieval operation binding the contract event 0x457a63bb4fcc8a2d38fb6141645ce9afbb084a1e63edee2799243c96a0498a0e.
//
// Solidity: event ValidatorLeft(address indexed operator)
func (_Validatorpool *ValidatorpoolFilterer) FilterValidatorLeft(opts *bind.FilterOpts, operator []common.Address) (*ValidatorpoolValidatorLeftIterator, error) {
	var operatorRule []interface{}
	for _, operatorItem := range operator {
		operatorRule = append(operatorRule, operatorItem)
	}

	logs, sub, err := _Validatorpool.contract.FilterLogs(opts, "ValidatorLeft", operatorRule)
	if err != nil {
		return nil, err
	}
	return &ValidatorpoolValidatorLeftIterator{contract: _Validatorpool.contract, event: "ValidatorLeft", logs: logs, sub: sub}, nil
}

// WatchValidatorLeft is a free log subscription operation binding the contract event 0x457a63bb4fcc8a2d38fb6141645ce9afbb084a1e63edee2799243c96a0498a0e.
//
// Solidity: event ValidatorLeft(address indexed operator)
func (_Validatorpool *ValidatorpoolFilterer) WatchValidatorLeft(opts *bind.WatchOpts, sink chan<- *ValidatorpoolValidatorLeft, operator []common.Address) (event.Subscription, error) {
	var operatorRule []interface{}
	for _, operatorItem := range operator {
		operatorRule = append(operatorRule, operatorItem)
	}

	logs, sub, err := _Validatorpool.contract.WatchLogs(opts, "ValidatorLeft", operatorRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ValidatorpoolValidatorLeft)
				if err := _Validatorpool.contract.UnpackLog(event, "ValidatorLeft", log); err != nil {
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

// ParseValidatorLeft is a log parse operation binding the contract event 0x457a63bb4fcc8a2d38fb6141645ce9afbb084a1e63edee2799243c96a0498a0e.
//
// Solidity: event ValidatorLeft(address indexed operator)
func (_Validatorpool *ValidatorpoolFilterer) ParseValidatorLeft(log types.Log) (*ValidatorpoolValidatorLeft, error) {
	event := new(ValidatorpoolValidatorLeft)
	if err := _Validatorpool.contract.UnpackLog(event, "ValidatorLeft", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
