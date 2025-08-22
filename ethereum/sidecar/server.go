package sidecar

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"net"
	"sort"
	"sync"
	"time"

	"cosmossdk.io/log"
	sdkmath "cosmossdk.io/math"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	ethconfig "github.com/keep-network/keep-common/pkg/chain/ethereum"
	ethconnect "github.com/mezo-org/mezod/ethereum"
	"github.com/mezo-org/mezod/ethereum/bindings/portal"
	"github.com/mezo-org/mezod/ethereum/sidecar/mezotime"
	pb "github.com/mezo-org/mezod/ethereum/sidecar/types"
	"github.com/mezo-org/mezod/version"
	bridgetypes "github.com/mezo-org/mezod/x/bridge/types"
	"google.golang.org/grpc"
)

var (
	// mezoBridgeName is the name of the MezoBridge contract.
	mezoBridgeName = "MezoBridge"

	// searchedRange is a number of blocks to look back for AssetsLocked events in
	// the MezoBridge contract. This is approximately 30 days window for 12 sec.
	// block time.
	searchedRange = new(big.Int).SetInt64(220000)

	// cachedEventsLimit is a number of events to keep in the cache.
	// Size of Sequence: 32 bytes
	// Size of Recipient: 42 bytes
	// Size of Amount: 32 bytes
	// Struct overhead and padding: ~16bytes
	// Total size of one event: 122 bytes (0.12KB)
	// Assuming we want to allocate ~64MB for the cache, we can store ~546k events.
	// For simplicity, let's make 500k events our limit. Even if deposits hit 1000
	// daily mark, that would give 50 days of cached events which should be more than
	// enough.
	cachedEventsLimit = 500000

	// assetsUnlockedLookBackPeriod is the look-back period used when fetching
	// AssetsUnlocked events from the Mezo chain. It defines how far back we
	// look when searching for unconfirmed events.
	assetsUnlockedLookBackPeriod = 30 * 24 * time.Hour // ~1 month

	// assetsUnlockConfirmedLookBackBlocks is the number of blocks used when
	// fetching AssetsUnlockConfirmed events from the Ethereum chain. It defines
	// how far back we look when searching for the events. Its value should
	// cover at least one month of Ethereum blocks; we add a 5-day margin.
	assetsUnlockConfirmedLookBackBlocks = uint64(252000) // ~35 days

	// assetsUnlockedBatchSize is the number of AssetsUnlocked events we can
	// fetch from the Mezo blockchain in one gRPC call.
	assetsUnlockedBatchSize = 100

	// assetsUnlockedFetchingPeriod is the time period defining how often new
	// AssetsUnlocked events should be fetched.
	assetsUnlockedFetchingPeriod = 1 * time.Minute

	// errSequenceGap is the error reported when there is a gap between
	// sequences of events.
	errSequenceGap = fmt.Errorf("sequence gap between events")

	// errInvalidEvents is the error reported when events are invalid.
	errInvalidEvents = fmt.Errorf("invalid AssetsLocked events")
)

// AssetsUnlockedEndpoint is a client enabling communication with the `Mezo`
// chain.
type AssetsUnlockedEndpoint interface {
	// GetAssetsUnlockedSequenceTip gets the assets unlocked sequence tip from
	// the Mezo chain. The returned sequence tip is equal to the number of
	// AssetsUnlocked events made so far. It is also equal to the value of the
	// unlock sequence in the newest AssetsUnlocked event.
	GetAssetsUnlockedSequenceTip(ctx context.Context) (sdkmath.Int, error)

	// GetAssetsUnlockedEvents gets the AssetsUnlocked events from the Mezo
	// chain. The requested range of events is inclusive on the lower side and
	// exclusive on the upper side.
	GetAssetsUnlockedEvents(
		ctx context.Context,
		sequenceStart sdkmath.Int,
		sequenceEnd sdkmath.Int,
	) ([]bridgetypes.AssetsUnlockedEvent, error)
}

// Server coordinates bridge data between Ethereum and Mezo.
// It handles both directions:
//   - Bridge-in (Ethereum to Mezo): watches AssetsLocked events emitted by the
//     MezoBridge contract and serves them via gRPC to `mezod` validator nodes.
//   - Bridge-out (Mezo to Ethereum): monitors AssetsUnlocked events on Mezo and
//     attests on Ethereum. Events are read from the chain via validators gRPC
//     endpoints.
//
// Server is intended to run as a separate process.
type Server struct {
	logger log.Logger

	// bridging-in
	grpcServer *grpc.Server

	assetsLockedEventsMutex sync.RWMutex
	assetsLockedEvents      []bridgetypes.AssetsLockedEvent

	lastFinalizedBlockMutex sync.RWMutex
	lastFinalizedBlock      *big.Int

	bridgeContract ethconnect.BridgeContract
	chain          ethconnect.Chain

	batchSize         uint64
	requestsPerMinute uint64

	// Channel used to indicate whether the data-heavy part of AssetsLocked
	// observation routine is done. Once this channel is closed we can proceed
	// with the AssetsUnlocked observation routine. Delaying launch helps avoid
	// overwhelming Ethereum data providers with requests.
	assetsLockedReady chan struct{}

	// bridging-out
	assetsUnlockedEndpoint AssetsUnlockedEndpoint

	assetsUnlockedLookBackPeriod time.Duration
	assetsUnlockedBatchSize      int

	attestationMutex sync.RWMutex
	attestationQueue []bridgetypes.AssetsUnlockedEvent

	// Unguarded by mutex as only the AssetsUnlock event observation
	// routine uses it.
	lastAssetsUnlockedSequence sdkmath.Int

	// privateKey is an optional ECDSA private key extracted from keyring
	privateKey *ecdsa.PrivateKey

	attestationValidator *attestationValidator
	blockHeightWaiter    ethconfig.BlockHeightWaiter
	submissionQueue      *submissionQueue
}

