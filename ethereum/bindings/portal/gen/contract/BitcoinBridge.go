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
var bbLogger = log.Logger("keep-contract-BitcoinBridge")

type BitcoinBridge struct {
	contract          *abi.BitcoinBridge
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

func NewBitcoinBridge(
	contractAddress common.Address,
	chainId *big.Int,
	accountKey *keystore.Key,
	backend bind.ContractBackend,
	nonceManager *ethereum.NonceManager,
	miningWaiter *chainutil.MiningWaiter,
	blockCounter *ethereum.BlockCounter,
	transactionMutex *sync.Mutex,
) (*BitcoinBridge, error) {
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

	contract, err := abi.NewBitcoinBridge(
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

	contractABI, err := hostchainabi.JSON(strings.NewReader(abi.BitcoinBridgeABI))
	if err != nil {
		return nil, fmt.Errorf("failed to instantiate ABI: [%v]", err)
	}

	return &BitcoinBridge{
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
func (bb *BitcoinBridge) AcceptOwnership(

	transactionOptions ...chainutil.TransactionOptions,
) (*types.Transaction, error) {
	bbLogger.Debug(
		"submitting transaction acceptOwnership",
	)

	bb.transactionMutex.Lock()
	defer bb.transactionMutex.Unlock()

	// create a copy
	transactorOptions := new(bind.TransactOpts)
	*transactorOptions = *bb.transactorOptions

	if len(transactionOptions) > 1 {
		return nil, fmt.Errorf(
			"could not process multiple transaction options sets",
		)
	} else if len(transactionOptions) > 0 {
		transactionOptions[0].Apply(transactorOptions)
	}

	nonce, err := bb.nonceManager.CurrentNonce()
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve account nonce: %v", err)
	}

	transactorOptions.Nonce = new(big.Int).SetUint64(nonce)

	transaction, err := bb.contract.AcceptOwnership(
		transactorOptions,
	)
	if err != nil {
		return transaction, bb.errorResolver.ResolveError(
			err,
			bb.transactorOptions.From,
			nil,
			"acceptOwnership",
		)
	}

	bbLogger.Infof(
		"submitted transaction acceptOwnership with id: [%s] and nonce [%v]",
		transaction.Hash(),
		transaction.Nonce(),
	)

	go bb.miningWaiter.ForceMining(
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

			transaction, err := bb.contract.AcceptOwnership(
				newTransactorOptions,
			)
			if err != nil {
				return nil, bb.errorResolver.ResolveError(
					err,
					bb.transactorOptions.From,
					nil,
					"acceptOwnership",
				)
			}

			bbLogger.Infof(
				"submitted transaction acceptOwnership with id: [%s] and nonce [%v]",
				transaction.Hash(),
				transaction.Nonce(),
			)

			return transaction, nil
		},
	)

	bb.nonceManager.IncrementNonce()

	return transaction, err
}

// Non-mutating call, not a transaction submission.
func (bb *BitcoinBridge) CallAcceptOwnership(
	blockNumber *big.Int,
) error {
	var result interface{} = nil

	err := chainutil.CallAtBlock(
		bb.transactorOptions.From,
		blockNumber, nil,
		bb.contractABI,
		bb.caller,
		bb.errorResolver,
		bb.contractAddress,
		"acceptOwnership",
		&result,
	)

	return err
}

func (bb *BitcoinBridge) AcceptOwnershipGasEstimate() (uint64, error) {
	var result uint64

	result, err := chainutil.EstimateGas(
		bb.callerOptions.From,
		bb.contractAddress,
		"acceptOwnership",
		bb.contractABI,
		bb.transactor,
	)

	return result, err
}

// Transaction submission.
func (bb *BitcoinBridge) BridgeTBTC(
	arg_amount *big.Int,
	arg_recipient common.Address,

	transactionOptions ...chainutil.TransactionOptions,
) (*types.Transaction, error) {
	bbLogger.Debug(
		"submitting transaction bridgeTBTC",
		" params: ",
		fmt.Sprint(
			arg_amount,
			arg_recipient,
		),
	)

	bb.transactionMutex.Lock()
	defer bb.transactionMutex.Unlock()

	// create a copy
	transactorOptions := new(bind.TransactOpts)
	*transactorOptions = *bb.transactorOptions

	if len(transactionOptions) > 1 {
		return nil, fmt.Errorf(
			"could not process multiple transaction options sets",
		)
	} else if len(transactionOptions) > 0 {
		transactionOptions[0].Apply(transactorOptions)
	}

	nonce, err := bb.nonceManager.CurrentNonce()
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve account nonce: %v", err)
	}

	transactorOptions.Nonce = new(big.Int).SetUint64(nonce)

	transaction, err := bb.contract.BridgeTBTC(
		transactorOptions,
		arg_amount,
		arg_recipient,
	)
	if err != nil {
		return transaction, bb.errorResolver.ResolveError(
			err,
			bb.transactorOptions.From,
			nil,
			"bridgeTBTC",
			arg_amount,
			arg_recipient,
		)
	}

	bbLogger.Infof(
		"submitted transaction bridgeTBTC with id: [%s] and nonce [%v]",
		transaction.Hash(),
		transaction.Nonce(),
	)

	go bb.miningWaiter.ForceMining(
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

			transaction, err := bb.contract.BridgeTBTC(
				newTransactorOptions,
				arg_amount,
				arg_recipient,
			)
			if err != nil {
				return nil, bb.errorResolver.ResolveError(
					err,
					bb.transactorOptions.From,
					nil,
					"bridgeTBTC",
					arg_amount,
					arg_recipient,
				)
			}

			bbLogger.Infof(
				"submitted transaction bridgeTBTC with id: [%s] and nonce [%v]",
				transaction.Hash(),
				transaction.Nonce(),
			)

			return transaction, nil
		},
	)

	bb.nonceManager.IncrementNonce()

	return transaction, err
}

// Non-mutating call, not a transaction submission.
func (bb *BitcoinBridge) CallBridgeTBTC(
	arg_amount *big.Int,
	arg_recipient common.Address,
	blockNumber *big.Int,
) error {
	var result interface{} = nil

	err := chainutil.CallAtBlock(
		bb.transactorOptions.From,
		blockNumber, nil,
		bb.contractABI,
		bb.caller,
		bb.errorResolver,
		bb.contractAddress,
		"bridgeTBTC",
		&result,
		arg_amount,
		arg_recipient,
	)

	return err
}

func (bb *BitcoinBridge) BridgeTBTCGasEstimate(
	arg_amount *big.Int,
	arg_recipient common.Address,
) (uint64, error) {
	var result uint64

	result, err := chainutil.EstimateGas(
		bb.callerOptions.From,
		bb.contractAddress,
		"bridgeTBTC",
		bb.contractABI,
		bb.transactor,
		arg_amount,
		arg_recipient,
	)

	return result, err
}

