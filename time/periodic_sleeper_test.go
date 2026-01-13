package time

import (
	"testing"
	"time"
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
	so(avgSleepTime, ge, percentage(averageIntervalFloat(intervals), 0.9))
	so(avgSleepTime, le, percentage(averageIntervalFloat(intervals), 1.1))

	t.Logf(
		"总共进行了 %v, 平均间隔时间 %v, 理论平均间隔时间 %v",
		UpTime()-start, avgSleepTime, time.Duration(averageIntervalFloat(intervals)),
	)
}

func averageIntervalFloat(intervals []time.Duration) float64 {
	total := float64(0)
	for _, interval := range intervals {
		total += float64(interval)
	}
	return total / float64(len(intervals))
}
