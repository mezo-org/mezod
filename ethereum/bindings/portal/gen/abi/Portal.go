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

// PortalDepositInfo is an auto generated low-level Go binding around an user-defined struct.
type PortalDepositInfo struct {
	Balance            *big.Int
	UnlockAt           uint32
	ReceiptMinted      *big.Int
	FeeOwed            *big.Int
	LastFeeIntegral    *big.Int
	TbtcMigrationState uint8
}

// PortalDepositToMigrate is an auto generated low-level Go binding around an user-defined struct.
type PortalDepositToMigrate struct {
	Depositor common.Address
	DepositId *big.Int
}

// PortalSupportedToken is an auto generated low-level Go binding around an user-defined struct.
type PortalSupportedToken struct {
	Token        common.Address
	TokenAbility uint8
}

// PortalMetaData contains all meta data concerning the Portal contract.
var PortalMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"target\",\"type\":\"address\"}],\"name\":\"AddressEmptyCode\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"}],\"name\":\"AddressInsufficientBalance\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"asset\",\"type\":\"address\"}],\"name\":\"AssetNotManagedByLiquidityTreasury\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint32\",\"name\":\"unlockAt\",\"type\":\"uint32\"}],\"name\":\"DepositLocked\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"DepositNotFound\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"FailedInnerCall\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"IncorrectAmount\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"depositor\",\"type\":\"address\"}],\"name\":\"IncorrectDepositor\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"lockPeriod\",\"type\":\"uint256\"}],\"name\":\"IncorrectLockPeriod\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"receiptToken\",\"type\":\"address\"}],\"name\":\"IncorrectReceiptTokenDecimals\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"enumPortal.TokenAbility\",\"name\":\"ability\",\"type\":\"uint8\"}],\"name\":\"IncorrectTokenAbility\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"token\",\"type\":\"address\"}],\"name\":\"IncorrectTokenAddress\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"token\",\"type\":\"address\"},{\"internalType\":\"enumPortal.TokenAbility\",\"name\":\"tokenAbility\",\"type\":\"uint8\"}],\"name\":\"InsufficientTokenAbility\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"InvalidInitialization\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint32\",\"name\":\"lockPeriod\",\"type\":\"uint32\"}],\"name\":\"LockPeriodOutOfRange\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint32\",\"name\":\"lockPeriod\",\"type\":\"uint32\"},{\"internalType\":\"uint32\",\"name\":\"newUnlockAt\",\"type\":\"uint32\"},{\"internalType\":\"uint32\",\"name\":\"existingUnlockAt\",\"type\":\"uint32\"}],\"name\":\"LockPeriodTooShort\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint8\",\"name\":\"annualFee\",\"type\":\"uint8\"}],\"name\":\"MaxAnnualFeeExceeded\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint8\",\"name\":\"mintCap\",\"type\":\"uint8\"}],\"name\":\"MaxReceiptMintCapExceeded\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"NotInitializing\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"}],\"name\":\"OwnableInvalidOwner\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"}],\"name\":\"OwnableUnauthorizedAccount\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"depositAmount\",\"type\":\"uint256\"}],\"name\":\"PartialWithdrawalAmountTooHigh\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"feeOwed\",\"type\":\"uint256\"}],\"name\":\"ReceiptFeeOwed\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"mintLimit\",\"type\":\"uint256\"},{\"internalType\":\"uint96\",\"name\":\"currentlyMinted\",\"type\":\"uint96\"},{\"internalType\":\"uint256\",\"name\":\"feeOwed\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"ReceiptMintLimitExceeded\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"ReceiptMintingDisabled\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"receiptMinted\",\"type\":\"uint256\"}],\"name\":\"ReceiptNotRepaid\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"ReceiptTokenAlreadyInitialized\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint96\",\"name\":\"mintedDebt\",\"type\":\"uint96\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"RepayAmountExceededDebt\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"token\",\"type\":\"address\"}],\"name\":\"SafeERC20FailedOperation\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"sender\",\"type\":\"address\"}],\"name\":\"SenderNotLiquidityTreasury\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"SenderNotTbtcMigrationTreasury\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"TbtcCanNotBeMigrated\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"TbtcMigrationAndLiquidityManagementConflict\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"TbtcMigrationNotAllowed\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"TbtcMigrationNotCompleted\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"TbtcMigrationRequestedErr\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"TbtcTokenAddressAlreadySet\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"TbtcTokenAddressNotSet\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"token\",\"type\":\"address\"},{\"internalType\":\"enumPortal.TokenAbility\",\"name\":\"tokenAbility\",\"type\":\"uint8\"}],\"name\":\"TokenAlreadySupported\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"token\",\"type\":\"address\"}],\"name\":\"TokenNotSupported\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"depositId\",\"type\":\"uint256\"},{\"internalType\":\"enumPortal.TbtcMigrationState\",\"name\":\"currentState\",\"type\":\"uint8\"},{\"internalType\":\"enumPortal.TbtcMigrationState\",\"name\":\"expectedState\",\"type\":\"uint8\"}],\"name\":\"UnexpectedTbtcMigrationState\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"token\",\"type\":\"address\"}],\"name\":\"UnknownTokenDecimals\",\"type\":\"error\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"depositor\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"token\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"depositId\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"Deposited\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"depositor\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"token\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"depositId\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"fee\",\"type\":\"uint256\"}],\"name\":\"FeeCollected\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"depositor\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"token\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"address\",\"name\":\"tbtcToken\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"depositId\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"feeInTbtc\",\"type\":\"uint256\"}],\"name\":\"FeeCollectedTbtcMigrated\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"FundedFromTbtcMigration\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint64\",\"name\":\"version\",\"type\":\"uint64\"}],\"name\":\"Initialized\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"asset\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"bool\",\"name\":\"isManaged\",\"type\":\"bool\"}],\"name\":\"LiquidityTreasuryManagedAssetUpdated\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"previousLiquidityTreasury\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"newLiquidityTreasury\",\"type\":\"address\"}],\"name\":\"LiquidityTreasuryUpdated\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"depositor\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"token\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"depositId\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint32\",\"name\":\"unlockAt\",\"type\":\"uint32\"},{\"indexed\":false,\"internalType\":\"uint32\",\"name\":\"lockPeriod\",\"type\":\"uint32\"}],\"name\":\"Locked\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint32\",\"name\":\"maxLockPeriod\",\"type\":\"uint32\"}],\"name\":\"MaxLockPeriodUpdated\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint32\",\"name\":\"minLockPeriod\",\"type\":\"uint32\"}],\"name\":\"MinLockPeriodUpdated\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"previousOwner\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"OwnershipTransferStarted\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"previousOwner\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"OwnershipTransferred\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"depositor\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"token\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"depositId\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"ReceiptMinted\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"token\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint8\",\"name\":\"annualFee\",\"type\":\"uint8\"},{\"indexed\":false,\"internalType\":\"uint8\",\"name\":\"mintCap\",\"type\":\"uint8\"},{\"indexed\":false,\"internalType\":\"address\",\"name\":\"receiptToken\",\"type\":\"address\"}],\"name\":\"ReceiptParamsUpdated\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"depositor\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"token\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"depositId\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"ReceiptRepaid\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"token\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"enumPortal.TokenAbility\",\"name\":\"tokenAbility\",\"type\":\"uint8\"}],\"name\":\"SupportedTokenAdded\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"token\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"bool\",\"name\":\"isAllowed\",\"type\":\"bool\"}],\"name\":\"TbtcMigrationAllowedUpdated\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"depositor\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"token\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"depositId\",\"type\":\"uint256\"}],\"name\":\"TbtcMigrationCompleted\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"depositor\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"token\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"depositId\",\"type\":\"uint256\"}],\"name\":\"TbtcMigrationRequested\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"depositor\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"token\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"depositId\",\"type\":\"uint256\"}],\"name\":\"TbtcMigrationStarted\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"previousMigrationTreasury\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"newMigrationTreasury\",\"type\":\"address\"}],\"name\":\"TbtcMigrationTreasuryUpdated\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"tbtc\",\"type\":\"address\"}],\"name\":\"TbtcTokenAddressSet\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"depositor\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"token\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"depositId\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"Withdrawn\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"token\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"WithdrawnByLiquidityTreasury\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"token\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"WithdrawnForTbtcMigration\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"depositor\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"token\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"address\",\"name\":\"tbtcToken\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"depositId\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amountInTbtc\",\"type\":\"uint256\"}],\"name\":\"WithdrawnTbtcMigrated\",\"type\":\"event\"},{\"inputs\":[],\"name\":\"acceptOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"components\":[{\"internalType\":\"address\",\"name\":\"token\",\"type\":\"address\"},{\"internalType\":\"enumPortal.TokenAbility\",\"name\":\"tokenAbility\",\"type\":\"uint8\"}],\"internalType\":\"structPortal.SupportedToken\",\"name\":\"supportedToken\",\"type\":\"tuple\"}],\"name\":\"addSupportedToken\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"token\",\"type\":\"address\"},{\"components\":[{\"internalType\":\"address\",\"name\":\"depositor\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"depositId\",\"type\":\"uint256\"}],\"internalType\":\"structPortal.DepositToMigrate[]\",\"name\":\"migratedDeposits\",\"type\":\"tuple[]\"}],\"name\":\"completeTbtcMigration\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"token\",\"type\":\"address\"},{\"internalType\":\"uint96\",\"name\":\"amount\",\"type\":\"uint96\"},{\"internalType\":\"uint32\",\"name\":\"lockPeriod\",\"type\":\"uint32\"}],\"name\":\"deposit\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"depositCount\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"depositOwner\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"token\",\"type\":\"address\"},{\"internalType\":\"uint96\",\"name\":\"amount\",\"type\":\"uint96\"},{\"internalType\":\"uint32\",\"name\":\"lockPeriod\",\"type\":\"uint32\"}],\"name\":\"depositFor\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"deposits\",\"outputs\":[{\"internalType\":\"uint96\",\"name\":\"balance\",\"type\":\"uint96\"},{\"internalType\":\"uint32\",\"name\":\"unlockAt\",\"type\":\"uint32\"},{\"internalType\":\"uint96\",\"name\":\"receiptMinted\",\"type\":\"uint96\"},{\"internalType\":\"uint96\",\"name\":\"feeOwed\",\"type\":\"uint96\"},{\"internalType\":\"uint88\",\"name\":\"lastFeeIntegral\",\"type\":\"uint88\"},{\"internalType\":\"enumPortal.TbtcMigrationState\",\"name\":\"tbtcMigrationState\",\"type\":\"uint8\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"name\":\"feeInfo\",\"outputs\":[{\"internalType\":\"uint96\",\"name\":\"totalMinted\",\"type\":\"uint96\"},{\"internalType\":\"uint32\",\"name\":\"lastFeeUpdateAt\",\"type\":\"uint32\"},{\"internalType\":\"uint88\",\"name\":\"feeIntegral\",\"type\":\"uint88\"},{\"internalType\":\"uint8\",\"name\":\"annualFee\",\"type\":\"uint8\"},{\"internalType\":\"uint8\",\"name\":\"mintCap\",\"type\":\"uint8\"},{\"internalType\":\"address\",\"name\":\"receiptToken\",\"type\":\"address\"},{\"internalType\":\"uint96\",\"name\":\"feeCollected\",\"type\":\"uint96\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"depositor\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"token\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"depositId\",\"type\":\"uint256\"}],\"name\":\"getDeposit\",\"outputs\":[{\"components\":[{\"internalType\":\"uint96\",\"name\":\"balance\",\"type\":\"uint96\"},{\"internalType\":\"uint32\",\"name\":\"unlockAt\",\"type\":\"uint32\"},{\"internalType\":\"uint96\",\"name\":\"receiptMinted\",\"type\":\"uint96\"},{\"internalType\":\"uint96\",\"name\":\"feeOwed\",\"type\":\"uint96\"},{\"internalType\":\"uint88\",\"name\":\"lastFeeIntegral\",\"type\":\"uint88\"},{\"internalType\":\"enumPortal.TbtcMigrationState\",\"name\":\"tbtcMigrationState\",\"type\":\"uint8\"}],\"internalType\":\"structPortal.DepositInfo\",\"name\":\"\",\"type\":\"tuple\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"components\":[{\"internalType\":\"address\",\"name\":\"token\",\"type\":\"address\"},{\"internalType\":\"enumPortal.TokenAbility\",\"name\":\"tokenAbility\",\"type\":\"uint8\"}],\"internalType\":\"structPortal.SupportedToken[]\",\"name\":\"supportedTokens\",\"type\":\"tuple[]\"}],\"name\":\"initialize\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"liquidityTreasury\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"name\":\"liquidityTreasuryManaged\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"token\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"depositId\",\"type\":\"uint256\"},{\"internalType\":\"uint32\",\"name\":\"lockPeriod\",\"type\":\"uint32\"}],\"name\":\"lock\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"maxLockPeriod\",\"outputs\":[{\"internalType\":\"uint32\",\"name\":\"\",\"type\":\"uint32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"minLockPeriod\",\"outputs\":[{\"internalType\":\"uint32\",\"name\":\"\",\"type\":\"uint32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"token\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"depositId\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"mintReceipt\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"owner\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"pendingOwner\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"from\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"token\",\"type\":\"address\"},{\"internalType\":\"bytes\",\"name\":\"data\",\"type\":\"bytes\"}],\"name\":\"receiveApproval\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"renounceOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"token\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"depositId\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"repayReceipt\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"token\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"depositId\",\"type\":\"uint256\"}],\"name\":\"requestTbtcMigration\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"asset\",\"type\":\"address\"},{\"internalType\":\"bool\",\"name\":\"isManaged\",\"type\":\"bool\"}],\"name\":\"setAssetAsLiquidityTreasuryManaged\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"asset\",\"type\":\"address\"},{\"internalType\":\"bool\",\"name\":\"isAllowed\",\"type\":\"bool\"}],\"name\":\"setAssetTbtcMigrationAllowed\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_liquidityTreasury\",\"type\":\"address\"}],\"name\":\"setLiquidityTreasury\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint32\",\"name\":\"_maxLockPeriod\",\"type\":\"uint32\"}],\"name\":\"setMaxLockPeriod\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint32\",\"name\":\"_minLockPeriod\",\"type\":\"uint32\"}],\"name\":\"setMinLockPeriod\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"token\",\"type\":\"address\"},{\"internalType\":\"uint8\",\"name\":\"annualFee\",\"type\":\"uint8\"},{\"internalType\":\"uint8\",\"name\":\"mintCap\",\"type\":\"uint8\"},{\"internalType\":\"address\",\"name\":\"receiptToken\",\"type\":\"address\"}],\"name\":\"setReceiptParams\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_tbtcMigrationTreasury\",\"type\":\"address\"}],\"name\":\"setTbtcMigrationTreasury\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_tbtcToken\",\"type\":\"address\"}],\"name\":\"setTbtcTokenAddress\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"tbtcMigrationTreasury\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"name\":\"tbtcMigrations\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"isAllowed\",\"type\":\"bool\"},{\"internalType\":\"uint96\",\"name\":\"totalMigrating\",\"type\":\"uint96\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"tbtcToken\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"name\":\"tokenAbility\",\"outputs\":[{\"internalType\":\"enumPortal.TokenAbility\",\"name\":\"\",\"type\":\"uint8\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"transferOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"token\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"depositId\",\"type\":\"uint256\"}],\"name\":\"withdraw\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"token\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"withdrawAsLiquidityTreasury\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"token\",\"type\":\"address\"},{\"components\":[{\"internalType\":\"address\",\"name\":\"depositor\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"depositId\",\"type\":\"uint256\"}],\"internalType\":\"structPortal.DepositToMigrate[]\",\"name\":\"depositsToMigrate\",\"type\":\"tuple[]\"}],\"name\":\"withdrawForTbtcMigration\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"token\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"depositId\",\"type\":\"uint256\"},{\"internalType\":\"uint96\",\"name\":\"amount\",\"type\":\"uint96\"}],\"name\":\"withdrawPartially\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]",
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
	parsed, err := PortalMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
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
// Solidity: function deposits(address , address , uint256 ) view returns(uint96 balance, uint32 unlockAt, uint96 receiptMinted, uint96 feeOwed, uint88 lastFeeIntegral, uint8 tbtcMigrationState)
func (_Portal *PortalCaller) Deposits(opts *bind.CallOpts, arg0 common.Address, arg1 common.Address, arg2 *big.Int) (struct {
	Balance            *big.Int
	UnlockAt           uint32
	ReceiptMinted      *big.Int
	FeeOwed            *big.Int
	LastFeeIntegral    *big.Int
	TbtcMigrationState uint8
}, error) {
	var out []interface{}
	err := _Portal.contract.Call(opts, &out, "deposits", arg0, arg1, arg2)

	outstruct := new(struct {
		Balance            *big.Int
		UnlockAt           uint32
		ReceiptMinted      *big.Int
		FeeOwed            *big.Int
		LastFeeIntegral    *big.Int
		TbtcMigrationState uint8
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.Balance = *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)
	outstruct.UnlockAt = *abi.ConvertType(out[1], new(uint32)).(*uint32)
	outstruct.ReceiptMinted = *abi.ConvertType(out[2], new(*big.Int)).(**big.Int)
	outstruct.FeeOwed = *abi.ConvertType(out[3], new(*big.Int)).(**big.Int)
	outstruct.LastFeeIntegral = *abi.ConvertType(out[4], new(*big.Int)).(**big.Int)
	outstruct.TbtcMigrationState = *abi.ConvertType(out[5], new(uint8)).(*uint8)

	return *outstruct, err

}

// Deposits is a free data retrieval call binding the contract method 0x5d93a3fc.
//
// Solidity: function deposits(address , address , uint256 ) view returns(uint96 balance, uint32 unlockAt, uint96 receiptMinted, uint96 feeOwed, uint88 lastFeeIntegral, uint8 tbtcMigrationState)
func (_Portal *PortalSession) Deposits(arg0 common.Address, arg1 common.Address, arg2 *big.Int) (struct {
	Balance            *big.Int
	UnlockAt           uint32
	ReceiptMinted      *big.Int
	FeeOwed            *big.Int
	LastFeeIntegral    *big.Int
	TbtcMigrationState uint8
}, error) {
	return _Portal.Contract.Deposits(&_Portal.CallOpts, arg0, arg1, arg2)
}

// Deposits is a free data retrieval call binding the contract method 0x5d93a3fc.
//
// Solidity: function deposits(address , address , uint256 ) view returns(uint96 balance, uint32 unlockAt, uint96 receiptMinted, uint96 feeOwed, uint88 lastFeeIntegral, uint8 tbtcMigrationState)
func (_Portal *PortalCallerSession) Deposits(arg0 common.Address, arg1 common.Address, arg2 *big.Int) (struct {
	Balance            *big.Int
	UnlockAt           uint32
	ReceiptMinted      *big.Int
	FeeOwed            *big.Int
	LastFeeIntegral    *big.Int
	TbtcMigrationState uint8
}, error) {
	return _Portal.Contract.Deposits(&_Portal.CallOpts, arg0, arg1, arg2)
}

// FeeInfo is a free data retrieval call binding the contract method 0xbfcfa66b.
//
// Solidity: function feeInfo(address ) view returns(uint96 totalMinted, uint32 lastFeeUpdateAt, uint88 feeIntegral, uint8 annualFee, uint8 mintCap, address receiptToken, uint96 feeCollected)
func (_Portal *PortalCaller) FeeInfo(opts *bind.CallOpts, arg0 common.Address) (struct {
	TotalMinted     *big.Int
	LastFeeUpdateAt uint32
	FeeIntegral     *big.Int
	AnnualFee       uint8
	MintCap         uint8
	ReceiptToken    common.Address
	FeeCollected    *big.Int
}, error) {
	var out []interface{}
	err := _Portal.contract.Call(opts, &out, "feeInfo", arg0)

	outstruct := new(struct {
		TotalMinted     *big.Int
		LastFeeUpdateAt uint32
		FeeIntegral     *big.Int
		AnnualFee       uint8
		MintCap         uint8
		ReceiptToken    common.Address
		FeeCollected    *big.Int
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.TotalMinted = *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)
	outstruct.LastFeeUpdateAt = *abi.ConvertType(out[1], new(uint32)).(*uint32)
	outstruct.FeeIntegral = *abi.ConvertType(out[2], new(*big.Int)).(**big.Int)
	outstruct.AnnualFee = *abi.ConvertType(out[3], new(uint8)).(*uint8)
	outstruct.MintCap = *abi.ConvertType(out[4], new(uint8)).(*uint8)
	outstruct.ReceiptToken = *abi.ConvertType(out[5], new(common.Address)).(*common.Address)
	outstruct.FeeCollected = *abi.ConvertType(out[6], new(*big.Int)).(**big.Int)

	return *outstruct, err

}

// FeeInfo is a free data retrieval call binding the contract method 0xbfcfa66b.
//
// Solidity: function feeInfo(address ) view returns(uint96 totalMinted, uint32 lastFeeUpdateAt, uint88 feeIntegral, uint8 annualFee, uint8 mintCap, address receiptToken, uint96 feeCollected)
func (_Portal *PortalSession) FeeInfo(arg0 common.Address) (struct {
	TotalMinted     *big.Int
	LastFeeUpdateAt uint32
	FeeIntegral     *big.Int
	AnnualFee       uint8
	MintCap         uint8
	ReceiptToken    common.Address
	FeeCollected    *big.Int
}, error) {
	return _Portal.Contract.FeeInfo(&_Portal.CallOpts, arg0)
}

// FeeInfo is a free data retrieval call binding the contract method 0xbfcfa66b.
//
// Solidity: function feeInfo(address ) view returns(uint96 totalMinted, uint32 lastFeeUpdateAt, uint88 feeIntegral, uint8 annualFee, uint8 mintCap, address receiptToken, uint96 feeCollected)
func (_Portal *PortalCallerSession) FeeInfo(arg0 common.Address) (struct {
	TotalMinted     *big.Int
	LastFeeUpdateAt uint32
	FeeIntegral     *big.Int
	AnnualFee       uint8
	MintCap         uint8
	ReceiptToken    common.Address
	FeeCollected    *big.Int
}, error) {
	return _Portal.Contract.FeeInfo(&_Portal.CallOpts, arg0)
}

// GetDeposit is a free data retrieval call binding the contract method 0x563170e3.
//
// Solidity: function getDeposit(address depositor, address token, uint256 depositId) view returns((uint96,uint32,uint96,uint96,uint88,uint8))
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
// Solidity: function getDeposit(address depositor, address token, uint256 depositId) view returns((uint96,uint32,uint96,uint96,uint88,uint8))
func (_Portal *PortalSession) GetDeposit(depositor common.Address, token common.Address, depositId *big.Int) (PortalDepositInfo, error) {
	return _Portal.Contract.GetDeposit(&_Portal.CallOpts, depositor, token, depositId)
}

// GetDeposit is a free data retrieval call binding the contract method 0x563170e3.
//
// Solidity: function getDeposit(address depositor, address token, uint256 depositId) view returns((uint96,uint32,uint96,uint96,uint88,uint8))
func (_Portal *PortalCallerSession) GetDeposit(depositor common.Address, token common.Address, depositId *big.Int) (PortalDepositInfo, error) {
	return _Portal.Contract.GetDeposit(&_Portal.CallOpts, depositor, token, depositId)
}

// LiquidityTreasury is a free data retrieval call binding the contract method 0xb918e5f9.
//
// Solidity: function liquidityTreasury() view returns(address)
func (_Portal *PortalCaller) LiquidityTreasury(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _Portal.contract.Call(opts, &out, "liquidityTreasury")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// LiquidityTreasury is a free data retrieval call binding the contract method 0xb918e5f9.
//
// Solidity: function liquidityTreasury() view returns(address)
func (_Portal *PortalSession) LiquidityTreasury() (common.Address, error) {
	return _Portal.Contract.LiquidityTreasury(&_Portal.CallOpts)
}

// LiquidityTreasury is a free data retrieval call binding the contract method 0xb918e5f9.
//
// Solidity: function liquidityTreasury() view returns(address)
func (_Portal *PortalCallerSession) LiquidityTreasury() (common.Address, error) {
	return _Portal.Contract.LiquidityTreasury(&_Portal.CallOpts)
}

// LiquidityTreasuryManaged is a free data retrieval call binding the contract method 0x19808171.
//
// Solidity: function liquidityTreasuryManaged(address ) view returns(bool)
func (_Portal *PortalCaller) LiquidityTreasuryManaged(opts *bind.CallOpts, arg0 common.Address) (bool, error) {
	var out []interface{}
	err := _Portal.contract.Call(opts, &out, "liquidityTreasuryManaged", arg0)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// LiquidityTreasuryManaged is a free data retrieval call binding the contract method 0x19808171.
//
// Solidity: function liquidityTreasuryManaged(address ) view returns(bool)
func (_Portal *PortalSession) LiquidityTreasuryManaged(arg0 common.Address) (bool, error) {
	return _Portal.Contract.LiquidityTreasuryManaged(&_Portal.CallOpts, arg0)
}

// LiquidityTreasuryManaged is a free data retrieval call binding the contract method 0x19808171.
//
// Solidity: function liquidityTreasuryManaged(address ) view returns(bool)
func (_Portal *PortalCallerSession) LiquidityTreasuryManaged(arg0 common.Address) (bool, error) {
	return _Portal.Contract.LiquidityTreasuryManaged(&_Portal.CallOpts, arg0)
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

// TbtcMigrationTreasury is a free data retrieval call binding the contract method 0xa2b7e2dd.
//
// Solidity: function tbtcMigrationTreasury() view returns(address)
func (_Portal *PortalCaller) TbtcMigrationTreasury(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _Portal.contract.Call(opts, &out, "tbtcMigrationTreasury")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// TbtcMigrationTreasury is a free data retrieval call binding the contract method 0xa2b7e2dd.
//
// Solidity: function tbtcMigrationTreasury() view returns(address)
func (_Portal *PortalSession) TbtcMigrationTreasury() (common.Address, error) {
	return _Portal.Contract.TbtcMigrationTreasury(&_Portal.CallOpts)
}

// TbtcMigrationTreasury is a free data retrieval call binding the contract method 0xa2b7e2dd.
//
// Solidity: function tbtcMigrationTreasury() view returns(address)
func (_Portal *PortalCallerSession) TbtcMigrationTreasury() (common.Address, error) {
	return _Portal.Contract.TbtcMigrationTreasury(&_Portal.CallOpts)
}

// TbtcMigrations is a free data retrieval call binding the contract method 0x6889d5d0.
//
// Solidity: function tbtcMigrations(address ) view returns(bool isAllowed, uint96 totalMigrating)
func (_Portal *PortalCaller) TbtcMigrations(opts *bind.CallOpts, arg0 common.Address) (struct {
	IsAllowed      bool
	TotalMigrating *big.Int
}, error) {
	var out []interface{}
	err := _Portal.contract.Call(opts, &out, "tbtcMigrations", arg0)

	outstruct := new(struct {
		IsAllowed      bool
		TotalMigrating *big.Int
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.IsAllowed = *abi.ConvertType(out[0], new(bool)).(*bool)
	outstruct.TotalMigrating = *abi.ConvertType(out[1], new(*big.Int)).(**big.Int)

	return *outstruct, err

}

// TbtcMigrations is a free data retrieval call binding the contract method 0x6889d5d0.
//
// Solidity: function tbtcMigrations(address ) view returns(bool isAllowed, uint96 totalMigrating)
func (_Portal *PortalSession) TbtcMigrations(arg0 common.Address) (struct {
	IsAllowed      bool
	TotalMigrating *big.Int
}, error) {
	return _Portal.Contract.TbtcMigrations(&_Portal.CallOpts, arg0)
}

// TbtcMigrations is a free data retrieval call binding the contract method 0x6889d5d0.
//
// Solidity: function tbtcMigrations(address ) view returns(bool isAllowed, uint96 totalMigrating)
func (_Portal *PortalCallerSession) TbtcMigrations(arg0 common.Address) (struct {
	IsAllowed      bool
	TotalMigrating *big.Int
}, error) {
	return _Portal.Contract.TbtcMigrations(&_Portal.CallOpts, arg0)
}

// TbtcToken is a free data retrieval call binding the contract method 0xe5d3d714.
//
// Solidity: function tbtcToken() view returns(address)
func (_Portal *PortalCaller) TbtcToken(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _Portal.contract.Call(opts, &out, "tbtcToken")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// TbtcToken is a free data retrieval call binding the contract method 0xe5d3d714.
//
// Solidity: function tbtcToken() view returns(address)
func (_Portal *PortalSession) TbtcToken() (common.Address, error) {
	return _Portal.Contract.TbtcToken(&_Portal.CallOpts)
}

// TbtcToken is a free data retrieval call binding the contract method 0xe5d3d714.
//
// Solidity: function tbtcToken() view returns(address)
func (_Portal *PortalCallerSession) TbtcToken() (common.Address, error) {
	return _Portal.Contract.TbtcToken(&_Portal.CallOpts)
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

// CompleteTbtcMigration is a paid mutator transaction binding the contract method 0xc74b552e.
//
// Solidity: function completeTbtcMigration(address token, (address,uint256)[] migratedDeposits) returns()
func (_Portal *PortalTransactor) CompleteTbtcMigration(opts *bind.TransactOpts, token common.Address, migratedDeposits []PortalDepositToMigrate) (*types.Transaction, error) {
	return _Portal.contract.Transact(opts, "completeTbtcMigration", token, migratedDeposits)
}

// CompleteTbtcMigration is a paid mutator transaction binding the contract method 0xc74b552e.
//
// Solidity: function completeTbtcMigration(address token, (address,uint256)[] migratedDeposits) returns()
func (_Portal *PortalSession) CompleteTbtcMigration(token common.Address, migratedDeposits []PortalDepositToMigrate) (*types.Transaction, error) {
	return _Portal.Contract.CompleteTbtcMigration(&_Portal.TransactOpts, token, migratedDeposits)
}

// CompleteTbtcMigration is a paid mutator transaction binding the contract method 0xc74b552e.
//
// Solidity: function completeTbtcMigration(address token, (address,uint256)[] migratedDeposits) returns()
func (_Portal *PortalTransactorSession) CompleteTbtcMigration(token common.Address, migratedDeposits []PortalDepositToMigrate) (*types.Transaction, error) {
	return _Portal.Contract.CompleteTbtcMigration(&_Portal.TransactOpts, token, migratedDeposits)
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

// MintReceipt is a paid mutator transaction binding the contract method 0x5fc7f7c5.
//
// Solidity: function mintReceipt(address token, uint256 depositId, uint256 amount) returns()
func (_Portal *PortalTransactor) MintReceipt(opts *bind.TransactOpts, token common.Address, depositId *big.Int, amount *big.Int) (*types.Transaction, error) {
	return _Portal.contract.Transact(opts, "mintReceipt", token, depositId, amount)
}

// MintReceipt is a paid mutator transaction binding the contract method 0x5fc7f7c5.
//
// Solidity: function mintReceipt(address token, uint256 depositId, uint256 amount) returns()
func (_Portal *PortalSession) MintReceipt(token common.Address, depositId *big.Int, amount *big.Int) (*types.Transaction, error) {
	return _Portal.Contract.MintReceipt(&_Portal.TransactOpts, token, depositId, amount)
}

// MintReceipt is a paid mutator transaction binding the contract method 0x5fc7f7c5.
//
// Solidity: function mintReceipt(address token, uint256 depositId, uint256 amount) returns()
func (_Portal *PortalTransactorSession) MintReceipt(token common.Address, depositId *big.Int, amount *big.Int) (*types.Transaction, error) {
	return _Portal.Contract.MintReceipt(&_Portal.TransactOpts, token, depositId, amount)
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

// RepayReceipt is a paid mutator transaction binding the contract method 0xfa4e00c0.
//
// Solidity: function repayReceipt(address token, uint256 depositId, uint256 amount) returns()
func (_Portal *PortalTransactor) RepayReceipt(opts *bind.TransactOpts, token common.Address, depositId *big.Int, amount *big.Int) (*types.Transaction, error) {
	return _Portal.contract.Transact(opts, "repayReceipt", token, depositId, amount)
}

// RepayReceipt is a paid mutator transaction binding the contract method 0xfa4e00c0.
//
// Solidity: function repayReceipt(address token, uint256 depositId, uint256 amount) returns()
func (_Portal *PortalSession) RepayReceipt(token common.Address, depositId *big.Int, amount *big.Int) (*types.Transaction, error) {
	return _Portal.Contract.RepayReceipt(&_Portal.TransactOpts, token, depositId, amount)
}

// RepayReceipt is a paid mutator transaction binding the contract method 0xfa4e00c0.
//
// Solidity: function repayReceipt(address token, uint256 depositId, uint256 amount) returns()
func (_Portal *PortalTransactorSession) RepayReceipt(token common.Address, depositId *big.Int, amount *big.Int) (*types.Transaction, error) {
	return _Portal.Contract.RepayReceipt(&_Portal.TransactOpts, token, depositId, amount)
}

// RequestTbtcMigration is a paid mutator transaction binding the contract method 0x7b498fc1.
//
// Solidity: function requestTbtcMigration(address token, uint256 depositId) returns()
func (_Portal *PortalTransactor) RequestTbtcMigration(opts *bind.TransactOpts, token common.Address, depositId *big.Int) (*types.Transaction, error) {
	return _Portal.contract.Transact(opts, "requestTbtcMigration", token, depositId)
}

// RequestTbtcMigration is a paid mutator transaction binding the contract method 0x7b498fc1.
//
// Solidity: function requestTbtcMigration(address token, uint256 depositId) returns()
func (_Portal *PortalSession) RequestTbtcMigration(token common.Address, depositId *big.Int) (*types.Transaction, error) {
	return _Portal.Contract.RequestTbtcMigration(&_Portal.TransactOpts, token, depositId)
}

// RequestTbtcMigration is a paid mutator transaction binding the contract method 0x7b498fc1.
//
// Solidity: function requestTbtcMigration(address token, uint256 depositId) returns()
func (_Portal *PortalTransactorSession) RequestTbtcMigration(token common.Address, depositId *big.Int) (*types.Transaction, error) {
	return _Portal.Contract.RequestTbtcMigration(&_Portal.TransactOpts, token, depositId)
}

// SetAssetAsLiquidityTreasuryManaged is a paid mutator transaction binding the contract method 0x584c970b.
//
// Solidity: function setAssetAsLiquidityTreasuryManaged(address asset, bool isManaged) returns()
func (_Portal *PortalTransactor) SetAssetAsLiquidityTreasuryManaged(opts *bind.TransactOpts, asset common.Address, isManaged bool) (*types.Transaction, error) {
	return _Portal.contract.Transact(opts, "setAssetAsLiquidityTreasuryManaged", asset, isManaged)
}

// SetAssetAsLiquidityTreasuryManaged is a paid mutator transaction binding the contract method 0x584c970b.
//
// Solidity: function setAssetAsLiquidityTreasuryManaged(address asset, bool isManaged) returns()
func (_Portal *PortalSession) SetAssetAsLiquidityTreasuryManaged(asset common.Address, isManaged bool) (*types.Transaction, error) {
	return _Portal.Contract.SetAssetAsLiquidityTreasuryManaged(&_Portal.TransactOpts, asset, isManaged)
}

// SetAssetAsLiquidityTreasuryManaged is a paid mutator transaction binding the contract method 0x584c970b.
//
// Solidity: function setAssetAsLiquidityTreasuryManaged(address asset, bool isManaged) returns()
func (_Portal *PortalTransactorSession) SetAssetAsLiquidityTreasuryManaged(asset common.Address, isManaged bool) (*types.Transaction, error) {
	return _Portal.Contract.SetAssetAsLiquidityTreasuryManaged(&_Portal.TransactOpts, asset, isManaged)
}

// SetAssetTbtcMigrationAllowed is a paid mutator transaction binding the contract method 0x639fcbc2.
//
// Solidity: function setAssetTbtcMigrationAllowed(address asset, bool isAllowed) returns()
func (_Portal *PortalTransactor) SetAssetTbtcMigrationAllowed(opts *bind.TransactOpts, asset common.Address, isAllowed bool) (*types.Transaction, error) {
	return _Portal.contract.Transact(opts, "setAssetTbtcMigrationAllowed", asset, isAllowed)
}

// SetAssetTbtcMigrationAllowed is a paid mutator transaction binding the contract method 0x639fcbc2.
//
// Solidity: function setAssetTbtcMigrationAllowed(address asset, bool isAllowed) returns()
func (_Portal *PortalSession) SetAssetTbtcMigrationAllowed(asset common.Address, isAllowed bool) (*types.Transaction, error) {
	return _Portal.Contract.SetAssetTbtcMigrationAllowed(&_Portal.TransactOpts, asset, isAllowed)
}

// SetAssetTbtcMigrationAllowed is a paid mutator transaction binding the contract method 0x639fcbc2.
//
// Solidity: function setAssetTbtcMigrationAllowed(address asset, bool isAllowed) returns()
func (_Portal *PortalTransactorSession) SetAssetTbtcMigrationAllowed(asset common.Address, isAllowed bool) (*types.Transaction, error) {
	return _Portal.Contract.SetAssetTbtcMigrationAllowed(&_Portal.TransactOpts, asset, isAllowed)
}

// SetLiquidityTreasury is a paid mutator transaction binding the contract method 0x25e102a9.
//
// Solidity: function setLiquidityTreasury(address _liquidityTreasury) returns()
func (_Portal *PortalTransactor) SetLiquidityTreasury(opts *bind.TransactOpts, _liquidityTreasury common.Address) (*types.Transaction, error) {
	return _Portal.contract.Transact(opts, "setLiquidityTreasury", _liquidityTreasury)
}

// SetLiquidityTreasury is a paid mutator transaction binding the contract method 0x25e102a9.
//
// Solidity: function setLiquidityTreasury(address _liquidityTreasury) returns()
func (_Portal *PortalSession) SetLiquidityTreasury(_liquidityTreasury common.Address) (*types.Transaction, error) {
	return _Portal.Contract.SetLiquidityTreasury(&_Portal.TransactOpts, _liquidityTreasury)
}

// SetLiquidityTreasury is a paid mutator transaction binding the contract method 0x25e102a9.
//
// Solidity: function setLiquidityTreasury(address _liquidityTreasury) returns()
func (_Portal *PortalTransactorSession) SetLiquidityTreasury(_liquidityTreasury common.Address) (*types.Transaction, error) {
	return _Portal.Contract.SetLiquidityTreasury(&_Portal.TransactOpts, _liquidityTreasury)
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

// SetReceiptParams is a paid mutator transaction binding the contract method 0x012f180e.
//
// Solidity: function setReceiptParams(address token, uint8 annualFee, uint8 mintCap, address receiptToken) returns()
func (_Portal *PortalTransactor) SetReceiptParams(opts *bind.TransactOpts, token common.Address, annualFee uint8, mintCap uint8, receiptToken common.Address) (*types.Transaction, error) {
	return _Portal.contract.Transact(opts, "setReceiptParams", token, annualFee, mintCap, receiptToken)
}

// SetReceiptParams is a paid mutator transaction binding the contract method 0x012f180e.
//
// Solidity: function setReceiptParams(address token, uint8 annualFee, uint8 mintCap, address receiptToken) returns()
func (_Portal *PortalSession) SetReceiptParams(token common.Address, annualFee uint8, mintCap uint8, receiptToken common.Address) (*types.Transaction, error) {
	return _Portal.Contract.SetReceiptParams(&_Portal.TransactOpts, token, annualFee, mintCap, receiptToken)
}

// SetReceiptParams is a paid mutator transaction binding the contract method 0x012f180e.
//
// Solidity: function setReceiptParams(address token, uint8 annualFee, uint8 mintCap, address receiptToken) returns()
func (_Portal *PortalTransactorSession) SetReceiptParams(token common.Address, annualFee uint8, mintCap uint8, receiptToken common.Address) (*types.Transaction, error) {
	return _Portal.Contract.SetReceiptParams(&_Portal.TransactOpts, token, annualFee, mintCap, receiptToken)
}

// SetTbtcMigrationTreasury is a paid mutator transaction binding the contract method 0x3c2b8745.
//
// Solidity: function setTbtcMigrationTreasury(address _tbtcMigrationTreasury) returns()
func (_Portal *PortalTransactor) SetTbtcMigrationTreasury(opts *bind.TransactOpts, _tbtcMigrationTreasury common.Address) (*types.Transaction, error) {
	return _Portal.contract.Transact(opts, "setTbtcMigrationTreasury", _tbtcMigrationTreasury)
}

// SetTbtcMigrationTreasury is a paid mutator transaction binding the contract method 0x3c2b8745.
//
// Solidity: function setTbtcMigrationTreasury(address _tbtcMigrationTreasury) returns()
func (_Portal *PortalSession) SetTbtcMigrationTreasury(_tbtcMigrationTreasury common.Address) (*types.Transaction, error) {
	return _Portal.Contract.SetTbtcMigrationTreasury(&_Portal.TransactOpts, _tbtcMigrationTreasury)
}

// SetTbtcMigrationTreasury is a paid mutator transaction binding the contract method 0x3c2b8745.
//
// Solidity: function setTbtcMigrationTreasury(address _tbtcMigrationTreasury) returns()
func (_Portal *PortalTransactorSession) SetTbtcMigrationTreasury(_tbtcMigrationTreasury common.Address) (*types.Transaction, error) {
	return _Portal.Contract.SetTbtcMigrationTreasury(&_Portal.TransactOpts, _tbtcMigrationTreasury)
}

// SetTbtcTokenAddress is a paid mutator transaction binding the contract method 0xbc06b965.
//
// Solidity: function setTbtcTokenAddress(address _tbtcToken) returns()
func (_Portal *PortalTransactor) SetTbtcTokenAddress(opts *bind.TransactOpts, _tbtcToken common.Address) (*types.Transaction, error) {
	return _Portal.contract.Transact(opts, "setTbtcTokenAddress", _tbtcToken)
}

// SetTbtcTokenAddress is a paid mutator transaction binding the contract method 0xbc06b965.
//
// Solidity: function setTbtcTokenAddress(address _tbtcToken) returns()
func (_Portal *PortalSession) SetTbtcTokenAddress(_tbtcToken common.Address) (*types.Transaction, error) {
	return _Portal.Contract.SetTbtcTokenAddress(&_Portal.TransactOpts, _tbtcToken)
}

// SetTbtcTokenAddress is a paid mutator transaction binding the contract method 0xbc06b965.
//
// Solidity: function setTbtcTokenAddress(address _tbtcToken) returns()
func (_Portal *PortalTransactorSession) SetTbtcTokenAddress(_tbtcToken common.Address) (*types.Transaction, error) {
	return _Portal.Contract.SetTbtcTokenAddress(&_Portal.TransactOpts, _tbtcToken)
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

// Withdraw is a paid mutator transaction binding the contract method 0xf3fef3a3.
//
// Solidity: function withdraw(address token, uint256 depositId) returns()
func (_Portal *PortalTransactor) Withdraw(opts *bind.TransactOpts, token common.Address, depositId *big.Int) (*types.Transaction, error) {
	return _Portal.contract.Transact(opts, "withdraw", token, depositId)
}

// Withdraw is a paid mutator transaction binding the contract method 0xf3fef3a3.
//
// Solidity: function withdraw(address token, uint256 depositId) returns()
func (_Portal *PortalSession) Withdraw(token common.Address, depositId *big.Int) (*types.Transaction, error) {
	return _Portal.Contract.Withdraw(&_Portal.TransactOpts, token, depositId)
}

// Withdraw is a paid mutator transaction binding the contract method 0xf3fef3a3.
//
// Solidity: function withdraw(address token, uint256 depositId) returns()
func (_Portal *PortalTransactorSession) Withdraw(token common.Address, depositId *big.Int) (*types.Transaction, error) {
	return _Portal.Contract.Withdraw(&_Portal.TransactOpts, token, depositId)
}

// WithdrawAsLiquidityTreasury is a paid mutator transaction binding the contract method 0x554e6e21.
//
// Solidity: function withdrawAsLiquidityTreasury(address token, uint256 amount) returns()
func (_Portal *PortalTransactor) WithdrawAsLiquidityTreasury(opts *bind.TransactOpts, token common.Address, amount *big.Int) (*types.Transaction, error) {
	return _Portal.contract.Transact(opts, "withdrawAsLiquidityTreasury", token, amount)
}

// WithdrawAsLiquidityTreasury is a paid mutator transaction binding the contract method 0x554e6e21.
//
// Solidity: function withdrawAsLiquidityTreasury(address token, uint256 amount) returns()
func (_Portal *PortalSession) WithdrawAsLiquidityTreasury(token common.Address, amount *big.Int) (*types.Transaction, error) {
	return _Portal.Contract.WithdrawAsLiquidityTreasury(&_Portal.TransactOpts, token, amount)
}

// WithdrawAsLiquidityTreasury is a paid mutator transaction binding the contract method 0x554e6e21.
//
// Solidity: function withdrawAsLiquidityTreasury(address token, uint256 amount) returns()
func (_Portal *PortalTransactorSession) WithdrawAsLiquidityTreasury(token common.Address, amount *big.Int) (*types.Transaction, error) {
	return _Portal.Contract.WithdrawAsLiquidityTreasury(&_Portal.TransactOpts, token, amount)
}

// WithdrawForTbtcMigration is a paid mutator transaction binding the contract method 0xfe10b9fb.
//
// Solidity: function withdrawForTbtcMigration(address token, (address,uint256)[] depositsToMigrate) returns()
func (_Portal *PortalTransactor) WithdrawForTbtcMigration(opts *bind.TransactOpts, token common.Address, depositsToMigrate []PortalDepositToMigrate) (*types.Transaction, error) {
	return _Portal.contract.Transact(opts, "withdrawForTbtcMigration", token, depositsToMigrate)
}

// WithdrawForTbtcMigration is a paid mutator transaction binding the contract method 0xfe10b9fb.
//
// Solidity: function withdrawForTbtcMigration(address token, (address,uint256)[] depositsToMigrate) returns()
func (_Portal *PortalSession) WithdrawForTbtcMigration(token common.Address, depositsToMigrate []PortalDepositToMigrate) (*types.Transaction, error) {
	return _Portal.Contract.WithdrawForTbtcMigration(&_Portal.TransactOpts, token, depositsToMigrate)
}

// WithdrawForTbtcMigration is a paid mutator transaction binding the contract method 0xfe10b9fb.
//
// Solidity: function withdrawForTbtcMigration(address token, (address,uint256)[] depositsToMigrate) returns()
func (_Portal *PortalTransactorSession) WithdrawForTbtcMigration(token common.Address, depositsToMigrate []PortalDepositToMigrate) (*types.Transaction, error) {
	return _Portal.Contract.WithdrawForTbtcMigration(&_Portal.TransactOpts, token, depositsToMigrate)
}

// WithdrawPartially is a paid mutator transaction binding the contract method 0x1ae41f84.
//
// Solidity: function withdrawPartially(address token, uint256 depositId, uint96 amount) returns()
func (_Portal *PortalTransactor) WithdrawPartially(opts *bind.TransactOpts, token common.Address, depositId *big.Int, amount *big.Int) (*types.Transaction, error) {
	return _Portal.contract.Transact(opts, "withdrawPartially", token, depositId, amount)
}

// WithdrawPartially is a paid mutator transaction binding the contract method 0x1ae41f84.
//
// Solidity: function withdrawPartially(address token, uint256 depositId, uint96 amount) returns()
func (_Portal *PortalSession) WithdrawPartially(token common.Address, depositId *big.Int, amount *big.Int) (*types.Transaction, error) {
	return _Portal.Contract.WithdrawPartially(&_Portal.TransactOpts, token, depositId, amount)
}

// WithdrawPartially is a paid mutator transaction binding the contract method 0x1ae41f84.
//
// Solidity: function withdrawPartially(address token, uint256 depositId, uint96 amount) returns()
func (_Portal *PortalTransactorSession) WithdrawPartially(token common.Address, depositId *big.Int, amount *big.Int) (*types.Transaction, error) {
	return _Portal.Contract.WithdrawPartially(&_Portal.TransactOpts, token, depositId, amount)
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

// PortalFeeCollectedIterator is returned from FilterFeeCollected and is used to iterate over the raw logs and unpacked data for FeeCollected events raised by the Portal contract.
type PortalFeeCollectedIterator struct {
	Event *PortalFeeCollected // Event containing the contract specifics and raw log

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
func (it *PortalFeeCollectedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(PortalFeeCollected)
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
		it.Event = new(PortalFeeCollected)
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
func (it *PortalFeeCollectedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *PortalFeeCollectedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// PortalFeeCollected represents a FeeCollected event raised by the Portal contract.
type PortalFeeCollected struct {
	Depositor common.Address
	Token     common.Address
	DepositId *big.Int
	Fee       *big.Int
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterFeeCollected is a free log retrieval operation binding the contract event 0x205442d60b70af1203d43cab62352c3b69b94f091be32fe683198057282b5c92.
//
// Solidity: event FeeCollected(address indexed depositor, address indexed token, uint256 indexed depositId, uint256 fee)
func (_Portal *PortalFilterer) FilterFeeCollected(opts *bind.FilterOpts, depositor []common.Address, token []common.Address, depositId []*big.Int) (*PortalFeeCollectedIterator, error) {

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

	logs, sub, err := _Portal.contract.FilterLogs(opts, "FeeCollected", depositorRule, tokenRule, depositIdRule)
	if err != nil {
		return nil, err
	}
	return &PortalFeeCollectedIterator{contract: _Portal.contract, event: "FeeCollected", logs: logs, sub: sub}, nil
}

// WatchFeeCollected is a free log subscription operation binding the contract event 0x205442d60b70af1203d43cab62352c3b69b94f091be32fe683198057282b5c92.
//
// Solidity: event FeeCollected(address indexed depositor, address indexed token, uint256 indexed depositId, uint256 fee)
func (_Portal *PortalFilterer) WatchFeeCollected(opts *bind.WatchOpts, sink chan<- *PortalFeeCollected, depositor []common.Address, token []common.Address, depositId []*big.Int) (event.Subscription, error) {

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

	logs, sub, err := _Portal.contract.WatchLogs(opts, "FeeCollected", depositorRule, tokenRule, depositIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(PortalFeeCollected)
				if err := _Portal.contract.UnpackLog(event, "FeeCollected", log); err != nil {
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

// ParseFeeCollected is a log parse operation binding the contract event 0x205442d60b70af1203d43cab62352c3b69b94f091be32fe683198057282b5c92.
//
// Solidity: event FeeCollected(address indexed depositor, address indexed token, uint256 indexed depositId, uint256 fee)
func (_Portal *PortalFilterer) ParseFeeCollected(log types.Log) (*PortalFeeCollected, error) {
	event := new(PortalFeeCollected)
	if err := _Portal.contract.UnpackLog(event, "FeeCollected", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// PortalFeeCollectedTbtcMigratedIterator is returned from FilterFeeCollectedTbtcMigrated and is used to iterate over the raw logs and unpacked data for FeeCollectedTbtcMigrated events raised by the Portal contract.
type PortalFeeCollectedTbtcMigratedIterator struct {
	Event *PortalFeeCollectedTbtcMigrated // Event containing the contract specifics and raw log

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
func (it *PortalFeeCollectedTbtcMigratedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(PortalFeeCollectedTbtcMigrated)
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
		it.Event = new(PortalFeeCollectedTbtcMigrated)
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
func (it *PortalFeeCollectedTbtcMigratedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *PortalFeeCollectedTbtcMigratedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// PortalFeeCollectedTbtcMigrated represents a FeeCollectedTbtcMigrated event raised by the Portal contract.
type PortalFeeCollectedTbtcMigrated struct {
	Depositor common.Address
	Token     common.Address
	TbtcToken common.Address
	DepositId *big.Int
	FeeInTbtc *big.Int
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterFeeCollectedTbtcMigrated is a free log retrieval operation binding the contract event 0x5f82682eb95ce785b4c40b5c57de2b7ae2ca818ac5f1e7ab89300e6142215d8f.
//
// Solidity: event FeeCollectedTbtcMigrated(address indexed depositor, address indexed token, address tbtcToken, uint256 indexed depositId, uint256 feeInTbtc)
func (_Portal *PortalFilterer) FilterFeeCollectedTbtcMigrated(opts *bind.FilterOpts, depositor []common.Address, token []common.Address, depositId []*big.Int) (*PortalFeeCollectedTbtcMigratedIterator, error) {

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

	logs, sub, err := _Portal.contract.FilterLogs(opts, "FeeCollectedTbtcMigrated", depositorRule, tokenRule, depositIdRule)
	if err != nil {
		return nil, err
	}
	return &PortalFeeCollectedTbtcMigratedIterator{contract: _Portal.contract, event: "FeeCollectedTbtcMigrated", logs: logs, sub: sub}, nil
}

// WatchFeeCollectedTbtcMigrated is a free log subscription operation binding the contract event 0x5f82682eb95ce785b4c40b5c57de2b7ae2ca818ac5f1e7ab89300e6142215d8f.
//
// Solidity: event FeeCollectedTbtcMigrated(address indexed depositor, address indexed token, address tbtcToken, uint256 indexed depositId, uint256 feeInTbtc)
func (_Portal *PortalFilterer) WatchFeeCollectedTbtcMigrated(opts *bind.WatchOpts, sink chan<- *PortalFeeCollectedTbtcMigrated, depositor []common.Address, token []common.Address, depositId []*big.Int) (event.Subscription, error) {

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

	logs, sub, err := _Portal.contract.WatchLogs(opts, "FeeCollectedTbtcMigrated", depositorRule, tokenRule, depositIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(PortalFeeCollectedTbtcMigrated)
				if err := _Portal.contract.UnpackLog(event, "FeeCollectedTbtcMigrated", log); err != nil {
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

// ParseFeeCollectedTbtcMigrated is a log parse operation binding the contract event 0x5f82682eb95ce785b4c40b5c57de2b7ae2ca818ac5f1e7ab89300e6142215d8f.
//
// Solidity: event FeeCollectedTbtcMigrated(address indexed depositor, address indexed token, address tbtcToken, uint256 indexed depositId, uint256 feeInTbtc)
func (_Portal *PortalFilterer) ParseFeeCollectedTbtcMigrated(log types.Log) (*PortalFeeCollectedTbtcMigrated, error) {
	event := new(PortalFeeCollectedTbtcMigrated)
	if err := _Portal.contract.UnpackLog(event, "FeeCollectedTbtcMigrated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// PortalFundedFromTbtcMigrationIterator is returned from FilterFundedFromTbtcMigration and is used to iterate over the raw logs and unpacked data for FundedFromTbtcMigration events raised by the Portal contract.
type PortalFundedFromTbtcMigrationIterator struct {
	Event *PortalFundedFromTbtcMigration // Event containing the contract specifics and raw log

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
func (it *PortalFundedFromTbtcMigrationIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(PortalFundedFromTbtcMigration)
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
		it.Event = new(PortalFundedFromTbtcMigration)
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
func (it *PortalFundedFromTbtcMigrationIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *PortalFundedFromTbtcMigrationIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// PortalFundedFromTbtcMigration represents a FundedFromTbtcMigration event raised by the Portal contract.
type PortalFundedFromTbtcMigration struct {
	Amount *big.Int
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterFundedFromTbtcMigration is a free log retrieval operation binding the contract event 0x4c0f021c587c95b1c98d00bd52fef4dc732158bc51f121461f3dc4e41990c563.
//
// Solidity: event FundedFromTbtcMigration(uint256 amount)
func (_Portal *PortalFilterer) FilterFundedFromTbtcMigration(opts *bind.FilterOpts) (*PortalFundedFromTbtcMigrationIterator, error) {

	logs, sub, err := _Portal.contract.FilterLogs(opts, "FundedFromTbtcMigration")
	if err != nil {
		return nil, err
	}
	return &PortalFundedFromTbtcMigrationIterator{contract: _Portal.contract, event: "FundedFromTbtcMigration", logs: logs, sub: sub}, nil
}

// WatchFundedFromTbtcMigration is a free log subscription operation binding the contract event 0x4c0f021c587c95b1c98d00bd52fef4dc732158bc51f121461f3dc4e41990c563.
//
// Solidity: event FundedFromTbtcMigration(uint256 amount)
func (_Portal *PortalFilterer) WatchFundedFromTbtcMigration(opts *bind.WatchOpts, sink chan<- *PortalFundedFromTbtcMigration) (event.Subscription, error) {

	logs, sub, err := _Portal.contract.WatchLogs(opts, "FundedFromTbtcMigration")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(PortalFundedFromTbtcMigration)
				if err := _Portal.contract.UnpackLog(event, "FundedFromTbtcMigration", log); err != nil {
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

// ParseFundedFromTbtcMigration is a log parse operation binding the contract event 0x4c0f021c587c95b1c98d00bd52fef4dc732158bc51f121461f3dc4e41990c563.
//
// Solidity: event FundedFromTbtcMigration(uint256 amount)
func (_Portal *PortalFilterer) ParseFundedFromTbtcMigration(log types.Log) (*PortalFundedFromTbtcMigration, error) {
	event := new(PortalFundedFromTbtcMigration)
	if err := _Portal.contract.UnpackLog(event, "FundedFromTbtcMigration", log); err != nil {
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

// PortalLiquidityTreasuryManagedAssetUpdatedIterator is returned from FilterLiquidityTreasuryManagedAssetUpdated and is used to iterate over the raw logs and unpacked data for LiquidityTreasuryManagedAssetUpdated events raised by the Portal contract.
type PortalLiquidityTreasuryManagedAssetUpdatedIterator struct {
	Event *PortalLiquidityTreasuryManagedAssetUpdated // Event containing the contract specifics and raw log

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
func (it *PortalLiquidityTreasuryManagedAssetUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(PortalLiquidityTreasuryManagedAssetUpdated)
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
		it.Event = new(PortalLiquidityTreasuryManagedAssetUpdated)
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
func (it *PortalLiquidityTreasuryManagedAssetUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *PortalLiquidityTreasuryManagedAssetUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// PortalLiquidityTreasuryManagedAssetUpdated represents a LiquidityTreasuryManagedAssetUpdated event raised by the Portal contract.
type PortalLiquidityTreasuryManagedAssetUpdated struct {
	Asset     common.Address
	IsManaged bool
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterLiquidityTreasuryManagedAssetUpdated is a free log retrieval operation binding the contract event 0x37757c6e0f561c1754a2bc68c5299e01bc49b31193e7928f6a6809920e6811e0.
//
// Solidity: event LiquidityTreasuryManagedAssetUpdated(address indexed asset, bool isManaged)
func (_Portal *PortalFilterer) FilterLiquidityTreasuryManagedAssetUpdated(opts *bind.FilterOpts, asset []common.Address) (*PortalLiquidityTreasuryManagedAssetUpdatedIterator, error) {

	var assetRule []interface{}
	for _, assetItem := range asset {
		assetRule = append(assetRule, assetItem)
	}

	logs, sub, err := _Portal.contract.FilterLogs(opts, "LiquidityTreasuryManagedAssetUpdated", assetRule)
	if err != nil {
		return nil, err
	}
	return &PortalLiquidityTreasuryManagedAssetUpdatedIterator{contract: _Portal.contract, event: "LiquidityTreasuryManagedAssetUpdated", logs: logs, sub: sub}, nil
}

// WatchLiquidityTreasuryManagedAssetUpdated is a free log subscription operation binding the contract event 0x37757c6e0f561c1754a2bc68c5299e01bc49b31193e7928f6a6809920e6811e0.
//
// Solidity: event LiquidityTreasuryManagedAssetUpdated(address indexed asset, bool isManaged)
func (_Portal *PortalFilterer) WatchLiquidityTreasuryManagedAssetUpdated(opts *bind.WatchOpts, sink chan<- *PortalLiquidityTreasuryManagedAssetUpdated, asset []common.Address) (event.Subscription, error) {

	var assetRule []interface{}
	for _, assetItem := range asset {
		assetRule = append(assetRule, assetItem)
	}

	logs, sub, err := _Portal.contract.WatchLogs(opts, "LiquidityTreasuryManagedAssetUpdated", assetRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(PortalLiquidityTreasuryManagedAssetUpdated)
				if err := _Portal.contract.UnpackLog(event, "LiquidityTreasuryManagedAssetUpdated", log); err != nil {
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

// ParseLiquidityTreasuryManagedAssetUpdated is a log parse operation binding the contract event 0x37757c6e0f561c1754a2bc68c5299e01bc49b31193e7928f6a6809920e6811e0.
//
// Solidity: event LiquidityTreasuryManagedAssetUpdated(address indexed asset, bool isManaged)
func (_Portal *PortalFilterer) ParseLiquidityTreasuryManagedAssetUpdated(log types.Log) (*PortalLiquidityTreasuryManagedAssetUpdated, error) {
	event := new(PortalLiquidityTreasuryManagedAssetUpdated)
	if err := _Portal.contract.UnpackLog(event, "LiquidityTreasuryManagedAssetUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// PortalLiquidityTreasuryUpdatedIterator is returned from FilterLiquidityTreasuryUpdated and is used to iterate over the raw logs and unpacked data for LiquidityTreasuryUpdated events raised by the Portal contract.
type PortalLiquidityTreasuryUpdatedIterator struct {
	Event *PortalLiquidityTreasuryUpdated // Event containing the contract specifics and raw log

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
func (it *PortalLiquidityTreasuryUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(PortalLiquidityTreasuryUpdated)
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
		it.Event = new(PortalLiquidityTreasuryUpdated)
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
func (it *PortalLiquidityTreasuryUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *PortalLiquidityTreasuryUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// PortalLiquidityTreasuryUpdated represents a LiquidityTreasuryUpdated event raised by the Portal contract.
type PortalLiquidityTreasuryUpdated struct {
	PreviousLiquidityTreasury common.Address
	NewLiquidityTreasury      common.Address
	Raw                       types.Log // Blockchain specific contextual infos
}

// FilterLiquidityTreasuryUpdated is a free log retrieval operation binding the contract event 0x087168495b2024a05f1e51c26b5abadc7eaa5984c24a419d3563f092693ca1d5.
//
// Solidity: event LiquidityTreasuryUpdated(address indexed previousLiquidityTreasury, address indexed newLiquidityTreasury)
func (_Portal *PortalFilterer) FilterLiquidityTreasuryUpdated(opts *bind.FilterOpts, previousLiquidityTreasury []common.Address, newLiquidityTreasury []common.Address) (*PortalLiquidityTreasuryUpdatedIterator, error) {

	var previousLiquidityTreasuryRule []interface{}
	for _, previousLiquidityTreasuryItem := range previousLiquidityTreasury {
		previousLiquidityTreasuryRule = append(previousLiquidityTreasuryRule, previousLiquidityTreasuryItem)
	}
	var newLiquidityTreasuryRule []interface{}
	for _, newLiquidityTreasuryItem := range newLiquidityTreasury {
		newLiquidityTreasuryRule = append(newLiquidityTreasuryRule, newLiquidityTreasuryItem)
	}

	logs, sub, err := _Portal.contract.FilterLogs(opts, "LiquidityTreasuryUpdated", previousLiquidityTreasuryRule, newLiquidityTreasuryRule)
	if err != nil {
		return nil, err
	}
	return &PortalLiquidityTreasuryUpdatedIterator{contract: _Portal.contract, event: "LiquidityTreasuryUpdated", logs: logs, sub: sub}, nil
}

// WatchLiquidityTreasuryUpdated is a free log subscription operation binding the contract event 0x087168495b2024a05f1e51c26b5abadc7eaa5984c24a419d3563f092693ca1d5.
//
// Solidity: event LiquidityTreasuryUpdated(address indexed previousLiquidityTreasury, address indexed newLiquidityTreasury)
func (_Portal *PortalFilterer) WatchLiquidityTreasuryUpdated(opts *bind.WatchOpts, sink chan<- *PortalLiquidityTreasuryUpdated, previousLiquidityTreasury []common.Address, newLiquidityTreasury []common.Address) (event.Subscription, error) {

	var previousLiquidityTreasuryRule []interface{}
	for _, previousLiquidityTreasuryItem := range previousLiquidityTreasury {
		previousLiquidityTreasuryRule = append(previousLiquidityTreasuryRule, previousLiquidityTreasuryItem)
	}
	var newLiquidityTreasuryRule []interface{}
	for _, newLiquidityTreasuryItem := range newLiquidityTreasury {
		newLiquidityTreasuryRule = append(newLiquidityTreasuryRule, newLiquidityTreasuryItem)
	}

	logs, sub, err := _Portal.contract.WatchLogs(opts, "LiquidityTreasuryUpdated", previousLiquidityTreasuryRule, newLiquidityTreasuryRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(PortalLiquidityTreasuryUpdated)
				if err := _Portal.contract.UnpackLog(event, "LiquidityTreasuryUpdated", log); err != nil {
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

// ParseLiquidityTreasuryUpdated is a log parse operation binding the contract event 0x087168495b2024a05f1e51c26b5abadc7eaa5984c24a419d3563f092693ca1d5.
//
// Solidity: event LiquidityTreasuryUpdated(address indexed previousLiquidityTreasury, address indexed newLiquidityTreasury)
func (_Portal *PortalFilterer) ParseLiquidityTreasuryUpdated(log types.Log) (*PortalLiquidityTreasuryUpdated, error) {
	event := new(PortalLiquidityTreasuryUpdated)
	if err := _Portal.contract.UnpackLog(event, "LiquidityTreasuryUpdated", log); err != nil {
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

// PortalReceiptMintedIterator is returned from FilterReceiptMinted and is used to iterate over the raw logs and unpacked data for ReceiptMinted events raised by the Portal contract.
type PortalReceiptMintedIterator struct {
	Event *PortalReceiptMinted // Event containing the contract specifics and raw log

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
func (it *PortalReceiptMintedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(PortalReceiptMinted)
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
		it.Event = new(PortalReceiptMinted)
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
func (it *PortalReceiptMintedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *PortalReceiptMintedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// PortalReceiptMinted represents a ReceiptMinted event raised by the Portal contract.
type PortalReceiptMinted struct {
	Depositor common.Address
	Token     common.Address
	DepositId *big.Int
	Amount    *big.Int
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterReceiptMinted is a free log retrieval operation binding the contract event 0x862f9b789dcac5ebaaeece5aa03067d5588c6d3f84c140d527894495a028b173.
//
// Solidity: event ReceiptMinted(address indexed depositor, address indexed token, uint256 indexed depositId, uint256 amount)
func (_Portal *PortalFilterer) FilterReceiptMinted(opts *bind.FilterOpts, depositor []common.Address, token []common.Address, depositId []*big.Int) (*PortalReceiptMintedIterator, error) {

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

	logs, sub, err := _Portal.contract.FilterLogs(opts, "ReceiptMinted", depositorRule, tokenRule, depositIdRule)
	if err != nil {
		return nil, err
	}
	return &PortalReceiptMintedIterator{contract: _Portal.contract, event: "ReceiptMinted", logs: logs, sub: sub}, nil
}

// WatchReceiptMinted is a free log subscription operation binding the contract event 0x862f9b789dcac5ebaaeece5aa03067d5588c6d3f84c140d527894495a028b173.
//
// Solidity: event ReceiptMinted(address indexed depositor, address indexed token, uint256 indexed depositId, uint256 amount)
func (_Portal *PortalFilterer) WatchReceiptMinted(opts *bind.WatchOpts, sink chan<- *PortalReceiptMinted, depositor []common.Address, token []common.Address, depositId []*big.Int) (event.Subscription, error) {

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

	logs, sub, err := _Portal.contract.WatchLogs(opts, "ReceiptMinted", depositorRule, tokenRule, depositIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(PortalReceiptMinted)
				if err := _Portal.contract.UnpackLog(event, "ReceiptMinted", log); err != nil {
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

// ParseReceiptMinted is a log parse operation binding the contract event 0x862f9b789dcac5ebaaeece5aa03067d5588c6d3f84c140d527894495a028b173.
//
// Solidity: event ReceiptMinted(address indexed depositor, address indexed token, uint256 indexed depositId, uint256 amount)
func (_Portal *PortalFilterer) ParseReceiptMinted(log types.Log) (*PortalReceiptMinted, error) {
	event := new(PortalReceiptMinted)
	if err := _Portal.contract.UnpackLog(event, "ReceiptMinted", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// PortalReceiptParamsUpdatedIterator is returned from FilterReceiptParamsUpdated and is used to iterate over the raw logs and unpacked data for ReceiptParamsUpdated events raised by the Portal contract.
type PortalReceiptParamsUpdatedIterator struct {
	Event *PortalReceiptParamsUpdated // Event containing the contract specifics and raw log

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
func (it *PortalReceiptParamsUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(PortalReceiptParamsUpdated)
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
		it.Event = new(PortalReceiptParamsUpdated)
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
func (it *PortalReceiptParamsUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *PortalReceiptParamsUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// PortalReceiptParamsUpdated represents a ReceiptParamsUpdated event raised by the Portal contract.
type PortalReceiptParamsUpdated struct {
	Token        common.Address
	AnnualFee    uint8
	MintCap      uint8
	ReceiptToken common.Address
	Raw          types.Log // Blockchain specific contextual infos
}

// FilterReceiptParamsUpdated is a free log retrieval operation binding the contract event 0xcc06fae89af7176b522e80e0f792b25901e66006b16c4ccac33cc75324a16dd3.
//
// Solidity: event ReceiptParamsUpdated(address indexed token, uint8 annualFee, uint8 mintCap, address receiptToken)
func (_Portal *PortalFilterer) FilterReceiptParamsUpdated(opts *bind.FilterOpts, token []common.Address) (*PortalReceiptParamsUpdatedIterator, error) {

	var tokenRule []interface{}
	for _, tokenItem := range token {
		tokenRule = append(tokenRule, tokenItem)
	}

	logs, sub, err := _Portal.contract.FilterLogs(opts, "ReceiptParamsUpdated", tokenRule)
	if err != nil {
		return nil, err
	}
	return &PortalReceiptParamsUpdatedIterator{contract: _Portal.contract, event: "ReceiptParamsUpdated", logs: logs, sub: sub}, nil
}

// WatchReceiptParamsUpdated is a free log subscription operation binding the contract event 0xcc06fae89af7176b522e80e0f792b25901e66006b16c4ccac33cc75324a16dd3.
//
// Solidity: event ReceiptParamsUpdated(address indexed token, uint8 annualFee, uint8 mintCap, address receiptToken)
func (_Portal *PortalFilterer) WatchReceiptParamsUpdated(opts *bind.WatchOpts, sink chan<- *PortalReceiptParamsUpdated, token []common.Address) (event.Subscription, error) {

	var tokenRule []interface{}
	for _, tokenItem := range token {
		tokenRule = append(tokenRule, tokenItem)
	}

	logs, sub, err := _Portal.contract.WatchLogs(opts, "ReceiptParamsUpdated", tokenRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(PortalReceiptParamsUpdated)
				if err := _Portal.contract.UnpackLog(event, "ReceiptParamsUpdated", log); err != nil {
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

// ParseReceiptParamsUpdated is a log parse operation binding the contract event 0xcc06fae89af7176b522e80e0f792b25901e66006b16c4ccac33cc75324a16dd3.
//
// Solidity: event ReceiptParamsUpdated(address indexed token, uint8 annualFee, uint8 mintCap, address receiptToken)
func (_Portal *PortalFilterer) ParseReceiptParamsUpdated(log types.Log) (*PortalReceiptParamsUpdated, error) {
	event := new(PortalReceiptParamsUpdated)
	if err := _Portal.contract.UnpackLog(event, "ReceiptParamsUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// PortalReceiptRepaidIterator is returned from FilterReceiptRepaid and is used to iterate over the raw logs and unpacked data for ReceiptRepaid events raised by the Portal contract.
type PortalReceiptRepaidIterator struct {
	Event *PortalReceiptRepaid // Event containing the contract specifics and raw log

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
func (it *PortalReceiptRepaidIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(PortalReceiptRepaid)
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
		it.Event = new(PortalReceiptRepaid)
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
func (it *PortalReceiptRepaidIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *PortalReceiptRepaidIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// PortalReceiptRepaid represents a ReceiptRepaid event raised by the Portal contract.
type PortalReceiptRepaid struct {
	Depositor common.Address
	Token     common.Address
	DepositId *big.Int
	Amount    *big.Int
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterReceiptRepaid is a free log retrieval operation binding the contract event 0x6043289a72dfdddcba5a5eebd82a24572023a2344a1292dfcf3b56c1a142f606.
//
// Solidity: event ReceiptRepaid(address indexed depositor, address indexed token, uint256 indexed depositId, uint256 amount)
func (_Portal *PortalFilterer) FilterReceiptRepaid(opts *bind.FilterOpts, depositor []common.Address, token []common.Address, depositId []*big.Int) (*PortalReceiptRepaidIterator, error) {

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

	logs, sub, err := _Portal.contract.FilterLogs(opts, "ReceiptRepaid", depositorRule, tokenRule, depositIdRule)
	if err != nil {
		return nil, err
	}
	return &PortalReceiptRepaidIterator{contract: _Portal.contract, event: "ReceiptRepaid", logs: logs, sub: sub}, nil
}

// WatchReceiptRepaid is a free log subscription operation binding the contract event 0x6043289a72dfdddcba5a5eebd82a24572023a2344a1292dfcf3b56c1a142f606.
//
// Solidity: event ReceiptRepaid(address indexed depositor, address indexed token, uint256 indexed depositId, uint256 amount)
func (_Portal *PortalFilterer) WatchReceiptRepaid(opts *bind.WatchOpts, sink chan<- *PortalReceiptRepaid, depositor []common.Address, token []common.Address, depositId []*big.Int) (event.Subscription, error) {

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

	logs, sub, err := _Portal.contract.WatchLogs(opts, "ReceiptRepaid", depositorRule, tokenRule, depositIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(PortalReceiptRepaid)
				if err := _Portal.contract.UnpackLog(event, "ReceiptRepaid", log); err != nil {
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

// ParseReceiptRepaid is a log parse operation binding the contract event 0x6043289a72dfdddcba5a5eebd82a24572023a2344a1292dfcf3b56c1a142f606.
//
// Solidity: event ReceiptRepaid(address indexed depositor, address indexed token, uint256 indexed depositId, uint256 amount)
func (_Portal *PortalFilterer) ParseReceiptRepaid(log types.Log) (*PortalReceiptRepaid, error) {
	event := new(PortalReceiptRepaid)
	if err := _Portal.contract.UnpackLog(event, "ReceiptRepaid", log); err != nil {
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

// PortalTbtcMigrationAllowedUpdatedIterator is returned from FilterTbtcMigrationAllowedUpdated and is used to iterate over the raw logs and unpacked data for TbtcMigrationAllowedUpdated events raised by the Portal contract.
type PortalTbtcMigrationAllowedUpdatedIterator struct {
	Event *PortalTbtcMigrationAllowedUpdated // Event containing the contract specifics and raw log

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
func (it *PortalTbtcMigrationAllowedUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(PortalTbtcMigrationAllowedUpdated)
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
		it.Event = new(PortalTbtcMigrationAllowedUpdated)
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
func (it *PortalTbtcMigrationAllowedUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *PortalTbtcMigrationAllowedUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// PortalTbtcMigrationAllowedUpdated represents a TbtcMigrationAllowedUpdated event raised by the Portal contract.
type PortalTbtcMigrationAllowedUpdated struct {
	Token     common.Address
	IsAllowed bool
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterTbtcMigrationAllowedUpdated is a free log retrieval operation binding the contract event 0x58282641aa313d24bee632c3ec1cdcbc8b924460dbda396d88cfc2a579446ecf.
//
// Solidity: event TbtcMigrationAllowedUpdated(address indexed token, bool isAllowed)
func (_Portal *PortalFilterer) FilterTbtcMigrationAllowedUpdated(opts *bind.FilterOpts, token []common.Address) (*PortalTbtcMigrationAllowedUpdatedIterator, error) {

	var tokenRule []interface{}
	for _, tokenItem := range token {
		tokenRule = append(tokenRule, tokenItem)
	}

	logs, sub, err := _Portal.contract.FilterLogs(opts, "TbtcMigrationAllowedUpdated", tokenRule)
	if err != nil {
		return nil, err
	}
	return &PortalTbtcMigrationAllowedUpdatedIterator{contract: _Portal.contract, event: "TbtcMigrationAllowedUpdated", logs: logs, sub: sub}, nil
}

// WatchTbtcMigrationAllowedUpdated is a free log subscription operation binding the contract event 0x58282641aa313d24bee632c3ec1cdcbc8b924460dbda396d88cfc2a579446ecf.
//
// Solidity: event TbtcMigrationAllowedUpdated(address indexed token, bool isAllowed)
func (_Portal *PortalFilterer) WatchTbtcMigrationAllowedUpdated(opts *bind.WatchOpts, sink chan<- *PortalTbtcMigrationAllowedUpdated, token []common.Address) (event.Subscription, error) {

	var tokenRule []interface{}
	for _, tokenItem := range token {
		tokenRule = append(tokenRule, tokenItem)
	}

	logs, sub, err := _Portal.contract.WatchLogs(opts, "TbtcMigrationAllowedUpdated", tokenRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(PortalTbtcMigrationAllowedUpdated)
				if err := _Portal.contract.UnpackLog(event, "TbtcMigrationAllowedUpdated", log); err != nil {
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

// ParseTbtcMigrationAllowedUpdated is a log parse operation binding the contract event 0x58282641aa313d24bee632c3ec1cdcbc8b924460dbda396d88cfc2a579446ecf.
//
// Solidity: event TbtcMigrationAllowedUpdated(address indexed token, bool isAllowed)
func (_Portal *PortalFilterer) ParseTbtcMigrationAllowedUpdated(log types.Log) (*PortalTbtcMigrationAllowedUpdated, error) {
	event := new(PortalTbtcMigrationAllowedUpdated)
	if err := _Portal.contract.UnpackLog(event, "TbtcMigrationAllowedUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// PortalTbtcMigrationCompletedIterator is returned from FilterTbtcMigrationCompleted and is used to iterate over the raw logs and unpacked data for TbtcMigrationCompleted events raised by the Portal contract.
type PortalTbtcMigrationCompletedIterator struct {
	Event *PortalTbtcMigrationCompleted // Event containing the contract specifics and raw log

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
func (it *PortalTbtcMigrationCompletedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(PortalTbtcMigrationCompleted)
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
		it.Event = new(PortalTbtcMigrationCompleted)
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
func (it *PortalTbtcMigrationCompletedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *PortalTbtcMigrationCompletedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// PortalTbtcMigrationCompleted represents a TbtcMigrationCompleted event raised by the Portal contract.
type PortalTbtcMigrationCompleted struct {
	Depositor common.Address
	Token     common.Address
	DepositId *big.Int
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterTbtcMigrationCompleted is a free log retrieval operation binding the contract event 0x68d9ffd354ad98f5572d5a19eb60d1be4e8cb57a2d8337d10a3ecfca40b1ebe9.
//
// Solidity: event TbtcMigrationCompleted(address indexed depositor, address indexed token, uint256 indexed depositId)
func (_Portal *PortalFilterer) FilterTbtcMigrationCompleted(opts *bind.FilterOpts, depositor []common.Address, token []common.Address, depositId []*big.Int) (*PortalTbtcMigrationCompletedIterator, error) {

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

	logs, sub, err := _Portal.contract.FilterLogs(opts, "TbtcMigrationCompleted", depositorRule, tokenRule, depositIdRule)
	if err != nil {
		return nil, err
	}
	return &PortalTbtcMigrationCompletedIterator{contract: _Portal.contract, event: "TbtcMigrationCompleted", logs: logs, sub: sub}, nil
}

// WatchTbtcMigrationCompleted is a free log subscription operation binding the contract event 0x68d9ffd354ad98f5572d5a19eb60d1be4e8cb57a2d8337d10a3ecfca40b1ebe9.
//
// Solidity: event TbtcMigrationCompleted(address indexed depositor, address indexed token, uint256 indexed depositId)
func (_Portal *PortalFilterer) WatchTbtcMigrationCompleted(opts *bind.WatchOpts, sink chan<- *PortalTbtcMigrationCompleted, depositor []common.Address, token []common.Address, depositId []*big.Int) (event.Subscription, error) {

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

	logs, sub, err := _Portal.contract.WatchLogs(opts, "TbtcMigrationCompleted", depositorRule, tokenRule, depositIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(PortalTbtcMigrationCompleted)
				if err := _Portal.contract.UnpackLog(event, "TbtcMigrationCompleted", log); err != nil {
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

// ParseTbtcMigrationCompleted is a log parse operation binding the contract event 0x68d9ffd354ad98f5572d5a19eb60d1be4e8cb57a2d8337d10a3ecfca40b1ebe9.
//
// Solidity: event TbtcMigrationCompleted(address indexed depositor, address indexed token, uint256 indexed depositId)
func (_Portal *PortalFilterer) ParseTbtcMigrationCompleted(log types.Log) (*PortalTbtcMigrationCompleted, error) {
	event := new(PortalTbtcMigrationCompleted)
	if err := _Portal.contract.UnpackLog(event, "TbtcMigrationCompleted", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// PortalTbtcMigrationRequestedIterator is returned from FilterTbtcMigrationRequested and is used to iterate over the raw logs and unpacked data for TbtcMigrationRequested events raised by the Portal contract.
type PortalTbtcMigrationRequestedIterator struct {
	Event *PortalTbtcMigrationRequested // Event containing the contract specifics and raw log

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
func (it *PortalTbtcMigrationRequestedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(PortalTbtcMigrationRequested)
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
		it.Event = new(PortalTbtcMigrationRequested)
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
func (it *PortalTbtcMigrationRequestedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *PortalTbtcMigrationRequestedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// PortalTbtcMigrationRequested represents a TbtcMigrationRequested event raised by the Portal contract.
type PortalTbtcMigrationRequested struct {
	Depositor common.Address
	Token     common.Address
	DepositId *big.Int
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterTbtcMigrationRequested is a free log retrieval operation binding the contract event 0x4713d6a3ccd421deeb6fb632d8c97878f2e4ae58ac48de1e520b362040b4abf9.
//
// Solidity: event TbtcMigrationRequested(address indexed depositor, address indexed token, uint256 indexed depositId)
func (_Portal *PortalFilterer) FilterTbtcMigrationRequested(opts *bind.FilterOpts, depositor []common.Address, token []common.Address, depositId []*big.Int) (*PortalTbtcMigrationRequestedIterator, error) {

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

	logs, sub, err := _Portal.contract.FilterLogs(opts, "TbtcMigrationRequested", depositorRule, tokenRule, depositIdRule)
	if err != nil {
		return nil, err
	}
	return &PortalTbtcMigrationRequestedIterator{contract: _Portal.contract, event: "TbtcMigrationRequested", logs: logs, sub: sub}, nil
}

// WatchTbtcMigrationRequested is a free log subscription operation binding the contract event 0x4713d6a3ccd421deeb6fb632d8c97878f2e4ae58ac48de1e520b362040b4abf9.
//
// Solidity: event TbtcMigrationRequested(address indexed depositor, address indexed token, uint256 indexed depositId)
func (_Portal *PortalFilterer) WatchTbtcMigrationRequested(opts *bind.WatchOpts, sink chan<- *PortalTbtcMigrationRequested, depositor []common.Address, token []common.Address, depositId []*big.Int) (event.Subscription, error) {

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

	logs, sub, err := _Portal.contract.WatchLogs(opts, "TbtcMigrationRequested", depositorRule, tokenRule, depositIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(PortalTbtcMigrationRequested)
				if err := _Portal.contract.UnpackLog(event, "TbtcMigrationRequested", log); err != nil {
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

// ParseTbtcMigrationRequested is a log parse operation binding the contract event 0x4713d6a3ccd421deeb6fb632d8c97878f2e4ae58ac48de1e520b362040b4abf9.
//
// Solidity: event TbtcMigrationRequested(address indexed depositor, address indexed token, uint256 indexed depositId)
func (_Portal *PortalFilterer) ParseTbtcMigrationRequested(log types.Log) (*PortalTbtcMigrationRequested, error) {
	event := new(PortalTbtcMigrationRequested)
	if err := _Portal.contract.UnpackLog(event, "TbtcMigrationRequested", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// PortalTbtcMigrationStartedIterator is returned from FilterTbtcMigrationStarted and is used to iterate over the raw logs and unpacked data for TbtcMigrationStarted events raised by the Portal contract.
type PortalTbtcMigrationStartedIterator struct {
	Event *PortalTbtcMigrationStarted // Event containing the contract specifics and raw log

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
func (it *PortalTbtcMigrationStartedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(PortalTbtcMigrationStarted)
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
		it.Event = new(PortalTbtcMigrationStarted)
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
func (it *PortalTbtcMigrationStartedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *PortalTbtcMigrationStartedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// PortalTbtcMigrationStarted represents a TbtcMigrationStarted event raised by the Portal contract.
type PortalTbtcMigrationStarted struct {
	Depositor common.Address
	Token     common.Address
	DepositId *big.Int
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterTbtcMigrationStarted is a free log retrieval operation binding the contract event 0x41f6c6b107a872f7e7a62127f1104669af1b4b25a8eba2a4207a8266bd2b2c64.
//
// Solidity: event TbtcMigrationStarted(address indexed depositor, address indexed token, uint256 indexed depositId)
func (_Portal *PortalFilterer) FilterTbtcMigrationStarted(opts *bind.FilterOpts, depositor []common.Address, token []common.Address, depositId []*big.Int) (*PortalTbtcMigrationStartedIterator, error) {

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

	logs, sub, err := _Portal.contract.FilterLogs(opts, "TbtcMigrationStarted", depositorRule, tokenRule, depositIdRule)
	if err != nil {
		return nil, err
	}
	return &PortalTbtcMigrationStartedIterator{contract: _Portal.contract, event: "TbtcMigrationStarted", logs: logs, sub: sub}, nil
}

// WatchTbtcMigrationStarted is a free log subscription operation binding the contract event 0x41f6c6b107a872f7e7a62127f1104669af1b4b25a8eba2a4207a8266bd2b2c64.
//
// Solidity: event TbtcMigrationStarted(address indexed depositor, address indexed token, uint256 indexed depositId)
func (_Portal *PortalFilterer) WatchTbtcMigrationStarted(opts *bind.WatchOpts, sink chan<- *PortalTbtcMigrationStarted, depositor []common.Address, token []common.Address, depositId []*big.Int) (event.Subscription, error) {

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

	logs, sub, err := _Portal.contract.WatchLogs(opts, "TbtcMigrationStarted", depositorRule, tokenRule, depositIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(PortalTbtcMigrationStarted)
				if err := _Portal.contract.UnpackLog(event, "TbtcMigrationStarted", log); err != nil {
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

// ParseTbtcMigrationStarted is a log parse operation binding the contract event 0x41f6c6b107a872f7e7a62127f1104669af1b4b25a8eba2a4207a8266bd2b2c64.
//
// Solidity: event TbtcMigrationStarted(address indexed depositor, address indexed token, uint256 indexed depositId)
func (_Portal *PortalFilterer) ParseTbtcMigrationStarted(log types.Log) (*PortalTbtcMigrationStarted, error) {
	event := new(PortalTbtcMigrationStarted)
	if err := _Portal.contract.UnpackLog(event, "TbtcMigrationStarted", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// PortalTbtcMigrationTreasuryUpdatedIterator is returned from FilterTbtcMigrationTreasuryUpdated and is used to iterate over the raw logs and unpacked data for TbtcMigrationTreasuryUpdated events raised by the Portal contract.
type PortalTbtcMigrationTreasuryUpdatedIterator struct {
	Event *PortalTbtcMigrationTreasuryUpdated // Event containing the contract specifics and raw log

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
func (it *PortalTbtcMigrationTreasuryUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(PortalTbtcMigrationTreasuryUpdated)
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
		it.Event = new(PortalTbtcMigrationTreasuryUpdated)
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
func (it *PortalTbtcMigrationTreasuryUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *PortalTbtcMigrationTreasuryUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// PortalTbtcMigrationTreasuryUpdated represents a TbtcMigrationTreasuryUpdated event raised by the Portal contract.
type PortalTbtcMigrationTreasuryUpdated struct {
	PreviousMigrationTreasury common.Address
	NewMigrationTreasury      common.Address
	Raw                       types.Log // Blockchain specific contextual infos
}

// FilterTbtcMigrationTreasuryUpdated is a free log retrieval operation binding the contract event 0x2393a3a0213901ea187a0528e61d30bfd31577cb6efa698270cc0757e82cc28e.
//
// Solidity: event TbtcMigrationTreasuryUpdated(address indexed previousMigrationTreasury, address indexed newMigrationTreasury)
func (_Portal *PortalFilterer) FilterTbtcMigrationTreasuryUpdated(opts *bind.FilterOpts, previousMigrationTreasury []common.Address, newMigrationTreasury []common.Address) (*PortalTbtcMigrationTreasuryUpdatedIterator, error) {

	var previousMigrationTreasuryRule []interface{}
	for _, previousMigrationTreasuryItem := range previousMigrationTreasury {
		previousMigrationTreasuryRule = append(previousMigrationTreasuryRule, previousMigrationTreasuryItem)
	}
	var newMigrationTreasuryRule []interface{}
	for _, newMigrationTreasuryItem := range newMigrationTreasury {
		newMigrationTreasuryRule = append(newMigrationTreasuryRule, newMigrationTreasuryItem)
	}

	logs, sub, err := _Portal.contract.FilterLogs(opts, "TbtcMigrationTreasuryUpdated", previousMigrationTreasuryRule, newMigrationTreasuryRule)
	if err != nil {
		return nil, err
	}
	return &PortalTbtcMigrationTreasuryUpdatedIterator{contract: _Portal.contract, event: "TbtcMigrationTreasuryUpdated", logs: logs, sub: sub}, nil
}

// WatchTbtcMigrationTreasuryUpdated is a free log subscription operation binding the contract event 0x2393a3a0213901ea187a0528e61d30bfd31577cb6efa698270cc0757e82cc28e.
//
// Solidity: event TbtcMigrationTreasuryUpdated(address indexed previousMigrationTreasury, address indexed newMigrationTreasury)
func (_Portal *PortalFilterer) WatchTbtcMigrationTreasuryUpdated(opts *bind.WatchOpts, sink chan<- *PortalTbtcMigrationTreasuryUpdated, previousMigrationTreasury []common.Address, newMigrationTreasury []common.Address) (event.Subscription, error) {

	var previousMigrationTreasuryRule []interface{}
	for _, previousMigrationTreasuryItem := range previousMigrationTreasury {
		previousMigrationTreasuryRule = append(previousMigrationTreasuryRule, previousMigrationTreasuryItem)
	}
	var newMigrationTreasuryRule []interface{}
	for _, newMigrationTreasuryItem := range newMigrationTreasury {
		newMigrationTreasuryRule = append(newMigrationTreasuryRule, newMigrationTreasuryItem)
	}

	logs, sub, err := _Portal.contract.WatchLogs(opts, "TbtcMigrationTreasuryUpdated", previousMigrationTreasuryRule, newMigrationTreasuryRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(PortalTbtcMigrationTreasuryUpdated)
				if err := _Portal.contract.UnpackLog(event, "TbtcMigrationTreasuryUpdated", log); err != nil {
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

// ParseTbtcMigrationTreasuryUpdated is a log parse operation binding the contract event 0x2393a3a0213901ea187a0528e61d30bfd31577cb6efa698270cc0757e82cc28e.
//
// Solidity: event TbtcMigrationTreasuryUpdated(address indexed previousMigrationTreasury, address indexed newMigrationTreasury)
func (_Portal *PortalFilterer) ParseTbtcMigrationTreasuryUpdated(log types.Log) (*PortalTbtcMigrationTreasuryUpdated, error) {
	event := new(PortalTbtcMigrationTreasuryUpdated)
	if err := _Portal.contract.UnpackLog(event, "TbtcMigrationTreasuryUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// PortalTbtcTokenAddressSetIterator is returned from FilterTbtcTokenAddressSet and is used to iterate over the raw logs and unpacked data for TbtcTokenAddressSet events raised by the Portal contract.
type PortalTbtcTokenAddressSetIterator struct {
	Event *PortalTbtcTokenAddressSet // Event containing the contract specifics and raw log

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
func (it *PortalTbtcTokenAddressSetIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(PortalTbtcTokenAddressSet)
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
		it.Event = new(PortalTbtcTokenAddressSet)
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
func (it *PortalTbtcTokenAddressSetIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *PortalTbtcTokenAddressSetIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// PortalTbtcTokenAddressSet represents a TbtcTokenAddressSet event raised by the Portal contract.
type PortalTbtcTokenAddressSet struct {
	Tbtc common.Address
	Raw  types.Log // Blockchain specific contextual infos
}

// FilterTbtcTokenAddressSet is a free log retrieval operation binding the contract event 0x3d8b27d0955baa4924ce9638e61ff44b8fca3c80475d3dfc8fd6582c5df016cf.
//
// Solidity: event TbtcTokenAddressSet(address tbtc)
func (_Portal *PortalFilterer) FilterTbtcTokenAddressSet(opts *bind.FilterOpts) (*PortalTbtcTokenAddressSetIterator, error) {

	logs, sub, err := _Portal.contract.FilterLogs(opts, "TbtcTokenAddressSet")
	if err != nil {
		return nil, err
	}
	return &PortalTbtcTokenAddressSetIterator{contract: _Portal.contract, event: "TbtcTokenAddressSet", logs: logs, sub: sub}, nil
}

// WatchTbtcTokenAddressSet is a free log subscription operation binding the contract event 0x3d8b27d0955baa4924ce9638e61ff44b8fca3c80475d3dfc8fd6582c5df016cf.
//
// Solidity: event TbtcTokenAddressSet(address tbtc)
func (_Portal *PortalFilterer) WatchTbtcTokenAddressSet(opts *bind.WatchOpts, sink chan<- *PortalTbtcTokenAddressSet) (event.Subscription, error) {

	logs, sub, err := _Portal.contract.WatchLogs(opts, "TbtcTokenAddressSet")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(PortalTbtcTokenAddressSet)
				if err := _Portal.contract.UnpackLog(event, "TbtcTokenAddressSet", log); err != nil {
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

// ParseTbtcTokenAddressSet is a log parse operation binding the contract event 0x3d8b27d0955baa4924ce9638e61ff44b8fca3c80475d3dfc8fd6582c5df016cf.
//
// Solidity: event TbtcTokenAddressSet(address tbtc)
func (_Portal *PortalFilterer) ParseTbtcTokenAddressSet(log types.Log) (*PortalTbtcTokenAddressSet, error) {
	event := new(PortalTbtcTokenAddressSet)
	if err := _Portal.contract.UnpackLog(event, "TbtcTokenAddressSet", log); err != nil {
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

// PortalWithdrawnByLiquidityTreasuryIterator is returned from FilterWithdrawnByLiquidityTreasury and is used to iterate over the raw logs and unpacked data for WithdrawnByLiquidityTreasury events raised by the Portal contract.
type PortalWithdrawnByLiquidityTreasuryIterator struct {
	Event *PortalWithdrawnByLiquidityTreasury // Event containing the contract specifics and raw log

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
func (it *PortalWithdrawnByLiquidityTreasuryIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(PortalWithdrawnByLiquidityTreasury)
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
		it.Event = new(PortalWithdrawnByLiquidityTreasury)
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
func (it *PortalWithdrawnByLiquidityTreasuryIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *PortalWithdrawnByLiquidityTreasuryIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// PortalWithdrawnByLiquidityTreasury represents a WithdrawnByLiquidityTreasury event raised by the Portal contract.
type PortalWithdrawnByLiquidityTreasury struct {
	Token  common.Address
	Amount *big.Int
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterWithdrawnByLiquidityTreasury is a free log retrieval operation binding the contract event 0x9ab9b817afca6d91dd7d523c53a3d2af8939f0a0805d85d0f67b07585fed524b.
//
// Solidity: event WithdrawnByLiquidityTreasury(address indexed token, uint256 amount)
func (_Portal *PortalFilterer) FilterWithdrawnByLiquidityTreasury(opts *bind.FilterOpts, token []common.Address) (*PortalWithdrawnByLiquidityTreasuryIterator, error) {

	var tokenRule []interface{}
	for _, tokenItem := range token {
		tokenRule = append(tokenRule, tokenItem)
	}

	logs, sub, err := _Portal.contract.FilterLogs(opts, "WithdrawnByLiquidityTreasury", tokenRule)
	if err != nil {
		return nil, err
	}
	return &PortalWithdrawnByLiquidityTreasuryIterator{contract: _Portal.contract, event: "WithdrawnByLiquidityTreasury", logs: logs, sub: sub}, nil
}

// WatchWithdrawnByLiquidityTreasury is a free log subscription operation binding the contract event 0x9ab9b817afca6d91dd7d523c53a3d2af8939f0a0805d85d0f67b07585fed524b.
//
// Solidity: event WithdrawnByLiquidityTreasury(address indexed token, uint256 amount)
func (_Portal *PortalFilterer) WatchWithdrawnByLiquidityTreasury(opts *bind.WatchOpts, sink chan<- *PortalWithdrawnByLiquidityTreasury, token []common.Address) (event.Subscription, error) {

	var tokenRule []interface{}
	for _, tokenItem := range token {
		tokenRule = append(tokenRule, tokenItem)
	}

	logs, sub, err := _Portal.contract.WatchLogs(opts, "WithdrawnByLiquidityTreasury", tokenRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(PortalWithdrawnByLiquidityTreasury)
				if err := _Portal.contract.UnpackLog(event, "WithdrawnByLiquidityTreasury", log); err != nil {
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

// ParseWithdrawnByLiquidityTreasury is a log parse operation binding the contract event 0x9ab9b817afca6d91dd7d523c53a3d2af8939f0a0805d85d0f67b07585fed524b.
//
// Solidity: event WithdrawnByLiquidityTreasury(address indexed token, uint256 amount)
func (_Portal *PortalFilterer) ParseWithdrawnByLiquidityTreasury(log types.Log) (*PortalWithdrawnByLiquidityTreasury, error) {
	event := new(PortalWithdrawnByLiquidityTreasury)
	if err := _Portal.contract.UnpackLog(event, "WithdrawnByLiquidityTreasury", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// PortalWithdrawnForTbtcMigrationIterator is returned from FilterWithdrawnForTbtcMigration and is used to iterate over the raw logs and unpacked data for WithdrawnForTbtcMigration events raised by the Portal contract.
type PortalWithdrawnForTbtcMigrationIterator struct {
	Event *PortalWithdrawnForTbtcMigration // Event containing the contract specifics and raw log

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
func (it *PortalWithdrawnForTbtcMigrationIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(PortalWithdrawnForTbtcMigration)
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
		it.Event = new(PortalWithdrawnForTbtcMigration)
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
func (it *PortalWithdrawnForTbtcMigrationIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *PortalWithdrawnForTbtcMigrationIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// PortalWithdrawnForTbtcMigration represents a WithdrawnForTbtcMigration event raised by the Portal contract.
type PortalWithdrawnForTbtcMigration struct {
	Token  common.Address
	Amount *big.Int
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterWithdrawnForTbtcMigration is a free log retrieval operation binding the contract event 0xd9953834583f8ccc107d531dd2133b07f00bf5c8cebe8f594486930986996c98.
//
// Solidity: event WithdrawnForTbtcMigration(address indexed token, uint256 amount)
func (_Portal *PortalFilterer) FilterWithdrawnForTbtcMigration(opts *bind.FilterOpts, token []common.Address) (*PortalWithdrawnForTbtcMigrationIterator, error) {

	var tokenRule []interface{}
	for _, tokenItem := range token {
		tokenRule = append(tokenRule, tokenItem)
	}

	logs, sub, err := _Portal.contract.FilterLogs(opts, "WithdrawnForTbtcMigration", tokenRule)
	if err != nil {
		return nil, err
	}
	return &PortalWithdrawnForTbtcMigrationIterator{contract: _Portal.contract, event: "WithdrawnForTbtcMigration", logs: logs, sub: sub}, nil
}

// WatchWithdrawnForTbtcMigration is a free log subscription operation binding the contract event 0xd9953834583f8ccc107d531dd2133b07f00bf5c8cebe8f594486930986996c98.
//
// Solidity: event WithdrawnForTbtcMigration(address indexed token, uint256 amount)
func (_Portal *PortalFilterer) WatchWithdrawnForTbtcMigration(opts *bind.WatchOpts, sink chan<- *PortalWithdrawnForTbtcMigration, token []common.Address) (event.Subscription, error) {

	var tokenRule []interface{}
	for _, tokenItem := range token {
		tokenRule = append(tokenRule, tokenItem)
	}

	logs, sub, err := _Portal.contract.WatchLogs(opts, "WithdrawnForTbtcMigration", tokenRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(PortalWithdrawnForTbtcMigration)
				if err := _Portal.contract.UnpackLog(event, "WithdrawnForTbtcMigration", log); err != nil {
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

// ParseWithdrawnForTbtcMigration is a log parse operation binding the contract event 0xd9953834583f8ccc107d531dd2133b07f00bf5c8cebe8f594486930986996c98.
//
// Solidity: event WithdrawnForTbtcMigration(address indexed token, uint256 amount)
func (_Portal *PortalFilterer) ParseWithdrawnForTbtcMigration(log types.Log) (*PortalWithdrawnForTbtcMigration, error) {
	event := new(PortalWithdrawnForTbtcMigration)
	if err := _Portal.contract.UnpackLog(event, "WithdrawnForTbtcMigration", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// PortalWithdrawnTbtcMigratedIterator is returned from FilterWithdrawnTbtcMigrated and is used to iterate over the raw logs and unpacked data for WithdrawnTbtcMigrated events raised by the Portal contract.
type PortalWithdrawnTbtcMigratedIterator struct {
	Event *PortalWithdrawnTbtcMigrated // Event containing the contract specifics and raw log

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
func (it *PortalWithdrawnTbtcMigratedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(PortalWithdrawnTbtcMigrated)
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
		it.Event = new(PortalWithdrawnTbtcMigrated)
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
func (it *PortalWithdrawnTbtcMigratedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *PortalWithdrawnTbtcMigratedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// PortalWithdrawnTbtcMigrated represents a WithdrawnTbtcMigrated event raised by the Portal contract.
type PortalWithdrawnTbtcMigrated struct {
	Depositor    common.Address
	Token        common.Address
	TbtcToken    common.Address
	DepositId    *big.Int
	AmountInTbtc *big.Int
	Raw          types.Log // Blockchain specific contextual infos
}

// FilterWithdrawnTbtcMigrated is a free log retrieval operation binding the contract event 0xaabf355ccacfa8b7366b9f6a14af62036d7dd401797d7591faae42a5bbbc3db9.
//
// Solidity: event WithdrawnTbtcMigrated(address indexed depositor, address indexed token, address tbtcToken, uint256 indexed depositId, uint256 amountInTbtc)
func (_Portal *PortalFilterer) FilterWithdrawnTbtcMigrated(opts *bind.FilterOpts, depositor []common.Address, token []common.Address, depositId []*big.Int) (*PortalWithdrawnTbtcMigratedIterator, error) {

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

	logs, sub, err := _Portal.contract.FilterLogs(opts, "WithdrawnTbtcMigrated", depositorRule, tokenRule, depositIdRule)
	if err != nil {
		return nil, err
	}
	return &PortalWithdrawnTbtcMigratedIterator{contract: _Portal.contract, event: "WithdrawnTbtcMigrated", logs: logs, sub: sub}, nil
}

// WatchWithdrawnTbtcMigrated is a free log subscription operation binding the contract event 0xaabf355ccacfa8b7366b9f6a14af62036d7dd401797d7591faae42a5bbbc3db9.
//
// Solidity: event WithdrawnTbtcMigrated(address indexed depositor, address indexed token, address tbtcToken, uint256 indexed depositId, uint256 amountInTbtc)
func (_Portal *PortalFilterer) WatchWithdrawnTbtcMigrated(opts *bind.WatchOpts, sink chan<- *PortalWithdrawnTbtcMigrated, depositor []common.Address, token []common.Address, depositId []*big.Int) (event.Subscription, error) {

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

	logs, sub, err := _Portal.contract.WatchLogs(opts, "WithdrawnTbtcMigrated", depositorRule, tokenRule, depositIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(PortalWithdrawnTbtcMigrated)
				if err := _Portal.contract.UnpackLog(event, "WithdrawnTbtcMigrated", log); err != nil {
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

// ParseWithdrawnTbtcMigrated is a log parse operation binding the contract event 0xaabf355ccacfa8b7366b9f6a14af62036d7dd401797d7591faae42a5bbbc3db9.
//
// Solidity: event WithdrawnTbtcMigrated(address indexed depositor, address indexed token, address tbtcToken, uint256 indexed depositId, uint256 amountInTbtc)
func (_Portal *PortalFilterer) ParseWithdrawnTbtcMigrated(log types.Log) (*PortalWithdrawnTbtcMigrated, error) {
	event := new(PortalWithdrawnTbtcMigrated)
	if err := _Portal.contract.UnpackLog(event, "WithdrawnTbtcMigrated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
