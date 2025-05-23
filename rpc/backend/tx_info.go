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
package backend

import (
	"fmt"
	"math"
	"math/big"

	tmrpcclient "github.com/cometbft/cometbft/rpc/client"

	errorsmod "cosmossdk.io/errors"

	tmrpctypes "github.com/cometbft/cometbft/rpc/core/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/mezo-org/mezod/indexer"
	"github.com/mezo-org/mezod/precompile/assetsbridge"
	rpctypes "github.com/mezo-org/mezod/rpc/types"
	"github.com/mezo-org/mezod/types"
	bridgetypes "github.com/mezo-org/mezod/x/bridge/abci/types"
	evmtypes "github.com/mezo-org/mezod/x/evm/types"
	"github.com/pkg/errors"
)

// GetTransactionByHash returns the Ethereum format transaction identified by Ethereum transaction hash
func (b *Backend) GetTransactionByHash(txHash common.Hash) (*rpctypes.RPCTransaction, error) {
	res, err := b.GetTxByEthHash(txHash)
	hexTx := txHash.Hex()

	if err != nil {
		return b.getTransactionByHashPending(txHash)
	}

	block, err := b.TendermintBlockByNumber(rpctypes.BlockNumber(res.Height))
	if err != nil {
		return nil, err
	}

	// Special case for pseudo-transactions containing bridging information.
	if len(res.ExtraData) > 0 && res.ExtraData[0] == byte(indexer.BridgingInfoDescriptor) {
		return b.getPseudoTransaction(res, block)
	}

	tx, err := b.clientCtx.TxConfig.TxDecoder()(block.Block.Txs[res.TxIndex])
	if err != nil {
		return nil, err
	}

	// the `res.MsgIndex` is inferred from tx index, should be within the bound.
	msg, ok := tx.GetMsgs()[res.MsgIndex].(*evmtypes.MsgEthereumTx)
	if !ok {
		return nil, errors.New("invalid ethereum tx")
	}

	blockRes, err := b.TendermintBlockResultByNumber(&block.Block.Height)
	if err != nil {
		b.logger.Debug("block result not found", "height", block.Block.Height, "error", err.Error())
		return nil, nil
	}

	if res.EthTxIndex == -1 {
		// Fallback to find tx index by iterating all valid eth transactions
		msgs := b.EthMsgsFromTendermintBlock(block, blockRes)
		for i := range msgs {
			if msgs[i].Hash == hexTx {
				if i > math.MaxInt32 {
					return nil, errors.New("tx index overflow")
				}
				res.EthTxIndex = int32(i) //#nosec G701 -- checked for int overflow already
				break
			}
		}
	}
	// if we still unable to find the eth tx index, return error, shouldn't happen.
	if res.EthTxIndex == -1 {
		return nil, errors.New("can't find index of ethereum tx")
	}

	baseFee, err := b.BaseFee(blockRes)
	if err != nil {
		// handle the error for pruned node.
		b.logger.Error("failed to fetch Base Fee from prunned block. Check node prunning configuration", "height", blockRes.Height, "error", err)
	}

	height := uint64(res.Height)    //#nosec G701 -- checked for int overflow already
	index := uint64(res.EthTxIndex) //#nosec G701 -- checked for int overflow already
	return rpctypes.NewTransactionFromMsg(
		msg,
		common.BytesToHash(block.BlockID.Hash.Bytes()),
		height,
		index,
		baseFee,
		b.chainID,
	)
}

