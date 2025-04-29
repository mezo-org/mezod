// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package contract

import (
	"context"
	"fmt"
	"math/big"
	"strings"
	"sync"
	"time"

	hostchainabi "github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"
	"github.com/mezo-org/mezod/ethereum/bindings/portal/mainnet/gen/abi"

	"github.com/ipfs/go-log"

	"github.com/keep-network/keep-common/pkg/chain/ethereum"
	chainutil "github.com/keep-network/keep-common/pkg/chain/ethereum/ethutil"
	"github.com/keep-network/keep-common/pkg/subscription"
)

// Create a package-level logger for this contract. The logger exists at
// package level so that the logger is registered at startup and can be
// included or excluded from logging at startup by name.
var mbLogger = log.Logger("keep-contract-MezoBridge")

type MezoBridge struct {
	contract          *abi.MezoBridge
	contractAddress   common.Address
	contractABI       *hostchainabi.ABI
	caller            bind.ContractCaller
	transactor        bind.ContractTransactor
	callerOptions     *bind.CallOpts
	transactorOptions *bind.TransactOpts
	errorResolver     *chainutil.ErrorResolver
	nonceManager      *ethereum.NonceManager
	miningWaiter      *chainutil.MiningWaiter
	blockCounter      *ethereum.BlockCounter

	transactionMutex *sync.Mutex
}

func NewMezoBridge(
	contractAddress common.Address,
	chainId *big.Int,
	accountKey *keystore.Key,
	backend bind.ContractBackend,
	nonceManager *ethereum.NonceManager,
	miningWaiter *chainutil.MiningWaiter,
	blockCounter *ethereum.BlockCounter,
	transactionMutex *sync.Mutex,
) (*MezoBridge, error) {
	callerOptions := &bind.CallOpts{
		From: accountKey.Address,
	}

	transactorOptions, err := bind.NewKeyedTransactorWithChainID(
		accountKey.PrivateKey,
		chainId,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to instantiate transactor: [%v]", err)
	}

	contract, err := abi.NewMezoBridge(
		contractAddress,
		backend,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to instantiate contract at address: %s [%v]",
			contractAddress.String(),
			err,
		)
	}

	contractABI, err := hostchainabi.JSON(strings.NewReader(abi.MezoBridgeABI))
	if err != nil {
		return nil, fmt.Errorf("failed to instantiate ABI: [%v]", err)
	}

	return &MezoBridge{
		contract:          contract,
		contractAddress:   contractAddress,
		contractABI:       &contractABI,
		caller:            backend,
		transactor:        backend,
		callerOptions:     callerOptions,
		transactorOptions: transactorOptions,
		errorResolver:     chainutil.NewErrorResolver(backend, &contractABI, &contractAddress),
		nonceManager:      nonceManager,
		miningWaiter:      miningWaiter,
		blockCounter:      blockCounter,
		transactionMutex:  transactionMutex,
	}, nil
}

// ----- Non-const Methods ------

// Transaction submission.
func (mb *MezoBridge) AcceptOwnership(

	transactionOptions ...chainutil.TransactionOptions,
) (*types.Transaction, error) {
	mbLogger.Debug(
		"submitting transaction acceptOwnership",
	)

	mb.transactionMutex.Lock()
	defer mb.transactionMutex.Unlock()

	// create a copy
	transactorOptions := new(bind.TransactOpts)
	*transactorOptions = *mb.transactorOptions

	if len(transactionOptions) > 1 {
		return nil, fmt.Errorf(
			"could not process multiple transaction options sets",
		)
	} else if len(transactionOptions) > 0 {
		transactionOptions[0].Apply(transactorOptions)
	}

	nonce, err := mb.nonceManager.CurrentNonce()
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve account nonce: %v", err)
	}

	transactorOptions.Nonce = new(big.Int).SetUint64(nonce)

	transaction, err := mb.contract.AcceptOwnership(
		transactorOptions,
	)
	if err != nil {
		return transaction, mb.errorResolver.ResolveError(
			err,
			mb.transactorOptions.From,
			nil,
			"acceptOwnership",
		)
	}

	mbLogger.Infof(
		"submitted transaction acceptOwnership with id: [%s] and nonce [%v]",
		transaction.Hash(),
		transaction.Nonce(),
	)

	go mb.miningWaiter.ForceMining(
		transaction,
		transactorOptions,
		func(newTransactorOptions *bind.TransactOpts) (*types.Transaction, error) {
			// If original transactor options has a non-zero gas limit, that
			// means the client code set it on their own. In that case, we
			// should rewrite the gas limit from the original transaction
			// for each resubmission. If the gas limit is not set by the client
			// code, let the the submitter re-estimate the gas limit on each
			// resubmission.
			if transactorOptions.GasLimit != 0 {
				newTransactorOptions.GasLimit = transactorOptions.GasLimit
			}

			transaction, err := mb.contract.AcceptOwnership(
				newTransactorOptions,
			)
			if err != nil {
				return nil, mb.errorResolver.ResolveError(
					err,
					mb.transactorOptions.From,
					nil,
					"acceptOwnership",
				)
			}

			mbLogger.Infof(
				"submitted transaction acceptOwnership with id: [%s] and nonce [%v]",
				transaction.Hash(),
				transaction.Nonce(),
			)

			return transaction, nil
		},
	)

	mb.nonceManager.IncrementNonce()

	return transaction, err
}

// Non-mutating call, not a transaction submission.
func (mb *MezoBridge) CallAcceptOwnership(
	blockNumber *big.Int,
) error {
	var result interface{} = nil

	err := chainutil.CallAtBlock(
		mb.transactorOptions.From,
		blockNumber, nil,
		mb.contractABI,
		mb.caller,
		mb.errorResolver,
		mb.contractAddress,
		"acceptOwnership",
		&result,
	)

	return err
}

func (mb *MezoBridge) AcceptOwnershipGasEstimate() (uint64, error) {
	var result uint64

	result, err := chainutil.EstimateGas(
		mb.callerOptions.From,
		mb.contractAddress,
		"acceptOwnership",
		mb.contractABI,
		mb.transactor,
	)

	return result, err
}

// Transaction submission.
func (mb *MezoBridge) BridgeERC20(
	arg_ERC20Token common.Address,
	arg_amount *big.Int,
	arg_recipient common.Address,

	transactionOptions ...chainutil.TransactionOptions,
) (*types.Transaction, error) {
	mbLogger.Debug(
		"submitting transaction bridgeERC20",
		" params: ",
		fmt.Sprint(
			arg_ERC20Token,
			arg_amount,
			arg_recipient,
		),
	)

	mb.transactionMutex.Lock()
	defer mb.transactionMutex.Unlock()

	// create a copy
	transactorOptions := new(bind.TransactOpts)
	*transactorOptions = *mb.transactorOptions

	if len(transactionOptions) > 1 {
		return nil, fmt.Errorf(
			"could not process multiple transaction options sets",
		)
	} else if len(transactionOptions) > 0 {
		transactionOptions[0].Apply(transactorOptions)
	}

	nonce, err := mb.nonceManager.CurrentNonce()
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve account nonce: %v", err)
	}

	transactorOptions.Nonce = new(big.Int).SetUint64(nonce)

	transaction, err := mb.contract.BridgeERC20(
		transactorOptions,
		arg_ERC20Token,
		arg_amount,
		arg_recipient,
	)
	if err != nil {
		return transaction, mb.errorResolver.ResolveError(
			err,
			mb.transactorOptions.From,
			nil,
			"bridgeERC20",
			arg_ERC20Token,
			arg_amount,
			arg_recipient,
		)
	}

	mbLogger.Infof(
		"submitted transaction bridgeERC20 with id: [%s] and nonce [%v]",
		transaction.Hash(),
		transaction.Nonce(),
	)

	go mb.miningWaiter.ForceMining(
		transaction,
		transactorOptions,
		func(newTransactorOptions *bind.TransactOpts) (*types.Transaction, error) {
			// If original transactor options has a non-zero gas limit, that
			// means the client code set it on their own. In that case, we
			// should rewrite the gas limit from the original transaction
			// for each resubmission. If the gas limit is not set by the client
			// code, let the the submitter re-estimate the gas limit on each
			// resubmission.
			if transactorOptions.GasLimit != 0 {
				newTransactorOptions.GasLimit = transactorOptions.GasLimit
			}

			transaction, err := mb.contract.BridgeERC20(
				newTransactorOptions,
				arg_ERC20Token,
				arg_amount,
				arg_recipient,
			)
			if err != nil {
				return nil, mb.errorResolver.ResolveError(
					err,
					mb.transactorOptions.From,
					nil,
					"bridgeERC20",
					arg_ERC20Token,
					arg_amount,
					arg_recipient,
				)
			}

			mbLogger.Infof(
				"submitted transaction bridgeERC20 with id: [%s] and nonce [%v]",
				transaction.Hash(),
				transaction.Nonce(),
			)

			return transaction, nil
		},
	)

	mb.nonceManager.IncrementNonce()

	return transaction, err
}

// Non-mutating call, not a transaction submission.
func (mb *MezoBridge) CallBridgeERC20(
	arg_ERC20Token common.Address,
	arg_amount *big.Int,
	arg_recipient common.Address,
	blockNumber *big.Int,
) error {
	var result interface{} = nil

	err := chainutil.CallAtBlock(
		mb.transactorOptions.From,
		blockNumber, nil,
		mb.contractABI,
		mb.caller,
		mb.errorResolver,
		mb.contractAddress,
		"bridgeERC20",
		&result,
		arg_ERC20Token,
		arg_amount,
		arg_recipient,
	)

	return err
}

func (mb *MezoBridge) BridgeERC20GasEstimate(
	arg_ERC20Token common.Address,
	arg_amount *big.Int,
	arg_recipient common.Address,
) (uint64, error) {
	var result uint64

	result, err := chainutil.EstimateGas(
		mb.callerOptions.From,
		mb.contractAddress,
		"bridgeERC20",
		mb.contractABI,
		mb.transactor,
		arg_ERC20Token,
		arg_amount,
		arg_recipient,
	)

	return result, err
}

// Transaction submission.
func (mb *MezoBridge) BridgeTBTC(
	arg_amount *big.Int,
	arg_recipient common.Address,

	transactionOptions ...chainutil.TransactionOptions,
) (*types.Transaction, error) {
	mbLogger.Debug(
		"submitting transaction bridgeTBTC",
		" params: ",
		fmt.Sprint(
			arg_amount,
			arg_recipient,
		),
	)

	mb.transactionMutex.Lock()
	defer mb.transactionMutex.Unlock()

	// create a copy
	transactorOptions := new(bind.TransactOpts)
	*transactorOptions = *mb.transactorOptions

	if len(transactionOptions) > 1 {
		return nil, fmt.Errorf(
			"could not process multiple transaction options sets",
		)
	} else if len(transactionOptions) > 0 {
		transactionOptions[0].Apply(transactorOptions)
	}

	nonce, err := mb.nonceManager.CurrentNonce()
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve account nonce: %v", err)
	}

	transactorOptions.Nonce = new(big.Int).SetUint64(nonce)

	transaction, err := mb.contract.BridgeTBTC(
		transactorOptions,
		arg_amount,
		arg_recipient,
	)
	if err != nil {
		return transaction, mb.errorResolver.ResolveError(
			err,
			mb.transactorOptions.From,
			nil,
			"bridgeTBTC",
			arg_amount,
			arg_recipient,
		)
	}

	mbLogger.Infof(
		"submitted transaction bridgeTBTC with id: [%s] and nonce [%v]",
		transaction.Hash(),
		transaction.Nonce(),
	)

	go mb.miningWaiter.ForceMining(
		transaction,
		transactorOptions,
		func(newTransactorOptions *bind.TransactOpts) (*types.Transaction, error) {
			// If original transactor options has a non-zero gas limit, that
			// means the client code set it on their own. In that case, we
			// should rewrite the gas limit from the original transaction
			// for each resubmission. If the gas limit is not set by the client
			// code, let the the submitter re-estimate the gas limit on each
			// resubmission.
			if transactorOptions.GasLimit != 0 {
				newTransactorOptions.GasLimit = transactorOptions.GasLimit
			}

			transaction, err := mb.contract.BridgeTBTC(
				newTransactorOptions,
				arg_amount,
				arg_recipient,
			)
			if err != nil {
				return nil, mb.errorResolver.ResolveError(
					err,
					mb.transactorOptions.From,
					nil,
					"bridgeTBTC",
					arg_amount,
					arg_recipient,
				)
			}

			mbLogger.Infof(
				"submitted transaction bridgeTBTC with id: [%s] and nonce [%v]",
				transaction.Hash(),
				transaction.Nonce(),
			)

			return transaction, nil
		},
	)

	mb.nonceManager.IncrementNonce()

	return transaction, err
}

// Non-mutating call, not a transaction submission.
func (mb *MezoBridge) CallBridgeTBTC(
	arg_amount *big.Int,
	arg_recipient common.Address,
	blockNumber *big.Int,
) error {
	var result interface{} = nil

	err := chainutil.CallAtBlock(
		mb.transactorOptions.From,
		blockNumber, nil,
		mb.contractABI,
		mb.caller,
		mb.errorResolver,
		mb.contractAddress,
		"bridgeTBTC",
		&result,
		arg_amount,
		arg_recipient,
	)

	return err
}

func (mb *MezoBridge) BridgeTBTCGasEstimate(
	arg_amount *big.Int,
	arg_recipient common.Address,
) (uint64, error) {
	var result uint64

	result, err := chainutil.EstimateGas(
		mb.callerOptions.From,
		mb.contractAddress,
		"bridgeTBTC",
		mb.contractABI,
		mb.transactor,
		arg_amount,
		arg_recipient,
	)

	return result, err
}

