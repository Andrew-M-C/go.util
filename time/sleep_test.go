package time

import (
	"testing"
	"time"
)

func testSleepToNextSecond(*testing.T) {
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

func testSleepToNextSecondsN(*testing.T) {
	const N = 5
	start := time.Now()

	SleepToNextSecondsN(N)

	end := time.Now()

	so(end.Unix()-start.Unix(), eq, N)
}

func testWait(*testing.T) {
	const N = 4
	start := time.Now()

	Wait(func() bool {
		return time.Since(start) > (N * time.Second)
	})

	so(int(time.Since(start).Seconds()), eq, N)
}

func testSleep(*testing.T) {
	t0 := time.Now()

	Sleep(0.1)
	t1 := time.Now()
	so(t1.Sub(t0), ge, 100*time.Millisecond)
	so(t1.Sub(t0), le, 200*time.Millisecond)

	Sleep(1)
	t2 := time.Now()
	so(t2.Sub(t1), ge, time.Second)
	so(t2.Sub(t1), le, time.Second+100*time.Millisecond)
}
