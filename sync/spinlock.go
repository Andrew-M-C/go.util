package sync

import (
	"sync/atomic"
	"time"
)

// SpinLock 表示一个自旋锁
type SpinLock struct {
	lockFlag int32
	opt      option
}

// NewSpinLock 返回一个新的自旋锁
func NewSpinLock(opts ...Option) *SpinLock {
	lck := &SpinLock{}

	for _, o := range opts {
		o(&lck.opt)
	}
	return lck
}

func (lck *SpinLock) wait(waitStart time.Time) {
	intvl := lck.opt.interval
	if intvl < minSpinLockInterval {
		intvl = minSpinLockInterval
	}
	if time.Since(waitStart) > spinlockHungryThreshold {
		intvl = minSpinLockInterval
	}
	time.Sleep(intvl)
}

// Lock 加锁
func (lck *SpinLock) Lock() {
	waitStart := time.Now()
	locked := atomic.CompareAndSwapInt32(&lck.lockFlag, 0, 1)
	for !locked {
		lck.wait(waitStart)
		locked = atomic.CompareAndSwapInt32(&lck.lockFlag, 0, 1)
	}
}

// Unlock 解锁
func (lck *SpinLock) Unlock() {
	if atomic.AddInt32(&lck.lockFlag, -1) < 0 {
		panic("try to unlock an unlocked spinlock!")
	}
}

// TryLock 尝试是否加锁
func (lck *SpinLock) TryLock() bool {
	return atomic.CompareAndSwapInt32(&lck.lockFlag, 0, 1)
}
