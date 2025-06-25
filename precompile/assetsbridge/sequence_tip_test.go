package assetsbridge_test

import (
	"math/big"

	"cosmossdk.io/math"
)

func (s *PrecompileTestSuite) TestGetCurrentSequenceTip() {
	testcases := []TestCase{
		{
			name: "0 sequence tip, network bootstrap",
			run: func() []any {
				return nil
			},
			basicPass: true,
			output:    []any{new(big.Int).Set(big.NewInt(0))},
		},
		{
			name: "sequence tip have increased",
			run: func() []any {
				s.bridgeKeeper.setAssetsLockedSequenceTip(math.NewIntFromUint64(42))
				return nil
			},
			basicPass: true,
			output:    []any{new(big.Int).Set(big.NewInt(42))},
		},
	}

	s.RunMethodTestCases(testcases, "getCurrentSequenceTip")
}