// Transaction submission.
func (bb *BitcoinBridge) BridgeTBTCWithPermit(
	arg_amount *big.Int,
	arg_recipient common.Address,
	arg_deadline *big.Int,
	arg_v uint8,
	arg_r [32]byte,
	arg_s [32]byte,

	transactionOptions ...chainutil.TransactionOptions,
) (*types.Transaction, error) {
	bbLogger.Debug(
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

	bb.transactionMutex.Lock()
	defer bb.transactionMutex.Unlock()

	// create a copy
	transactorOptions := new(bind.TransactOpts)
	*transactorOptions = *bb.transactorOptions

	if len(transactionOptions) > 1 {
		return nil, fmt.Errorf(
			"could not process multiple transaction options sets",
		)
	} else if len(transactionOptions) > 0 {
		transactionOptions[0].Apply(transactorOptions)
	}

	nonce, err := bb.nonceManager.CurrentNonce()
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve account nonce: %v", err)
	}

	transactorOptions.Nonce = new(big.Int).SetUint64(nonce)

	transaction, err := bb.contract.BridgeTBTCWithPermit(
		transactorOptions,
		arg_amount,
		arg_recipient,
		arg_deadline,
		arg_v,
		arg_r,
		arg_s,
	)
	if err != nil {
		return transaction, bb.errorResolver.ResolveError(
			err,
			bb.transactorOptions.From,
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

	bbLogger.Infof(
		"submitted transaction bridgeTBTCWithPermit with id: [%s] and nonce [%v]",
		transaction.Hash(),
		transaction.Nonce(),
	)

	go bb.miningWaiter.ForceMining(
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

			transaction, err := bb.contract.BridgeTBTCWithPermit(
				newTransactorOptions,
				arg_amount,
				arg_recipient,
				arg_deadline,
				arg_v,
				arg_r,
				arg_s,
			)
			if err != nil {
				return nil, bb.errorResolver.ResolveError(
					err,
					bb.transactorOptions.From,
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

			bbLogger.Infof(
				"submitted transaction bridgeTBTCWithPermit with id: [%s] and nonce [%v]",
				transaction.Hash(),
				transaction.Nonce(),
			)

			return transaction, nil
		},
	)

	bb.nonceManager.IncrementNonce()

	return transaction, err
}

// Non-mutating call, not a transaction submission.
func (bb *BitcoinBridge) CallBridgeTBTCWithPermit(
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
		bb.transactorOptions.From,
		blockNumber, nil,
		bb.contractABI,
		bb.caller,
		bb.errorResolver,
		bb.contractAddress,
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

func (bb *BitcoinBridge) BridgeTBTCWithPermitGasEstimate(
	arg_amount *big.Int,
	arg_recipient common.Address,
	arg_deadline *big.Int,
	arg_v uint8,
	arg_r [32]byte,
	arg_s [32]byte,
) (uint64, error) {
	var result uint64

	result, err := chainutil.EstimateGas(
		bb.callerOptions.From,
		bb.contractAddress,
		"bridgeTBTCWithPermit",
		bb.contractABI,
		bb.transactor,
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
func (bb *BitcoinBridge) FinalizeBTCBridging(
	arg_depositKey *big.Int,
	arg_recipient common.Address,

	transactionOptions ...chainutil.TransactionOptions,
) (*types.Transaction, error) {
	bbLogger.Debug(
		"submitting transaction finalizeBTCBridging",
		" params: ",
		fmt.Sprint(
			arg_depositKey,
			arg_recipient,
		),
	)

	bb.transactionMutex.Lock()
	defer bb.transactionMutex.Unlock()

	// create a copy
	transactorOptions := new(bind.TransactOpts)
	*transactorOptions = *bb.transactorOptions

	if len(transactionOptions) > 1 {
		return nil, fmt.Errorf(
			"could not process multiple transaction options sets",
		)
	} else if len(transactionOptions) > 0 {
		transactionOptions[0].Apply(transactorOptions)
	}

	nonce, err := bb.nonceManager.CurrentNonce()
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve account nonce: %v", err)
	}

	transactorOptions.Nonce = new(big.Int).SetUint64(nonce)

	transaction, err := bb.contract.FinalizeBTCBridging(
		transactorOptions,
		arg_depositKey,
		arg_recipient,
	)
	if err != nil {
		return transaction, bb.errorResolver.ResolveError(
			err,
			bb.transactorOptions.From,
			nil,
			"finalizeBTCBridging",
			arg_depositKey,
			arg_recipient,
		)
	}

	bbLogger.Infof(
		"submitted transaction finalizeBTCBridging with id: [%s] and nonce [%v]",
		transaction.Hash(),
		transaction.Nonce(),
	)

	go bb.miningWaiter.ForceMining(
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

			transaction, err := bb.contract.FinalizeBTCBridging(
				newTransactorOptions,
				arg_depositKey,
				arg_recipient,
			)
			if err != nil {
				return nil, bb.errorResolver.ResolveError(
					err,
					bb.transactorOptions.From,
					nil,
					"finalizeBTCBridging",
					arg_depositKey,
					arg_recipient,
				)
			}

			bbLogger.Infof(
				"submitted transaction finalizeBTCBridging with id: [%s] and nonce [%v]",
				transaction.Hash(),
				transaction.Nonce(),
			)

			return transaction, nil
		},
	)

	bb.nonceManager.IncrementNonce()

	return transaction, err
}

// Non-mutating call, not a transaction submission.
func (bb *BitcoinBridge) CallFinalizeBTCBridging(
	arg_depositKey *big.Int,
	arg_recipient common.Address,
	blockNumber *big.Int,
) error {
	var result interface{} = nil

	err := chainutil.CallAtBlock(
		bb.transactorOptions.From,
		blockNumber, nil,
		bb.contractABI,
		bb.caller,
		bb.errorResolver,
		bb.contractAddress,
		"finalizeBTCBridging",
		&result,
		arg_depositKey,
		arg_recipient,
	)

	return err
}

func (bb *BitcoinBridge) FinalizeBTCBridgingGasEstimate(
	arg_depositKey *big.Int,
	arg_recipient common.Address,
) (uint64, error) {
	var result uint64

	result, err := chainutil.EstimateGas(
		bb.callerOptions.From,
		bb.contractAddress,
		"finalizeBTCBridging",
		bb.contractABI,
		bb.transactor,
		arg_depositKey,
		arg_recipient,
	)

	return result, err
}

// Transaction submission.
func (bb *BitcoinBridge) Initialize(
	arg__bridge common.Address,
	arg__tbtcVault common.Address,
	arg__tbtcToken common.Address,

	transactionOptions ...chainutil.TransactionOptions,
) (*types.Transaction, error) {
	bbLogger.Debug(
		"submitting transaction initialize",
		" params: ",
		fmt.Sprint(
			arg__bridge,
			arg__tbtcVault,
			arg__tbtcToken,
		),
	)

	bb.transactionMutex.Lock()
	defer bb.transactionMutex.Unlock()

	// create a copy
	transactorOptions := new(bind.TransactOpts)
	*transactorOptions = *bb.transactorOptions

	if len(transactionOptions) > 1 {
		return nil, fmt.Errorf(
			"could not process multiple transaction options sets",
		)
	} else if len(transactionOptions) > 0 {
		transactionOptions[0].Apply(transactorOptions)
	}

	nonce, err := bb.nonceManager.CurrentNonce()
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve account nonce: %v", err)
	}

	transactorOptions.Nonce = new(big.Int).SetUint64(nonce)

	transaction, err := bb.contract.Initialize(
		transactorOptions,
		arg__bridge,
		arg__tbtcVault,
		arg__tbtcToken,
	)
	if err != nil {
		return transaction, bb.errorResolver.ResolveError(
			err,
			bb.transactorOptions.From,
			nil,
			"initialize",
			arg__bridge,
			arg__tbtcVault,
			arg__tbtcToken,
		)
	}

	bbLogger.Infof(
		"submitted transaction initialize with id: [%s] and nonce [%v]",
		transaction.Hash(),
		transaction.Nonce(),
	)

	go bb.miningWaiter.ForceMining(
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

			transaction, err := bb.contract.Initialize(
				newTransactorOptions,
				arg__bridge,
				arg__tbtcVault,
				arg__tbtcToken,
			)
			if err != nil {
				return nil, bb.errorResolver.ResolveError(
					err,
					bb.transactorOptions.From,
					nil,
					"initialize",
					arg__bridge,
					arg__tbtcVault,
					arg__tbtcToken,
				)
			}

			bbLogger.Infof(
				"submitted transaction initialize with id: [%s] and nonce [%v]",
				transaction.Hash(),
				transaction.Nonce(),
			)

			return transaction, nil
		},
	)

	bb.nonceManager.IncrementNonce()

	return transaction, err
}

// Non-mutating call, not a transaction submission.
func (bb *BitcoinBridge) CallInitialize(
	arg__bridge common.Address,
	arg__tbtcVault common.Address,
	arg__tbtcToken common.Address,
	blockNumber *big.Int,
) error {
	var result interface{} = nil

	err := chainutil.CallAtBlock(
		bb.transactorOptions.From,
		blockNumber, nil,
		bb.contractABI,
		bb.caller,
		bb.errorResolver,
		bb.contractAddress,
		"initialize",
		&result,
		arg__bridge,
		arg__tbtcVault,
		arg__tbtcToken,
	)

	return err
}

func (bb *BitcoinBridge) InitializeGasEstimate(
	arg__bridge common.Address,
	arg__tbtcVault common.Address,
	arg__tbtcToken common.Address,
) (uint64, error) {
	var result uint64

	result, err := chainutil.EstimateGas(
		bb.callerOptions.From,
		bb.contractAddress,
		"initialize",
		bb.contractABI,
		bb.transactor,
		arg__bridge,
		arg__tbtcVault,
		arg__tbtcToken,
	)

	return result, err
}

// Transaction submission.
func (bb *BitcoinBridge) InitializeBTCBridging(
	arg_fundingTx abi.IBridgeTypesBitcoinTxInfo,
	arg_reveal abi.IBridgeTypesDepositRevealInfo,
	arg_recipient common.Address,

	transactionOptions ...chainutil.TransactionOptions,
) (*types.Transaction, error) {
	bbLogger.Debug(
		"submitting transaction initializeBTCBridging",
		" params: ",
		fmt.Sprint(
			arg_fundingTx,
			arg_reveal,
			arg_recipient,
		),
	)

	bb.transactionMutex.Lock()
	defer bb.transactionMutex.Unlock()

	// create a copy
	transactorOptions := new(bind.TransactOpts)
	*transactorOptions = *bb.transactorOptions

	if len(transactionOptions) > 1 {
		return nil, fmt.Errorf(
			"could not process multiple transaction options sets",
		)
	} else if len(transactionOptions) > 0 {
		transactionOptions[0].Apply(transactorOptions)
	}

	nonce, err := bb.nonceManager.CurrentNonce()
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve account nonce: %v", err)
	}

	transactorOptions.Nonce = new(big.Int).SetUint64(nonce)

	transaction, err := bb.contract.InitializeBTCBridging(
		transactorOptions,
		arg_fundingTx,
		arg_reveal,
		arg_recipient,
	)
	if err != nil {
		return transaction, bb.errorResolver.ResolveError(
			err,
			bb.transactorOptions.From,
			nil,
			"initializeBTCBridging",
			arg_fundingTx,
			arg_reveal,
			arg_recipient,
		)
	}

	bbLogger.Infof(
		"submitted transaction initializeBTCBridging with id: [%s] and nonce [%v]",
		transaction.Hash(),
		transaction.Nonce(),
	)

	go bb.miningWaiter.ForceMining(
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

			transaction, err := bb.contract.InitializeBTCBridging(
				newTransactorOptions,
				arg_fundingTx,
				arg_reveal,
				arg_recipient,
			)
			if err != nil {
				return nil, bb.errorResolver.ResolveError(
					err,
					bb.transactorOptions.From,
					nil,
					"initializeBTCBridging",
					arg_fundingTx,
					arg_reveal,
					arg_recipient,
				)
			}

			bbLogger.Infof(
				"submitted transaction initializeBTCBridging with id: [%s] and nonce [%v]",
				transaction.Hash(),
				transaction.Nonce(),
			)

			return transaction, nil
		},
	)

	bb.nonceManager.IncrementNonce()

	return transaction, err
}

