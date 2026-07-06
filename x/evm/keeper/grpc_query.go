// Copyright 2022 Evmos Foundation
// This file is part of the Evmos Network packages.
//
// Evmos is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The Evmos packages are distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the Evmos packages. If not, see https://github.com/evmos/evmos/blob/main/LICENSE
package keeper

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"sync/atomic"
	"time"

	"github.com/ethereum/go-ethereum/eth/tracers/logger"

	storetypes "cosmossdk.io/store/types"

	"github.com/ethereum/go-ethereum/eth/tracers"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
	ethparams "github.com/ethereum/go-ethereum/params"

	rpctypes "github.com/mezo-org/mezod/rpc/types"
	mezotypes "github.com/mezo-org/mezod/types"
	"github.com/mezo-org/mezod/x/evm/statedb"
	"github.com/mezo-org/mezod/x/evm/types"
)

var _ types.QueryServer = Keeper{}

const (
	defaultTraceTimeout = 5 * time.Second
	maxPredecessorsTxs  = 50
	maxTxGasLimit       = 10_000_000
)

// Account implements the Query/Account gRPC method
func (k Keeper) Account(c context.Context, req *types.QueryAccountRequest) (*types.QueryAccountResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if err := mezotypes.ValidateAddress(req.Address); err != nil {
		return nil, status.Error(
			codes.InvalidArgument, err.Error(),
		)
	}

	addr := common.HexToAddress(req.Address)

	ctx := sdk.UnwrapSDKContext(c)
	acct := k.GetAccountOrEmpty(ctx, addr)

	return &types.QueryAccountResponse{
		Balance:  acct.Balance.String(),
		CodeHash: common.BytesToHash(acct.CodeHash).Hex(),
		Nonce:    acct.Nonce,
	}, nil
}

func (k Keeper) CosmosAccount(c context.Context, req *types.QueryCosmosAccountRequest) (*types.QueryCosmosAccountResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if err := mezotypes.ValidateAddress(req.Address); err != nil {
		return nil, status.Error(
			codes.InvalidArgument, err.Error(),
		)
	}

	ctx := sdk.UnwrapSDKContext(c)

	ethAddr := common.HexToAddress(req.Address)
	cosmosAddr := sdk.AccAddress(ethAddr.Bytes())

	account := k.accountKeeper.GetAccount(ctx, cosmosAddr)
	res := types.QueryCosmosAccountResponse{
		CosmosAddress: cosmosAddr.String(),
	}

	if account != nil {
		res.Sequence = account.GetSequence()
		res.AccountNumber = account.GetAccountNumber()
	}

	return &res, nil
}

// ValidatorAccount implements the Query/Balance gRPC method
func (k Keeper) ValidatorAccount(c context.Context, req *types.QueryValidatorAccountRequest) (*types.QueryValidatorAccountResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	consAddr, err := sdk.ConsAddressFromBech32(req.ConsAddress)
	if err != nil {
		return nil, status.Error(
			codes.InvalidArgument, err.Error(),
		)
	}

	ctx := sdk.UnwrapSDKContext(c)

	validator, found := k.stakingKeeper.GetValidatorByConsAddr(ctx, consAddr)
	if !found {
		return nil, fmt.Errorf("validator not found for %s", consAddr.String())
	}

	accAddr := sdk.AccAddress(validator.GetOperator())

	res := types.QueryValidatorAccountResponse{
		AccountAddress: accAddr.String(),
	}

	account := k.accountKeeper.GetAccount(ctx, accAddr)
	if account != nil {
		res.Sequence = account.GetSequence()
		res.AccountNumber = account.GetAccountNumber()
	}

	return &res, nil
}

// Balance implements the Query/Balance gRPC method
func (k Keeper) Balance(c context.Context, req *types.QueryBalanceRequest) (*types.QueryBalanceResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if err := mezotypes.ValidateAddress(req.Address); err != nil {
		return nil, status.Error(
			codes.InvalidArgument,
			types.ErrZeroAddress.Error(),
		)
	}

	ctx := sdk.UnwrapSDKContext(c)

	balanceInt := k.GetBalance(ctx, common.HexToAddress(req.Address))

	return &types.QueryBalanceResponse{
		Balance: balanceInt.String(),
	}, nil
}

// Storage implements the Query/Storage gRPC method
func (k Keeper) Storage(c context.Context, req *types.QueryStorageRequest) (*types.QueryStorageResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if err := mezotypes.ValidateAddress(req.Address); err != nil {
		return nil, status.Error(
			codes.InvalidArgument,
			types.ErrZeroAddress.Error(),
		)
	}

	ctx := sdk.UnwrapSDKContext(c)

	address := common.HexToAddress(req.Address)
	key := common.HexToHash(req.Key)

	state := k.GetState(ctx, address, key)
	stateHex := state.Hex()

	return &types.QueryStorageResponse{
		Value: stateHex,
	}, nil
}

