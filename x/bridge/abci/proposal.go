package abci

import (
	"cosmossdk.io/log"
	"fmt"
	cmtabci "github.com/cometbft/cometbft/abci/types"
	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/mezo-org/mezod/x/bridge/abci/types"
	bridgekeeper "github.com/mezo-org/mezod/x/bridge/keeper"
	bridgetypes "github.com/mezo-org/mezod/x/bridge/types"
	"slices"
)

// VoteExtensionDecomposer is a function that decomposes a composite app-level
// vote extension and returns the part that is relevant to the bridge.
type VoteExtensionDecomposer func(compositeVoteExtensionBytes []byte) (
	voteExtensionBytes []byte,
	err error,
)

// ProposalHandler is the bridge-specific handler for the PrepareProposal and
// ProcessProposal ABCI requests.
type ProposalHandler struct {
	logger                  log.Logger
	valStore                bridgetypes.ValidatorStore
	voteExtensionDecomposer VoteExtensionDecomposer
	keeper                  bridgekeeper.Keeper
}

// NewProposalHandler creates a new ProposalHandler instance.
func NewProposalHandler(
	logger log.Logger,
	valStore bridgetypes.ValidatorStore,
	voteExtensionDecomposer VoteExtensionDecomposer,
	keeper bridgekeeper.Keeper,
) *ProposalHandler {
	return &ProposalHandler{
		logger:                  logger,
		valStore:                valStore,
		voteExtensionDecomposer: voteExtensionDecomposer,
		keeper:                  keeper,
	}
}

// PrepareProposalHandler returns the handler for the PrepareProposal ABCI request.
// This function:
// - Validates the signatures of the commit's vote extensions
// - Extracts the bridge-specific parts from the vote extensions that hold
//   the AssetsLocked events
// - Determines the sequence of canonical AssetsLocked events supported
//   by 2/3+ of the bridge validators and confirmed by 2/3+ of the non-bridge
//   validators
// - Injects the pseudo-transaction containing the canonical events at the
//   beginning of the original transaction list being part of the proposal.
//
// Dev note: It is fine to return a nil response and an error from this
// function in case of failure. The upstream app-level vote extension handler
// will handle the error gracefully and won't include the bridge-specific part
// in the app-level proposal.
func (ph *ProposalHandler) PrepareProposalHandler() sdk.PrepareProposalHandler {
	return func(
		ctx sdk.Context,
		req *cmtabci.RequestPrepareProposal,
	) (*cmtabci.ResponsePrepareProposal, error) {
		// According to the app-level proposal handler requirements, this
		// handler must validate signatures of the commit's vote extensions
		// on their own.
		err := baseapp.ValidateVoteExtensions(
			ctx,
			ph.valStore,
			req.Height,
			ctx.ChainID(),
			req.LocalLastCommit,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to validate vote extensions: %w", err)
		}

		bridgeValsConsAddrs := ph.valStore.GetValidatorsConsAddrsByPrivilege(
			bridgetypes.ValidatorPrivilege,
		)

		voteCounter := newAssetsLockedVoteCounter()

		// req.LocalLastCommit.Votes is actually a list of validators in the
		// last CometBFT validator set, with voting information regarding
		// the last block included. That means validators who did not vote
		// for the last block are not included in this list. Such votes
		// have an appropriate value of the BlockIdFlag set and their
		// vote extension is empty. This loop must take this into account.
		for _, vote := range req.LocalLastCommit.Votes {
			valConsAddr := sdk.ConsAddress(vote.Validator.Address)
			valVP := vote.Validator.Power

			isBridgeVal := slices.ContainsFunc(
				bridgeValsConsAddrs,
				func(address sdk.ConsAddress) bool {
					return address.Equals(valConsAddr)
				},
			)

			voteCounter.registerVoter(valVP, isBridgeVal)

			// The vote extension validators vote on and which is delivered
			// here is an app-level composite vote extension that contains
			// multiple vote extension parts. This bridge handler must decompose
			// it using the provided VoteExtensionDecomposer and process
			// just the vote extension part that is relevant to the bridge.
			compositeVoteExtension := vote.VoteExtension

			if len(compositeVoteExtension) == 0 {
				// TODO: Log this event.
				continue
			}

			voteExtensionBytes, err := ph.voteExtensionDecomposer(
				compositeVoteExtension,
			)
			if err != nil {
				// TODO: Log this event.
				continue
			}

			if len(voteExtensionBytes) == 0 {
				// TODO: Log this event.
				continue
			}

			var voteExtension types.VoteExtension
			if err := voteExtension.Unmarshal(voteExtensionBytes); err != nil {
				// TODO: Log this event.
				continue
			}

			if len(voteExtension.AssetsLockedEvents) == 0 {
				// TODO: Log this event.
				continue
			}

			// ABCI++ specification requires that PrepareProposal validates
			// vote extensions in the same manner as done in VerifyVoteExtension.
			// This is because extensions of votes included in the commit info
			// after the minimum of +2/3 had been reached are not verified.
			// See: https://docs.cometbft.com/v0.38/spec/abci/abci++_methods#prepareproposal
			if err := validateAssetsLockedEvents(voteExtension.AssetsLockedEvents); err != nil {
				// TODO: Log this event.
				continue
			}

			// addVote assumes the given event is valid and does not perform
			// any validation over it. We ensure validity by calling
			// validateAssetsLockedEvents above.
			for _, event := range voteExtension.AssetsLockedEvents {
				voteCounter.addVote(&event, valVP, isBridgeVal)
			}
		}

		// canonicalEvents returns the sequence of canonical AssetsLocked events
		// and guarantees that the sequence is strictly increasing by 1.
		canonicalEvents, err := voteCounter.canonicalEvents()
		if err != nil {
			return nil, fmt.Errorf(
				"failed to determine canonical AssetsLocked events: %w",
				err,
			)
		}

		// If there are no canonical events, we do not inject the pseudo-transaction
		// and return the proposal txs vector as is.
		if len(canonicalEvents) == 0 {
			return &cmtabci.ResponsePrepareProposal{Txs: req.Txs}, nil
		}

		sequenceTip := ph.keeper.GetAssetsLockedSequenceTip(ctx)
		// If sequence of canonical events does not start directly after the
		// current sequence tip, that means some earlier AssetsLocked events
		// are missing. In this case, we do not inject the pseudo-transaction
		// and return the proposal txs vector as is.
		if !canonicalEvents[0].Sequence.Equal(sequenceTip.AddRaw(1)) {
			return &cmtabci.ResponsePrepareProposal{Txs: req.Txs}, nil
		}

		extendedCommitInfo, err := req.LocalLastCommit.Marshal()
		if err != nil {
			return nil, fmt.Errorf("failed to marshal the extended commit info: %w", err)
		}

		// Construct the bridge-specific injected pseudo-transaction.
		injectedTx := types.InjectedTx{
			AssetsLockedEvents: canonicalEvents,
			ExtendedCommitInfo: extendedCommitInfo,
		}
		// Marshal the injected pseudo-transaction into bytes.
		injectedTxBytes, err := injectedTx.Marshal()
		if err != nil {
			// If marshaling fails, we cannot recover, so return an error.
			return nil, fmt.Errorf("failed to marshal injected tx: %w", err)
		}
		// Inject the pseudo-transaction at the beginning of the original
		// transaction list being part of the proposal. No need to check
		// for req.MaxTxBytes as this is done by the composite app-level
		// handler upstream.
		txs := append([][]byte{injectedTxBytes}, req.Txs...)

		return &cmtabci.ResponsePrepareProposal{Txs: txs}, nil
	}
}

