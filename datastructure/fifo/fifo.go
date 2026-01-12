// Package fifo 实现一个先入先出队列
package fifo

type Queue[T any] struct{
	noCopy noCopy

	head *node[T]
	tail *node[T]
	size int
}

type node[T any] struct {
	value T
	next  *node[T]
}

// New 新建一个 FIFO 队列
func New[T any](capacity int) *Queue[T] {
	return &Queue[T]{}
}

func (q *Queue[T]) Len() int {
	return q.size
}

func (q *Queue[T]) Push(value T) {
	newNode := &node[T]{
		value: value,
		next:  nil,
	}
	if q.head == nil {
		q.head = newNode
		q.tail = newNode
		q.size = 1
		return
	}
	// else 
	q.tail.next = newNode
	q.tail = newNode
	q.size++
}

func (q *Queue[T]) Pop() (T, bool) {
	if q.head == nil {
		var zero T
		return zero, false
	}
	head := q.head
	q.head = head.next
	q.size--
	
	// 当队列为空时，重置 tail 指针
	if q.head == nil {
		q.tail = nil
	}
	
	return head.value, true
}

