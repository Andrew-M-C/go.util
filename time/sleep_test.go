package time

import (
	"testing"
	"time"
)

func testSleepToNextSecond(t *testing.T) {
	const N = 5
	start := time.Now()

	for i := 0; i < N; i++ {
		SleepToNextSecond()
		// t.Logf("now: %v", time.Now())
		so(time.Now().Unix()-start.Unix(), eq, i+1)
	}

	end := time.Now()

	so(end.Unix()-start.Unix(), eq, N)
}

func testSleepToNextSecondsN(t *testing.T) {
	const N = 5
	start := time.Now()

	SleepToNextSecondsN(N)

	end := time.Now()

	so(end.Unix()-start.Unix(), eq, N)
}

func testWait(t *testing.T) {
	const N = 4
	start := time.Now()

	Wait(func() bool {
		return time.Since(start) > (N * time.Second)
	})

	so(int(time.Since(start).Seconds()), eq, N)
}
