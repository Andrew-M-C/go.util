// Package maps 提供原生 map 类型相关的一些工具
package maps

import (
	"sort"

	"golang.org/x/exp/constraints"
)

type KeyList[K constraints.Ordered] []K

// SortAsc 返回一个升序后的 key 列表
func (l KeyList[K]) SortAsc() KeyList[K] {
	res := make([]K, len(l))
	copy(res, l)
	sort.Slice(res, func(i, j int) bool {
		return res[i] < res[j]
	})
	return res
}

// SortDesc 返回一个降序后的 key 列表
func (l KeyList[K]) SortDesc() KeyList[K] {
	res := make(KeyList[K], len(l))
	copy(res, l)
	sort.Slice(res, func(i, j int) bool {
		return res[i] > res[j]
	})
	return res
}

// Deduplicate 去重
func (l KeyList[K]) Deduplicate() KeyList[K] {
	keys := make(KeyList[K], 0, len(l))
	set := NewSetWithCapacity[K](len(l))
	for _, k := range l {
		if set.Has(k) {
			continue
		}
		set.Add(k)
		keys = append(keys, k)
	}
	return keys
}

// StringKeys 返回所有的 key
func Keys[K constraints.Ordered, V any](m map[K]V) (keys KeyList[K]) {
	keys = make([]K, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