// Code implements the Query/Code gRPC method
func (k Keeper) Code(c context.Context, req *types.QueryCodeRequest) (*types.QueryCodeResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if err := mezotypes.ValidateAddress(req.Address); err != nil {
		return nil, status.Error(
			codes.InvalidArgument,
			types.ErrZeroAddress.Error(),
		)
	}

	ctx := sdk.UnwrapSDKContext(c)

	address := common.HexToAddress(req.Address)
	acct := k.GetAccountWithoutBalance(ctx, address)

	var code []byte
	if acct != nil && acct.IsContract() {
		code = k.GetCode(ctx, common.BytesToHash(acct.CodeHash))
	}

	return &types.QueryCodeResponse{
		Code: code,
	}, nil
}

// Params implements the Query/Params gRPC method
func (k Keeper) Params(c context.Context, _ *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	params := k.GetParams(ctx)

	return &types.QueryParamsResponse{
		Params: params,
	}, nil
}

// statusFromCtxErr maps a context cancellation/deadline error to the
// matching gRPC status. It is only meaningful for a non-nil context error
// (context.Canceled or context.DeadlineExceeded).
func statusFromCtxErr(err error) error {
	code := codes.Canceled
	if errors.Is(err, context.DeadlineExceeded) {
		code = codes.DeadlineExceeded
	}
	return status.Error(code, err.Error())
}

// EthCall implements eth_call rpc api.
func (k Keeper) EthCall(c context.Context, req *types.EthCallRequest) (*types.MsgEthereumTxResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(c)

	var args types.TransactionArgs
	err := json.Unmarshal(req.Args, &args)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	chainID, err := getChainID(ctx, req.ChainId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	cfg, err := k.EVMConfig(ctx, GetProposerAddress(ctx, req.ProposerAddress), chainID)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// ApplyMessageWithConfig expect correct nonce set in msg
	nonce := k.GetNonce(ctx, args.GetFrom())
	args.Nonce = (*hexutil.Uint64)(&nonce)

	gasCap := req.GasCap
	if k.ethCallGasCap != 0 && (gasCap == 0 || gasCap > k.ethCallGasCap) {
		gasCap = k.ethCallGasCap
	}

	msg, err := args.ToMessage(gasCap, cfg.BaseFee)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	txConfig := statedb.NewEmptyTxConfig(common.BytesToHash(ctx.HeaderHash()))

	var overrides types.StateOverride
	if len(req.StateOverride) > 0 {
		if err := json.Unmarshal(req.StateOverride, &overrides); err != nil {
			return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("invalid state override: %v", err))
		}
	}

	if err := k.bumpAccNonce(ctx, msg.From); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Derive a single call context used for the rest of EthCall. The node's
	// json-rpc.evm-timeout is a hard ceiling on execution time: WithTimeout
	// keeps any shorter caller-supplied deadline, so a caller can only tighten
	// it, never exceed it or run uncancelled. A zero server timeout disables
	// the ceiling ("0=infinite"). The watcher and the post-call check below
	// read only callCtx, never c.
	var (
		callCtx       context.Context
		cancelCallCtx context.CancelFunc
	)
	if k.ethCallTimeout > 0 {
		callCtx, cancelCallCtx = context.WithTimeout(c, k.ethCallTimeout)
	} else {
		callCtx, cancelCallCtx = context.WithCancel(c)
	}
	defer cancelCallCtx()

	// One watcher goroutine per EthCall. OnEVMConstructed publishes the
	// per-call *vm.EVM into liveEVM before execution starts, so request
	// cancellation can interrupt the running EVM call.
	var liveEVM atomic.Pointer[vm.EVM]
	go func() {
		<-callCtx.Done()
		if e := liveEVM.Load(); e != nil {
			e.Cancel()
		}
	}()

	res, err := k.SimulateMessage(
		ctx,
		WrapMessage(msg),
		nil,
		cfg,
		txConfig,
		overrides,
		func(evm *vm.EVM) {
			liveEVM.Store(evm)
			if callCtx.Err() != nil {
				evm.Cancel()
			}
		},
	)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	if ctxErr := callCtx.Err(); ctxErr != nil {
		return nil, statusFromCtxErr(ctxErr)
	}

	return res, nil
}

