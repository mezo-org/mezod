package sidecar

import (
	"context"
	"math/big"
	"testing"

	sdkmath "cosmossdk.io/math"
	pb "github.com/mezo-org/mezod/ethereum/sidecar/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	bridgetypes "github.com/mezo-org/mezod/x/bridge/types"
)

func TestAssetsLockedEvents(t *testing.T) {
	server := &Server{
		events: []bridgetypes.AssetsLockedEvent{
			{Sequence: sdkmath.NewIntFromBigInt(big.NewInt(1)), Recipient: "recipient1", Amount: sdkmath.NewIntFromBigInt(big.NewInt(100))},
			{Sequence: sdkmath.NewIntFromBigInt(big.NewInt(2)), Recipient: "recipient2", Amount: sdkmath.NewIntFromBigInt(big.NewInt(200))},
		},
	}

	req := &pb.AssetsLockedEventsRequest{
		SequenceStart: sdkmath.NewInt(1),
		SequenceEnd:   sdkmath.NewInt(3),
	}

	resp, err := server.AssetsLockedEvents(context.Background(), req)

	require.NoError(t, err)
	assert.Len(t, resp.Events, 2)
	assert.Equal(t, "recipient1", resp.Events[0].Recipient)
	assert.Equal(t, "recipient2", resp.Events[1].Recipient)
	assert.Equal(t, int64(100), resp.Events[0].Amount.Int64())
	assert.Equal(t, int64(200), resp.Events[1].Amount.Int64())
	assert.Equal(t, int64(1), resp.Events[0].Sequence.Int64())
	assert.Equal(t, int64(2), resp.Events[1].Sequence.Int64())
}
