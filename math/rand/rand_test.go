package rand_test

import (
	"fmt"
	"sync"
	"testing"

	"github.com/Andrew-M-C/go.util/math/rand"
	"github.com/smartystreets/goconvey/convey"
)

var (
	cv = convey.Convey
	so = convey.So
	eq = convey.ShouldEqual

	between = convey.ShouldBeBetween
)

func TestRand(t *testing.T) {
	cv("IndexByRatios 函数", t, func() { testIndexByRatios(t) })
}

func testIndexByRatios(t *testing.T) {
	cv("基本逻辑", func() {
		const total = 100000
		ratios := []int32{11, 22, 33, 44}
		result := make([]int, len(ratios))

		for i := 0; i < total; i++ {
			v := rand.IndexByRatios(ratios, func(_ int, v int32) int32 { return v })
			so(v, between, -1, len(ratios))
			result[v]++
		}

		t.Logf("result: %v", result)

		so(result[0], between, inRange(10000, 0.05)...)
		so(result[1], between, inRange(20000, 0.05)...)
		so(result[2], between, inRange(30000, 0.05)...)
		so(result[3], between, inRange(40000, 0.05)...)
	})

	cv("并发抽奖", func() {
		const eachTotal = 10000
		const concurrency = 256
		ratios := []int32{11, 22, 33, 44}
		result := make([]int, len(ratios))

		wg := &sync.WaitGroup{}
		wg.Add(concurrency)

		for c := 0; c < concurrency; c++ {
			go func() {
				defer wg.Done()
				for i := 0; i < eachTotal; i++ {
					v := rand.IndexByRatios(ratios, func(_ int, v int32) int32 { return v })
					fmt.Printf("%d", v)
					result[v]++
				}
			}()
		}

		wg.Wait()
		t.Logf("result: %v", result)

		so(result[0], between, inRange(1000*concurrency, 0.05)...)
		so(result[1], between, inRange(2000*concurrency, 0.05)...)
		so(result[2], between, inRange(3000*concurrency, 0.05)...)
		so(result[3], between, inRange(4000*concurrency, 0.05)...)
	})

	cv("部分值为零或者小于零的情况", func() {
		const total = 100000
		ratios := []int{-100, 22, 0, 33, 0}
		result := make([]int, len(ratios))

		for i := 0; i < total; i++ {
			v := rand.IndexByRatios(ratios, func(_ int, v int) int { return v })
			so(v, between, -1, len(ratios))
			result[v]++
		}

		t.Logf("result: %v", result)

		so(result[0], eq, 0)
		so(result[1], between, inRange(40000, 0.05)...)
		so(result[2], eq, 0)
		so(result[3], between, inRange(60000, 0.05)...)
		so(result[4], eq, 0)
	})
}

func inRange(target int, deviation float64) []any {
	lower := int(float64(target) * (1.0 - deviation))
	higher := int(float64(target) * (1.0 + deviation))
	return []any{lower, higher}
}
