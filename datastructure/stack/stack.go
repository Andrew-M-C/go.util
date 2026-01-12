// Package stack 实现一个栈 (先进后出 FILO 队列)
package stack

type Queue[T any] struct{
	noCopy noCopy

	queue []T
}

func New[T any]() *Queue[T] {
	return &Queue[T]{}
}

func (q *Queue[T]) Len() int {
	return len(q.queue)
}

func (q *Queue[T]) Push(value T) {
	q.queue = append(q.queue, value)
}

func (q *Queue[T]) Pop() (T, bool) {
	if q.Len() == 0 {
		var zero T
		return zero, false
	}
	last := q.queue[q.Len()-1]
	q.queue = q.queue[:q.Len()-1]
	return last, true
}
