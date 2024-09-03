package abci

import (
	"fmt"
	connectve "github.com/skip-mev/connect/v2/abci/ve"

	"cosmossdk.io/log"

	cmtabci "github.com/cometbft/cometbft/abci/types"
	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/mezo-org/mezod/app/abci/types"
	bridgeabci "github.com/mezo-org/mezod/x/bridge/abci"
)

// VoteExtensionPart is an enumeration of the different parts of the app-level
// vote extension.
type VoteExtensionPart uint32

const (
	// VoteExtensionPartUnknown is an unknown vote extension part.
	VoteExtensionPartUnknown VoteExtensionPart = iota
	// VoteExtensionPartBridge is the part of the vote extension that
	// is specific to the Bitcoin bridge.
	VoteExtensionPartBridge
	// VoteExtensionPartConnect is the part of the vote extension that is specific to the Connect oracle.
	VoteExtensionPartConnect
)

// String returns a string representation of the vote extension part.
func (vep VoteExtensionPart) String() string {
	switch vep {
	case VoteExtensionPartBridge:
		return "bridge"
	default:
		return fmt.Sprintf("unknown: %d", vep)
	}
}

// IVoteExtensionHandler is an interface representing a vote extension handler.
type IVoteExtensionHandler interface {
	// ExtendVoteHandler returns the handler for the ExtendVote ABCI request.
	ExtendVoteHandler() sdk.ExtendVoteHandler
	// VerifyVoteExtensionHandler returns the handler for the VerifyVoteExtension
	// ABCI request.
	VerifyVoteExtensionHandler() sdk.VerifyVoteExtensionHandler
}

// VoteExtensionHandler is a handler for the ExtendVote and
// VerifyVoteExtension ABCI requests. It is designed to be used with
// multiple sub-handlers, each responsible for a specific part of the
// vote extension. The handler itself is responsible for combining the
// results of the sub-handlers into a single vote extension.
type VoteExtensionHandler struct {
	logger      log.Logger
	subHandlers map[VoteExtensionPart]IVoteExtensionHandler
}

