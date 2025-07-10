package math

import "golang.org/x/exp/constraints"

// SqrtUint sqrt(x) but returns integer part only. If x is negative, it will return 0.
func SqrtUint[T constraints.Unsigned](x T) T {
	return T(newtonIntSqrt(uint64(x)))
}

// SqrtInt sqrt(x) and returns complex number. If x is negative, it will return
// complex number with imaginary part.
func SqrtInt[T constraints.Signed](x T) complex128 {
	if x >= 0 {
		res := SqrtUint(uint64(x))
		return complex(float64(res), 0.0)
	}

	x = -x
	res := SqrtUint(uint64(x))
	return complex(0, float64(res))
}

// newtonIntSqrt 使用牛顿迭代法进行开平方根
func newtonIntSqrt(x uint64) uint64 {
	x1 := x - 1
	s := 1
	var g0, g1 uint64

	if x1 > 0xFFFFFFFF {
		s += 16
		x1 = x1 >> 32
	}
	if x1 > 65535 {
		s += 8
		x1 = x1 >> 16
	}
	if x1 > 255 {
		s += 4
		x1 = x1 >> 8
	}
	if x1 > 15 {
		s += 2
		x1 = x1 >> 4
	}
	if x1 > 3 {
		s += 1
		// x1 = x1 >> 2
	}

	g0 = 1 << s
	g1 = (g0 + (x >> s)) >> 1

	for g1 < g0 {
		g0 = g1
		g1 = (g0 + x/g0) >> 1
	}

	return g0
}