func (b *Backend) getPseudoTransaction(
	txResult *types.TxResult,
	blockResult *tmrpctypes.ResultBlock,
) (
	*rpctypes.RPCTransaction,
	error,
) {
	blockHash := common.BytesToHash(blockResult.BlockID.Hash.Bytes())
	blockNumber := (*hexutil.Big)(new(big.Int).SetUint64(uint64(txResult.Height)))
	to := common.HexToAddress(assetsbridge.EvmAddress)
	index := hexutil.Uint64(txResult.EthTxIndex)
	chainID := (*hexutil.Big)(b.chainID)
	zero := (*hexutil.Big)(new(big.Int).SetUint64(0))

	tx := blockResult.Block.Txs[txResult.TxIndex]
	txHash := common.BytesToHash(tx.Hash())

	// Skip the descriptor byte.
	serializedEvents := txResult.ExtraData[1:]

	// Unmarshal the serialized event.
	var bridgeTx bridgetypes.InjectedTx
	b.clientCtx.Codec.MustUnmarshal(serializedEvents, &bridgeTx)

	events := make([]assetsbridge.AssetsLockedEvent, 0, len(bridgeTx.AssetsLockedEvents))
	for _, event := range bridgeTx.AssetsLockedEvents {
		accAddress, err := sdk.AccAddressFromBech32(event.Recipient)
		if err != nil {
			return nil, fmt.Errorf(
				"failed to convert Mezo address to account address: [%w]",
				err,
			)
		}

		recipient := common.BytesToAddress(accAddress)

		events = append(events, assetsbridge.AssetsLockedEvent{
			SequenceNumber: event.Sequence.BigInt(),
			Recipient:      recipient,
			Amount:         event.Amount.BigInt(),
			Token:          common.HexToAddress(event.Token),
		})
	}

	// Pack the events to an input of the precompile's `bridge` function.
	input, err := assetsbridge.PackEventsToInput(events)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare input: [%w]", err)
	}

	return &rpctypes.RPCTransaction{
		BlockHash:        &blockHash,
		BlockNumber:      blockNumber,
		GasPrice:         zero,
		GasFeeCap:        zero,
		GasTipCap:        zero,
		Hash:             txHash,
		Input:            input,
		To:               &to,
		Type:             2,
		TransactionIndex: &index,
		Value:            zero,
		ChainID:          chainID,
	}, nil
}

// getTransactionByHashPending find pending tx from mempool
func (b *Backend) getTransactionByHashPending(txHash common.Hash) (*rpctypes.RPCTransaction, error) {
	hexTx := txHash.Hex()
	// try to find tx in mempool
	txs, err := b.PendingTransactions()
	if err != nil {
		b.logger.Debug("tx not found", "hash", hexTx, "error", err.Error())
		return nil, nil
	}

	for _, tx := range txs {
		msg, err := evmtypes.UnwrapEthereumMsg(tx, txHash)
		if err != nil {
			// not ethereum tx
			continue
		}

		if msg.Hash == hexTx {
			// use zero block values since it's not included in a block yet
			rpctx, err := rpctypes.NewTransactionFromMsg(
				msg,
				common.Hash{},
				uint64(0),
				uint64(0),
				nil,
				b.chainID,
			)
			if err != nil {
				return nil, err
			}
			return rpctx, nil
		}
	}

	b.logger.Debug("tx not found", "hash", hexTx)
	return nil, nil
}

// GetGasUsed returns gasUsed from transaction
func (b *Backend) GetGasUsed(res *types.TxResult, price *big.Int, gas uint64) uint64 {
	// patch gasUsed if tx is reverted and happened before height on which fixed was introduced
	// to return real gas charged
	// more info at https://github.com/evmos/ethermint/pull/1557
	if res.Failed && res.Height < b.cfg.JSONRPC.FixRevertGasRefundHeight {
		return new(big.Int).Mul(price, new(big.Int).SetUint64(gas)).Uint64()
	}
	return res.GasUsed
}