// RunServer initializes the server, starts the event observing routine and
// starts the gRPC server.
func RunServer(
	logger log.Logger,
	grpcAddress string,
	providerURL string,
	ethereumNetwork string,
	batchSize uint64,
	requestsPerMinute uint64,
	assetsUnlockedEndpoint string,
	registry codectypes.InterfaceRegistry,
	privateKey *ecdsa.PrivateKey,
) {
	network := ethconnect.NetworkFromString(ethereumNetwork)
	mezoBridgeAddress := portal.MezoBridgeAddress(network)

	if mezoBridgeAddress == "" {
		panic(
			"cannot get address of the MezoBridge contract on Ethereum; " +
				"make sure you run 'make bindings' before building the binary",
		)
	}

	logger.Info(
		"Sidecar server resolved MezoBridge contract and Ethereum network",
		"mezo_bridge_address", mezoBridgeAddress,
		"ethereum_network", network,
	)

	ctx, cancelCtx := context.WithCancel(context.Background())
	defer cancelCtx()

	var err error
	// Connect to the Ethereum network
	chain, err := ethconnect.Connect(
		ctx,
		ethconfig.Config{
			Network:           network,
			URL:               providerURL,
			ContractAddresses: map[string]string{mezoBridgeName: mezoBridgeAddress},
		},
		privateKey,
	)
	if err != nil {
		panic(fmt.Sprintf("failed to connect to the Ethereum network: %v", err))
	}

	// Initialize the MezoBridge contract instance.
	bridgeContractBinding, err := initializeBridgeContract(common.HexToAddress(mezoBridgeAddress), chain)
	if err != nil {
		panic(fmt.Sprintf("failed to initialize MezoBridge contract: %v", err))
	}

	bridgeContract := NewBridgeContract(bridgeContractBinding)

	attestationValidator := newAttestationValidation(
		logger,
		bridgeContract,
		chain.Key().Address,
	)

	submissionQueue := newSubmissionQueue(
		logger,
		bridgeContract,
		chain.Key().Address,
	)

	assetsUnlockedGrpcEndpoint, err := NewAssetsUnlockedGrpcEndpoint(
		assetsUnlockedEndpoint,
		registry,
	)
	if err != nil {
		panic(fmt.Sprintf("failed to create assets unlocked endpoint: %v", err))
	}

	server := &Server{
		logger:                       logger,
		grpcServer:                   grpc.NewServer(),
		assetsLockedEvents:           make([]bridgetypes.AssetsLockedEvent, 0),
		lastFinalizedBlock:           new(big.Int),
		bridgeContract:               bridgeContract,
		chain:                        chain,
		batchSize:                    batchSize,
		requestsPerMinute:            requestsPerMinute,
		assetsLockedReady:            make(chan struct{}),
		assetsUnlockedEndpoint:       assetsUnlockedGrpcEndpoint,
		assetsUnlockedLookBackPeriod: assetsUnlockedLookBackPeriod,
		assetsUnlockedBatchSize:      assetsUnlockedBatchSize,
		attestationQueue:             []bridgetypes.AssetsUnlockedEvent{},
		privateKey:                   privateKey,
		attestationValidator:         attestationValidator,
		blockHeightWaiter:            chain.BlockCounter(),
		submissionQueue:              submissionQueue,
	}

	go func() {
		defer cancelCtx()
		err := server.observeAssetsLockedEvents(ctx)
		if err != nil {
			server.logger.Error(
				"AssetsLocked events observation routine failed",
				"err", err,
			)
		}

		server.logger.Info("AssetsLocked events observation routine stopped")
	}()

	go func() {
		defer cancelCtx()
		err := server.startGRPCServer(ctx, grpcAddress)
		if err != nil {
			server.logger.Error("gRPC server routine failed", "err", err)
		}

		server.logger.Info("gRPC server routine stopped")
	}()

	server.logger.Info(
		"Waiting for the initial AssetsLocked sync before launching " +
			"AssetsUnlocked routine",
	)

	// Wait until the initial synchronization of the AssetsLocked routine is
	// complete.
	select {
	case <-server.assetsLockedReady:
		server.logger.Info(
			"Initial AssetsLocked sync completed; proceeding with " +
				"AssetsUnlocked routines",
		)
	case <-ctx.Done():
		server.logger.Info(
			"Context canceled while waiting; exiting without launching " +
				"AssetsUnlocked routines",
		)
		return
	}

	go func() {
		// TODO: Only a subset of validators should be attesting AssetsUnlocked
		//       events. Decide whether we should run this goroutine by calling
		//       MezoBridge.bridgeValidators() with our Ethereum address.

		defer cancelCtx()
		err := server.observeAssetsUnlockedEvents(ctx)
		if err != nil {
			server.logger.Error(
				"AssetsUnlocked events observation routine failed",
				"err", err,
			)
		}

		server.logger.Info("AssetsUnlocked events observation routine stopped")
	}()

	go func() {
		// TODO: Only a subset of validators should be attesting AssetsUnlocked
		//       events. Decide whether we should run this goroutine by calling
		//       MezoBridge.bridgeValidators() with our Ethereum address.

		defer cancelCtx()
		server.attestAssetsUnlockedEvents(ctx)
		server.logger.Info("AssetsUnlocked events attestation routine stopped")
	}()

	<-ctx.Done()

	server.logger.Error("sidecar stopped")
}

