package maps

import (
	"sort"

	"golang.org/x/exp/constraints"
)

// SortOrder 排序方向, 用于 KVPair
type SortOrder bool

const (
	// Ascend 升序
	Ascend SortOrder = false
	// Descend 降序
	Descend SortOrder = true
)

// KVPair 表示键值对, 便于调用方排序用。
type KVPair[K comparable, V any] struct {
	K K
	V V
}

// KeyValues 获取键值对列表
func KeyValues[K comparable, V any](m map[K]V) []KVPair[K, V] {
	res := make([]KVPair[K, V], 0, len(m))
	for k, v := range m {
		res = append(res, KVPair[K, V]{
			K: k,
			V: v,
		})
	}
	return res
}

// KeyValuesAndSortByKeys 获取键值对列表, 并且按照 key 排序
func KeyValuesAndSortByKeys[K constraints.Ordered, V any](m map[K]V, sortOrder SortOrder) []KVPair[K, V] {
	kvs := KeyValues(m)
	if sortOrder == Ascend {
		sort.Slice(kvs, func(i, j int) bool {
			return kvs[i].K <= kvs[j].K
		})
	} else {
		sort.Slice(kvs, func(i, j int) bool {
			return kvs[i].K >= kvs[j].K
		})
	}
	return kvs
}

// KeyValuesAndSortByValues 获取键值对列表, 并且按照 key 排序
func KeyValuesAndSortByValues[K comparable, V constraints.Ordered](m map[K]V, sortOrder SortOrder) []KVPair[K, V] {
	kvs := KeyValues(m)
	if sortOrder == Ascend {
		sort.Slice(kvs, func(i, j int) bool {
			return kvs[i].V <= kvs[j].V
		})
	} else {
		sort.Slice(kvs, func(i, j int) bool {
			return kvs[i].V >= kvs[j].V
		})
	}
	return kvs
}
