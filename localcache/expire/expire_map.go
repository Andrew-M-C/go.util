// Package expire 用于实现超时 map
package expire

import (
	"sync"
	"time"

	"github.com/Andrew-M-C/go.util/channel"
	timeutil "github.com/Andrew-M-C/go.util/time"
	list "github.com/emirpasic/gods/v2/lists/doublylinkedlist"
	rbt "github.com/emirpasic/gods/v2/trees/redblacktree"
)

// Map 实现一个超时即删除的 map。需要注意的是, 暂时还未实现退出机制, 也就是说这个 map
// 会常驻内存。
type Map[K comparable, V any] struct {
	lock sync.RWMutex

	initOK bool

	data    map[K]*valueWithDeadline[V]
	opt     *option
	trigger chan struct{}

	keysByDeadline *rbt.Tree[time.Duration, *list.List[K]]
}

// NewMap 新建一个过期 map
func NewMap[K comparable, V any](opts ...Option) *Map[K, V] {
	opt := defaultOption[K, V]()
	opt = opt.merge(opts)

	m := &Map[K, V]{}
	m.opt = opt
	m.lazyInit()

	return m
}

type valueWithDeadline[V any] struct {
	deadline time.Duration
	value    *V
	opt      *option
}

func (m *Map[K, V]) lazyInit() {
	if m.initOK {
		return
	}
	m.lock.Lock()
	defer m.lock.Unlock()

	if m.initOK {
		return
	}

	m.data = map[K]*valueWithDeadline[V]{}
	m.trigger = make(chan struct{}, 2)
	m.keysByDeadline = rbt.New[time.Duration, *list.List[K]]()

	if m.opt == nil {
		m.opt = defaultOption[K, V]()
	}

	go m.doExpireRoutine()
	m.initOK = true
}

func (m *Map[K, V]) Load(key K) (value *V, exist bool) {
	m.lazyInit()

	m.lock.RLock()
	defer m.lock.RUnlock()

	v, exist := m.data[key]
	if !exist {
		return nil, false
	}
	return v.value, true
}

func (m *Map[K, V]) LoadOrNew(key K, opts ...Option) (value *V, loaded bool) {
	m.lazyInit()

	if res, exist := m.Load(key); exist {
		return res, true
	}

	opt := m.opt.copy().merge(opts)

	m.lock.Lock()
	defer m.lock.Unlock()

	if v, exist := m.data[key]; exist {
		return v.value, true
	}

	newer, _ := opt.callback.newer.(func(K) *V)
	value = newer(key)

	deadline := timeutil.UpTime() + opt.timeout
	v := &valueWithDeadline[V]{
		deadline: deadline,
		value:    value,
		opt:      opt,
	}

	m.store(key, v)
	m.doTrigger()
	return value, false
}

func (m *Map[K, V]) Delete(key K) {
	_, _ = m.Drain(key)
}

func (m *Map[K, V]) Drain(key K) (value *V, drained bool) {
	m.lazyInit()

	m.lock.Lock()
	defer m.lock.Unlock()

	data, exist := m.data[key]
	if !exist {
		return nil, false
	}

	m.delete(key)
	m.doTrigger()
	return data.value, true
}

func (m *Map[K, V]) load(key K) (*valueWithDeadline[V], bool) {
	res, exist := m.data[key]
	return res, exist
}

func (m *Map[K, V]) store(key K, value *valueWithDeadline[V]) {
	// 存入并记录超时树
	m.data[key] = value

	if lst, exist := m.keysByDeadline.Get(value.deadline); exist {
		lst.Add(key)
	} else {
		lst = list.New(key)
		m.keysByDeadline.Put(value.deadline, lst)
	}
}

func (m *Map[K, V]) delete(key K) {
	data, exist := m.data[key]
	if !exist {
		return
	}

	m.opt.debug("删除 key '%v'", key)
	delete(m.data, key)

	lst, exist := m.keysByDeadline.Get(data.deadline)
	if !exist {
		return
	}

	lst.Remove(lst.IndexOf(key))
	if s := lst.Size(); s == 0 {
		m.opt.debug("deadline %v 上已没有 key, 删除之", key)
		m.keysByDeadline.Remove(data.deadline)
	} else {
		m.opt.debug("deadline %v 上剩余 %d 个 key, 保留", key, s)
	}
}

func (m *Map[K, V]) Store(key K, value *V, opts ...Option) {
	_, _ = m.Swap(key, value, opts...)
}

func (m *Map[K, V]) Swap(key K, newValue *V, opts ...Option) (oldValue *V, swapped bool) {
	m.lazyInit()
	opt := m.opt.copy().merge(opts)

	deadline := timeutil.UpTime() + opt.timeout
	v := &valueWithDeadline[V]{
		deadline: deadline,
		value:    newValue,
		opt:      opt,
	}

	m.lock.Lock()
	defer m.lock.Unlock()

	oldOne, exist := m.load(key)
	if exist {
		m.opt.debug("删除旧 key 值 '%v'", key)
		oldValue = oldOne.value
		m.delete(key)
	}

	m.store(key, v)
	m.doTrigger()
	return oldValue, true
}

func (m *Map[K, V]) doExpireRoutine() {
	tm := time.NewTimer(m.opt.timeout)

	updateNextTime := func() {
		m.lock.RLock()
		defer m.lock.RUnlock()
		first := m.keysByDeadline.Left()
		if first == nil {
			tm.Reset(time.Hour)
			return
		}
		next := first.Key - timeutil.UpTime()
		m.opt.debug("下一次过期时间: %v", next)
		tm.Reset(next)
	}

	for {
		select {
		case <-tm.C:
		case <-m.trigger:
		}
		m.checkExpire()
		updateNextTime()
	}
}

func (m *Map[K, V]) doTrigger() {
	_, _ = channel.WriteNonBlocked(m.trigger, struct{}{})
}

type kv[K comparable, V any] struct {
	Key   K
	Value V
}

func (m *Map[K, V]) checkExpire() {
	m.lock.Lock()
	defer m.lock.Unlock()

	var expireKVs []kv[K, *valueWithDeadline[V]]

	// 寻找过期节点
	for {
		now := timeutil.UpTime()
		left := m.keysByDeadline.Left()
		if left == nil {
			m.opt.debug("待过期列表为空")
			break
		}
		if now < left.Key {
			m.opt.debug("当前 uptime %v, 最近的一个 deadline %v, 暂时没有需要过期的数据, 忽略之", now, left.Key)
			break
		}

		for it := left.Value.Iterator(); it.Next(); {
			k := it.Value()

			v, exist := m.data[k]
			if !exist {
				continue
			}

			expireKVs = append(expireKVs, kv[K, *valueWithDeadline[V]]{
				Key:   k,
				Value: v,
			})
		}

		m.keysByDeadline.Remove(left.Key)
	}

	if len(expireKVs) > 0 {
		m.detachTimeoutCallback(expireKVs)
	}
}

func (m *Map[K, V]) detachTimeoutCallback(kvs []kv[K, *valueWithDeadline[V]]) {
	for _, kv := range kvs {
		v := kv.Value.opt.callback.timeout
		if v == nil {
			m.opt.debug("没有超时回调, 无需调用 %d 个数据", len(kvs))
		}
		callback, _ := v.(func(K, *V))
		go callback(kv.Key, kv.Value.value)
	}
}