// Transaction submission.
func (mb *MezoBridge) BridgeTBTCWithPermit(
	arg_amount *big.Int,
	arg_recipient common.Address,
	arg_deadline *big.Int,
	arg_v uint8,
	arg_r [32]byte,
	arg_s [32]byte,

	transactionOptions ...chainutil.TransactionOptions,
) (*types.Transaction, error) {
	mbLogger.Debug(
		"submitting transaction bridgeTBTCWithPermit",
		" params: ",
		fmt.Sprint(
			arg_amount,
			arg_recipient,
			arg_deadline,
			arg_v,
			arg_r,
			arg_s,
		),
	)

	mb.transactionMutex.Lock()
	defer mb.transactionMutex.Unlock()

	// create a copy
	transactorOptions := new(bind.TransactOpts)
	*transactorOptions = *mb.transactorOptions

	if len(transactionOptions) > 1 {
		return nil, fmt.Errorf(
			"could not process multiple transaction options sets",
		)
	} else if len(transactionOptions) > 0 {
		transactionOptions[0].Apply(transactorOptions)
	}

	nonce, err := mb.nonceManager.CurrentNonce()
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve account nonce: %v", err)
	}

	transactorOptions.Nonce = new(big.Int).SetUint64(nonce)

	transaction, err := mb.contract.BridgeTBTCWithPermit(
		transactorOptions,
		arg_amount,
		arg_recipient,
		arg_deadline,
		arg_v,
		arg_r,
		arg_s,
	)
	if err != nil {
		return transaction, mb.errorResolver.ResolveError(
			err,
			mb.transactorOptions.From,
			nil,
			"bridgeTBTCWithPermit",
			arg_amount,
			arg_recipient,
			arg_deadline,
			arg_v,
			arg_r,
			arg_s,
		)
	}

	mbLogger.Infof(
		"submitted transaction bridgeTBTCWithPermit with id: [%s] and nonce [%v]",
		transaction.Hash(),
		transaction.Nonce(),
	)

	go mb.miningWaiter.ForceMining(
		transaction,
		transactorOptions,
		func(newTransactorOptions *bind.TransactOpts) (*types.Transaction, error) {
			// If original transactor options has a non-zero gas limit, that
			// means the client code set it on their own. In that case, we
			// should rewrite the gas limit from the original transaction
			// for each resubmission. If the gas limit is not set by the client
			// code, let the the submitter re-estimate the gas limit on each
			// resubmission.
			if transactorOptions.GasLimit != 0 {
				newTransactorOptions.GasLimit = transactorOptions.GasLimit
			}

			transaction, err := mb.contract.BridgeTBTCWithPermit(
				newTransactorOptions,
				arg_amount,
				arg_recipient,
				arg_deadline,
				arg_v,
				arg_r,
				arg_s,
			)
			if err != nil {
				return nil, mb.errorResolver.ResolveError(
					err,
					mb.transactorOptions.From,
					nil,
					"bridgeTBTCWithPermit",
					arg_amount,
					arg_recipient,
					arg_deadline,
					arg_v,
					arg_r,
					arg_s,
				)
			}

			mbLogger.Infof(
				"submitted transaction bridgeTBTCWithPermit with id: [%s] and nonce [%v]",
				transaction.Hash(),
				transaction.Nonce(),
			)

			return transaction, nil
		},
	)

	mb.nonceManager.IncrementNonce()

	return transaction, err
}

// Non-mutating call, not a transaction submission.
func (mb *MezoBridge) CallBridgeTBTCWithPermit(
	arg_amount *big.Int,
	arg_recipient common.Address,
	arg_deadline *big.Int,
	arg_v uint8,
	arg_r [32]byte,
	arg_s [32]byte,
	blockNumber *big.Int,
) error {
	var result interface{} = nil

	err := chainutil.CallAtBlock(
		mb.transactorOptions.From,
		blockNumber, nil,
		mb.contractABI,
		mb.caller,
		mb.errorResolver,
		mb.contractAddress,
		"bridgeTBTCWithPermit",
		&result,
		arg_amount,
		arg_recipient,
		arg_deadline,
		arg_v,
		arg_r,
		arg_s,
	)

	return err
}

func (mb *MezoBridge) BridgeTBTCWithPermitGasEstimate(
	arg_amount *big.Int,
	arg_recipient common.Address,
	arg_deadline *big.Int,
	arg_v uint8,
	arg_r [32]byte,
	arg_s [32]byte,
) (uint64, error) {
	var result uint64

	result, err := chainutil.EstimateGas(
		mb.callerOptions.From,
		mb.contractAddress,
		"bridgeTBTCWithPermit",
		mb.contractABI,
		mb.transactor,
		arg_amount,
		arg_recipient,
		arg_deadline,
		arg_v,
		arg_r,
		arg_s,
	)

	return result, err
}

// Transaction submission.
func (mb *MezoBridge) DisableERC20Token(
	arg_ERC20Token common.Address,

	transactionOptions ...chainutil.TransactionOptions,
) (*types.Transaction, error) {
	mbLogger.Debug(
		"submitting transaction disableERC20Token",
		" params: ",
		fmt.Sprint(
			arg_ERC20Token,
		),
	)

	mb.transactionMutex.Lock()
	defer mb.transactionMutex.Unlock()

	// create a copy
	transactorOptions := new(bind.TransactOpts)
	*transactorOptions = *mb.transactorOptions

	if len(transactionOptions) > 1 {
		return nil, fmt.Errorf(
			"could not process multiple transaction options sets",
		)
	} else if len(transactionOptions) > 0 {
		transactionOptions[0].Apply(transactorOptions)
	}

	nonce, err := mb.nonceManager.CurrentNonce()
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve account nonce: %v", err)
	}

	transactorOptions.Nonce = new(big.Int).SetUint64(nonce)

	transaction, err := mb.contract.DisableERC20Token(
		transactorOptions,
		arg_ERC20Token,
	)
	if err != nil {
		return transaction, mb.errorResolver.ResolveError(
			err,
			mb.transactorOptions.From,
			nil,
			"disableERC20Token",
			arg_ERC20Token,
		)
	}

	mbLogger.Infof(
		"submitted transaction disableERC20Token with id: [%s] and nonce [%v]",
		transaction.Hash(),
		transaction.Nonce(),
	)

	go mb.miningWaiter.ForceMining(
		transaction,
		transactorOptions,
		func(newTransactorOptions *bind.TransactOpts) (*types.Transaction, error) {
			// If original transactor options has a non-zero gas limit, that
			// means the client code set it on their own. In that case, we
			// should rewrite the gas limit from the original transaction
			// for each resubmission. If the gas limit is not set by the client
			// code, let the the submitter re-estimate the gas limit on each
			// resubmission.
			if transactorOptions.GasLimit != 0 {
				newTransactorOptions.GasLimit = transactorOptions.GasLimit
			}

			transaction, err := mb.contract.DisableERC20Token(
				newTransactorOptions,
				arg_ERC20Token,
			)
			if err != nil {
				return nil, mb.errorResolver.ResolveError(
					err,
					mb.transactorOptions.From,
					nil,
					"disableERC20Token",
					arg_ERC20Token,
				)
			}

			mbLogger.Infof(
				"submitted transaction disableERC20Token with id: [%s] and nonce [%v]",
				transaction.Hash(),
				transaction.Nonce(),
			)

			return transaction, nil
		},
	)

	mb.nonceManager.IncrementNonce()

	return transaction, err
}

// Non-mutating call, not a transaction submission.
func (mb *MezoBridge) CallDisableERC20Token(
	arg_ERC20Token common.Address,
	blockNumber *big.Int,
) error {
	var result interface{} = nil

	err := chainutil.CallAtBlock(
		mb.transactorOptions.From,
		blockNumber, nil,
		mb.contractABI,
		mb.caller,
		mb.errorResolver,
		mb.contractAddress,
		"disableERC20Token",
		&result,
		arg_ERC20Token,
	)

	return err
}

func (mb *MezoBridge) DisableERC20TokenGasEstimate(
	arg_ERC20Token common.Address,
) (uint64, error) {
	var result uint64

	result, err := chainutil.EstimateGas(
		mb.callerOptions.From,
		mb.contractAddress,
		"disableERC20Token",
		mb.contractABI,
		mb.transactor,
		arg_ERC20Token,
	)

	return result, err
}

// Transaction submission.
func (mb *MezoBridge) EnableERC20Token(
	arg_ERC20Token common.Address,
	arg_minERC20Amount *big.Int,

	transactionOptions ...chainutil.TransactionOptions,
) (*types.Transaction, error) {
	mbLogger.Debug(
		"submitting transaction enableERC20Token",
		" params: ",
		fmt.Sprint(
			arg_ERC20Token,
			arg_minERC20Amount,
		),
	)

	mb.transactionMutex.Lock()
	defer mb.transactionMutex.Unlock()

	// create a copy
	transactorOptions := new(bind.TransactOpts)
	*transactorOptions = *mb.transactorOptions

	if len(transactionOptions) > 1 {
		return nil, fmt.Errorf(
			"could not process multiple transaction options sets",
		)
	} else if len(transactionOptions) > 0 {
		transactionOptions[0].Apply(transactorOptions)
	}

	nonce, err := mb.nonceManager.CurrentNonce()
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve account nonce: %v", err)
	}

	transactorOptions.Nonce = new(big.Int).SetUint64(nonce)

	transaction, err := mb.contract.EnableERC20Token(
		transactorOptions,
		arg_ERC20Token,
		arg_minERC20Amount,
	)
	if err != nil {
		return transaction, mb.errorResolver.ResolveError(
			err,
			mb.transactorOptions.From,
			nil,
			"enableERC20Token",
			arg_ERC20Token,
			arg_minERC20Amount,
		)
	}

	mbLogger.Infof(
		"submitted transaction enableERC20Token with id: [%s] and nonce [%v]",
		transaction.Hash(),
		transaction.Nonce(),
	)

	go mb.miningWaiter.ForceMining(
		transaction,
		transactorOptions,
		func(newTransactorOptions *bind.TransactOpts) (*types.Transaction, error) {
			// If original transactor options has a non-zero gas limit, that
			// means the client code set it on their own. In that case, we
			// should rewrite the gas limit from the original transaction
			// for each resubmission. If the gas limit is not set by the client
			// code, let the the submitter re-estimate the gas limit on each
			// resubmission.
			if transactorOptions.GasLimit != 0 {
				newTransactorOptions.GasLimit = transactorOptions.GasLimit
			}

			transaction, err := mb.contract.EnableERC20Token(
				newTransactorOptions,
				arg_ERC20Token,
				arg_minERC20Amount,
			)
			if err != nil {
				return nil, mb.errorResolver.ResolveError(
					err,
					mb.transactorOptions.From,
					nil,
					"enableERC20Token",
					arg_ERC20Token,
					arg_minERC20Amount,
				)
			}

			mbLogger.Infof(
				"submitted transaction enableERC20Token with id: [%s] and nonce [%v]",
				transaction.Hash(),
				transaction.Nonce(),
			)

			return transaction, nil
		},
	)

	mb.nonceManager.IncrementNonce()

	return transaction, err
}

// Non-mutating call, not a transaction submission.
func (mb *MezoBridge) CallEnableERC20Token(
	arg_ERC20Token common.Address,
	arg_minERC20Amount *big.Int,
	blockNumber *big.Int,
) error {
	var result interface{} = nil

	err := chainutil.CallAtBlock(
		mb.transactorOptions.From,
		blockNumber, nil,
		mb.contractABI,
		mb.caller,
		mb.errorResolver,
		mb.contractAddress,
		"enableERC20Token",
		&result,
		arg_ERC20Token,
		arg_minERC20Amount,
	)

	return err
}

func (mb *MezoBridge) EnableERC20TokenGasEstimate(
	arg_ERC20Token common.Address,
	arg_minERC20Amount *big.Int,
) (uint64, error) {
	var result uint64

	result, err := chainutil.EstimateGas(
		mb.callerOptions.From,
		mb.contractAddress,
		"enableERC20Token",
		mb.contractABI,
		mb.transactor,
		arg_ERC20Token,
		arg_minERC20Amount,
	)

	return result, err
}

// Transaction submission.
func (mb *MezoBridge) FinalizeBTCBridging(
	arg_btcDepositKey *big.Int,
	arg_recipient common.Address,

	transactionOptions ...chainutil.TransactionOptions,
) (*types.Transaction, error) {
	mbLogger.Debug(
		"submitting transaction finalizeBTCBridging",
		" params: ",
		fmt.Sprint(
			arg_btcDepositKey,
			arg_recipient,
		),
	)

	mb.transactionMutex.Lock()
	defer mb.transactionMutex.Unlock()

	// create a copy
	transactorOptions := new(bind.TransactOpts)
	*transactorOptions = *mb.transactorOptions

	if len(transactionOptions) > 1 {
		return nil, fmt.Errorf(
			"could not process multiple transaction options sets",
		)
	} else if len(transactionOptions) > 0 {
		transactionOptions[0].Apply(transactorOptions)
	}

	nonce, err := mb.nonceManager.CurrentNonce()
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve account nonce: %v", err)
	}

	transactorOptions.Nonce = new(big.Int).SetUint64(nonce)

	transaction, err := mb.contract.FinalizeBTCBridging(
		transactorOptions,
		arg_btcDepositKey,
		arg_recipient,
	)
	if err != nil {
		return transaction, mb.errorResolver.ResolveError(
			err,
			mb.transactorOptions.From,
			nil,
			"finalizeBTCBridging",
			arg_btcDepositKey,
			arg_recipient,
		)
	}

	mbLogger.Infof(
		"submitted transaction finalizeBTCBridging with id: [%s] and nonce [%v]",
		transaction.Hash(),
		transaction.Nonce(),
	)

	go mb.miningWaiter.ForceMining(
		transaction,
		transactorOptions,
		func(newTransactorOptions *bind.TransactOpts) (*types.Transaction, error) {
			// If original transactor options has a non-zero gas limit, that
			// means the client code set it on their own. In that case, we
			// should rewrite the gas limit from the original transaction
			// for each resubmission. If the gas limit is not set by the client
			// code, let the the submitter re-estimate the gas limit on each
			// resubmission.
			if transactorOptions.GasLimit != 0 {
				newTransactorOptions.GasLimit = transactorOptions.GasLimit
			}

			transaction, err := mb.contract.FinalizeBTCBridging(
				newTransactorOptions,
				arg_btcDepositKey,
				arg_recipient,
			)
			if err != nil {
				return nil, mb.errorResolver.ResolveError(
					err,
					mb.transactorOptions.From,
					nil,
					"finalizeBTCBridging",
					arg_btcDepositKey,
					arg_recipient,
				)
			}

			mbLogger.Infof(
				"submitted transaction finalizeBTCBridging with id: [%s] and nonce [%v]",
				transaction.Hash(),
				transaction.Nonce(),
			)

			return transaction, nil
		},
	)

	mb.nonceManager.IncrementNonce()

	return transaction, err
}