// EstimateGas implements eth_estimateGas rpc api.
func (k Keeper) EstimateGas(c context.Context, req *types.EthCallRequest) (*types.EstimateGasResponse, error) {
	return k.EstimateGasInternal(c, req, types.RPC)
}

// EstimateGasInternal returns the gas estimation for the corresponding request.
// This function is called from the RPC client (eth_estimateGas).
// When called from the RPC client, we need to reset the gas meter before
// simulating the transaction to have an accurate gas estimation for EVM
// extensions transactions.
func (k Keeper) EstimateGasInternal(c context.Context, req *types.EthCallRequest, fromType types.CallType) (*types.EstimateGasResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(c)
	chainID, err := getChainID(ctx, req.ChainId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if req.GasCap < ethparams.TxGas {
		return nil, status.Errorf(codes.InvalidArgument, "gas cap cannot be lower than %d", ethparams.TxGas)
	}

	// Clamp the caller-supplied gas cap to the server-side limit so a query
	// cannot request an unbounded gas budget; hi (the binary-search ceiling)
	// is recapped to req.GasCap below. A zero limit disables the clamp
	// ("0=infinite").
	if k.ethCallGasCap != 0 && req.GasCap > k.ethCallGasCap {
		req.GasCap = k.ethCallGasCap
	}

	var args types.TransactionArgs
	err = json.Unmarshal(req.Args, &args)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// Binary search the gas requirement, as it may be higher than the amount used
	var (
		lo     = ethparams.TxGas - 1
		hi     uint64
		gasCap uint64
	)

	// Determine the highest gas limit can be used during the estimation.
	if args.Gas != nil && uint64(*args.Gas) >= ethparams.TxGas {
		hi = uint64(*args.Gas)
	} else {
		// Query block gas limit
		paramsreq, err := k.consensusKeeper.Params(ctx, nil)
		if err != nil {
			return nil, status.Error(codes.Internal, "failed to get consensus params")
		}
		params := paramsreq.GetParams()
		if params != nil && params.Block != nil && params.Block.MaxGas > 0 {
			hi = uint64(params.Block.MaxGas) //nolint:gosec
		} else {
			hi = req.GasCap
		}
	}

	// TODO: Recap the highest gas limit with account's available balance.

	// Recap the highest gas allowance with specified gascap.
	if req.GasCap != 0 && hi > req.GasCap {
		hi = req.GasCap
	}

	gasCap = hi
	cfg, err := k.EVMConfig(ctx, GetProposerAddress(ctx, req.ProposerAddress), chainID)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to load evm config")
	}

	// ApplyMessageWithConfig expect correct nonce set in msg
	nonce := k.GetNonce(ctx, args.GetFrom())
	args.Nonce = (*hexutil.Uint64)(&nonce)

	txConfig := statedb.NewEmptyTxConfig(common.BytesToHash(ctx.HeaderHash()))

	// convert the tx args to an ethereum message
	msg, err := args.ToMessage(req.GasCap, cfg.BaseFee)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Deserialize state overrides once, before the binary search loop.
	var overrides types.StateOverride
	if len(req.StateOverride) > 0 {
		if err := json.Unmarshal(req.StateOverride, &overrides); err != nil {
			return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("invalid state override: %v", err))
		}
	}

	// Derive a single call context for the binary search. The node's
	// json-rpc.evm-timeout is a hard ceiling on execution time: WithTimeout
	// keeps any shorter caller-supplied deadline, so a caller can only tighten
	// it, never exceed it or run uncancelled. A zero server timeout disables
	// the ceiling ("0=infinite"). The watcher cancels whichever EVM call is
	// currently running (one per binary-search iteration).
	var (
		callCtx       context.Context
		cancelCallCtx context.CancelFunc
	)
	if k.ethCallTimeout > 0 {
		callCtx, cancelCallCtx = context.WithTimeout(c, k.ethCallTimeout)
	} else {
		callCtx, cancelCallCtx = context.WithCancel(c)
	}
	defer cancelCallCtx()

	var liveEVM atomic.Pointer[vm.EVM]
	go func() {
		<-callCtx.Done()
		if e := liveEVM.Load(); e != nil {
			e.Cancel()
		}
	}()

	// NOTE: the errors from the executable below should be consistent with go-ethereum,
	// so we don't wrap them with the gRPC status code

	// Create a helper to check if a gas allowance results in an executable transaction
	executable := func(gas uint64) (vmError bool, rsp *types.MsgEthereumTxResponse, err error) {
		// Stop the search promptly once the request context is done.
		if err := callCtx.Err(); err != nil {
			return true, nil, err
		}
		// update the message with the new gas value
		msg = core.Message{
			From:                  msg.From,
			To:                    msg.To,
			Nonce:                 msg.Nonce,
			Value:                 msg.Value,
			GasLimit:              gas,
			GasPrice:              msg.GasPrice,
			GasFeeCap:             msg.GasFeeCap,
			GasTipCap:             msg.GasTipCap,
			Data:                  msg.Data,
			AccessList:            msg.AccessList,
			SetCodeAuthorizations: msg.SetCodeAuthorizations,
		}

		tmpCtx := ctx
		if fromType == types.RPC {
			tmpCtx, _ = ctx.CacheContext()

			if err := k.bumpAccNonce(tmpCtx, msg.From); err != nil {
				return true, nil, err
			}
			// Resetting the gasMeter after increasing the sequence to have an accurate gas estimation on transactions against EVM precompiles.
			gasMeter := mezotypes.NewInfiniteGasMeterWithLimit(msg.GasLimit)
			tmpCtx = tmpCtx.WithGasMeter(gasMeter).
				WithKVGasConfig(storetypes.GasConfig{}).
				WithTransientKVGasConfig(storetypes.GasConfig{})
		}

		rsp, err = k.SimulateMessage(tmpCtx, WrapMessage(msg), nil, cfg, txConfig, overrides,
			func(evm *vm.EVM) {
				liveEVM.Store(evm)
				if callCtx.Err() != nil {
					evm.Cancel()
				}
			},
		)
		if err != nil {
			if errors.Is(err, core.ErrIntrinsicGas) || errors.Is(err, core.ErrFloorDataGas) {
				return true, nil, nil // Special case, raise gas limit
			}
			return true, nil, err // Bail out
		}
		return len(rsp.VmError) > 0, rsp, nil
	}

	// Pre-seed lo above the EIP-7623 floor so BinSearch skips the
	// iterations between TxGas and the floor on calldata-heavy txs. The
	// `floor-1 < hi` guard handles the caller-supplied `args.Gas < floor`
	// edge case: without it, `lo` could land at `floor-1 >= hi`, BinSearch
	// would exit immediately and return an unverified `hi`. With the
	// guard, the final-allowance check below re-runs `executable(hi)` and
	// surfaces the failure.
	floor, err := k.GetEthFloorDataGas(ctx, msg, cfg.ChainConfig)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	if floor > 0 && floor-1 < hi && floor-1 > lo {
		lo = floor - 1
	}

	// Execute the binary search and hone in on an executable gas limit
	hi, err = types.BinSearch(lo, hi, executable)
	if err != nil {
		if ctxErr := callCtx.Err(); ctxErr != nil {
			return nil, statusFromCtxErr(ctxErr)
		}
		return nil, err
	}

	// Reject the transaction as invalid if it still fails at the highest allowance
	if hi == gasCap {
		failed, result, err := executable(hi)
		if err != nil {
			if ctxErr := callCtx.Err(); ctxErr != nil {
				return nil, statusFromCtxErr(ctxErr)
			}
			return nil, err
		}

		if failed {
			if result != nil && result.VmError != vm.ErrOutOfGas.Error() {
				if result.VmError == vm.ErrExecutionReverted.Error() {
					return nil, types.NewExecErrorWithReason(result.Ret)
				}
				return nil, errors.New(result.VmError)
			}
			// Otherwise, the specified gas cap is too low
			return nil, fmt.Errorf("gas required exceeds allowance (%d)", gasCap)
		}
	}
	return &types.EstimateGasResponse{Gas: hi}, nil
}

