// Package constraints 参照 golang.org/x/exp/constraints, 但后者迭代太快经常使用新版本,
// 因此放弃此依赖转而自己写一个
package constraints

type Signed interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64
}

type Unsigned interface {
	~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64
}

type Integer interface {
	Signed | Unsigned
}

type Float interface {
	~float32 | ~float64
}


type Complex interface {
	~complex64 | ~complex128
}

type Ordered interface {
	Integer | Float | ~string
}

// Abs 返回绝对值
func Abs[T Signed | Float](x T) T {
	if x >= 0 {
		return x
	}
	return -x
}

// Min 返回两个数中的最小值, 实际上在 1.21 开始就已经有内置的 min 了, 以防开发者不知道
func Min[T Ordered](a, b T) T {
	return min(a, b)
}

// Max 返回两个数中的最大值, 实际上在 1.21 开始就已经有内置的 max 了, 以防开发者不知道
func Max[T Ordered](a, b T) T {
	return max(a, b)
}