// Non-mutating call, not a transaction submission.
func (mb *MezoBridge) CallFinalizeBTCBridging(
	arg_btcDepositKey *big.Int,
	arg_recipient common.Address,
	blockNumber *big.Int,
) error {
	var result interface{} = nil

	err := chainutil.CallAtBlock(
		mb.transactorOptions.From,
		blockNumber, nil,
		mb.contractABI,
		mb.caller,
		mb.errorResolver,
		mb.contractAddress,
		"finalizeBTCBridging",
		&result,
		arg_btcDepositKey,
		arg_recipient,
	)

	return err
}

func (mb *MezoBridge) FinalizeBTCBridgingGasEstimate(
	arg_btcDepositKey *big.Int,
	arg_recipient common.Address,
) (uint64, error) {
	var result uint64

	result, err := chainutil.EstimateGas(
		mb.callerOptions.From,
		mb.contractAddress,
		"finalizeBTCBridging",
		mb.contractABI,
		mb.transactor,
		arg_btcDepositKey,
		arg_recipient,
	)

	return result, err
}

// Transaction submission.
func (mb *MezoBridge) Initialize(
	arg__tbtcBridge common.Address,
	arg__tbtcVault common.Address,
	arg__tbtcToken common.Address,
	arg__initialSequence *big.Int,

	transactionOptions ...chainutil.TransactionOptions,
) (*types.Transaction, error) {
	mbLogger.Debug(
		"submitting transaction initialize",
		" params: ",
		fmt.Sprint(
			arg__tbtcBridge,
			arg__tbtcVault,
			arg__tbtcToken,
			arg__initialSequence,
		),
	)

	mb.transactionMutex.Lock()
	defer mb.transactionMutex.Unlock()

	// create a copy
	transactorOptions := new(bind.TransactOpts)
	*transactorOptions = *mb.transactorOptions

	if len(transactionOptions) > 1 {
		return nil, fmt.Errorf(
			"could not process multiple transaction options sets",
		)
	} else if len(transactionOptions) > 0 {
		transactionOptions[0].Apply(transactorOptions)
	}

	nonce, err := mb.nonceManager.CurrentNonce()
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve account nonce: %v", err)
	}

	transactorOptions.Nonce = new(big.Int).SetUint64(nonce)

	transaction, err := mb.contract.Initialize(
		transactorOptions,
		arg__tbtcBridge,
		arg__tbtcVault,
		arg__tbtcToken,
		arg__initialSequence,
	)
	if err != nil {
		return transaction, mb.errorResolver.ResolveError(
			err,
			mb.transactorOptions.From,
			nil,
			"initialize",
			arg__tbtcBridge,
			arg__tbtcVault,
			arg__tbtcToken,
			arg__initialSequence,
		)
	}

	mbLogger.Infof(
		"submitted transaction initialize with id: [%s] and nonce [%v]",
		transaction.Hash(),
		transaction.Nonce(),
	)

	go mb.miningWaiter.ForceMining(
		transaction,
		transactorOptions,
		func(newTransactorOptions *bind.TransactOpts) (*types.Transaction, error) {
			// If original transactor options has a non-zero gas limit, that
			// means the client code set it on their own. In that case, we
			// should rewrite the gas limit from the original transaction
			// for each resubmission. If the gas limit is not set by the client
			// code, let the the submitter re-estimate the gas limit on each
			// resubmission.
			if transactorOptions.GasLimit != 0 {
				newTransactorOptions.GasLimit = transactorOptions.GasLimit
			}

			transaction, err := mb.contract.Initialize(
				newTransactorOptions,
				arg__tbtcBridge,
				arg__tbtcVault,
				arg__tbtcToken,
				arg__initialSequence,
			)
			if err != nil {
				return nil, mb.errorResolver.ResolveError(
					err,
					mb.transactorOptions.From,
					nil,
					"initialize",
					arg__tbtcBridge,
					arg__tbtcVault,
					arg__tbtcToken,
					arg__initialSequence,
				)
			}

			mbLogger.Infof(
				"submitted transaction initialize with id: [%s] and nonce [%v]",
				transaction.Hash(),
				transaction.Nonce(),
			)

			return transaction, nil
		},
	)

	mb.nonceManager.IncrementNonce()

	return transaction, err
}

// Non-mutating call, not a transaction submission.
func (mb *MezoBridge) CallInitialize(
	arg__tbtcBridge common.Address,
	arg__tbtcVault common.Address,
	arg__tbtcToken common.Address,
	arg__initialSequence *big.Int,
	blockNumber *big.Int,
) error {
	var result interface{} = nil

	err := chainutil.CallAtBlock(
		mb.transactorOptions.From,
		blockNumber, nil,
		mb.contractABI,
		mb.caller,
		mb.errorResolver,
		mb.contractAddress,
		"initialize",
		&result,
		arg__tbtcBridge,
		arg__tbtcVault,
		arg__tbtcToken,
		arg__initialSequence,
	)

	return err
}

func (mb *MezoBridge) InitializeGasEstimate(
	arg__tbtcBridge common.Address,
	arg__tbtcVault common.Address,
	arg__tbtcToken common.Address,
	arg__initialSequence *big.Int,
) (uint64, error) {
	var result uint64

	result, err := chainutil.EstimateGas(
		mb.callerOptions.From,
		mb.contractAddress,
		"initialize",
		mb.contractABI,
		mb.transactor,
		arg__tbtcBridge,
		arg__tbtcVault,
		arg__tbtcToken,
		arg__initialSequence,
	)

	return result, err
}

// Transaction submission.
func (mb *MezoBridge) InitializeBTCBridging(
	arg_fundingTx abi.IBridgeTypesBitcoinTxInfo,
	arg_reveal abi.IBridgeTypesDepositRevealInfo,
	arg_recipient common.Address,

	transactionOptions ...chainutil.TransactionOptions,
) (*types.Transaction, error) {
	mbLogger.Debug(
		"submitting transaction initializeBTCBridging",
		" params: ",
		fmt.Sprint(
			arg_fundingTx,
			arg_reveal,
			arg_recipient,
		),
	)

	mb.transactionMutex.Lock()
	defer mb.transactionMutex.Unlock()

	// create a copy
	transactorOptions := new(bind.TransactOpts)
	*transactorOptions = *mb.transactorOptions

	if len(transactionOptions) > 1 {
		return nil, fmt.Errorf(
			"could not process multiple transaction options sets",
		)
	} else if len(transactionOptions) > 0 {
		transactionOptions[0].Apply(transactorOptions)
	}

	nonce, err := mb.nonceManager.CurrentNonce()
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve account nonce: %v", err)
	}

	transactorOptions.Nonce = new(big.Int).SetUint64(nonce)

	transaction, err := mb.contract.InitializeBTCBridging(
		transactorOptions,
		arg_fundingTx,
		arg_reveal,
		arg_recipient,
	)
	if err != nil {
		return transaction, mb.errorResolver.ResolveError(
			err,
			mb.transactorOptions.From,
			nil,
			"initializeBTCBridging",
			arg_fundingTx,
			arg_reveal,
			arg_recipient,
		)
	}

	mbLogger.Infof(
		"submitted transaction initializeBTCBridging with id: [%s] and nonce [%v]",
		transaction.Hash(),
		transaction.Nonce(),
	)

	go mb.miningWaiter.ForceMining(
		transaction,
		transactorOptions,
		func(newTransactorOptions *bind.TransactOpts) (*types.Transaction, error) {
			// If original transactor options has a non-zero gas limit, that
			// means the client code set it on their own. In that case, we
			// should rewrite the gas limit from the original transaction
			// for each resubmission. If the gas limit is not set by the client
			// code, let the the submitter re-estimate the gas limit on each
			// resubmission.
			if transactorOptions.GasLimit != 0 {
				newTransactorOptions.GasLimit = transactorOptions.GasLimit
			}

			transaction, err := mb.contract.InitializeBTCBridging(
				newTransactorOptions,
				arg_fundingTx,
				arg_reveal,
				arg_recipient,
			)
			if err != nil {
				return nil, mb.errorResolver.ResolveError(
					err,
					mb.transactorOptions.From,
					nil,
					"initializeBTCBridging",
					arg_fundingTx,
					arg_reveal,
					arg_recipient,
				)
			}

			mbLogger.Infof(
				"submitted transaction initializeBTCBridging with id: [%s] and nonce [%v]",
				transaction.Hash(),
				transaction.Nonce(),
			)

			return transaction, nil
		},
	)

	mb.nonceManager.IncrementNonce()

	return transaction, err
}

// Non-mutating call, not a transaction submission.
func (mb *MezoBridge) CallInitializeBTCBridging(
	arg_fundingTx abi.IBridgeTypesBitcoinTxInfo,
	arg_reveal abi.IBridgeTypesDepositRevealInfo,
	arg_recipient common.Address,
	blockNumber *big.Int,
) error {
	var result interface{} = nil

	err := chainutil.CallAtBlock(
		mb.transactorOptions.From,
		blockNumber, nil,
		mb.contractABI,
		mb.caller,
		mb.errorResolver,
		mb.contractAddress,
		"initializeBTCBridging",
		&result,
		arg_fundingTx,
		arg_reveal,
		arg_recipient,
	)

	return err
}

func (mb *MezoBridge) InitializeBTCBridgingGasEstimate(
	arg_fundingTx abi.IBridgeTypesBitcoinTxInfo,
	arg_reveal abi.IBridgeTypesDepositRevealInfo,
	arg_recipient common.Address,
) (uint64, error) {
	var result uint64

	result, err := chainutil.EstimateGas(
		mb.callerOptions.From,
		mb.contractAddress,
		"initializeBTCBridging",
		mb.contractABI,
		mb.transactor,
		arg_fundingTx,
		arg_reveal,
		arg_recipient,
	)

	return result, err
}

// Transaction submission.
func (mb *MezoBridge) RenounceOwnership(

	transactionOptions ...chainutil.TransactionOptions,
) (*types.Transaction, error) {
	mbLogger.Debug(
		"submitting transaction renounceOwnership",
	)

	mb.transactionMutex.Lock()
	defer mb.transactionMutex.Unlock()

	// create a copy
	transactorOptions := new(bind.TransactOpts)
	*transactorOptions = *mb.transactorOptions

	if len(transactionOptions) > 1 {
		return nil, fmt.Errorf(
			"could not process multiple transaction options sets",
		)
	} else if len(transactionOptions) > 0 {
		transactionOptions[0].Apply(transactorOptions)
	}

	nonce, err := mb.nonceManager.CurrentNonce()
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve account nonce: %v", err)
	}

	transactorOptions.Nonce = new(big.Int).SetUint64(nonce)

	transaction, err := mb.contract.RenounceOwnership(
		transactorOptions,
	)
	if err != nil {
		return transaction, mb.errorResolver.ResolveError(
			err,
			mb.transactorOptions.From,
			nil,
			"renounceOwnership",
		)
	}

	mbLogger.Infof(
		"submitted transaction renounceOwnership with id: [%s] and nonce [%v]",
		transaction.Hash(),
		transaction.Nonce(),
	)

	go mb.miningWaiter.ForceMining(
		transaction,
		transactorOptions,
		func(newTransactorOptions *bind.TransactOpts) (*types.Transaction, error) {
			// If original transactor options has a non-zero gas limit, that
			// means the client code set it on their own. In that case, we
			// should rewrite the gas limit from the original transaction
			// for each resubmission. If the gas limit is not set by the client
			// code, let the the submitter re-estimate the gas limit on each
			// resubmission.
			if transactorOptions.GasLimit != 0 {
				newTransactorOptions.GasLimit = transactorOptions.GasLimit
			}

			transaction, err := mb.contract.RenounceOwnership(
				newTransactorOptions,
			)
			if err != nil {
				return nil, mb.errorResolver.ResolveError(
					err,
					mb.transactorOptions.From,
					nil,
					"renounceOwnership",
				)
			}

			mbLogger.Infof(
				"submitted transaction renounceOwnership with id: [%s] and nonce [%v]",
				transaction.Hash(),
				transaction.Nonce(),
			)

			return transaction, nil
		},
	)

	mb.nonceManager.IncrementNonce()

	return transaction, err
}

// Non-mutating call, not a transaction submission.
func (mb *MezoBridge) CallRenounceOwnership(
	blockNumber *big.Int,
) error {
	var result interface{} = nil

	err := chainutil.CallAtBlock(
		mb.transactorOptions.From,
		blockNumber, nil,
		mb.contractABI,
		mb.caller,
		mb.errorResolver,
		mb.contractAddress,
		"renounceOwnership",
		&result,
	)

	return err
}

func (mb *MezoBridge) RenounceOwnershipGasEstimate() (uint64, error) {
	var result uint64

	result, err := chainutil.EstimateGas(
		mb.callerOptions.From,
		mb.contractAddress,
		"renounceOwnership",
		mb.contractABI,
		mb.transactor,
	)

	return result, err
}

