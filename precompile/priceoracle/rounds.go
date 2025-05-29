package priceoracle

import (
	"fmt"
	"math/big"

	"github.com/mezo-org/mezod/precompile"
	oracletypes "github.com/skip-mev/connect/v2/x/oracle/types"
)

// LatestRoundDataMethodName is the name of the latestRoundData method. It
// matches the name of the method in the contract ABI.
const LatestRoundDataMethodName = "latestRoundData"

type LatestRoundDataMethod struct {
	oracleQueryServer OracleQueryServer
}

func newLatestRoundDataMethod(oracleQueryServer OracleQueryServer) *LatestRoundDataMethod {
	return &LatestRoundDataMethod{
		oracleQueryServer: oracleQueryServer,
	}
}

func (m *LatestRoundDataMethod) MethodName() string {
	return LatestRoundDataMethodName
}

func (m *LatestRoundDataMethod) MethodType() precompile.MethodType {
	return precompile.Read
}

func (m *LatestRoundDataMethod) RequiredGas(_ []byte) (uint64, bool) {
	// Fallback to the default gas calculation.
	return 0, false
}

func (m *LatestRoundDataMethod) Payable() bool {
	return false
}

func (m *LatestRoundDataMethod) Run(
	ctx *precompile.RunContext,
	inputs precompile.MethodInputs,
) (precompile.MethodOutputs, error) {
	if err := precompile.ValidateMethodInputsCount(inputs, 0); err != nil {
		return nil, err
	}

	priceData, err := m.oracleQueryServer.GetPrice(
		ctx.SdkCtx(),
		&oracletypes.GetPriceRequest{
			CurrencyPair: "BTC/USD",
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get price: [%w]", err)
	}

	// Just in case to avoid unexpected panics.
	if priceData == nil || priceData.Price == nil {
		return nil, fmt.Errorf("price data is nil")
	}

	priceValue := priceData.Price.Price

	if priceValue.IsNil() || !priceValue.IsPositive() {
		// Better to fail fast and revert the upstream transaction than
		// feed it with zero price.
		return nil, fmt.Errorf("price value is nil or non-positive")
	}

	targetDecimals := uint64(Decimals)
	actualDecimals := priceData.Decimals
	deltaDecimals := int64(targetDecimals - actualDecimals) //nolint:gosec

	var answer *big.Int
	if deltaDecimals >= 0 {
		// Adjust to the desired precision up: priceValue * 10^deltaDecimals.
		answer = new(big.Int).Mul(
			priceValue.BigInt(),
			new(big.Int).Exp(big.NewInt(10), big.NewInt(deltaDecimals), nil),
		)
	} else {
		// Adjust to the desired precision down: priceValue / 10^|deltaDecimals|.
		deltaDecimalsAbs := new(big.Int).Abs(big.NewInt(deltaDecimals))
		answer = new(big.Int).Div(
			priceValue.BigInt(),
			new(big.Int).Exp(big.NewInt(10), deltaDecimalsAbs, nil),
		)
	}

	roundID := new(big.Int).SetUint64(priceData.Nonce)
	startedAt := big.NewInt(priceData.Price.BlockTimestamp.Unix())
	updatedAt := startedAt
	answeredInRound := big.NewInt(0) // deprecated field, returning 0 for consistency

	return precompile.MethodOutputs{
		roundID,
		answer,
		startedAt,
		updatedAt,
		answeredInRound,
	}, nil
}
