package freshness

import (
	"time"

	"github.com/Andrew-M-C/go.util/maps/constraints"
)

type Option func(*options)

type options struct {
	timeout     time.Duration
	renew       time.Duration
	newCallback any
	expCallback any

	debug func(string, ...any)
}

func mergeOption(opts []Option) *options {
	opt := &options{
		timeout: time.Second,
		debug:   func(string, ...any) {},
	}
	for _, f := range opts {
		if f != nil {
			f(opt)
		}
	}

	if opt.timeout <= 0 {
		opt.timeout = time.Second
	}
	if opt.renew <= 0 {
		opt.renew = opt.timeout
	}
	return opt
}

func WithTimeout(tm time.Duration) Option {
	return func(o *options) {
		o.timeout = tm
	}
}

func WithRenewTime(tm time.Duration) Option {
	return func(o *options) {
		o.renew = tm
	}
}

func WithExpireCallback[K constraints.Ordered, V any](f func(key K, value V)) Option {
	return func(o *options) {
		o.expCallback = f
	}
}

func WithNewer[K constraints.Ordered, V any](f func(key K) (V, error)) Option {
	return func(o *options) {
		o.newCallback = f
	}
}

func WithDebug(f func(string, ...any)) Option {
	return func(o *options) {
		if f != nil {
			o.debug = f
		}
	}
}
