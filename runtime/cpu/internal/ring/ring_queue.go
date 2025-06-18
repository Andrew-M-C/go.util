package ring

// Queue 实现一个简易的环形队列
type Queue[T any] struct {
	itemCount int
	buffer    []T
	nextIndex int
}

// NewRingQueue 创建一个环形队列
func NewRingQueue[T any](capacity int) *Queue[T] {
	if capacity <= 0 {
		panic("capacity should be greater than 0")
	}
	q := &Queue[T]{
		itemCount: 0,
		buffer:    make([]T, capacity),
		nextIndex: 0,
	}
	return q
}

// Len 返回队列中的数据数量
func (q *Queue[T]) Len() int {
	return q.itemCount
}

// Capacity 返回队列总大小
func (q *Queue[T]) Capacity() int {
	return len(q.buffer)
}

// Push 往尾部添加一个值
func (q *Queue[T]) Push(v T) {
	q.buffer[q.nextIndex] = v
	if q.itemCount < len(q.buffer) {
		q.itemCount++
	}

	q.nextIndex++
	if q.nextIndex >= len(q.buffer) {
		q.nextIndex = 0
	}
}

// Clear 清除缓存中的所有数据
func (q *Queue[T]) Clear() {
	q.itemCount = 0
	q.nextIndex = 0
}

// GetAllValues 读取所有的值, 从新到旧排序, 也就是说 value[0] 是最新的
func (q *Queue[T]) GetAllValues() (values []T) {
	values = make([]T, 0, q.itemCount)
	for i := q.nextIndex - 1; i >= 0; i-- {
		values = append(values, q.buffer[i])
	}
	if q.itemCount < len(q.buffer) {
		return
	}
	for i := len(q.buffer) - 1; i >= q.nextIndex; i-- {
		values = append(values, q.buffer[i])
	}
	return
}
