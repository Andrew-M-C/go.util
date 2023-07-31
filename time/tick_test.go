package time

import (
	"testing"
	"time"
)

func testTick(t *testing.T) {
	last := time.Now()
	start := last
	callCount := 0

	ti, err := NewTickBeta(5*time.Millisecond, func(TickCallbackParam) {
		callCount++
		now := time.Now()
		t.Logf("间隔 %v", now.Sub(last))
		last = now
	})
	so(err, eq, nil)

	ti.Run()
	time.Sleep(30 * time.Second)
	ti.Stop()

	total := last.Sub(start)
	t.Logf("总共进行了 %v, 平均间隔时间 %v", total, total/time.Duration(callCount))
}
