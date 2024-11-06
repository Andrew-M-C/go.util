package atomic_test

import (
	"math/rand"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	util "github.com/Andrew-M-C/go.util/sync/atomic"
	"github.com/smartystreets/goconvey/convey"
)

var (
	cv = convey.Convey
	so = convey.So

	notPanic = convey.ShouldNotPanic
)

func TestValue(t *testing.T) {
	cv("atomic.Value", t, func() {
		handler := func() {
			v := &util.Value[[]byte]{}

			const concurrency = 1000
			const operationTime = 10 * time.Second

			loopCount := uint64(0)

			// 开多个协程并发读写
			wg := &sync.WaitGroup{}
			wg.Add(concurrency)

			for i := 0; i < concurrency; i++ {
				go func() {
					defer wg.Done()
					start := time.Now()
					for time.Since(start) < operationTime {
						_ = v.Load()
						v.Store(make([]byte, rand.Intn(4096)))
						_ = atomic.AddUint64(&loopCount, 1)
					}
				}()
			}

			start := time.Now()
			for time.Since(start) < operationTime {
				time.Sleep(time.Second)
				t.Log(time.Now())
			}

			wg.Wait()

			t.Log("共循环次数:", loopCount)
		}

		so(handler, notPanic)
	})
}
