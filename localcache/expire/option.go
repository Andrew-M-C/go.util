package expire

import "time"

type option struct {
	timeout time.Duration
	debug   func(string, ...any)

	callback struct {
		newer   any
		timeout any
	}
}

// Option 额外选项
type Option func(*option)

func defaultOption[K comparable, V any]() *option {
	o := &option{
		timeout: 5 * time.Minute,
		debug:   func(string, ...any) { /* do nothing */ },
	}
	o.callback.newer = func(K) *V {
		return new(V)
	}
	return o
}

func (o *option) copy() *option {
	res := &option{}
	res.timeout = o.timeout
	res.callback.newer = o.callback.newer
	res.callback.timeout = o.callback.timeout
	return res
}

func (o *option) merge(opts []Option) *option {
	for _, f := range opts {
		if f != nil {
			f(o)
		}
	}
	return o
}

func WithTimeout(timeout time.Duration) Option {
	return func(o *option) {
		if timeout > 0 {
			o.timeout = timeout
		}
	}
}

func WithTimeoutCallback[K comparable, V any](f func(key K, value *V)) Option {
	return func(o *option) {
		o.callback.timeout = f
	}
}

func WithNewer[K comparable, V any](f func(key K) *V) Option {
	return func(o *option) {
		if f == nil {
			return
		}
		o.callback.newer = f
	}
}

func WithDebugger(f func(string, ...any)) Option {
	return func(o *option) {
		if f != nil {
			o.debug = f
		}
	}
}