// observeAssetsLockedEvents monitors and processes AssetsLocked events from the MezoBridge smart contract.
//
//   - Initializes a MezoBridge contract instance using the provided blockchain connection and contract address.
//   - Retrieves the most recent finalized block number from the blockchain.
//   - Calculates a start block that is two weeks prior to the current finalized block. It then fetches
//     `AssetsLocked` events for this range.
//   - Sets up a ticker channel to continuously monitor new finalized blocks.
//   - On each new block notification from the ticker channel, calls `processEvents` to handle new events
//     since the last finalized block.
func (s *Server) observeAssetsLockedEvents(ctx context.Context) error {
	finalizedBlock, err := s.chain.FinalizedBlock(ctx)
	if err != nil {
		return fmt.Errorf("failed to get the finalized block: [%w]", err)
	}

	var startBlock uint64
	if searchedRange.Cmp(finalizedBlock) <= 0 {
		// Fetch AssetsLocked events two weeks back from the finalized block
		startBlock = new(big.Int).Sub(finalizedBlock, searchedRange).Uint64()
	} else {
		startBlock = 0
	}
	err = s.fetchFinalizedAssetsLockedEvents(startBlock, finalizedBlock.Uint64())
	if err != nil {
		return fmt.Errorf("failed to fetch historical events: [%w]", err)
	}

	s.lastFinalizedBlockMutex.Lock()
	s.lastFinalizedBlock = finalizedBlock
	s.lastFinalizedBlockMutex.Unlock()

	// Signal that initial synchronization is ready.
	close(s.assetsLockedReady)

	// Start a ticker to periodically check the current block number
	tickerChan := s.chain.WatchBlocks(ctx)

	for {
		select {
		case <-ctx.Done(): // Handle context cancellation
			s.logger.Info(
				"stopping assets locked event observation due to context cancellation",
			)
			return nil
		case <-tickerChan:
			// On each tick check if the current finalized block is greater than the last
			// finalized block.
			// TODO: Add a basic counter to manage validation issues that may occur
			//	     when processing events. This counter should allow a few retry
			//		 attempts to handle temporary connection issues, but should halt
			//		 the sidecar if an error arises specifically with event processing.
			err := s.processAssetsLockedEvents(ctx)
			if err != nil {
				s.logger.Error(
					"failed to monitor newly emitted assets locked events",
					"err", err,
				)
			}
		}
	}
}

// processAssetsLockedEvents processes AssetsLocked events from Ethereum by
// fetching the AssetsLocked finalized events within a specified block range and
// managing memory usage for cached events.
func (s *Server) processAssetsLockedEvents(ctx context.Context) error {
	currentFinalizedBlock, err := s.chain.FinalizedBlock(ctx)
	if err != nil {
		return fmt.Errorf("cannot get finalized block: [%w]", err)
	}

	s.lastFinalizedBlockMutex.RLock()
	shouldFetchEvents := currentFinalizedBlock.Cmp(s.lastFinalizedBlock) > 0
	s.lastFinalizedBlockMutex.RUnlock()

	if shouldFetchEvents {
		// Specified range in FilterOps is inclusive.
		// 1 is added to the lastFinalizedBlock to make the range exclusive at
		// the beginning of the range.
		// E.g.
		// lastFinalizedBlock = 100
		// currentFinalizedBlock = 132
		// fetching events in the following range [101, 132]
		exclusiveLastFinalizedBlock := s.lastFinalizedBlock.Uint64() + 1
		err := s.fetchFinalizedAssetsLockedEvents(
			exclusiveLastFinalizedBlock,
			currentFinalizedBlock.Uint64(),
		)
		if err != nil {
			return fmt.Errorf(
				"cannot fetch finalized AssetsLocked events: [%w]",
				err,
			)
		}
		s.lastFinalizedBlockMutex.Lock()
		s.lastFinalizedBlock = currentFinalizedBlock
		s.lastFinalizedBlockMutex.Unlock()
	}

	// Free up memory up to the length that exceeds the cache size.
	if len(s.assetsLockedEvents) > cachedEventsLimit {
		s.assetsLockedEventsMutex.Lock()
		trim := len(s.assetsLockedEvents) - cachedEventsLimit
		s.assetsLockedEvents = s.assetsLockedEvents[trim:]
		s.assetsLockedEventsMutex.Unlock()
	}

	return nil
}

