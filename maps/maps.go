// Package maps 提供原生 map 类型相关的一些工具
package maps

import (
	"fmt"

	"github.com/Andrew-M-C/go.util/slice"
	"golang.org/x/exp/constraints"
)

// Keys 返回所有的 key
func Keys[K constraints.Ordered, V any](m map[K]V) (keys slice.List[K]) {
	keys = make([]K, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// Values 返回所有的 value
func Values[K constraints.Ordered, V any](m map[K]V) (values []V) {
	values = make([]V, 0, len(m))
	for _, v := range m {
		values = append(values, v)
	}
	return values
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

// GetOrDefault 从 map 中获取数据, 如果不存在则返回 default value
func GetOrDefault[K constraints.Ordered, V any](m map[K]V, key K, defaultValue V) V {
	if v, exist := m[key]; exist {
		return v
	}
	return defaultValue
}

// GetStringOrFormat 从 value 为 string 的 map 中获取数据, 如果不存在则使用 fmt.Sprintf(format, key)
// 的测试返回 value
func GetStringOrFormat[K constraints.Ordered, V ~string](m map[K]V, key K, format string) V {
	if v, exist := m[key]; exist {
		return v
	}
	return V(fmt.Sprintf(format, key))
}

// Clone 浅复制一个 map
func Clone[K constraints.Ordered, V any](m map[K]V) map[K]V {
	res := make(map[K]V, len(m))
	for k, v := range m {
		res[k] = v
	}
	return res
}

// GetOne 表示随便从 map 中取一个 KV 对
func GetOne[K constraints.Ordered, V any](m map[K]V) (key K, value V, exist bool) {
	for k, v := range m {
		return k, v, true
	}
	return
}
