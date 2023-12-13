// Package expire 实现超时 map
package expire

import (
	"errors"
	"sync"
	"time"

	"github.com/Andrew-M-C/go.util/channel"
	"github.com/Andrew-M-C/go.util/govet"
)

const (
	// DefaultExpire 全局默认超时时间, 如果不设置的话, ExpireMap 会使用这个超时时间
	DefaultExpire = 10 * time.Second
)

// ExpireMapNewer ExpireMap 值初始化器
type ExpireMapNewer[K comparable, V any] func(K) (*V, time.Duration)

// ExpireMapNotifier ExpireMap 超时通知器
type ExpireMapNotifier[K comparable, V any] func(K, *V)

// ExpireMap 表示一个定时删除的 map
type ExpireMap[K comparable, V any] struct {
	govet.NoCopy

	storage map[K]*expireMapItem[K, V]
	lock    sync.RWMutex

	defaultTTL      *time.Duration
	defaultNewer    ExpireMapNewer[K, V]
	defaultNotifier ExpireMapNotifier[K, V]
}

type expireMapItem[K comparable, V any] struct {
	ttl     *time.Duration
	access  chan struct{}
	payload *V
	notify  ExpireMapNotifier[K, V]
}

// SetDefaultNewer 设置默认初始化器
func (m *ExpireMap[K, V]) SetDefaultNewer(newer ExpireMapNewer[K, V]) {
	m.defaultNewer = newer
}

// SetDefaultExpire 设置默认超超时时间, 必须大于零。仅当 default newer 为空时生效
func (m *ExpireMap[K, V]) SetDefaultExpire(expire time.Duration) {
	if expire <= 0 {
		return
	}
	m.defaultTTL = &expire
}

// SetDefaultExpireNotifier 设置默认的超时通知
func (m *ExpireMap[K, V]) SetDefaultExpireNotifier(notifier ExpireMapNotifier[K, V]) {
	m.defaultNotifier = notifier
}

// LoadOrNew 加载或更新
func (m *ExpireMap[K, V]) LoadOrNew(key K, opts ...Option[K, V]) (value *V, loaded bool) {
	o := m.combineOptions(opts)

	m.lock.RLock()
	item, loaded := m.storage[key]
	m.lock.RUnlock()

	if loaded {
		_, _ = channel.WriteNonBlocked(item.access, struct{}{})
		return item.payload, true
	}

	m.lock.Lock()
	defer m.lock.Unlock()

	item, loaded = m.storage[key]
	if loaded {
		_, _ = channel.WriteNonBlocked(item.access, struct{}{})
		return item.payload, true
	}
	value, ttl := o.newer(key)
	if ttl <= 0 {
		panic(errors.New("expire time should greater than zero"))
	}

	item = &expireMapItem[K, V]{
		ttl:     &ttl,
		access:  make(chan struct{}, 1),
		payload: value,
		notify:  o.notifier,
	}
	if m.storage == nil {
		m.storage = map[K]*expireMapItem[K, V]{}
	}
	m.storage[key] = item

	go m.watchItem(key, item)
	return value, false
}

// ResetExpiration 重新设置超市时间, 从当前时间重新算起
func (m *ExpireMap[K, V]) ResetExpiration(key K, tm time.Duration) (updated bool) {
	if tm <= 0 {
		tm = 0
	}

	m.lock.RLock()
	item, updated := m.storage[key]
	m.lock.RUnlock()

	if !updated {
		return false
	}
	item.ttl = &tm
	_, _ = channel.WriteNonBlocked(item.access, struct{}{})

	return true
}

func (m *ExpireMap[K, V]) watchItem(key K, item *expireMapItem[K, V]) {
	timer := time.NewTimer(*item.ttl)
	timeout := false

	for !timeout {
		select {
		case <-timer.C:
			timeout = true

		case <-item.access:
			if !timer.Stop() {
				<-timer.C
			}
			timer.Reset(*item.ttl)
		}
	}

	m.lock.Lock()
	delete(m.storage, key)
	m.lock.Unlock()

	close(item.access)
	if item.notify != nil {
		item.notify(key, item.payload)
	}
}