// GetTransactionReceipt returns the transaction receipt identified by hash.
func (b *Backend) GetTransactionReceipt(hash common.Hash) (map[string]interface{}, error) {
	hexTx := hash.Hex()
	b.logger.Debug("eth_getTransactionReceipt", "hash", hexTx)

	res, err := b.GetTxByEthHash(hash)
	if err != nil {
		b.logger.Debug("tx not found", "hash", hexTx, "error", err.Error())
		return nil, nil
	}
	resBlock, err := b.TendermintBlockByNumber(rpctypes.BlockNumber(res.Height))
	if err != nil {
		b.logger.Debug("block not found", "height", res.Height, "error", err.Error())
		return nil, nil
	}

	// Special case for pseudo-transactions containing bridging information.
	if len(res.ExtraData) > 0 && res.ExtraData[0] == byte(indexer.BridgingInfoDescriptor) {
		return b.getPseudoTransactionReceipt(hash, res, resBlock), nil
	}

	tx, err := b.clientCtx.TxConfig.TxDecoder()(resBlock.Block.Txs[res.TxIndex])
	if err != nil {
		b.logger.Debug("decoding failed", "error", err.Error())
		return nil, fmt.Errorf("failed to decode tx: %w", err)
	}
	ethMsg := tx.GetMsgs()[res.MsgIndex].(*evmtypes.MsgEthereumTx)

	txData, err := evmtypes.UnpackTxData(ethMsg.Data)
	if err != nil {
		b.logger.Error("failed to unpack tx data", "error", err.Error())
		return nil, err
	}

	cumulativeGasUsed := uint64(0)
	blockRes, err := b.TendermintBlockResultByNumber(&res.Height)
	if err != nil {
		b.logger.Debug("failed to retrieve block results", "height", res.Height, "error", err.Error())
		return nil, nil
	}
	for _, txResult := range blockRes.TxsResults[0:res.TxIndex] {
		cumulativeGasUsed += uint64(txResult.GasUsed) // #nosec G701 -- checked for int overflow already
	}
	cumulativeGasUsed += res.CumulativeGasUsed

	var status hexutil.Uint
	if res.Failed {
		status = hexutil.Uint(ethtypes.ReceiptStatusFailed)
	} else {
		status = hexutil.Uint(ethtypes.ReceiptStatusSuccessful)
	}
	chainID, err := b.ChainID()
	if err != nil {
		return nil, err
	}

	from, err := ethMsg.GetSender(chainID.ToInt())
	if err != nil {
		return nil, err
	}

	// parse tx logs from events
	msgIndex := int(res.MsgIndex) // #nosec G701 -- checked for int overflow already

	logs, err := TxLogsFromEvents(blockRes.TxsResults[res.TxIndex].Events, msgIndex)
	if err != nil {
		b.logger.Debug("failed to parse logs", "hash", hexTx, "error", err.Error())
	}

	// Check if the block contains a pseudo-transaction.
	pseudoTxResult := b.GetPseudoTransactionResult(resBlock)

	// Adjust the transaction index to account for the pseudo-transaction.
	if pseudoTxResult != nil {
		for _, log := range logs {
			log.TxIndex++
		}
	}

	if res.EthTxIndex == -1 {
		// Fallback to find tx index by iterating all valid eth transactions
		msgs := b.EthMsgsFromTendermintBlock(resBlock, blockRes)
		for i := range msgs {
			if msgs[i].Hash == hexTx {
				res.EthTxIndex = int32(i) // #nosec G701
				break
			}
		}
	}
	// return error if still unable to find the eth tx index
	if res.EthTxIndex == -1 {
		return nil, errors.New("can't find index of ethereum tx")
	}

	receipt := map[string]interface{}{
		// Consensus fields: These fields are defined by the Yellow Paper
		"status":            status,
		"cumulativeGasUsed": hexutil.Uint64(cumulativeGasUsed),
		"logsBloom":         ethtypes.BytesToBloom(ethtypes.LogsBloom(logs)),
		"logs":              logs,

		// Implementation fields: These fields are added by geth when processing a transaction.
		// They are stored in the chain database.
		"transactionHash": hash,
		"contractAddress": nil,
		"gasUsed":         hexutil.Uint64(b.GetGasUsed(res, txData.GetGasPrice(), txData.GetGas())),

		// Inclusion information: These fields provide information about the inclusion of the
		// transaction corresponding to this receipt.
		"blockHash":        common.BytesToHash(resBlock.Block.Header.Hash()).Hex(),
		"blockNumber":      hexutil.Uint64(res.Height),
		"transactionIndex": hexutil.Uint64(res.EthTxIndex),

		// sender and receiver (contract or EOA) addreses
		"from": from,
		"to":   txData.GetTo(),
		"type": hexutil.Uint(ethMsg.AsTransaction().Type()),
	}

	if logs == nil {
		receipt["logs"] = [][]*ethtypes.Log{}
	}

	// If the ContractAddress is 20 0x0 bytes, assume it is not a contract creation
	if txData.GetTo() == nil {
		receipt["contractAddress"] = crypto.CreateAddress(from, txData.GetNonce())
	}

	if dynamicTx, ok := txData.(*evmtypes.DynamicFeeTx); ok {
		baseFee, err := b.BaseFee(blockRes)
		if err != nil {
			// tolerate the error for pruned node.
			b.logger.Error("fetch basefee failed, node is pruned?", "height", res.Height, "error", err)
		} else {
			receipt["effectiveGasPrice"] = hexutil.Big(*dynamicTx.EffectiveGasPrice(baseFee))
		}
	}

	return receipt, nil
}

