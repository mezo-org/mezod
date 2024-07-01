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
	"github.com/evmos/evmos/v12/ethereum/bindings/portal/gen/abi"

	"github.com/ipfs/go-log"

	"github.com/keep-network/keep-common/pkg/chain/ethereum"
	chainutil "github.com/keep-network/keep-common/pkg/chain/ethereum/ethutil"
	"github.com/keep-network/keep-common/pkg/subscription"
)

// Create a package-level logger for this contract. The logger exists at
// package level so that the logger is registered at startup and can be
// included or excluded from logging at startup by name.
var pLogger = log.Logger("keep-contract-Portal")

type Portal struct {
	contract          *abi.Portal
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

func NewPortal(
	contractAddress common.Address,
	chainId *big.Int,
	accountKey *keystore.Key,
	backend bind.ContractBackend,
	nonceManager *ethereum.NonceManager,
	miningWaiter *chainutil.MiningWaiter,
	blockCounter *ethereum.BlockCounter,
	transactionMutex *sync.Mutex,
) (*Portal, error) {
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

	contract, err := abi.NewPortal(
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

	contractABI, err := hostchainabi.JSON(strings.NewReader(abi.PortalABI))
	if err != nil {
		return nil, fmt.Errorf("failed to instantiate ABI: [%v]", err)
	}

	return &Portal{
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
func (p *Portal) AcceptOwnership(

	transactionOptions ...chainutil.TransactionOptions,
) (*types.Transaction, error) {
	pLogger.Debug(
		"submitting transaction acceptOwnership",
	)

	p.transactionMutex.Lock()
	defer p.transactionMutex.Unlock()

	// create a copy
	transactorOptions := new(bind.TransactOpts)
	*transactorOptions = *p.transactorOptions

	if len(transactionOptions) > 1 {
		return nil, fmt.Errorf(
			"could not process multiple transaction options sets",
		)
	} else if len(transactionOptions) > 0 {
		transactionOptions[0].Apply(transactorOptions)
	}

	nonce, err := p.nonceManager.CurrentNonce()
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve account nonce: %v", err)
	}

	transactorOptions.Nonce = new(big.Int).SetUint64(nonce)

	transaction, err := p.contract.AcceptOwnership(
		transactorOptions,
	)
	if err != nil {
		return transaction, p.errorResolver.ResolveError(
			err,
			p.transactorOptions.From,
			nil,
			"acceptOwnership",
		)
	}

	pLogger.Infof(
		"submitted transaction acceptOwnership with id: [%s] and nonce [%v]",
		transaction.Hash(),
		transaction.Nonce(),
	)

	go p.miningWaiter.ForceMining(
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

			transaction, err := p.contract.AcceptOwnership(
				newTransactorOptions,
			)
			if err != nil {
				return nil, p.errorResolver.ResolveError(
					err,
					p.transactorOptions.From,
					nil,
					"acceptOwnership",
				)
			}

			pLogger.Infof(
				"submitted transaction acceptOwnership with id: [%s] and nonce [%v]",
				transaction.Hash(),
				transaction.Nonce(),
			)

			return transaction, nil
		},
	)

	p.nonceManager.IncrementNonce()

	return transaction, err
}

// Non-mutating call, not a transaction submission.
func (p *Portal) CallAcceptOwnership(
	blockNumber *big.Int,
) error {
	var result interface{} = nil

	err := chainutil.CallAtBlock(
		p.transactorOptions.From,
		blockNumber, nil,
		p.contractABI,
		p.caller,
		p.errorResolver,
		p.contractAddress,
		"acceptOwnership",
		&result,
	)

	return err
}

func (p *Portal) AcceptOwnershipGasEstimate() (uint64, error) {
	var result uint64

	result, err := chainutil.EstimateGas(
		p.callerOptions.From,
		p.contractAddress,
		"acceptOwnership",
		p.contractABI,
		p.transactor,
	)

	return result, err
}

// Transaction submission.
func (p *Portal) AddSupportedToken(
	arg_supportedToken abi.PortalSupportedToken,

	transactionOptions ...chainutil.TransactionOptions,
) (*types.Transaction, error) {
	pLogger.Debug(
		"submitting transaction addSupportedToken",
		" params: ",
		fmt.Sprint(
			arg_supportedToken,
		),
	)

	p.transactionMutex.Lock()
	defer p.transactionMutex.Unlock()

	// create a copy
	transactorOptions := new(bind.TransactOpts)
	*transactorOptions = *p.transactorOptions

	if len(transactionOptions) > 1 {
		return nil, fmt.Errorf(
			"could not process multiple transaction options sets",
		)
	} else if len(transactionOptions) > 0 {
		transactionOptions[0].Apply(transactorOptions)
	}

	nonce, err := p.nonceManager.CurrentNonce()
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve account nonce: %v", err)
	}

	transactorOptions.Nonce = new(big.Int).SetUint64(nonce)

	transaction, err := p.contract.AddSupportedToken(
		transactorOptions,
		arg_supportedToken,
	)
	if err != nil {
		return transaction, p.errorResolver.ResolveError(
			err,
			p.transactorOptions.From,
			nil,
			"addSupportedToken",
			arg_supportedToken,
		)
	}

	pLogger.Infof(
		"submitted transaction addSupportedToken with id: [%s] and nonce [%v]",
		transaction.Hash(),
		transaction.Nonce(),
	)

	go p.miningWaiter.ForceMining(
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

			transaction, err := p.contract.AddSupportedToken(
				newTransactorOptions,
				arg_supportedToken,
			)
			if err != nil {
				return nil, p.errorResolver.ResolveError(
					err,
					p.transactorOptions.From,
					nil,
					"addSupportedToken",
					arg_supportedToken,
				)
			}

			pLogger.Infof(
				"submitted transaction addSupportedToken with id: [%s] and nonce [%v]",
				transaction.Hash(),
				transaction.Nonce(),
			)

			return transaction, nil
		},
	)

	p.nonceManager.IncrementNonce()

	return transaction, err
}

// Non-mutating call, not a transaction submission.
func (p *Portal) CallAddSupportedToken(
	arg_supportedToken abi.PortalSupportedToken,
	blockNumber *big.Int,
) error {
	var result interface{} = nil

	err := chainutil.CallAtBlock(
		p.transactorOptions.From,
		blockNumber, nil,
		p.contractABI,
		p.caller,
		p.errorResolver,
		p.contractAddress,
		"addSupportedToken",
		&result,
		arg_supportedToken,
	)

	return err
}

func (p *Portal) AddSupportedTokenGasEstimate(
	arg_supportedToken abi.PortalSupportedToken,
) (uint64, error) {
	var result uint64

	result, err := chainutil.EstimateGas(
		p.callerOptions.From,
		p.contractAddress,
		"addSupportedToken",
		p.contractABI,
		p.transactor,
		arg_supportedToken,
	)

	return result, err
}

// Transaction submission.
func (p *Portal) Deposit(
	arg_token common.Address,
	arg_amount *big.Int,
	arg_lockPeriod uint32,

	transactionOptions ...chainutil.TransactionOptions,
) (*types.Transaction, error) {
	pLogger.Debug(
		"submitting transaction deposit",
		" params: ",
		fmt.Sprint(
			arg_token,
			arg_amount,
			arg_lockPeriod,
		),
	)

	p.transactionMutex.Lock()
	defer p.transactionMutex.Unlock()

	// create a copy
	transactorOptions := new(bind.TransactOpts)
	*transactorOptions = *p.transactorOptions

	if len(transactionOptions) > 1 {
		return nil, fmt.Errorf(
			"could not process multiple transaction options sets",
		)
	} else if len(transactionOptions) > 0 {
		transactionOptions[0].Apply(transactorOptions)
	}

	nonce, err := p.nonceManager.CurrentNonce()
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve account nonce: %v", err)
	}

	transactorOptions.Nonce = new(big.Int).SetUint64(nonce)

	transaction, err := p.contract.Deposit(
		transactorOptions,
		arg_token,
		arg_amount,
		arg_lockPeriod,
	)
	if err != nil {
		return transaction, p.errorResolver.ResolveError(
			err,
			p.transactorOptions.From,
			nil,
			"deposit",
			arg_token,
			arg_amount,
			arg_lockPeriod,
		)
	}

	pLogger.Infof(
		"submitted transaction deposit with id: [%s] and nonce [%v]",
		transaction.Hash(),
		transaction.Nonce(),
	)

	go p.miningWaiter.ForceMining(
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

			transaction, err := p.contract.Deposit(
				newTransactorOptions,
				arg_token,
				arg_amount,
				arg_lockPeriod,
			)
			if err != nil {
				return nil, p.errorResolver.ResolveError(
					err,
					p.transactorOptions.From,
					nil,
					"deposit",
					arg_token,
					arg_amount,
					arg_lockPeriod,
				)
			}

			pLogger.Infof(
				"submitted transaction deposit with id: [%s] and nonce [%v]",
				transaction.Hash(),
				transaction.Nonce(),
			)

			return transaction, nil
		},
	)

	p.nonceManager.IncrementNonce()

	return transaction, err
}

// Non-mutating call, not a transaction submission.
func (p *Portal) CallDeposit(
	arg_token common.Address,
	arg_amount *big.Int,
	arg_lockPeriod uint32,
	blockNumber *big.Int,
) error {
	var result interface{} = nil

	err := chainutil.CallAtBlock(
		p.transactorOptions.From,
		blockNumber, nil,
		p.contractABI,
		p.caller,
		p.errorResolver,
		p.contractAddress,
		"deposit",
		&result,
		arg_token,
		arg_amount,
		arg_lockPeriod,
	)

	return err
}

func (p *Portal) DepositGasEstimate(
	arg_token common.Address,
	arg_amount *big.Int,
	arg_lockPeriod uint32,
) (uint64, error) {
	var result uint64

	result, err := chainutil.EstimateGas(
		p.callerOptions.From,
		p.contractAddress,
		"deposit",
		p.contractABI,
		p.transactor,
		arg_token,
		arg_amount,
		arg_lockPeriod,
	)

	return result, err
}

