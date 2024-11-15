package types

import (
	"cosmossdk.io/math"
	"github.com/cometbft/cometbft/proto/tendermint/crypto"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// ValidatorCompat is a wrapper for the Poa Validator type.
// It is used in Connect as a compatibility wrapper for the SDK's x/staking keeper types.
type ValidatorCompat struct {
	Validator
}

// GetMoniker is not implemented
func (v ValidatorCompat) GetMoniker() string {
	panic("not implemented")
}

// IsJailed is not implemented
func (v ValidatorCompat) IsJailed() bool {
	panic("not implemented")
}

// GetStatus is not implemented
func (v ValidatorCompat) GetStatus() stakingtypes.BondStatus {
	panic("not implemented")
}

// IsBonded is not implemented
func (v ValidatorCompat) IsBonded() bool {
	panic("not implemented")
}

// IsUnbonded is not implemented
func (v ValidatorCompat) IsUnbonded() bool {
	panic("not implemented")
}

// IsUnbonding is not implemented
func (v ValidatorCompat) IsUnbonding() bool {
	panic("not implemented")
}

// GetOperator is not implemented
func (v ValidatorCompat) GetOperator() string {
	panic("not implemented")
}

// ConsPubKey is not implemented
func (v ValidatorCompat) ConsPubKey() (cryptotypes.PubKey, error) {
	panic("not implemented")
}

// TmConsPublicKey is not implemented
func (v ValidatorCompat) TmConsPublicKey() (crypto.PublicKey, error) {
	panic("not implemented")
}

// GetConsAddr is not implemented
func (v ValidatorCompat) GetConsAddr() ([]byte, error) {
	panic("not implemented")
}

// GetTokens is not implemented
func (v ValidatorCompat) GetTokens() math.Int {
	panic("not implemented")
}

// GetBondedTokens returns 1 for every validator since the PoaKeeper weights each validator equally.
func (v ValidatorCompat) GetBondedTokens() math.Int {
	// Always returns 1 since each validator has an equal stake
	return math.NewInt(1)
}

// GetConsensusPower is not implemented
func (v ValidatorCompat) GetConsensusPower(m math.Int) int64 {
	panic("not implemented")
}

// GetCommission is not implemented
func (v ValidatorCompat) GetCommission() math.LegacyDec {
	panic("not implemented")
}

// GetMinSelfDelegation is not implemented
func (v ValidatorCompat) GetMinSelfDelegation() math.Int {
	panic("not implemented")
}

// GetDelegatorShares is not implemented
func (v ValidatorCompat) GetDelegatorShares() math.LegacyDec {
	panic("not implemented")
}

// TokensFromShares is not implemented
func (v ValidatorCompat) TokensFromShares(dec math.LegacyDec) math.LegacyDec {
	panic("not implemented")
}

// TokensFromSharesTruncated is not implemented
func (v ValidatorCompat) TokensFromSharesTruncated(dec math.LegacyDec) math.LegacyDec {
	panic("not implemented")
}

// TokensFromSharesRoundUp is not implemented
func (v ValidatorCompat) TokensFromSharesRoundUp(dec math.LegacyDec) math.LegacyDec {
	panic("not implemented")
}

// SharesFromTokens is not implemented
func (v ValidatorCompat) SharesFromTokens(amt math.Int) (math.LegacyDec, error) {
	panic("not implemented")
}

// SharesFromTokensTruncated is not implemented
func (v ValidatorCompat) SharesFromTokensTruncated(amt math.Int) (math.LegacyDec, error) {
	panic("not implemented")
}
