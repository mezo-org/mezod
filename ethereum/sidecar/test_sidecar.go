package sidecar

import (
	"context"
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/mezo-org/mezod/crypto/ethsecp256k1"
	bridgetypes "github.com/mezo-org/mezod/x/bridge/types"
	"math/big"
	"math/rand"
	"time"
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
			eventsCount := rand.Intn(11) // [0, 10] events

			for i := 0; i < eventsCount; i++ {
				ts.sequenceTip = ts.sequenceTip.Add(sdkmath.OneInt())

				amount := new(big.Int).Mul(
					big.NewInt(rand.Int63n(10) + 1),
					precision,
				)

				key, err := ethsecp256k1.GenerateKey()
				if err != nil {
					panic(err)
				}

				recipient := sdk.AccAddress(key.PubKey().Address().Bytes())

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

