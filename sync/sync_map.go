package sync

import "sync"

// Map 是 sync.Map 的封装, 但是暴露的参数类型是泛型
type Map[K comparable, V comparable] interface {
	Map() map[K]V
	CompareAndDelete(key K, old V) (deleted bool)
	CompareAndSwap(key K, old, new V) bool
	Delete(key K)
	Load(key K) (value V, ok bool)
	LoadAndDelete(key K) (value V, loaded bool)
	LoadOrStore(key K, value V) (actual V, loaded bool)
	Range(f func(key K, value V) bool)
	Store(key K, value V)
	Swap(key K, value V) (previous V, loaded bool)
}

func NewMap[K comparable, V comparable]() Map[K, V] {
	return &syncMap[K, V]{
		m: &sync.Map{},
	}
}

type syncMap[K comparable, V any] struct {
	m *sync.Map
}

func (m *syncMap[K, V]) Map() map[K]V {
	res := make(map[K]V)
	m.Range(func(key K, value V) bool {
		res[key] = value
		return true
	})
	return res
}

func (m *syncMap[K, V]) CompareAndDelete(key K, old V) (deleted bool) {
	return m.m.CompareAndDelete(key, old)
}

func (m *syncMap[K, V]) CompareAndSwap(key K, old, new V) bool {
	return m.m.CompareAndSwap(key, old, new)
}

func (m *syncMap[K, V]) Delete(key K) {
	m.m.Delete(key)
}

func (m *syncMap[K, V]) Load(key K) (value V, ok bool) {
	v, ok := m.m.Load(key)
	if !ok {
		return
	}
	value, ok = v.(V)
	return
}

func (m *syncMap[K, V]) LoadAndDelete(key K) (value V, loaded bool) {
	v, loaded := m.m.LoadAndDelete(key)
	value, _ = v.(V)
	return
}

func (m *syncMap[K, V]) LoadOrStore(key K, value V) (actual V, loaded bool) {
	v, loaded := m.m.LoadOrStore(key, value)
	actual, _ = v.(V)
	return
}

func (m *syncMap[K, V]) Range(f func(key K, value V) bool) {
	m.m.Range(func(key, value any) bool {
		k, _ := key.(K)
		v, _ := value.(V)
		return f(k, v)
	})
}

func (m *syncMap[K, V]) Store(key K, value V) {
	m.m.Store(key, value)
}

func (m *syncMap[K, V]) Swap(key K, value V) (previous V, loaded bool) {
	v, loaded := m.m.Swap(key, value)
	previous, _ = v.(V)
	return
}
