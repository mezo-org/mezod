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
	_ = abi.ConvertType
)

// IBridgeTypesBitcoinTxInfo is an auto generated low-level Go binding around an user-defined struct.
type IBridgeTypesBitcoinTxInfo struct {
	Version      [4]byte
	InputVector  []byte
	OutputVector []byte
	Locktime     [4]byte
}

// IBridgeTypesDepositRevealInfo is an auto generated low-level Go binding around an user-defined struct.
type IBridgeTypesDepositRevealInfo struct {
	FundingOutputIndex uint32
	BlindingFactor     [8]byte
	WalletPubKeyHash   [20]byte
	RefundPubKeyHash   [20]byte
	RefundLocktime     [4]byte
	Vault              common.Address
}

// BitcoinBridgeMetaData contains all meta data concerning the BitcoinBridge contract.
var BitcoinBridgeMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"target\",\"type\":\"address\"}],\"name\":\"AddressEmptyCode\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"}],\"name\":\"AddressInsufficientBalance\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"AmountBelowMinTBTCAmount\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"FailedInnerCall\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"InvalidInitialization\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"MinTBTCAmountIsZero\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"NotInitializing\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"}],\"name\":\"OwnableInvalidOwner\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"}],\"name\":\"OwnableUnauthorizedAccount\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"RecipientIsZeroAddress\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"token\",\"type\":\"address\"}],\"name\":\"SafeERC20FailedOperation\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"TBTCTokenIsZeroAddress\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"enumBitcoinBridge.DepositState\",\"name\":\"actualState\",\"type\":\"uint8\"},{\"internalType\":\"enumBitcoinBridge.DepositState\",\"name\":\"expectedState\",\"type\":\"uint8\"}],\"name\":\"UnexpectedDepositState\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"actualExtraData\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"expectedExtraData\",\"type\":\"bytes32\"}],\"name\":\"UnexpectedExtraData\",\"type\":\"error\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"sequenceNumber\",\"type\":\"uint256\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"recipient\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"tbtcAmount\",\"type\":\"uint256\"}],\"name\":\"AssetsLocked\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"depositKey\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"initialAmount\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"tbtcAmount\",\"type\":\"uint256\"}],\"name\":\"DepositFinalized\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"depositKey\",\"type\":\"uint256\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"recipient\",\"type\":\"address\"}],\"name\":\"DepositInitialized\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint64\",\"name\":\"version\",\"type\":\"uint64\"}],\"name\":\"Initialized\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"minTBTCAmount\",\"type\":\"uint256\"}],\"name\":\"MinTBTCAmountUpdated\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"previousOwner\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"OwnershipTransferStarted\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"previousOwner\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"OwnershipTransferred\",\"type\":\"event\"},{\"inputs\":[],\"name\":\"SATOSHI_MULTIPLIER\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"acceptOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"bridge\",\"outputs\":[{\"internalType\":\"contractIBridge\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"recipient\",\"type\":\"address\"}],\"name\":\"bridgeTBTC\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"recipient\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"deadline\",\"type\":\"uint256\"},{\"internalType\":\"uint8\",\"name\":\"v\",\"type\":\"uint8\"},{\"internalType\":\"bytes32\",\"name\":\"r\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"s\",\"type\":\"bytes32\"}],\"name\":\"bridgeTBTCWithPermit\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"deposits\",\"outputs\":[{\"internalType\":\"enumBitcoinBridge.DepositState\",\"name\":\"\",\"type\":\"uint8\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"depositKey\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"recipient\",\"type\":\"address\"}],\"name\":\"finalizeBTCBridging\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_bridge\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"_tbtcVault\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"_tbtcToken\",\"type\":\"address\"}],\"name\":\"initialize\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"components\":[{\"internalType\":\"bytes4\",\"name\":\"version\",\"type\":\"bytes4\"},{\"internalType\":\"bytes\",\"name\":\"inputVector\",\"type\":\"bytes\"},{\"internalType\":\"bytes\",\"name\":\"outputVector\",\"type\":\"bytes\"},{\"internalType\":\"bytes4\",\"name\":\"locktime\",\"type\":\"bytes4\"}],\"internalType\":\"structIBridgeTypes.BitcoinTxInfo\",\"name\":\"fundingTx\",\"type\":\"tuple\"},{\"components\":[{\"internalType\":\"uint32\",\"name\":\"fundingOutputIndex\",\"type\":\"uint32\"},{\"internalType\":\"bytes8\",\"name\":\"blindingFactor\",\"type\":\"bytes8\"},{\"internalType\":\"bytes20\",\"name\":\"walletPubKeyHash\",\"type\":\"bytes20\"},{\"internalType\":\"bytes20\",\"name\":\"refundPubKeyHash\",\"type\":\"bytes20\"},{\"internalType\":\"bytes4\",\"name\":\"refundLocktime\",\"type\":\"bytes4\"},{\"internalType\":\"address\",\"name\":\"vault\",\"type\":\"address\"}],\"internalType\":\"structIBridgeTypes.DepositRevealInfo\",\"name\":\"reveal\",\"type\":\"tuple\"},{\"internalType\":\"address\",\"name\":\"recipient\",\"type\":\"address\"}],\"name\":\"initializeBTCBridging\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"minTBTCAmount\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"owner\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"pendingOwner\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"renounceOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"sequence\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"tbtcToken\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"tbtcVault\",\"outputs\":[{\"internalType\":\"contractITBTCVault\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"transferOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"newMinTBTCAmount\",\"type\":\"uint256\"}],\"name\":\"updateMinTBTCAmount\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]",
}

// BitcoinBridgeABI is the input ABI used to generate the binding from.
// Deprecated: Use BitcoinBridgeMetaData.ABI instead.
var BitcoinBridgeABI = BitcoinBridgeMetaData.ABI

// BitcoinBridge is an auto generated Go binding around an Ethereum contract.
type BitcoinBridge struct {
	BitcoinBridgeCaller     // Read-only binding to the contract
	BitcoinBridgeTransactor // Write-only binding to the contract
	BitcoinBridgeFilterer   // Log filterer for contract events
}

// BitcoinBridgeCaller is an auto generated read-only Go binding around an Ethereum contract.
type BitcoinBridgeCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// BitcoinBridgeTransactor is an auto generated write-only Go binding around an Ethereum contract.
type BitcoinBridgeTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// BitcoinBridgeFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type BitcoinBridgeFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// BitcoinBridgeSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type BitcoinBridgeSession struct {
	Contract     *BitcoinBridge    // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// BitcoinBridgeCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type BitcoinBridgeCallerSession struct {
	Contract *BitcoinBridgeCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts        // Call options to use throughout this session
}

// BitcoinBridgeTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type BitcoinBridgeTransactorSession struct {
	Contract     *BitcoinBridgeTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts        // Transaction auth options to use throughout this session
}

// BitcoinBridgeRaw is an auto generated low-level Go binding around an Ethereum contract.
type BitcoinBridgeRaw struct {
	Contract *BitcoinBridge // Generic contract binding to access the raw methods on
}

// BitcoinBridgeCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type BitcoinBridgeCallerRaw struct {
	Contract *BitcoinBridgeCaller // Generic read-only contract binding to access the raw methods on
}

// BitcoinBridgeTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type BitcoinBridgeTransactorRaw struct {
	Contract *BitcoinBridgeTransactor // Generic write-only contract binding to access the raw methods on
}

// NewBitcoinBridge creates a new instance of BitcoinBridge, bound to a specific deployed contract.
func NewBitcoinBridge(address common.Address, backend bind.ContractBackend) (*BitcoinBridge, error) {
	contract, err := bindBitcoinBridge(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &BitcoinBridge{BitcoinBridgeCaller: BitcoinBridgeCaller{contract: contract}, BitcoinBridgeTransactor: BitcoinBridgeTransactor{contract: contract}, BitcoinBridgeFilterer: BitcoinBridgeFilterer{contract: contract}}, nil
}

// NewBitcoinBridgeCaller creates a new read-only instance of BitcoinBridge, bound to a specific deployed contract.
func NewBitcoinBridgeCaller(address common.Address, caller bind.ContractCaller) (*BitcoinBridgeCaller, error) {
	contract, err := bindBitcoinBridge(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &BitcoinBridgeCaller{contract: contract}, nil
}

// NewBitcoinBridgeTransactor creates a new write-only instance of BitcoinBridge, bound to a specific deployed contract.
func NewBitcoinBridgeTransactor(address common.Address, transactor bind.ContractTransactor) (*BitcoinBridgeTransactor, error) {
	contract, err := bindBitcoinBridge(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &BitcoinBridgeTransactor{contract: contract}, nil
}

// NewBitcoinBridgeFilterer creates a new log filterer instance of BitcoinBridge, bound to a specific deployed contract.
func NewBitcoinBridgeFilterer(address common.Address, filterer bind.ContractFilterer) (*BitcoinBridgeFilterer, error) {
	contract, err := bindBitcoinBridge(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &BitcoinBridgeFilterer{contract: contract}, nil
}

// bindBitcoinBridge binds a generic wrapper to an already deployed contract.
func bindBitcoinBridge(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := BitcoinBridgeMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_BitcoinBridge *BitcoinBridgeRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _BitcoinBridge.Contract.BitcoinBridgeCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_BitcoinBridge *BitcoinBridgeRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _BitcoinBridge.Contract.BitcoinBridgeTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_BitcoinBridge *BitcoinBridgeRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _BitcoinBridge.Contract.BitcoinBridgeTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_BitcoinBridge *BitcoinBridgeCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _BitcoinBridge.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_BitcoinBridge *BitcoinBridgeTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _BitcoinBridge.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_BitcoinBridge *BitcoinBridgeTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _BitcoinBridge.Contract.contract.Transact(opts, method, params...)
}

// SATOSHIMULTIPLIER is a free data retrieval call binding the contract method 0xc7ba0347.
//
// Solidity: function SATOSHI_MULTIPLIER() view returns(uint256)
func (_BitcoinBridge *BitcoinBridgeCaller) SATOSHIMULTIPLIER(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _BitcoinBridge.contract.Call(opts, &out, "SATOSHI_MULTIPLIER")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// SATOSHIMULTIPLIER is a free data retrieval call binding the contract method 0xc7ba0347.
//
// Solidity: function SATOSHI_MULTIPLIER() view returns(uint256)
func (_BitcoinBridge *BitcoinBridgeSession) SATOSHIMULTIPLIER() (*big.Int, error) {
	return _BitcoinBridge.Contract.SATOSHIMULTIPLIER(&_BitcoinBridge.CallOpts)
}

// SATOSHIMULTIPLIER is a free data retrieval call binding the contract method 0xc7ba0347.
//
// Solidity: function SATOSHI_MULTIPLIER() view returns(uint256)
func (_BitcoinBridge *BitcoinBridgeCallerSession) SATOSHIMULTIPLIER() (*big.Int, error) {
	return _BitcoinBridge.Contract.SATOSHIMULTIPLIER(&_BitcoinBridge.CallOpts)
}

// Bridge is a free data retrieval call binding the contract method 0xe78cea92.
//
// Solidity: function bridge() view returns(address)
func (_BitcoinBridge *BitcoinBridgeCaller) Bridge(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _BitcoinBridge.contract.Call(opts, &out, "bridge")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Bridge is a free data retrieval call binding the contract method 0xe78cea92.
//
// Solidity: function bridge() view returns(address)
func (_BitcoinBridge *BitcoinBridgeSession) Bridge() (common.Address, error) {
	return _BitcoinBridge.Contract.Bridge(&_BitcoinBridge.CallOpts)
}

// Bridge is a free data retrieval call binding the contract method 0xe78cea92.
//
// Solidity: function bridge() view returns(address)
func (_BitcoinBridge *BitcoinBridgeCallerSession) Bridge() (common.Address, error) {
	return _BitcoinBridge.Contract.Bridge(&_BitcoinBridge.CallOpts)
}

// Deposits is a free data retrieval call binding the contract method 0xb02c43d0.
//
// Solidity: function deposits(uint256 ) view returns(uint8)
func (_BitcoinBridge *BitcoinBridgeCaller) Deposits(opts *bind.CallOpts, arg0 *big.Int) (uint8, error) {
	var out []interface{}
	err := _BitcoinBridge.contract.Call(opts, &out, "deposits", arg0)

	if err != nil {
		return *new(uint8), err
	}

	out0 := *abi.ConvertType(out[0], new(uint8)).(*uint8)

	return out0, err

}

// Deposits is a free data retrieval call binding the contract method 0xb02c43d0.
//
// Solidity: function deposits(uint256 ) view returns(uint8)
func (_BitcoinBridge *BitcoinBridgeSession) Deposits(arg0 *big.Int) (uint8, error) {
	return _BitcoinBridge.Contract.Deposits(&_BitcoinBridge.CallOpts, arg0)
}

// Deposits is a free data retrieval call binding the contract method 0xb02c43d0.
//
// Solidity: function deposits(uint256 ) view returns(uint8)
func (_BitcoinBridge *BitcoinBridgeCallerSession) Deposits(arg0 *big.Int) (uint8, error) {
	return _BitcoinBridge.Contract.Deposits(&_BitcoinBridge.CallOpts, arg0)
}

// MinTBTCAmount is a free data retrieval call binding the contract method 0xdab1b4bd.
//
// Solidity: function minTBTCAmount() view returns(uint256)
func (_BitcoinBridge *BitcoinBridgeCaller) MinTBTCAmount(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _BitcoinBridge.contract.Call(opts, &out, "minTBTCAmount")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// MinTBTCAmount is a free data retrieval call binding the contract method 0xdab1b4bd.
//
// Solidity: function minTBTCAmount() view returns(uint256)
func (_BitcoinBridge *BitcoinBridgeSession) MinTBTCAmount() (*big.Int, error) {
	return _BitcoinBridge.Contract.MinTBTCAmount(&_BitcoinBridge.CallOpts)
}

// MinTBTCAmount is a free data retrieval call binding the contract method 0xdab1b4bd.
//
// Solidity: function minTBTCAmount() view returns(uint256)
func (_BitcoinBridge *BitcoinBridgeCallerSession) MinTBTCAmount() (*big.Int, error) {
	return _BitcoinBridge.Contract.MinTBTCAmount(&_BitcoinBridge.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_BitcoinBridge *BitcoinBridgeCaller) Owner(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _BitcoinBridge.contract.Call(opts, &out, "owner")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_BitcoinBridge *BitcoinBridgeSession) Owner() (common.Address, error) {
	return _BitcoinBridge.Contract.Owner(&_BitcoinBridge.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_BitcoinBridge *BitcoinBridgeCallerSession) Owner() (common.Address, error) {
	return _BitcoinBridge.Contract.Owner(&_BitcoinBridge.CallOpts)
}

// PendingOwner is a free data retrieval call binding the contract method 0xe30c3978.
//
// Solidity: function pendingOwner() view returns(address)
func (_BitcoinBridge *BitcoinBridgeCaller) PendingOwner(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _BitcoinBridge.contract.Call(opts, &out, "pendingOwner")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// PendingOwner is a free data retrieval call binding the contract method 0xe30c3978.
//
// Solidity: function pendingOwner() view returns(address)
func (_BitcoinBridge *BitcoinBridgeSession) PendingOwner() (common.Address, error) {
	return _BitcoinBridge.Contract.PendingOwner(&_BitcoinBridge.CallOpts)
}

// PendingOwner is a free data retrieval call binding the contract method 0xe30c3978.
//
// Solidity: function pendingOwner() view returns(address)
func (_BitcoinBridge *BitcoinBridgeCallerSession) PendingOwner() (common.Address, error) {
	return _BitcoinBridge.Contract.PendingOwner(&_BitcoinBridge.CallOpts)
}

// Sequence is a free data retrieval call binding the contract method 0x529d15cc.
//
// Solidity: function sequence() view returns(uint256)
func (_BitcoinBridge *BitcoinBridgeCaller) Sequence(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _BitcoinBridge.contract.Call(opts, &out, "sequence")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// Sequence is a free data retrieval call binding the contract method 0x529d15cc.
//
// Solidity: function sequence() view returns(uint256)
func (_BitcoinBridge *BitcoinBridgeSession) Sequence() (*big.Int, error) {
	return _BitcoinBridge.Contract.Sequence(&_BitcoinBridge.CallOpts)
}

// Sequence is a free data retrieval call binding the contract method 0x529d15cc.
//
// Solidity: function sequence() view returns(uint256)
func (_BitcoinBridge *BitcoinBridgeCallerSession) Sequence() (*big.Int, error) {
	return _BitcoinBridge.Contract.Sequence(&_BitcoinBridge.CallOpts)
}

// TbtcToken is a free data retrieval call binding the contract method 0xe5d3d714.
//
// Solidity: function tbtcToken() view returns(address)
func (_BitcoinBridge *BitcoinBridgeCaller) TbtcToken(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _BitcoinBridge.contract.Call(opts, &out, "tbtcToken")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// TbtcToken is a free data retrieval call binding the contract method 0xe5d3d714.
//
// Solidity: function tbtcToken() view returns(address)
func (_BitcoinBridge *BitcoinBridgeSession) TbtcToken() (common.Address, error) {
	return _BitcoinBridge.Contract.TbtcToken(&_BitcoinBridge.CallOpts)
}

// TbtcToken is a free data retrieval call binding the contract method 0xe5d3d714.
//
// Solidity: function tbtcToken() view returns(address)
func (_BitcoinBridge *BitcoinBridgeCallerSession) TbtcToken() (common.Address, error) {
	return _BitcoinBridge.Contract.TbtcToken(&_BitcoinBridge.CallOpts)
}

// TbtcVault is a free data retrieval call binding the contract method 0x0f36403a.
//
// Solidity: function tbtcVault() view returns(address)
func (_BitcoinBridge *BitcoinBridgeCaller) TbtcVault(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _BitcoinBridge.contract.Call(opts, &out, "tbtcVault")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// TbtcVault is a free data retrieval call binding the contract method 0x0f36403a.
//
// Solidity: function tbtcVault() view returns(address)
func (_BitcoinBridge *BitcoinBridgeSession) TbtcVault() (common.Address, error) {
	return _BitcoinBridge.Contract.TbtcVault(&_BitcoinBridge.CallOpts)
}

// TbtcVault is a free data retrieval call binding the contract method 0x0f36403a.
//
// Solidity: function tbtcVault() view returns(address)
func (_BitcoinBridge *BitcoinBridgeCallerSession) TbtcVault() (common.Address, error) {
	return _BitcoinBridge.Contract.TbtcVault(&_BitcoinBridge.CallOpts)
}

// AcceptOwnership is a paid mutator transaction binding the contract method 0x79ba5097.
//
// Solidity: function acceptOwnership() returns()
func (_BitcoinBridge *BitcoinBridgeTransactor) AcceptOwnership(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _BitcoinBridge.contract.Transact(opts, "acceptOwnership")
}

// AcceptOwnership is a paid mutator transaction binding the contract method 0x79ba5097.
//
// Solidity: function acceptOwnership() returns()
func (_BitcoinBridge *BitcoinBridgeSession) AcceptOwnership() (*types.Transaction, error) {
	return _BitcoinBridge.Contract.AcceptOwnership(&_BitcoinBridge.TransactOpts)
}

// AcceptOwnership is a paid mutator transaction binding the contract method 0x79ba5097.
//
// Solidity: function acceptOwnership() returns()
func (_BitcoinBridge *BitcoinBridgeTransactorSession) AcceptOwnership() (*types.Transaction, error) {
	return _BitcoinBridge.Contract.AcceptOwnership(&_BitcoinBridge.TransactOpts)
}

// BridgeTBTC is a paid mutator transaction binding the contract method 0xdf4d4663.
//
// Solidity: function bridgeTBTC(uint256 amount, address recipient) returns()
func (_BitcoinBridge *BitcoinBridgeTransactor) BridgeTBTC(opts *bind.TransactOpts, amount *big.Int, recipient common.Address) (*types.Transaction, error) {
	return _BitcoinBridge.contract.Transact(opts, "bridgeTBTC", amount, recipient)
}

// BridgeTBTC is a paid mutator transaction binding the contract method 0xdf4d4663.
//
// Solidity: function bridgeTBTC(uint256 amount, address recipient) returns()
func (_BitcoinBridge *BitcoinBridgeSession) BridgeTBTC(amount *big.Int, recipient common.Address) (*types.Transaction, error) {
	return _BitcoinBridge.Contract.BridgeTBTC(&_BitcoinBridge.TransactOpts, amount, recipient)
}

// BridgeTBTC is a paid mutator transaction binding the contract method 0xdf4d4663.
//
// Solidity: function bridgeTBTC(uint256 amount, address recipient) returns()
func (_BitcoinBridge *BitcoinBridgeTransactorSession) BridgeTBTC(amount *big.Int, recipient common.Address) (*types.Transaction, error) {
	return _BitcoinBridge.Contract.BridgeTBTC(&_BitcoinBridge.TransactOpts, amount, recipient)
}

// BridgeTBTCWithPermit is a paid mutator transaction binding the contract method 0x427f9568.
//
// Solidity: function bridgeTBTCWithPermit(uint256 amount, address recipient, uint256 deadline, uint8 v, bytes32 r, bytes32 s) returns()
func (_BitcoinBridge *BitcoinBridgeTransactor) BridgeTBTCWithPermit(opts *bind.TransactOpts, amount *big.Int, recipient common.Address, deadline *big.Int, v uint8, r [32]byte, s [32]byte) (*types.Transaction, error) {
	return _BitcoinBridge.contract.Transact(opts, "bridgeTBTCWithPermit", amount, recipient, deadline, v, r, s)
}

// BridgeTBTCWithPermit is a paid mutator transaction binding the contract method 0x427f9568.
//
// Solidity: function bridgeTBTCWithPermit(uint256 amount, address recipient, uint256 deadline, uint8 v, bytes32 r, bytes32 s) returns()
func (_BitcoinBridge *BitcoinBridgeSession) BridgeTBTCWithPermit(amount *big.Int, recipient common.Address, deadline *big.Int, v uint8, r [32]byte, s [32]byte) (*types.Transaction, error) {
	return _BitcoinBridge.Contract.BridgeTBTCWithPermit(&_BitcoinBridge.TransactOpts, amount, recipient, deadline, v, r, s)
}

// BridgeTBTCWithPermit is a paid mutator transaction binding the contract method 0x427f9568.
//
// Solidity: function bridgeTBTCWithPermit(uint256 amount, address recipient, uint256 deadline, uint8 v, bytes32 r, bytes32 s) returns()
func (_BitcoinBridge *BitcoinBridgeTransactorSession) BridgeTBTCWithPermit(amount *big.Int, recipient common.Address, deadline *big.Int, v uint8, r [32]byte, s [32]byte) (*types.Transaction, error) {
	return _BitcoinBridge.Contract.BridgeTBTCWithPermit(&_BitcoinBridge.TransactOpts, amount, recipient, deadline, v, r, s)
}

// FinalizeBTCBridging is a paid mutator transaction binding the contract method 0x24f90de9.
//
// Solidity: function finalizeBTCBridging(uint256 depositKey, address recipient) returns()
func (_BitcoinBridge *BitcoinBridgeTransactor) FinalizeBTCBridging(opts *bind.TransactOpts, depositKey *big.Int, recipient common.Address) (*types.Transaction, error) {
	return _BitcoinBridge.contract.Transact(opts, "finalizeBTCBridging", depositKey, recipient)
}

// FinalizeBTCBridging is a paid mutator transaction binding the contract method 0x24f90de9.
//
// Solidity: function finalizeBTCBridging(uint256 depositKey, address recipient) returns()
func (_BitcoinBridge *BitcoinBridgeSession) FinalizeBTCBridging(depositKey *big.Int, recipient common.Address) (*types.Transaction, error) {
	return _BitcoinBridge.Contract.FinalizeBTCBridging(&_BitcoinBridge.TransactOpts, depositKey, recipient)
}

// FinalizeBTCBridging is a paid mutator transaction binding the contract method 0x24f90de9.
//
// Solidity: function finalizeBTCBridging(uint256 depositKey, address recipient) returns()
func (_BitcoinBridge *BitcoinBridgeTransactorSession) FinalizeBTCBridging(depositKey *big.Int, recipient common.Address) (*types.Transaction, error) {
	return _BitcoinBridge.Contract.FinalizeBTCBridging(&_BitcoinBridge.TransactOpts, depositKey, recipient)
}

// Initialize is a paid mutator transaction binding the contract method 0xc0c53b8b.
//
// Solidity: function initialize(address _bridge, address _tbtcVault, address _tbtcToken) returns()
func (_BitcoinBridge *BitcoinBridgeTransactor) Initialize(opts *bind.TransactOpts, _bridge common.Address, _tbtcVault common.Address, _tbtcToken common.Address) (*types.Transaction, error) {
	return _BitcoinBridge.contract.Transact(opts, "initialize", _bridge, _tbtcVault, _tbtcToken)
}

// Initialize is a paid mutator transaction binding the contract method 0xc0c53b8b.
//
// Solidity: function initialize(address _bridge, address _tbtcVault, address _tbtcToken) returns()
func (_BitcoinBridge *BitcoinBridgeSession) Initialize(_bridge common.Address, _tbtcVault common.Address, _tbtcToken common.Address) (*types.Transaction, error) {
	return _BitcoinBridge.Contract.Initialize(&_BitcoinBridge.TransactOpts, _bridge, _tbtcVault, _tbtcToken)
}

// Initialize is a paid mutator transaction binding the contract method 0xc0c53b8b.
//
// Solidity: function initialize(address _bridge, address _tbtcVault, address _tbtcToken) returns()
func (_BitcoinBridge *BitcoinBridgeTransactorSession) Initialize(_bridge common.Address, _tbtcVault common.Address, _tbtcToken common.Address) (*types.Transaction, error) {
	return _BitcoinBridge.Contract.Initialize(&_BitcoinBridge.TransactOpts, _bridge, _tbtcVault, _tbtcToken)
}

// InitializeBTCBridging is a paid mutator transaction binding the contract method 0x6f64aca2.
//
// Solidity: function initializeBTCBridging((bytes4,bytes,bytes,bytes4) fundingTx, (uint32,bytes8,bytes20,bytes20,bytes4,address) reveal, address recipient) returns()
func (_BitcoinBridge *BitcoinBridgeTransactor) InitializeBTCBridging(opts *bind.TransactOpts, fundingTx IBridgeTypesBitcoinTxInfo, reveal IBridgeTypesDepositRevealInfo, recipient common.Address) (*types.Transaction, error) {
	return _BitcoinBridge.contract.Transact(opts, "initializeBTCBridging", fundingTx, reveal, recipient)
}

// InitializeBTCBridging is a paid mutator transaction binding the contract method 0x6f64aca2.
//
// Solidity: function initializeBTCBridging((bytes4,bytes,bytes,bytes4) fundingTx, (uint32,bytes8,bytes20,bytes20,bytes4,address) reveal, address recipient) returns()
func (_BitcoinBridge *BitcoinBridgeSession) InitializeBTCBridging(fundingTx IBridgeTypesBitcoinTxInfo, reveal IBridgeTypesDepositRevealInfo, recipient common.Address) (*types.Transaction, error) {
	return _BitcoinBridge.Contract.InitializeBTCBridging(&_BitcoinBridge.TransactOpts, fundingTx, reveal, recipient)
}

// InitializeBTCBridging is a paid mutator transaction binding the contract method 0x6f64aca2.
//
// Solidity: function initializeBTCBridging((bytes4,bytes,bytes,bytes4) fundingTx, (uint32,bytes8,bytes20,bytes20,bytes4,address) reveal, address recipient) returns()
func (_BitcoinBridge *BitcoinBridgeTransactorSession) InitializeBTCBridging(fundingTx IBridgeTypesBitcoinTxInfo, reveal IBridgeTypesDepositRevealInfo, recipient common.Address) (*types.Transaction, error) {
	return _BitcoinBridge.Contract.InitializeBTCBridging(&_BitcoinBridge.TransactOpts, fundingTx, reveal, recipient)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_BitcoinBridge *BitcoinBridgeTransactor) RenounceOwnership(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _BitcoinBridge.contract.Transact(opts, "renounceOwnership")
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_BitcoinBridge *BitcoinBridgeSession) RenounceOwnership() (*types.Transaction, error) {
	return _BitcoinBridge.Contract.RenounceOwnership(&_BitcoinBridge.TransactOpts)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_BitcoinBridge *BitcoinBridgeTransactorSession) RenounceOwnership() (*types.Transaction, error) {
	return _BitcoinBridge.Contract.RenounceOwnership(&_BitcoinBridge.TransactOpts)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_BitcoinBridge *BitcoinBridgeTransactor) TransferOwnership(opts *bind.TransactOpts, newOwner common.Address) (*types.Transaction, error) {
	return _BitcoinBridge.contract.Transact(opts, "transferOwnership", newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_BitcoinBridge *BitcoinBridgeSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _BitcoinBridge.Contract.TransferOwnership(&_BitcoinBridge.TransactOpts, newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_BitcoinBridge *BitcoinBridgeTransactorSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _BitcoinBridge.Contract.TransferOwnership(&_BitcoinBridge.TransactOpts, newOwner)
}

// UpdateMinTBTCAmount is a paid mutator transaction binding the contract method 0x62fe53e1.
//
// Solidity: function updateMinTBTCAmount(uint256 newMinTBTCAmount) returns()
func (_BitcoinBridge *BitcoinBridgeTransactor) UpdateMinTBTCAmount(opts *bind.TransactOpts, newMinTBTCAmount *big.Int) (*types.Transaction, error) {
	return _BitcoinBridge.contract.Transact(opts, "updateMinTBTCAmount", newMinTBTCAmount)
}

// UpdateMinTBTCAmount is a paid mutator transaction binding the contract method 0x62fe53e1.
//
// Solidity: function updateMinTBTCAmount(uint256 newMinTBTCAmount) returns()
func (_BitcoinBridge *BitcoinBridgeSession) UpdateMinTBTCAmount(newMinTBTCAmount *big.Int) (*types.Transaction, error) {
	return _BitcoinBridge.Contract.UpdateMinTBTCAmount(&_BitcoinBridge.TransactOpts, newMinTBTCAmount)
}

// UpdateMinTBTCAmount is a paid mutator transaction binding the contract method 0x62fe53e1.
//
// Solidity: function updateMinTBTCAmount(uint256 newMinTBTCAmount) returns()
func (_BitcoinBridge *BitcoinBridgeTransactorSession) UpdateMinTBTCAmount(newMinTBTCAmount *big.Int) (*types.Transaction, error) {
	return _BitcoinBridge.Contract.UpdateMinTBTCAmount(&_BitcoinBridge.TransactOpts, newMinTBTCAmount)
}

// BitcoinBridgeAssetsLockedIterator is returned from FilterAssetsLocked and is used to iterate over the raw logs and unpacked data for AssetsLocked events raised by the BitcoinBridge contract.
type BitcoinBridgeAssetsLockedIterator struct {
	Event *BitcoinBridgeAssetsLocked // Event containing the contract specifics and raw log

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
func (it *BitcoinBridgeAssetsLockedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BitcoinBridgeAssetsLocked)
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
		it.Event = new(BitcoinBridgeAssetsLocked)
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
func (it *BitcoinBridgeAssetsLockedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BitcoinBridgeAssetsLockedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BitcoinBridgeAssetsLocked represents a AssetsLocked event raised by the BitcoinBridge contract.
type BitcoinBridgeAssetsLocked struct {
	SequenceNumber *big.Int
	Recipient      common.Address
	TbtcAmount     *big.Int
	Raw            types.Log // Blockchain specific contextual infos
}

// FilterAssetsLocked is a free log retrieval operation binding the contract event 0x3d389641a19dd29654913e26f18e4c2ee9897e70fd01f67ff46e8782b247116a.
//
// Solidity: event AssetsLocked(uint256 indexed sequenceNumber, address indexed recipient, uint256 tbtcAmount)
func (_BitcoinBridge *BitcoinBridgeFilterer) FilterAssetsLocked(opts *bind.FilterOpts, sequenceNumber []*big.Int, recipient []common.Address) (*BitcoinBridgeAssetsLockedIterator, error) {

	var sequenceNumberRule []interface{}
	for _, sequenceNumberItem := range sequenceNumber {
		sequenceNumberRule = append(sequenceNumberRule, sequenceNumberItem)
	}
	var recipientRule []interface{}
	for _, recipientItem := range recipient {
		recipientRule = append(recipientRule, recipientItem)
	}

	logs, sub, err := _BitcoinBridge.contract.FilterLogs(opts, "AssetsLocked", sequenceNumberRule, recipientRule)
	if err != nil {
		return nil, err
	}
	return &BitcoinBridgeAssetsLockedIterator{contract: _BitcoinBridge.contract, event: "AssetsLocked", logs: logs, sub: sub}, nil
}

// WatchAssetsLocked is a free log subscription operation binding the contract event 0x3d389641a19dd29654913e26f18e4c2ee9897e70fd01f67ff46e8782b247116a.
//
// Solidity: event AssetsLocked(uint256 indexed sequenceNumber, address indexed recipient, uint256 tbtcAmount)
func (_BitcoinBridge *BitcoinBridgeFilterer) WatchAssetsLocked(opts *bind.WatchOpts, sink chan<- *BitcoinBridgeAssetsLocked, sequenceNumber []*big.Int, recipient []common.Address) (event.Subscription, error) {

	var sequenceNumberRule []interface{}
	for _, sequenceNumberItem := range sequenceNumber {
		sequenceNumberRule = append(sequenceNumberRule, sequenceNumberItem)
	}
	var recipientRule []interface{}
	for _, recipientItem := range recipient {
		recipientRule = append(recipientRule, recipientItem)
	}

	logs, sub, err := _BitcoinBridge.contract.WatchLogs(opts, "AssetsLocked", sequenceNumberRule, recipientRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BitcoinBridgeAssetsLocked)
				if err := _BitcoinBridge.contract.UnpackLog(event, "AssetsLocked", log); err != nil {
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

// ParseAssetsLocked is a log parse operation binding the contract event 0x3d389641a19dd29654913e26f18e4c2ee9897e70fd01f67ff46e8782b247116a.
//
// Solidity: event AssetsLocked(uint256 indexed sequenceNumber, address indexed recipient, uint256 tbtcAmount)
func (_BitcoinBridge *BitcoinBridgeFilterer) ParseAssetsLocked(log types.Log) (*BitcoinBridgeAssetsLocked, error) {
	event := new(BitcoinBridgeAssetsLocked)
	if err := _BitcoinBridge.contract.UnpackLog(event, "AssetsLocked", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BitcoinBridgeDepositFinalizedIterator is returned from FilterDepositFinalized and is used to iterate over the raw logs and unpacked data for DepositFinalized events raised by the BitcoinBridge contract.
type BitcoinBridgeDepositFinalizedIterator struct {
	Event *BitcoinBridgeDepositFinalized // Event containing the contract specifics and raw log

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
func (it *BitcoinBridgeDepositFinalizedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BitcoinBridgeDepositFinalized)
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
		it.Event = new(BitcoinBridgeDepositFinalized)
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
func (it *BitcoinBridgeDepositFinalizedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BitcoinBridgeDepositFinalizedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BitcoinBridgeDepositFinalized represents a DepositFinalized event raised by the BitcoinBridge contract.
type BitcoinBridgeDepositFinalized struct {
	DepositKey    *big.Int
	InitialAmount *big.Int
	TbtcAmount    *big.Int
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterDepositFinalized is a free log retrieval operation binding the contract event 0x270245f29b0d103ec754b6f21b274417b4eb2513d463bd068eaeacaaf6fa985d.
//
// Solidity: event DepositFinalized(uint256 indexed depositKey, uint256 initialAmount, uint256 tbtcAmount)
func (_BitcoinBridge *BitcoinBridgeFilterer) FilterDepositFinalized(opts *bind.FilterOpts, depositKey []*big.Int) (*BitcoinBridgeDepositFinalizedIterator, error) {

	var depositKeyRule []interface{}
	for _, depositKeyItem := range depositKey {
		depositKeyRule = append(depositKeyRule, depositKeyItem)
	}

	logs, sub, err := _BitcoinBridge.contract.FilterLogs(opts, "DepositFinalized", depositKeyRule)
	if err != nil {
		return nil, err
	}
	return &BitcoinBridgeDepositFinalizedIterator{contract: _BitcoinBridge.contract, event: "DepositFinalized", logs: logs, sub: sub}, nil
}

// WatchDepositFinalized is a free log subscription operation binding the contract event 0x270245f29b0d103ec754b6f21b274417b4eb2513d463bd068eaeacaaf6fa985d.
//
// Solidity: event DepositFinalized(uint256 indexed depositKey, uint256 initialAmount, uint256 tbtcAmount)
func (_BitcoinBridge *BitcoinBridgeFilterer) WatchDepositFinalized(opts *bind.WatchOpts, sink chan<- *BitcoinBridgeDepositFinalized, depositKey []*big.Int) (event.Subscription, error) {

	var depositKeyRule []interface{}
	for _, depositKeyItem := range depositKey {
		depositKeyRule = append(depositKeyRule, depositKeyItem)
	}

	logs, sub, err := _BitcoinBridge.contract.WatchLogs(opts, "DepositFinalized", depositKeyRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BitcoinBridgeDepositFinalized)
				if err := _BitcoinBridge.contract.UnpackLog(event, "DepositFinalized", log); err != nil {
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

// ParseDepositFinalized is a log parse operation binding the contract event 0x270245f29b0d103ec754b6f21b274417b4eb2513d463bd068eaeacaaf6fa985d.
//
// Solidity: event DepositFinalized(uint256 indexed depositKey, uint256 initialAmount, uint256 tbtcAmount)
func (_BitcoinBridge *BitcoinBridgeFilterer) ParseDepositFinalized(log types.Log) (*BitcoinBridgeDepositFinalized, error) {
	event := new(BitcoinBridgeDepositFinalized)
	if err := _BitcoinBridge.contract.UnpackLog(event, "DepositFinalized", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BitcoinBridgeDepositInitializedIterator is returned from FilterDepositInitialized and is used to iterate over the raw logs and unpacked data for DepositInitialized events raised by the BitcoinBridge contract.
type BitcoinBridgeDepositInitializedIterator struct {
	Event *BitcoinBridgeDepositInitialized // Event containing the contract specifics and raw log

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
func (it *BitcoinBridgeDepositInitializedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BitcoinBridgeDepositInitialized)
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
		it.Event = new(BitcoinBridgeDepositInitialized)
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
func (it *BitcoinBridgeDepositInitializedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BitcoinBridgeDepositInitializedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BitcoinBridgeDepositInitialized represents a DepositInitialized event raised by the BitcoinBridge contract.
type BitcoinBridgeDepositInitialized struct {
	DepositKey *big.Int
	Recipient  common.Address
	Raw        types.Log // Blockchain specific contextual infos
}

// FilterDepositInitialized is a free log retrieval operation binding the contract event 0xcfdf802cde659b8a4443f70f5a844fbc968a833202a89e216034308721010890.
//
// Solidity: event DepositInitialized(uint256 indexed depositKey, address indexed recipient)
func (_BitcoinBridge *BitcoinBridgeFilterer) FilterDepositInitialized(opts *bind.FilterOpts, depositKey []*big.Int, recipient []common.Address) (*BitcoinBridgeDepositInitializedIterator, error) {

	var depositKeyRule []interface{}
	for _, depositKeyItem := range depositKey {
		depositKeyRule = append(depositKeyRule, depositKeyItem)
	}
	var recipientRule []interface{}
	for _, recipientItem := range recipient {
		recipientRule = append(recipientRule, recipientItem)
	}

	logs, sub, err := _BitcoinBridge.contract.FilterLogs(opts, "DepositInitialized", depositKeyRule, recipientRule)
	if err != nil {
		return nil, err
	}
	return &BitcoinBridgeDepositInitializedIterator{contract: _BitcoinBridge.contract, event: "DepositInitialized", logs: logs, sub: sub}, nil
}

// WatchDepositInitialized is a free log subscription operation binding the contract event 0xcfdf802cde659b8a4443f70f5a844fbc968a833202a89e216034308721010890.
//
// Solidity: event DepositInitialized(uint256 indexed depositKey, address indexed recipient)
func (_BitcoinBridge *BitcoinBridgeFilterer) WatchDepositInitialized(opts *bind.WatchOpts, sink chan<- *BitcoinBridgeDepositInitialized, depositKey []*big.Int, recipient []common.Address) (event.Subscription, error) {

	var depositKeyRule []interface{}
	for _, depositKeyItem := range depositKey {
		depositKeyRule = append(depositKeyRule, depositKeyItem)
	}
	var recipientRule []interface{}
	for _, recipientItem := range recipient {
		recipientRule = append(recipientRule, recipientItem)
	}

	logs, sub, err := _BitcoinBridge.contract.WatchLogs(opts, "DepositInitialized", depositKeyRule, recipientRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BitcoinBridgeDepositInitialized)
				if err := _BitcoinBridge.contract.UnpackLog(event, "DepositInitialized", log); err != nil {
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

// ParseDepositInitialized is a log parse operation binding the contract event 0xcfdf802cde659b8a4443f70f5a844fbc968a833202a89e216034308721010890.
//
// Solidity: event DepositInitialized(uint256 indexed depositKey, address indexed recipient)
func (_BitcoinBridge *BitcoinBridgeFilterer) ParseDepositInitialized(log types.Log) (*BitcoinBridgeDepositInitialized, error) {
	event := new(BitcoinBridgeDepositInitialized)
	if err := _BitcoinBridge.contract.UnpackLog(event, "DepositInitialized", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BitcoinBridgeInitializedIterator is returned from FilterInitialized and is used to iterate over the raw logs and unpacked data for Initialized events raised by the BitcoinBridge contract.
type BitcoinBridgeInitializedIterator struct {
	Event *BitcoinBridgeInitialized // Event containing the contract specifics and raw log

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
func (it *BitcoinBridgeInitializedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BitcoinBridgeInitialized)
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
		it.Event = new(BitcoinBridgeInitialized)
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
func (it *BitcoinBridgeInitializedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BitcoinBridgeInitializedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BitcoinBridgeInitialized represents a Initialized event raised by the BitcoinBridge contract.
type BitcoinBridgeInitialized struct {
	Version uint64
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterInitialized is a free log retrieval operation binding the contract event 0xc7f505b2f371ae2175ee4913f4499e1f2633a7b5936321eed1cdaeb6115181d2.
//
// Solidity: event Initialized(uint64 version)
func (_BitcoinBridge *BitcoinBridgeFilterer) FilterInitialized(opts *bind.FilterOpts) (*BitcoinBridgeInitializedIterator, error) {

	logs, sub, err := _BitcoinBridge.contract.FilterLogs(opts, "Initialized")
	if err != nil {
		return nil, err
	}
	return &BitcoinBridgeInitializedIterator{contract: _BitcoinBridge.contract, event: "Initialized", logs: logs, sub: sub}, nil
}

// WatchInitialized is a free log subscription operation binding the contract event 0xc7f505b2f371ae2175ee4913f4499e1f2633a7b5936321eed1cdaeb6115181d2.
//
// Solidity: event Initialized(uint64 version)
func (_BitcoinBridge *BitcoinBridgeFilterer) WatchInitialized(opts *bind.WatchOpts, sink chan<- *BitcoinBridgeInitialized) (event.Subscription, error) {

	logs, sub, err := _BitcoinBridge.contract.WatchLogs(opts, "Initialized")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BitcoinBridgeInitialized)
				if err := _BitcoinBridge.contract.UnpackLog(event, "Initialized", log); err != nil {
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
func (_BitcoinBridge *BitcoinBridgeFilterer) ParseInitialized(log types.Log) (*BitcoinBridgeInitialized, error) {
	event := new(BitcoinBridgeInitialized)
	if err := _BitcoinBridge.contract.UnpackLog(event, "Initialized", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BitcoinBridgeMinTBTCAmountUpdatedIterator is returned from FilterMinTBTCAmountUpdated and is used to iterate over the raw logs and unpacked data for MinTBTCAmountUpdated events raised by the BitcoinBridge contract.
type BitcoinBridgeMinTBTCAmountUpdatedIterator struct {
	Event *BitcoinBridgeMinTBTCAmountUpdated // Event containing the contract specifics and raw log

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
func (it *BitcoinBridgeMinTBTCAmountUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BitcoinBridgeMinTBTCAmountUpdated)
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
		it.Event = new(BitcoinBridgeMinTBTCAmountUpdated)
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
func (it *BitcoinBridgeMinTBTCAmountUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BitcoinBridgeMinTBTCAmountUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BitcoinBridgeMinTBTCAmountUpdated represents a MinTBTCAmountUpdated event raised by the BitcoinBridge contract.
type BitcoinBridgeMinTBTCAmountUpdated struct {
	MinTBTCAmount *big.Int
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterMinTBTCAmountUpdated is a free log retrieval operation binding the contract event 0xe64dbc80c2152cea46e3b80ba80f3e8c125114dc79194e9c947b480cfc80e59c.
//
// Solidity: event MinTBTCAmountUpdated(uint256 minTBTCAmount)
func (_BitcoinBridge *BitcoinBridgeFilterer) FilterMinTBTCAmountUpdated(opts *bind.FilterOpts) (*BitcoinBridgeMinTBTCAmountUpdatedIterator, error) {

	logs, sub, err := _BitcoinBridge.contract.FilterLogs(opts, "MinTBTCAmountUpdated")
	if err != nil {
		return nil, err
	}
	return &BitcoinBridgeMinTBTCAmountUpdatedIterator{contract: _BitcoinBridge.contract, event: "MinTBTCAmountUpdated", logs: logs, sub: sub}, nil
}

// WatchMinTBTCAmountUpdated is a free log subscription operation binding the contract event 0xe64dbc80c2152cea46e3b80ba80f3e8c125114dc79194e9c947b480cfc80e59c.
//
// Solidity: event MinTBTCAmountUpdated(uint256 minTBTCAmount)
func (_BitcoinBridge *BitcoinBridgeFilterer) WatchMinTBTCAmountUpdated(opts *bind.WatchOpts, sink chan<- *BitcoinBridgeMinTBTCAmountUpdated) (event.Subscription, error) {

	logs, sub, err := _BitcoinBridge.contract.WatchLogs(opts, "MinTBTCAmountUpdated")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BitcoinBridgeMinTBTCAmountUpdated)
				if err := _BitcoinBridge.contract.UnpackLog(event, "MinTBTCAmountUpdated", log); err != nil {
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

// ParseMinTBTCAmountUpdated is a log parse operation binding the contract event 0xe64dbc80c2152cea46e3b80ba80f3e8c125114dc79194e9c947b480cfc80e59c.
//
// Solidity: event MinTBTCAmountUpdated(uint256 minTBTCAmount)
func (_BitcoinBridge *BitcoinBridgeFilterer) ParseMinTBTCAmountUpdated(log types.Log) (*BitcoinBridgeMinTBTCAmountUpdated, error) {
	event := new(BitcoinBridgeMinTBTCAmountUpdated)
	if err := _BitcoinBridge.contract.UnpackLog(event, "MinTBTCAmountUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BitcoinBridgeOwnershipTransferStartedIterator is returned from FilterOwnershipTransferStarted and is used to iterate over the raw logs and unpacked data for OwnershipTransferStarted events raised by the BitcoinBridge contract.
type BitcoinBridgeOwnershipTransferStartedIterator struct {
	Event *BitcoinBridgeOwnershipTransferStarted // Event containing the contract specifics and raw log

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
func (it *BitcoinBridgeOwnershipTransferStartedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BitcoinBridgeOwnershipTransferStarted)
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
		it.Event = new(BitcoinBridgeOwnershipTransferStarted)
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
func (it *BitcoinBridgeOwnershipTransferStartedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BitcoinBridgeOwnershipTransferStartedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BitcoinBridgeOwnershipTransferStarted represents a OwnershipTransferStarted event raised by the BitcoinBridge contract.
type BitcoinBridgeOwnershipTransferStarted struct {
	PreviousOwner common.Address
	NewOwner      common.Address
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterOwnershipTransferStarted is a free log retrieval operation binding the contract event 0x38d16b8cac22d99fc7c124b9cd0de2d3fa1faef420bfe791d8c362d765e22700.
//
// Solidity: event OwnershipTransferStarted(address indexed previousOwner, address indexed newOwner)
func (_BitcoinBridge *BitcoinBridgeFilterer) FilterOwnershipTransferStarted(opts *bind.FilterOpts, previousOwner []common.Address, newOwner []common.Address) (*BitcoinBridgeOwnershipTransferStartedIterator, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _BitcoinBridge.contract.FilterLogs(opts, "OwnershipTransferStarted", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return &BitcoinBridgeOwnershipTransferStartedIterator{contract: _BitcoinBridge.contract, event: "OwnershipTransferStarted", logs: logs, sub: sub}, nil
}

// WatchOwnershipTransferStarted is a free log subscription operation binding the contract event 0x38d16b8cac22d99fc7c124b9cd0de2d3fa1faef420bfe791d8c362d765e22700.
//
// Solidity: event OwnershipTransferStarted(address indexed previousOwner, address indexed newOwner)
func (_BitcoinBridge *BitcoinBridgeFilterer) WatchOwnershipTransferStarted(opts *bind.WatchOpts, sink chan<- *BitcoinBridgeOwnershipTransferStarted, previousOwner []common.Address, newOwner []common.Address) (event.Subscription, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _BitcoinBridge.contract.WatchLogs(opts, "OwnershipTransferStarted", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BitcoinBridgeOwnershipTransferStarted)
				if err := _BitcoinBridge.contract.UnpackLog(event, "OwnershipTransferStarted", log); err != nil {
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
func (_BitcoinBridge *BitcoinBridgeFilterer) ParseOwnershipTransferStarted(log types.Log) (*BitcoinBridgeOwnershipTransferStarted, error) {
	event := new(BitcoinBridgeOwnershipTransferStarted)
	if err := _BitcoinBridge.contract.UnpackLog(event, "OwnershipTransferStarted", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BitcoinBridgeOwnershipTransferredIterator is returned from FilterOwnershipTransferred and is used to iterate over the raw logs and unpacked data for OwnershipTransferred events raised by the BitcoinBridge contract.
type BitcoinBridgeOwnershipTransferredIterator struct {
	Event *BitcoinBridgeOwnershipTransferred // Event containing the contract specifics and raw log

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
func (it *BitcoinBridgeOwnershipTransferredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BitcoinBridgeOwnershipTransferred)
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
		it.Event = new(BitcoinBridgeOwnershipTransferred)
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
func (it *BitcoinBridgeOwnershipTransferredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BitcoinBridgeOwnershipTransferredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BitcoinBridgeOwnershipTransferred represents a OwnershipTransferred event raised by the BitcoinBridge contract.
type BitcoinBridgeOwnershipTransferred struct {
	PreviousOwner common.Address
	NewOwner      common.Address
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterOwnershipTransferred is a free log retrieval operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_BitcoinBridge *BitcoinBridgeFilterer) FilterOwnershipTransferred(opts *bind.FilterOpts, previousOwner []common.Address, newOwner []common.Address) (*BitcoinBridgeOwnershipTransferredIterator, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _BitcoinBridge.contract.FilterLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return &BitcoinBridgeOwnershipTransferredIterator{contract: _BitcoinBridge.contract, event: "OwnershipTransferred", logs: logs, sub: sub}, nil
}

// WatchOwnershipTransferred is a free log subscription operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_BitcoinBridge *BitcoinBridgeFilterer) WatchOwnershipTransferred(opts *bind.WatchOpts, sink chan<- *BitcoinBridgeOwnershipTransferred, previousOwner []common.Address, newOwner []common.Address) (event.Subscription, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _BitcoinBridge.contract.WatchLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BitcoinBridgeOwnershipTransferred)
				if err := _BitcoinBridge.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
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
func (_BitcoinBridge *BitcoinBridgeFilterer) ParseOwnershipTransferred(log types.Log) (*BitcoinBridgeOwnershipTransferred, error) {
	event := new(BitcoinBridgeOwnershipTransferred)
	if err := _BitcoinBridge.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
