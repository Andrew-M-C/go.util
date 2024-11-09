package atomic

import "sync/atomic"

// Value 是 atomic.Value 的泛型化封装
type Value[T any] struct {
	v atomic.Value
}

func (v *Value[T]) Load() T {
	intf := v.v.Load()
	res, _ := intf.(T)
	return res
}

func (v *Value[T]) Store(val T) {
	v.v.Store(val)
}

func (v *Value[T]) Swap(new T) (old T) {
	oldIntf := v.v.Swap(new)
	old, _ = oldIntf.(T)
	return old
}

func (v *Value[T]) CompareAndSwap(old, new T) (swapped bool) {
	return v.v.CompareAndSwap(old, new)
}
