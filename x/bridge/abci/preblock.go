package abci

import (
	"fmt"

	cmtabci "github.com/cometbft/cometbft/abci/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/mezo-org/mezod/x/bridge/abci/types"
	"github.com/mezo-org/mezod/x/bridge/keeper"
)

// PreBlockHandler is the bridge-specific pre-block handler (part of the
// FinalizeBlock ABCI request)
type PreBlockHandler struct {
	bridgeKeeper keeper.Keeper
}

// NewPreBlockHandler returns a new PreBlockHandler.
func NewPreBlockHandler(bridgeKeeper keeper.Keeper) *PreBlockHandler {
	return &PreBlockHandler{
		bridgeKeeper: bridgeKeeper,
	}
}

// PreBlocker returns the pre-block handler (part of the FinalizeBlock ABCI request).
// This function:
//   - Extracts the sequence of canonical AssetsLocked events from the injected
//     pseudo-transaction. Minimum validation is performed to avoid unexpected
//     panics but no error is expected here as this handler is invoked after
//     the proposal phase which guarantees the pseudo-transaction is well-formed
//     and contain a valid sequence of canonical AssetsLocked events.
//   - Accepts the AssetsLocked events by calling the bridge keeper's AcceptAssetsLocked
//     method. This call updates the internal state of the bridge module to reflect
//     the new AssetsLocked events and mints new BTC as result.
//
// Dev note: In case of success, this function should return a non-nil pointer value
// (&sdk.ResponsePreBlock{}) as ResponsePreBlock and a nil error. The specific
// value of ResponsePreBlock is not relevant for the app-level handler but, we
// try to adhere to the Go best practices and return a non-nil pointer value
// in case of success. Conversely, in case of failure, this function should return
// a nil ResponsePreBlock and a non-nil error. BEWARE: An error returned from this
// function will be bubbled up by the app-level handler and will cause the FinalizeBlock
// ABCI request to fail. This, in turn, will lead to an unrecoverable consensus failure.
// Make sure you know what you are doing before returning an error.
func (pbh *PreBlockHandler) PreBlocker() sdk.PreBlocker {
	return func(
		ctx sdk.Context,
		req *cmtabci.RequestFinalizeBlock,
	) (*sdk.ResponsePreBlock, error) {
		if len(req.Txs) == 0 {
			// The app-level handler always passes a transaction vector
			// with at least one element (representing the possibly empty
			// bridge-specific pseudo-transaction). If the vector is empty,
			// something went wrong upstream and we cannot recover so, return
			// an error.
			return nil, fmt.Errorf("empty transaction vector in the block")
		}

		if len(req.Txs[0]) == 0 {
			// As said above, the app-level handler always passes a transaction
			// vector with at least one element (representing the possibly empty
			// bridge-specific pseudo-transaction). The first element can be an
			// empty byte slice if the bridge-level PrepareProposal handler
			// decided to not inject the pseudo-transaction. In this case, we
			// do not need to do anything, and we can short-circuit the processing.
			return &sdk.ResponsePreBlock{}, nil
		}

		var injectedTx types.InjectedTx
		if err := injectedTx.Unmarshal(req.Txs[0]); err != nil {
			// If unmarshaling of the injected pseudo-transaction fails,
			// we cannot recover, so return an error.
			return nil, fmt.Errorf("failed to unmarshal injected tx: %w", err)
		}

		// If the pseudo-transaction is present but holds an empty slice of
		// AssetsLocked events, we error out as this case is illegal with
		// the current proposal phase implementation.
		if len(injectedTx.AssetsLockedEvents) == 0 {
			return nil, fmt.Errorf("injected tx does not contain AssetsLocked events")
		}

		err := pbh.bridgeKeeper.AcceptAssetsLocked(ctx, injectedTx.AssetsLockedEvents)
		if err != nil {
			return nil, fmt.Errorf("cannot accept AssetsLocked events: %w", err)
		}

		return &sdk.ResponsePreBlock{}, nil
	}
}
