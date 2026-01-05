package heap

import "container/heap"

// Heap 实现一个堆, 必须使用 New 方法创建, 否则会 panic
type Heap[T any] struct {
	data *container[T]
}

// New 创建一个堆。lessFunc 用于比较两个元素的大小, 如果为空则 panic
func New[T any](lessFunc func(i, j T) bool) *Heap[T] {
	if lessFunc == nil {
		panic("lessFunc is nil")
	}
	h := &Heap[T]{
		data: &container[T]{
			lessFunc: lessFunc,
		},
	}
	heap.Init(h.data)
	return h
}

func (h *Heap[T]) Len() int {
	return h.data.Len()
}

func (h *Heap[T]) Push(x T) {
	heap.Push(h.data, x)
}

func (h *Heap[T]) Pop() T {
	v := heap.Pop(h.data)
	res, _ := v.(T)
	return res
}

// ---- 内部实现 ----

type container[T any] struct{
	data []T
	lessFunc func(i, j T) bool
}

func (c container[T]) Len() int {
	return len(c.data)
}

func (c container[T]) Less(i, j int) bool {
	return c.lessFunc(c.data[i], c.data[j])
}

func (c *container[T]) Swap(i, j int) {
	c.data[i], c.data[j] = c.data[j], c.data[i]
}

func (c *container[T]) Push(x any) {
	v, _ := x.(T)
	c.data = append(c.data, v)
}

func (c *container[T]) Pop() any {
	n := len(c.data)
	if n == 0 {
		return nil
	}
	res := c.data[n-1]
	c.data = c.data[:n-1]
	return res
}






