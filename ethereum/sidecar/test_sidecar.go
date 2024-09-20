package sidecar

import (
	"context"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	bridgetypes "github.com/mezo-org/mezod/x/bridge/types"
)

// TODO: Once real Ethereum sidecar is implemented, remove this file.

// precision is the number of decimal places in the amount of assets locked.
var precision = new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil)

// TestSidecar is a test implementation of the Ethereum sidecar. It generates
// random AssetsLockedEvents every second.
type TestSidecar struct {
	sequenceTip sdkmath.Int
	events      []bridgetypes.AssetsLockedEvent
}

// RunTestSidecar starts a new TestSidecar instance and returns a pointer to it.
// The sidecar will run until the provided context is canceled.
func RunTestSidecar(ctx context.Context) *TestSidecar {
	sidecar := &TestSidecar{
		sequenceTip: sdkmath.ZeroInt(),
		events:      make([]bridgetypes.AssetsLockedEvent, 0),
	}

	go sidecar.run(ctx)

	return sidecar
}

func (ts *TestSidecar) run(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			//nolint:gosec
			eventsCount := 10

			for i := 0; i < eventsCount; i++ {
				ts.sequenceTip = ts.sequenceTip.Add(sdkmath.OneInt())

				amount := new(big.Int).Mul(
					//nolint:gosec
					ts.sequenceTip.Mul(sdkmath.NewInt(10)).BigInt(),
					precision,
				)

				// Just an arbitrary address for testing purposes.
				recipient := sdk.AccAddress(
					common.HexToAddress("0x06EeCc4C2fAC5548a5d09e1905F8Dc21AA01E13A").Bytes(),
				)

				ts.events = append(
					ts.events,
					bridgetypes.AssetsLockedEvent{
						Sequence:  ts.sequenceTip,
						Recipient: recipient.String(),
						Amount:    sdkmath.NewIntFromBigInt(amount),
					},
				)

				// Prune old events. Once the cache reaches 10000 elements,
				// remove the oldest 100.
				if len(ts.events) > 10000 {
					ts.events = ts.events[100:]
				}
			}
		case <-ctx.Done():
			return
		}
	}
}

// GetAssetsLockedEvents is the implementation of
// x/bridge/abci/types.EthereumSidecarClient.GetAssetsLockedEvents.
func (ts *TestSidecar) GetAssetsLockedEvents(
	_ context.Context,
	sequenceStart *sdkmath.Int,
	sequenceEnd *sdkmath.Int,
) ([]bridgetypes.AssetsLockedEvent, error) {
	matchingEvents := make([]bridgetypes.AssetsLockedEvent, 0)

	for _, event := range ts.events {
		start := false
		if sequenceStart == nil || event.Sequence.GTE(*sequenceStart) {
			start = true
		}

		end := false
		if sequenceEnd == nil || event.Sequence.LT(*sequenceEnd) {
			end = true
		}

		if start && end {
			matchingEvents = append(matchingEvents, event)
		}
	}

	return matchingEvents, nil
}