// getPseudoTransactionReceipt creates a receipt for a pseudo-transaction
// with bridging info.
func (b *Backend) getPseudoTransactionReceipt(
	txHash common.Hash,
	txResult *types.TxResult,
	blockResult *tmrpctypes.ResultBlock,
) map[string]interface{} {
	return map[string]interface{}{
		// Consensus fields.
		"status":            hexutil.Uint(ethtypes.ReceiptStatusSuccessful),
		"cumulativeGasUsed": hexutil.Uint64(0),
		"logsBloom":         ethtypes.Bloom{},
		"logs":              []*ethtypes.Log{},

		// Implementation fields
		"transactionHash": txHash,
		"contractAddress": nil,
		"gasUsed":         hexutil.Uint64(0),

		// Inclusion information.
		"blockHash":        common.BytesToHash(blockResult.Block.Header.Hash()).Hex(),
		"blockNumber":      hexutil.Uint64(txResult.Height),
		"transactionIndex": hexutil.Uint64(txResult.EthTxIndex),

		// Sender and receiver (contract or EOA) addresses.
		"from": common.Address{},
		"to":   common.HexToAddress(assetsbridge.EvmAddress),
		"type": hexutil.Uint(0),
	}
}

// GetTransactionByBlockHashAndIndex returns the transaction identified by hash and index.
func (b *Backend) GetTransactionByBlockHashAndIndex(hash common.Hash, idx hexutil.Uint) (*rpctypes.RPCTransaction, error) {
	b.logger.Debug("eth_getTransactionByBlockHashAndIndex", "hash", hash.Hex(), "index", idx)

	signClient, ok := b.clientCtx.Client.(tmrpcclient.SignClient)
	if !ok {
		return nil, errors.New("unexpected RPC client type")
	}

	block, err := signClient.BlockByHash(b.ctx, hash.Bytes())
	if err != nil {
		b.logger.Debug("block not found", "hash", hash.Hex(), "error", err.Error())
		return nil, nil
	}

	if block.Block == nil {
		b.logger.Debug("block not found", "hash", hash.Hex())
		return nil, nil
	}

	return b.GetTransactionByBlockAndIndex(block, idx)
}

// GetTransactionByBlockNumberAndIndex returns the transaction identified by number and index.
func (b *Backend) GetTransactionByBlockNumberAndIndex(blockNum rpctypes.BlockNumber, idx hexutil.Uint) (*rpctypes.RPCTransaction, error) {
	b.logger.Debug("eth_getTransactionByBlockNumberAndIndex", "number", blockNum, "index", idx)

	block, err := b.TendermintBlockByNumber(blockNum)
	if err != nil {
		b.logger.Debug("block not found", "height", blockNum.Int64(), "error", err.Error())
		return nil, nil
	}

	if block.Block == nil {
		b.logger.Debug("block not found", "height", blockNum.Int64())
		return nil, nil
	}

	return b.GetTransactionByBlockAndIndex(block, idx)
}