func (ph *ProposalHandler) ProcessProposalHandler() sdk.ProcessProposalHandler {
	return func(
		ctx sdk.Context,
		req *cmtabci.RequestProcessProposal,
	) (*cmtabci.ResponseProcessProposal, error) {
		// TODO: Implement the ProcessProposalHandler.
		return nil, nil
	}
}

// assetsLockedVoteCounter is a helper structure that counts votes for
// AssetsLocked events. It is responsible for determining the sequence of
// canonical AssetsLocked events supported by 2/3+ of the bridge validators
// and confirmed by 2/3+ of the non-bridge validators.
type assetsLockedVoteCounter struct {
	// A map of AssetsLocked vote infos indexed by their sequence number.
	// In case validators have a different view of the world, the given
	// sequence number may have more than one item associated with it.
	voteInfoMap map[string][]*assetsLockedVoteInfo
	// The total voting power of validators with the bridge privilege
	// (totalVP - nonBridgeValsTotalVP).
	bridgeValsTotalVP int64
	// The total voting power of validators without the bridge privilege
	// (totalVP - bridgeValsTotalVP)
	nonBridgeValsTotalVP int64
}

// newAssetsLockedVoteCounter creates a new assetsLockedVoteCounter instance.
func newAssetsLockedVoteCounter() *assetsLockedVoteCounter {
	return &assetsLockedVoteCounter{
		voteInfoMap:          make(map[string][]*assetsLockedVoteInfo),
		bridgeValsTotalVP:    0,
		nonBridgeValsTotalVP: 0,
	}
}

// registerVoter registers a validator's voting power. It must be called for
// each validator able to vote, regardless of whether they voted or not.
func (alvc *assetsLockedVoteCounter) registerVoter(vp int64, isBridgeVal bool) {
	if isBridgeVal {
		alvc.bridgeValsTotalVP += vp
	} else {
		alvc.nonBridgeValsTotalVP += vp
	}
}

