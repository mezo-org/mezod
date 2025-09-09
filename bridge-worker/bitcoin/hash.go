package bitcoin

import (
	"crypto/sha256"
	"encoding/hex"
)

// HashByteLength is the byte length of the Hash type.
const HashByteLength = 32

// Hash represents the double SHA-256 of some arbitrary data using the
// InternalByteOrder.
type Hash [HashByteLength]byte

// String returns the unprefixed hexadecimal string representation of the Hash
// in the InternalByteOrder.
func (h Hash) String() string {
	return h.Hex(InternalByteOrder)
}

// Hex returns the unprefixed hexadecimal string representation of the Hash
// in the given ByteOrder.
func (h Hash) Hex(byteOrder ByteOrder) string {
	switch byteOrder {
	case InternalByteOrder:
		return hex.EncodeToString(h[:])
	case ReversedByteOrder:
		for i := 0; i < HashByteLength/2; i++ {
			h[i], h[HashByteLength-1-i] = h[HashByteLength-1-i], h[i]
		}
		return hex.EncodeToString(h[:])
	default:
		panic("unknown byte order")
	}
}

// ComputeHash computes the Hash for the provided data.
func ComputeHash(data []byte) Hash {
	first := sha256.Sum256(data)
	return sha256.Sum256(first[:])
}