// TraceTx configures a new tracer according to the provided configuration, and
// executes the given message in the provided environment. The return value will
// be tracer dependent.
func (k Keeper) TraceTx(c context.Context, req *types.QueryTraceTxRequest) (*types.QueryTraceTxResponse, error) {
	if req == nil || req.Msg == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if req.TraceConfig != nil && req.TraceConfig.Limit < 0 {
		return nil, status.Errorf(codes.InvalidArgument, "output limit cannot be negative, got %d", req.TraceConfig.Limit)
	}

	// Prevent processing transactions with too many predecessors to avoid DoS.
	if len(req.Predecessors) > maxPredecessorsTxs {
		return nil, status.Errorf(
			codes.ResourceExhausted,
			"too many predecessor transactions, got %d, max %d",
			len(req.Predecessors),
			maxPredecessorsTxs,
		)
	}

	// minus one to get the context of block beginning
	contextHeight := req.BlockNumber - 1
	if contextHeight < 1 {
		// 0 is a special value in `ContextWithHeight`
		contextHeight = 1
	}

	ctx := sdk.UnwrapSDKContext(c)
	ctx = ctx.WithBlockHeight(contextHeight)
	ctx = ctx.WithBlockTime(req.BlockTime)
	ctx = ctx.WithHeaderHash(common.Hex2Bytes(req.BlockHash))
	chainID, err := getChainID(ctx, req.ChainId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	cfg, err := k.EVMConfig(ctx, GetProposerAddress(ctx, req.ProposerAddress), chainID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to load evm config: %s", err.Error())
	}
	blockTime := big.NewInt(ctx.BlockTime().Unix())
	signer := ethtypes.MakeSigner(cfg.ChainConfig, big.NewInt(ctx.BlockHeight()), blockTime.Uint64())

	txConfig := statedb.NewEmptyTxConfig(common.BytesToHash(ctx.HeaderHash()))
	for i, tx := range req.Predecessors {
		ethTx := tx.AsTransaction()
		msg, err := core.TransactionToMessage(ethTx, signer, cfg.BaseFee)
		if err != nil {
			continue
		}

		// Prevent processing transactions with extremely high gas limit to
		// avoid DoS.
		if msg.GasLimit > maxTxGasLimit {
			return nil, status.Errorf(
				codes.ResourceExhausted,
				"gas limit in predecessor tx too high, got %d, max %d",
				msg.GasLimit,
				maxTxGasLimit,
			)
		}

		txConfig.TxHash = ethTx.Hash()
		txConfig.TxIndex = uint(i) //nolint:gosec
		tracer, err := types.NewNoopTracer()
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}

		if err := k.bumpAccNonce(ctx, msg.From); err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}

		rsp, _, err := k.ApplyMessageWithConfig(ctx, WrapMessageWithSource(*msg, ethTx), tracer, true, cfg, txConfig)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
		txConfig.LogIndex += uint(len(rsp.Logs))
	}

	tx := req.Msg.AsTransaction()
	txConfig.TxHash = tx.Hash()
	if len(req.Predecessors) > 0 {
		txConfig.TxIndex++
	}

	var tracerConfig json.RawMessage
	if req.TraceConfig != nil && req.TraceConfig.TracerJsonConfig != "" {
		// ignore error. default to no traceConfig
		_ = json.Unmarshal([]byte(req.TraceConfig.TracerJsonConfig), &tracerConfig)
	}

	result, _, err := k.traceTx(ctx, cfg, txConfig, signer, tx, req.TraceConfig, false, tracerConfig)
	if err != nil {
		// error will be returned with detail status from traceTx
		return nil, err
	}

	resultData, err := json.Marshal(result)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryTraceTxResponse{
		Data: resultData,
	}, nil
}