// Non-mutating call, not a transaction submission.
func (bb *BitcoinBridge) CallInitializeBTCBridging(
	arg_fundingTx abi.IBridgeTypesBitcoinTxInfo,
	arg_reveal abi.IBridgeTypesDepositRevealInfo,
	arg_recipient common.Address,
	blockNumber *big.Int,
) error {
	var result interface{} = nil

	err := chainutil.CallAtBlock(
		bb.transactorOptions.From,
		blockNumber, nil,
		bb.contractABI,
		bb.caller,
		bb.errorResolver,
		bb.contractAddress,
		"initializeBTCBridging",
		&result,
		arg_fundingTx,
		arg_reveal,
		arg_recipient,
	)

	return err
}

func (bb *BitcoinBridge) InitializeBTCBridgingGasEstimate(
	arg_fundingTx abi.IBridgeTypesBitcoinTxInfo,
	arg_reveal abi.IBridgeTypesDepositRevealInfo,
	arg_recipient common.Address,
) (uint64, error) {
	var result uint64

	result, err := chainutil.EstimateGas(
		bb.callerOptions.From,
		bb.contractAddress,
		"initializeBTCBridging",
		bb.contractABI,
		bb.transactor,
		arg_fundingTx,
		arg_reveal,
		arg_recipient,
	)

	return result, err
}

// Transaction submission.
func (bb *BitcoinBridge) RenounceOwnership(

	transactionOptions ...chainutil.TransactionOptions,
) (*types.Transaction, error) {
	bbLogger.Debug(
		"submitting transaction renounceOwnership",
	)

	bb.transactionMutex.Lock()
	defer bb.transactionMutex.Unlock()

	// create a copy
	transactorOptions := new(bind.TransactOpts)
	*transactorOptions = *bb.transactorOptions

	if len(transactionOptions) > 1 {
		return nil, fmt.Errorf(
			"could not process multiple transaction options sets",
		)
	} else if len(transactionOptions) > 0 {
		transactionOptions[0].Apply(transactorOptions)
	}

	nonce, err := bb.nonceManager.CurrentNonce()
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve account nonce: %v", err)
	}

	transactorOptions.Nonce = new(big.Int).SetUint64(nonce)

	transaction, err := bb.contract.RenounceOwnership(
		transactorOptions,
	)
	if err != nil {
		return transaction, bb.errorResolver.ResolveError(
			err,
			bb.transactorOptions.From,
			nil,
			"renounceOwnership",
		)
	}

	bbLogger.Infof(
		"submitted transaction renounceOwnership with id: [%s] and nonce [%v]",
		transaction.Hash(),
		transaction.Nonce(),
	)

	go bb.miningWaiter.ForceMining(
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

			transaction, err := bb.contract.RenounceOwnership(
				newTransactorOptions,
			)
			if err != nil {
				return nil, bb.errorResolver.ResolveError(
					err,
					bb.transactorOptions.From,
					nil,
					"renounceOwnership",
				)
			}

			bbLogger.Infof(
				"submitted transaction renounceOwnership with id: [%s] and nonce [%v]",
				transaction.Hash(),
				transaction.Nonce(),
			)

			return transaction, nil
		},
	)

	bb.nonceManager.IncrementNonce()

	return transaction, err
}

// Non-mutating call, not a transaction submission.
func (bb *BitcoinBridge) CallRenounceOwnership(
	blockNumber *big.Int,
) error {
	var result interface{} = nil

	err := chainutil.CallAtBlock(
		bb.transactorOptions.From,
		blockNumber, nil,
		bb.contractABI,
		bb.caller,
		bb.errorResolver,
		bb.contractAddress,
		"renounceOwnership",
		&result,
	)

	return err
}

func (bb *BitcoinBridge) RenounceOwnershipGasEstimate() (uint64, error) {
	var result uint64

	result, err := chainutil.EstimateGas(
		bb.callerOptions.From,
		bb.contractAddress,
		"renounceOwnership",
		bb.contractABI,
		bb.transactor,
	)

	return result, err
}