// Transaction submission.
func (mb *MezoBridge) TransferOwnership(
	arg_newOwner common.Address,

	transactionOptions ...chainutil.TransactionOptions,
) (*types.Transaction, error) {
	mbLogger.Debug(
		"submitting transaction transferOwnership",
		" params: ",
		fmt.Sprint(
			arg_newOwner,
		),
	)

	mb.transactionMutex.Lock()
	defer mb.transactionMutex.Unlock()

	// create a copy
	transactorOptions := new(bind.TransactOpts)
	*transactorOptions = *mb.transactorOptions

	if len(transactionOptions) > 1 {
		return nil, fmt.Errorf(
			"could not process multiple transaction options sets",
		)
	} else if len(transactionOptions) > 0 {
		transactionOptions[0].Apply(transactorOptions)
	}

	nonce, err := mb.nonceManager.CurrentNonce()
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve account nonce: %v", err)
	}

	transactorOptions.Nonce = new(big.Int).SetUint64(nonce)

	transaction, err := mb.contract.TransferOwnership(
		transactorOptions,
		arg_newOwner,
	)
	if err != nil {
		return transaction, mb.errorResolver.ResolveError(
			err,
			mb.transactorOptions.From,
			nil,
			"transferOwnership",
			arg_newOwner,
		)
	}

	mbLogger.Infof(
		"submitted transaction transferOwnership with id: [%s] and nonce [%v]",
		transaction.Hash(),
		transaction.Nonce(),
	)

	go mb.miningWaiter.ForceMining(
		transaction,
		transactorOptions,
		func(newTransactorOptions *bind.TransactOpts) (*types.Transaction, error) {
			// If original transactor options has a non-zero gas limit, that
			// means the client code set it on their own. In that case, we
			// should rewrite the gas limit from the original transaction
			// for each resubmission. If the gas limit is not set by the client
			// code, let the the submitter re-estimate the gas limit on each
			// resubmission.
			if transactorOptions.GasLimit != 0 {
				newTransactorOptions.GasLimit = transactorOptions.GasLimit
			}

			transaction, err := mb.contract.TransferOwnership(
				newTransactorOptions,
				arg_newOwner,
			)
			if err != nil {
				return nil, mb.errorResolver.ResolveError(
					err,
					mb.transactorOptions.From,
					nil,
					"transferOwnership",
					arg_newOwner,
				)
			}

			mbLogger.Infof(
				"submitted transaction transferOwnership with id: [%s] and nonce [%v]",
				transaction.Hash(),
				transaction.Nonce(),
			)

			return transaction, nil
		},
	)

	mb.nonceManager.IncrementNonce()

	return transaction, err
}

// Non-mutating call, not a transaction submission.
func (mb *MezoBridge) CallTransferOwnership(
	arg_newOwner common.Address,
	blockNumber *big.Int,
) error {
	var result interface{} = nil

	err := chainutil.CallAtBlock(
		mb.transactorOptions.From,
		blockNumber, nil,
		mb.contractABI,
		mb.caller,
		mb.errorResolver,
		mb.contractAddress,
		"transferOwnership",
		&result,
		arg_newOwner,
	)

	return err
}

func (mb *MezoBridge) TransferOwnershipGasEstimate(
	arg_newOwner common.Address,
) (uint64, error) {
	var result uint64

	result, err := chainutil.EstimateGas(
		mb.callerOptions.From,
		mb.contractAddress,
		"transferOwnership",
		mb.contractABI,
		mb.transactor,
		arg_newOwner,
	)

	return result, err
}

// Transaction submission.
func (mb *MezoBridge) UpdateMinERC20Amount(
	arg_ERC20Token common.Address,
	arg_newMinERC20Amount *big.Int,

	transactionOptions ...chainutil.TransactionOptions,
) (*types.Transaction, error) {
	mbLogger.Debug(
		"submitting transaction updateMinERC20Amount",
		" params: ",
		fmt.Sprint(
			arg_ERC20Token,
			arg_newMinERC20Amount,
		),
	)

	mb.transactionMutex.Lock()
	defer mb.transactionMutex.Unlock()

	// create a copy
	transactorOptions := new(bind.TransactOpts)
	*transactorOptions = *mb.transactorOptions

	if len(transactionOptions) > 1 {
		return nil, fmt.Errorf(
			"could not process multiple transaction options sets",
		)
	} else if len(transactionOptions) > 0 {
		transactionOptions[0].Apply(transactorOptions)
	}

	nonce, err := mb.nonceManager.CurrentNonce()
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve account nonce: %v", err)
	}

	transactorOptions.Nonce = new(big.Int).SetUint64(nonce)

	transaction, err := mb.contract.UpdateMinERC20Amount(
		transactorOptions,
		arg_ERC20Token,
		arg_newMinERC20Amount,
	)
	if err != nil {
		return transaction, mb.errorResolver.ResolveError(
			err,
			mb.transactorOptions.From,
			nil,
			"updateMinERC20Amount",
			arg_ERC20Token,
			arg_newMinERC20Amount,
		)
	}

	mbLogger.Infof(
		"submitted transaction updateMinERC20Amount with id: [%s] and nonce [%v]",
		transaction.Hash(),
		transaction.Nonce(),
	)

	go mb.miningWaiter.ForceMining(
		transaction,
		transactorOptions,
		func(newTransactorOptions *bind.TransactOpts) (*types.Transaction, error) {
			// If original transactor options has a non-zero gas limit, that
			// means the client code set it on their own. In that case, we
			// should rewrite the gas limit from the original transaction
			// for each resubmission. If the gas limit is not set by the client
			// code, let the the submitter re-estimate the gas limit on each
			// resubmission.
			if transactorOptions.GasLimit != 0 {
				newTransactorOptions.GasLimit = transactorOptions.GasLimit
			}

			transaction, err := mb.contract.UpdateMinERC20Amount(
				newTransactorOptions,
				arg_ERC20Token,
				arg_newMinERC20Amount,
			)
			if err != nil {
				return nil, mb.errorResolver.ResolveError(
					err,
					mb.transactorOptions.From,
					nil,
					"updateMinERC20Amount",
					arg_ERC20Token,
					arg_newMinERC20Amount,
				)
			}

			mbLogger.Infof(
				"submitted transaction updateMinERC20Amount with id: [%s] and nonce [%v]",
				transaction.Hash(),
				transaction.Nonce(),
			)

			return transaction, nil
		},
	)

	mb.nonceManager.IncrementNonce()

	return transaction, err
}

// Non-mutating call, not a transaction submission.
func (mb *MezoBridge) CallUpdateMinERC20Amount(
	arg_ERC20Token common.Address,
	arg_newMinERC20Amount *big.Int,
	blockNumber *big.Int,
) error {
	var result interface{} = nil

	err := chainutil.CallAtBlock(
		mb.transactorOptions.From,
		blockNumber, nil,
		mb.contractABI,
		mb.caller,
		mb.errorResolver,
		mb.contractAddress,
		"updateMinERC20Amount",
		&result,
		arg_ERC20Token,
		arg_newMinERC20Amount,
	)

	return err
}

func (mb *MezoBridge) UpdateMinERC20AmountGasEstimate(
	arg_ERC20Token common.Address,
	arg_newMinERC20Amount *big.Int,
) (uint64, error) {
	var result uint64

	result, err := chainutil.EstimateGas(
		mb.callerOptions.From,
		mb.contractAddress,
		"updateMinERC20Amount",
		mb.contractABI,
		mb.transactor,
		arg_ERC20Token,
		arg_newMinERC20Amount,
	)

	return result, err
}

// Transaction submission.
func (mb *MezoBridge) UpdateMinTBTCAmount(
	arg_newMinTBTCAmount *big.Int,

	transactionOptions ...chainutil.TransactionOptions,
) (*types.Transaction, error) {
	mbLogger.Debug(
		"submitting transaction updateMinTBTCAmount",
		" params: ",
		fmt.Sprint(
			arg_newMinTBTCAmount,
		),
	)

	mb.transactionMutex.Lock()
	defer mb.transactionMutex.Unlock()

	// create a copy
	transactorOptions := new(bind.TransactOpts)
	*transactorOptions = *mb.transactorOptions

	if len(transactionOptions) > 1 {
		return nil, fmt.Errorf(
			"could not process multiple transaction options sets",
		)
	} else if len(transactionOptions) > 0 {
		transactionOptions[0].Apply(transactorOptions)
	}

	nonce, err := mb.nonceManager.CurrentNonce()
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve account nonce: %v", err)
	}

	transactorOptions.Nonce = new(big.Int).SetUint64(nonce)

	transaction, err := mb.contract.UpdateMinTBTCAmount(
		transactorOptions,
		arg_newMinTBTCAmount,
	)
	if err != nil {
		return transaction, mb.errorResolver.ResolveError(
			err,
			mb.transactorOptions.From,
			nil,
			"updateMinTBTCAmount",
			arg_newMinTBTCAmount,
		)
	}

	mbLogger.Infof(
		"submitted transaction updateMinTBTCAmount with id: [%s] and nonce [%v]",
		transaction.Hash(),
		transaction.Nonce(),
	)

	go mb.miningWaiter.ForceMining(
		transaction,
		transactorOptions,
		func(newTransactorOptions *bind.TransactOpts) (*types.Transaction, error) {
			// If original transactor options has a non-zero gas limit, that
			// means the client code set it on their own. In that case, we
			// should rewrite the gas limit from the original transaction
			// for each resubmission. If the gas limit is not set by the client
			// code, let the the submitter re-estimate the gas limit on each
			// resubmission.
			if transactorOptions.GasLimit != 0 {
				newTransactorOptions.GasLimit = transactorOptions.GasLimit
			}

			transaction, err := mb.contract.UpdateMinTBTCAmount(
				newTransactorOptions,
				arg_newMinTBTCAmount,
			)
			if err != nil {
				return nil, mb.errorResolver.ResolveError(
					err,
					mb.transactorOptions.From,
					nil,
					"updateMinTBTCAmount",
					arg_newMinTBTCAmount,
				)
			}

			mbLogger.Infof(
				"submitted transaction updateMinTBTCAmount with id: [%s] and nonce [%v]",
				transaction.Hash(),
				transaction.Nonce(),
			)

			return transaction, nil
		},
	)

	mb.nonceManager.IncrementNonce()

	return transaction, err
}

// Non-mutating call, not a transaction submission.
func (mb *MezoBridge) CallUpdateMinTBTCAmount(
	arg_newMinTBTCAmount *big.Int,
	blockNumber *big.Int,
) error {
	var result interface{} = nil

	err := chainutil.CallAtBlock(
		mb.transactorOptions.From,
		blockNumber, nil,
		mb.contractABI,
		mb.caller,
		mb.errorResolver,
		mb.contractAddress,
		"updateMinTBTCAmount",
		&result,
		arg_newMinTBTCAmount,
	)

	return err
}

func (mb *MezoBridge) UpdateMinTBTCAmountGasEstimate(
	arg_newMinTBTCAmount *big.Int,
) (uint64, error) {
	var result uint64

	result, err := chainutil.EstimateGas(
		mb.callerOptions.From,
		mb.contractAddress,
		"updateMinTBTCAmount",
		mb.contractABI,
		mb.transactor,
		arg_newMinTBTCAmount,
	)

	return result, err
}

// ----- Const Methods ------

func (mb *MezoBridge) Bridge() (common.Address, error) {
	result, err := mb.contract.Bridge(
		mb.callerOptions,
	)

	if err != nil {
		return result, mb.errorResolver.ResolveError(
			err,
			mb.callerOptions.From,
			nil,
			"bridge",
		)
	}

	return result, err
}

func (mb *MezoBridge) BridgeAtBlock(
	blockNumber *big.Int,
) (common.Address, error) {
	var result common.Address

	err := chainutil.CallAtBlock(
		mb.callerOptions.From,
		blockNumber,
		nil,
		mb.contractABI,
		mb.caller,
		mb.errorResolver,
		mb.contractAddress,
		"bridge",
		&result,
	)

	return result, err
}

func (mb *MezoBridge) BtcDeposits(
	arg0 *big.Int,
) (uint8, error) {
	result, err := mb.contract.BtcDeposits(
		mb.callerOptions,
		arg0,
	)

	if err != nil {
		return result, mb.errorResolver.ResolveError(
			err,
			mb.callerOptions.From,
			nil,
			"btcDeposits",
			arg0,
		)
	}

	return result, err
}

func (mb *MezoBridge) BtcDepositsAtBlock(
	arg0 *big.Int,
	blockNumber *big.Int,
) (uint8, error) {
	var result uint8

	err := chainutil.CallAtBlock(
		mb.callerOptions.From,
		blockNumber,
		nil,
		mb.contractABI,
		mb.caller,
		mb.errorResolver,
		mb.contractAddress,
		"btcDeposits",
		&result,
		arg0,
	)

	return result, err
}

func (mb *MezoBridge) ERC20Tokens(
	arg0 common.Address,
) (*big.Int, error) {
	result, err := mb.contract.ERC20Tokens(
		mb.callerOptions,
		arg0,
	)

	if err != nil {
		return result, mb.errorResolver.ResolveError(
			err,
			mb.callerOptions.From,
			nil,
			"eRC20Tokens",
			arg0,
		)
	}

	return result, err
}

func (mb *MezoBridge) ERC20TokensAtBlock(
	arg0 common.Address,
	blockNumber *big.Int,
) (*big.Int, error) {
	var result *big.Int

	err := chainutil.CallAtBlock(
		mb.callerOptions.From,
		blockNumber,
		nil,
		mb.contractABI,
		mb.caller,
		mb.errorResolver,
		mb.contractAddress,
		"eRC20Tokens",
		&result,
		arg0,
	)

	return result, err
}

func (mb *MezoBridge) ERC20TokensCount() (*big.Int, error) {
	result, err := mb.contract.ERC20TokensCount(
		mb.callerOptions,
	)

	if err != nil {
		return result, mb.errorResolver.ResolveError(
			err,
			mb.callerOptions.From,
			nil,
			"eRC20TokensCount",
		)
	}

	return result, err
}

func (mb *MezoBridge) ERC20TokensCountAtBlock(
	blockNumber *big.Int,
) (*big.Int, error) {
	var result *big.Int

	err := chainutil.CallAtBlock(
		mb.callerOptions.From,
		blockNumber,
		nil,
		mb.contractABI,
		mb.caller,
		mb.errorResolver,
		mb.contractAddress,
		"eRC20TokensCount",
		&result,
	)

	return result, err
}

