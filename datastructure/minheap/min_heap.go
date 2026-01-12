// Package minheap 实现一个最小堆
package minheap

import (
	"container/heap"
	"sync"

	"github.com/Andrew-M-C/go.util/datastructure/constraints"
)

// Basic 表示基于基础可比较类型的最小堆
type Basic[T constraints.Ordered] struct {
	noCopy noCopy

	list *list[T]
	once sync.Once
}

func (b *Basic[T]) init() {
	b.once.Do(func() {
		b.list = &list[T]{
			lessFunc: func(i, j T) bool {
				return i < j
			},
		}
		heap.Init(b.list)
	})
}

func (b *Basic[T]) Len() int {
	b.init()
	return b.list.Len()
}

func (b *Basic[T]) Push(value T) {
	b.init()
	heap.Push(b.list, value)
}

func (b *Basic[T]) Pop() (T, bool) {
	b.init()
	if b.list.Len() == 0 {
		var zero T
		return zero, false
	}
	val, _ := heap.Pop(b.list).(T)
	return val, true
}

type Heap[S constraints.Ordered, T any] struct {
	noCopy noCopy

	list *list[scoredNode[S, T]]
	once sync.Once
}

type scoredNode[S constraints.Ordered, T any] struct {
	score S
	value T
}

func (h *Heap[S, T]) init() {
	h.once.Do(func() {
		h.list = &list[scoredNode[S, T]]{
			lessFunc: func(i, j scoredNode[S, T]) bool {
				return i.score < j.score
			},
		}
		heap.Init(h.list)
	})
}

func (h *Heap[S, T]) Len() int {
	h.init()
	return h.list.Len()
}

func (h *Heap[S, T]) Push(score S, value T) {
	h.init()
	node := scoredNode[S, T]{
		score: score,
		value: value,
	}
	heap.Push(h.list, node)
}

func (h *Heap[S, T]) Pop() (S, T, bool) {
	h.init()
	if h.list.Len() == 0 {
		var zeroS S
		var zeroT T
		return zeroS, zeroT, false
	}
	node, _ := heap.Pop(h.list).(scoredNode[S, T])
	return node.score, node.value, true
}
