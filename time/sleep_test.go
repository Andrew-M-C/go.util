package time

import (
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func testSleepToNextSecond(t *testing.T) {
	const N = 5
	start := time.Now()

	for i := 0; i < N; i++ {
		SleepToNextSecond()
		// t.Logf("now: %v", time.Now())
		So(time.Now().Unix()-start.Unix(), ShouldEqual, i+1)
	}

	end := time.Now()

	So(end.Unix()-start.Unix(), ShouldEqual, N)
}

func testSleepToNextSecondsN(t *testing.T) {
	const N = 5
	start := time.Now()

	SleepToNextSecondsN(N)

	end := time.Now()

	So(end.Unix()-start.Unix(), ShouldEqual, N)
}

func testWait(t *testing.T) {
	const N = 4
	start := time.Now()

	Wait(func() bool {
		return time.Since(start) > (N * time.Second)
	})

	So(int(time.Since(start).Seconds()), ShouldEqual, N)
}
