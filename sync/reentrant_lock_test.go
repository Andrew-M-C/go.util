package sync

import (
	"bytes"
	"math/rand"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func testReentrantLock(t *testing.T) {
	cv("普通的锁操作", func() { testReentrantLockGeneral(t) })
	cv("高并发读写验证不 panic", func() { testReentrantLockConcurrency(t) })
	cv("高并发可重入读写验证不 panic", func() { testReentrantLockConcurrencyReentrantly(t) })
}

func testReentrantLockGeneral(t *testing.T) {
	lck := ReentrantLock{}

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

func testReentrantLockConcurrency(t *testing.T) {

	const concurrency = 500
	const tm = time.Second
	m := make(map[int64]bool)
	lck := ReentrantLock{}

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

func testReentrantLockConcurrencyReentrantly(t *testing.T) {

	const concurrency = 500
	const tm = time.Second
	m := make(map[int64]bool)

	var notifyCnt int64
	var expectedNotifyCnt int64

	notify := func(_ int64) {
		atomic.AddInt64(&notifyCnt, 1)
	}

	lck := NewReentrantLock(WithReentrantNotification(notify))
	// lck := sync.Mutex{}

	so(func() {
		start := time.Now()
		wg := sync.WaitGroup{}
		wg.Add(concurrency)
		for i := 0; i < concurrency; i++ {
			go func() {
				defer wg.Done()

				repeat := bytes.Repeat([]byte{'0'}, rand.Intn(10)+1)
				// repeat := []int{1}
				lock := func() {
					atomic.AddInt64(&expectedNotifyCnt, int64(len(repeat)-1))
					for range repeat {
						lck.Lock()
					}
				}
				unlock := func() {
					for range repeat {
						lck.Unlock()
					}
				}

				for time.Since(start) < tm {
					// 设置两个
					key := rand.Int63()
					lock()
					m[key] = true
					unlock()

					key = rand.Int63()
					lock()
					m[key] = true
					unlock()

					// 删除一个
					lock()
					for k := range m {
						delete(m, k)
						break
					}
					unlock()
				}
			}()
		}

		wg.Wait()
	}, notPanic)

	so(notifyCnt, eq, expectedNotifyCnt)
}