// fetchFinalizedAssetsLockedEvents retrieves and processes finalized
// `AssetsLocked` events from the Ethereum network, within a specified block
// range. It uses the provided MezoBridge contract to filter these events.
// Each event is transformed into an `AssetsLockedEvent` type compatible with
// the bridgetypes package and added to the server's event list with mutex
// protection.
func (s *Server) fetchFinalizedAssetsLockedEvents(startBlock uint64, endBlock uint64) error {
	abiEvents, err := s.fetchAssetsLockedABIEvents(startBlock, endBlock)
	if err != nil {
		return fmt.Errorf("failed to fetch AssetsLocked ABI events: [%w]", err)
	}

	//nolint:all
	var bufferedEvents []bridgetypes.AssetsLockedEvent
	for _, abiEvent := range abiEvents {
		event := bridgetypes.AssetsLockedEvent{
			Sequence:  sdkmath.NewIntFromBigInt(abiEvent.SequenceNumber),
			Recipient: sdk.AccAddress(abiEvent.Recipient.Bytes()).String(),
			Token:     abiEvent.Token.Hex(),
			Amount:    sdkmath.NewIntFromBigInt(abiEvent.Amount),
		}
		bufferedEvents = append(bufferedEvents, event)
		s.logger.Info(
			"finalized AssetsLocked event",
			"sequence", event.Sequence.String(),
			"recipient", abiEvent.Recipient,
			"token", abiEvent.Token,
			"amount", event.Amount.String(),
		)
	}

	if len(bufferedEvents) == 0 {
		s.logger.Info("no new events to process")
		return nil
	}

	// Make sure the events are sorted in ascending order by the sequence number
	sort.Slice(bufferedEvents, func(i, j int) bool {
		return bufferedEvents[i].Sequence.LT(bufferedEvents[j].Sequence)
	})

	s.assetsLockedEventsMutex.Lock()
	defer s.assetsLockedEventsMutex.Unlock()

	// Make sure there are no gaps between events and bufferedEvents lists
	if len(s.assetsLockedEvents) > 0 && len(bufferedEvents) > 0 {
		lastEvent := s.assetsLockedEvents[len(s.assetsLockedEvents)-1]
		firstEvent := bufferedEvents[0]
		if !lastEvent.Sequence.Add(sdkmath.NewInt(1)).Equal(firstEvent.Sequence) {
			return errSequenceGap
		}
	}

	if !bridgetypes.AssetsLockedEvents(bufferedEvents).IsValid() {
		return errInvalidEvents
	}

	s.assetsLockedEvents = append(s.assetsLockedEvents, bufferedEvents...)

	return nil
}

// fetchAssetsLockedABIEvents retrieves raw `AssetsLocked` ABI events from the
// MezoBridge contract within a specified block range. The function fetches
// events in batches if the entire range is too large to fetch at once.
func (s *Server) fetchAssetsLockedABIEvents(
	startBlock uint64,
	endBlock uint64,
) ([]*portal.MezoBridgeAssetsLocked, error) {
	s.logger.Info(
		"fetching AssetsLocked events from range",
		"startBlock", startBlock,
		"endBlock", endBlock,
	)

	result := make([]*portal.MezoBridgeAssetsLocked, 0)

	ticker := time.NewTicker(time.Minute / time.Duration(s.requestsPerMinute)) //nolint:gosec
	defer ticker.Stop()

	events, err := s.bridgeContract.PastAssetsLockedEvents(
		startBlock,
		&endBlock,
		nil,
		nil,
		nil,
	)
	if err != nil {
		s.logger.Warn(
			"failed to fetch AssetsLocked events from the entire range; "+
				"falling back to batched events fetch",
			"startBlock", startBlock,
			"endBlock", endBlock,
			"err", err,
		)

		batchStartBlock := startBlock

		for batchStartBlock <= endBlock {
			batchEndBlock := batchStartBlock + s.batchSize
			if batchEndBlock > endBlock {
				batchEndBlock = endBlock
			}

			s.logger.Info(
				"fetching a batch of AssetsLocked events from range",
				"batchStartBlock", batchStartBlock,
				"batchEndBlock", batchEndBlock,
			)

			<-ticker.C

			batchEvents, batchErr := s.bridgeContract.PastAssetsLockedEvents(
				batchStartBlock,
				&batchEndBlock,
				nil,
				nil,
				nil,
			)
			if batchErr != nil {
				return nil, fmt.Errorf(
					"batched AssetsLocked fetch failed: [%w]; giving up",
					batchErr,
				)
			}

			result = append(result, batchEvents...)

			batchStartBlock = batchEndBlock + 1
		}
	} else {
		result = append(result, events...)
	}

	return result, nil
}

