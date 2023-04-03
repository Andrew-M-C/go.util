package recovery

import (
	"context"

	"github.com/Andrew-M-C/go.util/runtime/caller"
)

type option struct {
	ctx          context.Context
	withErrorLog bool
	callback     PanicCallback
}

// PanicCallback 发生 panic 时的回调函数
type PanicCallback func(info any, stack []caller.Caller)

func mergeOptions(opts []Option) *option {
	o := &option{}
	for _, f := range opts {
		if f != nil {
			f(o)
		}
	}
	return o
}

// Option 表示额外参数
type Option func(o *option)

// WithErrorLog 表示按照默认格式默认写入错误日志
func WithErrorLog() Option {
	return func(o *option) {
		o.withErrorLog = true
	}
}

// WithContext 填充 context。当拥有 ctx 时, 输出日志将会调用 ErrorContext
func WithContext(ctx context.Context) Option {
	return func(o *option) {
		o.ctx = ctx
	}
}

// WithCallback 出现 panic 时回调
func WithCallback(f PanicCallback) Option {
	return func(o *option) {
		o.callback = f
	}
}
