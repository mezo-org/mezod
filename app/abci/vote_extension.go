package abci

import (
	"cosmossdk.io/log"
	"fmt"
	cmtabci "github.com/cometbft/cometbft/abci/types"
	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/mezo-org/mezod/app/abci/types"
	bridgeabci "github.com/mezo-org/mezod/x/bridge/abci"
)

// VoteExtensionPart is an enumeration of the different parts of the app-level
// vote extension.
type VoteExtensionPart = uint32

const (
	// VoteExtensionPartUnknown is an unknown vote extension part.
	VoteExtensionPartUnknown VoteExtensionPart = iota
	// VoteExtensionPartBridge is the part of the vote extension that
	// is specific to the Bitcoin bridge.
	VoteExtensionPartBridge
)

// VoteExtensionHandler is a handler for the ExtendVote and
// VerifyVoteExtension ABCI requests. It is designed to be used with
// multiple sub-handlers, each responsible for a specific part of the
// vote extension. The handler itself is responsible for combining the
// results of the sub-handlers into a single vote extension.
type VoteExtensionHandler struct {
	logger      log.Logger
	subHandlers map[VoteExtensionPart]*bridgeabci.VoteExtensionHandler
}

// NewVoteExtensionHandler creates a new VoteExtensionHandler instance.
func NewVoteExtensionHandler(
	logger log.Logger,
	bridgeSubHandler *bridgeabci.VoteExtensionHandler,
) *VoteExtensionHandler {
	subHandlers := map[VoteExtensionPart]*bridgeabci.VoteExtensionHandler{
		VoteExtensionPartBridge: bridgeSubHandler,
	}

	return &VoteExtensionHandler{
		logger:      logger,
		subHandlers: subHandlers,
	}
}

// SetHandlers sets the ExtendVote and VerifyVoteExtension handlers on the
// provided base app.
func (veh *VoteExtensionHandler) SetHandlers(baseApp *baseapp.BaseApp) {
	baseApp.SetExtendVoteHandler(veh.ExtendVoteHandler())
	baseApp.SetVerifyVoteExtensionHandler(veh.VerifyVoteExtensionHandler())
}

// ExtendVoteHandler returns the handler for the ExtendVote ABCI request.
// It triggers the ExtendVote sub-handlers for each part of the vote extension
// and combines the results into a single vote extension. If any sub-handler
// fails, its error is logged and the part of the vote extension it was
// responsible for is left empty. This way, handlers can fail independently
// of each other.
//
// Dev note: It is fine to return a nil response and an error from this
// function in case of failure. Cosmos SDK will log the error and return an
// empty vote extension to the CometBFT engine.
func (veh *VoteExtensionHandler) ExtendVoteHandler() sdk.ExtendVoteHandler {
	return func(
		ctx sdk.Context,
		req *cmtabci.RequestExtendVote,
	) (*cmtabci.ResponseExtendVote, error) {
		voteExtensionParts := make(map[VoteExtensionPart][]byte)

		// TODO: Consider running sub-handlers concurrently to speed up execution.

		for part, subHandler := range veh.subHandlers {
			// Trigger the ExtendVote sub-handler for the given vote extension part.
			res, err := subHandler.ExtendVoteHandler()(ctx, req)
			if err != nil {
				// Just log the error and continue execution in case there
				// are other sub-handlers in the queue. We do not want
				// make sub-handlers dependent on each other.
				veh.logger.Error(
					"sub-handler failed to extend vote with part",
					"part", part,
					"err", err,
				)
				continue
			}

			voteExtensionParts[part] = res.VoteExtension
		}

		if len(voteExtensionParts) == 0 {
			// Short-circuit if all sub-handlers failed.
			return nil, fmt.Errorf("all sub-handlers failed to extend vote")
		}

		// Construct the final vote extension from the parts.
		voteExtension := types.VoteExtension{
			Height: req.Height,
			Parts:  voteExtensionParts,
		}
		// Marshal the vote extension into bytes.
		voteExtensionBytes, err := voteExtension.Marshal()
		if err != nil {
			// If marshalling fails, we cannot recover, so return an error.
			return nil, fmt.Errorf("failed to marshal vote extension: %w", err)
		}

		return &cmtabci.ResponseExtendVote{VoteExtension: voteExtensionBytes}, nil
	}
}

// VerifyVoteExtensionHandler returns the handler for the VerifyVoteExtension
// ABCI request.
//
// Dev note: It is fine to return a nil response and an error from this
// function in case of failure. Cosmos SDK will log the error and REJECT
// the vote extension. Rejecting the vote extension is a serious event
// and has liveness implications for the CometBFT engine. Make sure you
// know what you are doing before returning an error from this function.
func (veh *VoteExtensionHandler) VerifyVoteExtensionHandler() sdk.VerifyVoteExtensionHandler {
	return func(
		ctx sdk.Context,
		req *cmtabci.RequestVerifyVoteExtension,
	) (*cmtabci.ResponseVerifyVoteExtension, error) {
		if len(req.VoteExtension) == 0 {
			// Accept empty vote extensions. This is necessary given that
			// Cosmos SDK turns ExtendVote errors into empty vote extensions.
			return &cmtabci.ResponseVerifyVoteExtension{
				Status: cmtabci.ResponseVerifyVoteExtension_ACCEPT,
			}, nil
		}

		// Unmarshal the vote extension.
		var voteExtension types.VoteExtension
		if err := voteExtension.Unmarshal(req.VoteExtension); err != nil {
			// If the vote extension cannot be unmarshalled, we cannot recover.
			return nil, fmt.Errorf("failed to unmarshal vote extension: %w", err)
		}

		if voteExtension.Height != req.Height {
			return nil, fmt.Errorf(
				"vote extension height does not match request height; expected: %d, got: %d",
				req.Height,
				voteExtension.Height,
			)
		}

		if len(voteExtension.Parts) == 0 {
			// Reject vote extensions with no parts. This is done in order to
			// ensure parity with the ExtendVoteHandler, which always produces
			// a vote extension with at least one part.
			return nil, fmt.Errorf("vote extension has no parts")
		}

		for part, partBytes := range voteExtension.Parts {
			subHandler, ok := veh.subHandlers[part]
			if !ok {
				// Make sure the vote extension part is recognized. If not,
				// reject the vote extension as something is clearly wrong.
				return nil, fmt.Errorf("unknown vote extension part: %d", part)
			}

			res, err := subHandler.VerifyVoteExtensionHandler()(
				ctx,
				&cmtabci.RequestVerifyVoteExtension{
					Hash:             req.Hash,
					ValidatorAddress: req.ValidatorAddress,
					Height:           req.Height,
					VoteExtension:    partBytes,
				},
			)
			if err != nil {
				// If a sub-handler fails to verify its part, reject the whole vote extension.
				return nil, fmt.Errorf(
					"sub-handler failed to verify vote extension part %v: %w",
					part,
					err,
				)
			}
			if res.Status != cmtabci.ResponseVerifyVoteExtension_ACCEPT {
				// If a sub-handler rejects its part, reject the whole vote extension.
				return nil, fmt.Errorf(
					"sub-handler rejected vote extension part %v",
					part,
				)
			}
		}

		return &cmtabci.ResponseVerifyVoteExtension{
			Status: cmtabci.ResponseVerifyVoteExtension_ACCEPT,
		}, nil
	}
}