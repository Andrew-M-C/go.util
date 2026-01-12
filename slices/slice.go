package slices

import (
	"math/rand"
	"sort"

	"github.com/Andrew-M-C/go.util/slices/constraints"
)

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
func Sum[T constraints.Number](numbers []T) T {
	var res T
	for _, n := range numbers {
		res += n
	}
	return res
}

// AverageFloat 求平均值, 返回值是浮点数
func AverageFloat[T constraints.Number](numbers []T) float64 {
	sum := Sum(numbers)
	return float64(sum) / float64(len(numbers))
}

// Minimum 找最小值
func Minimum[T constraints.Number](numbers []T) T {
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
func Maximum[T constraints.Number](numbers []T) T {
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

// Reverse a slice
func Reverse[T any](sli []T) {
	for i, j := 0, len(sli)-1; i < j; i, j = i+1, j-1 {
		sli[i], sli[j] = sli[j], sli[i]
	}
}

// Copy makes a copy
func Copy[T any](sli []T) []T {
	if sli == nil {
		return nil
	}
	res := make([]T, len(sli))
	copy(res, sli)
	return res
}

// Shuffle shuffles a slice
func Shuffle[T any](tgt []T) {
	rand.Shuffle(len(tgt), func(i, j int) {
		tgt[i], tgt[j] = tgt[j], tgt[i]
	})
}

// CutIntoSectors 阿照 limit 分成一段一段的
func CutIntoSectors[T any](sli []T, sectorLimit int) [][]T {
	if sectorLimit <= 0 {
		return [][]T{sli}
	}
	var res [][]T
	for len(sli) > sectorLimit {
		res = append(res, sli[:sectorLimit])
		sli = sli[sectorLimit:]
	}
	if len(sli) > 0 {
		res = append(res, sli)
	}
	return res
}

// Sort 封装 sort.Slice
func Sort[T any](sli []T, lessFunc func(i, j T) bool) {
	sort.Slice(sli, func(i, j int) bool {
		return lessFunc(sli[i], sli[j])
	})
}

// EnsureLength 确保切片长度至少为 length, 如果已经有足够长度了, 则什么都不做。如果长度不够,
// 则使用默认值填充。
func EnsureLength[T any](sli []T, length int, defaultValue ...T) []T {
	if length <= 0 || len(sli) >= length {
		return sli
	}
	var v T
	if len(defaultValue) > 0 {
		v = defaultValue[0]
	}
	for i := len(sli); i < length; i++ {
		sli = append(sli, v)
	}
	return sli
}

// Insert 在指定位置插入一个值。如果指定位置大于切片的长度, 则什么都不做。参数 index 可以小于
// 0, 表示从后面算起, -1 表示插入到最后一个元素的前面。但如果从末尾算起也超过切片的长度的话,
// 依然什么都不做
func Insert[T any](sli []T, index int, value T) []T {
	if index >= 0 {
		return insertForward(sli, index, value)
	}
	return insertBackward(sli, index, value)
}

// 正向插入
func insertForward[T any](sli []T, index int, value T) []T {
	if index >= len(sli) {
		return sli // 什么都不做
	}
	var emptyValue T
	sli = append(sli, emptyValue)
	copy(sli[index+1:], sli[index:])
	sli[index] = value
	return sli
}

// 从末尾插入
func insertBackward[T any](sli []T, index int, value T) []T {
	index += len(sli)
	if index < 0 {
		return sli // 什么都不做
	}
	return insertForward(sli, index, value)
}

// Remove 删除指定位置的值。如果指定位置大于切片的长度, 则什么都不做。参数 index 可以小于
// 0, 表示从后面算起, -1 表示删除最后一个元素。但如果从末尾算起也超过切片的长度的话, 依然什么都不做
func Remove[T any](sli []T, index int) []T {
	if index >= 0 {
		return removeForward(sli, index)
	}
	return removeBackward(sli, index)
}

func removeForward[T any](sli []T, index int) []T {
	if index >= len(sli) {
		return sli // 什么都不做
	}
	copy(sli[index:], sli[index+1:])
	sli = sli[:len(sli)-1]
	return sli
}

func removeBackward[T any](sli []T, index int) []T {
	index += len(sli)
	if index < 0 {
		return sli // 什么都不做
	}
	return removeForward(sli, index)
}
