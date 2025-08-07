package sidecar

import (
	"context"
	"fmt"

	sdkmath "cosmossdk.io/math"
	bridgetypes "github.com/mezo-org/mezod/x/bridge/types"
)

func newLocalAssetsUnlockedEndpoint() *localAssetsUnlockedEndpoint {
	return &localAssetsUnlockedEndpoint{}
}

type localAssetsUnlockedEndpoint struct {
	events []bridgetypes.AssetsUnlockedEvent
}

func (laue *localAssetsUnlockedEndpoint) GetAssetsUnlockedSequenceTip(
	_ context.Context,
) (sdkmath.Int, error) {
	tip := sdkmath.NewInt(0)

	for _, event := range laue.events {
		if event.UnlockSequence.GT(tip) {
			tip = event.UnlockSequence
		}
	}

	return tip, nil
}

func (laue *localAssetsUnlockedEndpoint) GetAssetsUnlockedEvents(
	_ context.Context,
	sequenceStart sdkmath.Int,
	sequenceEnd sdkmath.Int,
) ([]bridgetypes.AssetsUnlockedEvent, error) {
	if sequenceStart.LT(sdkmath.OneInt()) || sequenceStart.GTE(sequenceEnd) {
		return nil, fmt.Errorf("wrong sequence range")
	}

	events := []bridgetypes.AssetsUnlockedEvent{}
	for _, event := range laue.events {
		if event.UnlockSequence.GTE(sequenceStart) &&
			event.UnlockSequence.LT(sequenceEnd) {
			events = append(events, event)
		}
	}

	return events, nil
}

func (laue *localAssetsUnlockedEndpoint) SetAssetsUnlockedEvents(
	events []bridgetypes.AssetsUnlockedEvent,
) {
	laue.events = events
}
