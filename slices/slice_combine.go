package slices

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
	// 为了减少浮点数计算, 我们算比较短的 B 插入位置的步长
	step := float64(len(s1)) / float64(len(s2)+1)

	// 按照步长插入 B
	next := step
	for _, v := range s2 {
		pos := round64(next)
		out[pos] = v
		inserted[pos] = true
		next += step + 1
	}

	// 剩余位置用 A 插入
	i1 := 0
	for i, notNil := range inserted {
		if notNil {
			continue
		}
		out[i] = s1[i1]
		i1++
	}

	return out
}

func round64(f float64) int {
	f += 0.5
	return int(f)
}
