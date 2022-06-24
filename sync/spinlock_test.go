// Package sync 提供一些额外的、非常规的 sync 功能
package sync

import (
	"testing"
	"time"
)

func testSpinLock(t *testing.T) {
	cv("普通加解锁", func() {
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
	})

	cv("trylock", func() {
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
		t.Log(lck.flag)
		time.Sleep(100 * time.Millisecond)
		so(entered2, isFalse)
		lck.Unlock()

		so(true, isTrue)

		time.Sleep(100 * time.Millisecond)
		so(try2, isFalse)
		so(entered2, isTrue)
		so(exited2, isTrue)
	})
}
