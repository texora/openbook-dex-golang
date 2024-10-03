package openbookdexgolang

import "math"

func saturatingAdd(a, b int64) int64 {
	// Check for overflow in addition
	if b > 0 && a > math.MaxInt64-b {
		return math.MaxInt64
	} else if b < 0 && a < math.MinInt64-b {
		return math.MinInt64
	}
	return a + b
}
