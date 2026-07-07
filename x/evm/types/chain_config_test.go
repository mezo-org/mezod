package types

import (
	"math/big"
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/stretchr/testify/require"

	"github.com/ethereum/go-ethereum/common"
)

var defaultEIP150Hash = common.Hash{}.String()

func newIntPtr(i int64) *sdkmath.Int {
	v := sdkmath.NewInt(i)
	return &v
}

// validBaselineConfig returns a ChainConfig that passes Validate(). Time
// forks past Osaka are left nil because params.DefaultBlobSchedule has no
// entries for them yet and CheckConfigForkOrder rejects configs that
// activate a time fork without a matching blob schedule entry.
func validBaselineConfig() ChainConfig {
	return ChainConfig{
		HomesteadBlock:      newIntPtr(0),
		DAOForkBlock:        newIntPtr(0),
		EIP150Block:         newIntPtr(0),
		EIP150Hash:          defaultEIP150Hash,
		EIP155Block:         newIntPtr(0),
		EIP158Block:         newIntPtr(0),
		ByzantiumBlock:      newIntPtr(0),
		ConstantinopleBlock: newIntPtr(0),
		PetersburgBlock:     newIntPtr(0),
		IstanbulBlock:       newIntPtr(0),
		MuirGlacierBlock:    newIntPtr(0),
		BerlinBlock:         newIntPtr(0),
		LondonBlock:         newIntPtr(0),
		ArrowGlacierBlock:   newIntPtr(0),
		GrayGlacierBlock:    newIntPtr(0),
		MergeNetsplitBlock:  newIntPtr(0),
		ShanghaiTime:        newIntPtr(0),
		CancunTime:          newIntPtr(0),
		PragueTime:          newIntPtr(0),
		OsakaTime:           newIntPtr(0),
	}
}

// mutate returns a copy of base with mutator applied.
func mutate(base ChainConfig, mutator func(*ChainConfig)) ChainConfig {
	cfg := base
	mutator(&cfg)
	return cfg
}

