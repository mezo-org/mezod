package priceoracle_test

import (
	"math/big"
	"time"

	sdkmath "cosmossdk.io/math"
)

func (s *PrecompileTestSuite) TestLatestRoundData() {
	blockTimestamp := time.Date(2025, 1, 15, 10, 30, 25, 0, time.UTC)

	// 55555 USD with 10^18 precision
	expectedPrice, ok := new(big.Int).SetString(
		"55555000000000000000000",
		10,
	)
	s.Require().True(ok)

	// This is expected to be 0 but go-ethereum ABI machinery it as
	// a big.Int with empty bits slice. That said, we cannot use big.NewInt(0)
	// as this produces a big.Int with nil bits slice so the equality check fails.
	expectedAnsweredInRound := new(big.Int).SetBits([]big.Word{})

	testcases := []TestCase{
		{
			name: "argument count mismatch",
			run: func() []interface{} {
				return []interface{}{
					1,
				}
			},
			errContains: "argument count mismatch",
		},
		{
			name: "price is negative",
			run: func() []interface{} {
				// Set the oracle mock to return the desired price.
				s.oracleQueryServer.SetPrice(
					sdkmath.NewInt(-1),
					time.Time{},
					0,
					0,
					0,
				)

				return nil
			},
			basicPass:   true,
			revert:      true,
			errContains: "price value is nil or non-positive",
		},
		{
			name: "price is zero",
			run: func() []interface{} {
				// Set the oracle mock to return the desired price.
				s.oracleQueryServer.SetPrice(
					sdkmath.NewInt(0),
					time.Time{},
					0,
					0,
					0,
				)

				return nil
			},
			basicPass:   true,
			revert:      true,
			errContains: "price value is nil or non-positive",
		},
		{
			name: "precision lesser than desired",
			run: func() []interface{} {
				// Set the oracle mock to return the desired price.
				s.oracleQueryServer.SetPrice(
					// 55555 USD with 10^5 precision
					sdkmath.NewInt(5555500000),
					blockTimestamp,
					150,
					// precision of 5 is less than desired 18
					5,
					1,
				)

				return nil
			},
			basicPass: true,
			output: []interface{}{
				// round id should be equal to nonce
				big.NewInt(150),
				expectedPrice,
				// started at should be the block timestamp as UNIX time
				big.NewInt(blockTimestamp.Unix()),
				// updated at should be the block timestamp as UNIX time
				big.NewInt(blockTimestamp.Unix()),
				expectedAnsweredInRound,
			},
		},
		{
			name: "precision equal to desired",
			run: func() []interface{} {
				// 55555 USD with 10^18 precision
				price, priceOk := new(big.Int).SetString(
					"55555000000000000000000",
					10,
				)
				s.Require().True(priceOk)

				// Set the oracle mock to return the desired price.
				s.oracleQueryServer.SetPrice(
					sdkmath.NewIntFromBigInt(price),
					blockTimestamp,
					150,
					// precision of 18 is equal to desired 18
					18,
					1,
				)

				return nil
			},
			basicPass: true,
			output: []interface{}{
				// round id should be equal to nonce
				big.NewInt(150),
				expectedPrice,
				// started at should be the block timestamp as UNIX time
				big.NewInt(blockTimestamp.Unix()),
				// updated at should be the block timestamp as UNIX time
				big.NewInt(blockTimestamp.Unix()),
				expectedAnsweredInRound,
			},
		},
		{
			name: "precision greater than desired",
			run: func() []interface{} {
				// 55555 USD with 10^20 precision
				price, priceOk := new(big.Int).SetString(
					"5555500000000000000000055",
					10,
				)
				s.Require().True(priceOk)

				// Set the oracle mock to return the desired price.
				s.oracleQueryServer.SetPrice(
					sdkmath.NewIntFromBigInt(price),
					blockTimestamp,
					150,
					// precision of 20 is more than desired 18
					20,
					1,
				)

				return nil
			},
			basicPass: true,
			output: []interface{}{
				// round id should be equal to nonce
				big.NewInt(150),
				expectedPrice,
				// started at should be the block timestamp as UNIX time
				big.NewInt(blockTimestamp.Unix()),
				// updated at should be the block timestamp as UNIX time
				big.NewInt(blockTimestamp.Unix()),
				expectedAnsweredInRound,
			},
		},
	}

	s.RunMethodTestCases(testcases, "latestRoundData")
}