// startGRPCServer starts the gRPC server and registers the Ethereum sidecar
// service.
func (s *Server) startGRPCServer(
	ctx context.Context,
	address string,
) error {
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return fmt.Errorf("failed to listen: [%w]", err)
	}

	pb.RegisterEthereumSidecarServer(s.grpcServer, s)

	s.logger.Info(
		"gRPC server started",
		"address", address,
	)

	defer s.grpcServer.GracefulStop()

	errChan := make(chan error)

	go func() {
		err := s.grpcServer.Serve(listener)
		if err != nil {
			errChan <- err
		}
	}()

	select {
	case err := <-errChan:
		return fmt.Errorf("serve failed: [%w]", err)
	case <-ctx.Done():
		return nil
	}
}

// observeAssetsUnlockedEvents monitors `AssetsUnlocked` events emitted on the
// Mezo chain and stores unconfirmed `AssetsUnlocked` events preparing them for
// attestation. This routine consists of two parts:
//   - initial check for unconfirmed `AssetsUnlocked` events emitted when the
//     sidecar was turned off
//   - periodic check for new unconfirmed `AssetsUnlocked` events
func (s *Server) observeAssetsUnlockedEvents(ctx context.Context) error {
	// At the start of the routine we need to learn which of the recent
	// `AssetsUnlocked` events might require attestation in the `MezoBridge`
	// smart contract. It's possible that the sidecar has been turned off
	// for a significant amount of time while the Mezo chain was processing
	// bridge-out requests.

	// First fetch recent events. We consider such events as possibly
	// unconfirmed. The events are already sorted by their unlock sequence in
	// ascending order.
	recentEvents, err := s.fetchRecentAssetsUnlockedEvents(ctx)
	if err != nil {
		return fmt.Errorf(
			"failed to fetch recent AssetsUnlocked events: [%w]",
			err,
		)
	}

	unconfirmedEvents, err := s.findUnconfirmedAssetsUnlockedEvents(
		ctx,
		recentEvents,
	)
	if err != nil {
		return fmt.Errorf(
			"failed to find unconfirmed AssetsUnlocked events: [%w]",
			err,
		)
	}

	s.attestationMutex.Lock()
	s.attestationQueue = unconfirmedEvents
	s.attestationMutex.Unlock()

	// Save the unlock sequence of the last event as the starting point for
	// further event fetching.
	s.lastAssetsUnlockedSequence = sdkmath.ZeroInt()
	if len(recentEvents) > 0 {
		s.lastAssetsUnlockedSequence = recentEvents[len(recentEvents)-1].UnlockSequence
	}

	s.logger.Info(
		"Initial search for recent unconfirmed AssetsUnlocked events",
		"recent_events", len(recentEvents),
		"unconfirmed_events", len(unconfirmedEvents),
		"unlock_sequence_tip", s.lastAssetsUnlockedSequence.String(),
	)

	ticker := time.NewTicker(assetsUnlockedFetchingPeriod)
	defer ticker.Stop()
	tickerChan := ticker.C

	for {
		select {
		case <-ctx.Done(): // Handle context cancellation
			s.logger.Info(
				"stopping assets unlocked event observation due to context cancellation",
			)
			return nil
		case <-tickerChan:
			err := s.fetchNewAssetsUnlockedEvents(ctx)
			if err != nil {
				s.logger.Error(
					"failed to monitor newly emitted assets unlocked events",
					"err", err,
				)
			}
		}
	}
}

// fetchRecentAssetsUnlockedEvents fetches AssetsUnlocked events that entered
// the Mezo blockchain within the AssetsUnlocked look-back period. We consider
// events within that range as possibly unconfirmed. The returned events are
// sorted in the ascending order of their unlock sequences (from the lowest
// to the highest).
func (s *Server) fetchRecentAssetsUnlockedEvents(ctx context.Context) (
	[]bridgetypes.AssetsUnlockedEvent,
	error,
) {
	// Fetching events starts from the current sequence tip.
	sequenceTip, err := s.assetsUnlockedEndpoint.GetAssetsUnlockedSequenceTip(ctx)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to fetch assets unlocked sequence tip: [%w]",
			err,
		)
	}

	// The sequence tip is zero, meaning no events have been made.
	if sequenceTip.IsZero() {
		return []bridgetypes.AssetsUnlockedEvent{}, nil
	}

	// Cut-off block time in UNIX seconds.
	cutOffBlockTime := uint32(mezotime.Now().Add(-s.assetsUnlockedLookBackPeriod).Unix()) //nolint:gosec
	recentEvents := []bridgetypes.AssetsUnlockedEvent{}

	// Walk backwards from the current sequence tip in windows of at most
	// `assetsUnlockedBatchSize`.
