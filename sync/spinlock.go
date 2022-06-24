package sync

import (
	"sync/atomic"
	"time"
)

const (
	minSpinLockInterval = 10 * time.Microsecond
)

// SpinLock 表示一个自旋锁
type SpinLock struct {
	flag int32
	opt  option
}

// NewSpinLock 返回一个新的自旋锁
func NewSpinLock(opts ...Option) *SpinLock {
	lck := &SpinLock{}

	for _, o := range opts {
		o(&lck.opt)
	}
	return lck
}

// Lock 加锁
func (lck *SpinLock) Lock() {
	intvl := lck.opt.interval
	if intvl < minSpinLockInterval {
		intvl = minSpinLockInterval
	}
	locked := atomic.CompareAndSwapInt32(&lck.flag, 0, 1)
	for !locked {
		time.Sleep(intvl)
		locked = atomic.CompareAndSwapInt32(&lck.flag, 0, 1)
	}
}

// Unlock 解锁
func (lck *SpinLock) Unlock() {
	if atomic.AddInt32(&lck.flag, -1) < 0 {
		panic("try to unlock an unlocked spinlock!")
	}
}

// TryLock 尝试是否加锁
func (lck *SpinLock) TryLock() bool {
	return atomic.CompareAndSwapInt32(&lck.flag, 0, 1)
}