func TestChainConfigValidate(t *testing.T) {
	baseline := validBaselineConfig()

	testCases := []struct {
		name     string
		config   ChainConfig
		expError bool
	}{
		{"default", DefaultChainConfig(), false},
		{"valid baseline", baseline, false},
		{
			"valid with nil values",
			ChainConfig{
				EIP150Hash: defaultEIP150Hash,
			},
			false,
		},
		{"empty", ChainConfig{}, false},
		{
			"invalid HomesteadBlock",
			mutate(baseline, func(c *ChainConfig) { c.HomesteadBlock = newIntPtr(-1) }),
			true,
		},
		{
			"invalid DAOForkBlock",
			mutate(baseline, func(c *ChainConfig) { c.DAOForkBlock = newIntPtr(-1) }),
			true,
		},
		{
			"invalid EIP150Block",
			mutate(baseline, func(c *ChainConfig) { c.EIP150Block = newIntPtr(-1) }),
			true,
		},
		{
			"invalid EIP150Hash",
			mutate(baseline, func(c *ChainConfig) { c.EIP150Hash = "  " }),
			true,
		},
		{
			"invalid EIP155Block",
			mutate(baseline, func(c *ChainConfig) { c.EIP155Block = newIntPtr(-1) }),
			true,
		},
		{
			"invalid EIP158Block",
			mutate(baseline, func(c *ChainConfig) { c.EIP158Block = newIntPtr(-1) }),
			true,
		},
		{
			"invalid ByzantiumBlock",
			mutate(baseline, func(c *ChainConfig) { c.ByzantiumBlock = newIntPtr(-1) }),
			true,
		},
		{
			"invalid ConstantinopleBlock",
			mutate(baseline, func(c *ChainConfig) { c.ConstantinopleBlock = newIntPtr(-1) }),
			true,
		},
		{
			"invalid PetersburgBlock",
			mutate(baseline, func(c *ChainConfig) { c.PetersburgBlock = newIntPtr(-1) }),
			true,
		},
		{
			"invalid IstanbulBlock",
			mutate(baseline, func(c *ChainConfig) { c.IstanbulBlock = newIntPtr(-1) }),
			true,
		},
		{
			"invalid MuirGlacierBlock",
			mutate(baseline, func(c *ChainConfig) { c.MuirGlacierBlock = newIntPtr(-1) }),
			true,
		},
		{
			"invalid BerlinBlock",
			mutate(baseline, func(c *ChainConfig) { c.BerlinBlock = newIntPtr(-1) }),
			true,
		},
		{
			"invalid LondonBlock",
			mutate(baseline, func(c *ChainConfig) { c.LondonBlock = newIntPtr(-1) }),
			true,
		},
		{
			"invalid ArrowGlacierBlock",
			mutate(baseline, func(c *ChainConfig) { c.ArrowGlacierBlock = newIntPtr(-1) }),
			true,
		},
		{
			"invalid GrayGlacierBlock",
			mutate(baseline, func(c *ChainConfig) { c.GrayGlacierBlock = newIntPtr(-1) }),
			true,
		},
		{
			"invalid MergeNetsplitBlock",
			mutate(baseline, func(c *ChainConfig) { c.MergeNetsplitBlock = newIntPtr(-1) }),
			true,
		},
		{
			"invalid fork order - skip HomesteadBlock",
			mutate(baseline, func(c *ChainConfig) { c.HomesteadBlock = nil }),
			true,
		},
		{
			"invalid ShanghaiTime",
			mutate(baseline, func(c *ChainConfig) { c.ShanghaiTime = newIntPtr(-1) }),
			true,
		},
		{
			"invalid CancunTime",
			mutate(baseline, func(c *ChainConfig) { c.CancunTime = newIntPtr(-1) }),
			true,
		},
		{
			"invalid PragueTime",
			mutate(baseline, func(c *ChainConfig) { c.PragueTime = newIntPtr(-1) }),
			true,
		},
		{
			"invalid OsakaTime",
			mutate(baseline, func(c *ChainConfig) { c.OsakaTime = newIntPtr(-1) }),
			true,
		},
		{
			"invalid BPO1Time",
			mutate(baseline, func(c *ChainConfig) { c.BPO1Time = newIntPtr(-1) }),
			true,
		},
		{
			"invalid BPO2Time",
			mutate(baseline, func(c *ChainConfig) { c.BPO2Time = newIntPtr(-1) }),
			true,
		},
		{
			"invalid BPO3Time",
			mutate(baseline, func(c *ChainConfig) { c.BPO3Time = newIntPtr(-1) }),
			true,
		},
		{
			"invalid BPO4Time",
			mutate(baseline, func(c *ChainConfig) { c.BPO4Time = newIntPtr(-1) }),
			true,
		},
		{
			"invalid BPO5Time",
			mutate(baseline, func(c *ChainConfig) { c.BPO5Time = newIntPtr(-1) }),
			true,
		},
		{
			"invalid AmsterdamTime",
			mutate(baseline, func(c *ChainConfig) { c.AmsterdamTime = newIntPtr(-1) }),
			true,
		},
		{
			"invalid VerkleTime",
			mutate(baseline, func(c *ChainConfig) { c.VerkleTime = newIntPtr(-1) }),
			true,
		},
	}

	for _, tc := range testCases {
		err := tc.config.Validate()

		if tc.expError {
			require.Error(t, err, tc.name)
		} else {
			require.NoError(t, err, tc.name)
		}
	}
}