// Transaction submission.
func (p *Portal) DepositFor(
	arg_depositOwner common.Address,
	arg_token common.Address,
	arg_amount *big.Int,
	arg_lockPeriod uint32,

	transactionOptions ...chainutil.TransactionOptions,
) (*types.Transaction, error) {
	pLogger.Debug(
		"submitting transaction depositFor",
		" params: ",
		fmt.Sprint(
			arg_depositOwner,
			arg_token,
			arg_amount,
			arg_lockPeriod,
		),
	)

	p.transactionMutex.Lock()
	defer p.transactionMutex.Unlock()

	// create a copy
	transactorOptions := new(bind.TransactOpts)
	*transactorOptions = *p.transactorOptions

	if len(transactionOptions) > 1 {
		return nil, fmt.Errorf(
			"could not process multiple transaction options sets",
		)
	} else if len(transactionOptions) > 0 {
		transactionOptions[0].Apply(transactorOptions)
	}

	nonce, err := p.nonceManager.CurrentNonce()
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve account nonce: %v", err)
	}

	transactorOptions.Nonce = new(big.Int).SetUint64(nonce)

	transaction, err := p.contract.DepositFor(
		transactorOptions,
		arg_depositOwner,
		arg_token,
		arg_amount,
		arg_lockPeriod,
	)
	if err != nil {
		return transaction, p.errorResolver.ResolveError(
			err,
			p.transactorOptions.From,
			nil,
			"depositFor",
			arg_depositOwner,
			arg_token,
			arg_amount,
			arg_lockPeriod,
		)
	}

	pLogger.Infof(
		"submitted transaction depositFor with id: [%s] and nonce [%v]",
		transaction.Hash(),
		transaction.Nonce(),
	)

	go p.miningWaiter.ForceMining(
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

			transaction, err := p.contract.DepositFor(
				newTransactorOptions,
				arg_depositOwner,
				arg_token,
				arg_amount,
				arg_lockPeriod,
			)
			if err != nil {
				return nil, p.errorResolver.ResolveError(
					err,
					p.transactorOptions.From,
					nil,
					"depositFor",
					arg_depositOwner,
					arg_token,
					arg_amount,
					arg_lockPeriod,
				)
			}

			pLogger.Infof(
				"submitted transaction depositFor with id: [%s] and nonce [%v]",
				transaction.Hash(),
				transaction.Nonce(),
			)

			return transaction, nil
		},
	)

	p.nonceManager.IncrementNonce()

	return transaction, err
}

// Non-mutating call, not a transaction submission.
func (p *Portal) CallDepositFor(
	arg_depositOwner common.Address,
	arg_token common.Address,
	arg_amount *big.Int,
	arg_lockPeriod uint32,
	blockNumber *big.Int,
) error {
	var result interface{} = nil

	err := chainutil.CallAtBlock(
		p.transactorOptions.From,
		blockNumber, nil,
		p.contractABI,
		p.caller,
		p.errorResolver,
		p.contractAddress,
		"depositFor",
		&result,
		arg_depositOwner,
		arg_token,
		arg_amount,
		arg_lockPeriod,
	)

	return err
}

func (p *Portal) DepositForGasEstimate(
	arg_depositOwner common.Address,
	arg_token common.Address,
	arg_amount *big.Int,
	arg_lockPeriod uint32,
) (uint64, error) {
	var result uint64

	result, err := chainutil.EstimateGas(
		p.callerOptions.From,
		p.contractAddress,
		"depositFor",
		p.contractABI,
		p.transactor,
		arg_depositOwner,
		arg_token,
		arg_amount,
		arg_lockPeriod,
	)

	return result, err
}

// Transaction submission.
func (p *Portal) Initialize(
	arg_supportedTokens []abi.PortalSupportedToken,

	transactionOptions ...chainutil.TransactionOptions,
) (*types.Transaction, error) {
	pLogger.Debug(
		"submitting transaction initialize",
		" params: ",
		fmt.Sprint(
			arg_supportedTokens,
		),
	)

	p.transactionMutex.Lock()
	defer p.transactionMutex.Unlock()

	// create a copy
	transactorOptions := new(bind.TransactOpts)
	*transactorOptions = *p.transactorOptions

	if len(transactionOptions) > 1 {
		return nil, fmt.Errorf(
			"could not process multiple transaction options sets",
		)
	} else if len(transactionOptions) > 0 {
		transactionOptions[0].Apply(transactorOptions)
	}

	nonce, err := p.nonceManager.CurrentNonce()
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve account nonce: %v", err)
	}

	transactorOptions.Nonce = new(big.Int).SetUint64(nonce)

	transaction, err := p.contract.Initialize(
		transactorOptions,
		arg_supportedTokens,
	)
	if err != nil {
		return transaction, p.errorResolver.ResolveError(
			err,
			p.transactorOptions.From,
			nil,
			"initialize",
			arg_supportedTokens,
		)
	}

	pLogger.Infof(
		"submitted transaction initialize with id: [%s] and nonce [%v]",
		transaction.Hash(),
		transaction.Nonce(),
	)

	go p.miningWaiter.ForceMining(
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

			transaction, err := p.contract.Initialize(
				newTransactorOptions,
				arg_supportedTokens,
			)
			if err != nil {
				return nil, p.errorResolver.ResolveError(
					err,
					p.transactorOptions.From,
					nil,
					"initialize",
					arg_supportedTokens,
				)
			}

			pLogger.Infof(
				"submitted transaction initialize with id: [%s] and nonce [%v]",
				transaction.Hash(),
				transaction.Nonce(),
			)

			return transaction, nil
		},
	)

	p.nonceManager.IncrementNonce()

	return transaction, err
}

// Non-mutating call, not a transaction submission.
func (p *Portal) CallInitialize(
	arg_supportedTokens []abi.PortalSupportedToken,
	blockNumber *big.Int,
) error {
	var result interface{} = nil

	err := chainutil.CallAtBlock(
		p.transactorOptions.From,
		blockNumber, nil,
		p.contractABI,
		p.caller,
		p.errorResolver,
		p.contractAddress,
		"initialize",
		&result,
		arg_supportedTokens,
	)

	return err
}

func (p *Portal) InitializeGasEstimate(
	arg_supportedTokens []abi.PortalSupportedToken,
) (uint64, error) {
	var result uint64

	result, err := chainutil.EstimateGas(
		p.callerOptions.From,
		p.contractAddress,
		"initialize",
		p.contractABI,
		p.transactor,
		arg_supportedTokens,
	)

	return result, err
}

// Transaction submission.
func (p *Portal) Lock(
	arg_token common.Address,
	arg_depositId *big.Int,
	arg_lockPeriod uint32,

	transactionOptions ...chainutil.TransactionOptions,
) (*types.Transaction, error) {
	pLogger.Debug(
		"submitting transaction lock",
		" params: ",
		fmt.Sprint(
			arg_token,
			arg_depositId,
			arg_lockPeriod,
		),
	)

	p.transactionMutex.Lock()
	defer p.transactionMutex.Unlock()

	// create a copy
	transactorOptions := new(bind.TransactOpts)
	*transactorOptions = *p.transactorOptions

	if len(transactionOptions) > 1 {
		return nil, fmt.Errorf(
			"could not process multiple transaction options sets",
		)
	} else if len(transactionOptions) > 0 {
		transactionOptions[0].Apply(transactorOptions)
	}

	nonce, err := p.nonceManager.CurrentNonce()
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve account nonce: %v", err)
	}

	transactorOptions.Nonce = new(big.Int).SetUint64(nonce)

	transaction, err := p.contract.Lock(
		transactorOptions,
		arg_token,
		arg_depositId,
		arg_lockPeriod,
	)
	if err != nil {
		return transaction, p.errorResolver.ResolveError(
			err,
			p.transactorOptions.From,
			nil,
			"lock",
			arg_token,
			arg_depositId,
			arg_lockPeriod,
		)
	}

	pLogger.Infof(
		"submitted transaction lock with id: [%s] and nonce [%v]",
		transaction.Hash(),
		transaction.Nonce(),
	)

	go p.miningWaiter.ForceMining(
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

			transaction, err := p.contract.Lock(
				newTransactorOptions,
				arg_token,
				arg_depositId,
				arg_lockPeriod,
			)
			if err != nil {
				return nil, p.errorResolver.ResolveError(
					err,
					p.transactorOptions.From,
					nil,
					"lock",
					arg_token,
					arg_depositId,
					arg_lockPeriod,
				)
			}

			pLogger.Infof(
				"submitted transaction lock with id: [%s] and nonce [%v]",
				transaction.Hash(),
				transaction.Nonce(),
			)

			return transaction, nil
		},
	)

	p.nonceManager.IncrementNonce()

	return transaction, err
}

// Non-mutating call, not a transaction submission.
func (p *Portal) CallLock(
	arg_token common.Address,
	arg_depositId *big.Int,
	arg_lockPeriod uint32,
	blockNumber *big.Int,
) error {
	var result interface{} = nil

	err := chainutil.CallAtBlock(
		p.transactorOptions.From,
		blockNumber, nil,
		p.contractABI,
		p.caller,
		p.errorResolver,
		p.contractAddress,
		"lock",
		&result,
		arg_token,
		arg_depositId,
		arg_lockPeriod,
	)

	return err
}

func (p *Portal) LockGasEstimate(
	arg_token common.Address,
	arg_depositId *big.Int,
	arg_lockPeriod uint32,
) (uint64, error) {
	var result uint64

	result, err := chainutil.EstimateGas(
		p.callerOptions.From,
		p.contractAddress,
		"lock",
		p.contractABI,
		p.transactor,
		arg_token,
		arg_depositId,
		arg_lockPeriod,
	)

	return result, err
}

