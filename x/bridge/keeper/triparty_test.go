package keeper

import (
	"testing"

	"cosmossdk.io/math"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	evmtypes "github.com/mezo-org/mezod/x/evm/types"
	"github.com/mezo-org/mezod/x/bridge/types"
	"github.com/stretchr/testify/require"
)

const (
	testTripartyRecipient  = "0x0101010101010101010101010101010101010101"
	testTripartyController = "0x0202020202020202020202020202020202020202"
)

func TestTripartyBlockDelayManagement(t *testing.T) {
	ctx, keeper := mockContext()

	// Initially should return default value of 1
	delay := keeper.GetTripartyBlockDelay(ctx)
	require.Equal(t, uint64(1), delay, "initial delay should be 1")

	// Set a delay
	keeper.SetTripartyBlockDelay(ctx, 100)

	// Get the delay
	delay = keeper.GetTripartyBlockDelay(ctx)
	require.Equal(t, uint64(100), delay, "delay should match what was set")

	// Update the delay
	keeper.SetTripartyBlockDelay(ctx, 200)
	delay = keeper.GetTripartyBlockDelay(ctx)
	require.Equal(t, uint64(200), delay, "delay should be updated")

	// Set back to minimum
	keeper.SetTripartyBlockDelay(ctx, 1)
	delay = keeper.GetTripartyBlockDelay(ctx)
	require.Equal(t, uint64(1), delay, "delay should be 1")
}

func TestTripartyPerRequestLimitManagement(t *testing.T) {
	ctx, keeper := mockContext()

	// Initially should return zero
	limit := keeper.GetTripartyPerRequestLimit(ctx)
	require.True(t, limit.IsZero(), "initial limit should be zero")

	// Set a limit
	keeper.SetTripartyPerRequestLimit(ctx, math.NewInt(1000000))

	// Get the limit
	limit = keeper.GetTripartyPerRequestLimit(ctx)
	require.Equal(t, math.NewInt(1000000), limit, "limit should match what was set")

	// Update the limit
	keeper.SetTripartyPerRequestLimit(ctx, math.NewInt(2000000))
	limit = keeper.GetTripartyPerRequestLimit(ctx)
	require.Equal(t, math.NewInt(2000000), limit, "limit should be updated")

	// Set zero limit
	keeper.SetTripartyPerRequestLimit(ctx, math.ZeroInt())
	limit = keeper.GetTripartyPerRequestLimit(ctx)
	require.True(t, limit.IsZero(), "limit should be zero")
}

func TestTripartyWindowLimitManagement(t *testing.T) {
	ctx, keeper := mockContext()

	// Initially should return zero
	limit := keeper.GetTripartyWindowLimit(ctx)
	require.True(t, limit.IsZero(), "initial limit should be zero")

	// Set a limit
	keeper.SetTripartyWindowLimit(ctx, math.NewInt(5000000))

	// Get the limit
	limit = keeper.GetTripartyWindowLimit(ctx)
	require.Equal(t, math.NewInt(5000000), limit, "limit should match what was set")

	// Update the limit
	keeper.SetTripartyWindowLimit(ctx, math.NewInt(10000000))
	limit = keeper.GetTripartyWindowLimit(ctx)
	require.Equal(t, math.NewInt(10000000), limit, "limit should be updated")

	// Set zero limit
	keeper.SetTripartyWindowLimit(ctx, math.ZeroInt())
	limit = keeper.GetTripartyWindowLimit(ctx)
	require.True(t, limit.IsZero(), "limit should be zero")
}

func TestCreateTripartyBridgeRequest(t *testing.T) {
	ctx, keeper := mockContext()
	// Set a specific block height for testing.
	ctx = ctx.WithBlockHeader(tmproto.Header{Height: 100})

	amount := math.NewInt(1000)
	callbackData := []byte("test-callback")

	keeper.AllowTripartyController(ctx, evmtypes.HexAddressToBytes(testTripartyController), true)

	// First request should get sequence 1.
	reqID1, err := keeper.CreateTripartyBridgeRequest(
		ctx, testTripartyRecipient, amount, callbackData, testTripartyController,
	)
	require.NoError(t, err)
	require.Equal(t, math.NewInt(1), reqID1)

	// Second request should get sequence 2.
	reqID2, err := keeper.CreateTripartyBridgeRequest(
		ctx, testTripartyRecipient, amount, nil, testTripartyController,

	)
	require.NoError(t, err)
	require.Equal(t, math.NewInt(2), reqID2)

	// Sequence tip should now be 2 (last assigned).
	require.Equal(t, math.NewInt(2), keeper.GetTripartySequenceTip(ctx))

	// Verify the first stored request.
	req1, found := keeper.GetTripartyBridgeRequest(ctx, reqID1)
	require.True(t, found)
	require.Equal(t, int64(100), req1.BlockHeight)
	require.Equal(t, amount, req1.Amount)
	require.Equal(t, callbackData, req1.CallbackData)
	require.Equal(t, testTripartyRecipient, req1.Recipient)
	require.Equal(t, testTripartyController, req1.Controller)

	// Verify the second stored request (nil callback data).
	req2, found := keeper.GetTripartyBridgeRequest(ctx, reqID2)
	require.True(t, found)
	require.Equal(t, int64(100), req2.BlockHeight)
	require.Equal(t, amount, req2.Amount)
	require.Empty(t, req2.CallbackData)
	require.Equal(t, testTripartyRecipient, req2.Recipient)
	require.Equal(t, testTripartyController, req2.Controller)
}

