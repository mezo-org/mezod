package abci

import (
	"cosmossdk.io/log"
	cmtabci "github.com/cometbft/cometbft/abci/types"
	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/mezo-org/mezod/app/abci/types"
	bridgeabci "github.com/mezo-org/mezod/x/bridge/abci"
)

// VoteExtensionHandler is a handler for the ExtendVote and
// VerifyVoteExtension ABCI requests. It is designed to be used with
// multiple sub-handlers, each responsible for a specific part of the
// vote extension. The handler itself is responsible for combining the
// results of the sub-handlers into a single vote extension.
type VoteExtensionHandler struct {
	logger log.Logger
	bridgeSubHandler *bridgeabci.VoteExtensionHandler
}

// NewVoteExtensionHandler creates a new VoteExtensionHandler instance.
func NewVoteExtensionHandler(
	logger log.Logger,
	bridgeSubHandler *bridgeabci.VoteExtensionHandler,
) *VoteExtensionHandler {
	return &VoteExtensionHandler{
		logger: logger,
		bridgeSubHandler: bridgeSubHandler,
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
func (veh *VoteExtensionHandler) ExtendVoteHandler() sdk.ExtendVoteHandler {
	return func(
		ctx sdk.Context,
		req *cmtabci.RequestExtendVote,
	) (*cmtabci.ResponseExtendVote, error) {
		var voteExtension types.VoteExtension

		// TODO: Consider running sub-handlers concurrently to speed up execution.

		// Trigger the ExtendVote sub-handler for the Bitcoin bridge.
		if res, err := veh.bridgeSubHandler.ExtendVoteHandler()(ctx, req); err != nil {
			// Just log the error and continue execution in case there
			// are other sub-handlers.
			veh.logger.Error("bridge handler failed to extend vote", "err", err)
		} else {
			voteExtension.BridgeVoteExtension = res.VoteExtension
		}

		voteExtensionBytes, err := voteExtension.Marshal()
		if err != nil {
			// If marshalling fails, return an empty vote extension and
			// the error to indicate failure of the whole handler.
			veh.logger.Error("failed to marshal vote extension", "err", err)
			return &cmtabci.ResponseExtendVote{VoteExtension: []byte{}}, err
		}

		return &cmtabci.ResponseExtendVote{VoteExtension: voteExtensionBytes}, nil
	}
}

// VerifyVoteExtensionHandler returns the handler for the VerifyVoteExtension
// ABCI request.
func (veh *VoteExtensionHandler) VerifyVoteExtensionHandler() sdk.VerifyVoteExtensionHandler {
	return func(
		ctx sdk.Context,
		req *cmtabci.RequestVerifyVoteExtension,
	) (*cmtabci.ResponseVerifyVoteExtension, error) {
		// TODO: Implement verification logic.
		return &cmtabci.ResponseVerifyVoteExtension{Status: cmtabci.ResponseVerifyVoteExtension_ACCEPT}, nil
	}
}