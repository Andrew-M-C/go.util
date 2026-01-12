package minheap

type noCopy struct{}

func (*noCopy) Lock()   {}
func (*noCopy) Unlock() {}

type list[T any] struct {
	data     []T
	lessFunc func(i, j T) bool
}

func (l *list[T]) Len() int {
	return len(l.data)
}

func (l *list[T]) Less(i, j int) bool {
	return l.lessFunc(l.data[i], l.data[j])
}

func (l *list[T]) Swap(i, j int) {
	l.data[i], l.data[j] = l.data[j], l.data[i]
}

func (l *list[T]) Push(x any) {
	val, _ := x.(T)
	l.data = append(l.data, val)
}

func (l *list[T]) Pop() any {
	old := l.data
	n := len(old)
	x := old[n-1]
	l.data = old[0 : n-1]
	return x
}
