package sidecar

import (
	"context"

	sdkmath "cosmossdk.io/math"
	bridgetypes "github.com/mezo-org/mezod/x/bridge/types"
)

// ClientMock is the mock implementation of the Ethereum sidecar client used
// in contexts where there is no need for a server connection.
type ClientMock struct{}

func NewClientMock() *ClientMock {
	return &ClientMock{}
}

func (cm *ClientMock) GetAssetsLockedEvents(
	_ context.Context,
	_ *sdkmath.Int,
	_ *sdkmath.Int,
) ([]bridgetypes.AssetsLockedEvent, error) {
	return nil, nil
}
