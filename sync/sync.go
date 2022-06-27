// Package sync 提供一些额外的、非常规的 sync 功能
package sync

import "time"

const (
	spinlockHungryThreshold = 50 * time.Millisecond
	minSpinLockInterval     = 10 * time.Microsecond
)
