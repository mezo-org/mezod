package types

import "fmt"

const (
	// DefaultMaxERC20TokensMappings is the default maximum number of ERC20 tokens
	// mappings that can be supported by the bridge.
	DefaultMaxERC20TokensMappings = uint32(20)

	// DefaultBtcSupplyAssertionEnabled is the default value for the flag
	// steering the BTC supply assertion.
	DefaultBtcSupplyAssertionEnabled = true
)

// NewParams creates a new Params instance.
func NewParams(
	maxERC20TokensMappings uint32,
	btcSupplyAssertionEnabled bool,
) Params {
	return Params{
		MaxErc20TokensMappings:    maxERC20TokensMappings,
		BtcSupplyAssertionEnabled: btcSupplyAssertionEnabled,
	}
}

// DefaultParams returns a default set of parameters.
func DefaultParams() Params {
	return NewParams(
		DefaultMaxERC20TokensMappings,
		DefaultBtcSupplyAssertionEnabled,
	)
}

// Validate validates the set of params.
func (p Params) Validate() error {
	return nil
}

// String implements the Stringer interface.
func (p Params) String() string {
	return fmt.Sprintf(
		"Params(MaxErc20TokensMappings: %d)",
		p.MaxErc20TokensMappings,
	)
}