func (mb *MezoBridge) MAXERC20TOKENS() (*big.Int, error) {
	result, err := mb.contract.MAXERC20TOKENS(
		mb.callerOptions,
	)

	if err != nil {
		return result, mb.errorResolver.ResolveError(
			err,
			mb.callerOptions.From,
			nil,
			"mAXERC20TOKENS",
		)
	}

	return result, err
}

func (mb *MezoBridge) MAXERC20TOKENSAtBlock(
	blockNumber *big.Int,
) (*big.Int, error) {
	var result *big.Int

	err := chainutil.CallAtBlock(
		mb.callerOptions.From,
		blockNumber,
		nil,
		mb.contractABI,
		mb.caller,
		mb.errorResolver,
		mb.contractAddress,
		"mAXERC20TOKENS",
		&result,
	)

	return result, err
}

func (mb *MezoBridge) MinTBTCAmount() (*big.Int, error) {
	result, err := mb.contract.MinTBTCAmount(
		mb.callerOptions,
	)

	if err != nil {
		return result, mb.errorResolver.ResolveError(
			err,
			mb.callerOptions.From,
			nil,
			"minTBTCAmount",
		)
	}

	return result, err
}

func (mb *MezoBridge) MinTBTCAmountAtBlock(
	blockNumber *big.Int,
) (*big.Int, error) {
	var result *big.Int

	err := chainutil.CallAtBlock(
		mb.callerOptions.From,
		blockNumber,
		nil,
		mb.contractABI,
		mb.caller,
		mb.errorResolver,
		mb.contractAddress,
		"minTBTCAmount",
		&result,
	)

	return result, err
}

func (mb *MezoBridge) Owner() (common.Address, error) {
	result, err := mb.contract.Owner(
		mb.callerOptions,
	)

	if err != nil {
		return result, mb.errorResolver.ResolveError(
			err,
			mb.callerOptions.From,
			nil,
			"owner",
		)
	}

	return result, err
}

func (mb *MezoBridge) OwnerAtBlock(
	blockNumber *big.Int,
) (common.Address, error) {
	var result common.Address

	err := chainutil.CallAtBlock(
		mb.callerOptions.From,
		blockNumber,
		nil,
		mb.contractABI,
		mb.caller,
		mb.errorResolver,
		mb.contractAddress,
		"owner",
		&result,
	)

	return result, err
}

func (mb *MezoBridge) PendingOwner() (common.Address, error) {
	result, err := mb.contract.PendingOwner(
		mb.callerOptions,
	)

	if err != nil {
		return result, mb.errorResolver.ResolveError(
			err,
			mb.callerOptions.From,
			nil,
			"pendingOwner",
		)
	}

	return result, err
}

func (mb *MezoBridge) PendingOwnerAtBlock(
	blockNumber *big.Int,
) (common.Address, error) {
	var result common.Address

	err := chainutil.CallAtBlock(
		mb.callerOptions.From,
		blockNumber,
		nil,
		mb.contractABI,
		mb.caller,
		mb.errorResolver,
		mb.contractAddress,
		"pendingOwner",
		&result,
	)

	return result, err
}

func (mb *MezoBridge) SATOSHIMULTIPLIER() (*big.Int, error) {
	result, err := mb.contract.SATOSHIMULTIPLIER(
		mb.callerOptions,
	)

	if err != nil {
		return result, mb.errorResolver.ResolveError(
			err,
			mb.callerOptions.From,
			nil,
			"sATOSHIMULTIPLIER",
		)
	}

	return result, err
}

func (mb *MezoBridge) SATOSHIMULTIPLIERAtBlock(
	blockNumber *big.Int,
) (*big.Int, error) {
	var result *big.Int

	err := chainutil.CallAtBlock(
		mb.callerOptions.From,
		blockNumber,
		nil,
		mb.contractABI,
		mb.caller,
		mb.errorResolver,
		mb.contractAddress,
		"sATOSHIMULTIPLIER",
		&result,
	)

	return result, err
}

func (mb *MezoBridge) Sequence() (*big.Int, error) {
	result, err := mb.contract.Sequence(
		mb.callerOptions,
	)

	if err != nil {
		return result, mb.errorResolver.ResolveError(
			err,
			mb.callerOptions.From,
			nil,
			"sequence",
		)
	}

	return result, err
}

func (mb *MezoBridge) SequenceAtBlock(
	blockNumber *big.Int,
) (*big.Int, error) {
	var result *big.Int

	err := chainutil.CallAtBlock(
		mb.callerOptions.From,
		blockNumber,
		nil,
		mb.contractABI,
		mb.caller,
		mb.errorResolver,
		mb.contractAddress,
		"sequence",
		&result,
	)

	return result, err
}

func (mb *MezoBridge) TbtcToken() (common.Address, error) {
	result, err := mb.contract.TbtcToken(
		mb.callerOptions,
	)

	if err != nil {
		return result, mb.errorResolver.ResolveError(
			err,
			mb.callerOptions.From,
			nil,
			"tbtcToken",
		)
	}

	return result, err
}

func (mb *MezoBridge) TbtcTokenAtBlock(
	blockNumber *big.Int,
) (common.Address, error) {
	var result common.Address

	err := chainutil.CallAtBlock(
		mb.callerOptions.From,
		blockNumber,
		nil,
		mb.contractABI,
		mb.caller,
		mb.errorResolver,
		mb.contractAddress,
		"tbtcToken",
		&result,
	)

	return result, err
}

func (mb *MezoBridge) TbtcVault() (common.Address, error) {
	result, err := mb.contract.TbtcVault(
		mb.callerOptions,
	)

	if err != nil {
		return result, mb.errorResolver.ResolveError(
			err,
			mb.callerOptions.From,
			nil,
			"tbtcVault",
		)
	}

	return result, err
}

func (mb *MezoBridge) TbtcVaultAtBlock(
	blockNumber *big.Int,
) (common.Address, error) {
	var result common.Address

	err := chainutil.CallAtBlock(
		mb.callerOptions.From,
		blockNumber,
		nil,
		mb.contractABI,
		mb.caller,
		mb.errorResolver,
		mb.contractAddress,
		"tbtcVault",
		&result,
	)

	return result, err
}

// ------ Events -------

func (mb *MezoBridge) AssetsLockedEvent(
	opts *ethereum.SubscribeOpts,
	sequenceNumberFilter []*big.Int,
	recipientFilter []common.Address,
	tokenFilter []common.Address,
) *MbAssetsLockedSubscription {
	if opts == nil {
		opts = new(ethereum.SubscribeOpts)
	}
	if opts.Tick == 0 {
		opts.Tick = chainutil.DefaultSubscribeOptsTick
	}
	if opts.PastBlocks == 0 {
		opts.PastBlocks = chainutil.DefaultSubscribeOptsPastBlocks
	}

	return &MbAssetsLockedSubscription{
		mb,
		opts,
		sequenceNumberFilter,
		recipientFilter,
		tokenFilter,
	}
}

type MbAssetsLockedSubscription struct {
	contract             *MezoBridge
	opts                 *ethereum.SubscribeOpts
	sequenceNumberFilter []*big.Int
	recipientFilter      []common.Address
	tokenFilter          []common.Address
}

type mezoBridgeAssetsLockedFunc func(
	SequenceNumber *big.Int,
	Recipient common.Address,
	Token common.Address,
	Amount *big.Int,
	blockNumber uint64,
)

func (als *MbAssetsLockedSubscription) OnEvent(
	handler mezoBridgeAssetsLockedFunc,
) subscription.EventSubscription {
	eventChan := make(chan *abi.MezoBridgeAssetsLocked)
	ctx, cancelCtx := context.WithCancel(context.Background())

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case event := <-eventChan:
				handler(
					event.SequenceNumber,
					event.Recipient,
					event.Token,
					event.Amount,
					event.Raw.BlockNumber,
				)
			}
		}
	}()

	sub := als.Pipe(eventChan)
	return subscription.NewEventSubscription(func() {
		sub.Unsubscribe()
		cancelCtx()
	})
}

func (als *MbAssetsLockedSubscription) Pipe(
	sink chan *abi.MezoBridgeAssetsLocked,
) subscription.EventSubscription {
	ctx, cancelCtx := context.WithCancel(context.Background())
	go func() {
		ticker := time.NewTicker(als.opts.Tick)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				lastBlock, err := als.contract.blockCounter.CurrentBlock()
				if err != nil {
					mbLogger.Errorf(
						"subscription failed to pull events: [%v]",
						err,
					)
				}
				fromBlock := lastBlock - als.opts.PastBlocks

				mbLogger.Infof(
					"subscription monitoring fetching past AssetsLocked events "+
						"starting from block [%v]",
					fromBlock,
				)
				events, err := als.contract.PastAssetsLockedEvents(
					fromBlock,
					nil,
					als.sequenceNumberFilter,
					als.recipientFilter,
					als.tokenFilter,
				)
				if err != nil {
					mbLogger.Errorf(
						"subscription failed to pull events: [%v]",
						err,
					)
					continue
				}
				mbLogger.Infof(
					"subscription monitoring fetched [%v] past AssetsLocked events",
					len(events),
				)

				for _, event := range events {
					sink <- event
				}
			}
		}
	}()

	sub := als.contract.watchAssetsLocked(
		sink,
		als.sequenceNumberFilter,
		als.recipientFilter,
		als.tokenFilter,
	)

	return subscription.NewEventSubscription(func() {
		sub.Unsubscribe()
		cancelCtx()
	})
}

func (mb *MezoBridge) watchAssetsLocked(
	sink chan *abi.MezoBridgeAssetsLocked,
	sequenceNumberFilter []*big.Int,
	recipientFilter []common.Address,
	tokenFilter []common.Address,
) event.Subscription {
	subscribeFn := func(ctx context.Context) (event.Subscription, error) {
		return mb.contract.WatchAssetsLocked(
			&bind.WatchOpts{Context: ctx},
			sink,
			sequenceNumberFilter,
			recipientFilter,
			tokenFilter,
		)
	}

	thresholdViolatedFn := func(elapsed time.Duration) {
		mbLogger.Warnf(
			"subscription to event AssetsLocked had to be "+
				"retried [%s] since the last attempt; please inspect "+
				"host chain connectivity",
			elapsed,
		)
	}

	subscriptionFailedFn := func(err error) {
		mbLogger.Errorf(
			"subscription to event AssetsLocked failed "+
				"with error: [%v]; resubscription attempt will be "+
				"performed",
			err,
		)
	}

	return chainutil.WithResubscription(
		chainutil.SubscriptionBackoffMax,
		subscribeFn,
		chainutil.SubscriptionAlertThreshold,
		thresholdViolatedFn,
		subscriptionFailedFn,
	)
}

func (mb *MezoBridge) PastAssetsLockedEvents(
	startBlock uint64,
	endBlock *uint64,
	sequenceNumberFilter []*big.Int,
	recipientFilter []common.Address,
	tokenFilter []common.Address,
) ([]*abi.MezoBridgeAssetsLocked, error) {
	iterator, err := mb.contract.FilterAssetsLocked(
		&bind.FilterOpts{
			Start: startBlock,
			End:   endBlock,
		},
		sequenceNumberFilter,
		recipientFilter,
		tokenFilter,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"error retrieving past AssetsLocked events: [%v]",
			err,
		)
	}

	events := make([]*abi.MezoBridgeAssetsLocked, 0)

	for iterator.Next() {
		event := iterator.Event
		events = append(events, event)
	}

	return events, nil
}

func (mb *MezoBridge) BTCDepositFinalizedEvent(
	opts *ethereum.SubscribeOpts,
	btcDepositKeyFilter []*big.Int,
) *MbBTCDepositFinalizedSubscription {
	if opts == nil {
		opts = new(ethereum.SubscribeOpts)
	}
	if opts.Tick == 0 {
		opts.Tick = chainutil.DefaultSubscribeOptsTick
	}
	if opts.PastBlocks == 0 {
		opts.PastBlocks = chainutil.DefaultSubscribeOptsPastBlocks
	}

	return &MbBTCDepositFinalizedSubscription{
		mb,
		opts,
		btcDepositKeyFilter,
	}
}

type MbBTCDepositFinalizedSubscription struct {
	contract            *MezoBridge
	opts                *ethereum.SubscribeOpts
	btcDepositKeyFilter []*big.Int
}

type mezoBridgeBTCDepositFinalizedFunc func(
	BtcDepositKey *big.Int,
	InitialAmount *big.Int,
	TbtcAmount *big.Int,
	blockNumber uint64,
)

func (btcdfs *MbBTCDepositFinalizedSubscription) OnEvent(
	handler mezoBridgeBTCDepositFinalizedFunc,
) subscription.EventSubscription {
	eventChan := make(chan *abi.MezoBridgeBTCDepositFinalized)
	ctx, cancelCtx := context.WithCancel(context.Background())

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case event := <-eventChan:
				handler(
					event.BtcDepositKey,
					event.InitialAmount,
					event.TbtcAmount,
					event.Raw.BlockNumber,
				)
			}
		}
	}()

	sub := btcdfs.Pipe(eventChan)
	return subscription.NewEventSubscription(func() {
		sub.Unsubscribe()
		cancelCtx()
	})
}

func (btcdfs *MbBTCDepositFinalizedSubscription) Pipe(
	sink chan *abi.MezoBridgeBTCDepositFinalized,
) subscription.EventSubscription {
	ctx, cancelCtx := context.WithCancel(context.Background())
	go func() {
		ticker := time.NewTicker(btcdfs.opts.Tick)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				lastBlock, err := btcdfs.contract.blockCounter.CurrentBlock()
				if err != nil {
					mbLogger.Errorf(
						"subscription failed to pull events: [%v]",
						err,
					)
				}
				fromBlock := lastBlock - btcdfs.opts.PastBlocks

				mbLogger.Infof(
					"subscription monitoring fetching past BTCDepositFinalized events "+
						"starting from block [%v]",
					fromBlock,
				)
				events, err := btcdfs.contract.PastBTCDepositFinalizedEvents(
					fromBlock,
					nil,
					btcdfs.btcDepositKeyFilter,
				)
				if err != nil {
					mbLogger.Errorf(
						"subscription failed to pull events: [%v]",
						err,
					)
					continue
				}
				mbLogger.Infof(
					"subscription monitoring fetched [%v] past BTCDepositFinalized events",
					len(events),
				)

				for _, event := range events {
					sink <- event
				}
			}
		}
	}()

	sub := btcdfs.contract.watchBTCDepositFinalized(
		sink,
		btcdfs.btcDepositKeyFilter,
	)

	return subscription.NewEventSubscription(func() {
		sub.Unsubscribe()
		cancelCtx()
	})
}

