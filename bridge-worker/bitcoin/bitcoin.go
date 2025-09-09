package bitcoin

// ByteOrder represents the byte order used by the Bitcoin byte arrays. The
// Bitcoin ecosystem is not totally consistent in this regard and different
// byte orders are used depending on the purpose.
type ByteOrder int

const (
	// InternalByteOrder represents the internal byte order used by the Bitcoin
	// protocol. This is the primary byte order that is suitable for the
	// use cases related with the protocol logic and cryptography. Byte arrays
	// using this byte order should be converted to numbers according to
	// the little-endian sequence.
	InternalByteOrder ByteOrder = iota

	// ReversedByteOrder represents the "human" byte order. This is the
	// byte order that is typically used by the third party services like
	// block explorers or Bitcoin chain clients. Byte arrays using this byte
	// order should be converted to numbers according to the big-endian
	// sequence. This type is also known as the `RPC Byte Order` in the
	// Bitcoin specification.
	ReversedByteOrder
)