// Transaction submission.
func (bb *BitcoinBridge) TransferOwnership(
	arg_newOwner common.Address,

	transactionOptions ...chainutil.TransactionOptions,
) (*types.Transaction, error) {
	bbLogger.Debug(
		"submitting transaction transferOwnership",
		" params: ",
		fmt.Sprint(
			arg_newOwner,
		),
	)

	bb.transactionMutex.Lock()
	defer bb.transactionMutex.Unlock()

	// create a copy
	transactorOptions := new(bind.TransactOpts)
	*transactorOptions = *bb.transactorOptions

	if len(transactionOptions) > 1 {
		return nil, fmt.Errorf(
			"could not process multiple transaction options sets",
		)
	} else if len(transactionOptions) > 0 {
		transactionOptions[0].Apply(transactorOptions)
	}

	nonce, err := bb.nonceManager.CurrentNonce()
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve account nonce: %v", err)
	}

	transactorOptions.Nonce = new(big.Int).SetUint64(nonce)

	transaction, err := bb.contract.TransferOwnership(
		transactorOptions,
		arg_newOwner,
	)
	if err != nil {
		return transaction, bb.errorResolver.ResolveError(
			err,
			bb.transactorOptions.From,
			nil,
			"transferOwnership",
			arg_newOwner,
		)
	}

	bbLogger.Infof(
		"submitted transaction transferOwnership with id: [%s] and nonce [%v]",
		transaction.Hash(),
		transaction.Nonce(),
	)

	go bb.miningWaiter.ForceMining(
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

			transaction, err := bb.contract.TransferOwnership(
				newTransactorOptions,
				arg_newOwner,
			)
			if err != nil {
				return nil, bb.errorResolver.ResolveError(
					err,
					bb.transactorOptions.From,
					nil,
					"transferOwnership",
					arg_newOwner,
				)
			}

			bbLogger.Infof(
				"submitted transaction transferOwnership with id: [%s] and nonce [%v]",
				transaction.Hash(),
				transaction.Nonce(),
			)

			return transaction, nil
		},
	)

	bb.nonceManager.IncrementNonce()

	return transaction, err
}

// Non-mutating call, not a transaction submission.
func (bb *BitcoinBridge) CallTransferOwnership(
	arg_newOwner common.Address,
	blockNumber *big.Int,
) error {
	var result interface{} = nil

	err := chainutil.CallAtBlock(
		bb.transactorOptions.From,
		blockNumber, nil,
		bb.contractABI,
		bb.caller,
		bb.errorResolver,
		bb.contractAddress,
		"transferOwnership",
		&result,
		arg_newOwner,
	)

	return err
}

func (bb *BitcoinBridge) TransferOwnershipGasEstimate(
	arg_newOwner common.Address,
) (uint64, error) {
	var result uint64

	result, err := chainutil.EstimateGas(
		bb.callerOptions.From,
		bb.contractAddress,
		"transferOwnership",
		bb.contractABI,
		bb.transactor,
		arg_newOwner,
	)

	return result, err
}

// Transaction submission.
func (bb *BitcoinBridge) UpdateMinTBTCAmount(
	arg_newMinTBTCAmount *big.Int,

	transactionOptions ...chainutil.TransactionOptions,
) (*types.Transaction, error) {
	bbLogger.Debug(
		"submitting transaction updateMinTBTCAmount",
		" params: ",
		fmt.Sprint(
			arg_newMinTBTCAmount,
		),
	)

	bb.transactionMutex.Lock()
	defer bb.transactionMutex.Unlock()

	// create a copy
	transactorOptions := new(bind.TransactOpts)
	*transactorOptions = *bb.transactorOptions

	if len(transactionOptions) > 1 {
		return nil, fmt.Errorf(
			"could not process multiple transaction options sets",
		)
	} else if len(transactionOptions) > 0 {
		transactionOptions[0].Apply(transactorOptions)
	}

	nonce, err := bb.nonceManager.CurrentNonce()
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve account nonce: %v", err)
	}

	transactorOptions.Nonce = new(big.Int).SetUint64(nonce)

	transaction, err := bb.contract.UpdateMinTBTCAmount(
		transactorOptions,
		arg_newMinTBTCAmount,
	)
	if err != nil {
		return transaction, bb.errorResolver.ResolveError(
			err,
			bb.transactorOptions.From,
			nil,
			"updateMinTBTCAmount",
			arg_newMinTBTCAmount,
		)
	}

	bbLogger.Infof(
		"submitted transaction updateMinTBTCAmount with id: [%s] and nonce [%v]",
		transaction.Hash(),
		transaction.Nonce(),
	)

	go bb.miningWaiter.ForceMining(
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

			transaction, err := bb.contract.UpdateMinTBTCAmount(
				newTransactorOptions,
				arg_newMinTBTCAmount,
			)
			if err != nil {
				return nil, bb.errorResolver.ResolveError(
					err,
					bb.transactorOptions.From,
					nil,
					"updateMinTBTCAmount",
					arg_newMinTBTCAmount,
				)
			}

			bbLogger.Infof(
				"submitted transaction updateMinTBTCAmount with id: [%s] and nonce [%v]",
				transaction.Hash(),
				transaction.Nonce(),
			)

			return transaction, nil
		},
	)

	bb.nonceManager.IncrementNonce()

	return transaction, err
}

// Non-mutating call, not a transaction submission.
func (bb *BitcoinBridge) CallUpdateMinTBTCAmount(
	arg_newMinTBTCAmount *big.Int,
	blockNumber *big.Int,
) error {
	var result interface{} = nil

	err := chainutil.CallAtBlock(
		bb.transactorOptions.From,
		blockNumber, nil,
		bb.contractABI,
		bb.caller,
		bb.errorResolver,
		bb.contractAddress,
		"updateMinTBTCAmount",
		&result,
		arg_newMinTBTCAmount,
	)

	return err
}

func (bb *BitcoinBridge) UpdateMinTBTCAmountGasEstimate(
	arg_newMinTBTCAmount *big.Int,
) (uint64, error) {
	var result uint64

	result, err := chainutil.EstimateGas(
		bb.callerOptions.From,
		bb.contractAddress,
		"updateMinTBTCAmount",
		bb.contractABI,
		bb.transactor,
		arg_newMinTBTCAmount,
	)

	return result, err
}

// ----- Const Methods ------

func (bb *BitcoinBridge) Bridge() (common.Address, error) {
	result, err := bb.contract.Bridge(
		bb.callerOptions,
	)

	if err != nil {
		return result, bb.errorResolver.ResolveError(
			err,
			bb.callerOptions.From,
			nil,
			"bridge",
		)
	}

	return result, err
}

func (bb *BitcoinBridge) BridgeAtBlock(
	blockNumber *big.Int,
) (common.Address, error) {
	var result common.Address

	err := chainutil.CallAtBlock(
		bb.callerOptions.From,
		blockNumber,
		nil,
		bb.contractABI,
		bb.caller,
		bb.errorResolver,
		bb.contractAddress,
		"bridge",
		&result,
	)

	return result, err
}

func (bb *BitcoinBridge) Deposits(
	arg0 *big.Int,
) (uint8, error) {
	result, err := bb.contract.Deposits(
		bb.callerOptions,
		arg0,
	)

	if err != nil {
		return result, bb.errorResolver.ResolveError(
			err,
			bb.callerOptions.From,
			nil,
			"deposits",
			arg0,
		)
	}

	return result, err
}

func (bb *BitcoinBridge) DepositsAtBlock(
	arg0 *big.Int,
	blockNumber *big.Int,
) (uint8, error) {
	var result uint8

	err := chainutil.CallAtBlock(
		bb.callerOptions.From,
		blockNumber,
		nil,
		bb.contractABI,
		bb.caller,
		bb.errorResolver,
		bb.contractAddress,
		"deposits",
		&result,
		arg0,
	)

	return result, err
}

