// Package rand 提供官方 math/rand 的一些小工具封装
package rand

import (
	"math/rand"

	"github.com/Andrew-M-C/go.util/constraints"
)

// IndexByRatios 抽奖。
// 按照给定的比例, 随机选择一个选项。如果一个选项 <= 0 则表示不可能被选中。如果选项列表长度为
// 0 或者所有选项比例的总和 <= 0 则返回 -1。
func IndexByRatios[T any, N constraints.Integer](config []T, ratioGetter func(i int, v T) N) int {
	if len(config) == 0 {
		return -1
	}

	sumRatios := make([]N, len(config))
	for i, v := range config {
		pre := N(0)
		if i > 0 {
			pre = sumRatios[i-1]
		}
		cur := ratioGetter(i, v)
		if cur > 0 {
			sumRatios[i] = pre + cur
		} else {
			sumRatios[i] = pre
		}
	}

	// 如果总和概率分母为零, 那什么都无法选中
	total := sumRatios[len(sumRatios)-1]
	if total == 0 {
		return -1
	}

	// 随机抽奖
	r := N(rand.Int63n(int64(total)))
	return binarySearch(sumRatios, r)
}

func binarySearch[N constraints.Integer](sumRatios []N, r N) int {
	// O(log N) 二分搜索实现
	left, right := 0, len(sumRatios)-1
	result := -1

	for left <= right {
		mid := left + (right-left)/2
		if sumRatios[mid] > r {
			result = mid
			right = mid - 1 // 继续在左半部分寻找更小的满足条件的索引
		} else {
			left = mid + 1 // 在右半部分寻找
		}
	}

	return result // 因为 Rand 的特性和 sumRatios 的构造，这里不会返回 -1
}
