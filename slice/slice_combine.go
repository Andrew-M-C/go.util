package slice

func CombineEvenly[T comparable](s1, s2 []T) []T {
	if len(s1) < len(s2) {
		s1, s2 = s2, s1
	}
	if len(s1) == 0 {
		return []T{}
	}
	if len(s1) == len(s2) {
		res := make([]T, 0, len(s1)*2)
		for i := 0; i < len(s1); i++ {
			res = append(res, s1[i])
			res = append(res, s2[i])
		}
		return res
	}

	total := len(s1) + len(s2)
	inserted := make([]bool, total)
	out := make([]T, total)

	// 由于 lenA >= lenB，因此第一个位置必然是 A。
	// 首先计算出 A 插入位置的步长
	step := float64(len(s1)+len(s2)-1) / float64(len(s1)-1)

	// 第一个位置必然是 A
	out[0] = s1[0]
	inserted[0] = true

	// 后续位置按照步长插入
	next := step
	for i := 1; i < len(s1); i++ {
		pos := round64(next)
		if pos >= total {
			break
		}
		out[pos] = s1[i]
		inserted[pos] = true
		next += step
	}

	// 剩余位置用 B 插入
	v2Index := 0
	for i, notNil := range inserted {
		if notNil {
			continue
		}
		if v2Index > len(s2)-1 {
			break
		}
		out[i] = s2[v2Index]
		v2Index++
	}

	return out
}

func round64(f float64) int {
	f += 0.5
	return int(f)
}
