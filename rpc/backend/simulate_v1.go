package backend

import (
	"context"
	"encoding/json"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	rpctypes "github.com/mezo-org/mezod/rpc/types"
	evmtypes "github.com/mezo-org/mezod/x/evm/types"
)

// SimulateV1 runs `eth_simulateV1`. Timeout wiring mirrors DoCall; the
// eth_simulateV1 path inherits the same RPCEVMTimeout semantics.
func (b *Backend) SimulateV1(
	opts evmtypes.SimOpts,
	blockNrOrHash *rpctypes.BlockNumberOrHash,
) ([]*evmtypes.SimBlockResult, error) {
	optsBz, err := json.Marshal(opts)
	if err != nil {
		return nil, err
	}

	latest := rpctypes.EthLatestBlockNumber
	effectiveBnh := rpctypes.BlockNumberOrHash{BlockNumber: &latest}
	if blockNrOrHash != nil && (blockNrOrHash.BlockNumber != nil || blockNrOrHash.BlockHash != nil) {
		effectiveBnh = *blockNrOrHash
	}

	blockNr, err := b.BlockNumberFromTendermint(effectiveBnh)
	if err != nil {
		return nil, err
	}
	header, err := b.TendermintBlockByNumber(blockNr)
	if err != nil {
		return nil, fmt.Errorf("header not found: %w", err)
	}

	// Resolve the sentinel / hash form to a concrete height before
	// marshaling so the keeper's anchor validator compares the same
	// numeric value it anchors ctx at. BlockNumber's JSON round-trip
	// does not preserve sentinels, so emitting them here would
	// misfire the anchor check.
	resolvedBn := rpctypes.BlockNumber(header.Block.Height)
	resolvedBnh := rpctypes.BlockNumberOrHash{BlockNumber: &resolvedBn}
	bnhBz, err := json.Marshal(resolvedBnh)
	if err != nil {
		return nil, err
	}

	// Use the canonical CometBFT block hash so the envelope's
	// parentHash agrees with eth_getBlockByNumber. FormatBlock — the
	// path eth_getBlockByNumber takes — surfaces the Tendermint
	// header hash, not the hash of an Eth-formatted shadow header, so
	// anything derived via EthHeaderFromTendermint().Hash() is a
	// different scheme entirely and will not match. The keeper cannot
	// resolve this hash on its own: at FinalizeBlock the SDK only
	// puts a truncated cmtproto.Header on ctx (lacking LastBlockID,
	// DataHash, and others), and it's exactly that truncated header
	// PoA's TrackHistoricalInfo persists — so neither ctx.BlockHeader()
	// nor stakingKeeper.GetHistoricalInfo can produce the canonical
	// hash CometBFT's block store carries. Forwarding it from the
	// rpc layer is the only way the two surfaces line up.
	baseBlockHash := header.Block.Hash().Bytes()

	timeout := b.RPCEVMTimeout()

	req := &evmtypes.SimulateV1Request{
		Opts:              optsBz,
		BlockNumberOrHash: bnhBz,
		GasCap:            b.RPCGasCap(),
		ProposerAddress:   sdk.ConsAddress(header.Block.ProposerAddress),
		ChainId:           b.chainID.Int64(),
		TimeoutMs:         timeout.Milliseconds(),
		BaseBlockHash:     baseBlockHash,
	}

	ctx := rpctypes.ContextWithHeight(header.Block.Height)
	var cancel context.CancelFunc
	if timeout > 0 {
		ctx, cancel = context.WithTimeout(ctx, timeout)
	} else {
		ctx, cancel = context.WithCancel(ctx)
	}
	defer cancel()

	res, err := b.queryClient.QueryClient.SimulateV1(ctx, req)
	if err != nil {
		return nil, err
	}

	// A structured error on the response carries a spec-reserved
	// JSON-RPC code verbatim; return it as *evmtypes.SimError so
	// geth's RPC server emits {code, message, data} through the
	// error-interface methods.
	if res.Error != nil {
		return nil, res.Error
	}

	if len(res.Result) == 0 {
		return []*evmtypes.SimBlockResult{}, nil
	}

	var out []*evmtypes.SimBlockResult
	if err := json.Unmarshal(res.Result, &out); err != nil {
		return nil, err
	}
	return out, nil
}
