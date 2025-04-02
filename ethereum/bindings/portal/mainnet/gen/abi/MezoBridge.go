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

// MezoBridgeMetaData contains all meta data concerning the MezoBridge contract.
var MezoBridgeMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"target\",\"type\":\"address\"}],\"name\":\"AddressEmptyCode\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"}],\"name\":\"AddressInsufficientBalance\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"AmountBelowMinERC20Amount\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"AmountBelowMinTBTCAmount\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"BTCRecipientIsZeroAddress\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"ERC20RecipientIsZeroAddress\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"ERC20TokenAlreadyEnabled\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"ERC20TokenIsZeroAddress\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"ERC20TokenNotEnabled\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"FailedInnerCall\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"InvalidInitialization\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"MaxERC20TokensReached\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"MinERC20AmountIsZero\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"MinTBTCAmountIsZero\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"NotInitializing\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"}],\"name\":\"OwnableInvalidOwner\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"}],\"name\":\"OwnableUnauthorizedAccount\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"ReentrancyGuardReentrantCall\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"token\",\"type\":\"address\"}],\"name\":\"SafeERC20FailedOperation\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"TBTCTokenIsZeroAddress\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"enumBitcoinBridge.BTCDepositState\",\"name\":\"actualState\",\"type\":\"uint8\"},{\"internalType\":\"enumBitcoinBridge.BTCDepositState\",\"name\":\"expectedState\",\"type\":\"uint8\"}],\"name\":\"UnexpectedBTCDepositState\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"actualExtraData\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"expectedExtraData\",\"type\":\"bytes32\"}],\"name\":\"UnexpectedExtraData\",\"type\":\"error\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"sequenceNumber\",\"type\":\"uint256\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"recipient\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"token\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"AssetsLocked\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"btcDepositKey\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"initialAmount\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"tbtcAmount\",\"type\":\"uint256\"}],\"name\":\"BTCDepositFinalized\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"btcDepositKey\",\"type\":\"uint256\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"recipient\",\"type\":\"address\"}],\"name\":\"BTCDepositInitialized\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"ERC20Token\",\"type\":\"address\"}],\"name\":\"ERC20TokenDisabled\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"ERC20Token\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"minERC20Amount\",\"type\":\"uint256\"}],\"name\":\"ERC20TokenEnabled\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint64\",\"name\":\"version\",\"type\":\"uint64\"}],\"name\":\"Initialized\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"ERC20Token\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"newMinERC20Amount\",\"type\":\"uint256\"}],\"name\":\"MinERC20AmountUpdated\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"minTBTCAmount\",\"type\":\"uint256\"}],\"name\":\"MinTBTCAmountUpdated\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"previousOwner\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"OwnershipTransferStarted\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"previousOwner\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"OwnershipTransferred\",\"type\":\"event\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"name\":\"ERC20Tokens\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"ERC20TokensCount\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"MAX_ERC20_TOKENS\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"SATOSHI_MULTIPLIER\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"acceptOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"bridge\",\"outputs\":[{\"internalType\":\"contractIBridge\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"ERC20Token\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"recipient\",\"type\":\"address\"}],\"name\":\"bridgeERC20\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"recipient\",\"type\":\"address\"}],\"name\":\"bridgeTBTC\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"recipient\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"deadline\",\"type\":\"uint256\"},{\"internalType\":\"uint8\",\"name\":\"v\",\"type\":\"uint8\"},{\"internalType\":\"bytes32\",\"name\":\"r\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"s\",\"type\":\"bytes32\"}],\"name\":\"bridgeTBTCWithPermit\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"btcDeposits\",\"outputs\":[{\"internalType\":\"enumBitcoinBridge.BTCDepositState\",\"name\":\"\",\"type\":\"uint8\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"ERC20Token\",\"type\":\"address\"}],\"name\":\"disableERC20Token\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"ERC20Token\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"minERC20Amount\",\"type\":\"uint256\"}],\"name\":\"enableERC20Token\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"btcDepositKey\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"recipient\",\"type\":\"address\"}],\"name\":\"finalizeBTCBridging\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_tbtcBridge\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"_tbtcVault\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"_tbtcToken\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"_initialSequence\",\"type\":\"uint256\"}],\"name\":\"initialize\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"components\":[{\"internalType\":\"bytes4\",\"name\":\"version\",\"type\":\"bytes4\"},{\"internalType\":\"bytes\",\"name\":\"inputVector\",\"type\":\"bytes\"},{\"internalType\":\"bytes\",\"name\":\"outputVector\",\"type\":\"bytes\"},{\"internalType\":\"bytes4\",\"name\":\"locktime\",\"type\":\"bytes4\"}],\"internalType\":\"structIBridgeTypes.BitcoinTxInfo\",\"name\":\"fundingTx\",\"type\":\"tuple\"},{\"components\":[{\"internalType\":\"uint32\",\"name\":\"fundingOutputIndex\",\"type\":\"uint32\"},{\"internalType\":\"bytes8\",\"name\":\"blindingFactor\",\"type\":\"bytes8\"},{\"internalType\":\"bytes20\",\"name\":\"walletPubKeyHash\",\"type\":\"bytes20\"},{\"internalType\":\"bytes20\",\"name\":\"refundPubKeyHash\",\"type\":\"bytes20\"},{\"internalType\":\"bytes4\",\"name\":\"refundLocktime\",\"type\":\"bytes4\"},{\"internalType\":\"address\",\"name\":\"vault\",\"type\":\"address\"}],\"internalType\":\"structIBridgeTypes.DepositRevealInfo\",\"name\":\"reveal\",\"type\":\"tuple\"},{\"internalType\":\"address\",\"name\":\"recipient\",\"type\":\"address\"}],\"name\":\"initializeBTCBridging\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"minTBTCAmount\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"owner\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"pendingOwner\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"renounceOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"sequence\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"tbtcToken\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"tbtcVault\",\"outputs\":[{\"internalType\":\"contractITBTCVault\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"transferOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"ERC20Token\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"newMinERC20Amount\",\"type\":\"uint256\"}],\"name\":\"updateMinERC20Amount\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"newMinTBTCAmount\",\"type\":\"uint256\"}],\"name\":\"updateMinTBTCAmount\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]",
}

// MezoBridgeABI is the input ABI used to generate the binding from.
// Deprecated: Use MezoBridgeMetaData.ABI instead.
var MezoBridgeABI = MezoBridgeMetaData.ABI

// MezoBridge is an auto generated Go binding around an Ethereum contract.
type MezoBridge struct {
	MezoBridgeCaller     // Read-only binding to the contract
	MezoBridgeTransactor // Write-only binding to the contract
	MezoBridgeFilterer   // Log filterer for contract events
}

// MezoBridgeCaller is an auto generated read-only Go binding around an Ethereum contract.
type MezoBridgeCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// MezoBridgeTransactor is an auto generated write-only Go binding around an Ethereum contract.
type MezoBridgeTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// MezoBridgeFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type MezoBridgeFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// MezoBridgeSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type MezoBridgeSession struct {
	Contract     *MezoBridge       // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// MezoBridgeCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type MezoBridgeCallerSession struct {
	Contract *MezoBridgeCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts     // Call options to use throughout this session
}

// MezoBridgeTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type MezoBridgeTransactorSession struct {
	Contract     *MezoBridgeTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts     // Transaction auth options to use throughout this session
}

// MezoBridgeRaw is an auto generated low-level Go binding around an Ethereum contract.
type MezoBridgeRaw struct {
	Contract *MezoBridge // Generic contract binding to access the raw methods on
}

// MezoBridgeCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type MezoBridgeCallerRaw struct {
	Contract *MezoBridgeCaller // Generic read-only contract binding to access the raw methods on
}

// MezoBridgeTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type MezoBridgeTransactorRaw struct {
	Contract *MezoBridgeTransactor // Generic write-only contract binding to access the raw methods on
}

