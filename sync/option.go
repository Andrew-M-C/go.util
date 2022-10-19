package sync

import "time"

type option struct {
	interval time.Duration

	reentrantNotify ReentrantNotifyFunc
}

// Option 表示额外参数, 只能通过本 package 调用
type Option func(opt *option)

// WithRetryInterval 指定重试的时间间隔
func WithRetryInterval(intvl time.Duration) Option {
	return func(opt *option) {
		if intvl > 0 {
			opt.interval = intvl
		}
	}
}

// ReentrantNotifyFunc 表示发生重入时的通知函数类型
type ReentrantNotifyFunc func(goroutineID int64)

// WithReentrantNotification 指定发生重入时, 调用回调
func WithReentrantNotification(f ReentrantNotifyFunc) Option {
	return func(opt *option) {
		opt.reentrantNotify = f
	}
}
