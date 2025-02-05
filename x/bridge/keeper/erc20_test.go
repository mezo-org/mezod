package keeper

import (
	"testing"

	evmtypes "github.com/mezo-org/mezod/x/evm/types"

	"github.com/mezo-org/mezod/x/bridge/types"
	"github.com/stretchr/testify/require"
)

func TestGetERC20TokensMappings(t *testing.T) {
	ctx, k := mockContext()

	mapping := types.NewERC20TokenMapping(
		evmtypes.HexAddressToBytes(testSourceERC20Token1),
		evmtypes.HexAddressToBytes(testMezoERC20Token1),
	)
	k.setERC20TokenMapping(ctx, mapping)

	actualMappings := k.GetERC20TokensMappings(ctx)
	require.Len(t, actualMappings, 1)
	require.Equal(t, mapping, actualMappings[0])
}

func TestGetERC20TokenMapping(t *testing.T) {
	ctx, k := mockContext()

	actualMapping, found := k.GetERC20TokenMapping(ctx, evmtypes.HexAddressToBytes(testSourceERC20Token1))
	require.False(t, found)
	require.Nil(t, actualMapping)

	mapping := types.NewERC20TokenMapping(
		evmtypes.HexAddressToBytes(testSourceERC20Token1),
		evmtypes.HexAddressToBytes(testMezoERC20Token1),
	)
	k.setERC20TokenMapping(ctx, mapping)

	actualMapping, found = k.GetERC20TokenMapping(ctx, evmtypes.HexAddressToBytes(testSourceERC20Token1))
	require.True(t, found)
	require.Equal(t, mapping, actualMapping)
}

func TestCreateERC20TokenMapping(t *testing.T) {
	ctx, k := mockContext()
	sourceToken := evmtypes.HexAddressToBytes(testSourceERC20Token1)
	mezoToken := evmtypes.HexAddressToBytes(testMezoERC20Token1)
	zeroToken := evmtypes.HexAddressToBytes("0x0000000000000000000000000000000000000000")

	// Set max mappings to 1.
	err := k.SetParams(ctx, types.Params{MaxErc20TokensMappings: 1})
	require.NoError(t, err)

	// Test zero source token.
	err = k.CreateERC20TokenMapping(ctx, zeroToken, mezoToken)
	require.ErrorContains(t, err, types.ErrZeroEVMAddress.Error())

	// Test zero mezo token.
	err = k.CreateERC20TokenMapping(ctx, sourceToken, zeroToken)
	require.ErrorContains(t, err, types.ErrZeroEVMAddress.Error())

	// Test mezo token is not a contract.
	k.evmKeeper.(*mockEvmKeeper).On("IsContract", ctx, mezoToken).Return(false)
	err = k.CreateERC20TokenMapping(ctx, sourceToken, mezoToken)
	require.ErrorContains(t, err, types.ErrTokenNotContract.Error())

	// Test valid mapping creation.
	k.evmKeeper = newMockEvmKeeper() // reset the mock
	k.evmKeeper.(*mockEvmKeeper).On("IsContract", ctx, mezoToken).Return(true)
	err = k.CreateERC20TokenMapping(ctx, sourceToken, mezoToken)
	require.NoError(t, err)

	// Test duplicate mapping.
	err = k.CreateERC20TokenMapping(ctx, sourceToken, mezoToken)
	require.ErrorContains(t, err, types.ErrAlreadyMapping.Error())

	// Test max mappings reached.
	k.evmKeeper.(*mockEvmKeeper).On(
		"IsContract",
		ctx,
		evmtypes.HexAddressToBytes(testMezoERC20Token2),
	).Return(true)
	err = k.CreateERC20TokenMapping(
		ctx,
		evmtypes.HexAddressToBytes(testSourceERC20Token2),
		evmtypes.HexAddressToBytes(testMezoERC20Token2),
	)
	require.ErrorContains(t, err, types.ErrMaxMappingsReached.Error())
}

func TestDeleteERC20TokenMapping(t *testing.T) {
	ctx, k := mockContext()
	sourceToken := evmtypes.HexAddressToBytes(testSourceERC20Token1)
	mezoToken := evmtypes.HexAddressToBytes(testMezoERC20Token1)

	// Test non-existing mapping
	err := k.DeleteERC20TokenMapping(ctx, sourceToken)
	require.ErrorContains(t, err, types.ErrNotMapping.Error())

	// Add a mapping and test deletion
	mapping := types.NewERC20TokenMapping(sourceToken, mezoToken)
	k.setERC20TokenMapping(ctx, mapping)
	err = k.DeleteERC20TokenMapping(ctx, sourceToken)
	require.NoError(t, err)

	// Test mapping is deleted
	_, found := k.GetERC20TokenMapping(ctx, sourceToken)
	require.False(t, found)
}
