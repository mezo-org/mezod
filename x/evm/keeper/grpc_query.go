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

	msg, err := args.ToMessage(req.GasCap, cfg.BaseFee)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	txConfig := statedb.NewEmptyTxConfig(common.BytesToHash(ctx.HeaderHash()))

	// pass false to not commit StateDB
	res, err := k.ApplyMessageWithConfig(ctx, WrapMessage(msg), nil, false, cfg, txConfig)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
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

	// NOTE: the errors from the executable below should be consistent with go-ethereum,
	// so we don't wrap them with the gRPC status code

	// Create a helper to check if a gas allowance results in an executable transaction
	executable := func(gas uint64) (vmError bool, rsp *types.MsgEthereumTxResponse, err error) {
		// update the message with the new gas value
		msg = core.Message{
			From:       msg.From,
			To:         msg.To,
			Nonce:      msg.Nonce,
			Value:      msg.Value,
			GasLimit:   gas,
			GasPrice:   msg.GasPrice,
			GasFeeCap:  msg.GasFeeCap,
			GasTipCap:  msg.GasTipCap,
			Data:       msg.Data,
			AccessList: msg.AccessList,
		}

		tmpCtx := ctx
		if fromType == types.RPC {
			tmpCtx, _ = ctx.CacheContext()

			acct := k.GetAccount(tmpCtx, msg.From)

			from := msg.From
			if acct == nil {
				acc := k.accountKeeper.NewAccountWithAddress(tmpCtx, from[:])
				k.accountKeeper.SetAccount(tmpCtx, acc)
				acct = statedb.NewEmptyAccount()
			}
			// When submitting a transaction, the `EthIncrementSenderSequence` ante handler increases the account nonce
			acct.Nonce = nonce + 1
			err = k.SetAccount(tmpCtx, from, *acct)
			if err != nil {
				return true, nil, err
			}
			// Resetting the gasMeter after increasing the sequence to have an accurate gas estimation on transactions against EVM precompiles.
			gasMeter := mezotypes.NewInfiniteGasMeterWithLimit(msg.GasLimit)
			tmpCtx = tmpCtx.WithGasMeter(gasMeter).
				WithKVGasConfig(storetypes.GasConfig{}).
				WithTransientKVGasConfig(storetypes.GasConfig{})
		}

		// pass false to not commit StateDB
		rsp, err = k.ApplyMessageWithConfig(tmpCtx, WrapMessage(msg), nil, false, cfg, txConfig)
		if err != nil {
			if errors.Is(err, core.ErrIntrinsicGas) {
				return true, nil, nil // Special case, raise gas limit
			}
			return true, nil, err // Bail out
		}
		return len(rsp.VmError) > 0, rsp, nil
	}

	// Execute the binary search and hone in on an executable gas limit
	hi, err = types.BinSearch(lo, hi, executable)
	if err != nil {
		return nil, err
	}

	// Reject the transaction as invalid if it still fails at the highest allowance
	if hi == gasCap {
		failed, result, err := executable(hi)
		if err != nil {
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
		rsp, err := k.ApplyMessageWithConfig(ctx, WrapMessageWithSource(*msg, ethTx), tracer, true, cfg, txConfig)
		if err != nil {
			continue
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
		result := types.TxTraceResult{}
		ethTx := tx.AsTransaction()
		txConfig.TxHash = ethTx.Hash()
		txConfig.TxIndex = uint(i) //nolint:gosec
		traceResult, logIndex, err := k.traceTx(ctx, cfg, txConfig, signer, ethTx, req.TraceConfig, true, nil)
		if err != nil {
			result.Error = err.Error()
		} else {
			txConfig.LogIndex = logIndex
			result.Result = traceResult
		}
		results = append(results, &result)
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
		Debug:            traceConfig.Debug,
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
		tracer, err = tracers.DefaultDirectory.New(traceConfig.Tracer, tCtx, tracerJSONConfig)
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
		if errors.Is(deadlineCtx.Err(), context.DeadlineExceeded) {
			tracer.Stop(errors.New("execution timeout"))
		}
	}()

	res, err := k.ApplyMessageWithConfig(ctx, WrapMessageWithSource(*msg, tx), tracer, commitMessage, cfg, txConfig)
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
