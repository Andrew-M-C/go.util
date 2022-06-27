// Package sync 提供一些额外的、非常规的 sync 功能
package sync

import (
	"math/rand"
	"sync"
	"testing"
	"time"
)

func testSpinLock(t *testing.T) {
	cv("普通的锁操作", func() { testSpinLockGeneral(t) })
	cv("TryLock 操作", func() { testSpinLockTryLock(t) })
	cv("高并发读写验证不 panic", func() { testSpinLockConcurrency(t) })
}

func testSpinLockGeneral(t *testing.T) {
	lck := SpinLock{}

	entered2 := false
	exited2 := false

	// 协程2
	go func() {
		defer func() {
			exited2 = true
		}()

		time.Sleep(50 * time.Millisecond)
		lck.Lock()
		defer lck.Unlock()

		entered2 = true
	}()

	// 协程1
	lck.Lock()
	so(true, isTrue)
	time.Sleep(100 * time.Millisecond)
	so(entered2, isFalse)
	lck.Unlock()

	so(true, isTrue)

	time.Sleep(100 * time.Millisecond)
	so(entered2, isTrue)
	so(exited2, isTrue)
}

func testSpinLockTryLock(t *testing.T) {
	lck := SpinLock{}

	try2 := true
	entered2 := false
	exited2 := false

	// 协程2
	go func() {
		defer func() {
			if entered2 {
				exited2 = true
			}
		}()

		time.Sleep(50 * time.Millisecond)
		try2 = lck.TryLock()
		t.Logf("try lock result: %v", try2)

		time.Sleep(100 * time.Millisecond)
		lck.Lock()
		defer lck.Unlock()
		entered2 = true
	}()

	// 协程1
	lck.Lock()
	so(true, isTrue)
	t.Log(lck.lockFlag)
	time.Sleep(100 * time.Millisecond)
	so(entered2, isFalse)
	lck.Unlock()

	so(true, isTrue)

	time.Sleep(100 * time.Millisecond)
	so(try2, isFalse)
	so(entered2, isTrue)
	so(exited2, isTrue)
}

func testSpinLockConcurrency(t *testing.T) {

	const concurrency = 500
	const tm = time.Second
	m := make(map[int64]bool)
	lck := SpinLock{}

	so(func() {
		start := time.Now()
		wg := sync.WaitGroup{}
		wg.Add(concurrency)
		for i := 0; i < concurrency; i++ {
			go func() {
				defer wg.Done()
				for time.Since(start) < tm {
					// 设置两个
					key := rand.Int63()
					lck.Lock()
					m[key] = true
					lck.Unlock()

					key = rand.Int63()
					lck.Lock()
					m[key] = true
					lck.Unlock()

					// 删除一个
					lck.Lock()
					for k := range m {
						delete(m, k)
						break
					}
					lck.Unlock()
				}
			}()
		}

		wg.Wait()
	}, notPanic)
}
