package localcache

import (
	"cmp"
	"time"

	timeutil "github.com/Andrew-M-C/go.util/time"
	rbt "github.com/emirpasic/gods/v2/trees/redblacktree"
)

type uptime = time.Duration

// ExpireMap 表示新鲜度 map, 每一次存取对应的值, 会延长新鲜度
type ExpireMap[K comparable, V any] struct {
	cache[K, V]

	m map[K]*value[V]

	expires *rbt.Tree[uptime, *node[K, V]]
	trigger chan struct{}
}

// NewExpireMap 新建一个 ExpireMap
func NewExpireMap[K cmp.Ordered, V any](timeout time.Duration, opts ...Option) (*ExpireMap[K, V], error) {
	m := &ExpireMap[K, V]{}
	if err := m.initialize(timeout, opts); err != nil {
		return nil, err
	}

	m.m = map[K]*value[V]{}
	m.expires = rbt.New[uptime, *node[K, V]]()
	m.trigger = make(chan struct{}, 2)

	go m.doExpire()
	return m, nil
}

// MARK: 超时逻辑

func (m *ExpireMap[K, V]) doExpire() {
	tm := time.NewTimer(m.defaultOption.timing.timeout)

	updateNextTime := func() {
		first := m.expires.Left()
		if first == nil {
			tm.Reset(time.Hour)
			return
		}
		next := timeutil.UpTime() - first.Key
		m.debug("下一次过期时间: %v", next)
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

func (m *ExpireMap[K, V]) checkExpire() {
	m.lock.Lock()
	defer m.lock.Unlock()

	var expireKeys []time.Duration
	var expireNodeLists []*node[K, V]

	// 寻找过期节点
	for it := m.expires.Iterator(); it.Next(); {
		now := timeutil.UpTime()
		tm := it.Key()

		if now < tm {
			m.debug("当前没有过期值")
			break
		}

		m.debug("当前 up time %v, 目标时间 %v 已过期", now, tm)
		expireKeys = append(expireKeys, tm)
		expireNodeLists = append(expireNodeLists, it.Value())
	}

	// 删除过期节点
	for _, k := range expireKeys {
		m.expires.Remove(k)
	}
	// 执行过期回调
	if len(expireNodeLists) == 0 {
		return
	}
	go func() {
		for _, nodes := range expireNodeLists {
			for n := nodes; n != nil; n = n.Next {
				m.invokeExpire(n.Value)
			}
		}
	}()
}

// MARK: 基础能力

// Load 加载, 不存在的话就不初始化
func (m *ExpireMap[K, V]) Load(key K) (res *V, exist bool) {
	m.lock.Lock()
	defer m.lock.Unlock()

	v, exist := m.m[key]
	if !exist {
		return nil, false
	}
	return v.Value, true
}

// LoadOrNew 加载或新建
func (m *ExpireMap[K, V]) LoadOrNew(key K, opt ...Option) (res *V, loaded bool) {
	m.lock.Lock()
	defer m.lock.Unlock()

	v, exist := m.m[key]
	if exist {
		return v.Value, true
	}

	o := m.defaultOption
	opts, _ := mergeOptions[V](&o, opt)
	newer, _ := opts.callback.newer.(func() *V)
	res = newer()

	v = &value[V]{
		Value: res,
	}
	v.Value = res
	v.Opts = *opts

}