func (bb *BitcoinBridge) MinTBTCAmount() (*big.Int, error) {
	result, err := bb.contract.MinTBTCAmount(
		bb.callerOptions,
	)

	if err != nil {
		return result, bb.errorResolver.ResolveError(
			err,
			bb.callerOptions.From,
			nil,
			"minTBTCAmount",
		)
	}

	return result, err
}

func (bb *BitcoinBridge) MinTBTCAmountAtBlock(
	blockNumber *big.Int,
) (*big.Int, error) {
	var result *big.Int

	err := chainutil.CallAtBlock(
		bb.callerOptions.From,
		blockNumber,
		nil,
		bb.contractABI,
		bb.caller,
		bb.errorResolver,
		bb.contractAddress,
		"minTBTCAmount",
		&result,
	)

	return result, err
}

func (bb *BitcoinBridge) Owner() (common.Address, error) {
	result, err := bb.contract.Owner(
		bb.callerOptions,
	)

	if err != nil {
		return result, bb.errorResolver.ResolveError(
			err,
			bb.callerOptions.From,
			nil,
			"owner",
		)
	}

	return result, err
}

func (bb *BitcoinBridge) OwnerAtBlock(
	blockNumber *big.Int,
) (common.Address, error) {
	var result common.Address

	err := chainutil.CallAtBlock(
		bb.callerOptions.From,
		blockNumber,
		nil,
		bb.contractABI,
		bb.caller,
		bb.errorResolver,
		bb.contractAddress,
		"owner",
		&result,
	)

	return result, err
}

func (bb *BitcoinBridge) PendingOwner() (common.Address, error) {
	result, err := bb.contract.PendingOwner(
		bb.callerOptions,
	)

	if err != nil {
		return result, bb.errorResolver.ResolveError(
			err,
			bb.callerOptions.From,
			nil,
			"pendingOwner",
		)
	}

	return result, err
}

func (bb *BitcoinBridge) PendingOwnerAtBlock(
	blockNumber *big.Int,
) (common.Address, error) {
	var result common.Address

	err := chainutil.CallAtBlock(
		bb.callerOptions.From,
		blockNumber,
		nil,
		bb.contractABI,
		bb.caller,
		bb.errorResolver,
		bb.contractAddress,
		"pendingOwner",
		&result,
	)

	return result, err
}

func (bb *BitcoinBridge) SATOSHIMULTIPLIER() (*big.Int, error) {
	result, err := bb.contract.SATOSHIMULTIPLIER(
		bb.callerOptions,
	)

	if err != nil {
		return result, bb.errorResolver.ResolveError(
			err,
			bb.callerOptions.From,
			nil,
			"sATOSHIMULTIPLIER",
		)
	}

	return result, err
}

func (bb *BitcoinBridge) SATOSHIMULTIPLIERAtBlock(
	blockNumber *big.Int,
) (*big.Int, error) {
	var result *big.Int

	err := chainutil.CallAtBlock(
		bb.callerOptions.From,
		blockNumber,
		nil,
		bb.contractABI,
		bb.caller,
		bb.errorResolver,
		bb.contractAddress,
		"sATOSHIMULTIPLIER",
		&result,
	)

	return result, err
}

func (bb *BitcoinBridge) Sequence() (*big.Int, error) {
	result, err := bb.contract.Sequence(
		bb.callerOptions,
	)

	if err != nil {
		return result, bb.errorResolver.ResolveError(
			err,
			bb.callerOptions.From,
			nil,
			"sequence",
		)
	}

	return result, err
}

func (bb *BitcoinBridge) SequenceAtBlock(
	blockNumber *big.Int,
) (*big.Int, error) {
	var result *big.Int

	err := chainutil.CallAtBlock(
		bb.callerOptions.From,
		blockNumber,
		nil,
		bb.contractABI,
		bb.caller,
		bb.errorResolver,
		bb.contractAddress,
		"sequence",
		&result,
	)

	return result, err
}

func (bb *BitcoinBridge) TbtcToken() (common.Address, error) {
	result, err := bb.contract.TbtcToken(
		bb.callerOptions,
	)

	if err != nil {
		return result, bb.errorResolver.ResolveError(
			err,
			bb.callerOptions.From,
			nil,
			"tbtcToken",
		)
	}

	return result, err
}

func (bb *BitcoinBridge) TbtcTokenAtBlock(
	blockNumber *big.Int,
) (common.Address, error) {
	var result common.Address

	err := chainutil.CallAtBlock(
		bb.callerOptions.From,
		blockNumber,
		nil,
		bb.contractABI,
		bb.caller,
		bb.errorResolver,
		bb.contractAddress,
		"tbtcToken",
		&result,
	)

	return result, err
}

func (bb *BitcoinBridge) TbtcVault() (common.Address, error) {
	result, err := bb.contract.TbtcVault(
		bb.callerOptions,
	)

	if err != nil {
		return result, bb.errorResolver.ResolveError(
			err,
			bb.callerOptions.From,
			nil,
			"tbtcVault",
		)
	}

	return result, err
}

func (bb *BitcoinBridge) TbtcVaultAtBlock(
	blockNumber *big.Int,
) (common.Address, error) {
	var result common.Address

	err := chainutil.CallAtBlock(
		bb.callerOptions.From,
		blockNumber,
		nil,
		bb.contractABI,
		bb.caller,
		bb.errorResolver,
		bb.contractAddress,
		"tbtcVault",
		&result,
	)

	return result, err
}

// ------ Events -------

func (bb *BitcoinBridge) AssetsLockedEvent(
	opts *ethereum.SubscribeOpts,
	sequenceNumberFilter []*big.Int,
	recipientFilter []common.Address,
) *BbAssetsLockedSubscription {
	if opts == nil {
		opts = new(ethereum.SubscribeOpts)
	}
	if opts.Tick == 0 {
		opts.Tick = chainutil.DefaultSubscribeOptsTick
	}
	if opts.PastBlocks == 0 {
		opts.PastBlocks = chainutil.DefaultSubscribeOptsPastBlocks
	}

	return &BbAssetsLockedSubscription{
		bb,
		opts,
		sequenceNumberFilter,
		recipientFilter,
	}
}

type BbAssetsLockedSubscription struct {
	contract             *BitcoinBridge
	opts                 *ethereum.SubscribeOpts
	sequenceNumberFilter []*big.Int
	recipientFilter      []common.Address
}

type bitcoinBridgeAssetsLockedFunc func(
	SequenceNumber *big.Int,
	Recipient common.Address,
	TbtcAmount *big.Int,
	blockNumber uint64,
)

func (als *BbAssetsLockedSubscription) OnEvent(
	handler bitcoinBridgeAssetsLockedFunc,
) subscription.EventSubscription {
	eventChan := make(chan *abi.BitcoinBridgeAssetsLocked)
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
					event.TbtcAmount,
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

func (als *BbAssetsLockedSubscription) Pipe(
	sink chan *abi.BitcoinBridgeAssetsLocked,
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
					bbLogger.Errorf(
						"subscription failed to pull events: [%v]",
						err,
					)
				}
				fromBlock := lastBlock - als.opts.PastBlocks

				bbLogger.Infof(
					"subscription monitoring fetching past AssetsLocked events "+
						"starting from block [%v]",
					fromBlock,
				)
				events, err := als.contract.PastAssetsLockedEvents(
					fromBlock,
					nil,
					als.sequenceNumberFilter,
					als.recipientFilter,
				)
				if err != nil {
					bbLogger.Errorf(
						"subscription failed to pull events: [%v]",
						err,
					)
					continue
				}
				bbLogger.Infof(
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
	)

	return subscription.NewEventSubscription(func() {
		sub.Unsubscribe()
		cancelCtx()
	})
}