// Transaction submission.
func (p *Portal) ReceiveApproval(
	arg_from common.Address,
	arg_amount *big.Int,
	arg_token common.Address,
	arg_data []byte,

	transactionOptions ...chainutil.TransactionOptions,
) (*types.Transaction, error) {
	pLogger.Debug(
		"submitting transaction receiveApproval",
		" params: ",
		fmt.Sprint(
			arg_from,
			arg_amount,
			arg_token,
			arg_data,
		),
	)

	p.transactionMutex.Lock()
	defer p.transactionMutex.Unlock()

	// create a copy
	transactorOptions := new(bind.TransactOpts)
	*transactorOptions = *p.transactorOptions

	if len(transactionOptions) > 1 {
		return nil, fmt.Errorf(
			"could not process multiple transaction options sets",
		)
	} else if len(transactionOptions) > 0 {
		transactionOptions[0].Apply(transactorOptions)
	}

	nonce, err := p.nonceManager.CurrentNonce()
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve account nonce: %v", err)
	}

	transactorOptions.Nonce = new(big.Int).SetUint64(nonce)

	transaction, err := p.contract.ReceiveApproval(
		transactorOptions,
		arg_from,
		arg_amount,
		arg_token,
		arg_data,
	)
	if err != nil {
		return transaction, p.errorResolver.ResolveError(
			err,
			p.transactorOptions.From,
			nil,
			"receiveApproval",
			arg_from,
			arg_amount,
			arg_token,
			arg_data,
		)
	}

	pLogger.Infof(
		"submitted transaction receiveApproval with id: [%s] and nonce [%v]",
		transaction.Hash(),
		transaction.Nonce(),
	)

	go p.miningWaiter.ForceMining(
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

			transaction, err := p.contract.ReceiveApproval(
				newTransactorOptions,
				arg_from,
				arg_amount,
				arg_token,
				arg_data,
			)
			if err != nil {
				return nil, p.errorResolver.ResolveError(
					err,
					p.transactorOptions.From,
					nil,
					"receiveApproval",
					arg_from,
					arg_amount,
					arg_token,
					arg_data,
				)
			}

			pLogger.Infof(
				"submitted transaction receiveApproval with id: [%s] and nonce [%v]",
				transaction.Hash(),
				transaction.Nonce(),
			)

			return transaction, nil
		},
	)

	p.nonceManager.IncrementNonce()

	return transaction, err
}

// Non-mutating call, not a transaction submission.
func (p *Portal) CallReceiveApproval(
	arg_from common.Address,
	arg_amount *big.Int,
	arg_token common.Address,
	arg_data []byte,
	blockNumber *big.Int,
) error {
	var result interface{} = nil

	err := chainutil.CallAtBlock(
		p.transactorOptions.From,
		blockNumber, nil,
		p.contractABI,
		p.caller,
		p.errorResolver,
		p.contractAddress,
		"receiveApproval",
		&result,
		arg_from,
		arg_amount,
		arg_token,
		arg_data,
	)

	return err
}

func (p *Portal) ReceiveApprovalGasEstimate(
	arg_from common.Address,
	arg_amount *big.Int,
	arg_token common.Address,
	arg_data []byte,
) (uint64, error) {
	var result uint64

	result, err := chainutil.EstimateGas(
		p.callerOptions.From,
		p.contractAddress,
		"receiveApproval",
		p.contractABI,
		p.transactor,
		arg_from,
		arg_amount,
		arg_token,
		arg_data,
	)

	return result, err
}

// Transaction submission.
func (p *Portal) RenounceOwnership(

	transactionOptions ...chainutil.TransactionOptions,
) (*types.Transaction, error) {
	pLogger.Debug(
		"submitting transaction renounceOwnership",
	)

	p.transactionMutex.Lock()
	defer p.transactionMutex.Unlock()

	// create a copy
	transactorOptions := new(bind.TransactOpts)
	*transactorOptions = *p.transactorOptions

	if len(transactionOptions) > 1 {
		return nil, fmt.Errorf(
			"could not process multiple transaction options sets",
		)
	} else if len(transactionOptions) > 0 {
		transactionOptions[0].Apply(transactorOptions)
	}

	nonce, err := p.nonceManager.CurrentNonce()
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve account nonce: %v", err)
	}

	transactorOptions.Nonce = new(big.Int).SetUint64(nonce)

	transaction, err := p.contract.RenounceOwnership(
		transactorOptions,
	)
	if err != nil {
		return transaction, p.errorResolver.ResolveError(
			err,
			p.transactorOptions.From,
			nil,
			"renounceOwnership",
		)
	}

	pLogger.Infof(
		"submitted transaction renounceOwnership with id: [%s] and nonce [%v]",
		transaction.Hash(),
		transaction.Nonce(),
	)

	go p.miningWaiter.ForceMining(
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

			transaction, err := p.contract.RenounceOwnership(
				newTransactorOptions,
			)
			if err != nil {
				return nil, p.errorResolver.ResolveError(
					err,
					p.transactorOptions.From,
					nil,
					"renounceOwnership",
				)
			}

			pLogger.Infof(
				"submitted transaction renounceOwnership with id: [%s] and nonce [%v]",
				transaction.Hash(),
				transaction.Nonce(),
			)

			return transaction, nil
		},
	)

	p.nonceManager.IncrementNonce()

	return transaction, err
}

// Non-mutating call, not a transaction submission.
func (p *Portal) CallRenounceOwnership(
	blockNumber *big.Int,
) error {
	var result interface{} = nil

	err := chainutil.CallAtBlock(
		p.transactorOptions.From,
		blockNumber, nil,
		p.contractABI,
		p.caller,
		p.errorResolver,
		p.contractAddress,
		"renounceOwnership",
		&result,
	)

	return err
}

func (p *Portal) RenounceOwnershipGasEstimate() (uint64, error) {
	var result uint64

	result, err := chainutil.EstimateGas(
		p.callerOptions.From,
		p.contractAddress,
		"renounceOwnership",
		p.contractABI,
		p.transactor,
	)

	return result, err
}

// Transaction submission.
func (p *Portal) SetMaxLockPeriod(
	arg__maxLockPeriod uint32,

	transactionOptions ...chainutil.TransactionOptions,
) (*types.Transaction, error) {
	pLogger.Debug(
		"submitting transaction setMaxLockPeriod",
		" params: ",
		fmt.Sprint(
			arg__maxLockPeriod,
		),
	)

	p.transactionMutex.Lock()
	defer p.transactionMutex.Unlock()

	// create a copy
	transactorOptions := new(bind.TransactOpts)
	*transactorOptions = *p.transactorOptions

	if len(transactionOptions) > 1 {
		return nil, fmt.Errorf(
			"could not process multiple transaction options sets",
		)
	} else if len(transactionOptions) > 0 {
		transactionOptions[0].Apply(transactorOptions)
	}

	nonce, err := p.nonceManager.CurrentNonce()
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve account nonce: %v", err)
	}

	transactorOptions.Nonce = new(big.Int).SetUint64(nonce)

	transaction, err := p.contract.SetMaxLockPeriod(
		transactorOptions,
		arg__maxLockPeriod,
	)
	if err != nil {
		return transaction, p.errorResolver.ResolveError(
			err,
			p.transactorOptions.From,
			nil,
			"setMaxLockPeriod",
			arg__maxLockPeriod,
		)
	}

	pLogger.Infof(
		"submitted transaction setMaxLockPeriod with id: [%s] and nonce [%v]",
		transaction.Hash(),
		transaction.Nonce(),
	)

	go p.miningWaiter.ForceMining(
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

			transaction, err := p.contract.SetMaxLockPeriod(
				newTransactorOptions,
				arg__maxLockPeriod,
			)
			if err != nil {
				return nil, p.errorResolver.ResolveError(
					err,
					p.transactorOptions.From,
					nil,
					"setMaxLockPeriod",
					arg__maxLockPeriod,
				)
			}

			pLogger.Infof(
				"submitted transaction setMaxLockPeriod with id: [%s] and nonce [%v]",
				transaction.Hash(),
				transaction.Nonce(),
			)

			return transaction, nil
		},
	)

	p.nonceManager.IncrementNonce()

	return transaction, err
}

// Non-mutating call, not a transaction submission.
func (p *Portal) CallSetMaxLockPeriod(
	arg__maxLockPeriod uint32,
	blockNumber *big.Int,
) error {
	var result interface{} = nil

	err := chainutil.CallAtBlock(
		p.transactorOptions.From,
		blockNumber, nil,
		p.contractABI,
		p.caller,
		p.errorResolver,
		p.contractAddress,
		"setMaxLockPeriod",
		&result,
		arg__maxLockPeriod,
	)

	return err
}

func (p *Portal) SetMaxLockPeriodGasEstimate(
	arg__maxLockPeriod uint32,
) (uint64, error) {
	var result uint64

	result, err := chainutil.EstimateGas(
		p.callerOptions.From,
		p.contractAddress,
		"setMaxLockPeriod",
		p.contractABI,
		p.transactor,
		arg__maxLockPeriod,
	)

	return result, err
}

// Transaction submission.
func (p *Portal) SetMinLockPeriod(
	arg__minLockPeriod uint32,

	transactionOptions ...chainutil.TransactionOptions,
) (*types.Transaction, error) {
	pLogger.Debug(
		"submitting transaction setMinLockPeriod",
		" params: ",
		fmt.Sprint(
			arg__minLockPeriod,
		),
	)

	p.transactionMutex.Lock()
	defer p.transactionMutex.Unlock()

	// create a copy
	transactorOptions := new(bind.TransactOpts)
	*transactorOptions = *p.transactorOptions

	if len(transactionOptions) > 1 {
		return nil, fmt.Errorf(
			"could not process multiple transaction options sets",
		)
	} else if len(transactionOptions) > 0 {
		transactionOptions[0].Apply(transactorOptions)
	}

	nonce, err := p.nonceManager.CurrentNonce()
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve account nonce: %v", err)
	}

	transactorOptions.Nonce = new(big.Int).SetUint64(nonce)

	transaction, err := p.contract.SetMinLockPeriod(
		transactorOptions,
		arg__minLockPeriod,
	)
	if err != nil {
		return transaction, p.errorResolver.ResolveError(
			err,
			p.transactorOptions.From,
			nil,
			"setMinLockPeriod",
			arg__minLockPeriod,
		)
	}

	pLogger.Infof(
		"submitted transaction setMinLockPeriod with id: [%s] and nonce [%v]",
		transaction.Hash(),
		transaction.Nonce(),
	)

	go p.miningWaiter.ForceMining(
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

			transaction, err := p.contract.SetMinLockPeriod(
				newTransactorOptions,
				arg__minLockPeriod,
			)
			if err != nil {
				return nil, p.errorResolver.ResolveError(
					err,
					p.transactorOptions.From,
					nil,
					"setMinLockPeriod",
					arg__minLockPeriod,
				)
			}

			pLogger.Infof(
				"submitted transaction setMinLockPeriod with id: [%s] and nonce [%v]",
				transaction.Hash(),
				transaction.Nonce(),
			)

			return transaction, nil
		},
	)

	p.nonceManager.IncrementNonce()

	return transaction, err
}

