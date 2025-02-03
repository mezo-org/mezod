package types

import (
	"fmt"

	evmtypes "github.com/mezo-org/mezod/x/evm/types"

	sdkmath "cosmossdk.io/math"
)

// DefaultGenesis returns the default genesis state.
//
// WARNING: The default genesis state has an empty source BTC token, hence
// it is invalid (Validate will fail). A proper BTC token must be set at
// later stages, before running the network.
func DefaultGenesis() *GenesisState {
	return &GenesisState{
		Params:                  DefaultParams(),
		AssetsLockedSequenceTip: sdkmath.NewInt(0),
		SourceBtcToken:          "",
		Erc20TokensMappings:     nil,
	}
}

// Validate performs basic genesis state validation returning an error upon any
// failure.
func (gs GenesisState) Validate() error {
	if err := gs.Params.Validate(); err != nil {
		return err
	}

	if gs.AssetsLockedSequenceTip.IsNegative() {
		return fmt.Errorf(
			"genesis sequence tip cannot be negative: %s",
			gs.AssetsLockedSequenceTip,
		)
	}

	if len(gs.SourceBtcToken) == 0 {
		return fmt.Errorf("source btc token cannot be empty")
	}

	if !evmtypes.IsHexAddress(gs.SourceBtcToken) {
		return fmt.Errorf("source btc token must be a valid hex-encoded EVM address")
	}

	for i, mapping := range gs.Erc20TokensMappings {
		if !evmtypes.IsHexAddress(mapping.SourceToken) {
			return fmt.Errorf(
				"source token of ERC20 mapping %d must be a valid hex-encoded EVM address",
				i,
			)
		}

		if !evmtypes.IsHexAddress(mapping.MezoToken) {
			return fmt.Errorf(
				"mezo token of ERC20 mapping %d must be a valid hex-encoded EVM address",
				i,
			)
		}
	}

	return nil
}
