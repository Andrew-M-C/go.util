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