outer:
	for sequenceTip.IsPositive() {
		// Sequence end is exclusive. We must add `1`.
		seqEnd := sequenceTip.AddRaw(1)

		// Sequence start is inclusive. It cannot be lower than `1`.
		seqStart := seqEnd.SubRaw(int64(s.assetsUnlockedBatchSize))
		if seqStart.LT(sdkmath.OneInt()) {
			seqStart = sdkmath.OneInt()
		}

		events, err := s.assetsUnlockedEndpoint.GetAssetsUnlockedEvents(
			ctx,
			seqStart,
			seqEnd,
		)
		if err != nil {
			return nil, fmt.Errorf(
				"failed to fetch AssetsUnlocked events [%s, %s): %w",
				seqStart.String(),
				seqEnd.String(),
				err,
			)
		}

		for i := len(events) - 1; i >= 0; i-- {
			event := events[i]
			if event.BlockTime < cutOffBlockTime {
				// We found an event older than cut-off time.
				break outer
			}
			recentEvents = append(recentEvents, event)
		}

		// If the length of fetched events was smaller than the batch size,
		// there are no more events to fetch.
		if len(events) < s.assetsUnlockedBatchSize {
			break
		}

		sequenceTip = seqStart.SubRaw(1)
	}

	// Ensure events are sorted in ascending order of their unlock sequences.
	sort.Slice(recentEvents, func(i, j int) bool {
		return recentEvents[i].UnlockSequence.LT(recentEvents[j].UnlockSequence)
	})

	return recentEvents, nil
}

// findUnconfirmedAssetsUnlockedEvents finds unconfirmed `AssetsUnlocked` events
// from the passed input events. An event is considered confirmed if there
// was an `AssetsUnlockConfirmed` event emitted for it on Ethereum within
// finalized range of blocks.
func (s *Server) findUnconfirmedAssetsUnlockedEvents(
	ctx context.Context,
	inputEvents []bridgetypes.AssetsUnlockedEvent,
) ([]bridgetypes.AssetsUnlockedEvent, error) {
	// If there are no input events, return early with an empty list.
	if len(inputEvents) == 0 {
		return []bridgetypes.AssetsUnlockedEvent{}, nil
	}

	// Finalized block is considered safe from reorgs as it is 64-96 blocks
	// behind the current tip. It can be used as the end block.
	finalizedBlock, err := s.chain.FinalizedBlock(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get the finalized block: [%w]", err)
	}
	endBlock := finalizedBlock.Uint64()

	// The search range can be limited to avoid excessive chain data usage.
	// When defining the start block we must make sure we cover the entire range
	// in which the input events could have been confirmed on Ethereum.
	startBlock := uint64(0)
	if endBlock > assetsUnlockConfirmedLookBackBlocks {
		startBlock = endBlock - assetsUnlockConfirmedLookBackBlocks
	}

	confirmedEvents, err := s.fetchAssetsUnlockConfirmedEvents(
		startBlock,
		endBlock,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to fetch AssetsUnlockConfirmed events: [%w]",
			err,
		)
	}

	// Store the unlock sequences in a map for an easy look-up.
	confirmedUnlockSequences := make(map[string]bool)
	for _, confirmedEvent := range confirmedEvents {
		eventUnlockSequence := confirmedEvent.UnlockSequenceNumber.String()
		confirmedUnlockSequences[eventUnlockSequence] = true
	}

	// If the event's unlock sequence is not among the confirmed sequences
	// consider the event unconfirmed.
	unconfirmedEvents := []bridgetypes.AssetsUnlockedEvent{}
	for _, event := range inputEvents {
		if !confirmedUnlockSequences[event.UnlockSequence.String()] {
			unconfirmedEvents = append(unconfirmedEvents, event)
		}
	}

	// Sort the unconfirmed events by their unlock sequences in ascending order.
	sort.Slice(unconfirmedEvents, func(i, j int) bool {
		return unconfirmedEvents[i].UnlockSequence.
			LT(unconfirmedEvents[j].UnlockSequence)
	})

	return unconfirmedEvents, nil
}