func TestCreateTripartyBridgeRequestUnauthorizedController(t *testing.T) {
	ctx, keeper := mockContext()

	// Controller is not authorized — should be rejected.
	_, err := keeper.CreateTripartyBridgeRequest(
		ctx, testTripartyRecipient, math.NewInt(1000), nil, testTripartyController,
	)
	require.ErrorIs(t, err, types.ErrTripartyControllerNotAllowed)

	// Sequence tip should not have advanced.
	require.True(t, keeper.GetTripartySequenceTip(ctx).IsZero())
}

func TestCreateTripartyBridgeRequestPerRequestLimit(t *testing.T) {
	ctx, keeper := mockContext()

	keeper.AllowTripartyController(ctx, evmtypes.HexAddressToBytes(testTripartyController), true)
	keeper.SetTripartyPerRequestLimit(ctx, math.NewInt(500))

	// Amount exceeding the limit should be rejected.
	_, err := keeper.CreateTripartyBridgeRequest(
		ctx, testTripartyRecipient, math.NewInt(1000), nil, testTripartyController,
	)
	require.ErrorIs(t, err, types.ErrTripartyPerRequestLimitExceeded)

	// Sequence tip should not have advanced.
	require.True(t, keeper.GetTripartySequenceTip(ctx).IsZero())

	// Amount equal to the limit should succeed.
	reqID, err := keeper.CreateTripartyBridgeRequest(
		ctx, testTripartyRecipient, math.NewInt(500), nil, testTripartyController,
	)
	require.NoError(t, err)
	require.Equal(t, math.NewInt(1), reqID)

	// Zero limit disables the check.
	keeper.SetTripartyPerRequestLimit(ctx, math.ZeroInt())

	_, err = keeper.CreateTripartyBridgeRequest(
		ctx, testTripartyRecipient, math.NewInt(999999), nil, testTripartyController,
	)
	require.NoError(t, err)
}

func TestGetTripartyBridgeRequest(t *testing.T) {
	ctx, keeper := mockContext()

	amount := math.NewInt(500)

	keeper.AllowTripartyController(ctx, evmtypes.HexAddressToBytes(testTripartyController), true)

	// Non-existent request returns false.
	_, found := keeper.GetTripartyBridgeRequest(ctx, math.NewInt(1))
	require.False(t, found)

	// Create a request and retrieve it.
	reqID, err := keeper.CreateTripartyBridgeRequest(
		ctx, testTripartyRecipient, amount, nil, testTripartyController,
	)
	require.NoError(t, err)

	req, found := keeper.GetTripartyBridgeRequest(ctx, reqID)
	require.True(t, found)
	require.True(t, reqID.Equal(req.Sequence))
	require.Equal(t, testTripartyRecipient, req.Recipient)
	require.Equal(t, amount, req.Amount)
	require.Empty(t, req.CallbackData)
	require.Equal(t, testTripartyController, req.Controller)
}