// GetTxByEthHash uses `/tx_query` to find transaction by ethereum tx hash
// TODO: Don't need to convert once hashing is fixed on Tendermint
// https://github.com/tendermint/tendermint/issues/6539
func (b *Backend) GetTxByEthHash(hash common.Hash) (*types.TxResult, error) {
	if b.indexer != nil {
		// When getting the transaction result via the custom indexer, we do not
		// have to make any adjustments as the indexer is aware of the possibility
		// of a pseudo-transaction in the block and accounts for that.
		return b.indexer.GetByTxHash(hash)
	}

	// Fallback to Tendermint tx indexer - it is possible that the transaction
	// ETH index may need to be adjusted.
	query := fmt.Sprintf("%s.%s='%s'", evmtypes.TypeMsgEthereumTx, evmtypes.AttributeKeyEthereumTxHash, hash.Hex())
	txResult, err := b.queryTendermintTxIndexer(query, func(txs *rpctypes.ParsedTxs) *rpctypes.ParsedTx {
		return txs.GetTxByHash(hash)
	})
	if err != nil {
		return nil, errorsmod.Wrapf(err, "GetTxByEthHash %s", hash.Hex())
	}

	block, err := b.TendermintBlockByNumber(rpctypes.BlockNumber(txResult.Height))
	if err != nil {
		return nil, errorsmod.Wrapf(err, "GetTxByEthHash %s", hash.Hex())
	}

	// Check if the block contains a pseudo-transaction.
	pseudoTxResult := b.GetPseudoTransactionResult(block)

	// Adjust the ETH transaction index to account for the pseudo-transaction.
	if pseudoTxResult != nil {
		txResult.EthTxIndex++
	}

	return txResult, nil
}

// queryTendermintTxIndexer query tx in tendermint tx indexer
func (b *Backend) queryTendermintTxIndexer(query string, txGetter func(*rpctypes.ParsedTxs) *rpctypes.ParsedTx) (*types.TxResult, error) {
	resTxs, err := b.clientCtx.Client.TxSearch(b.ctx, query, false, nil, nil, "")
	if err != nil {
		return nil, err
	}
	if len(resTxs.Txs) == 0 {
		return nil, errors.New("ethereum tx not found")
	}
	txResult := resTxs.Txs[0]
	if !rpctypes.TxSuccessOrExceedsBlockGasLimit(&txResult.TxResult) {
		return nil, errors.New("invalid ethereum tx")
	}

	var tx sdk.Tx
	if txResult.TxResult.Code != 0 {
		// it's only needed when the tx exceeds block gas limit
		tx, err = b.clientCtx.TxConfig.TxDecoder()(txResult.Tx)
		if err != nil {
			return nil, fmt.Errorf("invalid ethereum tx")
		}
	}

	return rpctypes.ParseTxIndexerResult(txResult, tx, txGetter)
}

// getPseudoTransactionResult attempts to parse the pseudo-transaction from
// the given block. If the block does not contain the pseudo-transaction, it
// returns nil.
func (b *Backend) GetPseudoTransactionResult(
	block *tmrpctypes.ResultBlock,
) *types.TxResult {
	if block == nil {
		return nil
	}

	if len(block.Block.Txs) == 0 {
		// There are no transactions in the block. Therefore there is no
		// pseudo-transaction.
		return nil
	}

	// The pseudo-transaction can only be located at index `0`.
	injectedTx, isPseudoTx := indexer.ParsePseudoTransaction(block.Block.Txs[0])
	if !isPseudoTx {
		return nil
	}

	if injectedTx == nil || len(injectedTx.AssetsLockedEvents) == 0 {
		// The pseudo-transaction does not contain any `AssetsLocked` events.
		// Do not consider it a valid pseudo-transaction.
		return nil
	}

	serializedTx := b.clientCtx.Codec.MustMarshal(injectedTx)

	extraData := append(
		[]byte{indexer.BridgingInfoDescriptor},
		serializedTx...,
	)

	return &types.TxResult{
		Height:     block.Block.Height,
		TxIndex:    0,
		EthTxIndex: 0,
		ExtraData:  extraData,
	}
}

