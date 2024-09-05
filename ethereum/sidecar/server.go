package sidecar

import (
	"context"
	"math/big"
	"math/rand"
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/mezo-org/mezod/crypto/ethsecp256k1"
	bridgetypes "github.com/mezo-org/mezod/x/bridge/types"
)

// Server observes events emitted by the Mezo `BitcoinBridge` contract on the
// Ethereum chain. It enables retrieval of information on the assets locked by
// the contract. It is intended to be run as a separate process.
type Server struct {
	sequenceTip sdkmath.Int
	events      []bridgetypes.AssetsLockedEvent
}

func RunSever(ctx context.Context) *Server {
	server := &Server{
		sequenceTip: sdkmath.ZeroInt(),
		events:      make([]bridgetypes.AssetsLockedEvent, 0),
	}

	go server.run(ctx)

	return server
}

func (s *Server) run(ctx context.Context) {
	// TODO: Replace with the code that actually observes the Ethereum chain.
	//       Temporarily generate some dummy events.
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			//nolint:gosec
			eventsCount := rand.Intn(11) // [0, 10] events

			for i := 0; i < eventsCount; i++ {
				s.sequenceTip = s.sequenceTip.Add(sdkmath.OneInt())

				amount := new(big.Int).Mul(
					//nolint:gosec
					big.NewInt(rand.Int63n(10)+1),
					precision,
				)

				key, err := ethsecp256k1.GenerateKey()
				if err != nil {
					panic(err)
				}

				recipient := sdk.AccAddress(key.PubKey().Address().Bytes())

				s.events = append(
					s.events,
					bridgetypes.AssetsLockedEvent{
						Sequence:  s.sequenceTip,
						Recipient: recipient.String(),
						Amount:    sdkmath.NewIntFromBigInt(amount),
					},
				)

				// Prune old events. Once the cache reaches 10000 elements,
				// remove the oldest 100.
				if len(s.events) > 10000 {
					s.events = s.events[100:]
				}
			}
		case <-ctx.Done():
			return
		}
	}
}