func TestDeleteTripartyBridgeRequest(t *testing.T) {
	ctx, keeper := mockContext()

	keeper.AllowTripartyController(ctx, evmtypes.HexAddressToBytes(testTripartyController), true)

	reqID1, err := keeper.CreateTripartyBridgeRequest(
		ctx, testTripartyRecipient, math.NewInt(100), nil, testTripartyController,
	)
	require.NoError(t, err)
	reqID2, err := keeper.CreateTripartyBridgeRequest(
		ctx, testTripartyRecipient, math.NewInt(200), nil, testTripartyController,
	)
	require.NoError(t, err)

	// Both requests exist.
	_, found := keeper.GetTripartyBridgeRequest(ctx, reqID1)
	require.True(t, found)
	_, found = keeper.GetTripartyBridgeRequest(ctx, reqID2)
	require.True(t, found)

	// Deleting the second request while the first exists should fail.
	err = keeper.DeleteTripartyBridgeRequest(ctx, reqID2)
	require.Error(t, err)

	// Deleting the first (oldest) request should succeed.
	err = keeper.DeleteTripartyBridgeRequest(ctx, reqID1)
	require.NoError(t, err)
	_, found = keeper.GetTripartyBridgeRequest(ctx, reqID1)
	require.False(t, found)

	// Now the second request is the oldest; deleting it should succeed.
	err = keeper.DeleteTripartyBridgeRequest(ctx, reqID2)
	require.NoError(t, err)
	_, found = keeper.GetTripartyBridgeRequest(ctx, reqID2)
	require.False(t, found)
}

func TestGetPendingTripartyBridgeRequests(t *testing.T) {
	ctx, keeper := mockContext()

	keeper.AllowTripartyController(ctx, evmtypes.HexAddressToBytes(testTripartyController), true)

	// Create 5 requests.
	for i := 0; i < 5; i++ {
		_, err := keeper.CreateTripartyBridgeRequest(
			ctx,
			testTripartyRecipient,
			math.NewInt(int64(100*(i+1))),
			nil,
			testTripartyController,
		)
		require.NoError(t, err)
	}

	// Read all 5 with limit 10.
	requests := keeper.GetPendingTripartyBridgeRequests(ctx, math.NewInt(1), 10)
	require.Len(t, requests, 5)
	for i, req := range requests {
		require.True(t, math.NewInt(int64(i+1)).Equal(req.Sequence))
		require.Equal(t, math.NewInt(int64(100*(i+1))), req.Amount)
	}

	// Read with limit 3.
	requests = keeper.GetPendingTripartyBridgeRequests(ctx, math.NewInt(1), 3)
	require.Len(t, requests, 3)
	require.True(t, math.NewInt(1).Equal(requests[0].Sequence))
	require.True(t, math.NewInt(3).Equal(requests[2].Sequence))

	// Read starting from sequence 3.
	requests = keeper.GetPendingTripartyBridgeRequests(ctx, math.NewInt(3), 10)
	require.Len(t, requests, 3)
	require.True(t, math.NewInt(3).Equal(requests[0].Sequence))
	require.True(t, math.NewInt(5).Equal(requests[2].Sequence))

	// Read from non-existent sequence.
	requests = keeper.GetPendingTripartyBridgeRequests(ctx, math.NewInt(10), 5)
	require.Empty(t, requests)
}

func TestTripartySequenceTipIncrement(t *testing.T) {
	ctx, keeper := mockContext()

	// Default tip is 0.
	require.True(t, keeper.GetTripartySequenceTip(ctx).IsZero())

	// First increment returns 1.
	require.Equal(t, math.NewInt(1), keeper.incrementTripartySequenceTip(ctx))
	require.Equal(t, math.NewInt(1), keeper.GetTripartySequenceTip(ctx))

	// Second increment returns 2.
	require.Equal(t, math.NewInt(2), keeper.incrementTripartySequenceTip(ctx))
	require.Equal(t, math.NewInt(2), keeper.GetTripartySequenceTip(ctx))
}

func TestTripartySequenceTipDefault(t *testing.T) {
	ctx, keeper := mockContext()

	// Default sequence tip is 0 (no requests assigned yet).
	require.True(t, keeper.GetTripartySequenceTip(ctx).IsZero())
}

func TestTripartyBridgeRequestMarshalRoundtrip(t *testing.T) {
	req := &types.TripartyBridgeRequest{
		Sequence:     math.NewInt(42),
		BlockHeight:  12345,
		Recipient:    "0xAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA",
		Amount:       math.NewInt(999999999),
		CallbackData: []byte("some-callback-data"),
		Controller:   "0xBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBB",
	}

	bz, err := req.Marshal()
	require.NoError(t, err)

	decoded := &types.TripartyBridgeRequest{}
	err = decoded.Unmarshal(bz)
	require.NoError(t, err)

	require.True(t, req.Sequence.Equal(decoded.Sequence))
	require.Equal(t, req.BlockHeight, decoded.BlockHeight)
	require.Equal(t, req.Recipient, decoded.Recipient)
	require.Equal(t, req.Amount, decoded.Amount)
	require.Equal(t, req.CallbackData, decoded.CallbackData)
	require.Equal(t, req.Controller, decoded.Controller)
}

