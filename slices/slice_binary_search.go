package slices

import "github.com/Andrew-M-C/go.util/constraints"

// BinarySearchOne 二分查找一个。注意入参必须已经从小到大排了序, 否则结果无法确定。
// 函数返回 index, 如果 -1 则表示找不到
func BinarySearchOne[T any](sli []T, target T, compare func(a, b T) int) int {
	from, to := 0, len(sli)
	if to == 0 {
		return -1
	}
	if compare(target, sli[0]) < 0 {
		return -1
	}
	if compare(target, sli[to-1]) > 0 {
		return -1
	}

	for from < to-1 {
		mid := from + (to-from)/2
		n := sli[mid]
		comp := compare(n, target)
		switch {
		default:
			return mid
		case comp < 0:
			from = mid
		case comp > 0:
			to = mid
		}
	}

	if compare(sli[from], target) == 0 {
		return from
	}
	return -1
}

// BinarySearch 二分法搜索一段已排序的数
func BinarySearch[T constraints.Ordered](sli []T, target T) (from, to int, hit bool) {
	from, to = 0, len(sli)
	internal.debugf("input slice %v, length %d, target %v", sli, to, target)
	if to == 0 {
		return 0, 0, false
	}
	if target < sli[0] {
		return 0, 0, false
	}
	if target > sli[to-1] {
		return to, to, false
	}

	// 二分
	for from < to-1 {
		half := from + (to-from)/2
		n := sli[half]
		internal.debugf("from %d, to %d, half %d (%v)", from, to, half, n)
		switch {
		default:
			internal.debugf("Hit index %d (%v)", half, n)
			from, to = half, half+1 // causing break
		case n < target:
			from = half
		case n > target:
			to = half
		}
	}

	if sli[from] != target {
		return from, to, false
	}

	// 扩展命中的值域
	for ; from > 0 && sli[from-1] == target; from-- {
		// nothing
	}
	for ; to < len(sli) && sli[to] == target; to++ {
		// nothing
	}

	return from, to, true
}
