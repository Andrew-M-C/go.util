// Package sync 提供一些额外的、非常规的 sync 功能
package sync

import "time"

type option struct {
	interval time.Duration
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
