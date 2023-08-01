package time

import (
	"testing"
	"time"
)

func testTick(t *testing.T) {
	last := time.Now()
	start := last
	callCount := 0
	interval := 2 * time.Millisecond

	ti, err := NewTickBeta(interval, func(TickCallbackParam) {
		callCount++
		now := time.Now()

		if callCount%123 == 0 {
			t.Logf("%v - 间隔 %v", time.Now(), now.Sub(last))
		}

		last = now
	})
	so(err, eq, nil)

	ti.Run()
	time.Sleep(20*time.Second + interval)
	ti.Stop()

	total := last.Sub(start)
	t.Logf("总共进行了 %v, 平均间隔时间 %v", total, total/time.Duration(callCount))
}
