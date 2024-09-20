package abci

import (
	"fmt"
	"slices"

	"cosmossdk.io/log"

	cmtabci "github.com/cometbft/cometbft/abci/types"
	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/mezo-org/mezod/x/bridge/abci/types"
	bridgekeeper "github.com/mezo-org/mezod/x/bridge/keeper"
	bridgetypes "github.com/mezo-org/mezod/x/bridge/types"
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
//   - Validates the signatures of the commit's vote extensions
//   - Extracts the bridge-specific parts from the vote extensions that hold
//     the AssetsLocked events
//   - Determines the sequence of canonical AssetsLocked events supported
//     by 2/3+ of the bridge validators and confirmed by 2/3+ of the non-bridge
//     validators
//   - Injects the pseudo-transaction containing the canonical events at the
//     beginning of the original transaction list being part of the proposal.
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
		// TODO: Consider changing logging to debug level once this code matures.

		ph.logger.Info(
			"bridge is preparing proposal",
			"height", req.Height,
		)

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

		ph.logger.Info(
			"last commit's vote extensions validated successfully",
			"height", req.Height,
		)

		// determineCanonicalEvents determines the sequence of canonical
		// AssetsLocked events and guarantees that the sequence is strictly
		// increasing by 1.
		canonicalEvents, err := ph.determineCanonicalEvents(
			ctx,
			req.LocalLastCommit,
			req.Height,
		)
		if err != nil {
			return nil, fmt.Errorf(
				"failed to determine canonical AssetsLocked events: %w",
				err,
			)
		}

		// If there are no canonical events, we do not inject the pseudo-transaction
		// and return the proposal txs vector as is.
		if len(canonicalEvents) == 0 {
			ph.logger.Info(
				"canonical AssetsLocked events sequence is empty",
				"height", req.Height,
			)

			return &cmtabci.ResponsePrepareProposal{Txs: req.Txs}, nil
		}

		ph.logger.Info(
			"canonical AssetsLocked events sequence extracted",
			"height", req.Height,
			"events_count", len(canonicalEvents),
			"events_sequence_start", canonicalEvents[0].Sequence,
		)

		sequenceTip := ph.keeper.GetAssetsLockedSequenceTip(ctx)
		// If sequence of canonical events does not start directly after the
		// current sequence tip, that means some earlier AssetsLocked events
		// are missing. In this case, we do not inject the pseudo-transaction
		// and return the proposal txs vector as is.
		if !canonicalEvents[0].Sequence.Equal(sequenceTip.AddRaw(1)) {
			ph.logger.Info(
				"canonical AssetsLocked events sequence does not "+
					"stick to the current sequence tip",
				"height", req.Height,
				"events_count", len(canonicalEvents),
				"events_sequence_start", canonicalEvents[0].Sequence,
				"sequence_tip", sequenceTip,
			)

			return &cmtabci.ResponsePrepareProposal{Txs: req.Txs}, nil
		}

		ph.logger.Info(
			"canonical AssetsLocked events sequence sticks to the current sequence tip",
			"height", req.Height,
			"events_count", len(canonicalEvents),
			"events_sequence_start", canonicalEvents[0].Sequence,
			"sequence_tip", sequenceTip,
		)

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

		ph.logger.Info(
			"bridge prepared proposal",
			"height", req.Height,
			"injected_tx_byte_length", len(injectedTxBytes),
		)

		return &cmtabci.ResponsePrepareProposal{Txs: txs}, nil
	}
}

