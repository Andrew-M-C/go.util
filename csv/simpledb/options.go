package simpledb

import "time"

// Option 表示 simpledb.DB 的各种配置参数
type Option func(*options)

type options struct {
	// 异步写入时间。<= 0 表示同步写入
	asyncTime time.Duration
	// 需要保持唯一的列
	uniqueColumns map[string]struct{}
	// 调试函数
	debugf func(string, ...any)
}

// WithAsyncTime 设置异步写入时间, <= 0 表示同步写入
func WithAsyncTime(tm time.Duration) Option {
	return func(o *options) {
		o.asyncTime = tm
	}
}

// WithUniqueColumn 设置需要保持唯一的列名
func WithUniqueColumns[T ~string](columns ...T) Option {
	return func(o *options) {
		if o.uniqueColumns == nil {
			o.uniqueColumns = make(map[string]struct{})
		}
		for _, column := range columns {
			o.uniqueColumns[string(column)] = struct{}{}
		}
	}
}

// WithDebugger 设置调试器
func WithDebugger(f func(string, ...any)) Option {
	return func(o *options) {
		if f != nil {
			o.debugf = f
		}
	}
}

func mergeOptions(opts []Option) *options {
	o := &options{
		debugf: func(string, ...any) {},
	}
	for _, opt := range opts {
		if opt != nil {
			opt(o)
		}
	}
	return o
}
