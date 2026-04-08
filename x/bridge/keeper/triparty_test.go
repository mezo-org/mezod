package keeper

import (
	"fmt"
	"testing"

	"cosmossdk.io/math"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/mezo-org/mezod/cmd/config"
	"github.com/mezo-org/mezod/x/bridge/types"
	evmtypes "github.com/mezo-org/mezod/x/evm/types"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

const (
	testTripartyRecipient   = "0x0101010101010101010101010101010101010101"
	testTripartyController  = "0x0202020202020202020202020202020202020202"
	testTripartyController1 = "0x1111111111111111111111111111111111111111"
	testTripartyController2 = "0x2222222222222222222222222222222222222222"
)

var testTripartyRecipientAddr = sdk.AccAddress(evmtypes.HexAddressToBytes(testTripartyRecipient))

func TestTripartyBlockDelayManagement(t *testing.T) {
	ctx, keeper := mockContext()

	// Initially should return default value of 1
	delay := keeper.GetTripartyBlockDelay(ctx)
	require.Equal(t, int64(1), delay, "initial delay should be 1")

	// Set a delay
	require.NoError(t, keeper.SetTripartyBlockDelay(ctx, 100))

	// Get the delay
	delay = keeper.GetTripartyBlockDelay(ctx)
	require.Equal(t, int64(100), delay, "delay should match what was set")

	// Update the delay
	require.NoError(t, keeper.SetTripartyBlockDelay(ctx, 200))
	delay = keeper.GetTripartyBlockDelay(ctx)
	require.Equal(t, int64(200), delay, "delay should be updated")

	// Set back to minimum
	require.NoError(t, keeper.SetTripartyBlockDelay(ctx, 1))
	delay = keeper.GetTripartyBlockDelay(ctx)
	require.Equal(t, int64(1), delay, "delay should be 1")

	// Delay < 1 should return an error
	require.Error(t, keeper.SetTripartyBlockDelay(ctx, 0))
	require.Error(t, keeper.SetTripartyBlockDelay(ctx, -1))
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

	amount := MinTripartyAmount
	callbackData := []byte("test-callback")

	keeper.AllowTripartyController(ctx, evmtypes.HexAddressToBytes(testTripartyController), true)
	keeper.SetTripartyWindowLimit(ctx, to18Dec(100))

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
	require.Equal(t, math.NewInt(2), keeper.GetTripartyRequestSequenceTip(ctx))

	// Verify the first stored request.
	req1, found := keeper.getTripartyBridgeRequest(ctx, reqID1)
	require.True(t, found)
	require.Equal(t, int64(100), req1.BlockHeight)
	require.Equal(t, amount, req1.Amount)
	require.Equal(t, callbackData, req1.CallbackData)
	require.Equal(t, testTripartyRecipient, req1.Recipient)
	require.Equal(t, testTripartyController, req1.Controller)

	// Verify the second stored request (nil callback data).
	req2, found := keeper.getTripartyBridgeRequest(ctx, reqID2)
	require.True(t, found)
	require.Equal(t, int64(100), req2.BlockHeight)
	require.Equal(t, amount, req2.Amount)
	require.Empty(t, req2.CallbackData)
	require.Equal(t, testTripartyRecipient, req2.Recipient)
	require.Equal(t, testTripartyController, req2.Controller)
}

func TestCreateTripartyBridgeRequestPaused(t *testing.T) {
	ctx, keeper := mockContext()

	keeper.AllowTripartyController(ctx, evmtypes.HexAddressToBytes(testTripartyController), true)
	keeper.SetTripartyWindowLimit(ctx, to18Dec(100))
	keeper.SetTripartyPaused(ctx, true)

	// Should be rejected when paused.
	_, err := keeper.CreateTripartyBridgeRequest(
		ctx, testTripartyRecipient, math.NewInt(1000), nil, testTripartyController,
	)
	require.ErrorIs(t, err, types.ErrTripartyPaused)

	// Sequence tip should not have advanced.
	require.True(t, keeper.GetTripartyRequestSequenceTip(ctx).IsZero())

	// Unpause and verify create succeeds.
	keeper.SetTripartyPaused(ctx, false)

	reqID, err := keeper.CreateTripartyBridgeRequest(
		ctx, testTripartyRecipient, MinTripartyAmount, nil, testTripartyController,
	)
	require.NoError(t, err)
	require.Equal(t, math.NewInt(1), reqID)
}

func TestCreateTripartyBridgeRequestInvalidRecipient(t *testing.T) {
	ctx, keeper := mockContext()

	keeper.AllowTripartyController(ctx, evmtypes.HexAddressToBytes(testTripartyController), true)

	// Invalid hex string should be rejected.
	_, err := keeper.CreateTripartyBridgeRequest(
		ctx, "not-a-hex-address", math.NewInt(1000), nil, testTripartyController,
	)
	require.ErrorIs(t, err, types.ErrInvalidEVMAddress)

	// Sequence tip should not have advanced.
	require.True(t, keeper.GetTripartyRequestSequenceTip(ctx).IsZero())
}

func TestCreateTripartyBridgeRequestZeroRecipient(t *testing.T) {
	ctx, keeper := mockContext()

	keeper.AllowTripartyController(ctx, evmtypes.HexAddressToBytes(testTripartyController), true)

	// Zero address should be rejected.
	_, err := keeper.CreateTripartyBridgeRequest(
		ctx, "0x0000000000000000000000000000000000000000", math.NewInt(1000), nil, testTripartyController,
	)
	require.ErrorIs(t, err, types.ErrZeroEVMAddress)

	// Sequence tip should not have advanced.
	require.True(t, keeper.GetTripartyRequestSequenceTip(ctx).IsZero())
}

func TestCreateTripartyBridgeRequestInvalidController(t *testing.T) {
	ctx, keeper := mockContext()

	// Invalid hex string should be rejected.
	_, err := keeper.CreateTripartyBridgeRequest(
		ctx, testTripartyRecipient, math.NewInt(1000), nil, "bad-controller",
	)
	require.ErrorIs(t, err, types.ErrInvalidEVMAddress)

	// Sequence tip should not have advanced.
	require.True(t, keeper.GetTripartyRequestSequenceTip(ctx).IsZero())
}

func TestCreateTripartyBridgeRequestCallbackDataTooLarge(t *testing.T) {
	ctx, keeper := mockContext()

	keeper.AllowTripartyController(ctx, evmtypes.HexAddressToBytes(testTripartyController), true)
	keeper.SetTripartyWindowLimit(ctx, to18Dec(100))

	// 321 bytes exceeds the 320-byte limit.
	_, err := keeper.CreateTripartyBridgeRequest(
		ctx, testTripartyRecipient, MinTripartyAmount, make([]byte, 321), testTripartyController,
	)
	require.ErrorIs(t, err, types.ErrTripartyCallbackDataTooLarge)

	// Sequence tip should not have advanced.
	require.True(t, keeper.GetTripartyRequestSequenceTip(ctx).IsZero())

	// Exactly 320 bytes should succeed.
	reqID, err := keeper.CreateTripartyBridgeRequest(
		ctx, testTripartyRecipient, MinTripartyAmount, make([]byte, 320), testTripartyController,
	)
	require.NoError(t, err)
	require.Equal(t, math.NewInt(1), reqID)
}

func TestCreateTripartyBridgeRequestAmountNotPositive(t *testing.T) {
	ctx, keeper := mockContext()

	keeper.AllowTripartyController(ctx, evmtypes.HexAddressToBytes(testTripartyController), true)

	// Zero amount should be rejected.
	_, err := keeper.CreateTripartyBridgeRequest(
		ctx, testTripartyRecipient, math.ZeroInt(), nil, testTripartyController,
	)
	require.ErrorIs(t, err, types.ErrTripartyAmountNotPositive)

	// Negative amount should be rejected.
	_, err = keeper.CreateTripartyBridgeRequest(
		ctx, testTripartyRecipient, math.NewInt(-1), nil, testTripartyController,
	)
	require.ErrorIs(t, err, types.ErrTripartyAmountNotPositive)

	// Sequence tip should not have advanced.
	require.True(t, keeper.GetTripartyRequestSequenceTip(ctx).IsZero())
}

func TestCreateTripartyBridgeRequestAmountBelowMinimum(t *testing.T) {
	ctx, keeper := mockContext()

	keeper.AllowTripartyController(ctx, evmtypes.HexAddressToBytes(testTripartyController), true)
	keeper.SetTripartyWindowLimit(ctx, to18Dec(100))

	// Amount just below the minimum should be rejected.
	_, err := keeper.CreateTripartyBridgeRequest(
		ctx, testTripartyRecipient, MinTripartyAmount.SubRaw(1), nil, testTripartyController,
	)
	require.ErrorIs(t, err, types.ErrTripartyAmountBelowMinimum)

	// Sequence tip should not have advanced.
	require.True(t, keeper.GetTripartyRequestSequenceTip(ctx).IsZero())

	// Amount exactly at the minimum should succeed.
	reqID, err := keeper.CreateTripartyBridgeRequest(
		ctx, testTripartyRecipient, MinTripartyAmount, nil, testTripartyController,
	)
	require.NoError(t, err)
	require.Equal(t, math.NewInt(1), reqID)
}

func TestCreateTripartyBridgeRequestUnauthorizedController(t *testing.T) {
	ctx, keeper := mockContext()

	// Controller is not authorized — should be rejected.
	_, err := keeper.CreateTripartyBridgeRequest(
		ctx, testTripartyRecipient, MinTripartyAmount, nil, testTripartyController,
	)
	require.ErrorIs(t, err, types.ErrTripartyControllerNotAllowed)

	// Sequence tip should not have advanced.
	require.True(t, keeper.GetTripartyRequestSequenceTip(ctx).IsZero())
}

func TestCreateTripartyBridgeRequestPerRequestLimit(t *testing.T) {
	ctx, keeper := mockContext()

	keeper.AllowTripartyController(ctx, evmtypes.HexAddressToBytes(testTripartyController), true)
	keeper.SetTripartyWindowLimit(ctx, to18Dec(100))
	keeper.SetTripartyPerRequestLimit(ctx, MinTripartyAmount)

	// Amount exceeding the limit should be rejected.
	_, err := keeper.CreateTripartyBridgeRequest(
		ctx, testTripartyRecipient, MinTripartyAmount.AddRaw(1), nil, testTripartyController,
	)
	require.ErrorIs(t, err, types.ErrTripartyPerRequestLimitExceeded)

	// Sequence tip should not have advanced.
	require.True(t, keeper.GetTripartyRequestSequenceTip(ctx).IsZero())

	// Amount equal to the limit should succeed.
	reqID, err := keeper.CreateTripartyBridgeRequest(
		ctx, testTripartyRecipient, MinTripartyAmount, nil, testTripartyController,
	)
	require.NoError(t, err)
	require.Equal(t, math.NewInt(1), reqID)

	// Zero limit disables the check.
	keeper.SetTripartyPerRequestLimit(ctx, math.ZeroInt())

	_, err = keeper.CreateTripartyBridgeRequest(
		ctx, testTripartyRecipient, MinTripartyAmount.MulRaw(10), nil, testTripartyController,
	)
	require.NoError(t, err)
}

func TestGetTripartyBridgeRequest(t *testing.T) {
	ctx, keeper := mockContext()

	amount := MinTripartyAmount

	keeper.AllowTripartyController(ctx, evmtypes.HexAddressToBytes(testTripartyController), true)
	keeper.SetTripartyWindowLimit(ctx, to18Dec(100))

	// Non-existent request returns false.
	_, found := keeper.getTripartyBridgeRequest(ctx, math.NewInt(1))
	require.False(t, found)

	// Create a request and retrieve it.
	reqID, err := keeper.CreateTripartyBridgeRequest(
		ctx, testTripartyRecipient, amount, nil, testTripartyController,
	)
	require.NoError(t, err)

	req, found := keeper.getTripartyBridgeRequest(ctx, reqID)
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
	keeper.SetTripartyWindowLimit(ctx, to18Dec(100))

	reqID1, err := keeper.CreateTripartyBridgeRequest(
		ctx, testTripartyRecipient, MinTripartyAmount, nil, testTripartyController,
	)
	require.NoError(t, err)
	reqID2, err := keeper.CreateTripartyBridgeRequest(
		ctx, testTripartyRecipient, MinTripartyAmount.MulRaw(2), nil, testTripartyController,
	)
	require.NoError(t, err)

	// Both requests exist.
	_, found := keeper.getTripartyBridgeRequest(ctx, reqID1)
	require.True(t, found)
	_, found = keeper.getTripartyBridgeRequest(ctx, reqID2)
	require.True(t, found)

	// Deleting the second request while the first exists should fail.
	err = keeper.deleteTripartyBridgeRequest(ctx, reqID2)
	require.Error(t, err)

	// Deleting the first (oldest) request should succeed.
	err = keeper.deleteTripartyBridgeRequest(ctx, reqID1)
	require.NoError(t, err)
	_, found = keeper.getTripartyBridgeRequest(ctx, reqID1)
	require.False(t, found)

	// Now the second request is the oldest; deleting it should succeed.
	err = keeper.deleteTripartyBridgeRequest(ctx, reqID2)
	require.NoError(t, err)
	_, found = keeper.getTripartyBridgeRequest(ctx, reqID2)
	require.False(t, found)
}

func TestTripartyRequestSequenceTip(t *testing.T) {
	ctx, keeper := mockContext()

	// Default tip is 0.
	require.True(t, keeper.GetTripartyRequestSequenceTip(ctx).IsZero())

	// First increment returns 1.
	require.Equal(t, math.NewInt(1), keeper.incrementTripartyRequestSequenceTip(ctx))
	require.Equal(t, math.NewInt(1), keeper.GetTripartyRequestSequenceTip(ctx))

	// Second increment returns 2.
	require.Equal(t, math.NewInt(2), keeper.incrementTripartyRequestSequenceTip(ctx))
	require.Equal(t, math.NewInt(2), keeper.GetTripartyRequestSequenceTip(ctx))
}

func TestTripartyRequestSequenceTipManagement(t *testing.T) {
	ctx, keeper := mockContext()

	// Initially should be zero.
	require.True(t, keeper.GetTripartyRequestSequenceTip(ctx).IsZero())

	// Set sequence tip.
	keeper.setTripartyRequestSequenceTip(ctx, math.NewInt(5))
	require.Equal(t, math.NewInt(5), keeper.GetTripartyRequestSequenceTip(ctx))

	// Update sequence tip.
	keeper.setTripartyRequestSequenceTip(ctx, math.NewInt(42))
	require.Equal(t, math.NewInt(42), keeper.GetTripartyRequestSequenceTip(ctx))
}

func TestTripartyRequestSequenceTipDefault(t *testing.T) {
	ctx, keeper := mockContext()

	// Default sequence tip is 0 (no requests assigned yet).
	require.True(t, keeper.GetTripartyRequestSequenceTip(ctx).IsZero())
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

func TestTripartyWindowConsumed(t *testing.T) {
	ctx, keeper := mockContext()

	// Initially zero.
	require.True(t, keeper.getTripartyWindowConsumed(ctx).IsZero())

	// Increase accumulates.
	keeper.increaseTripartyWindowConsumed(ctx, math.NewInt(100))
	require.Equal(t, math.NewInt(100), keeper.getTripartyWindowConsumed(ctx))

	keeper.increaseTripartyWindowConsumed(ctx, math.NewInt(250))
	require.Equal(t, math.NewInt(350), keeper.getTripartyWindowConsumed(ctx))

	// Reset clears to zero and records block height.
	ctx = ctx.WithBlockHeader(tmproto.Header{Height: 500})
	keeper.resetTripartyWindowConsumed(ctx)
	require.True(t, keeper.getTripartyWindowConsumed(ctx).IsZero())
	require.Equal(t, uint64(500), keeper.getTripartyWindowLastReset(ctx))
}

func TestTripartyWindowLastResetManagement(t *testing.T) {
	ctx, keeper := mockContext()

	// Initially should be zero.
	require.Equal(t, uint64(0), keeper.getTripartyWindowLastReset(ctx))

	// Set reset height.
	keeper.setTripartyWindowLastReset(ctx, 12345)
	require.Equal(t, uint64(12345), keeper.getTripartyWindowLastReset(ctx))
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

	// Partial consumption reduces capacity.
	keeper.increaseTripartyWindowConsumed(ctx, math.NewInt(300))
	capacity, _ = keeper.GetTripartyCapacity(ctx)
	require.Equal(t, math.NewInt(700), capacity)

	// Reset at block 50000 updates the reset height.
	ctx = ctx.WithBlockHeader(tmproto.Header{Height: 50000})
	keeper.resetTripartyWindowConsumed(ctx)
	_, resetHeight = keeper.GetTripartyCapacity(ctx)
	require.Equal(t, uint64(75000), resetHeight) // 50000 + 25000
}

func TestCheckTripartyCapacity(t *testing.T) {
	ctx, keeper := mockContext()

	keeper.SetTripartyWindowLimit(ctx, math.NewInt(500))

	// Within capacity - no error.
	require.NoError(t, keeper.checkTripartyCapacity(ctx, math.NewInt(500)))
	require.NoError(t, keeper.checkTripartyCapacity(ctx, math.NewInt(1)))

	// Exceeds capacity - error.
	require.Error(t, keeper.checkTripartyCapacity(ctx, math.NewInt(501)))

	// After partial consumption, remaining capacity shrinks.
	keeper.increaseTripartyWindowConsumed(ctx, math.NewInt(400))
	require.NoError(t, keeper.checkTripartyCapacity(ctx, math.NewInt(100)))
	require.Error(t, keeper.checkTripartyCapacity(ctx, math.NewInt(101)))
}

func TestTripartyTotalBTCMinted(t *testing.T) {
	ctx, keeper := mockContext()

	// Initially zero (no per-controller entries).
	require.True(t, keeper.GetTripartyTotalBTCMinted(ctx).IsZero())

	// Total is derived from per-controller entries.
	keeper.increaseTripartyControllerBTCMinted(ctx, testTripartyController1, math.NewInt(1000))
	require.Equal(t, math.NewInt(1000), keeper.GetTripartyTotalBTCMinted(ctx))

	keeper.increaseTripartyControllerBTCMinted(ctx, testTripartyController2, math.NewInt(2500))
	require.Equal(t, math.NewInt(3500), keeper.GetTripartyTotalBTCMinted(ctx))

	// Further increment to an existing controller reflects in total.
	keeper.increaseTripartyControllerBTCMinted(ctx, testTripartyController1, math.NewInt(500))
	require.Equal(t, math.NewInt(4000), keeper.GetTripartyTotalBTCMinted(ctx))
}

func TestTripartyControllerBTCMinted(t *testing.T) {
	ctx, keeper := mockContext()

	controller1 := testTripartyController1
	controller2 := testTripartyController2

	// Initially zero for any controller.
	require.True(t, keeper.GetTripartyControllerBTCMinted(ctx, controller1).IsZero())
	require.True(t, keeper.GetTripartyControllerBTCMinted(ctx, controller2).IsZero())

	// Increase accumulates per controller.
	keeper.increaseTripartyControllerBTCMinted(ctx, controller1, math.NewInt(1000))
	require.Equal(t, math.NewInt(1000), keeper.GetTripartyControllerBTCMinted(ctx, controller1))
	require.True(t, keeper.GetTripartyControllerBTCMinted(ctx, controller2).IsZero())

	keeper.increaseTripartyControllerBTCMinted(ctx, controller2, math.NewInt(500))
	require.Equal(t, math.NewInt(1000), keeper.GetTripartyControllerBTCMinted(ctx, controller1))
	require.Equal(t, math.NewInt(500), keeper.GetTripartyControllerBTCMinted(ctx, controller2))

	// Further accumulation.
	keeper.increaseTripartyControllerBTCMinted(ctx, controller1, math.NewInt(2500))
	require.Equal(t, math.NewInt(3500), keeper.GetTripartyControllerBTCMinted(ctx, controller1))
	require.Equal(t, math.NewInt(500), keeper.GetTripartyControllerBTCMinted(ctx, controller2))
}

func TestGetAllTripartyControllerBTCMinted(t *testing.T) {
	ctx, keeper := mockContext()

	// Initially empty.
	require.Empty(t, keeper.getAllTripartyControllerBTCMinted(ctx))

	// After minting, returns all controllers with amounts.
	controller1 := testTripartyController1
	controller2 := testTripartyController2

	keeper.increaseTripartyControllerBTCMinted(ctx, controller1, math.NewInt(1000))
	keeper.increaseTripartyControllerBTCMinted(ctx, controller2, math.NewInt(500))

	all := keeper.getAllTripartyControllerBTCMinted(ctx)
	require.Len(t, all, 2)

	// Build a map for order-independent assertion.
	amounts := make(map[string]math.Int)
	for _, entry := range all {
		amounts[entry.Controller] = entry.Amount
	}
	require.Equal(t, math.NewInt(1000), amounts[testTripartyController1])
	require.Equal(t, math.NewInt(500), amounts[testTripartyController2])
}

func TestTripartyProcessedSequenceTip(t *testing.T) {
	ctx, keeper := mockContext()

	// Default is zero.
	require.True(t, keeper.GetTripartyProcessedSequenceTip(ctx).IsZero())

	// Set and get.
	keeper.setTripartyProcessedSequenceTip(ctx, math.NewInt(5))
	require.Equal(t, math.NewInt(5), keeper.GetTripartyProcessedSequenceTip(ctx))

	// Update.
	keeper.setTripartyProcessedSequenceTip(ctx, math.NewInt(42))
	require.Equal(t, math.NewInt(42), keeper.GetTripartyProcessedSequenceTip(ctx))
}

func TestGetAllAllowedTripartyControllers(t *testing.T) {
	ctx, keeper := mockContext()

	t.Run("returns empty slice when no controllers exist", func(t *testing.T) {
		controllers := keeper.getAllAllowedTripartyControllers(ctx)
		require.Empty(t, controllers)
	})

	t.Run("returns all allowed controllers", func(t *testing.T) {
		controller1 := "0x1111111111111111111111111111111111111111"
		controller2 := "0x2222222222222222222222222222222222222222"
		controller3 := "0x3333333333333333333333333333333333333333"

		keeper.AllowTripartyController(ctx, evmtypes.HexAddressToBytes(controller1), true)
		keeper.AllowTripartyController(ctx, evmtypes.HexAddressToBytes(controller2), true)
		keeper.AllowTripartyController(ctx, evmtypes.HexAddressToBytes(controller3), true)

		controllers := keeper.getAllAllowedTripartyControllers(ctx)
		require.Len(t, controllers, 3)
		require.ElementsMatch(t, []string{controller1, controller2, controller3}, controllers)
	})
}

func TestGetAllPendingTripartyBridgeRequests(t *testing.T) {
	ctx, keeper := mockContext()

	t.Run("returns empty slice when no requests exist", func(t *testing.T) {
		requests := keeper.getAllPendingTripartyBridgeRequests(ctx)
		require.Empty(t, requests)
	})

	t.Run("returns all pending requests", func(t *testing.T) {
		keeper.AllowTripartyController(ctx, evmtypes.HexAddressToBytes(testTripartyController), true)
		keeper.SetTripartyWindowLimit(ctx, to18Dec(100))

		reqID1, err := keeper.CreateTripartyBridgeRequest(
			ctx, testTripartyRecipient, MinTripartyAmount, []byte("callback-1"), testTripartyController,
		)
		require.NoError(t, err)

		reqID2, err := keeper.CreateTripartyBridgeRequest(
			ctx, testTripartyRecipient, MinTripartyAmount.MulRaw(2), []byte("callback-2"), testTripartyController,
		)
		require.NoError(t, err)

		requests := keeper.getAllPendingTripartyBridgeRequests(ctx)
		require.Len(t, requests, 2)

		requestsBySequence := make(map[string]*types.TripartyBridgeRequest)
		for _, req := range requests {
			requestsBySequence[req.Sequence.String()] = req
		}

		require.Equal(t, testTripartyRecipient, requestsBySequence[reqID1.String()].Recipient)
		require.Equal(t, MinTripartyAmount, requestsBySequence[reqID1.String()].Amount)
		require.Equal(t, []byte("callback-1"), requestsBySequence[reqID1.String()].CallbackData)
		require.Equal(t, testTripartyController, requestsBySequence[reqID1.String()].Controller)

		require.Equal(t, testTripartyRecipient, requestsBySequence[reqID2.String()].Recipient)
		require.Equal(t, MinTripartyAmount.MulRaw(2), requestsBySequence[reqID2.String()].Amount)
		require.Equal(t, []byte("callback-2"), requestsBySequence[reqID2.String()].CallbackData)
		require.Equal(t, testTripartyController, requestsBySequence[reqID2.String()].Controller)
	})
}

func TestCreateTripartyBridgeRequestWindowLimit(t *testing.T) {
	ctx, keeper := mockContext()

	keeper.AllowTripartyController(ctx, evmtypes.HexAddressToBytes(testTripartyController), true)
	// Set triparty window limit to twice minimum amount.
	keeper.SetTripartyWindowLimit(ctx, MinTripartyAmount.MulRaw(2))

	// First successful request.
	reqID, err := keeper.CreateTripartyBridgeRequest(
		ctx, testTripartyRecipient, MinTripartyAmount, nil, testTripartyController,
	)
	require.NoError(t, err)
	require.Equal(t, math.NewInt(1), reqID)
	require.Equal(t, MinTripartyAmount, keeper.getTripartyWindowConsumed(ctx))
	capacity, _ := keeper.GetTripartyCapacity(ctx)
	require.Equal(t, MinTripartyAmount, capacity)

	// Unsuccessful request using remaining capacity plus one.
	_, err = keeper.CreateTripartyBridgeRequest(
		ctx, testTripartyRecipient, MinTripartyAmount.AddRaw(1), nil, testTripartyController,
	)
	require.ErrorIs(t, err, types.ErrTripartyWindowLimitExceeded)
	require.Equal(t, math.NewInt(1), keeper.GetTripartyRequestSequenceTip(ctx))
	require.Equal(t, MinTripartyAmount, keeper.getTripartyWindowConsumed(ctx))
	capacity, _ = keeper.GetTripartyCapacity(ctx)
	require.Equal(t, MinTripartyAmount, capacity)

	// Second successful request.
	reqID, err = keeper.CreateTripartyBridgeRequest(
		ctx, testTripartyRecipient, MinTripartyAmount, nil, testTripartyController,
	)
	require.NoError(t, err)
	require.Equal(t, math.NewInt(2), reqID)
	require.Equal(t, MinTripartyAmount.MulRaw(2), keeper.getTripartyWindowConsumed(ctx))
	capacity, _ = keeper.GetTripartyCapacity(ctx)
	require.True(t, capacity.IsZero())
}

// setupTripartyProcessing is a helper that creates context, keeper, and
// configured mocks for ProcessTripartyBridgeRequests tests. The bank
// and evm keeper mocks are returned so the caller can set expectations.
func setupTripartyProcessing(t *testing.T) (
	sdk.Context, Keeper, *mockBankKeeper, *mockEvmKeeper,
) {
	t.Helper()

	ctx, k := mockContext()

	bankKeeper := newMockBankKeeper()
	evmKeeper := newMockEvmKeeper()
	k.bankKeeper = bankKeeper
	k.evmKeeper = evmKeeper

	// By default, all addresses are not precompiles.
	evmKeeper.On(
		"IsCustomPrecompileAddress", mock.Anything,
	).Return(false)

	// Allow the test controller.
	k.AllowTripartyController(
		ctx,
		evmtypes.HexAddressToBytes(testTripartyController),
		true,
	)
	k.SetTripartyWindowLimit(ctx, to18Dec(100))

	return ctx, k, bankKeeper, evmKeeper
}

// to18Dec converts a whole-BTC amount into 18-decimal units.
func to18Dec(amount int64) math.Int {
	return math.NewInt(amount).MulRaw(1_000_000_000_000_000_000)
}

// createTripartyRequest is a helper to create a request at a given block
// height.
func createTripartyRequest(
	t *testing.T,
	ctx sdk.Context,
	k Keeper,
	height int64,
	amount math.Int,
	callbackData []byte,
) {
	t.Helper()

	createTripartyRequestFromController(
		t, ctx, k, height, amount, callbackData, testTripartyController,
	)
}

// createTripartyRequestFromController is a helper to create a request from a
// specific controller at a given block height.
func createTripartyRequestFromController(
	t *testing.T,
	ctx sdk.Context,
	k Keeper,
	height int64,
	amount math.Int,
	callbackData []byte,
	controller string,
) {
	t.Helper()

	reqCtx := ctx.WithBlockHeader(tmproto.Header{Height: height})

	_, err := k.CreateTripartyBridgeRequest(
		reqCtx,
		testTripartyRecipient,
		amount,
		callbackData,
		controller,
	)
	require.NoError(t, err)
}

// expectMintBTC sets up mock expectations for a successful mintBTC call.
func expectMintBTC(
	bk *mockBankKeeper,
	ctx sdk.Context,
	recipient sdk.AccAddress,
	amount math.Int,
) {
	coins := sdk.NewCoins(sdk.NewCoin(evmtypes.DefaultEVMDenom, amount))

	bk.On("MintCoins", ctx, types.ModuleName, coins).Return(nil)
	bk.On(
		"SendCoinsFromModuleToAccount",
		ctx, types.ModuleName, recipient, coins,
	).Return(nil)
}

func TestProcessTripartyBridgeRequests_NoPendingRequests(t *testing.T) {
	ctx, k, bk, ek := setupTripartyProcessing(t)

	// Set block height past the delay.
	ctx = ctx.WithBlockHeader(tmproto.Header{Height: 100})

	err := k.ProcessTripartyBridgeRequests(ctx)
	require.NoError(t, err)

	// No interactions with bank or evm keepers.
	bk.AssertNotCalled(t, "MintCoins")
	ek.AssertNotCalled(t, "ExecuteContractCall")

	// Processed tip should not have advanced.
	require.True(t, k.GetTripartyProcessedSequenceTip(ctx).IsZero())
}

func TestProcessTripartyBridgeRequests_Paused(t *testing.T) {
	ctx, k, bk, ek := setupTripartyProcessing(t)

	// Create a mature request.
	createTripartyRequest(t, ctx, k, 1, to18Dec(1), nil)

	// Pause triparty.
	k.SetTripartyPaused(ctx, true)

	ctx = ctx.WithBlockHeader(tmproto.Header{Height: 100})

	err := k.ProcessTripartyBridgeRequests(ctx)
	require.NoError(t, err)

	// Nothing should have been processed.
	bk.AssertNotCalled(t, "MintCoins")
	ek.AssertNotCalled(t, "ExecuteContractCall")

	// Request should still exist.
	_, found := k.getTripartyBridgeRequest(ctx, math.NewInt(1))
	require.True(t, found)
}

func TestProcessTripartyBridgeRequests_AllImmature(t *testing.T) {
	ctx, k, bk, ek := setupTripartyProcessing(t)

	// Set block delay to 10.
	require.NoError(t, k.SetTripartyBlockDelay(ctx, 10))

	// Create requests at height 90.
	createTripartyRequest(t, ctx, k, 90, to18Dec(1), nil)
	createTripartyRequest(t, ctx, k, 90, to18Dec(2), nil)

	// Current height 95 — only 5 blocks have passed, need 10.
	ctx = ctx.WithBlockHeader(tmproto.Header{Height: 95})

	err := k.ProcessTripartyBridgeRequests(ctx)
	require.NoError(t, err)

	// Nothing should have been processed.
	bk.AssertNotCalled(t, "MintCoins")
	ek.AssertNotCalled(t, "ExecuteContractCall")

	// Both requests should still exist.
	_, found := k.getTripartyBridgeRequest(ctx, math.NewInt(1))
	require.True(t, found)
	_, found = k.getTripartyBridgeRequest(ctx, math.NewInt(2))
	require.True(t, found)

	// Processed tip should not have advanced.
	require.True(t, k.GetTripartyProcessedSequenceTip(ctx).IsZero())
}

func TestProcessTripartyBridgeRequests_MixedMaturity(t *testing.T) {
	ctx, k, bk, ek := setupTripartyProcessing(t)

	// Default delay is 1. Create 3 requests at different heights.
	createTripartyRequest(t, ctx, k, 10, to18Dec(1), nil) // mature at 11+
	createTripartyRequest(t, ctx, k, 10, to18Dec(2), nil) // mature at 11+
	createTripartyRequest(t, ctx, k, 20, to18Dec(3), nil) // mature at 21+

	// Process at height 20 — first two are mature, third is not (20-20=0 < 1).
	ctx = ctx.WithBlockHeader(tmproto.Header{Height: 20})

	// Expect mints for requests 1 and 2.
	expectMintBTC(bk, ctx, testTripartyRecipientAddr, to18Dec(1))
	expectMintBTC(bk, ctx, testTripartyRecipientAddr, to18Dec(2))

	// Expect callbacks (use mock.Anything for the call).
	ek.On("ExecuteContractCall", ctx, mock.Anything).Return(
		&evmtypes.MsgEthereumTxResponse{}, nil,
	)

	err := k.ProcessTripartyBridgeRequests(ctx)
	require.NoError(t, err)

	// First two deleted, third remains.
	_, found := k.getTripartyBridgeRequest(ctx, math.NewInt(1))
	require.False(t, found)
	_, found = k.getTripartyBridgeRequest(ctx, math.NewInt(2))
	require.False(t, found)
	_, found = k.getTripartyBridgeRequest(ctx, math.NewInt(3))
	require.True(t, found)

	// Processed tip advanced to 2 (not 3 because we stopped there).
	require.Equal(t, math.NewInt(2), k.GetTripartyProcessedSequenceTip(ctx))

	// Provenance counter updated for both.
	require.Equal(t, to18Dec(3), k.GetTripartyTotalBTCMinted(ctx))
}

func TestProcessTripartyBridgeRequests_BlockedRecipient(t *testing.T) {
	// Set bech32 prefixes so AccAddressFromBech32 works with "mezo1" prefix.
	cfg := sdk.GetConfig()
	config.SetBech32Prefixes(cfg)

	ctx, k, bankKeeper, evmKeeper := setupTripartyProcessing(t)

	// Compute the hex address that maps to the blocked bech32 address.
	blockedAddr, err := sdk.AccAddressFromBech32(testBlockedAddress)
	require.NoError(t, err)
	blockedHexAddr := evmtypes.BytesToHexAddress(blockedAddr.Bytes())

	// Write the blocked-recipient request directly to state (bypassing
	// CreateTripartyBridgeRequest which now rejects blocked addresses
	// at creation time). This simulates a request created before the
	// address was blocked.
	blockedReq := &types.TripartyBridgeRequest{
		Sequence:    k.incrementTripartyRequestSequenceTip(ctx),
		BlockHeight: 10,
		Recipient:   blockedHexAddr,
		Amount:      to18Dec(1),
		Controller:  testTripartyController,
	}
	bz, err := blockedReq.Marshal()
	require.NoError(t, err)
	ctx.KVStore(k.storeKey).Set(
		types.GetTripartyBridgeRequestKey(blockedReq.Sequence), bz,
	)

	// Create a second valid request.
	createTripartyRequest(t, ctx, k, 10, to18Dec(2), nil)

	// Process at height 20.
	ctx = ctx.WithBlockHeader(tmproto.Header{Height: 20})

	// Only the second request should be minted.
	expectMintBTC(bankKeeper, ctx, testTripartyRecipientAddr, to18Dec(2))
	evmKeeper.On("ExecuteContractCall", ctx, mock.Anything).Return(
		&evmtypes.MsgEthereumTxResponse{}, nil,
	)

	err = k.ProcessTripartyBridgeRequests(ctx)
	require.NoError(t, err)

	// Both deleted.
	_, found := k.getTripartyBridgeRequest(ctx, math.NewInt(1))
	require.False(t, found)
	_, found = k.getTripartyBridgeRequest(ctx, math.NewInt(2))
	require.False(t, found)

	// Only the valid request contributed to provenance.
	require.Equal(t, to18Dec(2), k.GetTripartyTotalBTCMinted(ctx))

	// Processed tip advanced to 2.
	require.Equal(t, math.NewInt(2), k.GetTripartyProcessedSequenceTip(ctx))
}

func TestProcessTripartyBridgeRequests_PrecompileRecipient(t *testing.T) {
	ctx, k, bankKeeper, evmKeeper := setupTripartyProcessing(t)

	// Create a request first (before the recipient becomes a precompile).
	createTripartyRequest(t, ctx, k, 10, to18Dec(1), nil)

	// Now override the mock so the recipient is considered a precompile
	// at processing time (simulating a precompile registered after request
	// creation).
	evmKeeper.ExpectedCalls = nil // Clear defaults.
	evmKeeper.On("IsCustomPrecompileAddress", testTripartyRecipient).Return(true)
	evmKeeper.On(
		"IsCustomPrecompileAddress", mock.MatchedBy(func(addr string) bool {
			return addr != testTripartyRecipient
		}),
	).Return(false)

	ctx = ctx.WithBlockHeader(tmproto.Header{Height: 20})

	err := k.ProcessTripartyBridgeRequests(ctx)
	require.NoError(t, err)

	// No mint should have happened.
	bankKeeper.AssertNotCalled(t, "MintCoins")

	// Request deleted.
	_, found := k.getTripartyBridgeRequest(ctx, math.NewInt(1))
	require.False(t, found)

	// Processed tip advanced.
	require.Equal(t, math.NewInt(1), k.GetTripartyProcessedSequenceTip(ctx))
}

func TestProcessTripartyBridgeRequests_DeauthorizedController(t *testing.T) {
	ctx, k, bk, ek := setupTripartyProcessing(t)

	// Create a request, then deauthorize the controller.
	createTripartyRequest(t, ctx, k, 10, to18Dec(1), nil)

	k.AllowTripartyController(
		ctx,
		evmtypes.HexAddressToBytes(testTripartyController),
		false,
	)

	ctx = ctx.WithBlockHeader(tmproto.Header{Height: 20})

	err := k.ProcessTripartyBridgeRequests(ctx)
	require.NoError(t, err)

	// No mint should have happened.
	bk.AssertNotCalled(t, "MintCoins")
	ek.AssertNotCalled(t, "ExecuteContractCall")

	// Request deleted.
	_, found := k.getTripartyBridgeRequest(ctx, math.NewInt(1))
	require.False(t, found)

	// Processed tip advanced.
	require.Equal(t, math.NewInt(1), k.GetTripartyProcessedSequenceTip(ctx))
}

func TestProcessTripartyBridgeRequests_PerRequestLimitExceeded(t *testing.T) {
	ctx, k, bk, ek := setupTripartyProcessing(t)

	// Create request, then lower the limit below its amount.
	createTripartyRequest(t, ctx, k, 10, to18Dec(1), nil)
	k.SetTripartyPerRequestLimit(ctx, to18Dec(1).SubRaw(1))

	ctx = ctx.WithBlockHeader(tmproto.Header{Height: 20})

	err := k.ProcessTripartyBridgeRequests(ctx)
	require.NoError(t, err)

	// No mint.
	bk.AssertNotCalled(t, "MintCoins")
	ek.AssertNotCalled(t, "ExecuteContractCall")

	// Request deleted.
	_, found := k.getTripartyBridgeRequest(ctx, math.NewInt(1))
	require.False(t, found)
}

func TestProcessTripartyBridgeRequests_SuccessfulMintAndCallback(t *testing.T) {
	ctx, k, bk, ek := setupTripartyProcessing(t)

	callbackData := []byte("test-callback-data")
	createTripartyRequest(t, ctx, k, 10, to18Dec(5), callbackData)

	ctx = ctx.WithBlockHeader(tmproto.Header{Height: 20})

	expectMintBTC(bk, ctx, testTripartyRecipientAddr, to18Dec(5))
	ek.On("ExecuteContractCall", ctx, mock.Anything).Return(
		&evmtypes.MsgEthereumTxResponse{}, nil,
	)

	err := k.ProcessTripartyBridgeRequests(ctx)
	require.NoError(t, err)

	// Request deleted.
	_, found := k.getTripartyBridgeRequest(ctx, math.NewInt(1))
	require.False(t, found)

	// Provenance counters updated.
	require.Equal(t, to18Dec(5), k.GetTripartyTotalBTCMinted(ctx))
	require.Equal(t, to18Dec(5), k.GetTripartyControllerBTCMinted(
		ctx, testTripartyController,
	))

	// BTCMinted counter updated (via mintBTC).
	require.Equal(t, to18Dec(5), k.GetBTCMinted(ctx))

	// Processed tip advanced.
	require.Equal(t, math.NewInt(1), k.GetTripartyProcessedSequenceTip(ctx))

	// Callback was issued.
	ek.AssertCalled(t, "ExecuteContractCall", ctx, mock.Anything)
}

func TestProcessTripartyBridgeRequests_CallbackFailure(t *testing.T) {
	ctx, k, bk, ek := setupTripartyProcessing(t)

	createTripartyRequest(t, ctx, k, 10, to18Dec(1), nil)

	ctx = ctx.WithBlockHeader(tmproto.Header{Height: 20})

	expectMintBTC(bk, ctx, testTripartyRecipientAddr, to18Dec(1))

	// Callback fails.
	ek.On("ExecuteContractCall", ctx, mock.Anything).Return(
		nil, fmt.Errorf("callback reverted"),
	)

	err := k.ProcessTripartyBridgeRequests(ctx)
	require.NoError(t, err) // Callback failure is non-fatal.

	// Mint still completed.
	require.Equal(t, to18Dec(1), k.GetBTCMinted(ctx))
	require.Equal(t, to18Dec(1), k.GetTripartyTotalBTCMinted(ctx))

	// Request deleted despite callback failure.
	_, found := k.getTripartyBridgeRequest(ctx, math.NewInt(1))
	require.False(t, found)
}

func TestProcessTripartyBridgeRequests_MintBTCFailure(t *testing.T) {
	ctx, k, bk, ek := setupTripartyProcessing(t)

	createTripartyRequest(t, ctx, k, 10, to18Dec(1), nil)

	ctx = ctx.WithBlockHeader(tmproto.Header{Height: 20})

	// Bank keeper fails to mint.
	coins := sdk.NewCoins(sdk.NewCoin(evmtypes.DefaultEVMDenom, to18Dec(1)))
	bk.On("MintCoins", ctx, types.ModuleName, coins).Return(
		fmt.Errorf("bank error"),
	)

	err := k.ProcessTripartyBridgeRequests(ctx)
	require.Error(t, err) // System error is fatal.
	require.ErrorContains(t, err, "failed to mint BTC")

	// No callback should have been issued.
	ek.AssertNotCalled(t, "ExecuteContractCall")
}

func TestProcessTripartyBridgeRequests_EmptyCallbackData(t *testing.T) {
	ctx, k, bk, ek := setupTripartyProcessing(t)

	// Create request with nil callback data.
	createTripartyRequest(t, ctx, k, 10, to18Dec(1), nil)

	ctx = ctx.WithBlockHeader(tmproto.Header{Height: 20})

	expectMintBTC(bk, ctx, testTripartyRecipientAddr, to18Dec(1))

	// Callback should still be issued with empty bytes.
	ek.On("ExecuteContractCall", ctx, mock.Anything).Return(
		&evmtypes.MsgEthereumTxResponse{}, nil,
	)

	err := k.ProcessTripartyBridgeRequests(ctx)
	require.NoError(t, err)

	// Callback was issued.
	ek.AssertCalled(t, "ExecuteContractCall", ctx, mock.Anything)
}

func TestProcessTripartyBridgeRequests_BatchCap(t *testing.T) {
	ctx, k, bk, ek := setupTripartyProcessing(t)

	// Create 6 mature requests — only 5 should be processed (TripartyBatch).
	for i := 0; i < 6; i++ {
		createTripartyRequest(t, ctx, k, 10, to18Dec(int64(i+1)), nil)
	}

	ctx = ctx.WithBlockHeader(tmproto.Header{Height: 20})

	// Expect mints for requests 1-5.
	for i := 0; i < 5; i++ {
		expectMintBTC(bk, ctx, testTripartyRecipientAddr, to18Dec(int64(i+1)))
	}
	ek.On("ExecuteContractCall", ctx, mock.Anything).Return(
		&evmtypes.MsgEthereumTxResponse{}, nil,
	)

	err := k.ProcessTripartyBridgeRequests(ctx)
	require.NoError(t, err)

	// Requests 1-5 deleted.
	for i := 1; i <= 5; i++ {
		_, found := k.getTripartyBridgeRequest(ctx, math.NewInt(int64(i)))
		require.False(t, found, "request %d should be deleted", i)
	}

	// Request 6 still exists.
	_, found := k.getTripartyBridgeRequest(ctx, math.NewInt(6))
	require.True(t, found, "request 6 should still exist")

	// Processed tip advanced to 5.
	require.Equal(t, math.NewInt(5), k.GetTripartyProcessedSequenceTip(ctx))

	// Provenance counters: 100+200+300+400+500 = 1500.
	require.Equal(t, to18Dec(15), k.GetTripartyTotalBTCMinted(ctx))
	require.Equal(t, to18Dec(15), k.GetTripartyControllerBTCMinted(
		ctx, testTripartyController,
	))
}

func TestProcessTripartyBridgeRequests_ResumesFromProcessedTip(t *testing.T) {
	ctx, k, bk, ek := setupTripartyProcessing(t)

	// Create 3 requests.
	createTripartyRequest(t, ctx, k, 10, to18Dec(1), nil)
	createTripartyRequest(t, ctx, k, 10, to18Dec(2), nil)
	createTripartyRequest(t, ctx, k, 10, to18Dec(3), nil)

	ctx = ctx.WithBlockHeader(tmproto.Header{Height: 20})

	// Process first batch — all 3.
	expectMintBTC(bk, ctx, testTripartyRecipientAddr, to18Dec(1))
	expectMintBTC(bk, ctx, testTripartyRecipientAddr, to18Dec(2))
	expectMintBTC(bk, ctx, testTripartyRecipientAddr, to18Dec(3))
	ek.On("ExecuteContractCall", ctx, mock.Anything).Return(
		&evmtypes.MsgEthereumTxResponse{}, nil,
	)

	err := k.ProcessTripartyBridgeRequests(ctx)
	require.NoError(t, err)
	require.Equal(t, math.NewInt(3), k.GetTripartyProcessedSequenceTip(ctx))

	// Create more requests.
	createTripartyRequest(t, ctx, k, 15, to18Dec(4), nil) // seq 4
	createTripartyRequest(t, ctx, k, 15, to18Dec(5), nil) // seq 5

	// Second processing run with fresh mocks.
	bankKeeper2 := newMockBankKeeper()
	evmKeeper2 := newMockEvmKeeper()
	evmKeeper2.On("IsCustomPrecompileAddress", mock.Anything).Return(false)
	k.bankKeeper = bankKeeper2
	k.evmKeeper = evmKeeper2

	expectMintBTC(bankKeeper2, ctx, testTripartyRecipientAddr, to18Dec(4))
	expectMintBTC(bankKeeper2, ctx, testTripartyRecipientAddr, to18Dec(5))
	evmKeeper2.On("ExecuteContractCall", ctx, mock.Anything).Return(
		&evmtypes.MsgEthereumTxResponse{}, nil,
	)

	err = k.ProcessTripartyBridgeRequests(ctx)
	require.NoError(t, err)

	// Processed tip advanced to 5.
	require.Equal(t, math.NewInt(5), k.GetTripartyProcessedSequenceTip(ctx))

	// Total provenance: 1+2+3+4+5 BTC.
	require.Equal(t, to18Dec(15), k.GetTripartyTotalBTCMinted(ctx))
}

func TestProcessTripartyBridgeRequests_MultipleControllers(t *testing.T) {
	ctx, k, bk, ek := setupTripartyProcessing(t)

	// Allow a second controller.
	controller2 := "0x0303030303030303030303030303030303030303"
	k.AllowTripartyController(
		ctx,
		evmtypes.HexAddressToBytes(controller2),
		true,
	)

	// Create requests from two different controllers.
	// Controller 1: 3 BTC and 7 BTC.
	createTripartyRequest(t, ctx, k, 10, to18Dec(3), nil)
	createTripartyRequestFromController(t, ctx, k, 10, to18Dec(5), nil, controller2)
	createTripartyRequest(t, ctx, k, 10, to18Dec(7), nil)
	createTripartyRequestFromController(t, ctx, k, 10, to18Dec(2), nil, controller2)

	ctx = ctx.WithBlockHeader(tmproto.Header{Height: 20})

	expectMintBTC(bk, ctx, testTripartyRecipientAddr, to18Dec(3))
	expectMintBTC(bk, ctx, testTripartyRecipientAddr, to18Dec(5))
	expectMintBTC(bk, ctx, testTripartyRecipientAddr, to18Dec(7))
	expectMintBTC(bk, ctx, testTripartyRecipientAddr, to18Dec(2))
	ek.On("ExecuteContractCall", ctx, mock.Anything).Return(
		&evmtypes.MsgEthereumTxResponse{}, nil,
	)

	err := k.ProcessTripartyBridgeRequests(ctx)
	require.NoError(t, err)

	// Global provenance: 3+5+7+2 = 17 BTC.
	require.Equal(t, to18Dec(17), k.GetTripartyTotalBTCMinted(ctx))

	// Per-controller provenance: controller1 = 3+7 = 10, controller2 = 5+2 = 7.
	require.Equal(t, to18Dec(10), k.GetTripartyControllerBTCMinted(
		ctx, testTripartyController,
	))
	require.Equal(t, to18Dec(7), k.GetTripartyControllerBTCMinted(
		ctx, controller2,
	))
}