func (bb *BitcoinBridge) watchAssetsLocked(
	sink chan *abi.BitcoinBridgeAssetsLocked,
	sequenceNumberFilter []*big.Int,
	recipientFilter []common.Address,
) event.Subscription {
	subscribeFn := func(ctx context.Context) (event.Subscription, error) {
		return bb.contract.WatchAssetsLocked(
			&bind.WatchOpts{Context: ctx},
			sink,
			sequenceNumberFilter,
			recipientFilter,
		)
	}

	thresholdViolatedFn := func(elapsed time.Duration) {
		bbLogger.Warnf(
			"subscription to event AssetsLocked had to be "+
				"retried [%s] since the last attempt; please inspect "+
				"host chain connectivity",
			elapsed,
		)
	}

	subscriptionFailedFn := func(err error) {
		bbLogger.Errorf(
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

func (bb *BitcoinBridge) PastAssetsLockedEvents(
	startBlock uint64,
	endBlock *uint64,
	sequenceNumberFilter []*big.Int,
	recipientFilter []common.Address,
) ([]*abi.BitcoinBridgeAssetsLocked, error) {
	iterator, err := bb.contract.FilterAssetsLocked(
		&bind.FilterOpts{
			Start: startBlock,
			End:   endBlock,
		},
		sequenceNumberFilter,
		recipientFilter,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"error retrieving past AssetsLocked events: [%v]",
			err,
		)
	}

	events := make([]*abi.BitcoinBridgeAssetsLocked, 0)

	for iterator.Next() {
		event := iterator.Event
		events = append(events, event)
	}

	return events, nil
}

func (bb *BitcoinBridge) DepositFinalizedEvent(
	opts *ethereum.SubscribeOpts,
	depositKeyFilter []*big.Int,
) *BbDepositFinalizedSubscription {
	if opts == nil {
		opts = new(ethereum.SubscribeOpts)
	}
	if opts.Tick == 0 {
		opts.Tick = chainutil.DefaultSubscribeOptsTick
	}
	if opts.PastBlocks == 0 {
		opts.PastBlocks = chainutil.DefaultSubscribeOptsPastBlocks
	}

	return &BbDepositFinalizedSubscription{
		bb,
		opts,
		depositKeyFilter,
	}
}

type BbDepositFinalizedSubscription struct {
	contract         *BitcoinBridge
	opts             *ethereum.SubscribeOpts
	depositKeyFilter []*big.Int
}

type bitcoinBridgeDepositFinalizedFunc func(
	DepositKey *big.Int,
	InitialAmount *big.Int,
	TbtcAmount *big.Int,
	blockNumber uint64,
)

func (dfs *BbDepositFinalizedSubscription) OnEvent(
	handler bitcoinBridgeDepositFinalizedFunc,
) subscription.EventSubscription {
	eventChan := make(chan *abi.BitcoinBridgeDepositFinalized)
	ctx, cancelCtx := context.WithCancel(context.Background())

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case event := <-eventChan:
				handler(
					event.DepositKey,
					event.InitialAmount,
					event.TbtcAmount,
					event.Raw.BlockNumber,
				)
			}
		}
	}()

	sub := dfs.Pipe(eventChan)
	return subscription.NewEventSubscription(func() {
		sub.Unsubscribe()
		cancelCtx()
	})
}

func (dfs *BbDepositFinalizedSubscription) Pipe(
	sink chan *abi.BitcoinBridgeDepositFinalized,
) subscription.EventSubscription {
	ctx, cancelCtx := context.WithCancel(context.Background())
	go func() {
		ticker := time.NewTicker(dfs.opts.Tick)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				lastBlock, err := dfs.contract.blockCounter.CurrentBlock()
				if err != nil {
					bbLogger.Errorf(
						"subscription failed to pull events: [%v]",
						err,
					)
				}
				fromBlock := lastBlock - dfs.opts.PastBlocks

				bbLogger.Infof(
					"subscription monitoring fetching past DepositFinalized events "+
						"starting from block [%v]",
					fromBlock,
				)
				events, err := dfs.contract.PastDepositFinalizedEvents(
					fromBlock,
					nil,
					dfs.depositKeyFilter,
				)
				if err != nil {
					bbLogger.Errorf(
						"subscription failed to pull events: [%v]",
						err,
					)
					continue
				}
				bbLogger.Infof(
					"subscription monitoring fetched [%v] past DepositFinalized events",
					len(events),
				)

				for _, event := range events {
					sink <- event
				}
			}
		}
	}()

	sub := dfs.contract.watchDepositFinalized(
		sink,
		dfs.depositKeyFilter,
	)

	return subscription.NewEventSubscription(func() {
		sub.Unsubscribe()
		cancelCtx()
	})
}

func (bb *BitcoinBridge) watchDepositFinalized(
	sink chan *abi.BitcoinBridgeDepositFinalized,
	depositKeyFilter []*big.Int,
) event.Subscription {
	subscribeFn := func(ctx context.Context) (event.Subscription, error) {
		return bb.contract.WatchDepositFinalized(
			&bind.WatchOpts{Context: ctx},
			sink,
			depositKeyFilter,
		)
	}

	thresholdViolatedFn := func(elapsed time.Duration) {
		bbLogger.Warnf(
			"subscription to event DepositFinalized had to be "+
				"retried [%s] since the last attempt; please inspect "+
				"host chain connectivity",
			elapsed,
		)
	}

	subscriptionFailedFn := func(err error) {
		bbLogger.Errorf(
			"subscription to event DepositFinalized failed "+
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

func (bb *BitcoinBridge) PastDepositFinalizedEvents(
	startBlock uint64,
	endBlock *uint64,
	depositKeyFilter []*big.Int,
) ([]*abi.BitcoinBridgeDepositFinalized, error) {
	iterator, err := bb.contract.FilterDepositFinalized(
		&bind.FilterOpts{
			Start: startBlock,
			End:   endBlock,
		},
		depositKeyFilter,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"error retrieving past DepositFinalized events: [%v]",
			err,
		)
	}

	events := make([]*abi.BitcoinBridgeDepositFinalized, 0)

	for iterator.Next() {
		event := iterator.Event
		events = append(events, event)
	}

	return events, nil
}

func (bb *BitcoinBridge) DepositInitializedEvent(
	opts *ethereum.SubscribeOpts,
	depositKeyFilter []*big.Int,
	recipientFilter []common.Address,
) *BbDepositInitializedSubscription {
	if opts == nil {
		opts = new(ethereum.SubscribeOpts)
	}
	if opts.Tick == 0 {
		opts.Tick = chainutil.DefaultSubscribeOptsTick
	}
	if opts.PastBlocks == 0 {
		opts.PastBlocks = chainutil.DefaultSubscribeOptsPastBlocks
	}

	return &BbDepositInitializedSubscription{
		bb,
		opts,
		depositKeyFilter,
		recipientFilter,
	}
}

type BbDepositInitializedSubscription struct {
	contract         *BitcoinBridge
	opts             *ethereum.SubscribeOpts
	depositKeyFilter []*big.Int
	recipientFilter  []common.Address
}

type bitcoinBridgeDepositInitializedFunc func(
	DepositKey *big.Int,
	Recipient common.Address,
	blockNumber uint64,
)

func (dis *BbDepositInitializedSubscription) OnEvent(
	handler bitcoinBridgeDepositInitializedFunc,
) subscription.EventSubscription {
	eventChan := make(chan *abi.BitcoinBridgeDepositInitialized)
	ctx, cancelCtx := context.WithCancel(context.Background())

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case event := <-eventChan:
				handler(
					event.DepositKey,
					event.Recipient,
					event.Raw.BlockNumber,
				)
			}
		}
	}()

	sub := dis.Pipe(eventChan)
	return subscription.NewEventSubscription(func() {
		sub.Unsubscribe()
		cancelCtx()
	})
}

