package json

import (
	"encoding/json"
	"sort"
)

// KeyableType 表示对于 go json 来说, 可以系列化为 object 的 key 的类型
type KeyableType interface {
	~string |
		~int8 | ~uint8 |
		~int16 | ~uint16 |
		~int32 | ~uint32 |
		~int64 | ~uint64 |
		~int | ~uint |
		~uintptr
}

// Map 表示各种基础和派生的 map 类型
type Map[K KeyableType, V any] interface {
	~map[K]V
}

// BoolValuedMap 表示各种基础和派生的 map[K]bool 理性
type BoolValuedMap[K KeyableType] interface {
	~map[K]bool
}

// Collection 特指 value 类型为 struct{} 的集合类型
type Collection[K KeyableType] interface {
	~map[K]struct{}
}

// MarshalMapKeyToArray 将一个 map 类型序列化为一个数组
func MarshalMapKeyToArray[K KeyableType, V any, M Map[K, V]](o M) ([]byte, error) {
	keys := make([]K, 0, len(o))
	for k := range o {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		return keys[i] < keys[j]
	})
	return json.Marshal(keys)
}

// MarshalBoolMapKeyToArray 将一个 map 类型序列化为一个数组, 如果 value 为 false 的话, 不写入数组
func MarshalBoolMapKeyToArray[K KeyableType, M BoolValuedMap[K]](o M) ([]byte, error) {
	keys := make([]K, 0, len(o))
	for k, b := range o {
		if !b {
			continue
		}
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		return keys[i] < keys[j]
	})
	return json.Marshal(keys)
}

// UnmarshalArrayToCollection 将一个数组反序列化为一个 map[xxx]struct{} 类型
func UnmarshalArrayToCollection[K KeyableType, M Collection[K]](b []byte, tgt *M) error {
	var keys []K
	if err := json.Unmarshal(b, &keys); err != nil {
		return err
	}
	if *tgt == nil {
		*tgt = make(map[K]struct{}, len(keys))
	}
	for _, k := range keys {
		(*tgt)[k] = struct{}{}
	}
	return nil
}

// UnmarshalArrayToBoolMap 将一个数组反序列化为一个 map[xxx]bool 类型, 所有的 value 均为 true
func UnmarshalArrayToBoolMap[K KeyableType, M BoolValuedMap[K]](b []byte, tgt *M) error {
	var keys []K
	if err := json.Unmarshal(b, &keys); err != nil {
		return err
	}
	if *tgt == nil {
		*tgt = make(map[K]bool, len(keys))
	}
	for _, k := range keys {
		(*tgt)[k] = true
	}
	return nil
}
