package slice

import "golang.org/x/exp/constraints"

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

// Element 读取切片中的值, 如果是负数, 表示从最后一个 (-1) 读起。不论是正数还是负数, 如果超出
// 范围, 返回的 value 均无效, 并且 inRange 返回 false。
func Element[T any, I constraints.Signed](sli []T, signedIndex I) (value T, inRange bool) {
	if signedIndex >= 0 {
		if int(signedIndex) >= len(sli) {
			return
		}
		return sli[signedIndex], true
	}

	// 从最后算起
	index := len(sli) + int(signedIndex)
	if index < 0 {
		return
	}
	return sli[index], true
}
