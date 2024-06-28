// Package cache 定义一些内存缓存, 但同时带各种特殊的过期逻辑。实验性工具
package localcache

import (
	"sync"
	"time"
)

type cache[K comparable, V any] struct {
	lock  sync.Mutex
	debug func(string, ...any)

	defaultOption options
}

func (c *cache[K, V]) initialize(timeout time.Duration, opts []Option) error {
	o := defaultOptions[V]()
	opts = append(opts, WithExpireTimeout(timeout))
	o, err := mergeOptions[V](o, opts)
	if err != nil {
		return err
	}
	c.defaultOption = *o
	c.debug = func(string, ...any) {}
	return nil
}

func (c *cache[_, V]) invokeExpire(v *value[V]) {
	if f, _ := v.Opts.callback.expire.(func(*V)); f != nil {
		f(v.Value)
	}
	if f, _ := c.defaultOption.callback.expire.(func(*V)); f != nil {
		f(v.Value)
	}
}

type value[V any] struct {
	Value *V
	Opts  options
}

type node[K comparable, V any] struct {
	Key   K
	Value *value[V]
	Next  *node[K, V]
}