// TraceBlock configures a new tracer according to the provided configuration, and
// executes the given message in the provided environment for all the transactions in the queried block.
// The return value will be tracer dependent.
func (k Keeper) TraceBlock(c context.Context, req *types.QueryTraceBlockRequest) (*types.QueryTraceBlockResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if req.TraceConfig != nil && req.TraceConfig.Limit < 0 {
		return nil, status.Errorf(codes.InvalidArgument, "output limit cannot be negative, got %d", req.TraceConfig.Limit)
	}

	// minus one to get the context of block beginning
	contextHeight := req.BlockNumber - 1
	if contextHeight < 1 {
		// 0 is a special value in `ContextWithHeight`
		contextHeight = 1
	}

	ctx := sdk.UnwrapSDKContext(c)
	ctx = ctx.WithBlockHeight(contextHeight)
	ctx = ctx.WithBlockTime(req.BlockTime)
	ctx = ctx.WithHeaderHash(common.Hex2Bytes(req.BlockHash))
	chainID, err := getChainID(ctx, req.ChainId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	cfg, err := k.EVMConfig(ctx, GetProposerAddress(ctx, req.ProposerAddress), chainID)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to load evm config")
	}
	blockTime := big.NewInt(ctx.BlockTime().Unix())
	signer := ethtypes.MakeSigner(cfg.ChainConfig, big.NewInt(ctx.BlockHeight()), blockTime.Uint64())
	txsLength := len(req.Txs)
	results := make([]*types.TxTraceResult, 0, txsLength)

	txConfig := statedb.NewEmptyTxConfig(common.BytesToHash(ctx.HeaderHash()))
	for i, tx := range req.Txs {
		ethTx := tx.AsTransaction()
		txConfig.TxHash = ethTx.Hash()
		txConfig.TxIndex = uint(i) //nolint:gosec
		traceResult, logIndex, err := k.traceTx(ctx, cfg, txConfig, signer, ethTx, req.TraceConfig, true, nil)
		if err != nil {
			return nil, err
		}
		txConfig.LogIndex = logIndex
		results = append(results, &types.TxTraceResult{Result: traceResult})
	}

	resultData, err := json.Marshal(results)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryTraceBlockResponse{
		Data: resultData,
	}, nil
}

