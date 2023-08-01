package slice

import "golang.org/x/exp/constraints"

// Number 表示所有的实数类型
type Number interface {
	constraints.Float | constraints.Integer
}

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

// SetElement 往一个切片中设置值, 如果是负数, 表示设置在最后一个 (-1) 位置。不论是正数还是负数,
// 如果超出范围, 均无法设置, 并且 inRange 返回 false。
func SetElement[T any, I constraints.Signed](sli []T, signedIndex I, value T) (inRange bool) {
	if signedIndex >= 0 {
		if int(signedIndex) >= len(sli) {
			return false
		}
		sli[signedIndex] = value
		return true
	}

	// 从最后算起
	index := len(sli) + int(signedIndex)
	if index < 0 {
		return false
	}
	sli[index] = value
	return true
}

// Sum 求和
func Sum[T Number](numbers []T) T {
	var res T
	for _, n := range numbers {
		res += n
	}
	return res
}

// AverageFloat 求平均值, 返回值是浮点数
func AverageFloat[T Number](numbers []T) float64 {
	sum := Sum(numbers)
	return float64(sum) / float64(len(numbers))
}

// Minimum 找最小值
func Minimum[T Number](numbers []T) T {
	if len(numbers) == 0 {
		return 0
	}
	min := numbers[0]
	le := len(numbers)
	for i := 1; i < le; i++ {
		if n := numbers[i]; n < min {
			min = n
		}
	}
	return min
}

// Maximum 找最小值
func Maximum[T Number](numbers []T) T {
	if len(numbers) == 0 {
		return 0
	}
	max := numbers[0]
	le := len(numbers)
	for i := 1; i < le; i++ {
		if n := numbers[i]; n > max {
			max = n
		}
	}
	return max
}