func (dis *BbDepositInitializedSubscription) Pipe(
	sink chan *abi.BitcoinBridgeDepositInitialized,
) subscription.EventSubscription {
	ctx, cancelCtx := context.WithCancel(context.Background())
	go func() {
		ticker := time.NewTicker(dis.opts.Tick)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				lastBlock, err := dis.contract.blockCounter.CurrentBlock()
				if err != nil {
					bbLogger.Errorf(
						"subscription failed to pull events: [%v]",
						err,
					)
				}
				fromBlock := lastBlock - dis.opts.PastBlocks

				bbLogger.Infof(
					"subscription monitoring fetching past DepositInitialized events "+
						"starting from block [%v]",
					fromBlock,
				)
				events, err := dis.contract.PastDepositInitializedEvents(
					fromBlock,
					nil,
					dis.depositKeyFilter,
					dis.recipientFilter,
				)
				if err != nil {
					bbLogger.Errorf(
						"subscription failed to pull events: [%v]",
						err,
					)
					continue
				}
				bbLogger.Infof(
					"subscription monitoring fetched [%v] past DepositInitialized events",
					len(events),
				)

				for _, event := range events {
					sink <- event
				}
			}
		}
	}()

	sub := dis.contract.watchDepositInitialized(
		sink,
		dis.depositKeyFilter,
		dis.recipientFilter,
	)

	return subscription.NewEventSubscription(func() {
		sub.Unsubscribe()
		cancelCtx()
	})
}

