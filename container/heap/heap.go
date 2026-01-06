package heap

import (
	"github.com/Andrew-M-C/go.util/constraints"
	"github.com/emirpasic/gods/trees/binaryheap"
)

// Heap 实现一个堆, 必须使用 New 方法创建, 否则会 panic
type Heap[T any] struct {
	heap     *binaryheap.Heap
	lessFunc func(i, j T) bool
}

// NewBasic 创建一个最小堆, 使用 constraints.Ordered 约束
func NewBasic[T constraints.Ordered]() *Heap[T] {
	return New(func(i, j T) bool {
		return i < j
	})
}

// New 创建一个堆。lessFunc 用于比较两个元素的大小, 如果为空则 panic
func New[T any](lessFunc func(i, j T) bool) *Heap[T] {
	if lessFunc == nil {
		panic("lessFunc is nil")
	}
	
	h := &Heap[T]{
		lessFunc: lessFunc,
	}
	
	// 将 lessFunc 转换为 gods 库要求的 comparator
	// comparator 要求: a < b 返回负数, a == b 返回 0, a > b 返回正数
	comparator := func(a, b any) int {
		aVal := a.(T)
		bVal := b.(T)
		if lessFunc(aVal, bVal) {
			return -1 // a < b
		} else if lessFunc(bVal, aVal) {
			return 1 // a > b
		}
		return 0 // a == b
	}
	
	h.heap = binaryheap.NewWith(comparator)
	return h
}

func (h *Heap[T]) Len() int {
	return h.heap.Size()
}

func (h *Heap[T]) Push(x T) {
	h.heap.Push(x)
}

func (h *Heap[T]) Pop() T {
	val, ok := h.heap.Pop()
	if !ok {
		var zero T
		return zero
	}
	res, _ := val.(T)
	return res
}






