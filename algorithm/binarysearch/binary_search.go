// Package binarysearch 提供简单的二分查找法实现
package binarysearch

import (
	"github.com/Andrew-M-C/go.util/algorithm/constraints"
)

// CompareFunction 表示比较函数
type CompareFunction[T any] func(a, b T) int

// ---- SearchOne ----

// SearchOne 搜索一个可比较的切片
func SearchOne[T constraints.Ordered](s []T, tgt T) int {
	return SearchOneByFunc(s, tgt, func(a, b T) int {
		switch {
		case a < b:
			return -1
		case a > b:
			return 1
		default:
			return 0
		}
	})
}

// SearchOneByFunc 使用比较函数进行大小判断来搜索一个值, 不论是一系列的左值还是右值
func SearchOneByFunc[T any](s []T, tgt T, compareFunc CompareFunction[T]) int {
	if len(s) == 0 {
		return -1
	}
	left, right := 0, len(s)-1

	for left < right-1 {
		mid := (left + right) / 2
		compare := compareFunc(tgt, s[mid])
		switch {
		case compare < 0: // tgt < mid
			right = mid
		case compare > 0: // tgt > mid
			left = mid
		default: // tgt == mid
			return mid
		}
	}
	if left < right {
		if compareFunc(tgt, s[right]) == 0 {
			return right
		}
	}
	if compareFunc(s[left], tgt) == 0 {
		return left
	}
	return -1
}

// ---- SearchFloor ----

// SearchFloor 搜索等于 tgt 的最左边的值
func SearchFloor[T constraints.Ordered](s []T, tgt T) int {
	return SearchFloorByFunc(s, tgt, func(a, b T) int {
		switch {
		case a < b:
			return -1
		case a > b:
			return 1
		default:
			return 0
		}
	})
}

// SearchFloorByFunc 搜索等于 tgt 的最左边的值
func SearchFloorByFunc[T any](s []T, tgt T, compareFunc CompareFunction[T]) int {
	if len(s) == 0 {
		return -1
	}
	left, right := 0, len(s)-1

	for left < right-1 {
		mid := (left + right) / 2
		compare := compareFunc(tgt, s[mid])
		switch {
		case compare < 0: // tgt < mid
			right = mid
			continue
		case compare > 0: // tgt > mid
			left = mid
			continue
		default:
			// found, go on
		}

		// 到边了
		if mid == 0 {
			return 0
		}

		// 通过左值判断是不是找到边了
		if compareFunc(tgt, s[mid-1]) != 0 {
			return mid
		}
		right = mid - 1
	}

	if compareFunc(tgt, s[left]) == 0 {
		return left
	}
	if compareFunc(tgt, s[right]) == 0 {
		return right
	}
	return -1
}

// ---- SearchCeiling ----

// SearchCeiling 搜索等于 tgt 的最左边的值
func SearchCeiling[T constraints.Ordered](s []T, tgt T) int {
	return SearchCeilingByFunc(s, tgt, func(a, b T) int {
		switch {
		case a < b:
			return -1
		case a > b:
			return 1
		default:
			return 0
		}
	})
}

// SearchCeilingByFunc 搜索等于 tgt 的最左边的值
func SearchCeilingByFunc[T any](s []T, tgt T, compareFunc CompareFunction[T]) int {
	if len(s) == 0 {
		return -1
	}
	left, right := 0, len(s)-1

	for left < right-1 {
		mid := (left + right) / 2
		compare := compareFunc(tgt, s[mid])
		switch {
		case compare < 0: // tgt < mid
			right = mid
			continue
		case compare > 0: // tgt > mid
			left = mid
			continue
		default:
			// found, go on
		}

		// 到边了
		if mid == len(s)-1 {
			return len(s) - 1
		}

		// 通过右值判断是不是找到边了
		if compareFunc(s[mid+1], tgt) != 0 {
			return mid
		}
		left = mid + 1
	}

	if compareFunc(tgt, s[right]) == 0 {
		return right
	}
	if compareFunc(tgt, s[left]) == 0 {
		return left
	}
	return -1
}
