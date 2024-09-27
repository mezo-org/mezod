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
	"github.com/mezo-org/mezod/ethereum/bindings/portal/gen/abi"

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
func (p *Portal) CompleteTbtcMigration(
	arg_token common.Address,
	arg_migratedDeposits []abi.PortalDepositToMigrate,

	transactionOptions ...chainutil.TransactionOptions,
) (*types.Transaction, error) {
	pLogger.Debug(
		"submitting transaction completeTbtcMigration",
		" params: ",
		fmt.Sprint(
			arg_token,
			arg_migratedDeposits,
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

	transaction, err := p.contract.CompleteTbtcMigration(
		transactorOptions,
		arg_token,
		arg_migratedDeposits,
	)
	if err != nil {
		return transaction, p.errorResolver.ResolveError(
			err,
			p.transactorOptions.From,
			nil,
			"completeTbtcMigration",
			arg_token,
			arg_migratedDeposits,
		)
	}

	pLogger.Infof(
		"submitted transaction completeTbtcMigration with id: [%s] and nonce [%v]",
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

			transaction, err := p.contract.CompleteTbtcMigration(
				newTransactorOptions,
				arg_token,
				arg_migratedDeposits,
			)
			if err != nil {
				return nil, p.errorResolver.ResolveError(
					err,
					p.transactorOptions.From,
					nil,
					"completeTbtcMigration",
					arg_token,
					arg_migratedDeposits,
				)
			}

			pLogger.Infof(
				"submitted transaction completeTbtcMigration with id: [%s] and nonce [%v]",
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
func (p *Portal) CallCompleteTbtcMigration(
	arg_token common.Address,
	arg_migratedDeposits []abi.PortalDepositToMigrate,
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
		"completeTbtcMigration",
		&result,
		arg_token,
		arg_migratedDeposits,
	)

	return err
}

func (p *Portal) CompleteTbtcMigrationGasEstimate(
	arg_token common.Address,
	arg_migratedDeposits []abi.PortalDepositToMigrate,
) (uint64, error) {
	var result uint64

	result, err := chainutil.EstimateGas(
		p.callerOptions.From,
		p.contractAddress,
		"completeTbtcMigration",
		p.contractABI,
		p.transactor,
		arg_token,
		arg_migratedDeposits,
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
func (p *Portal) MintReceipt(
	arg_token common.Address,
	arg_depositId *big.Int,
	arg_amount *big.Int,

	transactionOptions ...chainutil.TransactionOptions,
) (*types.Transaction, error) {
	pLogger.Debug(
		"submitting transaction mintReceipt",
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

	transaction, err := p.contract.MintReceipt(
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
			"mintReceipt",
			arg_token,
			arg_depositId,
			arg_amount,
		)
	}

	pLogger.Infof(
		"submitted transaction mintReceipt with id: [%s] and nonce [%v]",
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

			transaction, err := p.contract.MintReceipt(
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
					"mintReceipt",
					arg_token,
					arg_depositId,
					arg_amount,
				)
			}

			pLogger.Infof(
				"submitted transaction mintReceipt with id: [%s] and nonce [%v]",
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
func (p *Portal) CallMintReceipt(
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
		"mintReceipt",
		&result,
		arg_token,
		arg_depositId,
		arg_amount,
	)

	return err
}

func (p *Portal) MintReceiptGasEstimate(
	arg_token common.Address,
	arg_depositId *big.Int,
	arg_amount *big.Int,
) (uint64, error) {
	var result uint64

	result, err := chainutil.EstimateGas(
		p.callerOptions.From,
		p.contractAddress,
		"mintReceipt",
		p.contractABI,
		p.transactor,
		arg_token,
		arg_depositId,
		arg_amount,
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
func (p *Portal) RepayReceipt(
	arg_token common.Address,
	arg_depositId *big.Int,
	arg_amount *big.Int,

	transactionOptions ...chainutil.TransactionOptions,
) (*types.Transaction, error) {
	pLogger.Debug(
		"submitting transaction repayReceipt",
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

	transaction, err := p.contract.RepayReceipt(
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
			"repayReceipt",
			arg_token,
			arg_depositId,
			arg_amount,
		)
	}

	pLogger.Infof(
		"submitted transaction repayReceipt with id: [%s] and nonce [%v]",
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

			transaction, err := p.contract.RepayReceipt(
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
					"repayReceipt",
					arg_token,
					arg_depositId,
					arg_amount,
				)
			}

			pLogger.Infof(
				"submitted transaction repayReceipt with id: [%s] and nonce [%v]",
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
func (p *Portal) CallRepayReceipt(
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
		"repayReceipt",
		&result,
		arg_token,
		arg_depositId,
		arg_amount,
	)

	return err
}

func (p *Portal) RepayReceiptGasEstimate(
	arg_token common.Address,
	arg_depositId *big.Int,
	arg_amount *big.Int,
) (uint64, error) {
	var result uint64

	result, err := chainutil.EstimateGas(
		p.callerOptions.From,
		p.contractAddress,
		"repayReceipt",
		p.contractABI,
		p.transactor,
		arg_token,
		arg_depositId,
		arg_amount,
	)

	return result, err
}

// Transaction submission.
func (p *Portal) RequestTbtcMigration(
	arg_token common.Address,
	arg_depositId *big.Int,

	transactionOptions ...chainutil.TransactionOptions,
) (*types.Transaction, error) {
	pLogger.Debug(
		"submitting transaction requestTbtcMigration",
		" params: ",
		fmt.Sprint(
			arg_token,
			arg_depositId,
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

	transaction, err := p.contract.RequestTbtcMigration(
		transactorOptions,
		arg_token,
		arg_depositId,
	)
	if err != nil {
		return transaction, p.errorResolver.ResolveError(
			err,
			p.transactorOptions.From,
			nil,
			"requestTbtcMigration",
			arg_token,
			arg_depositId,
		)
	}

	pLogger.Infof(
		"submitted transaction requestTbtcMigration with id: [%s] and nonce [%v]",
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

			transaction, err := p.contract.RequestTbtcMigration(
				newTransactorOptions,
				arg_token,
				arg_depositId,
			)
			if err != nil {
				return nil, p.errorResolver.ResolveError(
					err,
					p.transactorOptions.From,
					nil,
					"requestTbtcMigration",
					arg_token,
					arg_depositId,
				)
			}

			pLogger.Infof(
				"submitted transaction requestTbtcMigration with id: [%s] and nonce [%v]",
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
func (p *Portal) CallRequestTbtcMigration(
	arg_token common.Address,
	arg_depositId *big.Int,
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
		"requestTbtcMigration",
		&result,
		arg_token,
		arg_depositId,
	)

	return err
}

func (p *Portal) RequestTbtcMigrationGasEstimate(
	arg_token common.Address,
	arg_depositId *big.Int,
) (uint64, error) {
	var result uint64

	result, err := chainutil.EstimateGas(
		p.callerOptions.From,
		p.contractAddress,
		"requestTbtcMigration",
		p.contractABI,
		p.transactor,
		arg_token,
		arg_depositId,
	)

	return result, err
}

// Transaction submission.
func (p *Portal) SetAssetAsLiquidityTreasuryManaged(
	arg_asset common.Address,
	arg_isManaged bool,

	transactionOptions ...chainutil.TransactionOptions,
) (*types.Transaction, error) {
	pLogger.Debug(
		"submitting transaction setAssetAsLiquidityTreasuryManaged",
		" params: ",
		fmt.Sprint(
			arg_asset,
			arg_isManaged,
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

	transaction, err := p.contract.SetAssetAsLiquidityTreasuryManaged(
		transactorOptions,
		arg_asset,
		arg_isManaged,
	)
	if err != nil {
		return transaction, p.errorResolver.ResolveError(
			err,
			p.transactorOptions.From,
			nil,
			"setAssetAsLiquidityTreasuryManaged",
			arg_asset,
			arg_isManaged,
		)
	}

	pLogger.Infof(
		"submitted transaction setAssetAsLiquidityTreasuryManaged with id: [%s] and nonce [%v]",
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

			transaction, err := p.contract.SetAssetAsLiquidityTreasuryManaged(
				newTransactorOptions,
				arg_asset,
				arg_isManaged,
			)
			if err != nil {
				return nil, p.errorResolver.ResolveError(
					err,
					p.transactorOptions.From,
					nil,
					"setAssetAsLiquidityTreasuryManaged",
					arg_asset,
					arg_isManaged,
				)
			}

			pLogger.Infof(
				"submitted transaction setAssetAsLiquidityTreasuryManaged with id: [%s] and nonce [%v]",
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
func (p *Portal) CallSetAssetAsLiquidityTreasuryManaged(
	arg_asset common.Address,
	arg_isManaged bool,
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
		"setAssetAsLiquidityTreasuryManaged",
		&result,
		arg_asset,
		arg_isManaged,
	)

	return err
}

func (p *Portal) SetAssetAsLiquidityTreasuryManagedGasEstimate(
	arg_asset common.Address,
	arg_isManaged bool,
) (uint64, error) {
	var result uint64

	result, err := chainutil.EstimateGas(
		p.callerOptions.From,
		p.contractAddress,
		"setAssetAsLiquidityTreasuryManaged",
		p.contractABI,
		p.transactor,
		arg_asset,
		arg_isManaged,
	)

	return result, err
}

// Transaction submission.
func (p *Portal) SetAssetTbtcMigrationAllowed(
	arg_asset common.Address,
	arg_isAllowed bool,

	transactionOptions ...chainutil.TransactionOptions,
) (*types.Transaction, error) {
	pLogger.Debug(
		"submitting transaction setAssetTbtcMigrationAllowed",
		" params: ",
		fmt.Sprint(
			arg_asset,
			arg_isAllowed,
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

	transaction, err := p.contract.SetAssetTbtcMigrationAllowed(
		transactorOptions,
		arg_asset,
		arg_isAllowed,
	)
	if err != nil {
		return transaction, p.errorResolver.ResolveError(
			err,
			p.transactorOptions.From,
			nil,
			"setAssetTbtcMigrationAllowed",
			arg_asset,
			arg_isAllowed,
		)
	}

	pLogger.Infof(
		"submitted transaction setAssetTbtcMigrationAllowed with id: [%s] and nonce [%v]",
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

			transaction, err := p.contract.SetAssetTbtcMigrationAllowed(
				newTransactorOptions,
				arg_asset,
				arg_isAllowed,
			)
			if err != nil {
				return nil, p.errorResolver.ResolveError(
					err,
					p.transactorOptions.From,
					nil,
					"setAssetTbtcMigrationAllowed",
					arg_asset,
					arg_isAllowed,
				)
			}

			pLogger.Infof(
				"submitted transaction setAssetTbtcMigrationAllowed with id: [%s] and nonce [%v]",
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
func (p *Portal) CallSetAssetTbtcMigrationAllowed(
	arg_asset common.Address,
	arg_isAllowed bool,
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
		"setAssetTbtcMigrationAllowed",
		&result,
		arg_asset,
		arg_isAllowed,
	)

	return err
}

func (p *Portal) SetAssetTbtcMigrationAllowedGasEstimate(
	arg_asset common.Address,
	arg_isAllowed bool,
) (uint64, error) {
	var result uint64

	result, err := chainutil.EstimateGas(
		p.callerOptions.From,
		p.contractAddress,
		"setAssetTbtcMigrationAllowed",
		p.contractABI,
		p.transactor,
		arg_asset,
		arg_isAllowed,
	)

	return result, err
}

// Transaction submission.
func (p *Portal) SetLiquidityTreasury(
	arg__liquidityTreasury common.Address,

	transactionOptions ...chainutil.TransactionOptions,
) (*types.Transaction, error) {
	pLogger.Debug(
		"submitting transaction setLiquidityTreasury",
		" params: ",
		fmt.Sprint(
			arg__liquidityTreasury,
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

	transaction, err := p.contract.SetLiquidityTreasury(
		transactorOptions,
		arg__liquidityTreasury,
	)
	if err != nil {
		return transaction, p.errorResolver.ResolveError(
			err,
			p.transactorOptions.From,
			nil,
			"setLiquidityTreasury",
			arg__liquidityTreasury,
		)
	}

	pLogger.Infof(
		"submitted transaction setLiquidityTreasury with id: [%s] and nonce [%v]",
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

			transaction, err := p.contract.SetLiquidityTreasury(
				newTransactorOptions,
				arg__liquidityTreasury,
			)
			if err != nil {
				return nil, p.errorResolver.ResolveError(
					err,
					p.transactorOptions.From,
					nil,
					"setLiquidityTreasury",
					arg__liquidityTreasury,
				)
			}

			pLogger.Infof(
				"submitted transaction setLiquidityTreasury with id: [%s] and nonce [%v]",
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
func (p *Portal) CallSetLiquidityTreasury(
	arg__liquidityTreasury common.Address,
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
		"setLiquidityTreasury",
		&result,
		arg__liquidityTreasury,
	)

	return err
}

func (p *Portal) SetLiquidityTreasuryGasEstimate(
	arg__liquidityTreasury common.Address,
) (uint64, error) {
	var result uint64

	result, err := chainutil.EstimateGas(
		p.callerOptions.From,
		p.contractAddress,
		"setLiquidityTreasury",
		p.contractABI,
		p.transactor,
		arg__liquidityTreasury,
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
func (p *Portal) SetReceiptParams(
	arg_token common.Address,
	arg_annualFee uint8,
	arg_mintCap uint8,
	arg_receiptToken common.Address,

	transactionOptions ...chainutil.TransactionOptions,
) (*types.Transaction, error) {
	pLogger.Debug(
		"submitting transaction setReceiptParams",
		" params: ",
		fmt.Sprint(
			arg_token,
			arg_annualFee,
			arg_mintCap,
			arg_receiptToken,
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

	transaction, err := p.contract.SetReceiptParams(
		transactorOptions,
		arg_token,
		arg_annualFee,
		arg_mintCap,
		arg_receiptToken,
	)
	if err != nil {
		return transaction, p.errorResolver.ResolveError(
			err,
			p.transactorOptions.From,
			nil,
			"setReceiptParams",
			arg_token,
			arg_annualFee,
			arg_mintCap,
			arg_receiptToken,
		)
	}

	pLogger.Infof(
		"submitted transaction setReceiptParams with id: [%s] and nonce [%v]",
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

			transaction, err := p.contract.SetReceiptParams(
				newTransactorOptions,
				arg_token,
				arg_annualFee,
				arg_mintCap,
				arg_receiptToken,
			)
			if err != nil {
				return nil, p.errorResolver.ResolveError(
					err,
					p.transactorOptions.From,
					nil,
					"setReceiptParams",
					arg_token,
					arg_annualFee,
					arg_mintCap,
					arg_receiptToken,
				)
			}

			pLogger.Infof(
				"submitted transaction setReceiptParams with id: [%s] and nonce [%v]",
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
func (p *Portal) CallSetReceiptParams(
	arg_token common.Address,
	arg_annualFee uint8,
	arg_mintCap uint8,
	arg_receiptToken common.Address,
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
		"setReceiptParams",
		&result,
		arg_token,
		arg_annualFee,
		arg_mintCap,
		arg_receiptToken,
	)

	return err
}

func (p *Portal) SetReceiptParamsGasEstimate(
	arg_token common.Address,
	arg_annualFee uint8,
	arg_mintCap uint8,
	arg_receiptToken common.Address,
) (uint64, error) {
	var result uint64

	result, err := chainutil.EstimateGas(
		p.callerOptions.From,
		p.contractAddress,
		"setReceiptParams",
		p.contractABI,
		p.transactor,
		arg_token,
		arg_annualFee,
		arg_mintCap,
		arg_receiptToken,
	)

	return result, err
}

// Transaction submission.
func (p *Portal) SetTbtcMigrationTreasury(
	arg__tbtcMigrationTreasury common.Address,

	transactionOptions ...chainutil.TransactionOptions,
) (*types.Transaction, error) {
	pLogger.Debug(
		"submitting transaction setTbtcMigrationTreasury",
		" params: ",
		fmt.Sprint(
			arg__tbtcMigrationTreasury,
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

	transaction, err := p.contract.SetTbtcMigrationTreasury(
		transactorOptions,
		arg__tbtcMigrationTreasury,
	)
	if err != nil {
		return transaction, p.errorResolver.ResolveError(
			err,
			p.transactorOptions.From,
			nil,
			"setTbtcMigrationTreasury",
			arg__tbtcMigrationTreasury,
		)
	}

	pLogger.Infof(
		"submitted transaction setTbtcMigrationTreasury with id: [%s] and nonce [%v]",
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

			transaction, err := p.contract.SetTbtcMigrationTreasury(
				newTransactorOptions,
				arg__tbtcMigrationTreasury,
			)
			if err != nil {
				return nil, p.errorResolver.ResolveError(
					err,
					p.transactorOptions.From,
					nil,
					"setTbtcMigrationTreasury",
					arg__tbtcMigrationTreasury,
				)
			}

			pLogger.Infof(
				"submitted transaction setTbtcMigrationTreasury with id: [%s] and nonce [%v]",
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
func (p *Portal) CallSetTbtcMigrationTreasury(
	arg__tbtcMigrationTreasury common.Address,
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
		"setTbtcMigrationTreasury",
		&result,
		arg__tbtcMigrationTreasury,
	)

	return err
}

func (p *Portal) SetTbtcMigrationTreasuryGasEstimate(
	arg__tbtcMigrationTreasury common.Address,
) (uint64, error) {
	var result uint64

	result, err := chainutil.EstimateGas(
		p.callerOptions.From,
		p.contractAddress,
		"setTbtcMigrationTreasury",
		p.contractABI,
		p.transactor,
		arg__tbtcMigrationTreasury,
	)

	return result, err
}

// Transaction submission.
func (p *Portal) SetTbtcTokenAddress(
	arg__tbtcToken common.Address,

	transactionOptions ...chainutil.TransactionOptions,
) (*types.Transaction, error) {
	pLogger.Debug(
		"submitting transaction setTbtcTokenAddress",
		" params: ",
		fmt.Sprint(
			arg__tbtcToken,
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

	transaction, err := p.contract.SetTbtcTokenAddress(
		transactorOptions,
		arg__tbtcToken,
	)
	if err != nil {
		return transaction, p.errorResolver.ResolveError(
			err,
			p.transactorOptions.From,
			nil,
			"setTbtcTokenAddress",
			arg__tbtcToken,
		)
	}

	pLogger.Infof(
		"submitted transaction setTbtcTokenAddress with id: [%s] and nonce [%v]",
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

			transaction, err := p.contract.SetTbtcTokenAddress(
				newTransactorOptions,
				arg__tbtcToken,
			)
			if err != nil {
				return nil, p.errorResolver.ResolveError(
					err,
					p.transactorOptions.From,
					nil,
					"setTbtcTokenAddress",
					arg__tbtcToken,
				)
			}

			pLogger.Infof(
				"submitted transaction setTbtcTokenAddress with id: [%s] and nonce [%v]",
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
func (p *Portal) CallSetTbtcTokenAddress(
	arg__tbtcToken common.Address,
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
		"setTbtcTokenAddress",
		&result,
		arg__tbtcToken,
	)

	return err
}

func (p *Portal) SetTbtcTokenAddressGasEstimate(
	arg__tbtcToken common.Address,
) (uint64, error) {
	var result uint64

	result, err := chainutil.EstimateGas(
		p.callerOptions.From,
		p.contractAddress,
		"setTbtcTokenAddress",
		p.contractABI,
		p.transactor,
		arg__tbtcToken,
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

	transactionOptions ...chainutil.TransactionOptions,
) (*types.Transaction, error) {
	pLogger.Debug(
		"submitting transaction withdraw",
		" params: ",
		fmt.Sprint(
			arg_token,
			arg_depositId,
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
	)
	if err != nil {
		return transaction, p.errorResolver.ResolveError(
			err,
			p.transactorOptions.From,
			nil,
			"withdraw",
			arg_token,
			arg_depositId,
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
			)
			if err != nil {
				return nil, p.errorResolver.ResolveError(
					err,
					p.transactorOptions.From,
					nil,
					"withdraw",
					arg_token,
					arg_depositId,
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
	)

	return err
}

func (p *Portal) WithdrawGasEstimate(
	arg_token common.Address,
	arg_depositId *big.Int,
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
	)

	return result, err
}

// Transaction submission.
func (p *Portal) WithdrawAsLiquidityTreasury(
	arg_token common.Address,
	arg_amount *big.Int,

	transactionOptions ...chainutil.TransactionOptions,
) (*types.Transaction, error) {
	pLogger.Debug(
		"submitting transaction withdrawAsLiquidityTreasury",
		" params: ",
		fmt.Sprint(
			arg_token,
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

	transaction, err := p.contract.WithdrawAsLiquidityTreasury(
		transactorOptions,
		arg_token,
		arg_amount,
	)
	if err != nil {
		return transaction, p.errorResolver.ResolveError(
			err,
			p.transactorOptions.From,
			nil,
			"withdrawAsLiquidityTreasury",
			arg_token,
			arg_amount,
		)
	}

	pLogger.Infof(
		"submitted transaction withdrawAsLiquidityTreasury with id: [%s] and nonce [%v]",
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

			transaction, err := p.contract.WithdrawAsLiquidityTreasury(
				newTransactorOptions,
				arg_token,
				arg_amount,
			)
			if err != nil {
				return nil, p.errorResolver.ResolveError(
					err,
					p.transactorOptions.From,
					nil,
					"withdrawAsLiquidityTreasury",
					arg_token,
					arg_amount,
				)
			}

			pLogger.Infof(
				"submitted transaction withdrawAsLiquidityTreasury with id: [%s] and nonce [%v]",
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
func (p *Portal) CallWithdrawAsLiquidityTreasury(
	arg_token common.Address,
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
		"withdrawAsLiquidityTreasury",
		&result,
		arg_token,
		arg_amount,
	)

	return err
}

func (p *Portal) WithdrawAsLiquidityTreasuryGasEstimate(
	arg_token common.Address,
	arg_amount *big.Int,
) (uint64, error) {
	var result uint64

	result, err := chainutil.EstimateGas(
		p.callerOptions.From,
		p.contractAddress,
		"withdrawAsLiquidityTreasury",
		p.contractABI,
		p.transactor,
		arg_token,
		arg_amount,
	)

	return result, err
}

// Transaction submission.
func (p *Portal) WithdrawForTbtcMigration(
	arg_token common.Address,
	arg_depositsToMigrate []abi.PortalDepositToMigrate,

	transactionOptions ...chainutil.TransactionOptions,
) (*types.Transaction, error) {
	pLogger.Debug(
		"submitting transaction withdrawForTbtcMigration",
		" params: ",
		fmt.Sprint(
			arg_token,
			arg_depositsToMigrate,
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

	transaction, err := p.contract.WithdrawForTbtcMigration(
		transactorOptions,
		arg_token,
		arg_depositsToMigrate,
	)
	if err != nil {
		return transaction, p.errorResolver.ResolveError(
			err,
			p.transactorOptions.From,
			nil,
			"withdrawForTbtcMigration",
			arg_token,
			arg_depositsToMigrate,
		)
	}

	pLogger.Infof(
		"submitted transaction withdrawForTbtcMigration with id: [%s] and nonce [%v]",
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

			transaction, err := p.contract.WithdrawForTbtcMigration(
				newTransactorOptions,
				arg_token,
				arg_depositsToMigrate,
			)
			if err != nil {
				return nil, p.errorResolver.ResolveError(
					err,
					p.transactorOptions.From,
					nil,
					"withdrawForTbtcMigration",
					arg_token,
					arg_depositsToMigrate,
				)
			}

			pLogger.Infof(
				"submitted transaction withdrawForTbtcMigration with id: [%s] and nonce [%v]",
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
func (p *Portal) CallWithdrawForTbtcMigration(
	arg_token common.Address,
	arg_depositsToMigrate []abi.PortalDepositToMigrate,
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
		"withdrawForTbtcMigration",
		&result,
		arg_token,
		arg_depositsToMigrate,
	)

	return err
}

func (p *Portal) WithdrawForTbtcMigrationGasEstimate(
	arg_token common.Address,
	arg_depositsToMigrate []abi.PortalDepositToMigrate,
) (uint64, error) {
	var result uint64

	result, err := chainutil.EstimateGas(
		p.callerOptions.From,
		p.contractAddress,
		"withdrawForTbtcMigration",
		p.contractABI,
		p.transactor,
		arg_token,
		arg_depositsToMigrate,
	)

	return result, err
}

// Transaction submission.
func (p *Portal) WithdrawPartially(
	arg_token common.Address,
	arg_depositId *big.Int,
	arg_amount *big.Int,

	transactionOptions ...chainutil.TransactionOptions,
) (*types.Transaction, error) {
	pLogger.Debug(
		"submitting transaction withdrawPartially",
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

	transaction, err := p.contract.WithdrawPartially(
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
			"withdrawPartially",
			arg_token,
			arg_depositId,
			arg_amount,
		)
	}

	pLogger.Infof(
		"submitted transaction withdrawPartially with id: [%s] and nonce [%v]",
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

			transaction, err := p.contract.WithdrawPartially(
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
					"withdrawPartially",
					arg_token,
					arg_depositId,
					arg_amount,
				)
			}

			pLogger.Infof(
				"submitted transaction withdrawPartially with id: [%s] and nonce [%v]",
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
func (p *Portal) CallWithdrawPartially(
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
		"withdrawPartially",
		&result,
		arg_token,
		arg_depositId,
		arg_amount,
	)

	return err
}

func (p *Portal) WithdrawPartiallyGasEstimate(
	arg_token common.Address,
	arg_depositId *big.Int,
	arg_amount *big.Int,
) (uint64, error) {
	var result uint64

	result, err := chainutil.EstimateGas(
		p.callerOptions.From,
		p.contractAddress,
		"withdrawPartially",
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
	Balance            *big.Int
	UnlockAt           uint32
	ReceiptMinted      *big.Int
	FeeOwed            *big.Int
	LastFeeIntegral    *big.Int
	TbtcMigrationState uint8
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

type feeInfo struct {
	TotalMinted     *big.Int
	LastFeeUpdateAt uint32
	FeeIntegral     *big.Int
	AnnualFee       uint8
	MintCap         uint8
	ReceiptToken    common.Address
	FeeCollected    *big.Int
}

func (p *Portal) FeeInfo(
	arg0 common.Address,
) (feeInfo, error) {
	result, err := p.contract.FeeInfo(
		p.callerOptions,
		arg0,
	)

	if err != nil {
		return result, p.errorResolver.ResolveError(
			err,
			p.callerOptions.From,
			nil,
			"feeInfo",
			arg0,
		)
	}

	return result, err
}

func (p *Portal) FeeInfoAtBlock(
	arg0 common.Address,
	blockNumber *big.Int,
) (feeInfo, error) {
	var result feeInfo

	err := chainutil.CallAtBlock(
		p.callerOptions.From,
		blockNumber,
		nil,
		p.contractABI,
		p.caller,
		p.errorResolver,
		p.contractAddress,
		"feeInfo",
		&result,
		arg0,
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

func (p *Portal) LiquidityTreasury() (common.Address, error) {
	result, err := p.contract.LiquidityTreasury(
		p.callerOptions,
	)

	if err != nil {
		return result, p.errorResolver.ResolveError(
			err,
			p.callerOptions.From,
			nil,
			"liquidityTreasury",
		)
	}

	return result, err
}

func (p *Portal) LiquidityTreasuryAtBlock(
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
		"liquidityTreasury",
		&result,
	)

	return result, err
}

func (p *Portal) LiquidityTreasuryManaged(
	arg0 common.Address,
) (bool, error) {
	result, err := p.contract.LiquidityTreasuryManaged(
		p.callerOptions,
		arg0,
	)

	if err != nil {
		return result, p.errorResolver.ResolveError(
			err,
			p.callerOptions.From,
			nil,
			"liquidityTreasuryManaged",
			arg0,
		)
	}

	return result, err
}

func (p *Portal) LiquidityTreasuryManagedAtBlock(
	arg0 common.Address,
	blockNumber *big.Int,
) (bool, error) {
	var result bool

	err := chainutil.CallAtBlock(
		p.callerOptions.From,
		blockNumber,
		nil,
		p.contractABI,
		p.caller,
		p.errorResolver,
		p.contractAddress,
		"liquidityTreasuryManaged",
		&result,
		arg0,
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

func (p *Portal) TbtcMigrationTreasury() (common.Address, error) {
	result, err := p.contract.TbtcMigrationTreasury(
		p.callerOptions,
	)

	if err != nil {
		return result, p.errorResolver.ResolveError(
			err,
			p.callerOptions.From,
			nil,
			"tbtcMigrationTreasury",
		)
	}

	return result, err
}

func (p *Portal) TbtcMigrationTreasuryAtBlock(
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
		"tbtcMigrationTreasury",
		&result,
	)

	return result, err
}

type tbtcMigrations struct {
	IsAllowed      bool
	TotalMigrating *big.Int
}

func (p *Portal) TbtcMigrations(
	arg0 common.Address,
) (tbtcMigrations, error) {
	result, err := p.contract.TbtcMigrations(
		p.callerOptions,
		arg0,
	)

	if err != nil {
		return result, p.errorResolver.ResolveError(
			err,
			p.callerOptions.From,
			nil,
			"tbtcMigrations",
			arg0,
		)
	}

	return result, err
}

func (p *Portal) TbtcMigrationsAtBlock(
	arg0 common.Address,
	blockNumber *big.Int,
) (tbtcMigrations, error) {
	var result tbtcMigrations

	err := chainutil.CallAtBlock(
		p.callerOptions.From,
		blockNumber,
		nil,
		p.contractABI,
		p.caller,
		p.errorResolver,
		p.contractAddress,
		"tbtcMigrations",
		&result,
		arg0,
	)

	return result, err
}

func (p *Portal) TbtcToken() (common.Address, error) {
	result, err := p.contract.TbtcToken(
		p.callerOptions,
	)

	if err != nil {
		return result, p.errorResolver.ResolveError(
			err,
			p.callerOptions.From,
			nil,
			"tbtcToken",
		)
	}

	return result, err
}

func (p *Portal) TbtcTokenAtBlock(
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
		"tbtcToken",
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

func (p *Portal) FeeCollectedEvent(
	opts *ethereum.SubscribeOpts,
	depositorFilter []common.Address,
	tokenFilter []common.Address,
	depositIdFilter []*big.Int,
) *PFeeCollectedSubscription {
	if opts == nil {
		opts = new(ethereum.SubscribeOpts)
	}
	if opts.Tick == 0 {
		opts.Tick = chainutil.DefaultSubscribeOptsTick
	}
	if opts.PastBlocks == 0 {
		opts.PastBlocks = chainutil.DefaultSubscribeOptsPastBlocks
	}

	return &PFeeCollectedSubscription{
		p,
		opts,
		depositorFilter,
		tokenFilter,
		depositIdFilter,
	}
}

type PFeeCollectedSubscription struct {
	contract        *Portal
	opts            *ethereum.SubscribeOpts
	depositorFilter []common.Address
	tokenFilter     []common.Address
	depositIdFilter []*big.Int
}

type portalFeeCollectedFunc func(
	Depositor common.Address,
	Token common.Address,
	DepositId *big.Int,
	Fee *big.Int,
	blockNumber uint64,
)

func (fcs *PFeeCollectedSubscription) OnEvent(
	handler portalFeeCollectedFunc,
) subscription.EventSubscription {
	eventChan := make(chan *abi.PortalFeeCollected)
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
					event.Fee,
					event.Raw.BlockNumber,
				)
			}
		}
	}()

	sub := fcs.Pipe(eventChan)
	return subscription.NewEventSubscription(func() {
		sub.Unsubscribe()
		cancelCtx()
	})
}

func (fcs *PFeeCollectedSubscription) Pipe(
	sink chan *abi.PortalFeeCollected,
) subscription.EventSubscription {
	ctx, cancelCtx := context.WithCancel(context.Background())
	go func() {
		ticker := time.NewTicker(fcs.opts.Tick)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				lastBlock, err := fcs.contract.blockCounter.CurrentBlock()
				if err != nil {
					pLogger.Errorf(
						"subscription failed to pull events: [%v]",
						err,
					)
				}
				fromBlock := lastBlock - fcs.opts.PastBlocks

				pLogger.Infof(
					"subscription monitoring fetching past FeeCollected events "+
						"starting from block [%v]",
					fromBlock,
				)
				events, err := fcs.contract.PastFeeCollectedEvents(
					fromBlock,
					nil,
					fcs.depositorFilter,
					fcs.tokenFilter,
					fcs.depositIdFilter,
				)
				if err != nil {
					pLogger.Errorf(
						"subscription failed to pull events: [%v]",
						err,
					)
					continue
				}
				pLogger.Infof(
					"subscription monitoring fetched [%v] past FeeCollected events",
					len(events),
				)

				for _, event := range events {
					sink <- event
				}
			}
		}
	}()

	sub := fcs.contract.watchFeeCollected(
		sink,
		fcs.depositorFilter,
		fcs.tokenFilter,
		fcs.depositIdFilter,
	)

	return subscription.NewEventSubscription(func() {
		sub.Unsubscribe()
		cancelCtx()
	})
}

func (p *Portal) watchFeeCollected(
	sink chan *abi.PortalFeeCollected,
	depositorFilter []common.Address,
	tokenFilter []common.Address,
	depositIdFilter []*big.Int,
) event.Subscription {
	subscribeFn := func(ctx context.Context) (event.Subscription, error) {
		return p.contract.WatchFeeCollected(
			&bind.WatchOpts{Context: ctx},
			sink,
			depositorFilter,
			tokenFilter,
			depositIdFilter,
		)
	}

	thresholdViolatedFn := func(elapsed time.Duration) {
		pLogger.Warnf(
			"subscription to event FeeCollected had to be "+
				"retried [%s] since the last attempt; please inspect "+
				"host chain connectivity",
			elapsed,
		)
	}

	subscriptionFailedFn := func(err error) {
		pLogger.Errorf(
			"subscription to event FeeCollected failed "+
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

func (p *Portal) PastFeeCollectedEvents(
	startBlock uint64,
	endBlock *uint64,
	depositorFilter []common.Address,
	tokenFilter []common.Address,
	depositIdFilter []*big.Int,
) ([]*abi.PortalFeeCollected, error) {
	iterator, err := p.contract.FilterFeeCollected(
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
			"error retrieving past FeeCollected events: [%v]",
			err,
		)
	}

	events := make([]*abi.PortalFeeCollected, 0)

	for iterator.Next() {
		event := iterator.Event
		events = append(events, event)
	}

	return events, nil
}

func (p *Portal) FeeCollectedTbtcMigratedEvent(
	opts *ethereum.SubscribeOpts,
	depositorFilter []common.Address,
	tokenFilter []common.Address,
	depositIdFilter []*big.Int,
) *PFeeCollectedTbtcMigratedSubscription {
	if opts == nil {
		opts = new(ethereum.SubscribeOpts)
	}
	if opts.Tick == 0 {
		opts.Tick = chainutil.DefaultSubscribeOptsTick
	}
	if opts.PastBlocks == 0 {
		opts.PastBlocks = chainutil.DefaultSubscribeOptsPastBlocks
	}

	return &PFeeCollectedTbtcMigratedSubscription{
		p,
		opts,
		depositorFilter,
		tokenFilter,
		depositIdFilter,
	}
}

type PFeeCollectedTbtcMigratedSubscription struct {
	contract        *Portal
	opts            *ethereum.SubscribeOpts
	depositorFilter []common.Address
	tokenFilter     []common.Address
	depositIdFilter []*big.Int
}

type portalFeeCollectedTbtcMigratedFunc func(
	Depositor common.Address,
	Token common.Address,
	TbtcToken common.Address,
	DepositId *big.Int,
	FeeInTbtc *big.Int,
	blockNumber uint64,
)

func (fctms *PFeeCollectedTbtcMigratedSubscription) OnEvent(
	handler portalFeeCollectedTbtcMigratedFunc,
) subscription.EventSubscription {
	eventChan := make(chan *abi.PortalFeeCollectedTbtcMigrated)
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
					event.TbtcToken,
					event.DepositId,
					event.FeeInTbtc,
					event.Raw.BlockNumber,
				)
			}
		}
	}()

	sub := fctms.Pipe(eventChan)
	return subscription.NewEventSubscription(func() {
		sub.Unsubscribe()
		cancelCtx()
	})
}

func (fctms *PFeeCollectedTbtcMigratedSubscription) Pipe(
	sink chan *abi.PortalFeeCollectedTbtcMigrated,
) subscription.EventSubscription {
	ctx, cancelCtx := context.WithCancel(context.Background())
	go func() {
		ticker := time.NewTicker(fctms.opts.Tick)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				lastBlock, err := fctms.contract.blockCounter.CurrentBlock()
				if err != nil {
					pLogger.Errorf(
						"subscription failed to pull events: [%v]",
						err,
					)
				}
				fromBlock := lastBlock - fctms.opts.PastBlocks

				pLogger.Infof(
					"subscription monitoring fetching past FeeCollectedTbtcMigrated events "+
						"starting from block [%v]",
					fromBlock,
				)
				events, err := fctms.contract.PastFeeCollectedTbtcMigratedEvents(
					fromBlock,
					nil,
					fctms.depositorFilter,
					fctms.tokenFilter,
					fctms.depositIdFilter,
				)
				if err != nil {
					pLogger.Errorf(
						"subscription failed to pull events: [%v]",
						err,
					)
					continue
				}
				pLogger.Infof(
					"subscription monitoring fetched [%v] past FeeCollectedTbtcMigrated events",
					len(events),
				)

				for _, event := range events {
					sink <- event
				}
			}
		}
	}()

	sub := fctms.contract.watchFeeCollectedTbtcMigrated(
		sink,
		fctms.depositorFilter,
		fctms.tokenFilter,
		fctms.depositIdFilter,
	)

	return subscription.NewEventSubscription(func() {
		sub.Unsubscribe()
		cancelCtx()
	})
}

func (p *Portal) watchFeeCollectedTbtcMigrated(
	sink chan *abi.PortalFeeCollectedTbtcMigrated,
	depositorFilter []common.Address,
	tokenFilter []common.Address,
	depositIdFilter []*big.Int,
) event.Subscription {
	subscribeFn := func(ctx context.Context) (event.Subscription, error) {
		return p.contract.WatchFeeCollectedTbtcMigrated(
			&bind.WatchOpts{Context: ctx},
			sink,
			depositorFilter,
			tokenFilter,
			depositIdFilter,
		)
	}

	thresholdViolatedFn := func(elapsed time.Duration) {
		pLogger.Warnf(
			"subscription to event FeeCollectedTbtcMigrated had to be "+
				"retried [%s] since the last attempt; please inspect "+
				"host chain connectivity",
			elapsed,
		)
	}

	subscriptionFailedFn := func(err error) {
		pLogger.Errorf(
			"subscription to event FeeCollectedTbtcMigrated failed "+
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

func (p *Portal) PastFeeCollectedTbtcMigratedEvents(
	startBlock uint64,
	endBlock *uint64,
	depositorFilter []common.Address,
	tokenFilter []common.Address,
	depositIdFilter []*big.Int,
) ([]*abi.PortalFeeCollectedTbtcMigrated, error) {
	iterator, err := p.contract.FilterFeeCollectedTbtcMigrated(
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
			"error retrieving past FeeCollectedTbtcMigrated events: [%v]",
			err,
		)
	}

	events := make([]*abi.PortalFeeCollectedTbtcMigrated, 0)

	for iterator.Next() {
		event := iterator.Event
		events = append(events, event)
	}

	return events, nil
}

func (p *Portal) FundedFromTbtcMigrationEvent(
	opts *ethereum.SubscribeOpts,
) *PFundedFromTbtcMigrationSubscription {
	if opts == nil {
		opts = new(ethereum.SubscribeOpts)
	}
	if opts.Tick == 0 {
		opts.Tick = chainutil.DefaultSubscribeOptsTick
	}
	if opts.PastBlocks == 0 {
		opts.PastBlocks = chainutil.DefaultSubscribeOptsPastBlocks
	}

	return &PFundedFromTbtcMigrationSubscription{
		p,
		opts,
	}
}

type PFundedFromTbtcMigrationSubscription struct {
	contract *Portal
	opts     *ethereum.SubscribeOpts
}

type portalFundedFromTbtcMigrationFunc func(
	Amount *big.Int,
	blockNumber uint64,
)

func (fftms *PFundedFromTbtcMigrationSubscription) OnEvent(
	handler portalFundedFromTbtcMigrationFunc,
) subscription.EventSubscription {
	eventChan := make(chan *abi.PortalFundedFromTbtcMigration)
	ctx, cancelCtx := context.WithCancel(context.Background())

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case event := <-eventChan:
				handler(
					event.Amount,
					event.Raw.BlockNumber,
				)
			}
		}
	}()

	sub := fftms.Pipe(eventChan)
	return subscription.NewEventSubscription(func() {
		sub.Unsubscribe()
		cancelCtx()
	})
}

func (fftms *PFundedFromTbtcMigrationSubscription) Pipe(
	sink chan *abi.PortalFundedFromTbtcMigration,
) subscription.EventSubscription {
	ctx, cancelCtx := context.WithCancel(context.Background())
	go func() {
		ticker := time.NewTicker(fftms.opts.Tick)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				lastBlock, err := fftms.contract.blockCounter.CurrentBlock()
				if err != nil {
					pLogger.Errorf(
						"subscription failed to pull events: [%v]",
						err,
					)
				}
				fromBlock := lastBlock - fftms.opts.PastBlocks

				pLogger.Infof(
					"subscription monitoring fetching past FundedFromTbtcMigration events "+
						"starting from block [%v]",
					fromBlock,
				)
				events, err := fftms.contract.PastFundedFromTbtcMigrationEvents(
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
					"subscription monitoring fetched [%v] past FundedFromTbtcMigration events",
					len(events),
				)

				for _, event := range events {
					sink <- event
				}
			}
		}
	}()

	sub := fftms.contract.watchFundedFromTbtcMigration(
		sink,
	)

	return subscription.NewEventSubscription(func() {
		sub.Unsubscribe()
		cancelCtx()
	})
}

func (p *Portal) watchFundedFromTbtcMigration(
	sink chan *abi.PortalFundedFromTbtcMigration,
) event.Subscription {
	subscribeFn := func(ctx context.Context) (event.Subscription, error) {
		return p.contract.WatchFundedFromTbtcMigration(
			&bind.WatchOpts{Context: ctx},
			sink,
		)
	}

	thresholdViolatedFn := func(elapsed time.Duration) {
		pLogger.Warnf(
			"subscription to event FundedFromTbtcMigration had to be "+
				"retried [%s] since the last attempt; please inspect "+
				"host chain connectivity",
			elapsed,
		)
	}

	subscriptionFailedFn := func(err error) {
		pLogger.Errorf(
			"subscription to event FundedFromTbtcMigration failed "+
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

func (p *Portal) PastFundedFromTbtcMigrationEvents(
	startBlock uint64,
	endBlock *uint64,
) ([]*abi.PortalFundedFromTbtcMigration, error) {
	iterator, err := p.contract.FilterFundedFromTbtcMigration(
		&bind.FilterOpts{
			Start: startBlock,
			End:   endBlock,
		},
	)
	if err != nil {
		return nil, fmt.Errorf(
			"error retrieving past FundedFromTbtcMigration events: [%v]",
			err,
		)
	}

	events := make([]*abi.PortalFundedFromTbtcMigration, 0)

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

func (p *Portal) LiquidityTreasuryManagedAssetUpdatedEvent(
	opts *ethereum.SubscribeOpts,
	assetFilter []common.Address,
) *PLiquidityTreasuryManagedAssetUpdatedSubscription {
	if opts == nil {
		opts = new(ethereum.SubscribeOpts)
	}
	if opts.Tick == 0 {
		opts.Tick = chainutil.DefaultSubscribeOptsTick
	}
	if opts.PastBlocks == 0 {
		opts.PastBlocks = chainutil.DefaultSubscribeOptsPastBlocks
	}

	return &PLiquidityTreasuryManagedAssetUpdatedSubscription{
		p,
		opts,
		assetFilter,
	}
}

type PLiquidityTreasuryManagedAssetUpdatedSubscription struct {
	contract    *Portal
	opts        *ethereum.SubscribeOpts
	assetFilter []common.Address
}

type portalLiquidityTreasuryManagedAssetUpdatedFunc func(
	Asset common.Address,
	IsManaged bool,
	blockNumber uint64,
)

func (ltmaus *PLiquidityTreasuryManagedAssetUpdatedSubscription) OnEvent(
	handler portalLiquidityTreasuryManagedAssetUpdatedFunc,
) subscription.EventSubscription {
	eventChan := make(chan *abi.PortalLiquidityTreasuryManagedAssetUpdated)
	ctx, cancelCtx := context.WithCancel(context.Background())

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case event := <-eventChan:
				handler(
					event.Asset,
					event.IsManaged,
					event.Raw.BlockNumber,
				)
			}
		}
	}()

	sub := ltmaus.Pipe(eventChan)
	return subscription.NewEventSubscription(func() {
		sub.Unsubscribe()
		cancelCtx()
	})
}

func (ltmaus *PLiquidityTreasuryManagedAssetUpdatedSubscription) Pipe(
	sink chan *abi.PortalLiquidityTreasuryManagedAssetUpdated,
) subscription.EventSubscription {
	ctx, cancelCtx := context.WithCancel(context.Background())
	go func() {
		ticker := time.NewTicker(ltmaus.opts.Tick)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				lastBlock, err := ltmaus.contract.blockCounter.CurrentBlock()
				if err != nil {
					pLogger.Errorf(
						"subscription failed to pull events: [%v]",
						err,
					)
				}
				fromBlock := lastBlock - ltmaus.opts.PastBlocks

				pLogger.Infof(
					"subscription monitoring fetching past LiquidityTreasuryManagedAssetUpdated events "+
						"starting from block [%v]",
					fromBlock,
				)
				events, err := ltmaus.contract.PastLiquidityTreasuryManagedAssetUpdatedEvents(
					fromBlock,
					nil,
					ltmaus.assetFilter,
				)
				if err != nil {
					pLogger.Errorf(
						"subscription failed to pull events: [%v]",
						err,
					)
					continue
				}
				pLogger.Infof(
					"subscription monitoring fetched [%v] past LiquidityTreasuryManagedAssetUpdated events",
					len(events),
				)

				for _, event := range events {
					sink <- event
				}
			}
		}
	}()

	sub := ltmaus.contract.watchLiquidityTreasuryManagedAssetUpdated(
		sink,
		ltmaus.assetFilter,
	)

	return subscription.NewEventSubscription(func() {
		sub.Unsubscribe()
		cancelCtx()
	})
}

func (p *Portal) watchLiquidityTreasuryManagedAssetUpdated(
	sink chan *abi.PortalLiquidityTreasuryManagedAssetUpdated,
	assetFilter []common.Address,
) event.Subscription {
	subscribeFn := func(ctx context.Context) (event.Subscription, error) {
		return p.contract.WatchLiquidityTreasuryManagedAssetUpdated(
			&bind.WatchOpts{Context: ctx},
			sink,
			assetFilter,
		)
	}

	thresholdViolatedFn := func(elapsed time.Duration) {
		pLogger.Warnf(
			"subscription to event LiquidityTreasuryManagedAssetUpdated had to be "+
				"retried [%s] since the last attempt; please inspect "+
				"host chain connectivity",
			elapsed,
		)
	}

	subscriptionFailedFn := func(err error) {
		pLogger.Errorf(
			"subscription to event LiquidityTreasuryManagedAssetUpdated failed "+
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

func (p *Portal) PastLiquidityTreasuryManagedAssetUpdatedEvents(
	startBlock uint64,
	endBlock *uint64,
	assetFilter []common.Address,
) ([]*abi.PortalLiquidityTreasuryManagedAssetUpdated, error) {
	iterator, err := p.contract.FilterLiquidityTreasuryManagedAssetUpdated(
		&bind.FilterOpts{
			Start: startBlock,
			End:   endBlock,
		},
		assetFilter,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"error retrieving past LiquidityTreasuryManagedAssetUpdated events: [%v]",
			err,
		)
	}

	events := make([]*abi.PortalLiquidityTreasuryManagedAssetUpdated, 0)

	for iterator.Next() {
		event := iterator.Event
		events = append(events, event)
	}

	return events, nil
}

func (p *Portal) LiquidityTreasuryUpdatedEvent(
	opts *ethereum.SubscribeOpts,
	previousLiquidityTreasuryFilter []common.Address,
	newLiquidityTreasuryFilter []common.Address,
) *PLiquidityTreasuryUpdatedSubscription {
	if opts == nil {
		opts = new(ethereum.SubscribeOpts)
	}
	if opts.Tick == 0 {
		opts.Tick = chainutil.DefaultSubscribeOptsTick
	}
	if opts.PastBlocks == 0 {
		opts.PastBlocks = chainutil.DefaultSubscribeOptsPastBlocks
	}

	return &PLiquidityTreasuryUpdatedSubscription{
		p,
		opts,
		previousLiquidityTreasuryFilter,
		newLiquidityTreasuryFilter,
	}
}

type PLiquidityTreasuryUpdatedSubscription struct {
	contract                        *Portal
	opts                            *ethereum.SubscribeOpts
	previousLiquidityTreasuryFilter []common.Address
	newLiquidityTreasuryFilter      []common.Address
}

type portalLiquidityTreasuryUpdatedFunc func(
	PreviousLiquidityTreasury common.Address,
	NewLiquidityTreasury common.Address,
	blockNumber uint64,
)

func (ltus *PLiquidityTreasuryUpdatedSubscription) OnEvent(
	handler portalLiquidityTreasuryUpdatedFunc,
) subscription.EventSubscription {
	eventChan := make(chan *abi.PortalLiquidityTreasuryUpdated)
	ctx, cancelCtx := context.WithCancel(context.Background())

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case event := <-eventChan:
				handler(
					event.PreviousLiquidityTreasury,
					event.NewLiquidityTreasury,
					event.Raw.BlockNumber,
				)
			}
		}
	}()

	sub := ltus.Pipe(eventChan)
	return subscription.NewEventSubscription(func() {
		sub.Unsubscribe()
		cancelCtx()
	})
}

func (ltus *PLiquidityTreasuryUpdatedSubscription) Pipe(
	sink chan *abi.PortalLiquidityTreasuryUpdated,
) subscription.EventSubscription {
	ctx, cancelCtx := context.WithCancel(context.Background())
	go func() {
		ticker := time.NewTicker(ltus.opts.Tick)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				lastBlock, err := ltus.contract.blockCounter.CurrentBlock()
				if err != nil {
					pLogger.Errorf(
						"subscription failed to pull events: [%v]",
						err,
					)
				}
				fromBlock := lastBlock - ltus.opts.PastBlocks

				pLogger.Infof(
					"subscription monitoring fetching past LiquidityTreasuryUpdated events "+
						"starting from block [%v]",
					fromBlock,
				)
				events, err := ltus.contract.PastLiquidityTreasuryUpdatedEvents(
					fromBlock,
					nil,
					ltus.previousLiquidityTreasuryFilter,
					ltus.newLiquidityTreasuryFilter,
				)
				if err != nil {
					pLogger.Errorf(
						"subscription failed to pull events: [%v]",
						err,
					)
					continue
				}
				pLogger.Infof(
					"subscription monitoring fetched [%v] past LiquidityTreasuryUpdated events",
					len(events),
				)

				for _, event := range events {
					sink <- event
				}
			}
		}
	}()

	sub := ltus.contract.watchLiquidityTreasuryUpdated(
		sink,
		ltus.previousLiquidityTreasuryFilter,
		ltus.newLiquidityTreasuryFilter,
	)

	return subscription.NewEventSubscription(func() {
		sub.Unsubscribe()
		cancelCtx()
	})
}

func (p *Portal) watchLiquidityTreasuryUpdated(
	sink chan *abi.PortalLiquidityTreasuryUpdated,
	previousLiquidityTreasuryFilter []common.Address,
	newLiquidityTreasuryFilter []common.Address,
) event.Subscription {
	subscribeFn := func(ctx context.Context) (event.Subscription, error) {
		return p.contract.WatchLiquidityTreasuryUpdated(
			&bind.WatchOpts{Context: ctx},
			sink,
			previousLiquidityTreasuryFilter,
			newLiquidityTreasuryFilter,
		)
	}

	thresholdViolatedFn := func(elapsed time.Duration) {
		pLogger.Warnf(
			"subscription to event LiquidityTreasuryUpdated had to be "+
				"retried [%s] since the last attempt; please inspect "+
				"host chain connectivity",
			elapsed,
		)
	}

	subscriptionFailedFn := func(err error) {
		pLogger.Errorf(
			"subscription to event LiquidityTreasuryUpdated failed "+
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

func (p *Portal) PastLiquidityTreasuryUpdatedEvents(
	startBlock uint64,
	endBlock *uint64,
	previousLiquidityTreasuryFilter []common.Address,
	newLiquidityTreasuryFilter []common.Address,
) ([]*abi.PortalLiquidityTreasuryUpdated, error) {
	iterator, err := p.contract.FilterLiquidityTreasuryUpdated(
		&bind.FilterOpts{
			Start: startBlock,
			End:   endBlock,
		},
		previousLiquidityTreasuryFilter,
		newLiquidityTreasuryFilter,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"error retrieving past LiquidityTreasuryUpdated events: [%v]",
			err,
		)
	}

	events := make([]*abi.PortalLiquidityTreasuryUpdated, 0)

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

func (p *Portal) ReceiptMintedEvent(
	opts *ethereum.SubscribeOpts,
	depositorFilter []common.Address,
	tokenFilter []common.Address,
	depositIdFilter []*big.Int,
) *PReceiptMintedSubscription {
	if opts == nil {
		opts = new(ethereum.SubscribeOpts)
	}
	if opts.Tick == 0 {
		opts.Tick = chainutil.DefaultSubscribeOptsTick
	}
	if opts.PastBlocks == 0 {
		opts.PastBlocks = chainutil.DefaultSubscribeOptsPastBlocks
	}

	return &PReceiptMintedSubscription{
		p,
		opts,
		depositorFilter,
		tokenFilter,
		depositIdFilter,
	}
}

type PReceiptMintedSubscription struct {
	contract        *Portal
	opts            *ethereum.SubscribeOpts
	depositorFilter []common.Address
	tokenFilter     []common.Address
	depositIdFilter []*big.Int
}

type portalReceiptMintedFunc func(
	Depositor common.Address,
	Token common.Address,
	DepositId *big.Int,
	Amount *big.Int,
	blockNumber uint64,
)

func (rms *PReceiptMintedSubscription) OnEvent(
	handler portalReceiptMintedFunc,
) subscription.EventSubscription {
	eventChan := make(chan *abi.PortalReceiptMinted)
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

	sub := rms.Pipe(eventChan)
	return subscription.NewEventSubscription(func() {
		sub.Unsubscribe()
		cancelCtx()
	})
}

func (rms *PReceiptMintedSubscription) Pipe(
	sink chan *abi.PortalReceiptMinted,
) subscription.EventSubscription {
	ctx, cancelCtx := context.WithCancel(context.Background())
	go func() {
		ticker := time.NewTicker(rms.opts.Tick)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				lastBlock, err := rms.contract.blockCounter.CurrentBlock()
				if err != nil {
					pLogger.Errorf(
						"subscription failed to pull events: [%v]",
						err,
					)
				}
				fromBlock := lastBlock - rms.opts.PastBlocks

				pLogger.Infof(
					"subscription monitoring fetching past ReceiptMinted events "+
						"starting from block [%v]",
					fromBlock,
				)
				events, err := rms.contract.PastReceiptMintedEvents(
					fromBlock,
					nil,
					rms.depositorFilter,
					rms.tokenFilter,
					rms.depositIdFilter,
				)
				if err != nil {
					pLogger.Errorf(
						"subscription failed to pull events: [%v]",
						err,
					)
					continue
				}
				pLogger.Infof(
					"subscription monitoring fetched [%v] past ReceiptMinted events",
					len(events),
				)

				for _, event := range events {
					sink <- event
				}
			}
		}
	}()

	sub := rms.contract.watchReceiptMinted(
		sink,
		rms.depositorFilter,
		rms.tokenFilter,
		rms.depositIdFilter,
	)

	return subscription.NewEventSubscription(func() {
		sub.Unsubscribe()
		cancelCtx()
	})
}

func (p *Portal) watchReceiptMinted(
	sink chan *abi.PortalReceiptMinted,
	depositorFilter []common.Address,
	tokenFilter []common.Address,
	depositIdFilter []*big.Int,
) event.Subscription {
	subscribeFn := func(ctx context.Context) (event.Subscription, error) {
		return p.contract.WatchReceiptMinted(
			&bind.WatchOpts{Context: ctx},
			sink,
			depositorFilter,
			tokenFilter,
			depositIdFilter,
		)
	}

	thresholdViolatedFn := func(elapsed time.Duration) {
		pLogger.Warnf(
			"subscription to event ReceiptMinted had to be "+
				"retried [%s] since the last attempt; please inspect "+
				"host chain connectivity",
			elapsed,
		)
	}

	subscriptionFailedFn := func(err error) {
		pLogger.Errorf(
			"subscription to event ReceiptMinted failed "+
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

func (p *Portal) PastReceiptMintedEvents(
	startBlock uint64,
	endBlock *uint64,
	depositorFilter []common.Address,
	tokenFilter []common.Address,
	depositIdFilter []*big.Int,
) ([]*abi.PortalReceiptMinted, error) {
	iterator, err := p.contract.FilterReceiptMinted(
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
			"error retrieving past ReceiptMinted events: [%v]",
			err,
		)
	}

	events := make([]*abi.PortalReceiptMinted, 0)

	for iterator.Next() {
		event := iterator.Event
		events = append(events, event)
	}

	return events, nil
}

func (p *Portal) ReceiptParamsUpdatedEvent(
	opts *ethereum.SubscribeOpts,
	tokenFilter []common.Address,
) *PReceiptParamsUpdatedSubscription {
	if opts == nil {
		opts = new(ethereum.SubscribeOpts)
	}
	if opts.Tick == 0 {
		opts.Tick = chainutil.DefaultSubscribeOptsTick
	}
	if opts.PastBlocks == 0 {
		opts.PastBlocks = chainutil.DefaultSubscribeOptsPastBlocks
	}

	return &PReceiptParamsUpdatedSubscription{
		p,
		opts,
		tokenFilter,
	}
}

type PReceiptParamsUpdatedSubscription struct {
	contract    *Portal
	opts        *ethereum.SubscribeOpts
	tokenFilter []common.Address
}

type portalReceiptParamsUpdatedFunc func(
	Token common.Address,
	AnnualFee uint8,
	MintCap uint8,
	ReceiptToken common.Address,
	blockNumber uint64,
)

func (rpus *PReceiptParamsUpdatedSubscription) OnEvent(
	handler portalReceiptParamsUpdatedFunc,
) subscription.EventSubscription {
	eventChan := make(chan *abi.PortalReceiptParamsUpdated)
	ctx, cancelCtx := context.WithCancel(context.Background())

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case event := <-eventChan:
				handler(
					event.Token,
					event.AnnualFee,
					event.MintCap,
					event.ReceiptToken,
					event.Raw.BlockNumber,
				)
			}
		}
	}()

	sub := rpus.Pipe(eventChan)
	return subscription.NewEventSubscription(func() {
		sub.Unsubscribe()
		cancelCtx()
	})
}

func (rpus *PReceiptParamsUpdatedSubscription) Pipe(
	sink chan *abi.PortalReceiptParamsUpdated,
) subscription.EventSubscription {
	ctx, cancelCtx := context.WithCancel(context.Background())
	go func() {
		ticker := time.NewTicker(rpus.opts.Tick)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				lastBlock, err := rpus.contract.blockCounter.CurrentBlock()
				if err != nil {
					pLogger.Errorf(
						"subscription failed to pull events: [%v]",
						err,
					)
				}
				fromBlock := lastBlock - rpus.opts.PastBlocks

				pLogger.Infof(
					"subscription monitoring fetching past ReceiptParamsUpdated events "+
						"starting from block [%v]",
					fromBlock,
				)
				events, err := rpus.contract.PastReceiptParamsUpdatedEvents(
					fromBlock,
					nil,
					rpus.tokenFilter,
				)
				if err != nil {
					pLogger.Errorf(
						"subscription failed to pull events: [%v]",
						err,
					)
					continue
				}
				pLogger.Infof(
					"subscription monitoring fetched [%v] past ReceiptParamsUpdated events",
					len(events),
				)

				for _, event := range events {
					sink <- event
				}
			}
		}
	}()

	sub := rpus.contract.watchReceiptParamsUpdated(
		sink,
		rpus.tokenFilter,
	)

	return subscription.NewEventSubscription(func() {
		sub.Unsubscribe()
		cancelCtx()
	})
}

func (p *Portal) watchReceiptParamsUpdated(
	sink chan *abi.PortalReceiptParamsUpdated,
	tokenFilter []common.Address,
) event.Subscription {
	subscribeFn := func(ctx context.Context) (event.Subscription, error) {
		return p.contract.WatchReceiptParamsUpdated(
			&bind.WatchOpts{Context: ctx},
			sink,
			tokenFilter,
		)
	}

	thresholdViolatedFn := func(elapsed time.Duration) {
		pLogger.Warnf(
			"subscription to event ReceiptParamsUpdated had to be "+
				"retried [%s] since the last attempt; please inspect "+
				"host chain connectivity",
			elapsed,
		)
	}

	subscriptionFailedFn := func(err error) {
		pLogger.Errorf(
			"subscription to event ReceiptParamsUpdated failed "+
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

func (p *Portal) PastReceiptParamsUpdatedEvents(
	startBlock uint64,
	endBlock *uint64,
	tokenFilter []common.Address,
) ([]*abi.PortalReceiptParamsUpdated, error) {
	iterator, err := p.contract.FilterReceiptParamsUpdated(
		&bind.FilterOpts{
			Start: startBlock,
			End:   endBlock,
		},
		tokenFilter,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"error retrieving past ReceiptParamsUpdated events: [%v]",
			err,
		)
	}

	events := make([]*abi.PortalReceiptParamsUpdated, 0)

	for iterator.Next() {
		event := iterator.Event
		events = append(events, event)
	}

	return events, nil
}

func (p *Portal) ReceiptRepaidEvent(
	opts *ethereum.SubscribeOpts,
	depositorFilter []common.Address,
	tokenFilter []common.Address,
	depositIdFilter []*big.Int,
) *PReceiptRepaidSubscription {
	if opts == nil {
		opts = new(ethereum.SubscribeOpts)
	}
	if opts.Tick == 0 {
		opts.Tick = chainutil.DefaultSubscribeOptsTick
	}
	if opts.PastBlocks == 0 {
		opts.PastBlocks = chainutil.DefaultSubscribeOptsPastBlocks
	}

	return &PReceiptRepaidSubscription{
		p,
		opts,
		depositorFilter,
		tokenFilter,
		depositIdFilter,
	}
}

type PReceiptRepaidSubscription struct {
	contract        *Portal
	opts            *ethereum.SubscribeOpts
	depositorFilter []common.Address
	tokenFilter     []common.Address
	depositIdFilter []*big.Int
}

type portalReceiptRepaidFunc func(
	Depositor common.Address,
	Token common.Address,
	DepositId *big.Int,
	Amount *big.Int,
	blockNumber uint64,
)

func (rrs *PReceiptRepaidSubscription) OnEvent(
	handler portalReceiptRepaidFunc,
) subscription.EventSubscription {
	eventChan := make(chan *abi.PortalReceiptRepaid)
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

	sub := rrs.Pipe(eventChan)
	return subscription.NewEventSubscription(func() {
		sub.Unsubscribe()
		cancelCtx()
	})
}

func (rrs *PReceiptRepaidSubscription) Pipe(
	sink chan *abi.PortalReceiptRepaid,
) subscription.EventSubscription {
	ctx, cancelCtx := context.WithCancel(context.Background())
	go func() {
		ticker := time.NewTicker(rrs.opts.Tick)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				lastBlock, err := rrs.contract.blockCounter.CurrentBlock()
				if err != nil {
					pLogger.Errorf(
						"subscription failed to pull events: [%v]",
						err,
					)
				}
				fromBlock := lastBlock - rrs.opts.PastBlocks

				pLogger.Infof(
					"subscription monitoring fetching past ReceiptRepaid events "+
						"starting from block [%v]",
					fromBlock,
				)
				events, err := rrs.contract.PastReceiptRepaidEvents(
					fromBlock,
					nil,
					rrs.depositorFilter,
					rrs.tokenFilter,
					rrs.depositIdFilter,
				)
				if err != nil {
					pLogger.Errorf(
						"subscription failed to pull events: [%v]",
						err,
					)
					continue
				}
				pLogger.Infof(
					"subscription monitoring fetched [%v] past ReceiptRepaid events",
					len(events),
				)

				for _, event := range events {
					sink <- event
				}
			}
		}
	}()

	sub := rrs.contract.watchReceiptRepaid(
		sink,
		rrs.depositorFilter,
		rrs.tokenFilter,
		rrs.depositIdFilter,
	)

	return subscription.NewEventSubscription(func() {
		sub.Unsubscribe()
		cancelCtx()
	})
}

func (p *Portal) watchReceiptRepaid(
	sink chan *abi.PortalReceiptRepaid,
	depositorFilter []common.Address,
	tokenFilter []common.Address,
	depositIdFilter []*big.Int,
) event.Subscription {
	subscribeFn := func(ctx context.Context) (event.Subscription, error) {
		return p.contract.WatchReceiptRepaid(
			&bind.WatchOpts{Context: ctx},
			sink,
			depositorFilter,
			tokenFilter,
			depositIdFilter,
		)
	}

	thresholdViolatedFn := func(elapsed time.Duration) {
		pLogger.Warnf(
			"subscription to event ReceiptRepaid had to be "+
				"retried [%s] since the last attempt; please inspect "+
				"host chain connectivity",
			elapsed,
		)
	}

	subscriptionFailedFn := func(err error) {
		pLogger.Errorf(
			"subscription to event ReceiptRepaid failed "+
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

func (p *Portal) PastReceiptRepaidEvents(
	startBlock uint64,
	endBlock *uint64,
	depositorFilter []common.Address,
	tokenFilter []common.Address,
	depositIdFilter []*big.Int,
) ([]*abi.PortalReceiptRepaid, error) {
	iterator, err := p.contract.FilterReceiptRepaid(
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
			"error retrieving past ReceiptRepaid events: [%v]",
			err,
		)
	}

	events := make([]*abi.PortalReceiptRepaid, 0)

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

func (p *Portal) TbtcMigrationAllowedUpdatedEvent(
	opts *ethereum.SubscribeOpts,
	tokenFilter []common.Address,
) *PTbtcMigrationAllowedUpdatedSubscription {
	if opts == nil {
		opts = new(ethereum.SubscribeOpts)
	}
	if opts.Tick == 0 {
		opts.Tick = chainutil.DefaultSubscribeOptsTick
	}
	if opts.PastBlocks == 0 {
		opts.PastBlocks = chainutil.DefaultSubscribeOptsPastBlocks
	}

	return &PTbtcMigrationAllowedUpdatedSubscription{
		p,
		opts,
		tokenFilter,
	}
}

type PTbtcMigrationAllowedUpdatedSubscription struct {
	contract    *Portal
	opts        *ethereum.SubscribeOpts
	tokenFilter []common.Address
}

type portalTbtcMigrationAllowedUpdatedFunc func(
	Token common.Address,
	IsAllowed bool,
	blockNumber uint64,
)

func (tmaus *PTbtcMigrationAllowedUpdatedSubscription) OnEvent(
	handler portalTbtcMigrationAllowedUpdatedFunc,
) subscription.EventSubscription {
	eventChan := make(chan *abi.PortalTbtcMigrationAllowedUpdated)
	ctx, cancelCtx := context.WithCancel(context.Background())

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case event := <-eventChan:
				handler(
					event.Token,
					event.IsAllowed,
					event.Raw.BlockNumber,
				)
			}
		}
	}()

	sub := tmaus.Pipe(eventChan)
	return subscription.NewEventSubscription(func() {
		sub.Unsubscribe()
		cancelCtx()
	})
}

func (tmaus *PTbtcMigrationAllowedUpdatedSubscription) Pipe(
	sink chan *abi.PortalTbtcMigrationAllowedUpdated,
) subscription.EventSubscription {
	ctx, cancelCtx := context.WithCancel(context.Background())
	go func() {
		ticker := time.NewTicker(tmaus.opts.Tick)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				lastBlock, err := tmaus.contract.blockCounter.CurrentBlock()
				if err != nil {
					pLogger.Errorf(
						"subscription failed to pull events: [%v]",
						err,
					)
				}
				fromBlock := lastBlock - tmaus.opts.PastBlocks

				pLogger.Infof(
					"subscription monitoring fetching past TbtcMigrationAllowedUpdated events "+
						"starting from block [%v]",
					fromBlock,
				)
				events, err := tmaus.contract.PastTbtcMigrationAllowedUpdatedEvents(
					fromBlock,
					nil,
					tmaus.tokenFilter,
				)
				if err != nil {
					pLogger.Errorf(
						"subscription failed to pull events: [%v]",
						err,
					)
					continue
				}
				pLogger.Infof(
					"subscription monitoring fetched [%v] past TbtcMigrationAllowedUpdated events",
					len(events),
				)

				for _, event := range events {
					sink <- event
				}
			}
		}
	}()

	sub := tmaus.contract.watchTbtcMigrationAllowedUpdated(
		sink,
		tmaus.tokenFilter,
	)

	return subscription.NewEventSubscription(func() {
		sub.Unsubscribe()
		cancelCtx()
	})
}

func (p *Portal) watchTbtcMigrationAllowedUpdated(
	sink chan *abi.PortalTbtcMigrationAllowedUpdated,
	tokenFilter []common.Address,
) event.Subscription {
	subscribeFn := func(ctx context.Context) (event.Subscription, error) {
		return p.contract.WatchTbtcMigrationAllowedUpdated(
			&bind.WatchOpts{Context: ctx},
			sink,
			tokenFilter,
		)
	}

	thresholdViolatedFn := func(elapsed time.Duration) {
		pLogger.Warnf(
			"subscription to event TbtcMigrationAllowedUpdated had to be "+
				"retried [%s] since the last attempt; please inspect "+
				"host chain connectivity",
			elapsed,
		)
	}

	subscriptionFailedFn := func(err error) {
		pLogger.Errorf(
			"subscription to event TbtcMigrationAllowedUpdated failed "+
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

func (p *Portal) PastTbtcMigrationAllowedUpdatedEvents(
	startBlock uint64,
	endBlock *uint64,
	tokenFilter []common.Address,
) ([]*abi.PortalTbtcMigrationAllowedUpdated, error) {
	iterator, err := p.contract.FilterTbtcMigrationAllowedUpdated(
		&bind.FilterOpts{
			Start: startBlock,
			End:   endBlock,
		},
		tokenFilter,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"error retrieving past TbtcMigrationAllowedUpdated events: [%v]",
			err,
		)
	}

	events := make([]*abi.PortalTbtcMigrationAllowedUpdated, 0)

	for iterator.Next() {
		event := iterator.Event
		events = append(events, event)
	}

	return events, nil
}

func (p *Portal) TbtcMigrationCompletedEvent(
	opts *ethereum.SubscribeOpts,
	depositorFilter []common.Address,
	tokenFilter []common.Address,
	depositIdFilter []*big.Int,
) *PTbtcMigrationCompletedSubscription {
	if opts == nil {
		opts = new(ethereum.SubscribeOpts)
	}
	if opts.Tick == 0 {
		opts.Tick = chainutil.DefaultSubscribeOptsTick
	}
	if opts.PastBlocks == 0 {
		opts.PastBlocks = chainutil.DefaultSubscribeOptsPastBlocks
	}

	return &PTbtcMigrationCompletedSubscription{
		p,
		opts,
		depositorFilter,
		tokenFilter,
		depositIdFilter,
	}
}

type PTbtcMigrationCompletedSubscription struct {
	contract        *Portal
	opts            *ethereum.SubscribeOpts
	depositorFilter []common.Address
	tokenFilter     []common.Address
	depositIdFilter []*big.Int
}

type portalTbtcMigrationCompletedFunc func(
	Depositor common.Address,
	Token common.Address,
	DepositId *big.Int,
	blockNumber uint64,
)

func (tmcs *PTbtcMigrationCompletedSubscription) OnEvent(
	handler portalTbtcMigrationCompletedFunc,
) subscription.EventSubscription {
	eventChan := make(chan *abi.PortalTbtcMigrationCompleted)
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
					event.Raw.BlockNumber,
				)
			}
		}
	}()

	sub := tmcs.Pipe(eventChan)
	return subscription.NewEventSubscription(func() {
		sub.Unsubscribe()
		cancelCtx()
	})
}

func (tmcs *PTbtcMigrationCompletedSubscription) Pipe(
	sink chan *abi.PortalTbtcMigrationCompleted,
) subscription.EventSubscription {
	ctx, cancelCtx := context.WithCancel(context.Background())
	go func() {
		ticker := time.NewTicker(tmcs.opts.Tick)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				lastBlock, err := tmcs.contract.blockCounter.CurrentBlock()
				if err != nil {
					pLogger.Errorf(
						"subscription failed to pull events: [%v]",
						err,
					)
				}
				fromBlock := lastBlock - tmcs.opts.PastBlocks

				pLogger.Infof(
					"subscription monitoring fetching past TbtcMigrationCompleted events "+
						"starting from block [%v]",
					fromBlock,
				)
				events, err := tmcs.contract.PastTbtcMigrationCompletedEvents(
					fromBlock,
					nil,
					tmcs.depositorFilter,
					tmcs.tokenFilter,
					tmcs.depositIdFilter,
				)
				if err != nil {
					pLogger.Errorf(
						"subscription failed to pull events: [%v]",
						err,
					)
					continue
				}
				pLogger.Infof(
					"subscription monitoring fetched [%v] past TbtcMigrationCompleted events",
					len(events),
				)

				for _, event := range events {
					sink <- event
				}
			}
		}
	}()

	sub := tmcs.contract.watchTbtcMigrationCompleted(
		sink,
		tmcs.depositorFilter,
		tmcs.tokenFilter,
		tmcs.depositIdFilter,
	)

	return subscription.NewEventSubscription(func() {
		sub.Unsubscribe()
		cancelCtx()
	})
}

func (p *Portal) watchTbtcMigrationCompleted(
	sink chan *abi.PortalTbtcMigrationCompleted,
	depositorFilter []common.Address,
	tokenFilter []common.Address,
	depositIdFilter []*big.Int,
) event.Subscription {
	subscribeFn := func(ctx context.Context) (event.Subscription, error) {
		return p.contract.WatchTbtcMigrationCompleted(
			&bind.WatchOpts{Context: ctx},
			sink,
			depositorFilter,
			tokenFilter,
			depositIdFilter,
		)
	}

	thresholdViolatedFn := func(elapsed time.Duration) {
		pLogger.Warnf(
			"subscription to event TbtcMigrationCompleted had to be "+
				"retried [%s] since the last attempt; please inspect "+
				"host chain connectivity",
			elapsed,
		)
	}

	subscriptionFailedFn := func(err error) {
		pLogger.Errorf(
			"subscription to event TbtcMigrationCompleted failed "+
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

func (p *Portal) PastTbtcMigrationCompletedEvents(
	startBlock uint64,
	endBlock *uint64,
	depositorFilter []common.Address,
	tokenFilter []common.Address,
	depositIdFilter []*big.Int,
) ([]*abi.PortalTbtcMigrationCompleted, error) {
	iterator, err := p.contract.FilterTbtcMigrationCompleted(
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
			"error retrieving past TbtcMigrationCompleted events: [%v]",
			err,
		)
	}

	events := make([]*abi.PortalTbtcMigrationCompleted, 0)

	for iterator.Next() {
		event := iterator.Event
		events = append(events, event)
	}

	return events, nil
}

func (p *Portal) TbtcMigrationRequestedEvent(
	opts *ethereum.SubscribeOpts,
	depositorFilter []common.Address,
	tokenFilter []common.Address,
	depositIdFilter []*big.Int,
) *PTbtcMigrationRequestedSubscription {
	if opts == nil {
		opts = new(ethereum.SubscribeOpts)
	}
	if opts.Tick == 0 {
		opts.Tick = chainutil.DefaultSubscribeOptsTick
	}
	if opts.PastBlocks == 0 {
		opts.PastBlocks = chainutil.DefaultSubscribeOptsPastBlocks
	}

	return &PTbtcMigrationRequestedSubscription{
		p,
		opts,
		depositorFilter,
		tokenFilter,
		depositIdFilter,
	}
}

type PTbtcMigrationRequestedSubscription struct {
	contract        *Portal
	opts            *ethereum.SubscribeOpts
	depositorFilter []common.Address
	tokenFilter     []common.Address
	depositIdFilter []*big.Int
}

type portalTbtcMigrationRequestedFunc func(
	Depositor common.Address,
	Token common.Address,
	DepositId *big.Int,
	blockNumber uint64,
)

func (tmrs *PTbtcMigrationRequestedSubscription) OnEvent(
	handler portalTbtcMigrationRequestedFunc,
) subscription.EventSubscription {
	eventChan := make(chan *abi.PortalTbtcMigrationRequested)
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
					event.Raw.BlockNumber,
				)
			}
		}
	}()

	sub := tmrs.Pipe(eventChan)
	return subscription.NewEventSubscription(func() {
		sub.Unsubscribe()
		cancelCtx()
	})
}

func (tmrs *PTbtcMigrationRequestedSubscription) Pipe(
	sink chan *abi.PortalTbtcMigrationRequested,
) subscription.EventSubscription {
	ctx, cancelCtx := context.WithCancel(context.Background())
	go func() {
		ticker := time.NewTicker(tmrs.opts.Tick)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				lastBlock, err := tmrs.contract.blockCounter.CurrentBlock()
				if err != nil {
					pLogger.Errorf(
						"subscription failed to pull events: [%v]",
						err,
					)
				}
				fromBlock := lastBlock - tmrs.opts.PastBlocks

				pLogger.Infof(
					"subscription monitoring fetching past TbtcMigrationRequested events "+
						"starting from block [%v]",
					fromBlock,
				)
				events, err := tmrs.contract.PastTbtcMigrationRequestedEvents(
					fromBlock,
					nil,
					tmrs.depositorFilter,
					tmrs.tokenFilter,
					tmrs.depositIdFilter,
				)
				if err != nil {
					pLogger.Errorf(
						"subscription failed to pull events: [%v]",
						err,
					)
					continue
				}
				pLogger.Infof(
					"subscription monitoring fetched [%v] past TbtcMigrationRequested events",
					len(events),
				)

				for _, event := range events {
					sink <- event
				}
			}
		}
	}()

	sub := tmrs.contract.watchTbtcMigrationRequested(
		sink,
		tmrs.depositorFilter,
		tmrs.tokenFilter,
		tmrs.depositIdFilter,
	)

	return subscription.NewEventSubscription(func() {
		sub.Unsubscribe()
		cancelCtx()
	})
}

func (p *Portal) watchTbtcMigrationRequested(
	sink chan *abi.PortalTbtcMigrationRequested,
	depositorFilter []common.Address,
	tokenFilter []common.Address,
	depositIdFilter []*big.Int,
) event.Subscription {
	subscribeFn := func(ctx context.Context) (event.Subscription, error) {
		return p.contract.WatchTbtcMigrationRequested(
			&bind.WatchOpts{Context: ctx},
			sink,
			depositorFilter,
			tokenFilter,
			depositIdFilter,
		)
	}

	thresholdViolatedFn := func(elapsed time.Duration) {
		pLogger.Warnf(
			"subscription to event TbtcMigrationRequested had to be "+
				"retried [%s] since the last attempt; please inspect "+
				"host chain connectivity",
			elapsed,
		)
	}

	subscriptionFailedFn := func(err error) {
		pLogger.Errorf(
			"subscription to event TbtcMigrationRequested failed "+
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

func (p *Portal) PastTbtcMigrationRequestedEvents(
	startBlock uint64,
	endBlock *uint64,
	depositorFilter []common.Address,
	tokenFilter []common.Address,
	depositIdFilter []*big.Int,
) ([]*abi.PortalTbtcMigrationRequested, error) {
	iterator, err := p.contract.FilterTbtcMigrationRequested(
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
			"error retrieving past TbtcMigrationRequested events: [%v]",
			err,
		)
	}

	events := make([]*abi.PortalTbtcMigrationRequested, 0)

	for iterator.Next() {
		event := iterator.Event
		events = append(events, event)
	}

	return events, nil
}

func (p *Portal) TbtcMigrationStartedEvent(
	opts *ethereum.SubscribeOpts,
	depositorFilter []common.Address,
	tokenFilter []common.Address,
	depositIdFilter []*big.Int,
) *PTbtcMigrationStartedSubscription {
	if opts == nil {
		opts = new(ethereum.SubscribeOpts)
	}
	if opts.Tick == 0 {
		opts.Tick = chainutil.DefaultSubscribeOptsTick
	}
	if opts.PastBlocks == 0 {
		opts.PastBlocks = chainutil.DefaultSubscribeOptsPastBlocks
	}

	return &PTbtcMigrationStartedSubscription{
		p,
		opts,
		depositorFilter,
		tokenFilter,
		depositIdFilter,
	}
}

type PTbtcMigrationStartedSubscription struct {
	contract        *Portal
	opts            *ethereum.SubscribeOpts
	depositorFilter []common.Address
	tokenFilter     []common.Address
	depositIdFilter []*big.Int
}

type portalTbtcMigrationStartedFunc func(
	Depositor common.Address,
	Token common.Address,
	DepositId *big.Int,
	blockNumber uint64,
)

func (tmss *PTbtcMigrationStartedSubscription) OnEvent(
	handler portalTbtcMigrationStartedFunc,
) subscription.EventSubscription {
	eventChan := make(chan *abi.PortalTbtcMigrationStarted)
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
					event.Raw.BlockNumber,
				)
			}
		}
	}()

	sub := tmss.Pipe(eventChan)
	return subscription.NewEventSubscription(func() {
		sub.Unsubscribe()
		cancelCtx()
	})
}

func (tmss *PTbtcMigrationStartedSubscription) Pipe(
	sink chan *abi.PortalTbtcMigrationStarted,
) subscription.EventSubscription {
	ctx, cancelCtx := context.WithCancel(context.Background())
	go func() {
		ticker := time.NewTicker(tmss.opts.Tick)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				lastBlock, err := tmss.contract.blockCounter.CurrentBlock()
				if err != nil {
					pLogger.Errorf(
						"subscription failed to pull events: [%v]",
						err,
					)
				}
				fromBlock := lastBlock - tmss.opts.PastBlocks

				pLogger.Infof(
					"subscription monitoring fetching past TbtcMigrationStarted events "+
						"starting from block [%v]",
					fromBlock,
				)
				events, err := tmss.contract.PastTbtcMigrationStartedEvents(
					fromBlock,
					nil,
					tmss.depositorFilter,
					tmss.tokenFilter,
					tmss.depositIdFilter,
				)
				if err != nil {
					pLogger.Errorf(
						"subscription failed to pull events: [%v]",
						err,
					)
					continue
				}
				pLogger.Infof(
					"subscription monitoring fetched [%v] past TbtcMigrationStarted events",
					len(events),
				)

				for _, event := range events {
					sink <- event
				}
			}
		}
	}()

	sub := tmss.contract.watchTbtcMigrationStarted(
		sink,
		tmss.depositorFilter,
		tmss.tokenFilter,
		tmss.depositIdFilter,
	)

	return subscription.NewEventSubscription(func() {
		sub.Unsubscribe()
		cancelCtx()
	})
}

func (p *Portal) watchTbtcMigrationStarted(
	sink chan *abi.PortalTbtcMigrationStarted,
	depositorFilter []common.Address,
	tokenFilter []common.Address,
	depositIdFilter []*big.Int,
) event.Subscription {
	subscribeFn := func(ctx context.Context) (event.Subscription, error) {
		return p.contract.WatchTbtcMigrationStarted(
			&bind.WatchOpts{Context: ctx},
			sink,
			depositorFilter,
			tokenFilter,
			depositIdFilter,
		)
	}

	thresholdViolatedFn := func(elapsed time.Duration) {
		pLogger.Warnf(
			"subscription to event TbtcMigrationStarted had to be "+
				"retried [%s] since the last attempt; please inspect "+
				"host chain connectivity",
			elapsed,
		)
	}

	subscriptionFailedFn := func(err error) {
		pLogger.Errorf(
			"subscription to event TbtcMigrationStarted failed "+
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

func (p *Portal) PastTbtcMigrationStartedEvents(
	startBlock uint64,
	endBlock *uint64,
	depositorFilter []common.Address,
	tokenFilter []common.Address,
	depositIdFilter []*big.Int,
) ([]*abi.PortalTbtcMigrationStarted, error) {
	iterator, err := p.contract.FilterTbtcMigrationStarted(
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
			"error retrieving past TbtcMigrationStarted events: [%v]",
			err,
		)
	}

	events := make([]*abi.PortalTbtcMigrationStarted, 0)

	for iterator.Next() {
		event := iterator.Event
		events = append(events, event)
	}

	return events, nil
}

func (p *Portal) TbtcMigrationTreasuryUpdatedEvent(
	opts *ethereum.SubscribeOpts,
	previousMigrationTreasuryFilter []common.Address,
	newMigrationTreasuryFilter []common.Address,
) *PTbtcMigrationTreasuryUpdatedSubscription {
	if opts == nil {
		opts = new(ethereum.SubscribeOpts)
	}
	if opts.Tick == 0 {
		opts.Tick = chainutil.DefaultSubscribeOptsTick
	}
	if opts.PastBlocks == 0 {
		opts.PastBlocks = chainutil.DefaultSubscribeOptsPastBlocks
	}

	return &PTbtcMigrationTreasuryUpdatedSubscription{
		p,
		opts,
		previousMigrationTreasuryFilter,
		newMigrationTreasuryFilter,
	}
}

type PTbtcMigrationTreasuryUpdatedSubscription struct {
	contract                        *Portal
	opts                            *ethereum.SubscribeOpts
	previousMigrationTreasuryFilter []common.Address
	newMigrationTreasuryFilter      []common.Address
}

type portalTbtcMigrationTreasuryUpdatedFunc func(
	PreviousMigrationTreasury common.Address,
	NewMigrationTreasury common.Address,
	blockNumber uint64,
)

func (tmtus *PTbtcMigrationTreasuryUpdatedSubscription) OnEvent(
	handler portalTbtcMigrationTreasuryUpdatedFunc,
) subscription.EventSubscription {
	eventChan := make(chan *abi.PortalTbtcMigrationTreasuryUpdated)
	ctx, cancelCtx := context.WithCancel(context.Background())

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case event := <-eventChan:
				handler(
					event.PreviousMigrationTreasury,
					event.NewMigrationTreasury,
					event.Raw.BlockNumber,
				)
			}
		}
	}()

	sub := tmtus.Pipe(eventChan)
	return subscription.NewEventSubscription(func() {
		sub.Unsubscribe()
		cancelCtx()
	})
}

func (tmtus *PTbtcMigrationTreasuryUpdatedSubscription) Pipe(
	sink chan *abi.PortalTbtcMigrationTreasuryUpdated,
) subscription.EventSubscription {
	ctx, cancelCtx := context.WithCancel(context.Background())
	go func() {
		ticker := time.NewTicker(tmtus.opts.Tick)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				lastBlock, err := tmtus.contract.blockCounter.CurrentBlock()
				if err != nil {
					pLogger.Errorf(
						"subscription failed to pull events: [%v]",
						err,
					)
				}
				fromBlock := lastBlock - tmtus.opts.PastBlocks

				pLogger.Infof(
					"subscription monitoring fetching past TbtcMigrationTreasuryUpdated events "+
						"starting from block [%v]",
					fromBlock,
				)
				events, err := tmtus.contract.PastTbtcMigrationTreasuryUpdatedEvents(
					fromBlock,
					nil,
					tmtus.previousMigrationTreasuryFilter,
					tmtus.newMigrationTreasuryFilter,
				)
				if err != nil {
					pLogger.Errorf(
						"subscription failed to pull events: [%v]",
						err,
					)
					continue
				}
				pLogger.Infof(
					"subscription monitoring fetched [%v] past TbtcMigrationTreasuryUpdated events",
					len(events),
				)

				for _, event := range events {
					sink <- event
				}
			}
		}
	}()

	sub := tmtus.contract.watchTbtcMigrationTreasuryUpdated(
		sink,
		tmtus.previousMigrationTreasuryFilter,
		tmtus.newMigrationTreasuryFilter,
	)

	return subscription.NewEventSubscription(func() {
		sub.Unsubscribe()
		cancelCtx()
	})
}

func (p *Portal) watchTbtcMigrationTreasuryUpdated(
	sink chan *abi.PortalTbtcMigrationTreasuryUpdated,
	previousMigrationTreasuryFilter []common.Address,
	newMigrationTreasuryFilter []common.Address,
) event.Subscription {
	subscribeFn := func(ctx context.Context) (event.Subscription, error) {
		return p.contract.WatchTbtcMigrationTreasuryUpdated(
			&bind.WatchOpts{Context: ctx},
			sink,
			previousMigrationTreasuryFilter,
			newMigrationTreasuryFilter,
		)
	}

	thresholdViolatedFn := func(elapsed time.Duration) {
		pLogger.Warnf(
			"subscription to event TbtcMigrationTreasuryUpdated had to be "+
				"retried [%s] since the last attempt; please inspect "+
				"host chain connectivity",
			elapsed,
		)
	}

	subscriptionFailedFn := func(err error) {
		pLogger.Errorf(
			"subscription to event TbtcMigrationTreasuryUpdated failed "+
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

func (p *Portal) PastTbtcMigrationTreasuryUpdatedEvents(
	startBlock uint64,
	endBlock *uint64,
	previousMigrationTreasuryFilter []common.Address,
	newMigrationTreasuryFilter []common.Address,
) ([]*abi.PortalTbtcMigrationTreasuryUpdated, error) {
	iterator, err := p.contract.FilterTbtcMigrationTreasuryUpdated(
		&bind.FilterOpts{
			Start: startBlock,
			End:   endBlock,
		},
		previousMigrationTreasuryFilter,
		newMigrationTreasuryFilter,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"error retrieving past TbtcMigrationTreasuryUpdated events: [%v]",
			err,
		)
	}

	events := make([]*abi.PortalTbtcMigrationTreasuryUpdated, 0)

	for iterator.Next() {
		event := iterator.Event
		events = append(events, event)
	}

	return events, nil
}

func (p *Portal) TbtcTokenAddressSetEvent(
	opts *ethereum.SubscribeOpts,
) *PTbtcTokenAddressSetSubscription {
	if opts == nil {
		opts = new(ethereum.SubscribeOpts)
	}
	if opts.Tick == 0 {
		opts.Tick = chainutil.DefaultSubscribeOptsTick
	}
	if opts.PastBlocks == 0 {
		opts.PastBlocks = chainutil.DefaultSubscribeOptsPastBlocks
	}

	return &PTbtcTokenAddressSetSubscription{
		p,
		opts,
	}
}

type PTbtcTokenAddressSetSubscription struct {
	contract *Portal
	opts     *ethereum.SubscribeOpts
}

type portalTbtcTokenAddressSetFunc func(
	Tbtc common.Address,
	blockNumber uint64,
)

func (ttass *PTbtcTokenAddressSetSubscription) OnEvent(
	handler portalTbtcTokenAddressSetFunc,
) subscription.EventSubscription {
	eventChan := make(chan *abi.PortalTbtcTokenAddressSet)
	ctx, cancelCtx := context.WithCancel(context.Background())

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case event := <-eventChan:
				handler(
					event.Tbtc,
					event.Raw.BlockNumber,
				)
			}
		}
	}()

	sub := ttass.Pipe(eventChan)
	return subscription.NewEventSubscription(func() {
		sub.Unsubscribe()
		cancelCtx()
	})
}

func (ttass *PTbtcTokenAddressSetSubscription) Pipe(
	sink chan *abi.PortalTbtcTokenAddressSet,
) subscription.EventSubscription {
	ctx, cancelCtx := context.WithCancel(context.Background())
	go func() {
		ticker := time.NewTicker(ttass.opts.Tick)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				lastBlock, err := ttass.contract.blockCounter.CurrentBlock()
				if err != nil {
					pLogger.Errorf(
						"subscription failed to pull events: [%v]",
						err,
					)
				}
				fromBlock := lastBlock - ttass.opts.PastBlocks

				pLogger.Infof(
					"subscription monitoring fetching past TbtcTokenAddressSet events "+
						"starting from block [%v]",
					fromBlock,
				)
				events, err := ttass.contract.PastTbtcTokenAddressSetEvents(
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
					"subscription monitoring fetched [%v] past TbtcTokenAddressSet events",
					len(events),
				)

				for _, event := range events {
					sink <- event
				}
			}
		}
	}()

	sub := ttass.contract.watchTbtcTokenAddressSet(
		sink,
	)

	return subscription.NewEventSubscription(func() {
		sub.Unsubscribe()
		cancelCtx()
	})
}

func (p *Portal) watchTbtcTokenAddressSet(
	sink chan *abi.PortalTbtcTokenAddressSet,
) event.Subscription {
	subscribeFn := func(ctx context.Context) (event.Subscription, error) {
		return p.contract.WatchTbtcTokenAddressSet(
			&bind.WatchOpts{Context: ctx},
			sink,
		)
	}

	thresholdViolatedFn := func(elapsed time.Duration) {
		pLogger.Warnf(
			"subscription to event TbtcTokenAddressSet had to be "+
				"retried [%s] since the last attempt; please inspect "+
				"host chain connectivity",
			elapsed,
		)
	}

	subscriptionFailedFn := func(err error) {
		pLogger.Errorf(
			"subscription to event TbtcTokenAddressSet failed "+
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

func (p *Portal) PastTbtcTokenAddressSetEvents(
	startBlock uint64,
	endBlock *uint64,
) ([]*abi.PortalTbtcTokenAddressSet, error) {
	iterator, err := p.contract.FilterTbtcTokenAddressSet(
		&bind.FilterOpts{
			Start: startBlock,
			End:   endBlock,
		},
	)
	if err != nil {
		return nil, fmt.Errorf(
			"error retrieving past TbtcTokenAddressSet events: [%v]",
			err,
		)
	}

	events := make([]*abi.PortalTbtcTokenAddressSet, 0)

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

func (p *Portal) WithdrawnByLiquidityTreasuryEvent(
	opts *ethereum.SubscribeOpts,
	tokenFilter []common.Address,
) *PWithdrawnByLiquidityTreasurySubscription {
	if opts == nil {
		opts = new(ethereum.SubscribeOpts)
	}
	if opts.Tick == 0 {
		opts.Tick = chainutil.DefaultSubscribeOptsTick
	}
	if opts.PastBlocks == 0 {
		opts.PastBlocks = chainutil.DefaultSubscribeOptsPastBlocks
	}

	return &PWithdrawnByLiquidityTreasurySubscription{
		p,
		opts,
		tokenFilter,
	}
}

type PWithdrawnByLiquidityTreasurySubscription struct {
	contract    *Portal
	opts        *ethereum.SubscribeOpts
	tokenFilter []common.Address
}

type portalWithdrawnByLiquidityTreasuryFunc func(
	Token common.Address,
	Amount *big.Int,
	blockNumber uint64,
)

func (wblts *PWithdrawnByLiquidityTreasurySubscription) OnEvent(
	handler portalWithdrawnByLiquidityTreasuryFunc,
) subscription.EventSubscription {
	eventChan := make(chan *abi.PortalWithdrawnByLiquidityTreasury)
	ctx, cancelCtx := context.WithCancel(context.Background())

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case event := <-eventChan:
				handler(
					event.Token,
					event.Amount,
					event.Raw.BlockNumber,
				)
			}
		}
	}()

	sub := wblts.Pipe(eventChan)
	return subscription.NewEventSubscription(func() {
		sub.Unsubscribe()
		cancelCtx()
	})
}

func (wblts *PWithdrawnByLiquidityTreasurySubscription) Pipe(
	sink chan *abi.PortalWithdrawnByLiquidityTreasury,
) subscription.EventSubscription {
	ctx, cancelCtx := context.WithCancel(context.Background())
	go func() {
		ticker := time.NewTicker(wblts.opts.Tick)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				lastBlock, err := wblts.contract.blockCounter.CurrentBlock()
				if err != nil {
					pLogger.Errorf(
						"subscription failed to pull events: [%v]",
						err,
					)
				}
				fromBlock := lastBlock - wblts.opts.PastBlocks

				pLogger.Infof(
					"subscription monitoring fetching past WithdrawnByLiquidityTreasury events "+
						"starting from block [%v]",
					fromBlock,
				)
				events, err := wblts.contract.PastWithdrawnByLiquidityTreasuryEvents(
					fromBlock,
					nil,
					wblts.tokenFilter,
				)
				if err != nil {
					pLogger.Errorf(
						"subscription failed to pull events: [%v]",
						err,
					)
					continue
				}
				pLogger.Infof(
					"subscription monitoring fetched [%v] past WithdrawnByLiquidityTreasury events",
					len(events),
				)

				for _, event := range events {
					sink <- event
				}
			}
		}
	}()

	sub := wblts.contract.watchWithdrawnByLiquidityTreasury(
		sink,
		wblts.tokenFilter,
	)

	return subscription.NewEventSubscription(func() {
		sub.Unsubscribe()
		cancelCtx()
	})
}

func (p *Portal) watchWithdrawnByLiquidityTreasury(
	sink chan *abi.PortalWithdrawnByLiquidityTreasury,
	tokenFilter []common.Address,
) event.Subscription {
	subscribeFn := func(ctx context.Context) (event.Subscription, error) {
		return p.contract.WatchWithdrawnByLiquidityTreasury(
			&bind.WatchOpts{Context: ctx},
			sink,
			tokenFilter,
		)
	}

	thresholdViolatedFn := func(elapsed time.Duration) {
		pLogger.Warnf(
			"subscription to event WithdrawnByLiquidityTreasury had to be "+
				"retried [%s] since the last attempt; please inspect "+
				"host chain connectivity",
			elapsed,
		)
	}

	subscriptionFailedFn := func(err error) {
		pLogger.Errorf(
			"subscription to event WithdrawnByLiquidityTreasury failed "+
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

func (p *Portal) PastWithdrawnByLiquidityTreasuryEvents(
	startBlock uint64,
	endBlock *uint64,
	tokenFilter []common.Address,
) ([]*abi.PortalWithdrawnByLiquidityTreasury, error) {
	iterator, err := p.contract.FilterWithdrawnByLiquidityTreasury(
		&bind.FilterOpts{
			Start: startBlock,
			End:   endBlock,
		},
		tokenFilter,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"error retrieving past WithdrawnByLiquidityTreasury events: [%v]",
			err,
		)
	}

	events := make([]*abi.PortalWithdrawnByLiquidityTreasury, 0)

	for iterator.Next() {
		event := iterator.Event
		events = append(events, event)
	}

	return events, nil
}

func (p *Portal) WithdrawnForTbtcMigrationEvent(
	opts *ethereum.SubscribeOpts,
	tokenFilter []common.Address,
) *PWithdrawnForTbtcMigrationSubscription {
	if opts == nil {
		opts = new(ethereum.SubscribeOpts)
	}
	if opts.Tick == 0 {
		opts.Tick = chainutil.DefaultSubscribeOptsTick
	}
	if opts.PastBlocks == 0 {
		opts.PastBlocks = chainutil.DefaultSubscribeOptsPastBlocks
	}

	return &PWithdrawnForTbtcMigrationSubscription{
		p,
		opts,
		tokenFilter,
	}
}

type PWithdrawnForTbtcMigrationSubscription struct {
	contract    *Portal
	opts        *ethereum.SubscribeOpts
	tokenFilter []common.Address
}

type portalWithdrawnForTbtcMigrationFunc func(
	Token common.Address,
	Amount *big.Int,
	blockNumber uint64,
)

func (wftms *PWithdrawnForTbtcMigrationSubscription) OnEvent(
	handler portalWithdrawnForTbtcMigrationFunc,
) subscription.EventSubscription {
	eventChan := make(chan *abi.PortalWithdrawnForTbtcMigration)
	ctx, cancelCtx := context.WithCancel(context.Background())

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case event := <-eventChan:
				handler(
					event.Token,
					event.Amount,
					event.Raw.BlockNumber,
				)
			}
		}
	}()

	sub := wftms.Pipe(eventChan)
	return subscription.NewEventSubscription(func() {
		sub.Unsubscribe()
		cancelCtx()
	})
}

func (wftms *PWithdrawnForTbtcMigrationSubscription) Pipe(
	sink chan *abi.PortalWithdrawnForTbtcMigration,
) subscription.EventSubscription {
	ctx, cancelCtx := context.WithCancel(context.Background())
	go func() {
		ticker := time.NewTicker(wftms.opts.Tick)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				lastBlock, err := wftms.contract.blockCounter.CurrentBlock()
				if err != nil {
					pLogger.Errorf(
						"subscription failed to pull events: [%v]",
						err,
					)
				}
				fromBlock := lastBlock - wftms.opts.PastBlocks

				pLogger.Infof(
					"subscription monitoring fetching past WithdrawnForTbtcMigration events "+
						"starting from block [%v]",
					fromBlock,
				)
				events, err := wftms.contract.PastWithdrawnForTbtcMigrationEvents(
					fromBlock,
					nil,
					wftms.tokenFilter,
				)
				if err != nil {
					pLogger.Errorf(
						"subscription failed to pull events: [%v]",
						err,
					)
					continue
				}
				pLogger.Infof(
					"subscription monitoring fetched [%v] past WithdrawnForTbtcMigration events",
					len(events),
				)

				for _, event := range events {
					sink <- event
				}
			}
		}
	}()

	sub := wftms.contract.watchWithdrawnForTbtcMigration(
		sink,
		wftms.tokenFilter,
	)

	return subscription.NewEventSubscription(func() {
		sub.Unsubscribe()
		cancelCtx()
	})
}

func (p *Portal) watchWithdrawnForTbtcMigration(
	sink chan *abi.PortalWithdrawnForTbtcMigration,
	tokenFilter []common.Address,
) event.Subscription {
	subscribeFn := func(ctx context.Context) (event.Subscription, error) {
		return p.contract.WatchWithdrawnForTbtcMigration(
			&bind.WatchOpts{Context: ctx},
			sink,
			tokenFilter,
		)
	}

	thresholdViolatedFn := func(elapsed time.Duration) {
		pLogger.Warnf(
			"subscription to event WithdrawnForTbtcMigration had to be "+
				"retried [%s] since the last attempt; please inspect "+
				"host chain connectivity",
			elapsed,
		)
	}

	subscriptionFailedFn := func(err error) {
		pLogger.Errorf(
			"subscription to event WithdrawnForTbtcMigration failed "+
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

func (p *Portal) PastWithdrawnForTbtcMigrationEvents(
	startBlock uint64,
	endBlock *uint64,
	tokenFilter []common.Address,
) ([]*abi.PortalWithdrawnForTbtcMigration, error) {
	iterator, err := p.contract.FilterWithdrawnForTbtcMigration(
		&bind.FilterOpts{
			Start: startBlock,
			End:   endBlock,
		},
		tokenFilter,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"error retrieving past WithdrawnForTbtcMigration events: [%v]",
			err,
		)
	}

	events := make([]*abi.PortalWithdrawnForTbtcMigration, 0)

	for iterator.Next() {
		event := iterator.Event
		events = append(events, event)
	}

	return events, nil
}

func (p *Portal) WithdrawnTbtcMigratedEvent(
	opts *ethereum.SubscribeOpts,
	depositorFilter []common.Address,
	tokenFilter []common.Address,
	depositIdFilter []*big.Int,
) *PWithdrawnTbtcMigratedSubscription {
	if opts == nil {
		opts = new(ethereum.SubscribeOpts)
	}
	if opts.Tick == 0 {
		opts.Tick = chainutil.DefaultSubscribeOptsTick
	}
	if opts.PastBlocks == 0 {
		opts.PastBlocks = chainutil.DefaultSubscribeOptsPastBlocks
	}

	return &PWithdrawnTbtcMigratedSubscription{
		p,
		opts,
		depositorFilter,
		tokenFilter,
		depositIdFilter,
	}
}

type PWithdrawnTbtcMigratedSubscription struct {
	contract        *Portal
	opts            *ethereum.SubscribeOpts
	depositorFilter []common.Address
	tokenFilter     []common.Address
	depositIdFilter []*big.Int
}

type portalWithdrawnTbtcMigratedFunc func(
	Depositor common.Address,
	Token common.Address,
	TbtcToken common.Address,
	DepositId *big.Int,
	AmountInTbtc *big.Int,
	blockNumber uint64,
)

func (wtms *PWithdrawnTbtcMigratedSubscription) OnEvent(
	handler portalWithdrawnTbtcMigratedFunc,
) subscription.EventSubscription {
	eventChan := make(chan *abi.PortalWithdrawnTbtcMigrated)
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
					event.TbtcToken,
					event.DepositId,
					event.AmountInTbtc,
					event.Raw.BlockNumber,
				)
			}
		}
	}()

	sub := wtms.Pipe(eventChan)
	return subscription.NewEventSubscription(func() {
		sub.Unsubscribe()
		cancelCtx()
	})
}

func (wtms *PWithdrawnTbtcMigratedSubscription) Pipe(
	sink chan *abi.PortalWithdrawnTbtcMigrated,
) subscription.EventSubscription {
	ctx, cancelCtx := context.WithCancel(context.Background())
	go func() {
		ticker := time.NewTicker(wtms.opts.Tick)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				lastBlock, err := wtms.contract.blockCounter.CurrentBlock()
				if err != nil {
					pLogger.Errorf(
						"subscription failed to pull events: [%v]",
						err,
					)
				}
				fromBlock := lastBlock - wtms.opts.PastBlocks

				pLogger.Infof(
					"subscription monitoring fetching past WithdrawnTbtcMigrated events "+
						"starting from block [%v]",
					fromBlock,
				)
				events, err := wtms.contract.PastWithdrawnTbtcMigratedEvents(
					fromBlock,
					nil,
					wtms.depositorFilter,
					wtms.tokenFilter,
					wtms.depositIdFilter,
				)
				if err != nil {
					pLogger.Errorf(
						"subscription failed to pull events: [%v]",
						err,
					)
					continue
				}
				pLogger.Infof(
					"subscription monitoring fetched [%v] past WithdrawnTbtcMigrated events",
					len(events),
				)

				for _, event := range events {
					sink <- event
				}
			}
		}
	}()

	sub := wtms.contract.watchWithdrawnTbtcMigrated(
		sink,
		wtms.depositorFilter,
		wtms.tokenFilter,
		wtms.depositIdFilter,
	)

	return subscription.NewEventSubscription(func() {
		sub.Unsubscribe()
		cancelCtx()
	})
}

func (p *Portal) watchWithdrawnTbtcMigrated(
	sink chan *abi.PortalWithdrawnTbtcMigrated,
	depositorFilter []common.Address,
	tokenFilter []common.Address,
	depositIdFilter []*big.Int,
) event.Subscription {
	subscribeFn := func(ctx context.Context) (event.Subscription, error) {
		return p.contract.WatchWithdrawnTbtcMigrated(
			&bind.WatchOpts{Context: ctx},
			sink,
			depositorFilter,
			tokenFilter,
			depositIdFilter,
		)
	}

	thresholdViolatedFn := func(elapsed time.Duration) {
		pLogger.Warnf(
			"subscription to event WithdrawnTbtcMigrated had to be "+
				"retried [%s] since the last attempt; please inspect "+
				"host chain connectivity",
			elapsed,
		)
	}

	subscriptionFailedFn := func(err error) {
		pLogger.Errorf(
			"subscription to event WithdrawnTbtcMigrated failed "+
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

func (p *Portal) PastWithdrawnTbtcMigratedEvents(
	startBlock uint64,
	endBlock *uint64,
	depositorFilter []common.Address,
	tokenFilter []common.Address,
	depositIdFilter []*big.Int,
) ([]*abi.PortalWithdrawnTbtcMigrated, error) {
	iterator, err := p.contract.FilterWithdrawnTbtcMigrated(
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
			"error retrieving past WithdrawnTbtcMigrated events: [%v]",
			err,
		)
	}

	events := make([]*abi.PortalWithdrawnTbtcMigrated, 0)

	for iterator.Next() {
		event := iterator.Event
		events = append(events, event)
	}

	return events, nil
}