// Non-mutating call, not a transaction submission.
func (p *Portal) CallSetMinLockPeriod(
	arg__minLockPeriod uint32,
	blockNumber *big.Int,
) error {
	var result interface{} = nil

	err := chainutil.CallAtBlock(
		p.transactorOptions.From,
		blockNumber, nil,
		p.contractABI,
		p.caller,
		p.errorResolver,
		p.contractAddress,
		"setMinLockPeriod",
		&result,
		arg__minLockPeriod,
	)

	return err
}

func (p *Portal) SetMinLockPeriodGasEstimate(
	arg__minLockPeriod uint32,
) (uint64, error) {
	var result uint64

	result, err := chainutil.EstimateGas(
		p.callerOptions.From,
		p.contractAddress,
		"setMinLockPeriod",
		p.contractABI,
		p.transactor,
		arg__minLockPeriod,
	)

	return result, err
}

// Transaction submission.
func (p *Portal) TransferOwnership(
	arg_newOwner common.Address,

	transactionOptions ...chainutil.TransactionOptions,
) (*types.Transaction, error) {
	pLogger.Debug(
		"submitting transaction transferOwnership",
		" params: ",
		fmt.Sprint(
			arg_newOwner,
		),
	)

	p.transactionMutex.Lock()
	defer p.transactionMutex.Unlock()

	// create a copy
	transactorOptions := new(bind.TransactOpts)
	*transactorOptions = *p.transactorOptions

	if len(transactionOptions) > 1 {
		return nil, fmt.Errorf(
			"could not process multiple transaction options sets",
		)
	} else if len(transactionOptions) > 0 {
		transactionOptions[0].Apply(transactorOptions)
	}

	nonce, err := p.nonceManager.CurrentNonce()
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve account nonce: %v", err)
	}

	transactorOptions.Nonce = new(big.Int).SetUint64(nonce)

	transaction, err := p.contract.TransferOwnership(
		transactorOptions,
		arg_newOwner,
	)
	if err != nil {
		return transaction, p.errorResolver.ResolveError(
			err,
			p.transactorOptions.From,
			nil,
			"transferOwnership",
			arg_newOwner,
		)
	}

	pLogger.Infof(
		"submitted transaction transferOwnership with id: [%s] and nonce [%v]",
		transaction.Hash(),
		transaction.Nonce(),
	)

	go p.miningWaiter.ForceMining(
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

			transaction, err := p.contract.TransferOwnership(
				newTransactorOptions,
				arg_newOwner,
			)
			if err != nil {
				return nil, p.errorResolver.ResolveError(
					err,
					p.transactorOptions.From,
					nil,
					"transferOwnership",
					arg_newOwner,
				)
			}

			pLogger.Infof(
				"submitted transaction transferOwnership with id: [%s] and nonce [%v]",
				transaction.Hash(),
				transaction.Nonce(),
			)

			return transaction, nil
		},
	)

	p.nonceManager.IncrementNonce()

	return transaction, err
}

// Non-mutating call, not a transaction submission.
func (p *Portal) CallTransferOwnership(
	arg_newOwner common.Address,
	blockNumber *big.Int,
) error {
	var result interface{} = nil

	err := chainutil.CallAtBlock(
		p.transactorOptions.From,
		blockNumber, nil,
		p.contractABI,
		p.caller,
		p.errorResolver,
		p.contractAddress,
		"transferOwnership",
		&result,
		arg_newOwner,
	)

	return err
}

func (p *Portal) TransferOwnershipGasEstimate(
	arg_newOwner common.Address,
) (uint64, error) {
	var result uint64

	result, err := chainutil.EstimateGas(
		p.callerOptions.From,
		p.contractAddress,
		"transferOwnership",
		p.contractABI,
		p.transactor,
		arg_newOwner,
	)

	return result, err
}

// Transaction submission.
func (p *Portal) Withdraw(
	arg_token common.Address,
	arg_depositId *big.Int,
	arg_amount *big.Int,

	transactionOptions ...chainutil.TransactionOptions,
) (*types.Transaction, error) {
	pLogger.Debug(
		"submitting transaction withdraw",
		" params: ",
		fmt.Sprint(
			arg_token,
			arg_depositId,
			arg_amount,
		),
	)

	p.transactionMutex.Lock()
	defer p.transactionMutex.Unlock()

	// create a copy
	transactorOptions := new(bind.TransactOpts)
	*transactorOptions = *p.transactorOptions

	if len(transactionOptions) > 1 {
		return nil, fmt.Errorf(
			"could not process multiple transaction options sets",
		)
	} else if len(transactionOptions) > 0 {
		transactionOptions[0].Apply(transactorOptions)
	}

	nonce, err := p.nonceManager.CurrentNonce()
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve account nonce: %v", err)
	}

	transactorOptions.Nonce = new(big.Int).SetUint64(nonce)

	transaction, err := p.contract.Withdraw(
		transactorOptions,
		arg_token,
		arg_depositId,
		arg_amount,
	)
	if err != nil {
		return transaction, p.errorResolver.ResolveError(
			err,
			p.transactorOptions.From,
			nil,
			"withdraw",
			arg_token,
			arg_depositId,
			arg_amount,
		)
	}

	pLogger.Infof(
		"submitted transaction withdraw with id: [%s] and nonce [%v]",
		transaction.Hash(),
		transaction.Nonce(),
	)

	go p.miningWaiter.ForceMining(
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

			transaction, err := p.contract.Withdraw(
				newTransactorOptions,
				arg_token,
				arg_depositId,
				arg_amount,
			)
			if err != nil {
				return nil, p.errorResolver.ResolveError(
					err,
					p.transactorOptions.From,
					nil,
					"withdraw",
					arg_token,
					arg_depositId,
					arg_amount,
				)
			}

			pLogger.Infof(
				"submitted transaction withdraw with id: [%s] and nonce [%v]",
				transaction.Hash(),
				transaction.Nonce(),
			)

			return transaction, nil
		},
	)

	p.nonceManager.IncrementNonce()

	return transaction, err
}

// Non-mutating call, not a transaction submission.
func (p *Portal) CallWithdraw(
	arg_token common.Address,
	arg_depositId *big.Int,
	arg_amount *big.Int,
	blockNumber *big.Int,
) error {
	var result interface{} = nil

	err := chainutil.CallAtBlock(
		p.transactorOptions.From,
		blockNumber, nil,
		p.contractABI,
		p.caller,
		p.errorResolver,
		p.contractAddress,
		"withdraw",
		&result,
		arg_token,
		arg_depositId,
		arg_amount,
	)

	return err
}

func (p *Portal) WithdrawGasEstimate(
	arg_token common.Address,
	arg_depositId *big.Int,
	arg_amount *big.Int,
) (uint64, error) {
	var result uint64

	result, err := chainutil.EstimateGas(
		p.callerOptions.From,
		p.contractAddress,
		"withdraw",
		p.contractABI,
		p.transactor,
		arg_token,
		arg_depositId,
		arg_amount,
	)

	return result, err
}

// ----- Const Methods ------

func (p *Portal) DepositCount() (*big.Int, error) {
	result, err := p.contract.DepositCount(
		p.callerOptions,
	)

	if err != nil {
		return result, p.errorResolver.ResolveError(
			err,
			p.callerOptions.From,
			nil,
			"depositCount",
		)
	}

	return result, err
}

func (p *Portal) DepositCountAtBlock(
	blockNumber *big.Int,
) (*big.Int, error) {
	var result *big.Int

	err := chainutil.CallAtBlock(
		p.callerOptions.From,
		blockNumber,
		nil,
		p.contractABI,
		p.caller,
		p.errorResolver,
		p.contractAddress,
		"depositCount",
		&result,
	)

	return result, err
}

type deposits struct {
	Balance  *big.Int
	UnlockAt uint32
}

func (p *Portal) Deposits(
	arg0 common.Address,
	arg1 common.Address,
	arg2 *big.Int,
) (deposits, error) {
	result, err := p.contract.Deposits(
		p.callerOptions,
		arg0,
		arg1,
		arg2,
	)

	if err != nil {
		return result, p.errorResolver.ResolveError(
			err,
			p.callerOptions.From,
			nil,
			"deposits",
			arg0,
			arg1,
			arg2,
		)
	}

	return result, err
}

func (p *Portal) DepositsAtBlock(
	arg0 common.Address,
	arg1 common.Address,
	arg2 *big.Int,
	blockNumber *big.Int,
) (deposits, error) {
	var result deposits

	err := chainutil.CallAtBlock(
		p.callerOptions.From,
		blockNumber,
		nil,
		p.contractABI,
		p.caller,
		p.errorResolver,
		p.contractAddress,
		"deposits",
		&result,
		arg0,
		arg1,
		arg2,
	)

	return result, err
}

func (p *Portal) GetDeposit(
	arg_depositor common.Address,
	arg_token common.Address,
	arg_depositId *big.Int,
) (abi.PortalDepositInfo, error) {
	result, err := p.contract.GetDeposit(
		p.callerOptions,
		arg_depositor,
		arg_token,
		arg_depositId,
	)

	if err != nil {
		return result, p.errorResolver.ResolveError(
			err,
			p.callerOptions.From,
			nil,
			"getDeposit",
			arg_depositor,
			arg_token,
			arg_depositId,
		)
	}

	return result, err
}

