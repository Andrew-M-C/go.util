package slice

import (
	"sort"

	"golang.org/x/exp/constraints"
)

// List 表示一个 slice 类型
type List[T constraints.Ordered] []T

// ToList 转为 List 类型
func ToList[T constraints.Ordered](cli []T) List[T] {
	return List[T](cli)
}

// SortAsc 按照升序 (<=) 排序。为了便于链式调用, 返回自己
func (l List[T]) SortAsc() List[T] {
	sort.Slice(l, func(i, j int) bool {
		return l[i] <= l[j]
	})
	return l
}

// SortDesc 按照降序 (>=) 排序。为了便于链式调用, 返回自己
func (l List[T]) SortDesc() List[T] {
	sort.Slice(l, func(i, j int) bool {
		return l[i] >= l[j]
	})
	return l
}

// Equal 判断两个 list 各位置上的成员是否相等
func (l List[T]) Equal(another List[T]) bool {
	if len(l) != len(another) {
		return false
	}

	for i, left := range l {
		right := another[i]
		if left != right {
			return false
		}
	}
	return true
}

// Shuffle 乱序
func (l List[T]) Shuffle() {
	n := len(l)
	if n == 0 {
		return
	}

	i := n - 1
	for ; i > 1<<31-1-1; i-- {
		j := int(internalInt63n(int64(i + 1)))
		l[i], l[j] = l[j], l[i]
	}
	for ; i > 0; i-- {
		j := int(internalInt31n(int32(i + 1)))
		l[i], l[j] = l[j], l[i]
	}

	// rand.Shuffle(len(l), func(i, j int) {
	// 	l[i], l[j] = l[j], l[i]
	// })
}

// Copy 制作一个副本
func (l List[T]) Copy() List[T] {
	res := make(List[T], len(l))
	copy(res, l)
	return res
}

// Deduplicate 创建一个副本并去重
func (l List[T]) Deduplicate() List[T] {
	res := make(List[T], 0, len(l))
	set := make(map[T]struct{}, len(l))

	for _, v := range l {
		if _, exist := set[v]; exist {
			continue
		}
		res = append(res, v)
		set[v] = struct{}{}
	}

	return res
}
