package localcache

import (
	"errors"
	"time"
)

// Option 表示额外参数
type Option func(*options)

type options struct {
	callback struct {
		newer  any
		expire any
	}
	timing struct {
		timeout time.Duration
		renew   time.Duration
	}
}

func defaultOptions[V any]() *options {
	opt := &options{}
	opt.callback.newer = func() *V {
		return new(V)
	}
	opt.callback.expire = func(*V) {}
	return opt
}

func mergeOptions[V any](to *options, opts []Option) (*options, error) {
	// 自定义选项
	for _, o := range opts {
		if o != nil {
			o(to)
		}
	}

	// 合法性检查
	if to.timing.timeout <= 0 {
		return to, errors.New("没有指定有效的超时时间")
	}
	if to.timing.renew <= 0 {
		to.timing.renew = to.timing.timeout
	}
	if fu, _ := to.callback.newer.(func() *V); fu == nil {
		return to, errors.New("值初始化函数类型不合法")
	}
	if fu, _ := to.callback.expire.(func(*V)); fu == nil {
		return to, errors.New("超时回调函数类型不合法")
	}

	// 检查 OK 返回
	return to, nil
}

// WithNewer 指定初始化函数
func WithNewer[V any](fu func() *V) Option {
	return func(o *options) {
		if fu != nil {
			o.callback.newer = fu
		}
	}
}

// WithTimeoutCallback 指定超时回调函数
func WithTimeoutCallback[V any](fu func(*V)) Option {
	return func(o *options) {
		if fu != nil {
			o.callback.expire = fu
		}
	}
}

// WithExpireTimeout 指定超时回调函数
func WithExpireTimeout(tm time.Duration) Option {
	return func(o *options) {
		o.timing.timeout = tm
	}
}

// WithRenewTime 指定续约回调函数
func WithRenewTime[V any](tm time.Duration) Option {
	return func(o *options) {
		o.timing.renew = tm
	}
}
