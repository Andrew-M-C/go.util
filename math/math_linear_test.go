package math

import "testing"

func testLinear(t *testing.T) {
	cv("Normalize", func() { testLinearNormalize(t) })
}

func testLinearNormalize(t *testing.T) {
	in := []float64{0, 25, 200, 100, 150}
	out := Normalize(in, -1, 1)

	so(out[0], eq, -1)
	so(out[1], eq, -0.75)
	so(out[2], eq, 1)
	so(out[3], eq, 0)
	so(out[4], eq, 0.5)
}