// NewMezoBridge creates a new instance of MezoBridge, bound to a specific deployed contract.
func NewMezoBridge(address common.Address, backend bind.ContractBackend) (*MezoBridge, error) {
	contract, err := bindMezoBridge(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &MezoBridge{MezoBridgeCaller: MezoBridgeCaller{contract: contract}, MezoBridgeTransactor: MezoBridgeTransactor{contract: contract}, MezoBridgeFilterer: MezoBridgeFilterer{contract: contract}}, nil
}

// NewMezoBridgeCaller creates a new read-only instance of MezoBridge, bound to a specific deployed contract.
func NewMezoBridgeCaller(address common.Address, caller bind.ContractCaller) (*MezoBridgeCaller, error) {
	contract, err := bindMezoBridge(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &MezoBridgeCaller{contract: contract}, nil
}

// NewMezoBridgeTransactor creates a new write-only instance of MezoBridge, bound to a specific deployed contract.
func NewMezoBridgeTransactor(address common.Address, transactor bind.ContractTransactor) (*MezoBridgeTransactor, error) {
	contract, err := bindMezoBridge(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &MezoBridgeTransactor{contract: contract}, nil
}

// NewMezoBridgeFilterer creates a new log filterer instance of MezoBridge, bound to a specific deployed contract.
func NewMezoBridgeFilterer(address common.Address, filterer bind.ContractFilterer) (*MezoBridgeFilterer, error) {
	contract, err := bindMezoBridge(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &MezoBridgeFilterer{contract: contract}, nil
}

// bindMezoBridge binds a generic wrapper to an already deployed contract.
func bindMezoBridge(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := MezoBridgeMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_MezoBridge *MezoBridgeRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _MezoBridge.Contract.MezoBridgeCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_MezoBridge *MezoBridgeRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _MezoBridge.Contract.MezoBridgeTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_MezoBridge *MezoBridgeRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _MezoBridge.Contract.MezoBridgeTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_MezoBridge *MezoBridgeCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _MezoBridge.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_MezoBridge *MezoBridgeTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _MezoBridge.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_MezoBridge *MezoBridgeTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _MezoBridge.Contract.contract.Transact(opts, method, params...)
}

// ERC20Tokens is a free data retrieval call binding the contract method 0xd80687ef.
//
// Solidity: function ERC20Tokens(address ) view returns(uint256)
func (_MezoBridge *MezoBridgeCaller) ERC20Tokens(opts *bind.CallOpts, arg0 common.Address) (*big.Int, error) {
	var out []interface{}
	err := _MezoBridge.contract.Call(opts, &out, "ERC20Tokens", arg0)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// ERC20Tokens is a free data retrieval call binding the contract method 0xd80687ef.
//
// Solidity: function ERC20Tokens(address ) view returns(uint256)
func (_MezoBridge *MezoBridgeSession) ERC20Tokens(arg0 common.Address) (*big.Int, error) {
	return _MezoBridge.Contract.ERC20Tokens(&_MezoBridge.CallOpts, arg0)
}

// ERC20Tokens is a free data retrieval call binding the contract method 0xd80687ef.
//
// Solidity: function ERC20Tokens(address ) view returns(uint256)
func (_MezoBridge *MezoBridgeCallerSession) ERC20Tokens(arg0 common.Address) (*big.Int, error) {
	return _MezoBridge.Contract.ERC20Tokens(&_MezoBridge.CallOpts, arg0)
}

// ERC20TokensCount is a free data retrieval call binding the contract method 0xd252bb2c.
//
// Solidity: function ERC20TokensCount() view returns(uint256)
func (_MezoBridge *MezoBridgeCaller) ERC20TokensCount(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _MezoBridge.contract.Call(opts, &out, "ERC20TokensCount")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// ERC20TokensCount is a free data retrieval call binding the contract method 0xd252bb2c.
//
// Solidity: function ERC20TokensCount() view returns(uint256)
func (_MezoBridge *MezoBridgeSession) ERC20TokensCount() (*big.Int, error) {
	return _MezoBridge.Contract.ERC20TokensCount(&_MezoBridge.CallOpts)
}

// ERC20TokensCount is a free data retrieval call binding the contract method 0xd252bb2c.
//
// Solidity: function ERC20TokensCount() view returns(uint256)
func (_MezoBridge *MezoBridgeCallerSession) ERC20TokensCount() (*big.Int, error) {
	return _MezoBridge.Contract.ERC20TokensCount(&_MezoBridge.CallOpts)
}

// MAXERC20TOKENS is a free data retrieval call binding the contract method 0x5febd8eb.
//
// Solidity: function MAX_ERC20_TOKENS() view returns(uint256)
func (_MezoBridge *MezoBridgeCaller) MAXERC20TOKENS(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _MezoBridge.contract.Call(opts, &out, "MAX_ERC20_TOKENS")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// MAXERC20TOKENS is a free data retrieval call binding the contract method 0x5febd8eb.
//
// Solidity: function MAX_ERC20_TOKENS() view returns(uint256)
func (_MezoBridge *MezoBridgeSession) MAXERC20TOKENS() (*big.Int, error) {
	return _MezoBridge.Contract.MAXERC20TOKENS(&_MezoBridge.CallOpts)
}

// MAXERC20TOKENS is a free data retrieval call binding the contract method 0x5febd8eb.
//
// Solidity: function MAX_ERC20_TOKENS() view returns(uint256)
func (_MezoBridge *MezoBridgeCallerSession) MAXERC20TOKENS() (*big.Int, error) {
	return _MezoBridge.Contract.MAXERC20TOKENS(&_MezoBridge.CallOpts)
}

// SATOSHIMULTIPLIER is a free data retrieval call binding the contract method 0xc7ba0347.
//
// Solidity: function SATOSHI_MULTIPLIER() view returns(uint256)
func (_MezoBridge *MezoBridgeCaller) SATOSHIMULTIPLIER(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _MezoBridge.contract.Call(opts, &out, "SATOSHI_MULTIPLIER")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// SATOSHIMULTIPLIER is a free data retrieval call binding the contract method 0xc7ba0347.
//
// Solidity: function SATOSHI_MULTIPLIER() view returns(uint256)
func (_MezoBridge *MezoBridgeSession) SATOSHIMULTIPLIER() (*big.Int, error) {
	return _MezoBridge.Contract.SATOSHIMULTIPLIER(&_MezoBridge.CallOpts)
}

// SATOSHIMULTIPLIER is a free data retrieval call binding the contract method 0xc7ba0347.
//
// Solidity: function SATOSHI_MULTIPLIER() view returns(uint256)
func (_MezoBridge *MezoBridgeCallerSession) SATOSHIMULTIPLIER() (*big.Int, error) {
	return _MezoBridge.Contract.SATOSHIMULTIPLIER(&_MezoBridge.CallOpts)
}

// Bridge is a free data retrieval call binding the contract method 0xe78cea92.
//
// Solidity: function bridge() view returns(address)
func (_MezoBridge *MezoBridgeCaller) Bridge(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _MezoBridge.contract.Call(opts, &out, "bridge")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Bridge is a free data retrieval call binding the contract method 0xe78cea92.
//
// Solidity: function bridge() view returns(address)
func (_MezoBridge *MezoBridgeSession) Bridge() (common.Address, error) {
	return _MezoBridge.Contract.Bridge(&_MezoBridge.CallOpts)
}

// Bridge is a free data retrieval call binding the contract method 0xe78cea92.
//
// Solidity: function bridge() view returns(address)
func (_MezoBridge *MezoBridgeCallerSession) Bridge() (common.Address, error) {
	return _MezoBridge.Contract.Bridge(&_MezoBridge.CallOpts)
}

// BtcDeposits is a free data retrieval call binding the contract method 0x941b1f94.
//
// Solidity: function btcDeposits(uint256 ) view returns(uint8)
func (_MezoBridge *MezoBridgeCaller) BtcDeposits(opts *bind.CallOpts, arg0 *big.Int) (uint8, error) {
	var out []interface{}
	err := _MezoBridge.contract.Call(opts, &out, "btcDeposits", arg0)

	if err != nil {
		return *new(uint8), err
	}

	out0 := *abi.ConvertType(out[0], new(uint8)).(*uint8)

	return out0, err

}

// BtcDeposits is a free data retrieval call binding the contract method 0x941b1f94.
//
// Solidity: function btcDeposits(uint256 ) view returns(uint8)
func (_MezoBridge *MezoBridgeSession) BtcDeposits(arg0 *big.Int) (uint8, error) {
	return _MezoBridge.Contract.BtcDeposits(&_MezoBridge.CallOpts, arg0)
}

// BtcDeposits is a free data retrieval call binding the contract method 0x941b1f94.
//
// Solidity: function btcDeposits(uint256 ) view returns(uint8)
func (_MezoBridge *MezoBridgeCallerSession) BtcDeposits(arg0 *big.Int) (uint8, error) {
	return _MezoBridge.Contract.BtcDeposits(&_MezoBridge.CallOpts, arg0)
}

// MinTBTCAmount is a free data retrieval call binding the contract method 0xdab1b4bd.
//
// Solidity: function minTBTCAmount() view returns(uint256)
func (_MezoBridge *MezoBridgeCaller) MinTBTCAmount(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _MezoBridge.contract.Call(opts, &out, "minTBTCAmount")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// MinTBTCAmount is a free data retrieval call binding the contract method 0xdab1b4bd.
//
// Solidity: function minTBTCAmount() view returns(uint256)
func (_MezoBridge *MezoBridgeSession) MinTBTCAmount() (*big.Int, error) {
	return _MezoBridge.Contract.MinTBTCAmount(&_MezoBridge.CallOpts)
}

// MinTBTCAmount is a free data retrieval call binding the contract method 0xdab1b4bd.
//
// Solidity: function minTBTCAmount() view returns(uint256)
func (_MezoBridge *MezoBridgeCallerSession) MinTBTCAmount() (*big.Int, error) {
	return _MezoBridge.Contract.MinTBTCAmount(&_MezoBridge.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_MezoBridge *MezoBridgeCaller) Owner(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _MezoBridge.contract.Call(opts, &out, "owner")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_MezoBridge *MezoBridgeSession) Owner() (common.Address, error) {
	return _MezoBridge.Contract.Owner(&_MezoBridge.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_MezoBridge *MezoBridgeCallerSession) Owner() (common.Address, error) {
	return _MezoBridge.Contract.Owner(&_MezoBridge.CallOpts)
}

// PendingOwner is a free data retrieval call binding the contract method 0xe30c3978.
//
// Solidity: function pendingOwner() view returns(address)
func (_MezoBridge *MezoBridgeCaller) PendingOwner(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _MezoBridge.contract.Call(opts, &out, "pendingOwner")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// PendingOwner is a free data retrieval call binding the contract method 0xe30c3978.
//
// Solidity: function pendingOwner() view returns(address)
func (_MezoBridge *MezoBridgeSession) PendingOwner() (common.Address, error) {
	return _MezoBridge.Contract.PendingOwner(&_MezoBridge.CallOpts)
}

// PendingOwner is a free data retrieval call binding the contract method 0xe30c3978.
//
// Solidity: function pendingOwner() view returns(address)
func (_MezoBridge *MezoBridgeCallerSession) PendingOwner() (common.Address, error) {
	return _MezoBridge.Contract.PendingOwner(&_MezoBridge.CallOpts)
}

// Sequence is a free data retrieval call binding the contract method 0x529d15cc.
//
// Solidity: function sequence() view returns(uint256)
func (_MezoBridge *MezoBridgeCaller) Sequence(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _MezoBridge.contract.Call(opts, &out, "sequence")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// Sequence is a free data retrieval call binding the contract method 0x529d15cc.
//
// Solidity: function sequence() view returns(uint256)
func (_MezoBridge *MezoBridgeSession) Sequence() (*big.Int, error) {
	return _MezoBridge.Contract.Sequence(&_MezoBridge.CallOpts)
}

// Sequence is a free data retrieval call binding the contract method 0x529d15cc.
//
// Solidity: function sequence() view returns(uint256)
func (_MezoBridge *MezoBridgeCallerSession) Sequence() (*big.Int, error) {
	return _MezoBridge.Contract.Sequence(&_MezoBridge.CallOpts)
}

// TbtcToken is a free data retrieval call binding the contract method 0xe5d3d714.
//
// Solidity: function tbtcToken() view returns(address)
func (_MezoBridge *MezoBridgeCaller) TbtcToken(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _MezoBridge.contract.Call(opts, &out, "tbtcToken")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// TbtcToken is a free data retrieval call binding the contract method 0xe5d3d714.
//
// Solidity: function tbtcToken() view returns(address)
func (_MezoBridge *MezoBridgeSession) TbtcToken() (common.Address, error) {
	return _MezoBridge.Contract.TbtcToken(&_MezoBridge.CallOpts)
}

// TbtcToken is a free data retrieval call binding the contract method 0xe5d3d714.
//
// Solidity: function tbtcToken() view returns(address)
func (_MezoBridge *MezoBridgeCallerSession) TbtcToken() (common.Address, error) {
	return _MezoBridge.Contract.TbtcToken(&_MezoBridge.CallOpts)
}

// TbtcVault is a free data retrieval call binding the contract method 0x0f36403a.
//
// Solidity: function tbtcVault() view returns(address)
func (_MezoBridge *MezoBridgeCaller) TbtcVault(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _MezoBridge.contract.Call(opts, &out, "tbtcVault")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// TbtcVault is a free data retrieval call binding the contract method 0x0f36403a.
//
// Solidity: function tbtcVault() view returns(address)
func (_MezoBridge *MezoBridgeSession) TbtcVault() (common.Address, error) {
	return _MezoBridge.Contract.TbtcVault(&_MezoBridge.CallOpts)
}

// TbtcVault is a free data retrieval call binding the contract method 0x0f36403a.
//
// Solidity: function tbtcVault() view returns(address)
func (_MezoBridge *MezoBridgeCallerSession) TbtcVault() (common.Address, error) {
	return _MezoBridge.Contract.TbtcVault(&_MezoBridge.CallOpts)
}

// AcceptOwnership is a paid mutator transaction binding the contract method 0x79ba5097.
//
// Solidity: function acceptOwnership() returns()
func (_MezoBridge *MezoBridgeTransactor) AcceptOwnership(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _MezoBridge.contract.Transact(opts, "acceptOwnership")
}

// AcceptOwnership is a paid mutator transaction binding the contract method 0x79ba5097.
//
// Solidity: function acceptOwnership() returns()
func (_MezoBridge *MezoBridgeSession) AcceptOwnership() (*types.Transaction, error) {
	return _MezoBridge.Contract.AcceptOwnership(&_MezoBridge.TransactOpts)
}

// AcceptOwnership is a paid mutator transaction binding the contract method 0x79ba5097.
//
// Solidity: function acceptOwnership() returns()
func (_MezoBridge *MezoBridgeTransactorSession) AcceptOwnership() (*types.Transaction, error) {
	return _MezoBridge.Contract.AcceptOwnership(&_MezoBridge.TransactOpts)
}

// BridgeERC20 is a paid mutator transaction binding the contract method 0x61912174.
//
// Solidity: function bridgeERC20(address ERC20Token, uint256 amount, address recipient) returns()
func (_MezoBridge *MezoBridgeTransactor) BridgeERC20(opts *bind.TransactOpts, ERC20Token common.Address, amount *big.Int, recipient common.Address) (*types.Transaction, error) {
	return _MezoBridge.contract.Transact(opts, "bridgeERC20", ERC20Token, amount, recipient)
}

// BridgeERC20 is a paid mutator transaction binding the contract method 0x61912174.
//
// Solidity: function bridgeERC20(address ERC20Token, uint256 amount, address recipient) returns()
func (_MezoBridge *MezoBridgeSession) BridgeERC20(ERC20Token common.Address, amount *big.Int, recipient common.Address) (*types.Transaction, error) {
	return _MezoBridge.Contract.BridgeERC20(&_MezoBridge.TransactOpts, ERC20Token, amount, recipient)
}

// BridgeERC20 is a paid mutator transaction binding the contract method 0x61912174.
//
// Solidity: function bridgeERC20(address ERC20Token, uint256 amount, address recipient) returns()
func (_MezoBridge *MezoBridgeTransactorSession) BridgeERC20(ERC20Token common.Address, amount *big.Int, recipient common.Address) (*types.Transaction, error) {
	return _MezoBridge.Contract.BridgeERC20(&_MezoBridge.TransactOpts, ERC20Token, amount, recipient)
}

// BridgeTBTC is a paid mutator transaction binding the contract method 0xdf4d4663.
//
// Solidity: function bridgeTBTC(uint256 amount, address recipient) returns()
func (_MezoBridge *MezoBridgeTransactor) BridgeTBTC(opts *bind.TransactOpts, amount *big.Int, recipient common.Address) (*types.Transaction, error) {
	return _MezoBridge.contract.Transact(opts, "bridgeTBTC", amount, recipient)
}

// BridgeTBTC is a paid mutator transaction binding the contract method 0xdf4d4663.
//
// Solidity: function bridgeTBTC(uint256 amount, address recipient) returns()
func (_MezoBridge *MezoBridgeSession) BridgeTBTC(amount *big.Int, recipient common.Address) (*types.Transaction, error) {
	return _MezoBridge.Contract.BridgeTBTC(&_MezoBridge.TransactOpts, amount, recipient)
}

// BridgeTBTC is a paid mutator transaction binding the contract method 0xdf4d4663.
//
// Solidity: function bridgeTBTC(uint256 amount, address recipient) returns()
func (_MezoBridge *MezoBridgeTransactorSession) BridgeTBTC(amount *big.Int, recipient common.Address) (*types.Transaction, error) {
	return _MezoBridge.Contract.BridgeTBTC(&_MezoBridge.TransactOpts, amount, recipient)
}

// BridgeTBTCWithPermit is a paid mutator transaction binding the contract method 0x427f9568.
//
// Solidity: function bridgeTBTCWithPermit(uint256 amount, address recipient, uint256 deadline, uint8 v, bytes32 r, bytes32 s) returns()
func (_MezoBridge *MezoBridgeTransactor) BridgeTBTCWithPermit(opts *bind.TransactOpts, amount *big.Int, recipient common.Address, deadline *big.Int, v uint8, r [32]byte, s [32]byte) (*types.Transaction, error) {
	return _MezoBridge.contract.Transact(opts, "bridgeTBTCWithPermit", amount, recipient, deadline, v, r, s)
}

// BridgeTBTCWithPermit is a paid mutator transaction binding the contract method 0x427f9568.
//
// Solidity: function bridgeTBTCWithPermit(uint256 amount, address recipient, uint256 deadline, uint8 v, bytes32 r, bytes32 s) returns()
func (_MezoBridge *MezoBridgeSession) BridgeTBTCWithPermit(amount *big.Int, recipient common.Address, deadline *big.Int, v uint8, r [32]byte, s [32]byte) (*types.Transaction, error) {
	return _MezoBridge.Contract.BridgeTBTCWithPermit(&_MezoBridge.TransactOpts, amount, recipient, deadline, v, r, s)
}

// BridgeTBTCWithPermit is a paid mutator transaction binding the contract method 0x427f9568.
//
// Solidity: function bridgeTBTCWithPermit(uint256 amount, address recipient, uint256 deadline, uint8 v, bytes32 r, bytes32 s) returns()
func (_MezoBridge *MezoBridgeTransactorSession) BridgeTBTCWithPermit(amount *big.Int, recipient common.Address, deadline *big.Int, v uint8, r [32]byte, s [32]byte) (*types.Transaction, error) {
	return _MezoBridge.Contract.BridgeTBTCWithPermit(&_MezoBridge.TransactOpts, amount, recipient, deadline, v, r, s)
}

// DisableERC20Token is a paid mutator transaction binding the contract method 0x74ca1279.
//
// Solidity: function disableERC20Token(address ERC20Token) returns()
func (_MezoBridge *MezoBridgeTransactor) DisableERC20Token(opts *bind.TransactOpts, ERC20Token common.Address) (*types.Transaction, error) {
	return _MezoBridge.contract.Transact(opts, "disableERC20Token", ERC20Token)
}

// DisableERC20Token is a paid mutator transaction binding the contract method 0x74ca1279.
//
// Solidity: function disableERC20Token(address ERC20Token) returns()
func (_MezoBridge *MezoBridgeSession) DisableERC20Token(ERC20Token common.Address) (*types.Transaction, error) {
	return _MezoBridge.Contract.DisableERC20Token(&_MezoBridge.TransactOpts, ERC20Token)
}

// DisableERC20Token is a paid mutator transaction binding the contract method 0x74ca1279.
//
// Solidity: function disableERC20Token(address ERC20Token) returns()
func (_MezoBridge *MezoBridgeTransactorSession) DisableERC20Token(ERC20Token common.Address) (*types.Transaction, error) {
	return _MezoBridge.Contract.DisableERC20Token(&_MezoBridge.TransactOpts, ERC20Token)
}

// EnableERC20Token is a paid mutator transaction binding the contract method 0x67a68320.
//
// Solidity: function enableERC20Token(address ERC20Token, uint256 minERC20Amount) returns()
func (_MezoBridge *MezoBridgeTransactor) EnableERC20Token(opts *bind.TransactOpts, ERC20Token common.Address, minERC20Amount *big.Int) (*types.Transaction, error) {
	return _MezoBridge.contract.Transact(opts, "enableERC20Token", ERC20Token, minERC20Amount)
}

// EnableERC20Token is a paid mutator transaction binding the contract method 0x67a68320.
//
// Solidity: function enableERC20Token(address ERC20Token, uint256 minERC20Amount) returns()
func (_MezoBridge *MezoBridgeSession) EnableERC20Token(ERC20Token common.Address, minERC20Amount *big.Int) (*types.Transaction, error) {
	return _MezoBridge.Contract.EnableERC20Token(&_MezoBridge.TransactOpts, ERC20Token, minERC20Amount)
}

// EnableERC20Token is a paid mutator transaction binding the contract method 0x67a68320.
//
// Solidity: function enableERC20Token(address ERC20Token, uint256 minERC20Amount) returns()
func (_MezoBridge *MezoBridgeTransactorSession) EnableERC20Token(ERC20Token common.Address, minERC20Amount *big.Int) (*types.Transaction, error) {
	return _MezoBridge.Contract.EnableERC20Token(&_MezoBridge.TransactOpts, ERC20Token, minERC20Amount)
}

// FinalizeBTCBridging is a paid mutator transaction binding the contract method 0x24f90de9.
//
// Solidity: function finalizeBTCBridging(uint256 btcDepositKey, address recipient) returns()
func (_MezoBridge *MezoBridgeTransactor) FinalizeBTCBridging(opts *bind.TransactOpts, btcDepositKey *big.Int, recipient common.Address) (*types.Transaction, error) {
	return _MezoBridge.contract.Transact(opts, "finalizeBTCBridging", btcDepositKey, recipient)
}

// FinalizeBTCBridging is a paid mutator transaction binding the contract method 0x24f90de9.
//
// Solidity: function finalizeBTCBridging(uint256 btcDepositKey, address recipient) returns()
func (_MezoBridge *MezoBridgeSession) FinalizeBTCBridging(btcDepositKey *big.Int, recipient common.Address) (*types.Transaction, error) {
	return _MezoBridge.Contract.FinalizeBTCBridging(&_MezoBridge.TransactOpts, btcDepositKey, recipient)
}

// FinalizeBTCBridging is a paid mutator transaction binding the contract method 0x24f90de9.
//
// Solidity: function finalizeBTCBridging(uint256 btcDepositKey, address recipient) returns()
func (_MezoBridge *MezoBridgeTransactorSession) FinalizeBTCBridging(btcDepositKey *big.Int, recipient common.Address) (*types.Transaction, error) {
	return _MezoBridge.Contract.FinalizeBTCBridging(&_MezoBridge.TransactOpts, btcDepositKey, recipient)
}

// Initialize is a paid mutator transaction binding the contract method 0xcf756fdf.
//
// Solidity: function initialize(address _tbtcBridge, address _tbtcVault, address _tbtcToken, uint256 _initialSequence) returns()
func (_MezoBridge *MezoBridgeTransactor) Initialize(opts *bind.TransactOpts, _tbtcBridge common.Address, _tbtcVault common.Address, _tbtcToken common.Address, _initialSequence *big.Int) (*types.Transaction, error) {
	return _MezoBridge.contract.Transact(opts, "initialize", _tbtcBridge, _tbtcVault, _tbtcToken, _initialSequence)
}

// Initialize is a paid mutator transaction binding the contract method 0xcf756fdf.
//
// Solidity: function initialize(address _tbtcBridge, address _tbtcVault, address _tbtcToken, uint256 _initialSequence) returns()
func (_MezoBridge *MezoBridgeSession) Initialize(_tbtcBridge common.Address, _tbtcVault common.Address, _tbtcToken common.Address, _initialSequence *big.Int) (*types.Transaction, error) {
	return _MezoBridge.Contract.Initialize(&_MezoBridge.TransactOpts, _tbtcBridge, _tbtcVault, _tbtcToken, _initialSequence)
}

// Initialize is a paid mutator transaction binding the contract method 0xcf756fdf.
//
// Solidity: function initialize(address _tbtcBridge, address _tbtcVault, address _tbtcToken, uint256 _initialSequence) returns()
func (_MezoBridge *MezoBridgeTransactorSession) Initialize(_tbtcBridge common.Address, _tbtcVault common.Address, _tbtcToken common.Address, _initialSequence *big.Int) (*types.Transaction, error) {
	return _MezoBridge.Contract.Initialize(&_MezoBridge.TransactOpts, _tbtcBridge, _tbtcVault, _tbtcToken, _initialSequence)
}

// InitializeBTCBridging is a paid mutator transaction binding the contract method 0x6f64aca2.
//
// Solidity: function initializeBTCBridging((bytes4,bytes,bytes,bytes4) fundingTx, (uint32,bytes8,bytes20,bytes20,bytes4,address) reveal, address recipient) returns()
func (_MezoBridge *MezoBridgeTransactor) InitializeBTCBridging(opts *bind.TransactOpts, fundingTx IBridgeTypesBitcoinTxInfo, reveal IBridgeTypesDepositRevealInfo, recipient common.Address) (*types.Transaction, error) {
	return _MezoBridge.contract.Transact(opts, "initializeBTCBridging", fundingTx, reveal, recipient)
}

// InitializeBTCBridging is a paid mutator transaction binding the contract method 0x6f64aca2.
//
// Solidity: function initializeBTCBridging((bytes4,bytes,bytes,bytes4) fundingTx, (uint32,bytes8,bytes20,bytes20,bytes4,address) reveal, address recipient) returns()
func (_MezoBridge *MezoBridgeSession) InitializeBTCBridging(fundingTx IBridgeTypesBitcoinTxInfo, reveal IBridgeTypesDepositRevealInfo, recipient common.Address) (*types.Transaction, error) {
	return _MezoBridge.Contract.InitializeBTCBridging(&_MezoBridge.TransactOpts, fundingTx, reveal, recipient)
}

// InitializeBTCBridging is a paid mutator transaction binding the contract method 0x6f64aca2.
//
// Solidity: function initializeBTCBridging((bytes4,bytes,bytes,bytes4) fundingTx, (uint32,bytes8,bytes20,bytes20,bytes4,address) reveal, address recipient) returns()
func (_MezoBridge *MezoBridgeTransactorSession) InitializeBTCBridging(fundingTx IBridgeTypesBitcoinTxInfo, reveal IBridgeTypesDepositRevealInfo, recipient common.Address) (*types.Transaction, error) {
	return _MezoBridge.Contract.InitializeBTCBridging(&_MezoBridge.TransactOpts, fundingTx, reveal, recipient)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_MezoBridge *MezoBridgeTransactor) RenounceOwnership(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _MezoBridge.contract.Transact(opts, "renounceOwnership")
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_MezoBridge *MezoBridgeSession) RenounceOwnership() (*types.Transaction, error) {
	return _MezoBridge.Contract.RenounceOwnership(&_MezoBridge.TransactOpts)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_MezoBridge *MezoBridgeTransactorSession) RenounceOwnership() (*types.Transaction, error) {
	return _MezoBridge.Contract.RenounceOwnership(&_MezoBridge.TransactOpts)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_MezoBridge *MezoBridgeTransactor) TransferOwnership(opts *bind.TransactOpts, newOwner common.Address) (*types.Transaction, error) {
	return _MezoBridge.contract.Transact(opts, "transferOwnership", newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_MezoBridge *MezoBridgeSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _MezoBridge.Contract.TransferOwnership(&_MezoBridge.TransactOpts, newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_MezoBridge *MezoBridgeTransactorSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _MezoBridge.Contract.TransferOwnership(&_MezoBridge.TransactOpts, newOwner)
}

// UpdateMinERC20Amount is a paid mutator transaction binding the contract method 0x908d272b.
//
// Solidity: function updateMinERC20Amount(address ERC20Token, uint256 newMinERC20Amount) returns()
func (_MezoBridge *MezoBridgeTransactor) UpdateMinERC20Amount(opts *bind.TransactOpts, ERC20Token common.Address, newMinERC20Amount *big.Int) (*types.Transaction, error) {
	return _MezoBridge.contract.Transact(opts, "updateMinERC20Amount", ERC20Token, newMinERC20Amount)
}

// UpdateMinERC20Amount is a paid mutator transaction binding the contract method 0x908d272b.
//
// Solidity: function updateMinERC20Amount(address ERC20Token, uint256 newMinERC20Amount) returns()
func (_MezoBridge *MezoBridgeSession) UpdateMinERC20Amount(ERC20Token common.Address, newMinERC20Amount *big.Int) (*types.Transaction, error) {
	return _MezoBridge.Contract.UpdateMinERC20Amount(&_MezoBridge.TransactOpts, ERC20Token, newMinERC20Amount)
}

// UpdateMinERC20Amount is a paid mutator transaction binding the contract method 0x908d272b.
//
// Solidity: function updateMinERC20Amount(address ERC20Token, uint256 newMinERC20Amount) returns()
func (_MezoBridge *MezoBridgeTransactorSession) UpdateMinERC20Amount(ERC20Token common.Address, newMinERC20Amount *big.Int) (*types.Transaction, error) {
	return _MezoBridge.Contract.UpdateMinERC20Amount(&_MezoBridge.TransactOpts, ERC20Token, newMinERC20Amount)
}

// UpdateMinTBTCAmount is a paid mutator transaction binding the contract method 0x62fe53e1.
//
// Solidity: function updateMinTBTCAmount(uint256 newMinTBTCAmount) returns()
func (_MezoBridge *MezoBridgeTransactor) UpdateMinTBTCAmount(opts *bind.TransactOpts, newMinTBTCAmount *big.Int) (*types.Transaction, error) {
	return _MezoBridge.contract.Transact(opts, "updateMinTBTCAmount", newMinTBTCAmount)
}

// UpdateMinTBTCAmount is a paid mutator transaction binding the contract method 0x62fe53e1.
//
// Solidity: function updateMinTBTCAmount(uint256 newMinTBTCAmount) returns()
func (_MezoBridge *MezoBridgeSession) UpdateMinTBTCAmount(newMinTBTCAmount *big.Int) (*types.Transaction, error) {
	return _MezoBridge.Contract.UpdateMinTBTCAmount(&_MezoBridge.TransactOpts, newMinTBTCAmount)
}

// UpdateMinTBTCAmount is a paid mutator transaction binding the contract method 0x62fe53e1.
//
// Solidity: function updateMinTBTCAmount(uint256 newMinTBTCAmount) returns()
func (_MezoBridge *MezoBridgeTransactorSession) UpdateMinTBTCAmount(newMinTBTCAmount *big.Int) (*types.Transaction, error) {
	return _MezoBridge.Contract.UpdateMinTBTCAmount(&_MezoBridge.TransactOpts, newMinTBTCAmount)
}

// MezoBridgeAssetsLockedIterator is returned from FilterAssetsLocked and is used to iterate over the raw logs and unpacked data for AssetsLocked events raised by the MezoBridge contract.
type MezoBridgeAssetsLockedIterator struct {
	Event *MezoBridgeAssetsLocked // Event containing the contract specifics and raw log

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
func (it *MezoBridgeAssetsLockedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(MezoBridgeAssetsLocked)
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
		it.Event = new(MezoBridgeAssetsLocked)
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
func (it *MezoBridgeAssetsLockedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *MezoBridgeAssetsLockedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// MezoBridgeAssetsLocked represents a AssetsLocked event raised by the MezoBridge contract.
type MezoBridgeAssetsLocked struct {
	SequenceNumber *big.Int
	Recipient      common.Address
	Token          common.Address
	Amount         *big.Int
	Raw            types.Log // Blockchain specific contextual infos
}

// FilterAssetsLocked is a free log retrieval operation binding the contract event 0x75aa5616721471b8ab0c49ce59500cbad2b7ef1ad10e5eb9449c693c0a5c8fd1.
//
// Solidity: event AssetsLocked(uint256 indexed sequenceNumber, address indexed recipient, address indexed token, uint256 amount)
func (_MezoBridge *MezoBridgeFilterer) FilterAssetsLocked(opts *bind.FilterOpts, sequenceNumber []*big.Int, recipient []common.Address, token []common.Address) (*MezoBridgeAssetsLockedIterator, error) {

	var sequenceNumberRule []interface{}
	for _, sequenceNumberItem := range sequenceNumber {
		sequenceNumberRule = append(sequenceNumberRule, sequenceNumberItem)
	}
	var recipientRule []interface{}
	for _, recipientItem := range recipient {
		recipientRule = append(recipientRule, recipientItem)
	}
	var tokenRule []interface{}
	for _, tokenItem := range token {
		tokenRule = append(tokenRule, tokenItem)
	}

	logs, sub, err := _MezoBridge.contract.FilterLogs(opts, "AssetsLocked", sequenceNumberRule, recipientRule, tokenRule)
	if err != nil {
		return nil, err
	}
	return &MezoBridgeAssetsLockedIterator{contract: _MezoBridge.contract, event: "AssetsLocked", logs: logs, sub: sub}, nil
}

// WatchAssetsLocked is a free log subscription operation binding the contract event 0x75aa5616721471b8ab0c49ce59500cbad2b7ef1ad10e5eb9449c693c0a5c8fd1.
//
// Solidity: event AssetsLocked(uint256 indexed sequenceNumber, address indexed recipient, address indexed token, uint256 amount)
func (_MezoBridge *MezoBridgeFilterer) WatchAssetsLocked(opts *bind.WatchOpts, sink chan<- *MezoBridgeAssetsLocked, sequenceNumber []*big.Int, recipient []common.Address, token []common.Address) (event.Subscription, error) {

	var sequenceNumberRule []interface{}
	for _, sequenceNumberItem := range sequenceNumber {
		sequenceNumberRule = append(sequenceNumberRule, sequenceNumberItem)
	}
	var recipientRule []interface{}
	for _, recipientItem := range recipient {
		recipientRule = append(recipientRule, recipientItem)
	}
	var tokenRule []interface{}
	for _, tokenItem := range token {
		tokenRule = append(tokenRule, tokenItem)
	}

	logs, sub, err := _MezoBridge.contract.WatchLogs(opts, "AssetsLocked", sequenceNumberRule, recipientRule, tokenRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(MezoBridgeAssetsLocked)
				if err := _MezoBridge.contract.UnpackLog(event, "AssetsLocked", log); err != nil {
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

// ParseAssetsLocked is a log parse operation binding the contract event 0x75aa5616721471b8ab0c49ce59500cbad2b7ef1ad10e5eb9449c693c0a5c8fd1.
//
// Solidity: event AssetsLocked(uint256 indexed sequenceNumber, address indexed recipient, address indexed token, uint256 amount)
func (_MezoBridge *MezoBridgeFilterer) ParseAssetsLocked(log types.Log) (*MezoBridgeAssetsLocked, error) {
	event := new(MezoBridgeAssetsLocked)
	if err := _MezoBridge.contract.UnpackLog(event, "AssetsLocked", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// MezoBridgeBTCDepositFinalizedIterator is returned from FilterBTCDepositFinalized and is used to iterate over the raw logs and unpacked data for BTCDepositFinalized events raised by the MezoBridge contract.
type MezoBridgeBTCDepositFinalizedIterator struct {
	Event *MezoBridgeBTCDepositFinalized // Event containing the contract specifics and raw log

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
func (it *MezoBridgeBTCDepositFinalizedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(MezoBridgeBTCDepositFinalized)
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
		it.Event = new(MezoBridgeBTCDepositFinalized)
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
func (it *MezoBridgeBTCDepositFinalizedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *MezoBridgeBTCDepositFinalizedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// MezoBridgeBTCDepositFinalized represents a BTCDepositFinalized event raised by the MezoBridge contract.
type MezoBridgeBTCDepositFinalized struct {
	BtcDepositKey *big.Int
	InitialAmount *big.Int
	TbtcAmount    *big.Int
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterBTCDepositFinalized is a free log retrieval operation binding the contract event 0xa81d3c9594b1f3363bfc07d9277c4624e0da8dae3b42d466f1edc0718c62ab53.
//
// Solidity: event BTCDepositFinalized(uint256 indexed btcDepositKey, uint256 initialAmount, uint256 tbtcAmount)
func (_MezoBridge *MezoBridgeFilterer) FilterBTCDepositFinalized(opts *bind.FilterOpts, btcDepositKey []*big.Int) (*MezoBridgeBTCDepositFinalizedIterator, error) {

	var btcDepositKeyRule []interface{}
	for _, btcDepositKeyItem := range btcDepositKey {
		btcDepositKeyRule = append(btcDepositKeyRule, btcDepositKeyItem)
	}

	logs, sub, err := _MezoBridge.contract.FilterLogs(opts, "BTCDepositFinalized", btcDepositKeyRule)
	if err != nil {
		return nil, err
	}
	return &MezoBridgeBTCDepositFinalizedIterator{contract: _MezoBridge.contract, event: "BTCDepositFinalized", logs: logs, sub: sub}, nil
}

// WatchBTCDepositFinalized is a free log subscription operation binding the contract event 0xa81d3c9594b1f3363bfc07d9277c4624e0da8dae3b42d466f1edc0718c62ab53.
//
// Solidity: event BTCDepositFinalized(uint256 indexed btcDepositKey, uint256 initialAmount, uint256 tbtcAmount)
func (_MezoBridge *MezoBridgeFilterer) WatchBTCDepositFinalized(opts *bind.WatchOpts, sink chan<- *MezoBridgeBTCDepositFinalized, btcDepositKey []*big.Int) (event.Subscription, error) {

	var btcDepositKeyRule []interface{}
	for _, btcDepositKeyItem := range btcDepositKey {
		btcDepositKeyRule = append(btcDepositKeyRule, btcDepositKeyItem)
	}

	logs, sub, err := _MezoBridge.contract.WatchLogs(opts, "BTCDepositFinalized", btcDepositKeyRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(MezoBridgeBTCDepositFinalized)
				if err := _MezoBridge.contract.UnpackLog(event, "BTCDepositFinalized", log); err != nil {
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

// ParseBTCDepositFinalized is a log parse operation binding the contract event 0xa81d3c9594b1f3363bfc07d9277c4624e0da8dae3b42d466f1edc0718c62ab53.
//
// Solidity: event BTCDepositFinalized(uint256 indexed btcDepositKey, uint256 initialAmount, uint256 tbtcAmount)
func (_MezoBridge *MezoBridgeFilterer) ParseBTCDepositFinalized(log types.Log) (*MezoBridgeBTCDepositFinalized, error) {
	event := new(MezoBridgeBTCDepositFinalized)
	if err := _MezoBridge.contract.UnpackLog(event, "BTCDepositFinalized", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// MezoBridgeBTCDepositInitializedIterator is returned from FilterBTCDepositInitialized and is used to iterate over the raw logs and unpacked data for BTCDepositInitialized events raised by the MezoBridge contract.
type MezoBridgeBTCDepositInitializedIterator struct {
	Event *MezoBridgeBTCDepositInitialized // Event containing the contract specifics and raw log

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
func (it *MezoBridgeBTCDepositInitializedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(MezoBridgeBTCDepositInitialized)
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
		it.Event = new(MezoBridgeBTCDepositInitialized)
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
func (it *MezoBridgeBTCDepositInitializedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *MezoBridgeBTCDepositInitializedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// MezoBridgeBTCDepositInitialized represents a BTCDepositInitialized event raised by the MezoBridge contract.
type MezoBridgeBTCDepositInitialized struct {
	BtcDepositKey *big.Int
	Recipient     common.Address
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterBTCDepositInitialized is a free log retrieval operation binding the contract event 0x2fbc945bad45e66509bad2bda7b97993796881f9ac2543b827d2aaf69f186923.
//
// Solidity: event BTCDepositInitialized(uint256 indexed btcDepositKey, address indexed recipient)
func (_MezoBridge *MezoBridgeFilterer) FilterBTCDepositInitialized(opts *bind.FilterOpts, btcDepositKey []*big.Int, recipient []common.Address) (*MezoBridgeBTCDepositInitializedIterator, error) {

	var btcDepositKeyRule []interface{}
	for _, btcDepositKeyItem := range btcDepositKey {
		btcDepositKeyRule = append(btcDepositKeyRule, btcDepositKeyItem)
	}
	var recipientRule []interface{}
	for _, recipientItem := range recipient {
		recipientRule = append(recipientRule, recipientItem)
	}

	logs, sub, err := _MezoBridge.contract.FilterLogs(opts, "BTCDepositInitialized", btcDepositKeyRule, recipientRule)
	if err != nil {
		return nil, err
	}
	return &MezoBridgeBTCDepositInitializedIterator{contract: _MezoBridge.contract, event: "BTCDepositInitialized", logs: logs, sub: sub}, nil
}

// WatchBTCDepositInitialized is a free log subscription operation binding the contract event 0x2fbc945bad45e66509bad2bda7b97993796881f9ac2543b827d2aaf69f186923.
//
// Solidity: event BTCDepositInitialized(uint256 indexed btcDepositKey, address indexed recipient)
func (_MezoBridge *MezoBridgeFilterer) WatchBTCDepositInitialized(opts *bind.WatchOpts, sink chan<- *MezoBridgeBTCDepositInitialized, btcDepositKey []*big.Int, recipient []common.Address) (event.Subscription, error) {

	var btcDepositKeyRule []interface{}
	for _, btcDepositKeyItem := range btcDepositKey {
		btcDepositKeyRule = append(btcDepositKeyRule, btcDepositKeyItem)
	}
	var recipientRule []interface{}
	for _, recipientItem := range recipient {
		recipientRule = append(recipientRule, recipientItem)
	}

	logs, sub, err := _MezoBridge.contract.WatchLogs(opts, "BTCDepositInitialized", btcDepositKeyRule, recipientRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(MezoBridgeBTCDepositInitialized)
				if err := _MezoBridge.contract.UnpackLog(event, "BTCDepositInitialized", log); err != nil {
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

// ParseBTCDepositInitialized is a log parse operation binding the contract event 0x2fbc945bad45e66509bad2bda7b97993796881f9ac2543b827d2aaf69f186923.
//
// Solidity: event BTCDepositInitialized(uint256 indexed btcDepositKey, address indexed recipient)
func (_MezoBridge *MezoBridgeFilterer) ParseBTCDepositInitialized(log types.Log) (*MezoBridgeBTCDepositInitialized, error) {
	event := new(MezoBridgeBTCDepositInitialized)
	if err := _MezoBridge.contract.UnpackLog(event, "BTCDepositInitialized", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// MezoBridgeERC20TokenDisabledIterator is returned from FilterERC20TokenDisabled and is used to iterate over the raw logs and unpacked data for ERC20TokenDisabled events raised by the MezoBridge contract.
type MezoBridgeERC20TokenDisabledIterator struct {
	Event *MezoBridgeERC20TokenDisabled // Event containing the contract specifics and raw log

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
func (it *MezoBridgeERC20TokenDisabledIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(MezoBridgeERC20TokenDisabled)
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
		it.Event = new(MezoBridgeERC20TokenDisabled)
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
func (it *MezoBridgeERC20TokenDisabledIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *MezoBridgeERC20TokenDisabledIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// MezoBridgeERC20TokenDisabled represents a ERC20TokenDisabled event raised by the MezoBridge contract.
type MezoBridgeERC20TokenDisabled struct {
	ERC20Token common.Address
	Raw        types.Log // Blockchain specific contextual infos
}

// FilterERC20TokenDisabled is a free log retrieval operation binding the contract event 0x9c4edffd5782d54d432f513a2a7d944aac6f743c7ef4a83d8c6189ba21dd4299.
//
// Solidity: event ERC20TokenDisabled(address indexed ERC20Token)
func (_MezoBridge *MezoBridgeFilterer) FilterERC20TokenDisabled(opts *bind.FilterOpts, ERC20Token []common.Address) (*MezoBridgeERC20TokenDisabledIterator, error) {

	var ERC20TokenRule []interface{}
	for _, ERC20TokenItem := range ERC20Token {
		ERC20TokenRule = append(ERC20TokenRule, ERC20TokenItem)
	}

	logs, sub, err := _MezoBridge.contract.FilterLogs(opts, "ERC20TokenDisabled", ERC20TokenRule)
	if err != nil {
		return nil, err
	}
	return &MezoBridgeERC20TokenDisabledIterator{contract: _MezoBridge.contract, event: "ERC20TokenDisabled", logs: logs, sub: sub}, nil
}

// WatchERC20TokenDisabled is a free log subscription operation binding the contract event 0x9c4edffd5782d54d432f513a2a7d944aac6f743c7ef4a83d8c6189ba21dd4299.
//
// Solidity: event ERC20TokenDisabled(address indexed ERC20Token)
func (_MezoBridge *MezoBridgeFilterer) WatchERC20TokenDisabled(opts *bind.WatchOpts, sink chan<- *MezoBridgeERC20TokenDisabled, ERC20Token []common.Address) (event.Subscription, error) {

	var ERC20TokenRule []interface{}
	for _, ERC20TokenItem := range ERC20Token {
		ERC20TokenRule = append(ERC20TokenRule, ERC20TokenItem)
	}

	logs, sub, err := _MezoBridge.contract.WatchLogs(opts, "ERC20TokenDisabled", ERC20TokenRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(MezoBridgeERC20TokenDisabled)
				if err := _MezoBridge.contract.UnpackLog(event, "ERC20TokenDisabled", log); err != nil {
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

// ParseERC20TokenDisabled is a log parse operation binding the contract event 0x9c4edffd5782d54d432f513a2a7d944aac6f743c7ef4a83d8c6189ba21dd4299.
//
// Solidity: event ERC20TokenDisabled(address indexed ERC20Token)
func (_MezoBridge *MezoBridgeFilterer) ParseERC20TokenDisabled(log types.Log) (*MezoBridgeERC20TokenDisabled, error) {
	event := new(MezoBridgeERC20TokenDisabled)
	if err := _MezoBridge.contract.UnpackLog(event, "ERC20TokenDisabled", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// MezoBridgeERC20TokenEnabledIterator is returned from FilterERC20TokenEnabled and is used to iterate over the raw logs and unpacked data for ERC20TokenEnabled events raised by the MezoBridge contract.
type MezoBridgeERC20TokenEnabledIterator struct {
	Event *MezoBridgeERC20TokenEnabled // Event containing the contract specifics and raw log

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
func (it *MezoBridgeERC20TokenEnabledIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(MezoBridgeERC20TokenEnabled)
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
		it.Event = new(MezoBridgeERC20TokenEnabled)
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
func (it *MezoBridgeERC20TokenEnabledIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *MezoBridgeERC20TokenEnabledIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// MezoBridgeERC20TokenEnabled represents a ERC20TokenEnabled event raised by the MezoBridge contract.
type MezoBridgeERC20TokenEnabled struct {
	ERC20Token     common.Address
	MinERC20Amount *big.Int
	Raw            types.Log // Blockchain specific contextual infos
}

// FilterERC20TokenEnabled is a free log retrieval operation binding the contract event 0xf17d094161c4f2776fc9caa30094c8ebe1b86cd6f2108db5d9f1d46d8f85494c.
//
// Solidity: event ERC20TokenEnabled(address indexed ERC20Token, uint256 minERC20Amount)
func (_MezoBridge *MezoBridgeFilterer) FilterERC20TokenEnabled(opts *bind.FilterOpts, ERC20Token []common.Address) (*MezoBridgeERC20TokenEnabledIterator, error) {

	var ERC20TokenRule []interface{}
	for _, ERC20TokenItem := range ERC20Token {
		ERC20TokenRule = append(ERC20TokenRule, ERC20TokenItem)
	}

	logs, sub, err := _MezoBridge.contract.FilterLogs(opts, "ERC20TokenEnabled", ERC20TokenRule)
	if err != nil {
		return nil, err
	}
	return &MezoBridgeERC20TokenEnabledIterator{contract: _MezoBridge.contract, event: "ERC20TokenEnabled", logs: logs, sub: sub}, nil
}

// WatchERC20TokenEnabled is a free log subscription operation binding the contract event 0xf17d094161c4f2776fc9caa30094c8ebe1b86cd6f2108db5d9f1d46d8f85494c.
//
// Solidity: event ERC20TokenEnabled(address indexed ERC20Token, uint256 minERC20Amount)
func (_MezoBridge *MezoBridgeFilterer) WatchERC20TokenEnabled(opts *bind.WatchOpts, sink chan<- *MezoBridgeERC20TokenEnabled, ERC20Token []common.Address) (event.Subscription, error) {

	var ERC20TokenRule []interface{}
	for _, ERC20TokenItem := range ERC20Token {
		ERC20TokenRule = append(ERC20TokenRule, ERC20TokenItem)
	}

	logs, sub, err := _MezoBridge.contract.WatchLogs(opts, "ERC20TokenEnabled", ERC20TokenRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(MezoBridgeERC20TokenEnabled)
				if err := _MezoBridge.contract.UnpackLog(event, "ERC20TokenEnabled", log); err != nil {
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

// ParseERC20TokenEnabled is a log parse operation binding the contract event 0xf17d094161c4f2776fc9caa30094c8ebe1b86cd6f2108db5d9f1d46d8f85494c.
//
// Solidity: event ERC20TokenEnabled(address indexed ERC20Token, uint256 minERC20Amount)
func (_MezoBridge *MezoBridgeFilterer) ParseERC20TokenEnabled(log types.Log) (*MezoBridgeERC20TokenEnabled, error) {
	event := new(MezoBridgeERC20TokenEnabled)
	if err := _MezoBridge.contract.UnpackLog(event, "ERC20TokenEnabled", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// MezoBridgeInitializedIterator is returned from FilterInitialized and is used to iterate over the raw logs and unpacked data for Initialized events raised by the MezoBridge contract.
type MezoBridgeInitializedIterator struct {
	Event *MezoBridgeInitialized // Event containing the contract specifics and raw log

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
func (it *MezoBridgeInitializedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(MezoBridgeInitialized)
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
		it.Event = new(MezoBridgeInitialized)
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
func (it *MezoBridgeInitializedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *MezoBridgeInitializedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// MezoBridgeInitialized represents a Initialized event raised by the MezoBridge contract.
type MezoBridgeInitialized struct {
	Version uint64
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterInitialized is a free log retrieval operation binding the contract event 0xc7f505b2f371ae2175ee4913f4499e1f2633a7b5936321eed1cdaeb6115181d2.
//
// Solidity: event Initialized(uint64 version)
func (_MezoBridge *MezoBridgeFilterer) FilterInitialized(opts *bind.FilterOpts) (*MezoBridgeInitializedIterator, error) {

	logs, sub, err := _MezoBridge.contract.FilterLogs(opts, "Initialized")
	if err != nil {
		return nil, err
	}
	return &MezoBridgeInitializedIterator{contract: _MezoBridge.contract, event: "Initialized", logs: logs, sub: sub}, nil
}

// WatchInitialized is a free log subscription operation binding the contract event 0xc7f505b2f371ae2175ee4913f4499e1f2633a7b5936321eed1cdaeb6115181d2.
//
// Solidity: event Initialized(uint64 version)
func (_MezoBridge *MezoBridgeFilterer) WatchInitialized(opts *bind.WatchOpts, sink chan<- *MezoBridgeInitialized) (event.Subscription, error) {

	logs, sub, err := _MezoBridge.contract.WatchLogs(opts, "Initialized")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(MezoBridgeInitialized)
				if err := _MezoBridge.contract.UnpackLog(event, "Initialized", log); err != nil {
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
func (_MezoBridge *MezoBridgeFilterer) ParseInitialized(log types.Log) (*MezoBridgeInitialized, error) {
	event := new(MezoBridgeInitialized)
	if err := _MezoBridge.contract.UnpackLog(event, "Initialized", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// MezoBridgeMinERC20AmountUpdatedIterator is returned from FilterMinERC20AmountUpdated and is used to iterate over the raw logs and unpacked data for MinERC20AmountUpdated events raised by the MezoBridge contract.
type MezoBridgeMinERC20AmountUpdatedIterator struct {
	Event *MezoBridgeMinERC20AmountUpdated // Event containing the contract specifics and raw log

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
func (it *MezoBridgeMinERC20AmountUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(MezoBridgeMinERC20AmountUpdated)
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
		it.Event = new(MezoBridgeMinERC20AmountUpdated)
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
func (it *MezoBridgeMinERC20AmountUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *MezoBridgeMinERC20AmountUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// MezoBridgeMinERC20AmountUpdated represents a MinERC20AmountUpdated event raised by the MezoBridge contract.
type MezoBridgeMinERC20AmountUpdated struct {
	ERC20Token        common.Address
	NewMinERC20Amount *big.Int
	Raw               types.Log // Blockchain specific contextual infos
}

// FilterMinERC20AmountUpdated is a free log retrieval operation binding the contract event 0x886950a2d9ce5c7d214261968375335366c8547e3e5eb5e1744c3cb581c4a672.
//
// Solidity: event MinERC20AmountUpdated(address indexed ERC20Token, uint256 newMinERC20Amount)
func (_MezoBridge *MezoBridgeFilterer) FilterMinERC20AmountUpdated(opts *bind.FilterOpts, ERC20Token []common.Address) (*MezoBridgeMinERC20AmountUpdatedIterator, error) {

	var ERC20TokenRule []interface{}
	for _, ERC20TokenItem := range ERC20Token {
		ERC20TokenRule = append(ERC20TokenRule, ERC20TokenItem)
	}

	logs, sub, err := _MezoBridge.contract.FilterLogs(opts, "MinERC20AmountUpdated", ERC20TokenRule)
	if err != nil {
		return nil, err
	}
	return &MezoBridgeMinERC20AmountUpdatedIterator{contract: _MezoBridge.contract, event: "MinERC20AmountUpdated", logs: logs, sub: sub}, nil
}

// WatchMinERC20AmountUpdated is a free log subscription operation binding the contract event 0x886950a2d9ce5c7d214261968375335366c8547e3e5eb5e1744c3cb581c4a672.
//
// Solidity: event MinERC20AmountUpdated(address indexed ERC20Token, uint256 newMinERC20Amount)
func (_MezoBridge *MezoBridgeFilterer) WatchMinERC20AmountUpdated(opts *bind.WatchOpts, sink chan<- *MezoBridgeMinERC20AmountUpdated, ERC20Token []common.Address) (event.Subscription, error) {

	var ERC20TokenRule []interface{}
	for _, ERC20TokenItem := range ERC20Token {
		ERC20TokenRule = append(ERC20TokenRule, ERC20TokenItem)
	}

	logs, sub, err := _MezoBridge.contract.WatchLogs(opts, "MinERC20AmountUpdated", ERC20TokenRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(MezoBridgeMinERC20AmountUpdated)
				if err := _MezoBridge.contract.UnpackLog(event, "MinERC20AmountUpdated", log); err != nil {
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

// ParseMinERC20AmountUpdated is a log parse operation binding the contract event 0x886950a2d9ce5c7d214261968375335366c8547e3e5eb5e1744c3cb581c4a672.
//
// Solidity: event MinERC20AmountUpdated(address indexed ERC20Token, uint256 newMinERC20Amount)
func (_MezoBridge *MezoBridgeFilterer) ParseMinERC20AmountUpdated(log types.Log) (*MezoBridgeMinERC20AmountUpdated, error) {
	event := new(MezoBridgeMinERC20AmountUpdated)
	if err := _MezoBridge.contract.UnpackLog(event, "MinERC20AmountUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// MezoBridgeMinTBTCAmountUpdatedIterator is returned from FilterMinTBTCAmountUpdated and is used to iterate over the raw logs and unpacked data for MinTBTCAmountUpdated events raised by the MezoBridge contract.
type MezoBridgeMinTBTCAmountUpdatedIterator struct {
	Event *MezoBridgeMinTBTCAmountUpdated // Event containing the contract specifics and raw log

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
func (it *MezoBridgeMinTBTCAmountUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(MezoBridgeMinTBTCAmountUpdated)
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
		it.Event = new(MezoBridgeMinTBTCAmountUpdated)
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
func (it *MezoBridgeMinTBTCAmountUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *MezoBridgeMinTBTCAmountUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// MezoBridgeMinTBTCAmountUpdated represents a MinTBTCAmountUpdated event raised by the MezoBridge contract.
type MezoBridgeMinTBTCAmountUpdated struct {
	MinTBTCAmount *big.Int
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterMinTBTCAmountUpdated is a free log retrieval operation binding the contract event 0xe64dbc80c2152cea46e3b80ba80f3e8c125114dc79194e9c947b480cfc80e59c.
//
// Solidity: event MinTBTCAmountUpdated(uint256 minTBTCAmount)
func (_MezoBridge *MezoBridgeFilterer) FilterMinTBTCAmountUpdated(opts *bind.FilterOpts) (*MezoBridgeMinTBTCAmountUpdatedIterator, error) {

	logs, sub, err := _MezoBridge.contract.FilterLogs(opts, "MinTBTCAmountUpdated")
	if err != nil {
		return nil, err
	}
	return &MezoBridgeMinTBTCAmountUpdatedIterator{contract: _MezoBridge.contract, event: "MinTBTCAmountUpdated", logs: logs, sub: sub}, nil
}

// WatchMinTBTCAmountUpdated is a free log subscription operation binding the contract event 0xe64dbc80c2152cea46e3b80ba80f3e8c125114dc79194e9c947b480cfc80e59c.
//
// Solidity: event MinTBTCAmountUpdated(uint256 minTBTCAmount)
func (_MezoBridge *MezoBridgeFilterer) WatchMinTBTCAmountUpdated(opts *bind.WatchOpts, sink chan<- *MezoBridgeMinTBTCAmountUpdated) (event.Subscription, error) {

	logs, sub, err := _MezoBridge.contract.WatchLogs(opts, "MinTBTCAmountUpdated")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(MezoBridgeMinTBTCAmountUpdated)
				if err := _MezoBridge.contract.UnpackLog(event, "MinTBTCAmountUpdated", log); err != nil {
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
func (_MezoBridge *MezoBridgeFilterer) ParseMinTBTCAmountUpdated(log types.Log) (*MezoBridgeMinTBTCAmountUpdated, error) {
	event := new(MezoBridgeMinTBTCAmountUpdated)
	if err := _MezoBridge.contract.UnpackLog(event, "MinTBTCAmountUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// MezoBridgeOwnershipTransferStartedIterator is returned from FilterOwnershipTransferStarted and is used to iterate over the raw logs and unpacked data for OwnershipTransferStarted events raised by the MezoBridge contract.
type MezoBridgeOwnershipTransferStartedIterator struct {
	Event *MezoBridgeOwnershipTransferStarted // Event containing the contract specifics and raw log

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
func (it *MezoBridgeOwnershipTransferStartedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(MezoBridgeOwnershipTransferStarted)
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
		it.Event = new(MezoBridgeOwnershipTransferStarted)
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
func (it *MezoBridgeOwnershipTransferStartedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *MezoBridgeOwnershipTransferStartedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// MezoBridgeOwnershipTransferStarted represents a OwnershipTransferStarted event raised by the MezoBridge contract.
type MezoBridgeOwnershipTransferStarted struct {
	PreviousOwner common.Address
	NewOwner      common.Address
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterOwnershipTransferStarted is a free log retrieval operation binding the contract event 0x38d16b8cac22d99fc7c124b9cd0de2d3fa1faef420bfe791d8c362d765e22700.
//
// Solidity: event OwnershipTransferStarted(address indexed previousOwner, address indexed newOwner)
func (_MezoBridge *MezoBridgeFilterer) FilterOwnershipTransferStarted(opts *bind.FilterOpts, previousOwner []common.Address, newOwner []common.Address) (*MezoBridgeOwnershipTransferStartedIterator, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _MezoBridge.contract.FilterLogs(opts, "OwnershipTransferStarted", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return &MezoBridgeOwnershipTransferStartedIterator{contract: _MezoBridge.contract, event: "OwnershipTransferStarted", logs: logs, sub: sub}, nil
}

// WatchOwnershipTransferStarted is a free log subscription operation binding the contract event 0x38d16b8cac22d99fc7c124b9cd0de2d3fa1faef420bfe791d8c362d765e22700.
//
// Solidity: event OwnershipTransferStarted(address indexed previousOwner, address indexed newOwner)
func (_MezoBridge *MezoBridgeFilterer) WatchOwnershipTransferStarted(opts *bind.WatchOpts, sink chan<- *MezoBridgeOwnershipTransferStarted, previousOwner []common.Address, newOwner []common.Address) (event.Subscription, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _MezoBridge.contract.WatchLogs(opts, "OwnershipTransferStarted", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(MezoBridgeOwnershipTransferStarted)
				if err := _MezoBridge.contract.UnpackLog(event, "OwnershipTransferStarted", log); err != nil {
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
func (_MezoBridge *MezoBridgeFilterer) ParseOwnershipTransferStarted(log types.Log) (*MezoBridgeOwnershipTransferStarted, error) {
	event := new(MezoBridgeOwnershipTransferStarted)
	if err := _MezoBridge.contract.UnpackLog(event, "OwnershipTransferStarted", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// MezoBridgeOwnershipTransferredIterator is returned from FilterOwnershipTransferred and is used to iterate over the raw logs and unpacked data for OwnershipTransferred events raised by the MezoBridge contract.
type MezoBridgeOwnershipTransferredIterator struct {
	Event *MezoBridgeOwnershipTransferred // Event containing the contract specifics and raw log

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
func (it *MezoBridgeOwnershipTransferredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(MezoBridgeOwnershipTransferred)
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
		it.Event = new(MezoBridgeOwnershipTransferred)
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
func (it *MezoBridgeOwnershipTransferredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *MezoBridgeOwnershipTransferredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// MezoBridgeOwnershipTransferred represents a OwnershipTransferred event raised by the MezoBridge contract.
type MezoBridgeOwnershipTransferred struct {
	PreviousOwner common.Address
	NewOwner      common.Address
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterOwnershipTransferred is a free log retrieval operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_MezoBridge *MezoBridgeFilterer) FilterOwnershipTransferred(opts *bind.FilterOpts, previousOwner []common.Address, newOwner []common.Address) (*MezoBridgeOwnershipTransferredIterator, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _MezoBridge.contract.FilterLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return &MezoBridgeOwnershipTransferredIterator{contract: _MezoBridge.contract, event: "OwnershipTransferred", logs: logs, sub: sub}, nil
}

// WatchOwnershipTransferred is a free log subscription operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_MezoBridge *MezoBridgeFilterer) WatchOwnershipTransferred(opts *bind.WatchOpts, sink chan<- *MezoBridgeOwnershipTransferred, previousOwner []common.Address, newOwner []common.Address) (event.Subscription, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _MezoBridge.contract.WatchLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(MezoBridgeOwnershipTransferred)
				if err := _MezoBridge.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
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
func (_MezoBridge *MezoBridgeFilterer) ParseOwnershipTransferred(log types.Log) (*MezoBridgeOwnershipTransferred, error) {
	event := new(MezoBridgeOwnershipTransferred)
	if err := _MezoBridge.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