func (mb *MezoBridge) watchBTCDepositFinalized(
	sink chan *abi.MezoBridgeBTCDepositFinalized,
	btcDepositKeyFilter []*big.Int,
) event.Subscription {
	subscribeFn := func(ctx context.Context) (event.Subscription, error) {
		return mb.contract.WatchBTCDepositFinalized(
			&bind.WatchOpts{Context: ctx},
			sink,
			btcDepositKeyFilter,
		)
	}

	thresholdViolatedFn := func(elapsed time.Duration) {
		mbLogger.Warnf(
			"subscription to event BTCDepositFinalized had to be "+
				"retried [%s] since the last attempt; please inspect "+
				"host chain connectivity",
			elapsed,
		)
	}

	subscriptionFailedFn := func(err error) {
		mbLogger.Errorf(
			"subscription to event BTCDepositFinalized failed "+
				"with error: [%v]; resubscription attempt will be "+
				"performed",
			err,
		)
	}

	return chainutil.WithResubscription(
		chainutil.SubscriptionBackoffMax,
		subscribeFn,
		chainutil.SubscriptionAlertThreshold,
		thresholdViolatedFn,
		subscriptionFailedFn,
	)
}

func (mb *MezoBridge) PastBTCDepositFinalizedEvents(
	startBlock uint64,
	endBlock *uint64,
	btcDepositKeyFilter []*big.Int,
) ([]*abi.MezoBridgeBTCDepositFinalized, error) {
	iterator, err := mb.contract.FilterBTCDepositFinalized(
		&bind.FilterOpts{
			Start: startBlock,
			End:   endBlock,
		},
		btcDepositKeyFilter,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"error retrieving past BTCDepositFinalized events: [%v]",
			err,
		)
	}

	events := make([]*abi.MezoBridgeBTCDepositFinalized, 0)

	for iterator.Next() {
		event := iterator.Event
		events = append(events, event)
	}

	return events, nil
}

func (mb *MezoBridge) BTCDepositInitializedEvent(
	opts *ethereum.SubscribeOpts,
	btcDepositKeyFilter []*big.Int,
	recipientFilter []common.Address,
) *MbBTCDepositInitializedSubscription {
	if opts == nil {
		opts = new(ethereum.SubscribeOpts)
	}
	if opts.Tick == 0 {
		opts.Tick = chainutil.DefaultSubscribeOptsTick
	}
	if opts.PastBlocks == 0 {
		opts.PastBlocks = chainutil.DefaultSubscribeOptsPastBlocks
	}

	return &MbBTCDepositInitializedSubscription{
		mb,
		opts,
		btcDepositKeyFilter,
		recipientFilter,
	}
}

type MbBTCDepositInitializedSubscription struct {
	contract            *MezoBridge
	opts                *ethereum.SubscribeOpts
	btcDepositKeyFilter []*big.Int
	recipientFilter     []common.Address
}

type mezoBridgeBTCDepositInitializedFunc func(
	BtcDepositKey *big.Int,
	Recipient common.Address,
	blockNumber uint64,
)

func (btcdis *MbBTCDepositInitializedSubscription) OnEvent(
	handler mezoBridgeBTCDepositInitializedFunc,
) subscription.EventSubscription {
	eventChan := make(chan *abi.MezoBridgeBTCDepositInitialized)
	ctx, cancelCtx := context.WithCancel(context.Background())

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case event := <-eventChan:
				handler(
					event.BtcDepositKey,
					event.Recipient,
					event.Raw.BlockNumber,
				)
			}
		}
	}()

	sub := btcdis.Pipe(eventChan)
	return subscription.NewEventSubscription(func() {
		sub.Unsubscribe()
		cancelCtx()
	})
}

func (btcdis *MbBTCDepositInitializedSubscription) Pipe(
	sink chan *abi.MezoBridgeBTCDepositInitialized,
) subscription.EventSubscription {
	ctx, cancelCtx := context.WithCancel(context.Background())
	go func() {
		ticker := time.NewTicker(btcdis.opts.Tick)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				lastBlock, err := btcdis.contract.blockCounter.CurrentBlock()
				if err != nil {
					mbLogger.Errorf(
						"subscription failed to pull events: [%v]",
						err,
					)
				}
				fromBlock := lastBlock - btcdis.opts.PastBlocks

				mbLogger.Infof(
					"subscription monitoring fetching past BTCDepositInitialized events "+
						"starting from block [%v]",
					fromBlock,
				)
				events, err := btcdis.contract.PastBTCDepositInitializedEvents(
					fromBlock,
					nil,
					btcdis.btcDepositKeyFilter,
					btcdis.recipientFilter,
				)
				if err != nil {
					mbLogger.Errorf(
						"subscription failed to pull events: [%v]",
						err,
					)
					continue
				}
				mbLogger.Infof(
					"subscription monitoring fetched [%v] past BTCDepositInitialized events",
					len(events),
				)

				for _, event := range events {
					sink <- event
				}
			}
		}
	}()

	sub := btcdis.contract.watchBTCDepositInitialized(
		sink,
		btcdis.btcDepositKeyFilter,
		btcdis.recipientFilter,
	)

	return subscription.NewEventSubscription(func() {
		sub.Unsubscribe()
		cancelCtx()
	})
}

func (mb *MezoBridge) watchBTCDepositInitialized(
	sink chan *abi.MezoBridgeBTCDepositInitialized,
	btcDepositKeyFilter []*big.Int,
	recipientFilter []common.Address,
) event.Subscription {
	subscribeFn := func(ctx context.Context) (event.Subscription, error) {
		return mb.contract.WatchBTCDepositInitialized(
			&bind.WatchOpts{Context: ctx},
			sink,
			btcDepositKeyFilter,
			recipientFilter,
		)
	}

	thresholdViolatedFn := func(elapsed time.Duration) {
		mbLogger.Warnf(
			"subscription to event BTCDepositInitialized had to be "+
				"retried [%s] since the last attempt; please inspect "+
				"host chain connectivity",
			elapsed,
		)
	}

	subscriptionFailedFn := func(err error) {
		mbLogger.Errorf(
			"subscription to event BTCDepositInitialized failed "+
				"with error: [%v]; resubscription attempt will be "+
				"performed",
			err,
		)
	}

	return chainutil.WithResubscription(
		chainutil.SubscriptionBackoffMax,
		subscribeFn,
		chainutil.SubscriptionAlertThreshold,
		thresholdViolatedFn,
		subscriptionFailedFn,
	)
}

func (mb *MezoBridge) PastBTCDepositInitializedEvents(
	startBlock uint64,
	endBlock *uint64,
	btcDepositKeyFilter []*big.Int,
	recipientFilter []common.Address,
) ([]*abi.MezoBridgeBTCDepositInitialized, error) {
	iterator, err := mb.contract.FilterBTCDepositInitialized(
		&bind.FilterOpts{
			Start: startBlock,
			End:   endBlock,
		},
		btcDepositKeyFilter,
		recipientFilter,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"error retrieving past BTCDepositInitialized events: [%v]",
			err,
		)
	}

	events := make([]*abi.MezoBridgeBTCDepositInitialized, 0)

	for iterator.Next() {
		event := iterator.Event
		events = append(events, event)
	}

	return events, nil
}

func (mb *MezoBridge) ERC20TokenDisabledEvent(
	opts *ethereum.SubscribeOpts,
	ERC20TokenFilter []common.Address,
) *MbERC20TokenDisabledSubscription {
	if opts == nil {
		opts = new(ethereum.SubscribeOpts)
	}
	if opts.Tick == 0 {
		opts.Tick = chainutil.DefaultSubscribeOptsTick
	}
	if opts.PastBlocks == 0 {
		opts.PastBlocks = chainutil.DefaultSubscribeOptsPastBlocks
	}

	return &MbERC20TokenDisabledSubscription{
		mb,
		opts,
		ERC20TokenFilter,
	}
}

type MbERC20TokenDisabledSubscription struct {
	contract         *MezoBridge
	opts             *ethereum.SubscribeOpts
	ERC20TokenFilter []common.Address
}

type mezoBridgeERC20TokenDisabledFunc func(
	ERC20Token common.Address,
	blockNumber uint64,
)

func (erctds *MbERC20TokenDisabledSubscription) OnEvent(
	handler mezoBridgeERC20TokenDisabledFunc,
) subscription.EventSubscription {
	eventChan := make(chan *abi.MezoBridgeERC20TokenDisabled)
	ctx, cancelCtx := context.WithCancel(context.Background())

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case event := <-eventChan:
				handler(
					event.ERC20Token,
					event.Raw.BlockNumber,
				)
			}
		}
	}()

	sub := erctds.Pipe(eventChan)
	return subscription.NewEventSubscription(func() {
		sub.Unsubscribe()
		cancelCtx()
	})
}

func (erctds *MbERC20TokenDisabledSubscription) Pipe(
	sink chan *abi.MezoBridgeERC20TokenDisabled,
) subscription.EventSubscription {
	ctx, cancelCtx := context.WithCancel(context.Background())
	go func() {
		ticker := time.NewTicker(erctds.opts.Tick)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				lastBlock, err := erctds.contract.blockCounter.CurrentBlock()
				if err != nil {
					mbLogger.Errorf(
						"subscription failed to pull events: [%v]",
						err,
					)
				}
				fromBlock := lastBlock - erctds.opts.PastBlocks

				mbLogger.Infof(
					"subscription monitoring fetching past ERC20TokenDisabled events "+
						"starting from block [%v]",
					fromBlock,
				)
				events, err := erctds.contract.PastERC20TokenDisabledEvents(
					fromBlock,
					nil,
					erctds.ERC20TokenFilter,
				)
				if err != nil {
					mbLogger.Errorf(
						"subscription failed to pull events: [%v]",
						err,
					)
					continue
				}
				mbLogger.Infof(
					"subscription monitoring fetched [%v] past ERC20TokenDisabled events",
					len(events),
				)

				for _, event := range events {
					sink <- event
				}
			}
		}
	}()

	sub := erctds.contract.watchERC20TokenDisabled(
		sink,
		erctds.ERC20TokenFilter,
	)

	return subscription.NewEventSubscription(func() {
		sub.Unsubscribe()
		cancelCtx()
	})
}

func (mb *MezoBridge) watchERC20TokenDisabled(
	sink chan *abi.MezoBridgeERC20TokenDisabled,
	ERC20TokenFilter []common.Address,
) event.Subscription {
	subscribeFn := func(ctx context.Context) (event.Subscription, error) {
		return mb.contract.WatchERC20TokenDisabled(
			&bind.WatchOpts{Context: ctx},
			sink,
			ERC20TokenFilter,
		)
	}

	thresholdViolatedFn := func(elapsed time.Duration) {
		mbLogger.Warnf(
			"subscription to event ERC20TokenDisabled had to be "+
				"retried [%s] since the last attempt; please inspect "+
				"host chain connectivity",
			elapsed,
		)
	}

	subscriptionFailedFn := func(err error) {
		mbLogger.Errorf(
			"subscription to event ERC20TokenDisabled failed "+
				"with error: [%v]; resubscription attempt will be "+
				"performed",
			err,
		)
	}

	return chainutil.WithResubscription(
		chainutil.SubscriptionBackoffMax,
		subscribeFn,
		chainutil.SubscriptionAlertThreshold,
		thresholdViolatedFn,
		subscriptionFailedFn,
	)
}

func (mb *MezoBridge) PastERC20TokenDisabledEvents(
	startBlock uint64,
	endBlock *uint64,
	ERC20TokenFilter []common.Address,
) ([]*abi.MezoBridgeERC20TokenDisabled, error) {
	iterator, err := mb.contract.FilterERC20TokenDisabled(
		&bind.FilterOpts{
			Start: startBlock,
			End:   endBlock,
		},
		ERC20TokenFilter,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"error retrieving past ERC20TokenDisabled events: [%v]",
			err,
		)
	}

	events := make([]*abi.MezoBridgeERC20TokenDisabled, 0)

	for iterator.Next() {
		event := iterator.Event
		events = append(events, event)
	}

	return events, nil
}

func (mb *MezoBridge) ERC20TokenEnabledEvent(
	opts *ethereum.SubscribeOpts,
	ERC20TokenFilter []common.Address,
) *MbERC20TokenEnabledSubscription {
	if opts == nil {
		opts = new(ethereum.SubscribeOpts)
	}
	if opts.Tick == 0 {
		opts.Tick = chainutil.DefaultSubscribeOptsTick
	}
	if opts.PastBlocks == 0 {
		opts.PastBlocks = chainutil.DefaultSubscribeOptsPastBlocks
	}

	return &MbERC20TokenEnabledSubscription{
		mb,
		opts,
		ERC20TokenFilter,
	}
}

type MbERC20TokenEnabledSubscription struct {
	contract         *MezoBridge
	opts             *ethereum.SubscribeOpts
	ERC20TokenFilter []common.Address
}

type mezoBridgeERC20TokenEnabledFunc func(
	ERC20Token common.Address,
	MinERC20Amount *big.Int,
	blockNumber uint64,
)

func (erctes *MbERC20TokenEnabledSubscription) OnEvent(
	handler mezoBridgeERC20TokenEnabledFunc,
) subscription.EventSubscription {
	eventChan := make(chan *abi.MezoBridgeERC20TokenEnabled)
	ctx, cancelCtx := context.WithCancel(context.Background())

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case event := <-eventChan:
				handler(
					event.ERC20Token,
					event.MinERC20Amount,
					event.Raw.BlockNumber,
				)
			}
		}
	}()

	sub := erctes.Pipe(eventChan)
	return subscription.NewEventSubscription(func() {
		sub.Unsubscribe()
		cancelCtx()
	})
}