// addVote adds the given voting power to the sum of the bridge or non-bridge
// voting power behind the given AssetsLocked event. This function assumes
// the given event is valid and does not perform any validation over it.
// The caller is responsible for ensuring the event is actually valid.
func (alvc *assetsLockedVoteCounter) addVote(
	event *bridgetypes.AssetsLockedEvent,
	vp int64,
	isBridgeVal bool,
) {
	sequenceKey := event.Sequence.String()

	if index := slices.IndexFunc(
		alvc.voteInfoMap[sequenceKey],
		func(v *assetsLockedVoteInfo) bool {
			// It's enough to compare the recipient and the amount.
			// The sequence number is used as the votesInfo map key hence
			// it is the same for all items in the given value slice.
			return event.Recipient == v.event.Recipient &&
				event.Amount.Equal(v.event.Amount)
		},
	); index >= 0 {
		alvc.voteInfoMap[sequenceKey][index].add(vp, isBridgeVal)
	} else {
		voteInfo := newAssetsLockedVoteInfo(event)
		voteInfo.add(vp, isBridgeVal)
		alvc.voteInfoMap[sequenceKey] = append(
			alvc.voteInfoMap[sequenceKey],
			voteInfo,
		)
	}
}

// canonicalEvents returns the sequence of canonical AssetsLocked events
// supported by 2/3+ of the bridge validators and confirmed by 2/3+ of the
// non-bridge validators. If the sequence of canonical events cannot be
// determined, the function returns an error. Otherwise, it returns the
// canonical events in a sequence strictly increasing by 1.
func (alvc *assetsLockedVoteCounter) canonicalEvents() (
	[]bridgetypes.AssetsLockedEvent,
	error,
) {
	// Should never happen but just in case.
	if alvc.bridgeValsTotalVP <= 0 {
		return nil, fmt.Errorf("total bridge validators voting power must be positive")
	}
	if alvc.nonBridgeValsTotalVP <= 0 {
		return nil, fmt.Errorf("total non-bridge validators voting power must be positive")
	}

	requiredBridgeValsVP := ((alvc.bridgeValsTotalVP * 2) / 3) + 1
	requiredNonBridgeValsVP := ((alvc.nonBridgeValsTotalVP * 2) / 3) + 1

	canonicalEvents := make([]bridgetypes.AssetsLockedEvent, 0)

	for _, voteInfos := range alvc.voteInfoMap {
		// Filter out events that do not have the required super-majority
		// support from both bridge and non-bridge validators. Only 0 or 1
		// event should pass this filter.
		superMajorityEvents := make([]bridgetypes.AssetsLockedEvent, 0)
		for _, voteInfo := range voteInfos {
			if voteInfo.bridgeValsVP >= requiredBridgeValsVP &&
				voteInfo.nonBridgeValsVP >= requiredNonBridgeValsVP {
				superMajorityEvents = append(
					superMajorityEvents,
					*voteInfo.event,
				)
			}
		}

		switch len(superMajorityEvents) {
		case 0:
			// None of the events for the given sequence number have
			// received the required super-majority support from both
			// bridge and non-bridge validators. Skip this sequence number.
			continue
		case 1:
			// Exactly one event for the given sequence number has received
			// the required super-majority support from both bridge and
			// non-bridge validators. Add it to the canonical events slice.
			canonicalEvents = append(canonicalEvents, superMajorityEvents[0])
		default:
			// Multiple events for the given sequence number have received
			// the super-majority support. This case is not possible in the
			// current implementation of the bridge module. If this happens,
			// a serious bug has occurred and the function should panic.
			panic("multiple canonical AssetsLocked events for the same sequence")
		}
	}

	// Sort the canonical events slice by sequence number so achieve an
	// increasing sequence order. All events processed here are assumed to be
	// valid hence we sort without nil-checking the sequence number (see addVote).
	slices.SortFunc(canonicalEvents, func(a, b bridgetypes.AssetsLockedEvent) int {
		return a.Sequence.BigInt().Cmp(b.Sequence.BigInt())
	})

	// Ensure the sequence of canonical events is strictly increasing by 1.
	// In this context, we make sure there are no gaps between the sequence
	// numbers of the events and that each event is unique by its sequence.
	// We use IsStrictlyIncreasingSequence and not the full IsValid function
	// because we assume all events in the voteInfoMap are valid (see addVote).
	if !bridgetypes.AssetsLockedEvents(canonicalEvents).IsStrictlyIncreasingSequence() {
		return nil, fmt.Errorf("canonical events sequence is not strictly increasing")
	}

	return canonicalEvents, nil
}

// assetsLockedVoteInfo is a helper structure that holds the bridge and
// non-bridge voting power behind a single AssetsLocked event.
type assetsLockedVoteInfo struct {
	event           *bridgetypes.AssetsLockedEvent
	bridgeValsVP    int64
	nonBridgeValsVP int64
}

// newAssetsLockedVote creates a new assetsLockedVote instance.
func newAssetsLockedVoteInfo(
	event *bridgetypes.AssetsLockedEvent,
) *assetsLockedVoteInfo {
	return &assetsLockedVoteInfo{
		event:           event,
		bridgeValsVP:    0,
		nonBridgeValsVP: 0,
	}
}

// add adds the given voting power to the sum of the bridge or non-bridge
// voting power behind this AssetsLocked event.
func (alvi *assetsLockedVoteInfo) add(vp int64, isBridgeVal bool) {
	if isBridgeVal {
		alvi.bridgeValsVP += vp
	} else {
		alvi.nonBridgeValsVP += vp
	}
}