// traceTx do trace on one transaction, it returns a tuple: (traceResult, nextLogIndex, error).
func (k *Keeper) traceTx(
	ctx sdk.Context,
	cfg *statedb.EVMConfig,
	txConfig statedb.TxConfig,
	signer ethtypes.Signer,
	tx *ethtypes.Transaction,
	traceConfig *types.TraceConfig,
	commitMessage bool,
	tracerJSONConfig json.RawMessage,
) (*interface{}, uint, error) {
	// Assemble the structured logger or the JavaScript tracer
	var (
		tracer    *tracers.Tracer
		overrides *ethparams.ChainConfig
		err       error
		timeout   = defaultTraceTimeout
	)
	msg, err := core.TransactionToMessage(tx, signer, cfg.BaseFee)
	if err != nil {
		return nil, 0, status.Error(codes.Internal, err.Error())
	}

	// Prevent processing transactions with extremely high gas limit to
	// avoid DoS.
	if msg.GasLimit > maxTxGasLimit {
		return nil, 0, status.Errorf(
			codes.ResourceExhausted,
			"gas limit in tx too high, got %d, max %d",
			msg.GasLimit,
			maxTxGasLimit,
		)
	}

	if traceConfig == nil {
		traceConfig = &types.TraceConfig{}
	}

	if traceConfig.Overrides != nil {
		overrides = traceConfig.Overrides.EthereumConfig(cfg.ChainConfig.ChainID)
	}

	l := logger.NewStructLogger(&logger.Config{
		EnableMemory:     traceConfig.EnableMemory,
		DisableStack:     traceConfig.DisableStack,
		DisableStorage:   traceConfig.DisableStorage,
		EnableReturnData: traceConfig.EnableReturnData,
		Limit:            int(traceConfig.Limit),
		Overrides:        overrides,
	})
	tracer = &tracers.Tracer{
		Hooks:     l.Hooks(),
		GetResult: l.GetResult,
		Stop:      l.Stop,
	}

	tCtx := &tracers.Context{
		BlockHash: txConfig.BlockHash,
		TxIndex:   int(txConfig.TxIndex), //nolint:gosec
		TxHash:    txConfig.TxHash,
	}

	if len(traceConfig.Tracer) != 0 {
		tracer, err = tracers.DefaultDirectory.New(traceConfig.Tracer, tCtx, tracerJSONConfig, cfg.ChainConfig)
		if err != nil {
			return nil, 0, status.Error(codes.Internal, err.Error())
		}
	}

	// Define a meaningful timeout of a single transaction trace
	if traceConfig.Timeout != "" {
		if timeout, err = time.ParseDuration(traceConfig.Timeout); err != nil {
			return nil, 0, status.Errorf(codes.InvalidArgument, "timeout value: %s", err.Error())
		}
	}

	// Handle timeouts and RPC cancellations
	deadlineCtx, cancel := context.WithTimeout(ctx.Context(), timeout)
	defer cancel()

	go func() {
		<-deadlineCtx.Done()
		// Guard against tracers that leave Stop unset (e.g. the native
		// keccak256PreimageTracer). Calling a nil Stop here panics on this
		// spawned goroutine, which request-level recovery cannot catch, so
		// an unset Stop would crash the whole process.
		if errors.Is(deadlineCtx.Err(), context.DeadlineExceeded) && tracer.Stop != nil {
			tracer.Stop(errors.New("execution timeout"))
		}
	}()

	if err := k.bumpAccNonce(ctx, msg.From); err != nil {
		return nil, 0, status.Error(codes.Internal, err.Error())
	}

	res, _, err := k.ApplyMessageWithConfig(ctx, WrapMessageWithSource(*msg, tx), tracer, commitMessage, cfg, txConfig)
	if err != nil {
		return nil, 0, status.Error(codes.Internal, err.Error())
	}

	var result interface{}
	result, err = tracer.GetResult()
	if err != nil {
		return nil, 0, status.Error(codes.Internal, err.Error())
	}

	return &result, txConfig.LogIndex + uint(len(res.Logs)), nil
}

