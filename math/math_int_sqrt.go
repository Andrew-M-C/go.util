package math

// SqrtUint sqrt(x) but returns integer part only.
func SqrtUint(x uint64) uint64 {
	return newtonIntSqrt(x)
}

// SqrtUint sqrt(x) but returns integer part only.
func SqrtInt(x int64) complex128 {
	if x >= 0 {
		res := SqrtUint(uint64(x))
		return complex(float64(res), 0)
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
