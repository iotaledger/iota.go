package math

import (
	"math"
)

// AbsInt64 returns the absolute value of an int64 value.
func AbsInt64(v int64) uint64 {
	if v == math.MinInt64 {
		return math.MaxInt64 + 1
	}
	if v < 0 {
		return uint64(-v)
	}
	return uint64(v)
}
