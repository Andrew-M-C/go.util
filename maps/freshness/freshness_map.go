// Package freshness 实现一个新鲜度 map, 新鲜度指的是存取时间的新鲜度, 如果一个值经常被存取,
// 可能永远不会过期
package freshness

import (
	"reflect"
	"sync"
	"time"

	"github.com/Andrew-M-C/go.util/channel"
	timeutil "github.com/Andrew-M-C/go.util/time"
	"golang.org/x/exp/constraints"
)

// Map 实现一个新鲜度 map
type Map[K constraints.Ordered, V any] interface {
	Get(key K) (V, bool)
	Set(key K, value V) (swapped bool)
	GetOrNew(key K) (V, bool, error)
	Close()
}

func NewMap[K constraints.Ordered, V any](opts ...Option) Map[K, V] {
	o := mergeOption(opts)
	m := &mapImpl[K, V]{
		opts:  *o,
		m:     map[K]*valueWithExp[V]{},
		stop:  make(chan struct{}, 1),
		check: make(chan struct{}, 1),
	}
	m.nextCheckTime = timeutil.UpTime() + time.Minute
	go m.doExpire()
	return m
}

type valueWithExp[V any] struct {
	value  V
	expire time.Duration
}

type mapImpl[K constraints.Ordered, V any] struct {
	lock  sync.RWMutex
	opts  options
	m     map[K]*valueWithExp[V]
	stop  chan struct{}
	check chan struct{}

	nextCheckTime time.Duration
}

func (m *mapImpl[K, V]) Get(key K) (res V, exist bool) {
	m.lock.RLock()
	defer m.lock.RUnlock()

	v, exist := m.m[key]
	if !exist {
		return res, false
	}

	exp := timeutil.UpTime() + m.opts.renew
	if v.expire < exp {
		v.expire = exp
		m.requestCheck(exp)
	}
	return v.value, exist
}

func (m *mapImpl[K, V]) Set(key K, value V) bool {
	m.lock.Lock()
	defer m.lock.Unlock()

	_, exist := m.m[key]
	exp := timeutil.UpTime() + m.opts.timeout
	m.m[key] = &valueWithExp[V]{
		value:  value,
		expire: exp,
	}
	m.requestCheck(exp)
	return exist
}

func (m *mapImpl[K, V]) GetOrNew(key K) (res V, exist bool, err error) {
	if v, exist := m.Get(key); exist {
		return v, true, nil
	}

	m.lock.Lock()
	defer m.lock.Unlock()

	if v, exist := m.m[key]; exist {
		exp := timeutil.UpTime() + m.opts.renew
		if v.expire < exp {
			v.expire = exp
			m.requestCheck(exp)
		}
		return v.value, true, nil
	}

	v, err := m.new(key)
	if err != nil {
		return res, false, err
	}

	m.m[key] = v
	m.requestCheck(v.expire)
	return v.value, false, nil
}

func (m *mapImpl[K, V]) Close() {
	_, _ = channel.WriteNonBlocked(m.stop, struct{}{})
}

func (m *mapImpl[K, V]) requestCheck(to time.Duration) {
	if to > m.nextCheckTime {
		return
	}
	_, _ = channel.WriteNonBlocked(m.check, struct{}{})
}

func (m *mapImpl[K, V]) new(key K) (v *valueWithExp[V], err error) {
	expire := timeutil.UpTime() + m.opts.timeout

	if intf := m.opts.newCallback; intf != nil {
		fu, ok := intf.(func(key K) (V, error))
		if ok {
			value, err := fu(key)
			if err != nil {
				return nil, err
			}
			v = &valueWithExp[V]{
				value:  value,
				expire: expire,
			}
			return v, nil
		}
	}

	var temp V
	typ := reflect.TypeOf(temp)
	if typ.Kind() != reflect.Ptr {
		return &valueWithExp[V]{
			value:  temp,
			expire: expire,
		}, nil
	}

	res := reflect.New(typ.Elem())
	value, _ := res.Interface().(V)
	v = &valueWithExp[V]{
		value:  value,
		expire: expire,
	}
	return v, nil
}

func (m *mapImpl[K, V]) doExpire() {
	timer := time.NewTimer(m.nextCheckTime - timeutil.UpTime())

	for shouldExit := false; !shouldExit; {
		select {
		case <-m.stop:
			timer.Stop()
			_, _, _ = channel.ReadNonBlocked(timer.C)
			close(m.check)
			close(m.stop)
			shouldExit = true

		case <-m.check:
			next := m.checkAllTimeoutsAndDo()
			resetTimer(timer, next)
			m.nextCheckTime = next

		case <-timer.C:
			next := m.checkAllTimeoutsAndDo()
			resetTimer(timer, next)
			m.nextCheckTime = next
		}
	}

	var kt K
	var vt V
	m.opts.debug("freshness.Map[%v]%v exit", reflect.TypeOf(kt), reflect.TypeOf(vt))
}

func (m *mapImpl[K, V]) checkAllTimeoutsAndDo() (nextTimeout time.Duration) {
	m.lock.Lock()
	defer m.lock.Unlock()

	nextTimeout = timeutil.UpTime() + time.Minute
	now := timeutil.UpTime()
	var timeoutKeys []K
	var timeoutValues []*valueWithExp[V]

	for k, value := range m.m {
		if value.expire > now {
			if value.expire < nextTimeout {
				nextTimeout = value.expire
			}
			continue
		}
		m.opts.debug("key %v timeout at: %v, now %v", k, value.expire, now)
		timeoutKeys = append(timeoutKeys, k)
		timeoutValues = append(timeoutValues, value)
		delete(m.m, k)
	}

	callback := m.getExpCallback()
	if callback == nil {
		return
	}

	for i, v := range timeoutValues {
		go callback(timeoutKeys[i], v.value)
	}
	return nextTimeout
}

func (m *mapImpl[K, V]) getExpCallback() func(key K, value V) {
	callback := m.opts.expCallback
	if callback == nil {
		return nil
	}
	f, _ := callback.(func(key K, value V))
	return f
}

func resetTimer(timer *time.Timer, to time.Duration) {
	_, _, _ = channel.ReadNonBlocked(timer.C)
	timer.Reset(to - timeutil.UpTime())
}
