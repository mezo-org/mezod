package types

import "fmt"

// DefaultMaxERC20TokensMappings is the default maximum number of ERC20 tokens
// mappings that can be supported by the bridge.
const DefaultMaxERC20TokensMappings = uint32(20)

// NewParams creates a new Params instance.
func NewParams(maxERC20TokensMappings uint32) Params {
	return Params{
		MaxErc20TokensMappings: maxERC20TokensMappings,
	}
}

// DefaultParams returns a default set of parameters.
func DefaultParams() Params {
	return NewParams(DefaultMaxERC20TokensMappings)
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