func (erctes *MbERC20TokenEnabledSubscription) Pipe(
	sink chan *abi.MezoBridgeERC20TokenEnabled,
) subscription.EventSubscription {
	ctx, cancelCtx := context.WithCancel(context.Background())
	go func() {
		ticker := time.NewTicker(erctes.opts.Tick)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				lastBlock, err := erctes.contract.blockCounter.CurrentBlock()
				if err != nil {
					mbLogger.Errorf(
						"subscription failed to pull events: [%v]",
						err,
					)
				}
				fromBlock := lastBlock - erctes.opts.PastBlocks

				mbLogger.Infof(
					"subscription monitoring fetching past ERC20TokenEnabled events "+
						"starting from block [%v]",
					fromBlock,
				)
				events, err := erctes.contract.PastERC20TokenEnabledEvents(
					fromBlock,
					nil,
					erctes.ERC20TokenFilter,
				)
				if err != nil {
					mbLogger.Errorf(
						"subscription failed to pull events: [%v]",
						err,
					)
					continue
				}
				mbLogger.Infof(
					"subscription monitoring fetched [%v] past ERC20TokenEnabled events",
					len(events),
				)

				for _, event := range events {
					sink <- event
				}
			}
		}
	}()

	sub := erctes.contract.watchERC20TokenEnabled(
		sink,
		erctes.ERC20TokenFilter,
	)

	return subscription.NewEventSubscription(func() {
		sub.Unsubscribe()
		cancelCtx()
	})
}

func (mb *MezoBridge) watchERC20TokenEnabled(
	sink chan *abi.MezoBridgeERC20TokenEnabled,
	ERC20TokenFilter []common.Address,
) event.Subscription {
	subscribeFn := func(ctx context.Context) (event.Subscription, error) {
		return mb.contract.WatchERC20TokenEnabled(
			&bind.WatchOpts{Context: ctx},
			sink,
			ERC20TokenFilter,
		)
	}

	thresholdViolatedFn := func(elapsed time.Duration) {
		mbLogger.Warnf(
			"subscription to event ERC20TokenEnabled had to be "+
				"retried [%s] since the last attempt; please inspect "+
				"host chain connectivity",
			elapsed,
		)
	}

	subscriptionFailedFn := func(err error) {
		mbLogger.Errorf(
			"subscription to event ERC20TokenEnabled failed "+
				"with error: [%v]; resubscription attempt will be "+
				"performed",
			err,
		)
	}

	return chainutil.WithResubscription(
		chainutil.SubscriptionBackoffMax,
		subscribeFn,
		chainutil.SubscriptionAlertThreshold,
		thresholdViolatedFn,
		subscriptionFailedFn,
	)
}

func (mb *MezoBridge) PastERC20TokenEnabledEvents(
	startBlock uint64,
	endBlock *uint64,
	ERC20TokenFilter []common.Address,
) ([]*abi.MezoBridgeERC20TokenEnabled, error) {
	iterator, err := mb.contract.FilterERC20TokenEnabled(
		&bind.FilterOpts{
			Start: startBlock,
			End:   endBlock,
		},
		ERC20TokenFilter,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"error retrieving past ERC20TokenEnabled events: [%v]",
			err,
		)
	}

	events := make([]*abi.MezoBridgeERC20TokenEnabled, 0)

	for iterator.Next() {
		event := iterator.Event
		events = append(events, event)
	}

	return events, nil
}

func (mb *MezoBridge) InitializedEvent(
	opts *ethereum.SubscribeOpts,
) *MbInitializedSubscription {
	if opts == nil {
		opts = new(ethereum.SubscribeOpts)
	}
	if opts.Tick == 0 {
		opts.Tick = chainutil.DefaultSubscribeOptsTick
	}
	if opts.PastBlocks == 0 {
		opts.PastBlocks = chainutil.DefaultSubscribeOptsPastBlocks
	}

	return &MbInitializedSubscription{
		mb,
		opts,
	}
}

type MbInitializedSubscription struct {
	contract *MezoBridge
	opts     *ethereum.SubscribeOpts
}

type mezoBridgeInitializedFunc func(
	Version uint64,
	blockNumber uint64,
)

func (is *MbInitializedSubscription) OnEvent(
	handler mezoBridgeInitializedFunc,
) subscription.EventSubscription {
	eventChan := make(chan *abi.MezoBridgeInitialized)
	ctx, cancelCtx := context.WithCancel(context.Background())

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case event := <-eventChan:
				handler(
					event.Version,
					event.Raw.BlockNumber,
				)
			}
		}
	}()

	sub := is.Pipe(eventChan)
	return subscription.NewEventSubscription(func() {
		sub.Unsubscribe()
		cancelCtx()
	})
}

func (is *MbInitializedSubscription) Pipe(
	sink chan *abi.MezoBridgeInitialized,
) subscription.EventSubscription {
	ctx, cancelCtx := context.WithCancel(context.Background())
	go func() {
		ticker := time.NewTicker(is.opts.Tick)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				lastBlock, err := is.contract.blockCounter.CurrentBlock()
				if err != nil {
					mbLogger.Errorf(
						"subscription failed to pull events: [%v]",
						err,
					)
				}
				fromBlock := lastBlock - is.opts.PastBlocks

				mbLogger.Infof(
					"subscription monitoring fetching past Initialized events "+
						"starting from block [%v]",
					fromBlock,
				)
				events, err := is.contract.PastInitializedEvents(
					fromBlock,
					nil,
				)
				if err != nil {
					mbLogger.Errorf(
						"subscription failed to pull events: [%v]",
						err,
					)
					continue
				}
				mbLogger.Infof(
					"subscription monitoring fetched [%v] past Initialized events",
					len(events),
				)

				for _, event := range events {
					sink <- event
				}
			}
		}
	}()

	sub := is.contract.watchInitialized(
		sink,
	)

	return subscription.NewEventSubscription(func() {
		sub.Unsubscribe()
		cancelCtx()
	})
}

func (mb *MezoBridge) watchInitialized(
	sink chan *abi.MezoBridgeInitialized,
) event.Subscription {
	subscribeFn := func(ctx context.Context) (event.Subscription, error) {
		return mb.contract.WatchInitialized(
			&bind.WatchOpts{Context: ctx},
			sink,
		)
	}

	thresholdViolatedFn := func(elapsed time.Duration) {
		mbLogger.Warnf(
			"subscription to event Initialized had to be "+
				"retried [%s] since the last attempt; please inspect "+
				"host chain connectivity",
			elapsed,
		)
	}

	subscriptionFailedFn := func(err error) {
		mbLogger.Errorf(
			"subscription to event Initialized failed "+
				"with error: [%v]; resubscription attempt will be "+
				"performed",
			err,
		)
	}

	return chainutil.WithResubscription(
		chainutil.SubscriptionBackoffMax,
		subscribeFn,
		chainutil.SubscriptionAlertThreshold,
		thresholdViolatedFn,
		subscriptionFailedFn,
	)
}

func (mb *MezoBridge) PastInitializedEvents(
	startBlock uint64,
	endBlock *uint64,
) ([]*abi.MezoBridgeInitialized, error) {
	iterator, err := mb.contract.FilterInitialized(
		&bind.FilterOpts{
			Start: startBlock,
			End:   endBlock,
		},
	)
	if err != nil {
		return nil, fmt.Errorf(
			"error retrieving past Initialized events: [%v]",
			err,
		)
	}

	events := make([]*abi.MezoBridgeInitialized, 0)

	for iterator.Next() {
		event := iterator.Event
		events = append(events, event)
	}

	return events, nil
}

func (mb *MezoBridge) MinERC20AmountUpdatedEvent(
	opts *ethereum.SubscribeOpts,
	ERC20TokenFilter []common.Address,
) *MbMinERC20AmountUpdatedSubscription {
	if opts == nil {
		opts = new(ethereum.SubscribeOpts)
	}
	if opts.Tick == 0 {
		opts.Tick = chainutil.DefaultSubscribeOptsTick
	}
	if opts.PastBlocks == 0 {
		opts.PastBlocks = chainutil.DefaultSubscribeOptsPastBlocks
	}

	return &MbMinERC20AmountUpdatedSubscription{
		mb,
		opts,
		ERC20TokenFilter,
	}
}

type MbMinERC20AmountUpdatedSubscription struct {
	contract         *MezoBridge
	opts             *ethereum.SubscribeOpts
	ERC20TokenFilter []common.Address
}

type mezoBridgeMinERC20AmountUpdatedFunc func(
	ERC20Token common.Address,
	NewMinERC20Amount *big.Int,
	blockNumber uint64,
)

func (mercaus *MbMinERC20AmountUpdatedSubscription) OnEvent(
	handler mezoBridgeMinERC20AmountUpdatedFunc,
) subscription.EventSubscription {
	eventChan := make(chan *abi.MezoBridgeMinERC20AmountUpdated)
	ctx, cancelCtx := context.WithCancel(context.Background())

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case event := <-eventChan:
				handler(
					event.ERC20Token,
					event.NewMinERC20Amount,
					event.Raw.BlockNumber,
				)
			}
		}
	}()

	sub := mercaus.Pipe(eventChan)
	return subscription.NewEventSubscription(func() {
		sub.Unsubscribe()
		cancelCtx()
	})
}

func (mercaus *MbMinERC20AmountUpdatedSubscription) Pipe(
	sink chan *abi.MezoBridgeMinERC20AmountUpdated,
) subscription.EventSubscription {
	ctx, cancelCtx := context.WithCancel(context.Background())
	go func() {
		ticker := time.NewTicker(mercaus.opts.Tick)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				lastBlock, err := mercaus.contract.blockCounter.CurrentBlock()
				if err != nil {
					mbLogger.Errorf(
						"subscription failed to pull events: [%v]",
						err,
					)
				}
				fromBlock := lastBlock - mercaus.opts.PastBlocks

				mbLogger.Infof(
					"subscription monitoring fetching past MinERC20AmountUpdated events "+
						"starting from block [%v]",
					fromBlock,
				)
				events, err := mercaus.contract.PastMinERC20AmountUpdatedEvents(
					fromBlock,
					nil,
					mercaus.ERC20TokenFilter,
				)
				if err != nil {
					mbLogger.Errorf(
						"subscription failed to pull events: [%v]",
						err,
					)
					continue
				}
				mbLogger.Infof(
					"subscription monitoring fetched [%v] past MinERC20AmountUpdated events",
					len(events),
				)

				for _, event := range events {
					sink <- event
				}
			}
		}
	}()

	sub := mercaus.contract.watchMinERC20AmountUpdated(
		sink,
		mercaus.ERC20TokenFilter,
	)

	return subscription.NewEventSubscription(func() {
		sub.Unsubscribe()
		cancelCtx()
	})
}

func (mb *MezoBridge) watchMinERC20AmountUpdated(
	sink chan *abi.MezoBridgeMinERC20AmountUpdated,
	ERC20TokenFilter []common.Address,
) event.Subscription {
	subscribeFn := func(ctx context.Context) (event.Subscription, error) {
		return mb.contract.WatchMinERC20AmountUpdated(
			&bind.WatchOpts{Context: ctx},
			sink,
			ERC20TokenFilter,
		)
	}

	thresholdViolatedFn := func(elapsed time.Duration) {
		mbLogger.Warnf(
			"subscription to event MinERC20AmountUpdated had to be "+
				"retried [%s] since the last attempt; please inspect "+
				"host chain connectivity",
			elapsed,
		)
	}

	subscriptionFailedFn := func(err error) {
		mbLogger.Errorf(
			"subscription to event MinERC20AmountUpdated failed "+
				"with error: [%v]; resubscription attempt will be "+
				"performed",
			err,
		)
	}

	return chainutil.WithResubscription(
		chainutil.SubscriptionBackoffMax,
		subscribeFn,
		chainutil.SubscriptionAlertThreshold,
		thresholdViolatedFn,
		subscriptionFailedFn,
	)
}

func (mb *MezoBridge) PastMinERC20AmountUpdatedEvents(
	startBlock uint64,
	endBlock *uint64,
	ERC20TokenFilter []common.Address,
) ([]*abi.MezoBridgeMinERC20AmountUpdated, error) {
	iterator, err := mb.contract.FilterMinERC20AmountUpdated(
		&bind.FilterOpts{
			Start: startBlock,
			End:   endBlock,
		},
		ERC20TokenFilter,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"error retrieving past MinERC20AmountUpdated events: [%v]",
			err,
		)
	}

	events := make([]*abi.MezoBridgeMinERC20AmountUpdated, 0)

	for iterator.Next() {
		event := iterator.Event
		events = append(events, event)
	}

	return events, nil
}

func (mb *MezoBridge) MinTBTCAmountUpdatedEvent(
	opts *ethereum.SubscribeOpts,
) *MbMinTBTCAmountUpdatedSubscription {
	if opts == nil {
		opts = new(ethereum.SubscribeOpts)
	}
	if opts.Tick == 0 {
		opts.Tick = chainutil.DefaultSubscribeOptsTick
	}
	if opts.PastBlocks == 0 {
		opts.PastBlocks = chainutil.DefaultSubscribeOptsPastBlocks
	}

	return &MbMinTBTCAmountUpdatedSubscription{
		mb,
		opts,
	}
}

type MbMinTBTCAmountUpdatedSubscription struct {
	contract *MezoBridge
	opts     *ethereum.SubscribeOpts
}

type mezoBridgeMinTBTCAmountUpdatedFunc func(
	MinTBTCAmount *big.Int,
	blockNumber uint64,
)

