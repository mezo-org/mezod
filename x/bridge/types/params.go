package types

import "fmt"

// DefaultMaxErc20Tokens is the default maximum number of ERC20 tokens that can
// be supported by the bridge.
const DefaultMaxErc20Tokens = uint32(20)

// NewParams creates a new Params instance.
func NewParams(maxErc20Tokens uint32) Params {
	return Params{
		MaxErc20Tokens: maxErc20Tokens,
	}
}

// DefaultParams returns a default set of parameters.
func DefaultParams() Params {
	return NewParams(DefaultMaxErc20Tokens)
}

// Validate validates the set of params.
func (p Params) Validate() error {
	return nil
}

// String implements the Stringer interface.
func (p Params) String() string {
	return fmt.Sprintf("Params(MaxErc20Tokens: %d)", p.MaxErc20Tokens)
}
