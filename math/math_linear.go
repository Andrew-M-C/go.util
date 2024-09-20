package math

import "github.com/Andrew-M-C/go.util/constraints"

// Normalize 归一化
func Normalize[T constraints.Float](in []T, min, max T) []T {
	if min > max {
		min, max = max, min
	}
	if len(in) == 0 {
		return nil
	}
	if min == max {
		res := make([]T, len(in))
		for i := range res {
			res[i] = min
		}
		return res
	}

	inMin, inMax := in[0], in[0]
	for _, n := range in {
		if n < inMin {
			inMin = n
		} else if n > inMax {
			inMax = n
		}
	}

	mul := (max - min) / (inMax - inMin)
	out := make([]T, len(in))
	for i, n := range in {
		if n == inMin {
			out[i] = min
		} else if n == inMax {
			out[i] = max
		} else {
			out[i] = (n-inMin)*mul + min
		}
	}

	return out
}