func TestTripartyBridgeRequestMarshalEmptyCallbackData(t *testing.T) {
	req := &types.TripartyBridgeRequest{
		Sequence:     math.NewInt(1),
		BlockHeight:  100,
		Recipient:    testTripartyRecipient,
		Amount:       math.NewInt(500),
		CallbackData: nil,
		Controller:   testTripartyController,
	}

	bz, err := req.Marshal()
	require.NoError(t, err)

	decoded := &types.TripartyBridgeRequest{}
	err = decoded.Unmarshal(bz)
	require.NoError(t, err)

	require.True(t, req.Sequence.Equal(decoded.Sequence))
	require.Empty(t, decoded.CallbackData)
}

func TestTripartyWindowMinted(t *testing.T) {
	ctx, keeper := mockContext()

	// Initially zero.
	require.True(t, keeper.GetTripartyWindowMinted(ctx).IsZero())

	// Increase accumulates.
	keeper.IncreaseTripartyWindowMinted(ctx, math.NewInt(100))
	require.Equal(t, math.NewInt(100), keeper.GetTripartyWindowMinted(ctx))

	keeper.IncreaseTripartyWindowMinted(ctx, math.NewInt(250))
	require.Equal(t, math.NewInt(350), keeper.GetTripartyWindowMinted(ctx))

	// Reset clears to zero and records block height.
	ctx = ctx.WithBlockHeader(tmproto.Header{Height: 500})
	keeper.ResetTripartyWindowMinted(ctx)
	require.True(t, keeper.GetTripartyWindowMinted(ctx).IsZero())
	require.Equal(t, uint64(500), keeper.GetTripartyWindowLastReset(ctx))
}

func TestTripartyCapacity(t *testing.T) {
	ctx, keeper := mockContext()

	// No limit set, capacity is zero.
	capacity, resetHeight := keeper.GetTripartyCapacity(ctx)
	require.True(t, capacity.IsZero())
	require.Equal(t, uint64(25000), resetHeight) // 0 + TripartyWindowResetBlocks

	// Set window limit.
	keeper.SetTripartyWindowLimit(ctx, math.NewInt(1000))

	// Full capacity when nothing minted.
	capacity, _ = keeper.GetTripartyCapacity(ctx)
	require.Equal(t, math.NewInt(1000), capacity)

	// Partial mint reduces capacity.
	keeper.IncreaseTripartyWindowMinted(ctx, math.NewInt(300))
	capacity, _ = keeper.GetTripartyCapacity(ctx)
	require.Equal(t, math.NewInt(700), capacity)

	// Reset at block 50000 updates the reset height.
	ctx = ctx.WithBlockHeader(tmproto.Header{Height: 50000})
	keeper.ResetTripartyWindowMinted(ctx)
	_, resetHeight = keeper.GetTripartyCapacity(ctx)
	require.Equal(t, uint64(75000), resetHeight) // 50000 + 25000
}

func TestCheckTripartyCapacity(t *testing.T) {
	ctx, keeper := mockContext()

	keeper.SetTripartyWindowLimit(ctx, math.NewInt(500))

	// Within capacity - no error.
	require.NoError(t, keeper.CheckTripartyCapacity(ctx, math.NewInt(500)))
	require.NoError(t, keeper.CheckTripartyCapacity(ctx, math.NewInt(1)))

	// Exceeds capacity - error.
	require.Error(t, keeper.CheckTripartyCapacity(ctx, math.NewInt(501)))

	// After partial mint, remaining capacity shrinks.
	keeper.IncreaseTripartyWindowMinted(ctx, math.NewInt(400))
	require.NoError(t, keeper.CheckTripartyCapacity(ctx, math.NewInt(100)))
	require.Error(t, keeper.CheckTripartyCapacity(ctx, math.NewInt(101)))
}

func TestTripartyTotalBTCMinted(t *testing.T) {
	ctx, keeper := mockContext()

	// Initially zero.
	require.True(t, keeper.GetTripartyTotalBTCMinted(ctx).IsZero())

	// Increase accumulates.
	keeper.IncreaseTripartyTotalBTCMinted(ctx, math.NewInt(1000))
	require.Equal(t, math.NewInt(1000), keeper.GetTripartyTotalBTCMinted(ctx))

	keeper.IncreaseTripartyTotalBTCMinted(ctx, math.NewInt(2500))
	require.Equal(t, math.NewInt(3500), keeper.GetTripartyTotalBTCMinted(ctx))
}
