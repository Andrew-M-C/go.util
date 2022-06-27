// Package sync 提供一些额外的、非常规的 sync 功能
package sync

import (
	"fmt"
	"sync"

	"github.com/petermattis/goid"
)

// ReentrantLock 表示一个可重入锁
//
// Ref: [Go语言如何实现可重入锁？](https://segmentfault.com/a/1190000040092635)
type ReentrantLock struct {
	lock      sync.Mutex
	cond      *sync.Cond
	recursion int32
	host      int64
}

// NewReentrantLock 返回一个新的可重入锁
func NewReentrantLock() *ReentrantLock {
	lck := &ReentrantLock{}
	lck.lazyInitCond()
	return lck
}

func (lck *ReentrantLock) lazyInitCond() {
	if lck.cond == nil {
		lck.cond = sync.NewCond(&lck.lock)
	}
}

// Lock 加锁
func (lck *ReentrantLock) Lock() {
	gid := goid.Get()

	lck.lock.Lock()
	defer lck.lock.Unlock()

	lck.lazyInitCond()

	if lck.host == gid {
		lck.recursion++
		return
	}

	for lck.recursion != 0 {
		lck.cond.Wait()
	}
	lck.host = gid
	lck.recursion = 1

}

// Unlock 解锁
func (lck *ReentrantLock) Unlock() {
	gid := goid.Get()

	lck.lock.Lock()
	defer lck.lock.Unlock()

	if lck.recursion == 0 || lck.host != gid {
		err := fmt.Errorf(
			"unexpected call host: (%d); current_id: %d; recursion: %d",
			lck.host, gid, lck.recursion,
		)
		panic(err)
	}

	lck.recursion--
	if lck.recursion == 0 {
		lck.cond.Signal()
	}
}
