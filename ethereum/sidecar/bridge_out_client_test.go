package sidecar

import (
	"context"
	"fmt"

	sdkmath "cosmossdk.io/math"
	bridgetypes "github.com/mezo-org/mezod/x/bridge/types"
)

func newLocalBridgeOutClient() *localBridgeOutClient {
	return &localBridgeOutClient{}
}

type localBridgeOutClient struct {
	events []bridgetypes.AssetsUnlockedEvent
}

func (lboc *localBridgeOutClient) GetAssetsUnlockedSequenceTip(
	_ context.Context,
) (sdkmath.Int, error) {
	tip := sdkmath.NewInt(0)

	for _, event := range lboc.events {
		if event.UnlockSequence.GT(tip) {
			tip = event.UnlockSequence
		}
	}

	return tip, nil
}

func (lboc *localBridgeOutClient) GetAssetsUnlockedEvents(
	_ context.Context,
	sequenceStart sdkmath.Int,
	sequenceEnd sdkmath.Int,
) ([]bridgetypes.AssetsUnlockedEvent, error) {
	if sequenceStart.LT(sdkmath.OneInt()) || sequenceStart.GTE(sequenceEnd) {
		return nil, fmt.Errorf("wrong sequence range")
	}

	events := []bridgetypes.AssetsUnlockedEvent{}
	for _, event := range lboc.events {
		if event.UnlockSequence.GTE(sequenceStart) &&
			event.UnlockSequence.LT(sequenceEnd) {
			events = append(events, event)
		}
	}

	return events, nil
}

func (lboc *localBridgeOutClient) SetAssetsUnlockedEvents(
	events []bridgetypes.AssetsUnlockedEvent,
) {
	lboc.events = events
}
