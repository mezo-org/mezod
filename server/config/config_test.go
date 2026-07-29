package config

import (
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	require.True(t, cfg.JSONRPC.Enable)
	require.Equal(t, cfg.JSONRPC.Address, DefaultJSONRPCAddress)
	require.Equal(t, cfg.JSONRPC.WsAddress, DefaultJSONRPCWsAddress)
}

func TestDefaultConfigLimitsGetProofStorageKeys(t *testing.T) {
	// A zero cap means no limit, so the default must be a real number to keep
	// the work of a single `eth_getProof` call bounded out of the box.
	require.Positive(t, DefaultGetProofStorageKeysCap)
	require.Equal(t, DefaultGetProofStorageKeysCap, DefaultConfig().JSONRPC.GetProofStorageKeysCap)
}

func TestJSONRPCConfigValidate(t *testing.T) {
	t.Run("accepts the defaults", func(t *testing.T) {
		require.NoError(t, DefaultJSONRPCConfig().Validate())
	})

	t.Run("accepts a get proof storage keys cap of zero", func(t *testing.T) {
		cfg := DefaultJSONRPCConfig()
		cfg.GetProofStorageKeysCap = 0
		require.NoError(t, cfg.Validate())
	})

	t.Run("rejects a negative get proof storage keys cap", func(t *testing.T) {
		cfg := DefaultJSONRPCConfig()
		cfg.GetProofStorageKeysCap = -1
		require.EqualError(t, cfg.Validate(), "JSON-RPC get proof storage keys cap cannot be negative")
	})
}

// TestGetConfigReadsGetProofStorageKeysCap pins the name the setting is read
// under. Without it, reading the value under a name that does not match the
// `app.toml` key and the CLI flag would leave the setting with no effect.
func TestGetConfigReadsGetProofStorageKeysCap(t *testing.T) {
	v := viper.New()
	v.Set("json-rpc.get-proof-storage-keys-cap", 7)

	cfg, err := GetConfig(v)
	require.NoError(t, err)
	require.Equal(t, int32(7), cfg.JSONRPC.GetProofStorageKeysCap)
}
