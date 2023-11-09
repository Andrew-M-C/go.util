package maps

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func testRWSafeMap(t *testing.T) {
	cv("基本并发逻辑", func() {
		const concurrency = 500
		const maxValue = 10000
		const testTime = 5 * time.Second

		// 首先新建一个 map
		m := NewRWSafeMap[int, string](10000)
		wg := sync.WaitGroup{}
		start := time.Now()
		var iterateCount int64

		testContinue := func() bool {
			return time.Since(start) < testTime
		}

		// 首先开一堆协程并发读写这个 map
		wg.Add(concurrency)
		for i := 0; i < concurrency; i++ {
			go func() {
				defer wg.Done()

				for testContinue() {
					i := rand.Intn(maxValue)
					s := strings.Repeat(fmt.Sprint(i), rand.Intn(10)+1)
					_, _ = m.Swap(i, s)
					atomic.AddInt64(&iterateCount, 1)
				}
			}()
		}

		// 然后开一个协程慢悠悠地操作
		type testResult struct {
			i      int
			v      string
			loaded bool
		}
		results := make(chan testResult, 10000)
		wg.Add(1)
		go func() {
			defer wg.Done()

			for testContinue() {
				i := rand.Intn(maxValue)
				v, loaded := m.LoadOrStore(i, fmt.Sprint(i))

				results <- testResult{
					i, v, loaded,
				}

				time.Sleep(100 * time.Millisecond)
			}

			close(results)
		}()

		for res := range results {
			t.Logf("result: %v - %v", res.v, res.loaded)
			so(len(res.v), gt, 0)
			so(strings.ReplaceAll(res.v, fmt.Sprint(res.i), ""), eq, "")
		}
		wg.Wait()

		t.Logf("safe map size: %d, written times: %d", m.Size(), iterateCount)
	})

	cv("JSON 序列化和反序列化", func() {
		raw := `{"aaaa":1111,"bbbb":2222}`
		m := NewRWSafeMap[string, int]()
		err := json.Unmarshal([]byte(raw), &m)
		so(err, isNil)

		got, exist := m.Load("aaaa")
		so(exist, eq, true)
		so(got, eq, 1111)

		got, exist = m.Load("bbbb")
		so(exist, eq, true)
		so(got, eq, 2222)
	})
}
