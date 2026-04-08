package types

import (
	"fmt"

	evmtypes "github.com/mezo-org/mezod/x/evm/types"

	sdkmath "cosmossdk.io/math"
)

// DefaultGenesis returns the default genesis state.
//
// WARNING: The default genesis state has an empty source BTC token, hence
// it is invalid (Validate will fail). A proper BTC token must be set at
// later stages, before running the network.
func DefaultGenesis() *GenesisState {
	return &GenesisState{
		Params:                         DefaultParams(),
		AssetsLockedSequenceTip:        sdkmath.NewInt(0),
		AssetsUnlockedSequenceTip:      sdkmath.NewInt(0),
		SourceBtcToken:                 "",
		Erc20TokensMappings:            nil,
		InitialBtcSupply:               sdkmath.NewInt(0),
		AssetsUnlockedEvents:           nil,
		BitcoinChainMinBridgeOutAmount: sdkmath.NewInt(0),
		TokenMinBridgeOutAmounts:       nil,
		Pauser:                         evmtypes.ZeroHexAddress(),
		LastOutflowReset:               0,
		CurrentOutflowLimits:           nil,
		CurrentOutflowAmounts:          nil,
		AllowedTripartyControllers:     nil,
		TripartyPaused:                 false,
		TripartyBlockDelay:             1,
		TripartyPerRequestLimit:        sdkmath.NewInt(0),
		TripartyWindowLimit:            sdkmath.NewInt(0),
		TripartyRequestSequenceTip:     sdkmath.NewInt(0),
		TripartyProcessedSequenceTip:   sdkmath.NewInt(0),
		TripartyPendingRequests:        nil,
		TripartyWindowConsumed:         sdkmath.NewInt(0),
		TripartyWindowLastReset:        0,
		TripartyControllerBtcMinted:    nil,
	}
}