// GetTransactionByBlockAndIndex is the common code shared by `GetTransactionByBlockNumberAndIndex` and `GetTransactionByBlockHashAndIndex`.
func (b *Backend) GetTransactionByBlockAndIndex(block *tmrpctypes.ResultBlock, idx hexutil.Uint) (*rpctypes.RPCTransaction, error) {
	blockRes, err := b.TendermintBlockResultByNumber(&block.Block.Height)
	if err != nil {
		return nil, nil
	}

	pseudoTxResult := b.GetPseudoTransactionResult(block)
	if idx == 0 && pseudoTxResult != nil {
		// There is a pseudo-transaction in the block and it is requested.
		return b.getPseudoTransaction(pseudoTxResult, block)
	}

	// Function for getting the result for regular transactions either from
	// the custom indexer or Tendermint indexer.
	getTxByTxIndex := func() (*types.TxResult, error) {
		if b.indexer != nil {
			// The custom indexer is set. It stores both pseudo-transactions
			// and regular transactions. We do not have to adjust the index
			// when querying the custom indexer.
			return b.indexer.GetByBlockAndIndex(block.Block.Height, int32(idx))
		}

		// If the custom indexer is not set, query the Tendermint indexer.
		// As the Tendermint indexer is not aware of pseudo-transactions, we
		// may need to adjust the requested index: if the block contains a
		// valid pseudo-transaction, we must decrement the index.
		adjustedIndex := uint(idx)
		if pseudoTxResult != nil {
			// Notice that we can safely decrement the transaction index variable.
			// Due to the earlier check for pseudo-transaction, we are sure the
			// passed index is positive.
			adjustedIndex--
		}

		query := fmt.Sprintf("tx.height=%d AND %s.%s=%d",
			block.Block.Height, evmtypes.TypeMsgEthereumTx,
			evmtypes.AttributeKeyTxIndex, adjustedIndex,
		)

		txResult, err := b.queryTendermintTxIndexer(query, func(txs *rpctypes.ParsedTxs) *rpctypes.ParsedTx {
			return txs.GetTxByTxIndex(int(adjustedIndex)) // #nosec G701 -- checked for int overflow already
		})
		if err != nil {
			return nil, errorsmod.Wrapf(err, "GetTransactionByBlockAndIndex %d %d", block.Block.Height, adjustedIndex)
		}

		return txResult, nil
	}

	var msg *evmtypes.MsgEthereumTx

	res, err := getTxByTxIndex()
	if err == nil {
		tx, err := b.clientCtx.TxConfig.TxDecoder()(block.Block.Txs[res.TxIndex])
		if err != nil {
			b.logger.Debug("invalid ethereum tx", "height", block.Block.Header, "index", idx)
			return nil, nil
		}

		var ok bool
		// msgIndex is inferred from tx events, should be within bound.
		msg, ok = tx.GetMsgs()[res.MsgIndex].(*evmtypes.MsgEthereumTx)
		if !ok {
			b.logger.Debug("invalid ethereum tx", "height", block.Block.Header, "index", idx)
			return nil, nil
		}
	} else {
		// If it was impossible to get the transaction result from custom or
		// Tendermint indexer, find the transaction by iterating over the block's
		// Eth messages. If the block contains a pseudo-transaction, we must
		// adjust the index.
		i := uint(idx) // #nosec G701
		if pseudoTxResult != nil {
			// Notice that we can safely decrement the transaction index variable.
			// Due to the earlier check for pseudo-transaction, we are sure the
			// passed index is positive.
			i--
		}
		ethMsgs := b.EthMsgsFromTendermintBlock(block, blockRes)
		if i >= uint(len(ethMsgs)) {
			b.logger.Debug("block txs index out of bound", "index", i)
			return nil, nil
		}

		msg = ethMsgs[i]
	}

	baseFee, err := b.BaseFee(blockRes)
	if err != nil {
		// handle the error for pruned node.
		b.logger.Error("failed to fetch Base Fee from prunned block. Check node prunning configuration", "height", block.Block.Height, "error", err)
	}

	height := uint64(block.Block.Height) // #nosec G701 -- checked for int overflow already
	index := uint64(idx)                 // #nosec G701 -- checked for int overflow already
	return rpctypes.NewTransactionFromMsg(
		msg,
		common.BytesToHash(block.Block.Hash()),
		height,
		index,
		baseFee,
		b.chainID,
	)
}