func (p *Portal) GetDepositAtBlock(
	arg_depositor common.Address,
	arg_token common.Address,
	arg_depositId *big.Int,
	blockNumber *big.Int,
) (abi.PortalDepositInfo, error) {
	var result abi.PortalDepositInfo

	err := chainutil.CallAtBlock(
		p.callerOptions.From,
		blockNumber,
		nil,
		p.contractABI,
		p.caller,
		p.errorResolver,
		p.contractAddress,
		"getDeposit",
		&result,
		arg_depositor,
		arg_token,
		arg_depositId,
	)

	return result, err
}

func (p *Portal) MaxLockPeriod() (uint32, error) {
	result, err := p.contract.MaxLockPeriod(
		p.callerOptions,
	)

	if err != nil {
		return result, p.errorResolver.ResolveError(
			err,
			p.callerOptions.From,
			nil,
			"maxLockPeriod",
		)
	}

	return result, err
}

func (p *Portal) MaxLockPeriodAtBlock(
	blockNumber *big.Int,
) (uint32, error) {
	var result uint32

	err := chainutil.CallAtBlock(
		p.callerOptions.From,
		blockNumber,
		nil,
		p.contractABI,
		p.caller,
		p.errorResolver,
		p.contractAddress,
		"maxLockPeriod",
		&result,
	)

	return result, err
}

func (p *Portal) MinLockPeriod() (uint32, error) {
	result, err := p.contract.MinLockPeriod(
		p.callerOptions,
	)

	if err != nil {
		return result, p.errorResolver.ResolveError(
			err,
			p.callerOptions.From,
			nil,
			"minLockPeriod",
		)
	}

	return result, err
}

func (p *Portal) MinLockPeriodAtBlock(
	blockNumber *big.Int,
) (uint32, error) {
	var result uint32

	err := chainutil.CallAtBlock(
		p.callerOptions.From,
		blockNumber,
		nil,
		p.contractABI,
		p.caller,
		p.errorResolver,
		p.contractAddress,
		"minLockPeriod",
		&result,
	)

	return result, err
}

func (p *Portal) Owner() (common.Address, error) {
	result, err := p.contract.Owner(
		p.callerOptions,
	)

	if err != nil {
		return result, p.errorResolver.ResolveError(
			err,
			p.callerOptions.From,
			nil,
			"owner",
		)
	}

	return result, err
}

func (p *Portal) OwnerAtBlock(
	blockNumber *big.Int,
) (common.Address, error) {
	var result common.Address

	err := chainutil.CallAtBlock(
		p.callerOptions.From,
		blockNumber,
		nil,
		p.contractABI,
		p.caller,
		p.errorResolver,
		p.contractAddress,
		"owner",
		&result,
	)

	return result, err
}

func (p *Portal) PendingOwner() (common.Address, error) {
	result, err := p.contract.PendingOwner(
		p.callerOptions,
	)

	if err != nil {
		return result, p.errorResolver.ResolveError(
			err,
			p.callerOptions.From,
			nil,
			"pendingOwner",
		)
	}

	return result, err
}

func (p *Portal) PendingOwnerAtBlock(
	blockNumber *big.Int,
) (common.Address, error) {
	var result common.Address

	err := chainutil.CallAtBlock(
		p.callerOptions.From,
		blockNumber,
		nil,
		p.contractABI,
		p.caller,
		p.errorResolver,
		p.contractAddress,
		"pendingOwner",
		&result,
	)

	return result, err
}

func (p *Portal) TokenAbility(
	arg0 common.Address,
) (uint8, error) {
	result, err := p.contract.TokenAbility(
		p.callerOptions,
		arg0,
	)

	if err != nil {
		return result, p.errorResolver.ResolveError(
			err,
			p.callerOptions.From,
			nil,
			"tokenAbility",
			arg0,
		)
	}

	return result, err
}

func (p *Portal) TokenAbilityAtBlock(
	arg0 common.Address,
	blockNumber *big.Int,
) (uint8, error) {
	var result uint8

	err := chainutil.CallAtBlock(
		p.callerOptions.From,
		blockNumber,
		nil,
		p.contractABI,
		p.caller,
		p.errorResolver,
		p.contractAddress,
		"tokenAbility",
		&result,
		arg0,
	)

	return result, err
}

// ------ Events -------

func (p *Portal) DepositedEvent(
	opts *ethereum.SubscribeOpts,
	depositorFilter []common.Address,
	tokenFilter []common.Address,
	depositIdFilter []*big.Int,
) *PDepositedSubscription {
	if opts == nil {
		opts = new(ethereum.SubscribeOpts)
	}
	if opts.Tick == 0 {
		opts.Tick = chainutil.DefaultSubscribeOptsTick
	}
	if opts.PastBlocks == 0 {
		opts.PastBlocks = chainutil.DefaultSubscribeOptsPastBlocks
	}

	return &PDepositedSubscription{
		p,
		opts,
		depositorFilter,
		tokenFilter,
		depositIdFilter,
	}
}

type PDepositedSubscription struct {
	contract        *Portal
	opts            *ethereum.SubscribeOpts
	depositorFilter []common.Address
	tokenFilter     []common.Address
	depositIdFilter []*big.Int
}

type portalDepositedFunc func(
	Depositor common.Address,
	Token common.Address,
	DepositId *big.Int,
	Amount *big.Int,
	blockNumber uint64,
)

func (ds *PDepositedSubscription) OnEvent(
	handler portalDepositedFunc,
) subscription.EventSubscription {
	eventChan := make(chan *abi.PortalDeposited)
	ctx, cancelCtx := context.WithCancel(context.Background())

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case event := <-eventChan:
				handler(
					event.Depositor,
					event.Token,
					event.DepositId,
					event.Amount,
					event.Raw.BlockNumber,
				)
			}
		}
	}()

	sub := ds.Pipe(eventChan)
	return subscription.NewEventSubscription(func() {
		sub.Unsubscribe()
		cancelCtx()
	})
}

func (ds *PDepositedSubscription) Pipe(
	sink chan *abi.PortalDeposited,
) subscription.EventSubscription {
	ctx, cancelCtx := context.WithCancel(context.Background())
	go func() {
		ticker := time.NewTicker(ds.opts.Tick)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				lastBlock, err := ds.contract.blockCounter.CurrentBlock()
				if err != nil {
					pLogger.Errorf(
						"subscription failed to pull events: [%v]",
						err,
					)
				}
				fromBlock := lastBlock - ds.opts.PastBlocks

				pLogger.Infof(
					"subscription monitoring fetching past Deposited events "+
						"starting from block [%v]",
					fromBlock,
				)
				events, err := ds.contract.PastDepositedEvents(
					fromBlock,
					nil,
					ds.depositorFilter,
					ds.tokenFilter,
					ds.depositIdFilter,
				)
				if err != nil {
					pLogger.Errorf(
						"subscription failed to pull events: [%v]",
						err,
					)
					continue
				}
				pLogger.Infof(
					"subscription monitoring fetched [%v] past Deposited events",
					len(events),
				)

				for _, event := range events {
					sink <- event
				}
			}
		}
	}()

	sub := ds.contract.watchDeposited(
		sink,
		ds.depositorFilter,
		ds.tokenFilter,
		ds.depositIdFilter,
	)

	return subscription.NewEventSubscription(func() {
		sub.Unsubscribe()
		cancelCtx()
	})
}