// NewVoteExtensionHandler creates a new VoteExtensionHandler instance.
func NewVoteExtensionHandler(
	logger log.Logger,
	bridgeSubHandler *bridgeabci.VoteExtensionHandler,
	connectVEHandler *connectve.VoteExtensionHandler,

) *VoteExtensionHandler {
	subHandlers := map[VoteExtensionPart]IVoteExtensionHandler{
		VoteExtensionPartBridge:  bridgeSubHandler,
		VoteExtensionPartConnect: connectVEHandler,
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
// and combines the results into a single app-level vote extension. If any
// sub-handler fails, its error is logged and the part of the vote extension
// it was responsible for is not included in the final app-level vote extension.
// This way, handlers can fail independently of each other. A sub-handler
// can also return empty bytes for its part of the vote extension and such an
// empty part is included in the final app-level vote extension. This way
// presence of the given part in the app-level vote extension indicates
// the success/failure of the corresponding sub-handler and may be used
// for debugging purposes.
//
// Dev note: It is fine to return a nil response and an error from this
// function in case of failure. Cosmos SDK will log the error and return an
// empty app-level vote extension to the CometBFT engine.
//
// Invariants summary:
//  1. Sub-handler's response is considered valid if no error is returned,
//     regardless of whether the vote extension part being part of the response
//     is empty or not.
//  2. Returns an error if all sub-handlers return an invalid response.
//  3. Returns a non-empty app-level vote extension if at least one sub-handler
//     returns a valid response.
//  4. Returns an error if the app-level vote extension cannot be marshaled.
func (veh *VoteExtensionHandler) ExtendVoteHandler() sdk.ExtendVoteHandler {
	return func(
		ctx sdk.Context,
		req *cmtabci.RequestExtendVote,
	) (*cmtabci.ResponseExtendVote, error) {
		voteExtensionParts := make(map[uint32][]byte)

		// TODO: Consider running sub-handlers concurrently to speed up execution.
		//
		// TODO: Consider changing logging to debug level once this code matures.

		for part, subHandler := range veh.subHandlers {
			veh.logger.Info(
				"running sub-handler to extend vote",
				"height", req.Height,
				"part", part,
			)

			// Trigger the ExtendVote sub-handler for the given vote extension part.
			res, err := subHandler.ExtendVoteHandler()(ctx, req)
			if err != nil {
				// Just log the error and continue execution in case there
				// are other sub-handlers in the queue. We do not want
				// make sub-handlers dependent on each other.
				veh.logger.Error(
					"sub-handler failed to extend vote",
					"height", req.Height,
					"part", part,
					"err", err,
				)
				continue
			}

			veh.logger.Info(
				"sub-handler extended vote",
				"height", req.Height,
				"part", part,
				"part_byte_length", len(res.VoteExtension),
			)

			// Note that len(res.VoteExtension) may be 0 and this is a valid
			// case. Such an empty part is still included in the app-level
			// vote extension. Missing part indicates an error of the
			// given sub-handler.
			voteExtensionParts[uint32(part)] = res.VoteExtension
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
			// If marshaling fails, we cannot recover, so return an error.
			return nil, fmt.Errorf("failed to marshal vote extension: %w", err)
		}

		veh.logger.Info(
			"vote extended",
			"height", req.Height,
			"ve_byte_length", len(voteExtensionBytes),
		)

		return &cmtabci.ResponseExtendVote{VoteExtension: voteExtensionBytes}, nil
	}
}

// VerifyVoteExtensionHandler returns the handler for the VerifyVoteExtension
// ABCI request.
//
// Dev note: It is fine to return a nil response and an error from this
// function in case of failure. Cosmos SDK will log the error and REJECT
// the app-level vote extension. Rejecting the vote extension is a serious
// event and has liveness implications for the CometBFT engine. Make sure you
// know what you are doing before returning an error from this function.
//
// Invariants summary:
//  1. Accepts empty app-level vote extension.
//  2. Rejects app-level vote extension that cannot be unmarshaled.
//  3. Rejects app-level vote extension with incorrect height.
//  4. Rejects app-level vote extension with no parts.
//  5. Accepts app-level vote extension only if each part corresponds to
//     a known sub-handler.
//  6. Accepts app-level vote extension only if each part is accepted by the
//     corresponding sub-handler.
//
// TODO: Currently, this function either ACCEPT the vote extension or return
// an error - it does not REJECT anything explicitly. This is fine as
// Cosmos SDK handles the error gracefully by logging it on the error
// level and rejecting the vote extension. This is fine for now, but we
// may consider a more granular approach and start distinguish between
// a case when the validation completes successfully but the vote extension
// is invalid and a case when the validation fails (sub-handlers can
// distinguish between these cases by returning either REJECT + error
// holding the reason or nil + error, respectively). Having this distinction
// could allow us to log rejections on warn level and leave log errors
// for actual errors. The main motivation here is validator experience.
// Error logs should ideally lead to action items for the given validator
// while rejection warnings should stay informative and highlight potential
// misbehavior of other validators.
func (veh *VoteExtensionHandler) VerifyVoteExtensionHandler() sdk.VerifyVoteExtensionHandler {
	return func(
		ctx sdk.Context,
		req *cmtabci.RequestVerifyVoteExtension,
	) (*cmtabci.ResponseVerifyVoteExtension, error) {
		from := sdk.ConsAddress(req.ValidatorAddress).String()

		veh.logger.Debug(
			"verifying vote extension",
			"height", req.Height,
			"from", from,
		)

		if len(req.VoteExtension) == 0 {
			// Accept empty vote extensions. This is necessary given that
			// Cosmos SDK turns this handler's ExtendVote errors into empty
			// vote extensions.
			veh.logger.Debug(
				"accepted empty vote extension",
				"height", req.Height,
				"from", from,
			)

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

		for partUint, partBytes := range voteExtension.Parts {
			part := VoteExtensionPart(partUint)

			subHandler, ok := veh.subHandlers[part]
			if !ok {
				// Make sure the vote extension part is recognized. If not,
				// reject the vote extension as something is clearly wrong.
				return nil, fmt.Errorf("unknown vote extension part: %d", part)
			}

			veh.logger.Debug(
				"running sub-handler to verify vote extension",
				"height", req.Height,
				"from", from,
				"part", part,
			)

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

			veh.logger.Debug(
				"sub-handler verified and accepted vote extension",
				"height", req.Height,
				"from", from,
				"part", part,
			)
		}

		veh.logger.Debug(
			"accepted vote extension",
			"height", req.Height,
			"from", from,
		)

		return &cmtabci.ResponseVerifyVoteExtension{
			Status: cmtabci.ResponseVerifyVoteExtension_ACCEPT,
		}, nil
	}
}

// isVoteExtensionsEnabled returns true if the vote extensions are enabled
// at the given height.
func isVoteExtensionsEnabled(ctx sdk.Context, height int64) bool {
	cp := ctx.ConsensusParams()
	return cp.Abci != nil &&
		cp.Abci.VoteExtensionsEnableHeight > 0 &&
		height > cp.Abci.VoteExtensionsEnableHeight
}

// VoteExtensionDecomposer returns a function that decomposes a composite
// app-level vote extension and returns the given part.
func VoteExtensionDecomposer(part VoteExtensionPart) func([]byte) ([]byte, error) {
	return func(voteExtensionBytes []byte) ([]byte, error) {
		var voteExtension types.VoteExtension
		if err := voteExtension.Unmarshal(voteExtensionBytes); err != nil {
			return nil, fmt.Errorf(
				"failed to unmarshal composite vote extension: %w",
				err,
			)
		}

		return voteExtension.Parts[uint32(part)], nil
	}
}
