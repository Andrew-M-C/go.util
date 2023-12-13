package expire

import "time"

// Option 表示额外配置
type Option[K comparable, V any] func(*options[K, V])

type options[K comparable, V any] struct {
	newer    ExpireMapNewer[K, V]
	notifier ExpireMapNotifier[K, V]
}

// WithNewer 覆盖默认的初始化器
func WithNewer[K comparable, V any](newer ExpireMapNewer[K, V]) Option[K, V] {
	return func(o *options[K, V]) {
		o.newer = newer
	}
}

// WithExpireNotifier 覆盖默认的通知器
func WithExpireNotifier[K comparable, V any](notifier ExpireMapNotifier[K, V]) Option[K, V] {
	return func(o *options[K, V]) {
		o.notifier = notifier
	}
}

func (m *ExpireMap[K, V]) combineOptions(opt []Option[K, V]) *options[K, V] {
	o := &options[K, V]{}
	for _, f := range opt {
		if f != nil {
			f(o)
		}
	}

	if o.newer == nil {
		o.newer = m.defaultNewer
	}
	if o.newer == nil {
		o.newer = func(k K) (*V, time.Duration) {
			return new(V), m.getDefaultExpire()
		}
	}

	return o
}

func (m *ExpireMap[K, V]) getDefaultExpire() time.Duration {
	if ptr := m.defaultTTL; ptr != nil {
		return *ptr
	}
	return DefaultExpire
}