func (s *Server) fetchNewAssetsUnlockedEvents(ctx context.Context) error {
	sequenceTip, err := s.assetsUnlockedEndpoint.GetAssetsUnlockedSequenceTip(ctx)
	if err != nil {
		return fmt.Errorf(
			"failed to fetch assets unlocked sequence tip: [%w]",
			err,
		)
	}

	// This should never happen. Check just in case.
	if sequenceTip.LT(s.lastAssetsUnlockedSequence) {
		return fmt.Errorf(
			"current AssetsUnlock sequence tip on Mezo lower than "+
				"the last processed unlock sequence (%s vs %s)",
			sequenceTip.String(),
			s.lastAssetsUnlockedSequence.String(),
		)
	}

	// If the unlock sequence tip has not advanced on Mezo return early.
	if sequenceTip.Equal(s.lastAssetsUnlockedSequence) {
		s.logger.Info(
			"no new AssetsUnlocked events to fetch from Mezo",
			"unlock_sequence_tip", s.lastAssetsUnlockedSequence.String(),
		)
		return nil
	}

	// The unlock sequence tip on Mezo is greater than the unlock sequence from
	// the last fetched event. We need to fetch events from Mezo.
	seqStart := s.lastAssetsUnlockedSequence.AddRaw(1)
	seqEnd := sequenceTip.AddRaw(1)

	// ON THE MEZOD SIDE THERE IS A LIMIT OF 10,000 EVENTS THAT CAN BE FETCHED
	// IN ONE REQUEST. EXCEEDING THE LIMIT RESULTS IN AN ERROR.
	// However, since new events are fetched very frequently therefore there is
	// no need to split them into batches as realistically we should never
	// exceed the limit.
	events, err := s.assetsUnlockedEndpoint.GetAssetsUnlockedEvents(
		ctx,
		seqStart,
		seqEnd,
	)
	if err != nil {
		return fmt.Errorf(
			"failed to fetch AssetsUnlocked events [%s, %s): %w",
			seqStart.String(),
			seqEnd.String(),
			err,
		)
	}

	s.attestationMutex.Lock()
	s.attestationQueue = append(s.attestationQueue, events...)
	s.attestationMutex.Unlock()

	s.lastAssetsUnlockedSequence = sequenceTip

	s.logger.Info(
		"Fetched new AssetsUnlocked events from Mezo",
		"event_count", len(events),
		"unlock_sequence_tip", s.lastAssetsUnlockedSequence.String(),
	)

	return nil
}

// fetchAssetsUnlockConfirmedEvents retrieves raw `AssetsUnlockConfirmed` ABI
// events from the MezoBridge contract within a specified block range. The
// function fetches events in batches if the entire range is too large to fetch
// at once.
func (s *Server) fetchAssetsUnlockConfirmedEvents(
	startBlock uint64,
	endBlock uint64,
) ([]*portal.MezoBridgeAssetsUnlockConfirmed, error) {
	s.logger.Info(
		"fetching AssetsUnlockConfirmed events from range",
		"startBlock", startBlock,
		"endBlock", endBlock,
	)

	result := make([]*portal.MezoBridgeAssetsUnlockConfirmed, 0)

	ticker := time.NewTicker(time.Minute / time.Duration(s.requestsPerMinute)) //nolint:gosec
	defer ticker.Stop()

	events, err := s.bridgeContract.PastAssetsUnlockConfirmedEvents(
		startBlock,
		&endBlock,
		nil,
		nil,
		nil,
	)
	if err != nil {
		s.logger.Warn(
			"failed to fetch AssetsUnlockConfirmed events from the entire "+
				"range; falling back to batched events fetch",
			"startBlock", startBlock,
			"endBlock", endBlock,
			"err", err,
		)

		batchStartBlock := startBlock

		for batchStartBlock <= endBlock {
			batchEndBlock := batchStartBlock + s.batchSize
			if batchEndBlock > endBlock {
				batchEndBlock = endBlock
			}

			s.logger.Info(
				"fetching a batch of AssetsUnlockConfirmed events from range",
				"batchStartBlock", batchStartBlock,
				"batchEndBlock", batchEndBlock,
			)

			<-ticker.C

			batchEvents, batchErr := s.bridgeContract.PastAssetsUnlockConfirmedEvents(
				batchStartBlock,
				&batchEndBlock,
				nil,
				nil,
				nil,
			)
			if batchErr != nil {
				return nil, fmt.Errorf(
					"batched AssetsUnlockConfirmed fetch failed: [%w]; giving up",
					batchErr,
				)
			}

			result = append(result, batchEvents...)

			batchStartBlock = batchEndBlock + 1
		}
	} else {
		result = append(result, events...)
	}

	return result, nil
}

func (s *Server) unqueueAttestation() *bridgetypes.AssetsUnlockedEvent {
	s.attestationMutex.Lock()
	defer s.attestationMutex.Unlock()
	if len(s.attestationQueue) == 0 {
		return nil
	}

	attestation := s.attestationQueue[0]
	s.attestationQueue = s.attestationQueue[1:]

	return &attestation
}

