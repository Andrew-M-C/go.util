package maps

import (
	"encoding/json"
	"sync"
)

// RWSafeMap 即 Read-Write-Locked Safe Map, 表示这是一个使用读写锁锁住的安全的 map 类型。
// 适用于读多写少的情况。
type RWSafeMap[K comparable, V any] interface {
	json.Marshaler
	json.Unmarshaler

	Store(k K, v V)
	Load(k K) (V, bool)
	Delete(k K)

	LoadAndDelete(k K) (value V, loaded bool)
	LoadOrStore(k K, v V) (actual V, loaded bool)
	LoadOrNew(k K, newFunc func() V) (actual V, loaded bool)

	Swap(k K, v V) (previous V, loaded bool)
	Range(f func(key K, value V) bool)

	Size() int
}

// NewRWSafeMap 新建一个 RWSafeMap 实例。可选参数只有一个, 表示 capacity
func NewRWSafeMap[K comparable, V any](capacity ...int) RWSafeMap[K, V] {
	cap := 0
	if len(capacity) > 0 && capacity[0] > 0 {
		cap = capacity[0]
	}

	m := &rwSafeMapImpl[K, V]{
		ma: make(map[K]V, cap),
	}

	return m
}

type rwSafeMapImpl[K comparable, V any] struct {
	ma map[K]V
	mu sync.RWMutex
}

func (m *rwSafeMapImpl[K, V]) doReading(f func()) {
	m.mu.RLock()
	f()
	m.mu.RUnlock()
}

func (m *rwSafeMapImpl[K, V]) doWriting(f func()) {
	m.mu.Lock()
	f()
	m.mu.Unlock()
}

func (m *rwSafeMapImpl[K, V]) Store(k K, v V) {
	m.doWriting(func() {
		m.ma[k] = v
	})
}

func (m *rwSafeMapImpl[K, V]) Load(k K) (value V, exist bool) {
	m.doReading(func() {
		value, exist = m.ma[k]
	})
	return
}

func (m *rwSafeMapImpl[K, V]) Delete(k K) {
	m.doWriting(func() {
		delete(m.ma, k)
	})
}

func (m *rwSafeMapImpl[K, V]) LoadAndDelete(k K) (value V, loaded bool) {
	m.doWriting(func() {
		value, loaded = m.ma[k]
		if loaded {
			delete(m.ma, k)
		}
	})
	return
}

func (m *rwSafeMapImpl[K, V]) LoadOrStore(k K, v V) (actual V, loaded bool) {
	m.doWriting(func() {
		actual, loaded = m.ma[k]
		if !loaded {
			actual = v
			m.ma[k] = v
		}
	})
	return
}

func (m *rwSafeMapImpl[K, V]) LoadOrNew(k K, newFunc func() V) (actual V, loaded bool) {
	if newFunc == nil {
		return m.Load(k)
	}

	m.doWriting(func() {
		actual, loaded = m.ma[k]
		if !loaded {
			actual = newFunc()
			m.ma[k] = actual
		}
	})
	return
}

func (m *rwSafeMapImpl[K, V]) Swap(k K, v V) (previous V, loaded bool) {
	m.doWriting(func() {
		previous, loaded = m.ma[k]
		m.ma[k] = v
	})
	return
}

func (m *rwSafeMapImpl[K, V]) Size() (size int) {
	m.doReading(func() {
		size = len(m.ma)
	})
	return
}

func (m *rwSafeMapImpl[K, V]) MarshalJSON() (b []byte, err error) {
	m.doReading(func() {
		b, err = json.Marshal(m.ma)
	})
	return
}

func (m *rwSafeMapImpl[K, V]) UnmarshalJSON(b []byte) error {
	var newMap map[K]V
	if err := json.Unmarshal(b, &newMap); err != nil {
		return err
	}
	m.doWriting(func() {
		m.ma = newMap
	})
	return nil
}

func (m *rwSafeMapImpl[K, V]) Range(f func(key K, value V) bool) {
	var kvList []KVPair[K, V]
	m.doReading(func() {
		kvList = KeyValues(m.ma)
	})

	for _, kv := range kvList {
		if !f(kv.K, kv.V) {
			return
		}
	}
}