func (p *Portal) watchDeposited(
	sink chan *abi.PortalDeposited,
	depositorFilter []common.Address,
	tokenFilter []common.Address,
	depositIdFilter []*big.Int,
) event.Subscription {
	subscribeFn := func(ctx context.Context) (event.Subscription, error) {
		return p.contract.WatchDeposited(
			&bind.WatchOpts{Context: ctx},
			sink,
			depositorFilter,
			tokenFilter,
			depositIdFilter,
		)
	}

	thresholdViolatedFn := func(elapsed time.Duration) {
		pLogger.Warnf(
			"subscription to event Deposited had to be "+
				"retried [%s] since the last attempt; please inspect "+
				"host chain connectivity",
			elapsed,
		)
	}

	subscriptionFailedFn := func(err error) {
		pLogger.Errorf(
			"subscription to event Deposited failed "+
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

func (p *Portal) PastDepositedEvents(
	startBlock uint64,
	endBlock *uint64,
	depositorFilter []common.Address,
	tokenFilter []common.Address,
	depositIdFilter []*big.Int,
) ([]*abi.PortalDeposited, error) {
	iterator, err := p.contract.FilterDeposited(
		&bind.FilterOpts{
			Start: startBlock,
			End:   endBlock,
		},
		depositorFilter,
		tokenFilter,
		depositIdFilter,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"error retrieving past Deposited events: [%v]",
			err,
		)
	}

	events := make([]*abi.PortalDeposited, 0)

	for iterator.Next() {
		event := iterator.Event
		events = append(events, event)
	}

	return events, nil
}

func (p *Portal) InitializedEvent(
	opts *ethereum.SubscribeOpts,
) *PInitializedSubscription {
	if opts == nil {
		opts = new(ethereum.SubscribeOpts)
	}
	if opts.Tick == 0 {
		opts.Tick = chainutil.DefaultSubscribeOptsTick
	}
	if opts.PastBlocks == 0 {
		opts.PastBlocks = chainutil.DefaultSubscribeOptsPastBlocks
	}

	return &PInitializedSubscription{
		p,
		opts,
	}
}

type PInitializedSubscription struct {
	contract *Portal
	opts     *ethereum.SubscribeOpts
}

type portalInitializedFunc func(
	Version uint64,
	blockNumber uint64,
)

func (is *PInitializedSubscription) OnEvent(
	handler portalInitializedFunc,
) subscription.EventSubscription {
	eventChan := make(chan *abi.PortalInitialized)
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

func (is *PInitializedSubscription) Pipe(
	sink chan *abi.PortalInitialized,
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
					pLogger.Errorf(
						"subscription failed to pull events: [%v]",
						err,
					)
				}
				fromBlock := lastBlock - is.opts.PastBlocks

				pLogger.Infof(
					"subscription monitoring fetching past Initialized events "+
						"starting from block [%v]",
					fromBlock,
				)
				events, err := is.contract.PastInitializedEvents(
					fromBlock,
					nil,
				)
				if err != nil {
					pLogger.Errorf(
						"subscription failed to pull events: [%v]",
						err,
					)
					continue
				}
				pLogger.Infof(
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

func (p *Portal) watchInitialized(
	sink chan *abi.PortalInitialized,
) event.Subscription {
	subscribeFn := func(ctx context.Context) (event.Subscription, error) {
		return p.contract.WatchInitialized(
			&bind.WatchOpts{Context: ctx},
			sink,
		)
	}

	thresholdViolatedFn := func(elapsed time.Duration) {
		pLogger.Warnf(
			"subscription to event Initialized had to be "+
				"retried [%s] since the last attempt; please inspect "+
				"host chain connectivity",
			elapsed,
		)
	}

	subscriptionFailedFn := func(err error) {
		pLogger.Errorf(
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

func (p *Portal) PastInitializedEvents(
	startBlock uint64,
	endBlock *uint64,
) ([]*abi.PortalInitialized, error) {
	iterator, err := p.contract.FilterInitialized(
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

	events := make([]*abi.PortalInitialized, 0)

	for iterator.Next() {
		event := iterator.Event
		events = append(events, event)
	}

	return events, nil
}

func (p *Portal) LockedEvent(
	opts *ethereum.SubscribeOpts,
	depositorFilter []common.Address,
	tokenFilter []common.Address,
	depositIdFilter []*big.Int,
) *PLockedSubscription {
	if opts == nil {
		opts = new(ethereum.SubscribeOpts)
	}
	if opts.Tick == 0 {
		opts.Tick = chainutil.DefaultSubscribeOptsTick
	}
	if opts.PastBlocks == 0 {
		opts.PastBlocks = chainutil.DefaultSubscribeOptsPastBlocks
	}

	return &PLockedSubscription{
		p,
		opts,
		depositorFilter,
		tokenFilter,
		depositIdFilter,
	}
}

type PLockedSubscription struct {
	contract        *Portal
	opts            *ethereum.SubscribeOpts
	depositorFilter []common.Address
	tokenFilter     []common.Address
	depositIdFilter []*big.Int
}

type portalLockedFunc func(
	Depositor common.Address,
	Token common.Address,
	DepositId *big.Int,
	UnlockAt uint32,
	LockPeriod uint32,
	blockNumber uint64,
)

func (ls *PLockedSubscription) OnEvent(
	handler portalLockedFunc,
) subscription.EventSubscription {
	eventChan := make(chan *abi.PortalLocked)
	ctx, cancelCtx := context.WithCancel(context.Background())

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case event := <-eventChan:
				handler(
					event.Depositor,
					event.Token,
					event.DepositId,
					event.UnlockAt,
					event.LockPeriod,
					event.Raw.BlockNumber,
				)
			}
		}
	}()

	sub := ls.Pipe(eventChan)
	return subscription.NewEventSubscription(func() {
		sub.Unsubscribe()
		cancelCtx()
	})
}

func (ls *PLockedSubscription) Pipe(
	sink chan *abi.PortalLocked,
) subscription.EventSubscription {
	ctx, cancelCtx := context.WithCancel(context.Background())
	go func() {
		ticker := time.NewTicker(ls.opts.Tick)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				lastBlock, err := ls.contract.blockCounter.CurrentBlock()
				if err != nil {
					pLogger.Errorf(
						"subscription failed to pull events: [%v]",
						err,
					)
				}
				fromBlock := lastBlock - ls.opts.PastBlocks

				pLogger.Infof(
					"subscription monitoring fetching past Locked events "+
						"starting from block [%v]",
					fromBlock,
				)
				events, err := ls.contract.PastLockedEvents(
					fromBlock,
					nil,
					ls.depositorFilter,
					ls.tokenFilter,
					ls.depositIdFilter,
				)
				if err != nil {
					pLogger.Errorf(
						"subscription failed to pull events: [%v]",
						err,
					)
					continue
				}
				pLogger.Infof(
					"subscription monitoring fetched [%v] past Locked events",
					len(events),
				)

				for _, event := range events {
					sink <- event
				}
			}
		}
	}()

	sub := ls.contract.watchLocked(
		sink,
		ls.depositorFilter,
		ls.tokenFilter,
		ls.depositIdFilter,
	)

	return subscription.NewEventSubscription(func() {
		sub.Unsubscribe()
		cancelCtx()
	})
}

func (p *Portal) watchLocked(
	sink chan *abi.PortalLocked,
	depositorFilter []common.Address,
	tokenFilter []common.Address,
	depositIdFilter []*big.Int,
) event.Subscription {
	subscribeFn := func(ctx context.Context) (event.Subscription, error) {
		return p.contract.WatchLocked(
			&bind.WatchOpts{Context: ctx},
			sink,
			depositorFilter,
			tokenFilter,
			depositIdFilter,
		)
	}

	thresholdViolatedFn := func(elapsed time.Duration) {
		pLogger.Warnf(
			"subscription to event Locked had to be "+
				"retried [%s] since the last attempt; please inspect "+
				"host chain connectivity",
			elapsed,
		)
	}

	subscriptionFailedFn := func(err error) {
		pLogger.Errorf(
			"subscription to event Locked failed "+
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

func (p *Portal) PastLockedEvents(
	startBlock uint64,
	endBlock *uint64,
	depositorFilter []common.Address,
	tokenFilter []common.Address,
	depositIdFilter []*big.Int,
) ([]*abi.PortalLocked, error) {
	iterator, err := p.contract.FilterLocked(
		&bind.FilterOpts{
			Start: startBlock,
			End:   endBlock,
		},
		depositorFilter,
		tokenFilter,
		depositIdFilter,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"error retrieving past Locked events: [%v]",
			err,
		)
	}

	events := make([]*abi.PortalLocked, 0)

	for iterator.Next() {
		event := iterator.Event
		events = append(events, event)
	}

	return events, nil
}

func (p *Portal) MaxLockPeriodUpdatedEvent(
	opts *ethereum.SubscribeOpts,
) *PMaxLockPeriodUpdatedSubscription {
	if opts == nil {
		opts = new(ethereum.SubscribeOpts)
	}
	if opts.Tick == 0 {
		opts.Tick = chainutil.DefaultSubscribeOptsTick
	}
	if opts.PastBlocks == 0 {
		opts.PastBlocks = chainutil.DefaultSubscribeOptsPastBlocks
	}

	return &PMaxLockPeriodUpdatedSubscription{
		p,
		opts,
	}
}

type PMaxLockPeriodUpdatedSubscription struct {
	contract *Portal
	opts     *ethereum.SubscribeOpts
}

type portalMaxLockPeriodUpdatedFunc func(
	MaxLockPeriod uint32,
	blockNumber uint64,
)

func (mlpus *PMaxLockPeriodUpdatedSubscription) OnEvent(
	handler portalMaxLockPeriodUpdatedFunc,
) subscription.EventSubscription {
	eventChan := make(chan *abi.PortalMaxLockPeriodUpdated)
	ctx, cancelCtx := context.WithCancel(context.Background())

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case event := <-eventChan:
				handler(
					event.MaxLockPeriod,
					event.Raw.BlockNumber,
				)
			}
		}
	}()

	sub := mlpus.Pipe(eventChan)
	return subscription.NewEventSubscription(func() {
		sub.Unsubscribe()
		cancelCtx()
	})
}

func (mlpus *PMaxLockPeriodUpdatedSubscription) Pipe(
	sink chan *abi.PortalMaxLockPeriodUpdated,
) subscription.EventSubscription {
	ctx, cancelCtx := context.WithCancel(context.Background())
	go func() {
		ticker := time.NewTicker(mlpus.opts.Tick)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				lastBlock, err := mlpus.contract.blockCounter.CurrentBlock()
				if err != nil {
					pLogger.Errorf(
						"subscription failed to pull events: [%v]",
						err,
					)
				}
				fromBlock := lastBlock - mlpus.opts.PastBlocks

				pLogger.Infof(
					"subscription monitoring fetching past MaxLockPeriodUpdated events "+
						"starting from block [%v]",
					fromBlock,
				)
				events, err := mlpus.contract.PastMaxLockPeriodUpdatedEvents(
					fromBlock,
					nil,
				)
				if err != nil {
					pLogger.Errorf(
						"subscription failed to pull events: [%v]",
						err,
					)
					continue
				}
				pLogger.Infof(
					"subscription monitoring fetched [%v] past MaxLockPeriodUpdated events",
					len(events),
				)

				for _, event := range events {
					sink <- event
				}
			}
		}
	}()

	sub := mlpus.contract.watchMaxLockPeriodUpdated(
		sink,
	)

	return subscription.NewEventSubscription(func() {
		sub.Unsubscribe()
		cancelCtx()
	})
}

