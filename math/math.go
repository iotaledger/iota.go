package math

// Abs returns the absolute value of an int64 value.
func Abs(n int64) int64 {
	y := n >> 63
	return (n ^ y) - y
}
