package abci

import (
	cmtabci "github.com/cometbft/cometbft/abci/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// TODO: Implementation and documentation.

type VoteExtensionHandler struct {}

func NewVoteExtensionHandler() *VoteExtensionHandler {
	return &VoteExtensionHandler{}
}

func (veh *VoteExtensionHandler) ExtendVoteHandler() sdk.ExtendVoteHandler {
	return func(
		ctx sdk.Context,
		req *cmtabci.RequestExtendVote,
	) (*cmtabci.ResponseExtendVote, error) {
		return &cmtabci.ResponseExtendVote{VoteExtension: []byte{}}, nil
	}
}

func (veh *VoteExtensionHandler) VerifyVoteExtensionHandler() sdk.VerifyVoteExtensionHandler {
	return func(
		ctx sdk.Context,
		req *cmtabci.RequestVerifyVoteExtension,
	) (*cmtabci.ResponseVerifyVoteExtension, error) {
		return &cmtabci.ResponseVerifyVoteExtension{Status: cmtabci.ResponseVerifyVoteExtension_ACCEPT}, nil
	}
}