// SimulateV1 implements the eth_simulateV1 gRPC backend.
//
// The `simulate-disabled` operator kill switch is enforced at the
// JSON-RPC namespace handler (PublicAPI.SimulateV1) — not here. Direct
// gRPC peers bypass it; operators who need to suppress simulate
// entirely must additionally restrict the SDK gRPC port (default 9090).
// Same applies to the operator-derived RPCGasCap and RPCEVMTimeout
// defaults, which the RPC backend injects via req.GasCap and
// req.TimeoutMs. The keeper only sees the wire values and treats 0 as
// "no bound" — see newSimGasBudget and the WithTimeout block below.
func (k Keeper) SimulateV1(c context.Context, req *types.SimulateV1Request) (*types.SimulateV1Response, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(c)

	// Defense-in-depth for callers reaching the keeper gRPC directly
	// (bypassing rpc/backend, which ordinarily anchors ctx at the
	// resolved base height via rpctypes.ContextWithHeight).
	// rpc/backend.Backend.SimulateV1 (rpc/backend/simulate_v1.go)
	// always populates BlockNumberOrHash, defaulting to "latest" when
	// the caller omits it; an empty field at this point means we're on
	// a direct-gRPC path and the ctx is authoritative.
	if err := validateSimulateV1Anchor(ctx, req.BlockNumberOrHash); err != nil {
		return simulateV1ErrResponse(err)
	}

	chainID, err := getChainID(ctx, req.ChainId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	cfg, err := k.EVMConfig(ctx, GetProposerAddress(ctx, req.ProposerAddress), chainID)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	opts, err := types.UnmarshalSimOpts(req.Opts)
	if err != nil {
		return simulateV1ErrResponse(err)
	}

	// Reject oversized envelopes before sanitizeSimChain clones the
	// input slice. sanitizeSimChain only gap-fills with empty Calls, so
	// the post-sanitize call total always equals the pre-sanitize one;
	// catching it here is sufficient.
	if n := len(opts.BlockStateCalls); n > types.MaxSimulateBlocks {
		return simulateV1ErrResponse(types.NewSimBlockCountExceeded(n, types.MaxSimulateBlocks))
	}
	if n := types.CountSimCalls(opts.BlockStateCalls); n > types.MaxSimulateCalls {
		return simulateV1ErrResponse(types.NewSimCallLimitExceeded(n, types.MaxSimulateCalls))
	}

	baseGasLimit, err := k.simulateBaseGasLimit(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get consensus params")
	}

	var baseHash common.Hash
	if len(req.BaseBlockHash) > 0 {
		baseHash = common.BytesToHash(req.BaseBlockHash)
	}

	// The node's json-rpc.evm-timeout is a hard ceiling on execution time: the
	// caller may only request a shorter deadline, never a longer one (or none).
	// A zero server timeout disables the ceiling ("0=infinite"). WithTimeout
	// further keeps any shorter deadline already on c; the watcher in
	// processSimBlock keys off this ctx.
	timeout := time.Duration(req.TimeoutMs) * time.Millisecond
	if k.ethCallTimeout > 0 && (timeout <= 0 || timeout > k.ethCallTimeout) {
		timeout = k.ethCallTimeout
	}
	if timeout > 0 {
		var cancel context.CancelFunc
		c, cancel = context.WithTimeout(c, timeout)
		defer cancel()
	}

	// Clamp the caller-supplied gas cap to the server-side limit so a query
	// cannot request an unbounded gas budget (newSimGasBudget treats 0 as
	// unlimited). A zero limit disables the clamp ("0=infinite").
	gasCap := req.GasCap
	if k.ethCallGasCap != 0 && (gasCap == 0 || gasCap > k.ethCallGasCap) {
		gasCap = k.ethCallGasCap
	}

	results, err := k.simulateV1(c, ctx, cfg, baseHeaderFromContext(ctx, cfg, baseGasLimit), baseHash, opts, gasCap, timeout)
	if err != nil {
		return simulateV1ErrResponse(err)
	}

	payload, err := json.Marshal(results)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &types.SimulateV1Response{Result: payload}, nil
}

// simulateV1ErrResponse routes a driver-layer error onto the response's
// structured SimError field for spec-coded failures, or onto a gRPC
// Internal status for everything else. core.ErrIntrinsicGas is mapped
// to a structured -38013 SimError; CallResultFailure does not permit
// -38013 on a per-call entry so it must surface at the request level.
func simulateV1ErrResponse(err error) (*types.SimulateV1Response, error) {
	var simErr *types.SimError
	if errors.As(err, &simErr) {
		return &types.SimulateV1Response{Error: simErr}, nil
	}
	if errors.Is(err, core.ErrIntrinsicGas) || errors.Is(err, core.ErrFloorDataGas) {
		return &types.SimulateV1Response{Error: types.NewSimIntrinsicGas(0, 0)}, nil
	}
	return nil, status.Error(codes.Internal, err.Error())
}

// BaseFee implements the Query/BaseFee gRPC method
func (k Keeper) BaseFee(c context.Context, _ *types.QueryBaseFeeRequest) (*types.QueryBaseFeeResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	params := k.GetParams(ctx)
	ethCfg := params.ChainConfig.EthereumConfig(k.eip155ChainID)
	baseFee := k.GetBaseFee(ctx, ethCfg)

	res := &types.QueryBaseFeeResponse{}
	if baseFee != nil {
		aux := sdkmath.NewIntFromBigInt(baseFee)
		res.BaseFee = &aux
	}

	return res, nil
}

// getChainID parse chainID from current context if not provided
func getChainID(ctx sdk.Context, chainID int64) (*big.Int, error) {
	if chainID == 0 {
		return mezotypes.ParseChainID(ctx.ChainID())
	}
	return big.NewInt(chainID), nil
}

// validateSimulateV1Anchor performs a minimal consistency check between
// the request's BlockNumberOrHash field and the SDK context's block
// height. The backend (rpc/backend/simulate_v1.go) already resolves the
// field into a concrete height and anchors ctx via
// rpctypes.ContextWithHeight, so for the normal call path this is a
// no-op. Direct gRPC callers that desynchronize the two surfaces hit
// a spec-conformant -32602 instead of silently simulating against the
// wrong base.
//
// Sentinel BlockNumber values (Latest / Earliest / Pending / Finalized
// / Safe) and hash-only addressing skip the numeric comparison — the
// backend's resolution is trusted for those cases.
func validateSimulateV1Anchor(ctx sdk.Context, bnhBz []byte) error {
	if len(bnhBz) == 0 {
		return nil
	}
	var bnh rpctypes.BlockNumberOrHash
	if err := json.Unmarshal(bnhBz, &bnh); err != nil {
		return types.NewSimInvalidParams(fmt.Sprintf(
			"simulate: malformed blockNumberOrHash: %s", err.Error(),
		))
	}
	if bnh.BlockNumber == nil {
		return nil
	}
	requested := bnh.BlockNumber.Int64()
	if requested < 0 {
		// Sentinel values (latest, earliest, pending, finalized, safe)
		// are resolved in the backend; the ctx height is authoritative.
		return nil
	}
	if requested != ctx.BlockHeight() {
		return types.NewSimInvalidParams(fmt.Sprintf(
			"simulate: blockNumberOrHash (%d) does not match anchored context height (%d)",
			requested, ctx.BlockHeight(),
		))
	}
	return nil
}

// simulateBaseGasLimit reads the consensus block gas limit via the
// consensus keeper. baseapp.CreateQueryContext does not populate
// ctx.ConsensusParams or attach a BlockGasMeter for query-side calls,
// so the keeper-attached ConsensusParamsKeeper is the only path that
// returns the chain's configured limit here.
func (k Keeper) simulateBaseGasLimit(ctx sdk.Context) (uint64, error) {
	consensusParamsResp, err := k.consensusKeeper.Params(ctx, nil)
	if err != nil {
		return 0, err
	}

	switch maxGas := consensusParamsResp.GetParams().GetBlock().GetMaxGas(); {
	case maxGas == -1:
		// Consensus "unlimited" sentinel: surface max uint32 instead
		// of a full uint64 so JS dev tooling does not choke on a value
		// past 2^53.
		return uint64(^uint32(0)), nil
	case maxGas > 0:
		return uint64(maxGas), nil //nolint:gosec
	default:
		return 0, nil
	}
}

// baseHeaderFromContext synthesizes the execution-api base header from
// the SDK context that the gRPC call was anchored at. The returned
// header only populates the fields the simulate driver consumes
// (Number, Time, GasLimit, BaseFee, Difficulty, Coinbase). The ctx has
// already been anchored at the requested base height by the rpc/backend
// layer (via rpctypes.ContextWithHeight); newSimGetHashFn delegates
// canonical-range BLOCKHASH resolution to k.GetHashFn which reads the
// same ctx, so all three sources (ctx, base header fields, BLOCKHASH
// for base.Number) stay consistent.
func baseHeaderFromContext(ctx sdk.Context, cfg *statedb.EVMConfig, gasLimit uint64) *ethtypes.Header {
	return &ethtypes.Header{
		Number:     big.NewInt(ctx.BlockHeight()),
		Time:       uint64(ctx.BlockTime().Unix()), //nolint:gosec
		GasLimit:   gasLimit,
		BaseFee:    cfg.BaseFee,
		Difficulty: new(big.Int),
		// Match the non-simulate path so COINBASE returns the validator
		// operator address rather than zero for simulated blocks that
		// don't override FeeRecipient.
		Coinbase: cfg.CoinBase,
	}
}
