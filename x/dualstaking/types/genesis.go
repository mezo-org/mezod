package types

// DefaultGenesis returns the default genesis state
func DefaultGenesis() *GenesisState {
	return &GenesisState{
		StakingPositions:   []StakingPosition{},
		DelegationPositions: []DelegationPosition{},
	}
}

// Validate performs basic genesis state validation
func (gs GenesisState) Validate() error {
	// Add validation logic if needed
	return nil
}