func (mtbtcaus *MbMinTBTCAmountUpdatedSubscription) OnEvent(
	handler mezoBridgeMinTBTCAmountUpdatedFunc,
) subscription.EventSubscription {
	eventChan := make(chan *abi.MezoBridgeMinTBTCAmountUpdated)
	ctx, cancelCtx := context.WithCancel(context.Background())

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case event := <-eventChan:
				handler(
					event.MinTBTCAmount,
					event.Raw.BlockNumber,
				)
			}
		}
	}()

	sub := mtbtcaus.Pipe(eventChan)
	return subscription.NewEventSubscription(func() {
		sub.Unsubscribe()
		cancelCtx()
	})
}

func (mtbtcaus *MbMinTBTCAmountUpdatedSubscription) Pipe(
	sink chan *abi.MezoBridgeMinTBTCAmountUpdated,
) subscription.EventSubscription {
	ctx, cancelCtx := context.WithCancel(context.Background())
	go func() {
		ticker := time.NewTicker(mtbtcaus.opts.Tick)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				lastBlock, err := mtbtcaus.contract.blockCounter.CurrentBlock()
				if err != nil {
					mbLogger.Errorf(
						"subscription failed to pull events: [%v]",
						err,
					)
				}
				fromBlock := lastBlock - mtbtcaus.opts.PastBlocks

				mbLogger.Infof(
					"subscription monitoring fetching past MinTBTCAmountUpdated events "+
						"starting from block [%v]",
					fromBlock,
				)
				events, err := mtbtcaus.contract.PastMinTBTCAmountUpdatedEvents(
					fromBlock,
					nil,
				)
				if err != nil {
					mbLogger.Errorf(
						"subscription failed to pull events: [%v]",
						err,
					)
					continue
				}
				mbLogger.Infof(
					"subscription monitoring fetched [%v] past MinTBTCAmountUpdated events",
					len(events),
				)

				for _, event := range events {
					sink <- event
				}
			}
		}
	}()

	sub := mtbtcaus.contract.watchMinTBTCAmountUpdated(
		sink,
	)

	return subscription.NewEventSubscription(func() {
		sub.Unsubscribe()
		cancelCtx()
	})
}

func (mb *MezoBridge) watchMinTBTCAmountUpdated(
	sink chan *abi.MezoBridgeMinTBTCAmountUpdated,
) event.Subscription {
	subscribeFn := func(ctx context.Context) (event.Subscription, error) {
		return mb.contract.WatchMinTBTCAmountUpdated(
			&bind.WatchOpts{Context: ctx},
			sink,
		)
	}

	thresholdViolatedFn := func(elapsed time.Duration) {
		mbLogger.Warnf(
			"subscription to event MinTBTCAmountUpdated had to be "+
				"retried [%s] since the last attempt; please inspect "+
				"host chain connectivity",
			elapsed,
		)
	}

	subscriptionFailedFn := func(err error) {
		mbLogger.Errorf(
			"subscription to event MinTBTCAmountUpdated failed "+
				"with error: [%v]; resubscription attempt will be "+
				"performed",
			err,
		)
	}

	return chainutil.WithResubscription(
		chainutil.SubscriptionBackoffMax,
		subscribeFn,
		chainutil.SubscriptionAlertThreshold,
		thresholdViolatedFn,
		subscriptionFailedFn,
	)
}

func (mb *MezoBridge) PastMinTBTCAmountUpdatedEvents(
	startBlock uint64,
	endBlock *uint64,
) ([]*abi.MezoBridgeMinTBTCAmountUpdated, error) {
	iterator, err := mb.contract.FilterMinTBTCAmountUpdated(
		&bind.FilterOpts{
			Start: startBlock,
			End:   endBlock,
		},
	)
	if err != nil {
		return nil, fmt.Errorf(
			"error retrieving past MinTBTCAmountUpdated events: [%v]",
			err,
		)
	}

	events := make([]*abi.MezoBridgeMinTBTCAmountUpdated, 0)

	for iterator.Next() {
		event := iterator.Event
		events = append(events, event)
	}

	return events, nil
}

func (mb *MezoBridge) OwnershipTransferStartedEvent(
	opts *ethereum.SubscribeOpts,
	previousOwnerFilter []common.Address,
	newOwnerFilter []common.Address,
) *MbOwnershipTransferStartedSubscription {
	if opts == nil {
		opts = new(ethereum.SubscribeOpts)
	}
	if opts.Tick == 0 {
		opts.Tick = chainutil.DefaultSubscribeOptsTick
	}
	if opts.PastBlocks == 0 {
		opts.PastBlocks = chainutil.DefaultSubscribeOptsPastBlocks
	}

	return &MbOwnershipTransferStartedSubscription{
		mb,
		opts,
		previousOwnerFilter,
		newOwnerFilter,
	}
}

type MbOwnershipTransferStartedSubscription struct {
	contract            *MezoBridge
	opts                *ethereum.SubscribeOpts
	previousOwnerFilter []common.Address
	newOwnerFilter      []common.Address
}

type mezoBridgeOwnershipTransferStartedFunc func(
	PreviousOwner common.Address,
	NewOwner common.Address,
	blockNumber uint64,
)

func (otss *MbOwnershipTransferStartedSubscription) OnEvent(
	handler mezoBridgeOwnershipTransferStartedFunc,
) subscription.EventSubscription {
	eventChan := make(chan *abi.MezoBridgeOwnershipTransferStarted)
	ctx, cancelCtx := context.WithCancel(context.Background())

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case event := <-eventChan:
				handler(
					event.PreviousOwner,
					event.NewOwner,
					event.Raw.BlockNumber,
				)
			}
		}
	}()

	sub := otss.Pipe(eventChan)
	return subscription.NewEventSubscription(func() {
		sub.Unsubscribe()
		cancelCtx()
	})
}

func (otss *MbOwnershipTransferStartedSubscription) Pipe(
	sink chan *abi.MezoBridgeOwnershipTransferStarted,
) subscription.EventSubscription {
	ctx, cancelCtx := context.WithCancel(context.Background())
	go func() {
		ticker := time.NewTicker(otss.opts.Tick)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				lastBlock, err := otss.contract.blockCounter.CurrentBlock()
				if err != nil {
					mbLogger.Errorf(
						"subscription failed to pull events: [%v]",
						err,
					)
				}
				fromBlock := lastBlock - otss.opts.PastBlocks

				mbLogger.Infof(
					"subscription monitoring fetching past OwnershipTransferStarted events "+
						"starting from block [%v]",
					fromBlock,
				)
				events, err := otss.contract.PastOwnershipTransferStartedEvents(
					fromBlock,
					nil,
					otss.previousOwnerFilter,
					otss.newOwnerFilter,
				)
				if err != nil {
					mbLogger.Errorf(
						"subscription failed to pull events: [%v]",
						err,
					)
					continue
				}
				mbLogger.Infof(
					"subscription monitoring fetched [%v] past OwnershipTransferStarted events",
					len(events),
				)

				for _, event := range events {
					sink <- event
				}
			}
		}
	}()

	sub := otss.contract.watchOwnershipTransferStarted(
		sink,
		otss.previousOwnerFilter,
		otss.newOwnerFilter,
	)

	return subscription.NewEventSubscription(func() {
		sub.Unsubscribe()
		cancelCtx()
	})
}

func (mb *MezoBridge) watchOwnershipTransferStarted(
	sink chan *abi.MezoBridgeOwnershipTransferStarted,
	previousOwnerFilter []common.Address,
	newOwnerFilter []common.Address,
) event.Subscription {
	subscribeFn := func(ctx context.Context) (event.Subscription, error) {
		return mb.contract.WatchOwnershipTransferStarted(
			&bind.WatchOpts{Context: ctx},
			sink,
			previousOwnerFilter,
			newOwnerFilter,
		)
	}

	thresholdViolatedFn := func(elapsed time.Duration) {
		mbLogger.Warnf(
			"subscription to event OwnershipTransferStarted had to be "+
				"retried [%s] since the last attempt; please inspect "+
				"host chain connectivity",
			elapsed,
		)
	}

	subscriptionFailedFn := func(err error) {
		mbLogger.Errorf(
			"subscription to event OwnershipTransferStarted failed "+
				"with error: [%v]; resubscription attempt will be "+
				"performed",
			err,
		)
	}

	return chainutil.WithResubscription(
		chainutil.SubscriptionBackoffMax,
		subscribeFn,
		chainutil.SubscriptionAlertThreshold,
		thresholdViolatedFn,
		subscriptionFailedFn,
	)
}

func (mb *MezoBridge) PastOwnershipTransferStartedEvents(
	startBlock uint64,
	endBlock *uint64,
	previousOwnerFilter []common.Address,
	newOwnerFilter []common.Address,
) ([]*abi.MezoBridgeOwnershipTransferStarted, error) {
	iterator, err := mb.contract.FilterOwnershipTransferStarted(
		&bind.FilterOpts{
			Start: startBlock,
			End:   endBlock,
		},
		previousOwnerFilter,
		newOwnerFilter,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"error retrieving past OwnershipTransferStarted events: [%v]",
			err,
		)
	}

	events := make([]*abi.MezoBridgeOwnershipTransferStarted, 0)

	for iterator.Next() {
		event := iterator.Event
		events = append(events, event)
	}

	return events, nil
}

func (mb *MezoBridge) OwnershipTransferredEvent(
	opts *ethereum.SubscribeOpts,
	previousOwnerFilter []common.Address,
	newOwnerFilter []common.Address,
) *MbOwnershipTransferredSubscription {
	if opts == nil {
		opts = new(ethereum.SubscribeOpts)
	}
	if opts.Tick == 0 {
		opts.Tick = chainutil.DefaultSubscribeOptsTick
	}
	if opts.PastBlocks == 0 {
		opts.PastBlocks = chainutil.DefaultSubscribeOptsPastBlocks
	}

	return &MbOwnershipTransferredSubscription{
		mb,
		opts,
		previousOwnerFilter,
		newOwnerFilter,
	}
}

type MbOwnershipTransferredSubscription struct {
	contract            *MezoBridge
	opts                *ethereum.SubscribeOpts
	previousOwnerFilter []common.Address
	newOwnerFilter      []common.Address
}

type mezoBridgeOwnershipTransferredFunc func(
	PreviousOwner common.Address,
	NewOwner common.Address,
	blockNumber uint64,
)

func (ots *MbOwnershipTransferredSubscription) OnEvent(
	handler mezoBridgeOwnershipTransferredFunc,
) subscription.EventSubscription {
	eventChan := make(chan *abi.MezoBridgeOwnershipTransferred)
	ctx, cancelCtx := context.WithCancel(context.Background())

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case event := <-eventChan:
				handler(
					event.PreviousOwner,
					event.NewOwner,
					event.Raw.BlockNumber,
				)
			}
		}
	}()

	sub := ots.Pipe(eventChan)
	return subscription.NewEventSubscription(func() {
		sub.Unsubscribe()
		cancelCtx()
	})
}

func (ots *MbOwnershipTransferredSubscription) Pipe(
	sink chan *abi.MezoBridgeOwnershipTransferred,
) subscription.EventSubscription {
	ctx, cancelCtx := context.WithCancel(context.Background())
	go func() {
		ticker := time.NewTicker(ots.opts.Tick)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				lastBlock, err := ots.contract.blockCounter.CurrentBlock()
				if err != nil {
					mbLogger.Errorf(
						"subscription failed to pull events: [%v]",
						err,
					)
				}
				fromBlock := lastBlock - ots.opts.PastBlocks

				mbLogger.Infof(
					"subscription monitoring fetching past OwnershipTransferred events "+
						"starting from block [%v]",
					fromBlock,
				)
				events, err := ots.contract.PastOwnershipTransferredEvents(
					fromBlock,
					nil,
					ots.previousOwnerFilter,
					ots.newOwnerFilter,
				)
				if err != nil {
					mbLogger.Errorf(
						"subscription failed to pull events: [%v]",
						err,
					)
					continue
				}
				mbLogger.Infof(
					"subscription monitoring fetched [%v] past OwnershipTransferred events",
					len(events),
				)

				for _, event := range events {
					sink <- event
				}
			}
		}
	}()

	sub := ots.contract.watchOwnershipTransferred(
		sink,
		ots.previousOwnerFilter,
		ots.newOwnerFilter,
	)

	return subscription.NewEventSubscription(func() {
		sub.Unsubscribe()
		cancelCtx()
	})
}

func (mb *MezoBridge) watchOwnershipTransferred(
	sink chan *abi.MezoBridgeOwnershipTransferred,
	previousOwnerFilter []common.Address,
	newOwnerFilter []common.Address,
) event.Subscription {
	subscribeFn := func(ctx context.Context) (event.Subscription, error) {
		return mb.contract.WatchOwnershipTransferred(
			&bind.WatchOpts{Context: ctx},
			sink,
			previousOwnerFilter,
			newOwnerFilter,
		)
	}

	thresholdViolatedFn := func(elapsed time.Duration) {
		mbLogger.Warnf(
			"subscription to event OwnershipTransferred had to be "+
				"retried [%s] since the last attempt; please inspect "+
				"host chain connectivity",
			elapsed,
		)
	}

	subscriptionFailedFn := func(err error) {
		mbLogger.Errorf(
			"subscription to event OwnershipTransferred failed "+
				"with error: [%v]; resubscription attempt will be "+
				"performed",
			err,
		)
	}

	return chainutil.WithResubscription(
		chainutil.SubscriptionBackoffMax,
		subscribeFn,
		chainutil.SubscriptionAlertThreshold,
		thresholdViolatedFn,
		subscriptionFailedFn,
	)
}

func (mb *MezoBridge) PastOwnershipTransferredEvents(
	startBlock uint64,
	endBlock *uint64,
	previousOwnerFilter []common.Address,
	newOwnerFilter []common.Address,
) ([]*abi.MezoBridgeOwnershipTransferred, error) {
	iterator, err := mb.contract.FilterOwnershipTransferred(
		&bind.FilterOpts{
			Start: startBlock,
			End:   endBlock,
		},
		previousOwnerFilter,
		newOwnerFilter,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"error retrieving past OwnershipTransferred events: [%v]",
			err,
		)
	}

	events := make([]*abi.MezoBridgeOwnershipTransferred, 0)

	for iterator.Next() {
		event := iterator.Event
		events = append(events, event)
	}

	return events, nil
}
