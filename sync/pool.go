package sync

import "sync"

// Pool 封装官方 sync.Pool, 但范型化
type Pool[T any] interface {
	Get() T
	Put(T)
}

// PoolNewer 用在 Pool 类型中
type PoolNewer[T any] func() T

// NewPool 新建一个泛型化的 sync.Pool
func NewPool[T any](newer PoolNewer[T]) Pool[T] {
	return &pool[T]{
		pool: &sync.Pool{
			New: func() any { return newer() },
		},
	}
}

type pool[T any] struct {
	pool *sync.Pool
}

func (p *pool[T]) Get() T {
	v := p.pool.Get()
	res, _ := v.(T)
	return res
}

func (p *pool[T]) Put(obj T) {
	p.pool.Put(obj)
}
