package backend

import (
	"context"
	"encoding/json"

	"github.com/pkg/errors"

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
	bnhBz, err := json.Marshal(effectiveBnh)
	if err != nil {
		return nil, err
	}

	blockNr, err := b.BlockNumberFromTendermint(effectiveBnh)
	if err != nil {
		return nil, err
	}
	header, err := b.TendermintBlockByNumber(blockNr)
	if err != nil {
		return nil, errors.New("header not found")
	}

	timeout := b.RPCEVMTimeout()

	req := &evmtypes.SimulateV1Request{
		Opts:              optsBz,
		BlockNumberOrHash: bnhBz,
		GasCap:            b.RPCGasCap(),
		ProposerAddress:   sdk.ConsAddress(header.Block.ProposerAddress),
		ChainId:           b.chainID.Int64(),
		TimeoutMs:         timeout.Milliseconds(),
	}

	ctx := rpctypes.ContextWithHeight(blockNr.Int64())
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