func (s *Server) attestAssetsUnlockedEvents(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			for attestation := s.unqueueAttestation(); attestation != nil; attestation = s.unqueueAttestation() {
				bridgeAssetsUnlocked := &portal.MezoBridgeAssetsUnlocked{
					UnlockSequenceNumber: attestation.UnlockSequence.BigInt(),
					Recipient:            attestation.Recipient,
					Token:                common.HexToAddress(attestation.Token),
					Amount:               attestation.Amount.BigInt(),
					Chain:                uint8(attestation.Chain), //nolint:gosec // G115: Chain is known to be within uint8 range
				}

				ok, err := s.attestationValidator.IsConfirmed(bridgeAssetsUnlocked)
				if err != nil {
					// we are just logging here so we can move into the attestation loop anyway
					s.logger.Error("couldn't confirm attestation", "attestation", attestation.String(), "error", err)
				}
				if ok {
					s.logger.Info("attestation already confirmed", "attestation", attestation.String())
					continue
				}

				delay := s.submissionQueue.GetSubmissionDelay(bridgeAssetsUnlocked)

				s.logger.Info("waiting for attestation submission slot", "delay", delay)

				// wait for our turn to submit
				select {
				case <-time.After(delay):
					s.logger.Info("starting processing attestation", "attestation", attestation.String())
				case <-ctx.Done():
					s.logger.Info("stopping assets unlocked attestations to context cancellation")
					return
				}

				withBackoff := false
				for {
					if withBackoff {
						time.Sleep(10 * time.Second)
					} else {
						// start with backoff from next iteration
						withBackoff = true
					}

					ok, err := s.attestationValidator.IsConfirmed(bridgeAssetsUnlocked)
					if err != nil {
						s.logger.Error("couldn't confirm attestation", "attestation", attestation.String(), "error", err)
						continue
					}
					if ok {
						s.logger.Info("attestation already confirmed", "attestation", attestation.String())
						break
					}

					s.logger.Info("attestation status", "attestation", attestation.String(), "error", err)

					tx, err := s.bridgeContract.AttestBridgeOut(bridgeAssetsUnlocked)
					if err != nil {
						s.logger.Error("error sending attestation %cv to MezoBridge: %v", attestation.String(), err)
						// just log the error then try again
						continue
					}

					s.logger.Info("attestation sent, waiting for confirmation", "txHash", tx.Hash().Hex())

					// nil is latest block
					latestBlock, err := s.chain.LatestBlock(ctx)
					if err != nil {
						s.logger.Error("couldn't get chain latest block", "error", err)
						continue
					}

					ok, err = s.attestationValidator.WaitForAttestationConfirmation(
						s.blockHeightWaiter,
						latestBlock.Uint64(),
						32, // this is 1 epoch // safe block
						bridgeAssetsUnlocked,
					)
					if err != nil {
						s.logger.Error("couldn't confirm attestation transaction", attestation.String(), err)
						continue
					}
					if ok {
						s.logger.Info("attestation confirmed successfully")
						break
					}

				}
			}
		case <-ctx.Done():
			s.logger.Info(
				"stopping assets unlocked attestations due to context cancellation",
			)
			return
		}
	}
}

// Version return the current version of the ethereum sidecar.
func (s *Server) Version(
	_ context.Context,
	_ *pb.VersionRequest,
) (*pb.VersionResponse, error) {
	return &pb.VersionResponse{
		Version: version.AppVersion,
	}, nil
}

// AssetsLockedEvents returns a list of AssetsLocked events based on the
// passed request. It is executed by the gRPC server.
func (s *Server) AssetsLockedEvents(
	_ context.Context,
	req *pb.AssetsLockedEventsRequest,
) (
	*pb.AssetsLockedEventsResponse,
	error,
) {
	s.assetsLockedEventsMutex.RLock()
	defer s.assetsLockedEventsMutex.RUnlock()

	start, end := req.SequenceStart, req.SequenceEnd

	// The sequence start must be lower than the sequence end if both values are
	// non-nil.
	if !start.IsNil() && !end.IsNil() && start.GTE(end) {
		return nil, fmt.Errorf("sequence start is not lower than sequence end")
	}

	// The sequence start and end must be positive.
	if (!start.IsNil() && !start.IsPositive()) || (!end.IsNil() && !end.IsPositive()) {
		return nil, fmt.Errorf("invalid non positive sequence range")
	}

	// Filter events that fit into the requested range.
	filteredEvents := []*bridgetypes.AssetsLockedEvent{}
	for _, event := range s.assetsLockedEvents {
		if (start.IsNil() || event.Sequence.GTE(start)) && (end.IsNil() || event.Sequence.LT(end)) {
			filteredEvents = append(filteredEvents, &bridgetypes.AssetsLockedEvent{
				Sequence:  event.Sequence,
				Recipient: event.Recipient,
				Token:     event.Token,
				Amount:    event.Amount,
			})
		}
	}

	return &pb.AssetsLockedEventsResponse{
		Events: filteredEvents,
	}, nil
}

// Construct a new instance of the Ethereum MezoBridge contract.
func initializeBridgeContract(
	address common.Address,
	chain *ethconnect.BaseChain,
) (*portal.MezoBridge, error) {
	bridgeContract, err := portal.NewMezoBridge(
		address,
		chain.ChainID(),
		chain.Key(),
		chain.Client(),
		chain.NonceManager(),
		chain.MiningWaiter(),
		chain.BlockCounter(),
		chain.TransactionMutex(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to attach to MezoBridge contract. %v", err)
	}

	return bridgeContract, nil
}
