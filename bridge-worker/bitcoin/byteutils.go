package bitcoin

// Reverse reverses bytes order in the slice. The function copies the slice and
// returns a new reversed slice, keeping the source slice untouched.
func Reverse(bytes []byte) []byte {
	result := make([]byte, 0, len(bytes))
	for i := len(bytes) - 1; i >= 0; i-- {
		result = append(result, bytes[i])
	}
	return result
}