func (p *Portal) watchMaxLockPeriodUpdated(
	sink chan *abi.PortalMaxLockPeriodUpdated,
) event.Subscription {
	subscribeFn := func(ctx context.Context) (event.Subscription, error) {
		return p.contract.WatchMaxLockPeriodUpdated(
			&bind.WatchOpts{Context: ctx},
			sink,
		)
	}

	thresholdViolatedFn := func(elapsed time.Duration) {
		pLogger.Warnf(
			"subscription to event MaxLockPeriodUpdated had to be "+
				"retried [%s] since the last attempt; please inspect "+
				"host chain connectivity",
			elapsed,
		)
	}

	subscriptionFailedFn := func(err error) {
		pLogger.Errorf(
			"subscription to event MaxLockPeriodUpdated failed "+
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

func (p *Portal) PastMaxLockPeriodUpdatedEvents(
	startBlock uint64,
	endBlock *uint64,
) ([]*abi.PortalMaxLockPeriodUpdated, error) {
	iterator, err := p.contract.FilterMaxLockPeriodUpdated(
		&bind.FilterOpts{
			Start: startBlock,
			End:   endBlock,
		},
	)
	if err != nil {
		return nil, fmt.Errorf(
			"error retrieving past MaxLockPeriodUpdated events: [%v]",
			err,
		)
	}

	events := make([]*abi.PortalMaxLockPeriodUpdated, 0)

	for iterator.Next() {
		event := iterator.Event
		events = append(events, event)
	}

	return events, nil
}

func (p *Portal) MinLockPeriodUpdatedEvent(
	opts *ethereum.SubscribeOpts,
) *PMinLockPeriodUpdatedSubscription {
	if opts == nil {
		opts = new(ethereum.SubscribeOpts)
	}
	if opts.Tick == 0 {
		opts.Tick = chainutil.DefaultSubscribeOptsTick
	}
	if opts.PastBlocks == 0 {
		opts.PastBlocks = chainutil.DefaultSubscribeOptsPastBlocks
	}

	return &PMinLockPeriodUpdatedSubscription{
		p,
		opts,
	}
}

type PMinLockPeriodUpdatedSubscription struct {
	contract *Portal
	opts     *ethereum.SubscribeOpts
}

type portalMinLockPeriodUpdatedFunc func(
	MinLockPeriod uint32,
	blockNumber uint64,
)

func (mlpus *PMinLockPeriodUpdatedSubscription) OnEvent(
	handler portalMinLockPeriodUpdatedFunc,
) subscription.EventSubscription {
	eventChan := make(chan *abi.PortalMinLockPeriodUpdated)
	ctx, cancelCtx := context.WithCancel(context.Background())

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case event := <-eventChan:
				handler(
					event.MinLockPeriod,
					event.Raw.BlockNumber,
				)
			}
		}
	}()

	sub := mlpus.Pipe(eventChan)
	return subscription.NewEventSubscription(func() {
		sub.Unsubscribe()
		cancelCtx()
	})
}

func (mlpus *PMinLockPeriodUpdatedSubscription) Pipe(
	sink chan *abi.PortalMinLockPeriodUpdated,
) subscription.EventSubscription {
	ctx, cancelCtx := context.WithCancel(context.Background())
	go func() {
		ticker := time.NewTicker(mlpus.opts.Tick)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				lastBlock, err := mlpus.contract.blockCounter.CurrentBlock()
				if err != nil {
					pLogger.Errorf(
						"subscription failed to pull events: [%v]",
						err,
					)
				}
				fromBlock := lastBlock - mlpus.opts.PastBlocks

				pLogger.Infof(
					"subscription monitoring fetching past MinLockPeriodUpdated events "+
						"starting from block [%v]",
					fromBlock,
				)
				events, err := mlpus.contract.PastMinLockPeriodUpdatedEvents(
					fromBlock,
					nil,
				)
				if err != nil {
					pLogger.Errorf(
						"subscription failed to pull events: [%v]",
						err,
					)
					continue
				}
				pLogger.Infof(
					"subscription monitoring fetched [%v] past MinLockPeriodUpdated events",
					len(events),
				)

				for _, event := range events {
					sink <- event
				}
			}
		}
	}()

	sub := mlpus.contract.watchMinLockPeriodUpdated(
		sink,
	)

	return subscription.NewEventSubscription(func() {
		sub.Unsubscribe()
		cancelCtx()
	})
}

