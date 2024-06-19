package types

// DefaultGenesis returns the default genesis state
func DefaultGenesis() *GenesisState {
	return &GenesisState{
		Stakes:   []Stake{},
		Delegations: []Delegation{},
	}
}

// Validate performs basic genesis state validation
func (gs GenesisState) Validate() error {
	// Add validation logic if needed
	return nil
}
