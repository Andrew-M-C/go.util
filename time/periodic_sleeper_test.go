package time

import (
	"testing"
	"time"

	"github.com/Andrew-M-C/go.util/slices"
)

func testPeriodicSleeper(t *testing.T) {
	start := UpTime()
	intervals := []time.Duration{
		2 * time.Millisecond,
		3 * time.Millisecond,
		4 * time.Millisecond,
		5 * time.Millisecond,
	}

	const sleepCount = 500
	sl := NewPeriodicSleeper()

	for i := 0; i < sleepCount; i++ {
		d := intervals[i&0x3]
		sl.Sleep(d)
		t.Logf("%v sleep for %v", time.Now(), d)
	}

	avgSleepTime := (UpTime() - start) / sleepCount
	so(avgSleepTime, ge, percentage(slices.AverageFloat(intervals), 0.9))
	so(avgSleepTime, le, percentage(slices.AverageFloat(intervals), 1.1))

	t.Logf(
		"总共进行了 %v, 平均间隔时间 %v, 理论平均间隔时间 %v",
		UpTime()-start, avgSleepTime, time.Duration(slices.AverageFloat(intervals)),
	)
}