// Validate performs basic genesis state validation returning an error upon any
// failure.
func (gs GenesisState) Validate() error {
	if err := gs.Params.Validate(); err != nil {
		return err
	}

	if gs.AssetsLockedSequenceTip.IsNegative() {
		return fmt.Errorf(
			"genesis assets locked sequence tip cannot be negative: %s",
			gs.AssetsLockedSequenceTip,
		)
	}

	if gs.AssetsUnlockedSequenceTip.IsNegative() {
		return fmt.Errorf(
			"genesis assets unlocked sequence tip cannot be negative: %s",
			gs.AssetsUnlockedSequenceTip,
		)
	}

	if len(gs.SourceBtcToken) == 0 {
		return fmt.Errorf("source btc token cannot be empty")
	}

	if !evmtypes.IsHexAddress(gs.SourceBtcToken) {
		return fmt.Errorf("source btc token must be a valid hex-encoded EVM address")
	}

	for i, mapping := range gs.Erc20TokensMappings {
		if len(mapping.SourceToken) == 0 {
			return fmt.Errorf(
				"source token of ERC20 mapping %d cannot be empty",
				i,
			)
		}

		if !evmtypes.IsHexAddress(mapping.SourceToken) {
			return fmt.Errorf(
				"source token of ERC20 mapping %d must be a valid hex-encoded EVM address",
				i,
			)
		}

		if len(mapping.MezoToken) == 0 {
			return fmt.Errorf(
				"mezo token of ERC20 mapping %d cannot be empty",
				i,
			)
		}

		if !evmtypes.IsHexAddress(mapping.MezoToken) {
			return fmt.Errorf(
				"mezo token of ERC20 mapping %d must be a valid hex-encoded EVM address",
				i,
			)
		}
	}

	if gs.TripartyBlockDelay < 1 {
		return fmt.Errorf(
			"genesis triparty block delay cannot be less than 1: %d",
			gs.TripartyBlockDelay,
		)
	}

	if gs.TripartyPerRequestLimit.IsNegative() {
		return fmt.Errorf(
			"genesis triparty per-request limit cannot be negative: %s",
			gs.TripartyPerRequestLimit,
		)
	}

	if gs.TripartyWindowLimit.IsNegative() {
		return fmt.Errorf(
			"genesis triparty window limit cannot be negative: %s",
			gs.TripartyWindowLimit,
		)
	}

	if gs.TripartyRequestSequenceTip.IsNegative() {
		return fmt.Errorf(
			"genesis triparty request sequence tip cannot be negative: %s",
			gs.TripartyRequestSequenceTip,
		)
	}

	if gs.TripartyProcessedSequenceTip.IsNegative() {
		return fmt.Errorf(
			"genesis triparty processed sequence tip cannot be negative: %s",
			gs.TripartyProcessedSequenceTip,
		)
	}

	if gs.TripartyRequestSequenceTip.LT(gs.TripartyProcessedSequenceTip) {
		return fmt.Errorf(
			"genesis triparty request sequence tip cannot be less than processed sequence tip: %s < %s",
			gs.TripartyRequestSequenceTip,
			gs.TripartyProcessedSequenceTip,
		)
	}

	if gs.TripartyWindowConsumed.IsNegative() {
		return fmt.Errorf(
			"genesis triparty window consumed cannot be negative: %s",
			gs.TripartyWindowConsumed,
		)
	}

	seenControllerMinted := make(map[string]struct{}, len(gs.TripartyControllerBtcMinted))
	for i, entry := range gs.TripartyControllerBtcMinted {
		if len(entry.Controller) == 0 {
			return fmt.Errorf("triparty controller BTC minted entry %d controller cannot be empty", i)
		}

		if !evmtypes.IsHexAddress(entry.Controller) {
			return fmt.Errorf(
				"triparty controller BTC minted entry %d controller must be a valid hex-encoded EVM address",
				i,
			)
		}

		if evmtypes.IsZeroHexAddress(entry.Controller) {
			return fmt.Errorf(
				"triparty controller BTC minted entry %d controller cannot be the zero EVM address",
				i,
			)
		}

		if !entry.Amount.IsPositive() {
			return fmt.Errorf(
				"triparty controller BTC minted entry %d amount must be positive: %s",
				i,
				entry.Amount,
			)
		}

		normalizedController := evmtypes.BytesToHexAddress(evmtypes.HexAddressToBytes(entry.Controller))
		if _, ok := seenControllerMinted[normalizedController]; ok {
			return fmt.Errorf(
				"triparty controller BTC minted entry %d has duplicate controller: %s",
				i,
				entry.Controller,
			)
		}
		seenControllerMinted[normalizedController] = struct{}{}
	}

	for i, controller := range gs.AllowedTripartyControllers {
		if len(controller) == 0 {
			return fmt.Errorf("allowed triparty controller %d cannot be empty", i)
		}

		if !evmtypes.IsHexAddress(controller) {
			return fmt.Errorf(
				"allowed triparty controller %d must be a valid hex-encoded EVM address",
				i,
			)
		}

		if evmtypes.IsZeroHexAddress(controller) {
			return fmt.Errorf(
				"allowed triparty controller %d cannot be the zero EVM address",
				i,
			)
		}
	}

	for i, req := range gs.TripartyPendingRequests {
		if req.Sequence.IsNegative() || req.Sequence.IsZero() {
			return fmt.Errorf(
				"pending triparty request %d sequence must be positive: %s",
				i,
				req.Sequence,
			)
		}

		if req.BlockHeight < 0 {
			return fmt.Errorf(
				"pending triparty request %d block height cannot be negative: %d",
				i,
				req.BlockHeight,
			)
		}

		if len(req.Recipient) == 0 {
			return fmt.Errorf("pending triparty request %d recipient cannot be empty", i)
		}

		if !evmtypes.IsHexAddress(req.Recipient) {
			return fmt.Errorf(
				"pending triparty request %d recipient must be a valid hex-encoded EVM address",
				i,
			)
		}

		if evmtypes.IsZeroHexAddress(req.Recipient) {
			return fmt.Errorf(
				"pending triparty request %d recipient cannot be the zero EVM address",
				i,
			)
		}

		if len(req.Controller) == 0 {
			return fmt.Errorf("pending triparty request %d controller cannot be empty", i)
		}

		if !evmtypes.IsHexAddress(req.Controller) {
			return fmt.Errorf(
				"pending triparty request %d controller must be a valid hex-encoded EVM address",
				i,
			)
		}

		if evmtypes.IsZeroHexAddress(req.Controller) {
			return fmt.Errorf(
				"pending triparty request %d controller cannot be the zero EVM address",
				i,
			)
		}

		if req.Amount.IsNegative() || req.Amount.IsZero() {
			return fmt.Errorf(
				"pending triparty request %d amount must be positive: %s",
				i,
				req.Amount,
			)
		}

		if len(req.CallbackData) > MaxTripartyCallbackDataLength {
			return fmt.Errorf(
				"pending triparty request %d callback data exceeds maximum length",
				i,
			)
		}
	}

	for i, req := range gs.TripartyPendingRequests {
		if req.Sequence.LTE(gs.TripartyProcessedSequenceTip) {
			return fmt.Errorf(
				"pending triparty request %d sequence must be greater than processed sequence tip: %s <= %s",
				i,
				req.Sequence,
				gs.TripartyProcessedSequenceTip,
			)
		}

		if req.Sequence.GT(gs.TripartyRequestSequenceTip) {
			return fmt.Errorf(
				"pending triparty request %d sequence cannot be greater than request sequence tip: %s > %s",
				i,
				req.Sequence,
				gs.TripartyRequestSequenceTip,
			)
		}
	}

	expectedPendingRequests := gs.TripartyRequestSequenceTip.Sub(gs.TripartyProcessedSequenceTip)
	actualPendingRequests := sdkmath.NewInt(int64(len(gs.TripartyPendingRequests)))
	if !expectedPendingRequests.Equal(actualPendingRequests) {
		return fmt.Errorf(
			"pending triparty requests must form a gapless range between processed and request sequence tips",
		)
	}

	pendingSequences := make(map[string]struct{}, len(gs.TripartyPendingRequests))
	for _, req := range gs.TripartyPendingRequests {
		pendingSequences[req.Sequence.String()] = struct{}{}
	}

	for seq := gs.TripartyProcessedSequenceTip.AddRaw(1); !seq.GT(gs.TripartyRequestSequenceTip); seq = seq.AddRaw(1) {
		if _, ok := pendingSequences[seq.String()]; !ok {
			return fmt.Errorf(
				"pending triparty requests must form a gapless range between processed and request sequence tips",
			)
		}
	}

	return nil
}
