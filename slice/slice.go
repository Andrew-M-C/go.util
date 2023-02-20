package slice

// Equal 逐个比较两个切片里的值是否相等
func Equal[T comparable](a, b []T) bool {
	if len(a) != len(b) {
		return false
	}
	for i, va := range a {
		vb := b[i]
		if va != vb {
			return false
		}
	}
	return true
}

// HaveEqualValues 判断两个切片中是否拥有相同的值, 允许重复并且不考虑顺序
func HaveEqualValues[T comparable](a, b []T) bool {
	ka := make(map[T]struct{}, len(a))
	checkedKa := make(map[T]struct{}, len(a))
	for _, va := range a {
		ka[va] = struct{}{}
	}
	for _, vb := range b {
		if _, exist := ka[vb]; !exist {
			return false
		}
		checkedKa[vb] = struct{}{}
	}
	return len(ka) == len(checkedKa)
}