func (bb *BitcoinBridge) watchDepositInitialized(
	sink chan *abi.BitcoinBridgeDepositInitialized,
	depositKeyFilter []*big.Int,
	recipientFilter []common.Address,
) event.Subscription {
	subscribeFn := func(ctx context.Context) (event.Subscription, error) {
		return bb.contract.WatchDepositInitialized(
			&bind.WatchOpts{Context: ctx},
			sink,
			depositKeyFilter,
			recipientFilter,
		)
	}

	thresholdViolatedFn := func(elapsed time.Duration) {
		bbLogger.Warnf(
			"subscription to event DepositInitialized had to be "+
				"retried [%s] since the last attempt; please inspect "+
				"host chain connectivity",
			elapsed,
		)
	}

	subscriptionFailedFn := func(err error) {
		bbLogger.Errorf(
			"subscription to event DepositInitialized failed "+
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

func (bb *BitcoinBridge) PastDepositInitializedEvents(
	startBlock uint64,
	endBlock *uint64,
	depositKeyFilter []*big.Int,
	recipientFilter []common.Address,
) ([]*abi.BitcoinBridgeDepositInitialized, error) {
	iterator, err := bb.contract.FilterDepositInitialized(
		&bind.FilterOpts{
			Start: startBlock,
			End:   endBlock,
		},
		depositKeyFilter,
		recipientFilter,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"error retrieving past DepositInitialized events: [%v]",
			err,
		)
	}

	events := make([]*abi.BitcoinBridgeDepositInitialized, 0)

	for iterator.Next() {
		event := iterator.Event
		events = append(events, event)
	}

	return events, nil
}

func (bb *BitcoinBridge) InitializedEvent(
	opts *ethereum.SubscribeOpts,
) *BbInitializedSubscription {
	if opts == nil {
		opts = new(ethereum.SubscribeOpts)
	}
	if opts.Tick == 0 {
		opts.Tick = chainutil.DefaultSubscribeOptsTick
	}
	if opts.PastBlocks == 0 {
		opts.PastBlocks = chainutil.DefaultSubscribeOptsPastBlocks
	}

	return &BbInitializedSubscription{
		bb,
		opts,
	}
}

type BbInitializedSubscription struct {
	contract *BitcoinBridge
	opts     *ethereum.SubscribeOpts
}

type bitcoinBridgeInitializedFunc func(
	Version uint64,
	blockNumber uint64,
)

func (is *BbInitializedSubscription) OnEvent(
	handler bitcoinBridgeInitializedFunc,
) subscription.EventSubscription {
	eventChan := make(chan *abi.BitcoinBridgeInitialized)
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

func (is *BbInitializedSubscription) Pipe(
	sink chan *abi.BitcoinBridgeInitialized,
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
					bbLogger.Errorf(
						"subscription failed to pull events: [%v]",
						err,
					)
				}
				fromBlock := lastBlock - is.opts.PastBlocks

				bbLogger.Infof(
					"subscription monitoring fetching past Initialized events "+
						"starting from block [%v]",
					fromBlock,
				)
				events, err := is.contract.PastInitializedEvents(
					fromBlock,
					nil,
				)
				if err != nil {
					bbLogger.Errorf(
						"subscription failed to pull events: [%v]",
						err,
					)
					continue
				}
				bbLogger.Infof(
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

func (bb *BitcoinBridge) watchInitialized(
	sink chan *abi.BitcoinBridgeInitialized,
) event.Subscription {
	subscribeFn := func(ctx context.Context) (event.Subscription, error) {
		return bb.contract.WatchInitialized(
			&bind.WatchOpts{Context: ctx},
			sink,
		)
	}

	thresholdViolatedFn := func(elapsed time.Duration) {
		bbLogger.Warnf(
			"subscription to event Initialized had to be "+
				"retried [%s] since the last attempt; please inspect "+
				"host chain connectivity",
			elapsed,
		)
	}

	subscriptionFailedFn := func(err error) {
		bbLogger.Errorf(
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

func (bb *BitcoinBridge) PastInitializedEvents(
	startBlock uint64,
	endBlock *uint64,
) ([]*abi.BitcoinBridgeInitialized, error) {
	iterator, err := bb.contract.FilterInitialized(
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

	events := make([]*abi.BitcoinBridgeInitialized, 0)

	for iterator.Next() {
		event := iterator.Event
		events = append(events, event)
	}

	return events, nil
}

func (bb *BitcoinBridge) MinTBTCAmountUpdatedEvent(
	opts *ethereum.SubscribeOpts,
) *BbMinTBTCAmountUpdatedSubscription {
	if opts == nil {
		opts = new(ethereum.SubscribeOpts)
	}
	if opts.Tick == 0 {
		opts.Tick = chainutil.DefaultSubscribeOptsTick
	}
	if opts.PastBlocks == 0 {
		opts.PastBlocks = chainutil.DefaultSubscribeOptsPastBlocks
	}

	return &BbMinTBTCAmountUpdatedSubscription{
		bb,
		opts,
	}
}

type BbMinTBTCAmountUpdatedSubscription struct {
	contract *BitcoinBridge
	opts     *ethereum.SubscribeOpts
}

type bitcoinBridgeMinTBTCAmountUpdatedFunc func(
	MinTBTCAmount *big.Int,
	blockNumber uint64,
)

func (mtbtcaus *BbMinTBTCAmountUpdatedSubscription) OnEvent(
	handler bitcoinBridgeMinTBTCAmountUpdatedFunc,
) subscription.EventSubscription {
	eventChan := make(chan *abi.BitcoinBridgeMinTBTCAmountUpdated)
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

func (mtbtcaus *BbMinTBTCAmountUpdatedSubscription) Pipe(
	sink chan *abi.BitcoinBridgeMinTBTCAmountUpdated,
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
					bbLogger.Errorf(
						"subscription failed to pull events: [%v]",
						err,
					)
				}
				fromBlock := lastBlock - mtbtcaus.opts.PastBlocks

				bbLogger.Infof(
					"subscription monitoring fetching past MinTBTCAmountUpdated events "+
						"starting from block [%v]",
					fromBlock,
				)
				events, err := mtbtcaus.contract.PastMinTBTCAmountUpdatedEvents(
					fromBlock,
					nil,
				)
				if err != nil {
					bbLogger.Errorf(
						"subscription failed to pull events: [%v]",
						err,
					)
					continue
				}
				bbLogger.Infof(
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

func (bb *BitcoinBridge) watchMinTBTCAmountUpdated(
	sink chan *abi.BitcoinBridgeMinTBTCAmountUpdated,
) event.Subscription {
	subscribeFn := func(ctx context.Context) (event.Subscription, error) {
		return bb.contract.WatchMinTBTCAmountUpdated(
			&bind.WatchOpts{Context: ctx},
			sink,
		)
	}

	thresholdViolatedFn := func(elapsed time.Duration) {
		bbLogger.Warnf(
			"subscription to event MinTBTCAmountUpdated had to be "+
				"retried [%s] since the last attempt; please inspect "+
				"host chain connectivity",
			elapsed,
		)
	}

	subscriptionFailedFn := func(err error) {
		bbLogger.Errorf(
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

func (bb *BitcoinBridge) PastMinTBTCAmountUpdatedEvents(
	startBlock uint64,
	endBlock *uint64,
) ([]*abi.BitcoinBridgeMinTBTCAmountUpdated, error) {
	iterator, err := bb.contract.FilterMinTBTCAmountUpdated(
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

	events := make([]*abi.BitcoinBridgeMinTBTCAmountUpdated, 0)

	for iterator.Next() {
		event := iterator.Event
		events = append(events, event)
	}

	return events, nil
}

func (bb *BitcoinBridge) OwnershipTransferStartedEvent(
	opts *ethereum.SubscribeOpts,
	previousOwnerFilter []common.Address,
	newOwnerFilter []common.Address,
) *BbOwnershipTransferStartedSubscription {
	if opts == nil {
		opts = new(ethereum.SubscribeOpts)
	}
	if opts.Tick == 0 {
		opts.Tick = chainutil.DefaultSubscribeOptsTick
	}
	if opts.PastBlocks == 0 {
		opts.PastBlocks = chainutil.DefaultSubscribeOptsPastBlocks
	}

	return &BbOwnershipTransferStartedSubscription{
		bb,
		opts,
		previousOwnerFilter,
		newOwnerFilter,
	}
}

type BbOwnershipTransferStartedSubscription struct {
	contract            *BitcoinBridge
	opts                *ethereum.SubscribeOpts
	previousOwnerFilter []common.Address
	newOwnerFilter      []common.Address
}

type bitcoinBridgeOwnershipTransferStartedFunc func(
	PreviousOwner common.Address,
	NewOwner common.Address,
	blockNumber uint64,
)

func (otss *BbOwnershipTransferStartedSubscription) OnEvent(
	handler bitcoinBridgeOwnershipTransferStartedFunc,
) subscription.EventSubscription {
	eventChan := make(chan *abi.BitcoinBridgeOwnershipTransferStarted)
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

func (otss *BbOwnershipTransferStartedSubscription) Pipe(
	sink chan *abi.BitcoinBridgeOwnershipTransferStarted,
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
					bbLogger.Errorf(
						"subscription failed to pull events: [%v]",
						err,
					)
				}
				fromBlock := lastBlock - otss.opts.PastBlocks

				bbLogger.Infof(
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
					bbLogger.Errorf(
						"subscription failed to pull events: [%v]",
						err,
					)
					continue
				}
				bbLogger.Infof(
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

func (bb *BitcoinBridge) watchOwnershipTransferStarted(
	sink chan *abi.BitcoinBridgeOwnershipTransferStarted,
	previousOwnerFilter []common.Address,
	newOwnerFilter []common.Address,
) event.Subscription {
	subscribeFn := func(ctx context.Context) (event.Subscription, error) {
		return bb.contract.WatchOwnershipTransferStarted(
			&bind.WatchOpts{Context: ctx},
			sink,
			previousOwnerFilter,
			newOwnerFilter,
		)
	}

	thresholdViolatedFn := func(elapsed time.Duration) {
		bbLogger.Warnf(
			"subscription to event OwnershipTransferStarted had to be "+
				"retried [%s] since the last attempt; please inspect "+
				"host chain connectivity",
			elapsed,
		)
	}

	subscriptionFailedFn := func(err error) {
		bbLogger.Errorf(
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

func (bb *BitcoinBridge) PastOwnershipTransferStartedEvents(
	startBlock uint64,
	endBlock *uint64,
	previousOwnerFilter []common.Address,
	newOwnerFilter []common.Address,
) ([]*abi.BitcoinBridgeOwnershipTransferStarted, error) {
	iterator, err := bb.contract.FilterOwnershipTransferStarted(
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

	events := make([]*abi.BitcoinBridgeOwnershipTransferStarted, 0)

	for iterator.Next() {
		event := iterator.Event
		events = append(events, event)
	}

	return events, nil
}

func (bb *BitcoinBridge) OwnershipTransferredEvent(
	opts *ethereum.SubscribeOpts,
	previousOwnerFilter []common.Address,
	newOwnerFilter []common.Address,
) *BbOwnershipTransferredSubscription {
	if opts == nil {
		opts = new(ethereum.SubscribeOpts)
	}
	if opts.Tick == 0 {
		opts.Tick = chainutil.DefaultSubscribeOptsTick
	}
	if opts.PastBlocks == 0 {
		opts.PastBlocks = chainutil.DefaultSubscribeOptsPastBlocks
	}

	return &BbOwnershipTransferredSubscription{
		bb,
		opts,
		previousOwnerFilter,
		newOwnerFilter,
	}
}

type BbOwnershipTransferredSubscription struct {
	contract            *BitcoinBridge
	opts                *ethereum.SubscribeOpts
	previousOwnerFilter []common.Address
	newOwnerFilter      []common.Address
}

type bitcoinBridgeOwnershipTransferredFunc func(
	PreviousOwner common.Address,
	NewOwner common.Address,
	blockNumber uint64,
)

func (ots *BbOwnershipTransferredSubscription) OnEvent(
	handler bitcoinBridgeOwnershipTransferredFunc,
) subscription.EventSubscription {
	eventChan := make(chan *abi.BitcoinBridgeOwnershipTransferred)
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

func (ots *BbOwnershipTransferredSubscription) Pipe(
	sink chan *abi.BitcoinBridgeOwnershipTransferred,
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
					bbLogger.Errorf(
						"subscription failed to pull events: [%v]",
						err,
					)
				}
				fromBlock := lastBlock - ots.opts.PastBlocks

				bbLogger.Infof(
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
					bbLogger.Errorf(
						"subscription failed to pull events: [%v]",
						err,
					)
					continue
				}
				bbLogger.Infof(
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

func (bb *BitcoinBridge) watchOwnershipTransferred(
	sink chan *abi.BitcoinBridgeOwnershipTransferred,
	previousOwnerFilter []common.Address,
	newOwnerFilter []common.Address,
) event.Subscription {
	subscribeFn := func(ctx context.Context) (event.Subscription, error) {
		return bb.contract.WatchOwnershipTransferred(
			&bind.WatchOpts{Context: ctx},
			sink,
			previousOwnerFilter,
			newOwnerFilter,
		)
	}

	thresholdViolatedFn := func(elapsed time.Duration) {
		bbLogger.Warnf(
			"subscription to event OwnershipTransferred had to be "+
				"retried [%s] since the last attempt; please inspect "+
				"host chain connectivity",
			elapsed,
		)
	}

	subscriptionFailedFn := func(err error) {
		bbLogger.Errorf(
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

func (bb *BitcoinBridge) PastOwnershipTransferredEvents(
	startBlock uint64,
	endBlock *uint64,
	previousOwnerFilter []common.Address,
	newOwnerFilter []common.Address,
) ([]*abi.BitcoinBridgeOwnershipTransferred, error) {
	iterator, err := bb.contract.FilterOwnershipTransferred(
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

	events := make([]*abi.BitcoinBridgeOwnershipTransferred, 0)

	for iterator.Next() {
		event := iterator.Event
		events = append(events, event)
	}

	return events, nil
}