// TestChainConfigEthereumConfig asserts that ChainConfig.EthereumConfig
// maps each fork-time field to the corresponding field on the returned
// *params.ChainConfig. A cross-wired assignment (e.g. OsakaTime going
// to params.PragueTime) would silently activate the wrong fork on
// every fresh genesis; populating each field with a distinct sentinel
// makes any swap immediately visible.
func TestChainConfigEthereumConfig(t *testing.T) {
	// Each fork time gets a distinct sentinel so a cross-wire
	// (e.g. assigning OsakaTime to params.PragueTime) is detectable.
	const (
		shanghaiSentinel  uint64 = 1001
		cancunSentinel    uint64 = 1002
		pragueSentinel    uint64 = 1003
		osakaSentinel     uint64 = 1004
		bpo1Sentinel      uint64 = 1005
		bpo2Sentinel      uint64 = 1006
		bpo3Sentinel      uint64 = 1007
		bpo4Sentinel      uint64 = 1008
		bpo5Sentinel      uint64 = 1009
		amsterdamSentinel uint64 = 1010
		verkleSentinel    uint64 = 1011
	)

	cfg := ChainConfig{
		ShanghaiTime:  newIntPtr(int64(shanghaiSentinel)),
		CancunTime:    newIntPtr(int64(cancunSentinel)),
		PragueTime:    newIntPtr(int64(pragueSentinel)),
		OsakaTime:     newIntPtr(int64(osakaSentinel)),
		BPO1Time:      newIntPtr(int64(bpo1Sentinel)),
		BPO2Time:      newIntPtr(int64(bpo2Sentinel)),
		BPO3Time:      newIntPtr(int64(bpo3Sentinel)),
		BPO4Time:      newIntPtr(int64(bpo4Sentinel)),
		BPO5Time:      newIntPtr(int64(bpo5Sentinel)),
		AmsterdamTime: newIntPtr(int64(amsterdamSentinel)),
		VerkleTime:    newIntPtr(int64(verkleSentinel)),
	}

	chainID := big.NewInt(31611)
	out := cfg.EthereumConfig(chainID)

	require.NotNil(t, out)
	require.Equal(t, chainID, out.ChainID, "ChainID")

	// Each fork-time field must point to its dedicated sentinel; any
	// swap between two fields would surface as a mismatch here.
	assertTime := func(t *testing.T, name string, got *uint64, want uint64) {
		t.Helper()
		require.NotNil(t, got, "%s: expected non-nil pointer", name)
		require.Equal(t, want, *got, "%s mismatch (cross-wired field?)", name)
	}

	assertTime(t, "ShanghaiTime", out.ShanghaiTime, shanghaiSentinel)
	assertTime(t, "CancunTime", out.CancunTime, cancunSentinel)
	assertTime(t, "PragueTime", out.PragueTime, pragueSentinel)
	assertTime(t, "OsakaTime", out.OsakaTime, osakaSentinel)
	assertTime(t, "BPO1Time", out.BPO1Time, bpo1Sentinel)
	assertTime(t, "BPO2Time", out.BPO2Time, bpo2Sentinel)
	assertTime(t, "BPO3Time", out.BPO3Time, bpo3Sentinel)
	assertTime(t, "BPO4Time", out.BPO4Time, bpo4Sentinel)
	assertTime(t, "BPO5Time", out.BPO5Time, bpo5Sentinel)
	assertTime(t, "AmsterdamTime", out.AmsterdamTime, amsterdamSentinel)
	assertTime(t, "VerkleTime", out.VerkleTime, verkleSentinel)

	// nil time fields must remain nil on the returned config; a
	// regression that defaulted unset times to zero would silently
	// activate every fork at genesis.
	nilCfg := ChainConfig{}
	nilOut := nilCfg.EthereumConfig(chainID)
	require.Nil(t, nilOut.ShanghaiTime)
	require.Nil(t, nilOut.CancunTime)
	require.Nil(t, nilOut.PragueTime)
	require.Nil(t, nilOut.OsakaTime)
	require.Nil(t, nilOut.BPO1Time)
	require.Nil(t, nilOut.BPO2Time)
	require.Nil(t, nilOut.BPO3Time)
	require.Nil(t, nilOut.BPO4Time)
	require.Nil(t, nilOut.BPO5Time)
	require.Nil(t, nilOut.AmsterdamTime)
	require.Nil(t, nilOut.VerkleTime)
}
