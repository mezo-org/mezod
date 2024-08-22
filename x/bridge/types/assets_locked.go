package types

// AssetsLockedEventsCmp is a sequence-based comparator for AssetsLockedEvent
// that defines the natural order for a slice of AssetsLockedEvent.
func AssetsLockedEventsCmp(a, b AssetsLockedEvent) int {
	return a.Sequence.BigInt().Cmp(b.Sequence.BigInt())
}
