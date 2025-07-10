package math

import (
	"testing"
)

func testIntSqrt(t *testing.T) {
	cv("simple logic", func() {
		check := func(x uint64) {
			if t.Failed() {
				return
			}
			a := newtonIntSqrt(x)
			b := bitwiseSqrt(x)

			so(b, eq, a)
			t.Logf("sqrt(%d) = %d", x, a)
		}

		check(257)
		check(1000000000)
		check(0x1000000000000000)
		check(0xFFFFFFFFFFFFFFFF)
	})
}

// bitwiseSqrt 逐比特确认法
func bitwiseSqrt(x uint64) uint64 {
	if x <= 1 {
		return x
	}

	sqrt := uint64(0)
	shift := int64(31)
	sqrt2 := uint64(0)

	for shift >= 0 {
		sqrt2 = ((sqrt << 1) + (1 << shift)) << shift
		if sqrt2 <= x {
			sqrt += (1 << shift)
			x -= sqrt2
		}
		shift--
	}
	return sqrt
}
