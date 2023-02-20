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

// Equal 判断两个 map 是否完全一致
func Equal[K comparable, V comparable](a, b map[K]V) bool {
	if len(a) != len(b) {
		return false
	}
	for k, va := range a {
		vb, exist := b[k]
		if !exist {
			return false
		}
		if va != vb {
			return false
		}
	}
	return true
}

// KeysEqual 判断两个 map 是否拥有相同的 keys
func KeysEqual[K comparable, V1, V2 any](a map[K]V1, b map[K]V2) bool {
	if len(a) != len(b) {
		return false
	}
	for k := range a {
		if _, exist := b[k]; !exist {
			return false
		}
	}
	return true
}
