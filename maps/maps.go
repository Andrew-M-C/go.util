// Package maps 提供原生 map 类型相关的一些工具
package maps

import (
	"sort"

	"golang.org/x/exp/constraints"
)

// StringKeys 返回所有的 key
func Keys[K comparable, V any](m map[K]V) (keys []K) {
	keys = make([]K, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// KeysDeduplicated 返回所有的 key 并去重
func KeysDeduplicated[K comparable, V any](m map[K]V) (keys []K) {
	keys = make([]K, 0, len(m))
	set := make(map[K]struct{}, len(m))
	for k := range m {
		if _, exist := set[k]; exist {
			continue
		}
		set[k] = struct{}{}
		keys = append(keys, k)
	}
	return keys
}

// KeysSorted 在 Keys 基础上对返回值进行排序
func KeysSorted[K constraints.Ordered, V any](m map[K]V) (keys []K) {
	keys = Keys(m)
	sort.Slice(keys, func(i, j int) bool {
		return keys[i] < keys[j]
	})
	return keys
}

// KeysSortedAndDeduplicated 在 Keys 基础上对返回值进行排序和去重复
func KeysSortedAndDeduplicated[K constraints.Ordered, V any](m map[K]V) (keys []K) {
	keys = KeysDeduplicated(m)
	sort.Slice(keys, func(i, j int) bool {
		return keys[i] < keys[j]
	})
	return keys
}
