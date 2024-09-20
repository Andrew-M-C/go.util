// Package rand 提供官方 math/rand 的一些小工具封装
package rand

import (
	"math/rand"

	"github.com/Andrew-M-C/go.util/constraints"
)

type Number interface {
	constraints.Integer | constraints.Float
}

// 按照给定的比例, 随机选择一个选项。如果一个选项 <= 0 则表示不可能被选中。如果选项列表长度为
// 0 或者所有选项比例的总和 <= 0 则返回 -1。
func IndexByRatios[T Number](ratios []T) int {
	if len(ratios) == 0 {
		return -1
	}

	// 计算各选项
	options := make([]int, 0, len(ratios))
	sums := make([]float64, 0, len(ratios))
	for i, r := range ratios {
		if r <= 0 {
			continue
		}
		options = append(options, i)

		if len(sums) == 0 {
			sums = append(sums, float64(r))
		} else {
			sums = append(sums, float64(r)+sums[len(sums)-1])
		}
	}

	if len(options) == 0 {
		return -1
	}
	if len(options) == 1 {
		return options[0]
	}

	// 随机一个值, 然后匹配看看
	v := rand.Float64()
	v *= sums[len(sums)-1]

	for i, sum := range sums {
		if v < sum {
			return options[i]
		}
	}

	return options[len(options)-1]
}
