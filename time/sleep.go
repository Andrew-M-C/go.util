package time

import (
	"time"

	"golang.org/x/exp/constraints"
)

// SleepToNextSecond sleep 到墙上时钟的下一秒
func SleepToNextSecond() {
	now := time.Now()
	nano := time.Duration(now.Nanosecond()) * time.Nanosecond
	time.Sleep(time.Second - nano)
}

// SleepToNextSecondsN sleep x N 版
func SleepToNextSecondsN(n int) {
	for i := 0; i < n; i++ {
		SleepToNextSecond()
	}
}

// Wait 等待是否满足条件，如果不满足则自旋等待。interval 如果不传参的话，默认是 100ms。如果传参的话，最小则为 1ms
func Wait(f func() (done bool), interval ...time.Duration) {
	if f == nil {
		return
	}
	d := 100 * time.Millisecond

	if len(interval) > 0 {
		d = interval[0]
		if d < time.Millisecond {
			d = time.Microsecond
		}
	}

	for !f() {
		time.Sleep(d)
	}
}

// Sleep 按秒数 sleep
func Sleep[N constraints.Float | constraints.Integer](secs N) {
	duration := float64(secs) * float64(time.Second)
	time.Sleep(time.Duration(duration))
}