// determineCanonicalEvents determines the sequence of canonical AssetsLocked
// events supported by 2/3+ of the bridge validators and confirmed by 2/3+ of
// the non-bridge validators. If the sequence of canonical events cannot be
// determined, the function returns an error. Otherwise, it returns the canonical
// events in a sequence strictly increasing by 1.
func (ph *ProposalHandler) determineCanonicalEvents(
	ctx sdk.Context,
	extendedCommitInfo cmtabci.ExtendedCommitInfo,
	height int64,
) (bridgetypes.AssetsLockedEvents, error) {
	bridgeValsConsAddrs := ph.valStore.GetValidatorsConsAddrsByPrivilege(
		ctx,
		bridgetypes.ValidatorPrivilege,
	)

	voteCounter := newAssetsLockedVoteCounter()

	// extendedCommitInfo.Votes is actually a list of validators in the
	// last CometBFT validator set, with voting information regarding
	// the last block included. That means validators who did not vote
	// for the last block are not included in this list. Such votes
	// have an appropriate value of the BlockIdFlag set and their
	// vote extension is empty. This loop must take this into account.
	for _, vote := range extendedCommitInfo.Votes {
		valConsAddr := sdk.ConsAddress(vote.Validator.Address)
		valVP := vote.Validator.Power

		isBridgeVal := slices.ContainsFunc(
			bridgeValsConsAddrs,
			func(address sdk.ConsAddress) bool {
				return address.Equals(valConsAddr)
			},
		)

		voteCounter.registerVoter(valVP, isBridgeVal)

		logInvalidVoteExtension := func(cause string) {
			ph.logger.Debug(
				"invalid vote extension while determining "+
					"canonical AssetsLocked events",
				"height", height,
				"from", valConsAddr.String(),
				"cause", cause,
			)
		}

		// The vote extension validators vote on and which is delivered
		// here is an app-level composite vote extension that contains
		// multiple vote extension parts. This bridge handler must decompose
		// it using the provided VoteExtensionDecomposer and process
		// just the vote extension part that is relevant to the bridge.
		compositeVoteExtension := vote.VoteExtension

		if len(compositeVoteExtension) == 0 {
			logInvalidVoteExtension("empty composite vote extension")
			continue
		}

		voteExtensionBytes, err := ph.voteExtensionDecomposer(
			compositeVoteExtension,
		)
		if err != nil {
			logInvalidVoteExtension(fmt.Sprintf("decomposition error: %v", err))
			continue
		}

		if len(voteExtensionBytes) == 0 {
			logInvalidVoteExtension("empty bridge-specific vote extension")
			continue
		}

		var voteExtension types.VoteExtension
		if err := voteExtension.Unmarshal(voteExtensionBytes); err != nil {
			logInvalidVoteExtension(fmt.Sprintf("unmarshaling error: %v", err))
			continue
		}

		if len(voteExtension.AssetsLockedEvents) == 0 {
			logInvalidVoteExtension("no AssetsLocked events")
			continue
		}

		// ABCI++ specification requires that the proposal phase validates
		// vote extensions in the same manner as done in VerifyVoteExtension.
		// This is because extensions of votes included in the commit info
		// after the minimum of +2/3 had been reached are not verified.
		// See: https://docs.cometbft.com/v0.38/spec/abci/abci++_methods#prepareproposal
		if err := validateAssetsLockedEvents(voteExtension.AssetsLockedEvents); err != nil {
			logInvalidVoteExtension(fmt.Sprintf("invalid AssetsLocked sequence: %v", err))
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
	return voteCounter.canonicalEvents()
}

// ProcessProposalHandler returns the handler for the ProcessProposal ABCI request.
// This function validates the injected bridge-specific pseudo-tx to determine
// proposal acceptance or rejection. Specifically it:
//   - Makes sure the injected pseudo-tx exists and unmarshals correctly
//   - Verifies whether the commit info attached to the pseudo-tx unmarshals correctly
//   - Ensures injected pseudo-tx contains a non-empty slice of AssetsLocked events
//   - Validates the signatures of the vote extensions attached to the injected
//     pseudo-tx (as part of the commit info)
//   - Recreates the canonical sequence of AssetsLocked events using the attached
//     vote extensions and ensures it matches the AssetsLocked events from the
//     injected pseudo-tx
//   - Makes sure the AssetsLocked events from the injected pseudo-tx start directly
//     after the current sequence tip
//
// If the injected pseudo-tx is valid, the proposal is accepted. Empty pseudo-txs
// lead to proposal acceptance by default.
//
// Dev note: In case the injected pseudo-tx is invalid, we REJECT the proposal
// explicitly and return an error describing the reason. Due to the limitations
// of the Cosmos interface, REJECT without an error does not provide any details
// about the reason. Conversely, error without REJECT is confusing as it should
// rather denote a failure of the handler itself. The upstream app-level proposal
// handler will handle all non-ACCEPT cases gracefully and reject the app-level
// proposal.
//
// See Skip's price oracle ProcessProposal handler for a similar pattern:
// https://github.com/skip-mev/connect/blob/53b22d1ff50f5b60d50346640e32bbd77472e95e/abci/proposals/proposals.go#L255
func (ph *ProposalHandler) ProcessProposalHandler() sdk.ProcessProposalHandler {
	return func(
		ctx sdk.Context,
		req *cmtabci.RequestProcessProposal,
	) (*cmtabci.ResponseProcessProposal, error) {
		from := sdk.ConsAddress(req.ProposerAddress).String()

		ph.logger.Debug(
			"bridge is processing proposal",
			"height", req.Height,
			"from", from,
		)

		if len(req.Txs) == 0 {
			// The app-level handler always passes a transaction vector
			// with at least one element (representing the possibly empty
			// bridge-specific pseudo-transaction). If the vector is empty,
			// something went wrong upstream and we cannot recover so, return
			// an error.
			return &cmtabci.ResponseProcessProposal{
				Status: cmtabci.ResponseProcessProposal_REJECT,
			}, fmt.Errorf("empty transaction vector in the proposal")
		}

		if len(req.Txs[0]) == 0 {
			// As said above, the app-level handler always passes a transaction
			// vector with at least one element (representing the possibly empty
			// bridge-specific pseudo-transaction). The first element can be a
			// byte slice if the bridge-level PrepareProposal handler decided to
			// not inject the pseudo-transaction. In this case, we do not need to
			// do anything and return ACCEPT.
			ph.logger.Debug(
				"bridge accepted empty proposal",
				"height", req.Height,
				"from", from,
			)

			return &cmtabci.ResponseProcessProposal{
				Status: cmtabci.ResponseProcessProposal_ACCEPT,
			}, nil
		}

		var injectedTx types.InjectedTx
		if err := injectedTx.Unmarshal(req.Txs[0]); err != nil {
			// If unmarshaling of the injected pseudo-transaction fails,
			// we cannot recover, so return an error.
			return &cmtabci.ResponseProcessProposal{
				Status: cmtabci.ResponseProcessProposal_REJECT,
			}, fmt.Errorf("failed to unmarshal injected tx: %w", err)
		}

		// If the pseudo-transaction is present but holds an empty slice of
		// AssetsLocked events, we reject the proposal as the block proposer
		// misbehaved.
		if len(injectedTx.AssetsLockedEvents) == 0 {
			return &cmtabci.ResponseProcessProposal{
				Status: cmtabci.ResponseProcessProposal_REJECT,
			}, fmt.Errorf("injected tx does not contain AssetsLocked events")
		}

		var extendedCommitInfo cmtabci.ExtendedCommitInfo
		if err := extendedCommitInfo.Unmarshal(injectedTx.ExtendedCommitInfo); err != nil {
			// If unmarshaling of the attached commit info fails,
			// we cannot recover, so return an error.
			return &cmtabci.ResponseProcessProposal{
				Status: cmtabci.ResponseProcessProposal_REJECT,
			}, fmt.Errorf("failed to unmarshal commit info from injected tx: %w", err)
		}

		// According to the app-level proposal handler requirements, this
		// handler must re-validate signatures of the vote extensions
		// attached by the block proposer on their own.
		err := baseapp.ValidateVoteExtensions(
			ctx,
			ph.valStore,
			req.Height,
			ctx.ChainID(),
			extendedCommitInfo,
		)
		if err != nil {
			return &cmtabci.ResponseProcessProposal{
					Status: cmtabci.ResponseProcessProposal_REJECT,
				}, fmt.Errorf(
					"failed to validate vote extensions from injexted tx: %w",
					err,
				)
		}

		// Re-create the canonical events using vote extensions from the
		// injected pseudo-transaction. determineCanonicalEvents guarantees
		// that the sequence is valid and strictly increasing by 1.
		recreatedCanonicalEvents, err := ph.determineCanonicalEvents(
			ctx,
			extendedCommitInfo,
			req.Height,
		)
		if err != nil {
			return &cmtabci.ResponseProcessProposal{
				Status: cmtabci.ResponseProcessProposal_REJECT,
			}, fmt.Errorf("failed to recreate canonical AssetsLocked events: %w", err)
		}

		// Make sure the recreated canonical events match the injected ones.
		// This is a proof that the block proposer behaved correctly and used
		// signed vote extensions to inject the canonical AssetsLocked events.
		if !recreatedCanonicalEvents.Equal(injectedTx.AssetsLockedEvents) {
			return &cmtabci.ResponseProcessProposal{
					Status: cmtabci.ResponseProcessProposal_REJECT,
				}, fmt.Errorf(
					"recreated canonical AssetsLocked events do not match " +
						"events from injected tx",
				)
		}

		sequenceTip := ph.keeper.GetAssetsLockedSequenceTip(ctx)
		// If sequence of injected events does not start directly after the
		// current sequence tip, that means some earlier AssetsLocked events
		// are missing. That means the block proposer misbehaved and the proposal
		// should be rejected. Note that we could have done this check earlier
		// but, at this point, we know that the injected events match the recreated
		// canonical events that are guaranteed to be valid. This way, we can safely
		// compare the sequence of the first injected event with the sequence tip.
		if !injectedTx.AssetsLockedEvents[0].Sequence.Equal(sequenceTip.AddRaw(1)) {
			return &cmtabci.ResponseProcessProposal{
					Status: cmtabci.ResponseProcessProposal_REJECT,
				}, fmt.Errorf(
					"AssetsLocked events from injected tx do not start " +
						"after the current sequence tip",
				)
		}

		ph.logger.Debug(
			"bridge accepted proposal",
			"height", req.Height,
			"from", from,
		)

		return &cmtabci.ResponseProcessProposal{
			Status: cmtabci.ResponseProcessProposal_ACCEPT,
		}, nil
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
	if alvc.bridgeValsTotalVP <= 0 || alvc.nonBridgeValsTotalVP <= 0 {
		// This case means either:
		// - None of the bridge/non-bridge validators has voted for the block OR
		// - The bridge/non-bridge validators do not exist in the current validator set.
		// In both cases, bridging is not possible and, we return an empty
		// slice of canonical events.
		return []bridgetypes.AssetsLockedEvent{}, nil
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

