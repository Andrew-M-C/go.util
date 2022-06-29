// Package maps 提供原生 map 类型相关的一些工具
package maps

import (
	"reflect"
	"sort"
)

// StringKeys 简单返回某个 key 是 string 的 map 的所有 key
func StringKeys(m interface{}) (keys []string) {
	v := reflect.ValueOf(m)
	if v.Kind() != reflect.Map {
		return nil
	}
	if v.Type().Key().Kind() != reflect.String {
		return nil
	}

	vkeys := v.MapKeys()
	for _, k := range vkeys {
		keys = append(keys, k.String())
	}
	return keys
}

// StringKeysSorted 在 StringKeys 基础上对返回值进行排序
func StringKeysSorted(m interface{}) (keys []string) {
	keys = StringKeys(m)
	sort.Strings(keys)
	return keys
}

// IntKeys 简单返回某个 key 是有符号数的 map 的所有 key
func IntKeys(m interface{}) (keys []int64) {
	v := reflect.ValueOf(m)
	if v.Kind() != reflect.Map {
		return nil
	}
	if v.Type().Key().Kind() != reflect.Int {
		return nil
	}

	vkeys := v.MapKeys()
	for _, k := range vkeys {
		keys = append(keys, k.Int())
	}
	return keys
}

// IntKeysSorted 在 IntKeys 基础上对返回值进行排序
func IntKeysSorted(m interface{}) (keys []int64) {
	keys = IntKeys(m)
	sort.Slice(keys, func(i, j int) bool {
		return keys[i] <= keys[j]
	})
	return keys
}

// UintKeys 简单返回某个 key 是无符号数的 map 的所有 key
func UintKeys(m interface{}) (keys []uint64) {
	v := reflect.ValueOf(m)
	if v.Kind() != reflect.Map {
		return nil
	}
	if v.Type().Key().Kind() != reflect.Uint {
		return nil
	}

	vkeys := v.MapKeys()
	for _, k := range vkeys {
		keys = append(keys, k.Uint())
	}
	return keys
}

// UintKeysSorted 在 UintKeys 基础上对返回值进行排序
func UintKeysSorted(m interface{}) (keys []uint64) {
	keys = UintKeys(m)
	sort.Slice(keys, func(i, j int) bool {
		return keys[i] <= keys[j]
	})
	return keys
}