func (p *Portal) watchMinLockPeriodUpdated(
	sink chan *abi.PortalMinLockPeriodUpdated,
) event.Subscription {
	subscribeFn := func(ctx context.Context) (event.Subscription, error) {
		return p.contract.WatchMinLockPeriodUpdated(
			&bind.WatchOpts{Context: ctx},
			sink,
		)
	}

	thresholdViolatedFn := func(elapsed time.Duration) {
		pLogger.Warnf(
			"subscription to event MinLockPeriodUpdated had to be "+
				"retried [%s] since the last attempt; please inspect "+
				"host chain connectivity",
			elapsed,
		)
	}

	subscriptionFailedFn := func(err error) {
		pLogger.Errorf(
			"subscription to event MinLockPeriodUpdated failed "+
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

func (p *Portal) PastMinLockPeriodUpdatedEvents(
	startBlock uint64,
	endBlock *uint64,
) ([]*abi.PortalMinLockPeriodUpdated, error) {
	iterator, err := p.contract.FilterMinLockPeriodUpdated(
		&bind.FilterOpts{
			Start: startBlock,
			End:   endBlock,
		},
	)
	if err != nil {
		return nil, fmt.Errorf(
			"error retrieving past MinLockPeriodUpdated events: [%v]",
			err,
		)
	}

	events := make([]*abi.PortalMinLockPeriodUpdated, 0)

	for iterator.Next() {
		event := iterator.Event
		events = append(events, event)
	}

	return events, nil
}

func (p *Portal) OwnershipTransferStartedEvent(
	opts *ethereum.SubscribeOpts,
	previousOwnerFilter []common.Address,
	newOwnerFilter []common.Address,
) *POwnershipTransferStartedSubscription {
	if opts == nil {
		opts = new(ethereum.SubscribeOpts)
	}
	if opts.Tick == 0 {
		opts.Tick = chainutil.DefaultSubscribeOptsTick
	}
	if opts.PastBlocks == 0 {
		opts.PastBlocks = chainutil.DefaultSubscribeOptsPastBlocks
	}

	return &POwnershipTransferStartedSubscription{
		p,
		opts,
		previousOwnerFilter,
		newOwnerFilter,
	}
}

type POwnershipTransferStartedSubscription struct {
	contract            *Portal
	opts                *ethereum.SubscribeOpts
	previousOwnerFilter []common.Address
	newOwnerFilter      []common.Address
}

type portalOwnershipTransferStartedFunc func(
	PreviousOwner common.Address,
	NewOwner common.Address,
	blockNumber uint64,
)

func (otss *POwnershipTransferStartedSubscription) OnEvent(
	handler portalOwnershipTransferStartedFunc,
) subscription.EventSubscription {
	eventChan := make(chan *abi.PortalOwnershipTransferStarted)
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

func (otss *POwnershipTransferStartedSubscription) Pipe(
	sink chan *abi.PortalOwnershipTransferStarted,
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
					pLogger.Errorf(
						"subscription failed to pull events: [%v]",
						err,
					)
				}
				fromBlock := lastBlock - otss.opts.PastBlocks

				pLogger.Infof(
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
					pLogger.Errorf(
						"subscription failed to pull events: [%v]",
						err,
					)
					continue
				}
				pLogger.Infof(
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

func (p *Portal) watchOwnershipTransferStarted(
	sink chan *abi.PortalOwnershipTransferStarted,
	previousOwnerFilter []common.Address,
	newOwnerFilter []common.Address,
) event.Subscription {
	subscribeFn := func(ctx context.Context) (event.Subscription, error) {
		return p.contract.WatchOwnershipTransferStarted(
			&bind.WatchOpts{Context: ctx},
			sink,
			previousOwnerFilter,
			newOwnerFilter,
		)
	}

	thresholdViolatedFn := func(elapsed time.Duration) {
		pLogger.Warnf(
			"subscription to event OwnershipTransferStarted had to be "+
				"retried [%s] since the last attempt; please inspect "+
				"host chain connectivity",
			elapsed,
		)
	}

	subscriptionFailedFn := func(err error) {
		pLogger.Errorf(
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

func (p *Portal) PastOwnershipTransferStartedEvents(
	startBlock uint64,
	endBlock *uint64,
	previousOwnerFilter []common.Address,
	newOwnerFilter []common.Address,
) ([]*abi.PortalOwnershipTransferStarted, error) {
	iterator, err := p.contract.FilterOwnershipTransferStarted(
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

	events := make([]*abi.PortalOwnershipTransferStarted, 0)

	for iterator.Next() {
		event := iterator.Event
		events = append(events, event)
	}

	return events, nil
}

func (p *Portal) OwnershipTransferredEvent(
	opts *ethereum.SubscribeOpts,
	previousOwnerFilter []common.Address,
	newOwnerFilter []common.Address,
) *POwnershipTransferredSubscription {
	if opts == nil {
		opts = new(ethereum.SubscribeOpts)
	}
	if opts.Tick == 0 {
		opts.Tick = chainutil.DefaultSubscribeOptsTick
	}
	if opts.PastBlocks == 0 {
		opts.PastBlocks = chainutil.DefaultSubscribeOptsPastBlocks
	}

	return &POwnershipTransferredSubscription{
		p,
		opts,
		previousOwnerFilter,
		newOwnerFilter,
	}
}

type POwnershipTransferredSubscription struct {
	contract            *Portal
	opts                *ethereum.SubscribeOpts
	previousOwnerFilter []common.Address
	newOwnerFilter      []common.Address
}

type portalOwnershipTransferredFunc func(
	PreviousOwner common.Address,
	NewOwner common.Address,
	blockNumber uint64,
)

func (ots *POwnershipTransferredSubscription) OnEvent(
	handler portalOwnershipTransferredFunc,
) subscription.EventSubscription {
	eventChan := make(chan *abi.PortalOwnershipTransferred)
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

func (ots *POwnershipTransferredSubscription) Pipe(
	sink chan *abi.PortalOwnershipTransferred,
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
					pLogger.Errorf(
						"subscription failed to pull events: [%v]",
						err,
					)
				}
				fromBlock := lastBlock - ots.opts.PastBlocks

				pLogger.Infof(
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
					pLogger.Errorf(
						"subscription failed to pull events: [%v]",
						err,
					)
					continue
				}
				pLogger.Infof(
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

func (p *Portal) watchOwnershipTransferred(
	sink chan *abi.PortalOwnershipTransferred,
	previousOwnerFilter []common.Address,
	newOwnerFilter []common.Address,
) event.Subscription {
	subscribeFn := func(ctx context.Context) (event.Subscription, error) {
		return p.contract.WatchOwnershipTransferred(
			&bind.WatchOpts{Context: ctx},
			sink,
			previousOwnerFilter,
			newOwnerFilter,
		)
	}

	thresholdViolatedFn := func(elapsed time.Duration) {
		pLogger.Warnf(
			"subscription to event OwnershipTransferred had to be "+
				"retried [%s] since the last attempt; please inspect "+
				"host chain connectivity",
			elapsed,
		)
	}

	subscriptionFailedFn := func(err error) {
		pLogger.Errorf(
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

func (p *Portal) PastOwnershipTransferredEvents(
	startBlock uint64,
	endBlock *uint64,
	previousOwnerFilter []common.Address,
	newOwnerFilter []common.Address,
) ([]*abi.PortalOwnershipTransferred, error) {
	iterator, err := p.contract.FilterOwnershipTransferred(
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

	events := make([]*abi.PortalOwnershipTransferred, 0)

	for iterator.Next() {
		event := iterator.Event
		events = append(events, event)
	}

	return events, nil
}

func (p *Portal) SupportedTokenAddedEvent(
	opts *ethereum.SubscribeOpts,
	tokenFilter []common.Address,
) *PSupportedTokenAddedSubscription {
	if opts == nil {
		opts = new(ethereum.SubscribeOpts)
	}
	if opts.Tick == 0 {
		opts.Tick = chainutil.DefaultSubscribeOptsTick
	}
	if opts.PastBlocks == 0 {
		opts.PastBlocks = chainutil.DefaultSubscribeOptsPastBlocks
	}

	return &PSupportedTokenAddedSubscription{
		p,
		opts,
		tokenFilter,
	}
}

type PSupportedTokenAddedSubscription struct {
	contract    *Portal
	opts        *ethereum.SubscribeOpts
	tokenFilter []common.Address
}

type portalSupportedTokenAddedFunc func(
	Token common.Address,
	TokenAbility uint8,
	blockNumber uint64,
)

func (stas *PSupportedTokenAddedSubscription) OnEvent(
	handler portalSupportedTokenAddedFunc,
) subscription.EventSubscription {
	eventChan := make(chan *abi.PortalSupportedTokenAdded)
	ctx, cancelCtx := context.WithCancel(context.Background())

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case event := <-eventChan:
				handler(
					event.Token,
					event.TokenAbility,
					event.Raw.BlockNumber,
				)
			}
		}
	}()

	sub := stas.Pipe(eventChan)
	return subscription.NewEventSubscription(func() {
		sub.Unsubscribe()
		cancelCtx()
	})
}

func (stas *PSupportedTokenAddedSubscription) Pipe(
	sink chan *abi.PortalSupportedTokenAdded,
) subscription.EventSubscription {
	ctx, cancelCtx := context.WithCancel(context.Background())
	go func() {
		ticker := time.NewTicker(stas.opts.Tick)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				lastBlock, err := stas.contract.blockCounter.CurrentBlock()
				if err != nil {
					pLogger.Errorf(
						"subscription failed to pull events: [%v]",
						err,
					)
				}
				fromBlock := lastBlock - stas.opts.PastBlocks

				pLogger.Infof(
					"subscription monitoring fetching past SupportedTokenAdded events "+
						"starting from block [%v]",
					fromBlock,
				)
				events, err := stas.contract.PastSupportedTokenAddedEvents(
					fromBlock,
					nil,
					stas.tokenFilter,
				)
				if err != nil {
					pLogger.Errorf(
						"subscription failed to pull events: [%v]",
						err,
					)
					continue
				}
				pLogger.Infof(
					"subscription monitoring fetched [%v] past SupportedTokenAdded events",
					len(events),
				)

				for _, event := range events {
					sink <- event
				}
			}
		}
	}()

	sub := stas.contract.watchSupportedTokenAdded(
		sink,
		stas.tokenFilter,
	)

	return subscription.NewEventSubscription(func() {
		sub.Unsubscribe()
		cancelCtx()
	})
}

func (p *Portal) watchSupportedTokenAdded(
	sink chan *abi.PortalSupportedTokenAdded,
	tokenFilter []common.Address,
) event.Subscription {
	subscribeFn := func(ctx context.Context) (event.Subscription, error) {
		return p.contract.WatchSupportedTokenAdded(
			&bind.WatchOpts{Context: ctx},
			sink,
			tokenFilter,
		)
	}

	thresholdViolatedFn := func(elapsed time.Duration) {
		pLogger.Warnf(
			"subscription to event SupportedTokenAdded had to be "+
				"retried [%s] since the last attempt; please inspect "+
				"host chain connectivity",
			elapsed,
		)
	}

	subscriptionFailedFn := func(err error) {
		pLogger.Errorf(
			"subscription to event SupportedTokenAdded failed "+
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

func (p *Portal) PastSupportedTokenAddedEvents(
	startBlock uint64,
	endBlock *uint64,
	tokenFilter []common.Address,
) ([]*abi.PortalSupportedTokenAdded, error) {
	iterator, err := p.contract.FilterSupportedTokenAdded(
		&bind.FilterOpts{
			Start: startBlock,
			End:   endBlock,
		},
		tokenFilter,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"error retrieving past SupportedTokenAdded events: [%v]",
			err,
		)
	}

	events := make([]*abi.PortalSupportedTokenAdded, 0)

	for iterator.Next() {
		event := iterator.Event
		events = append(events, event)
	}

	return events, nil
}

func (p *Portal) WithdrawnEvent(
	opts *ethereum.SubscribeOpts,
	depositorFilter []common.Address,
	tokenFilter []common.Address,
	depositIdFilter []*big.Int,
) *PWithdrawnSubscription {
	if opts == nil {
		opts = new(ethereum.SubscribeOpts)
	}
	if opts.Tick == 0 {
		opts.Tick = chainutil.DefaultSubscribeOptsTick
	}
	if opts.PastBlocks == 0 {
		opts.PastBlocks = chainutil.DefaultSubscribeOptsPastBlocks
	}

	return &PWithdrawnSubscription{
		p,
		opts,
		depositorFilter,
		tokenFilter,
		depositIdFilter,
	}
}

type PWithdrawnSubscription struct {
	contract        *Portal
	opts            *ethereum.SubscribeOpts
	depositorFilter []common.Address
	tokenFilter     []common.Address
	depositIdFilter []*big.Int
}

type portalWithdrawnFunc func(
	Depositor common.Address,
	Token common.Address,
	DepositId *big.Int,
	Amount *big.Int,
	blockNumber uint64,
)

func (ws *PWithdrawnSubscription) OnEvent(
	handler portalWithdrawnFunc,
) subscription.EventSubscription {
	eventChan := make(chan *abi.PortalWithdrawn)
	ctx, cancelCtx := context.WithCancel(context.Background())

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case event := <-eventChan:
				handler(
					event.Depositor,
					event.Token,
					event.DepositId,
					event.Amount,
					event.Raw.BlockNumber,
				)
			}
		}
	}()

	sub := ws.Pipe(eventChan)
	return subscription.NewEventSubscription(func() {
		sub.Unsubscribe()
		cancelCtx()
	})
}

func (ws *PWithdrawnSubscription) Pipe(
	sink chan *abi.PortalWithdrawn,
) subscription.EventSubscription {
	ctx, cancelCtx := context.WithCancel(context.Background())
	go func() {
		ticker := time.NewTicker(ws.opts.Tick)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				lastBlock, err := ws.contract.blockCounter.CurrentBlock()
				if err != nil {
					pLogger.Errorf(
						"subscription failed to pull events: [%v]",
						err,
					)
				}
				fromBlock := lastBlock - ws.opts.PastBlocks

				pLogger.Infof(
					"subscription monitoring fetching past Withdrawn events "+
						"starting from block [%v]",
					fromBlock,
				)
				events, err := ws.contract.PastWithdrawnEvents(
					fromBlock,
					nil,
					ws.depositorFilter,
					ws.tokenFilter,
					ws.depositIdFilter,
				)
				if err != nil {
					pLogger.Errorf(
						"subscription failed to pull events: [%v]",
						err,
					)
					continue
				}
				pLogger.Infof(
					"subscription monitoring fetched [%v] past Withdrawn events",
					len(events),
				)

				for _, event := range events {
					sink <- event
				}
			}
		}
	}()

	sub := ws.contract.watchWithdrawn(
		sink,
		ws.depositorFilter,
		ws.tokenFilter,
		ws.depositIdFilter,
	)

	return subscription.NewEventSubscription(func() {
		sub.Unsubscribe()
		cancelCtx()
	})
}

func (p *Portal) watchWithdrawn(
	sink chan *abi.PortalWithdrawn,
	depositorFilter []common.Address,
	tokenFilter []common.Address,
	depositIdFilter []*big.Int,
) event.Subscription {
	subscribeFn := func(ctx context.Context) (event.Subscription, error) {
		return p.contract.WatchWithdrawn(
			&bind.WatchOpts{Context: ctx},
			sink,
			depositorFilter,
			tokenFilter,
			depositIdFilter,
		)
	}

	thresholdViolatedFn := func(elapsed time.Duration) {
		pLogger.Warnf(
			"subscription to event Withdrawn had to be "+
				"retried [%s] since the last attempt; please inspect "+
				"host chain connectivity",
			elapsed,
		)
	}

	subscriptionFailedFn := func(err error) {
		pLogger.Errorf(
			"subscription to event Withdrawn failed "+
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

func (p *Portal) PastWithdrawnEvents(
	startBlock uint64,
	endBlock *uint64,
	depositorFilter []common.Address,
	tokenFilter []common.Address,
	depositIdFilter []*big.Int,
) ([]*abi.PortalWithdrawn, error) {
	iterator, err := p.contract.FilterWithdrawn(
		&bind.FilterOpts{
			Start: startBlock,
			End:   endBlock,
		},
		depositorFilter,
		tokenFilter,
		depositIdFilter,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"error retrieving past Withdrawn events: [%v]",
			err,
		)
	}

	events := make([]*abi.PortalWithdrawn, 0)

	for iterator.Next() {
		event := iterator.Event
		events = append(events, event)
	}

	return events, nil
}
