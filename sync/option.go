// Package sync 提供一些额外的、非常规的 sync 功能
package sync

import "time"

type option struct {
	interval time.Duration

	reentrantNotify ReentrantNotifyFunc
}

type Option func(opt *option)

// WithRetryInterval 指定重试的时间间隔
func WithRetryInterval(intvl time.Duration) Option {
	return func(opt *option) {
		if intvl > 0 {
			opt.interval = intvl
		}
	}
}

// ReentrantNotifyFunc 表示发生重入时的通知函数
type ReentrantNotifyFunc func(goroutineID int64)

// WithReentrantNotification 指定发生
func WithReentrantNotification(f ReentrantNotifyFunc) Option {
	return func(opt *option) {
		opt.reentrantNotify = f
	}
}
