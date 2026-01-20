package maps

import (
	"iter"

	rbt "github.com/Andrew-M-C/go.util/datastructure/redblacktree"
	"github.com/Andrew-M-C/go.util/maps/constraints"
)

// OrderedMap 表示一个有序的 map, 底层使用红黑树实现。
// 请注意: 非协程安全
type OrderedMap[K constraints.Ordered, V any] struct {
	t *rbt.Tree[K, V]
}

func NewOrderedMap[K constraints.Ordered, V any]() *OrderedMap[K, V] {
	m := &OrderedMap[K, V]{}
	m.initialize()
	return m
}

func (m *OrderedMap[K, V]) initialize() {
	if m.t != nil {
		return
	}
	m.t = rbt.New[K, V]()
}

// Size 返回 map 的大小
func (m *OrderedMap[K, V]) Size() int {
	m.initialize()
	return m.t.Size()
}

// Len 返回 map 的大小
func (m *OrderedMap[K, V]) Len() int {
	return m.Size()
}

// Get 获取某个 key 的值
func (m *OrderedMap[K, V]) Get(k K) (v V, exist bool) {
	m.initialize()
	return m.t.Get(k)
}

// SetOrSwap 设置某个 key 的值, 如果 key 不存在, 则设置, 如果 key 存在, 则交换
func (m *OrderedMap[K, V]) SetOrSwap(k K, v V) (swapped bool) {
	m.initialize()
	_, swapped = m.t.Get(k)
	m.t.Set(k, v)
	return swapped
}

// Set 设置某个 key 的值
func (m *OrderedMap[K, V]) Set(k K, v V) {
	m.initialize()
	m.t.Set(k, v)
}

// Delete 删除某个 key
func (m *OrderedMap[K, V]) Delete(k K) {
	m.initialize()
	m.t.Remove(k)
}

// Range 遍历 map, 如果 f 返回 false, 则停止遍历
func (m *OrderedMap[K, V]) Range(f func(k K, v V) bool) {
	m.initialize()
	if f == nil {
		return
	}
	m.t.Iterate(func(k K, v V) bool {
		return f(k, v)
	})
}

// All 返回一个迭代器, 用于遍历所有的键值对, 按照 key 的升序排列
// 用法示例:
//
//	for k, v := range m.All() {
//		// 处理 k 和 v
//	}
func (m *OrderedMap[K, V]) All() iter.Seq2[K, V] {
	m.initialize()
	return func(yield func(K, V) bool) {
		m.t.Iterate(func(k K, v V) bool {
			return yield(k, v)
		})
	}
}

// Keys 返回一个迭代器, 用于遍历所有的 key, 按照 key 的升序排列
// 用法示例:
//
//	for k := range m.Keys() {
//		// 处理 k
//	}
func (m *OrderedMap[K, V]) Keys() iter.Seq[K] {
	m.initialize()
	return func(yield func(K) bool) {
		m.t.Iterate(func(k K, v V) bool {
			return yield(k)
		})
	}
}
