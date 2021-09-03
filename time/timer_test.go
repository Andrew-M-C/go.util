package time

import (
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
	"go.uber.org/atomic"
)

func TestTime(t *testing.T) {
	Convey("测试 SleepToNextSecond", t, func() { testSleepToNextSecond((t)) })
	Convey("测试 SleepToNextSecondsN", t, func() { testSleepToNextSecondsN((t)) })
	Convey("测试 Wait", t, func() { testWait((t)) })
	Convey("测试 Timer", t, func() { testTimer(t) })
	Convey("测试 Timer.Stop", t, func() { testTimer_Stop(t) })
}

func expectElapsed(t *testing.T, start time.Time, ela time.Duration) {
	// d := time.Since(start)
	// expectDuration(t, d, ela)
	// TODO:
}

func expectDuration(t *testing.T, d, expect time.Duration) {
	tolerance := 10 * time.Millisecond

	abs := func(d time.Duration) time.Duration {
		if d > 0 {
			return d
		}
		return -d
	}

	So(abs(expect-d), ShouldBeLessThanOrEqualTo, tolerance)
}

func testTimer(t *testing.T) {
	secs := 5

	duration := time.Duration(secs) * time.Second
	gotCb := atomic.NewBool(false)
	var start time.Time

	tm := NewTimer(duration, func() {
		gotCb.Store(true)

		expectElapsed(t, start, duration)
		t.Logf("Timeout received, elapsed %v", time.Since(start))
	})

	SleepToNextSecond()
	SleepToNextSecond()

	expectDuration(t, tm.Elapsed(), 0)

	tm.Run()
	start = time.Now()

	for i := 0; i < 5; i++ {
		ela := tm.Elapsed()
		rm := tm.Remain()
		expectDuration(t, ela, time.Duration(i)*time.Second)
		expectDuration(t, rm, time.Duration(secs-i)*time.Second)
		t.Logf("elapsed %v", ela)
		SleepToNextSecond()
	}

	time.Sleep(10 * time.Millisecond)
	if tm.Running() {
		t.Errorf("timer should be stopped")
		return
	}

	if !gotCb.Load() {
		t.Errorf("callback not invoked!")
		return
	}

	// succ
	t.Logf("callback invoked")
}

func testTimer_Stop(t *testing.T) {
	secs := 5
	duration := time.Duration(secs) * time.Second
	gotCb := atomic.NewBool(false)
	var start time.Time

	tm := NewTimer(duration, func() {
		gotCb.Store(true)

		// expectElapsed(t, start, duration)
		t.Logf("Timeout received, elapsed %v", time.Since(start))
	})

	SleepToNextSecond()
	SleepToNextSecond()
	tm.Run()

	SleepToNextSecond()
	SleepToNextSecond()

	expectDuration(t, tm.Elapsed(), 2*time.Second)

	tm.Stop()

	t.Logf("Timer stopped, now let us sleep for %v", duration)
	time.Sleep(duration)
	if gotCb.Load() {
		t.Errorf("callback should NOT be invoked!")
		return
	}
	t.Logf("callback is not invoked, this is expected")

	// try re-start timer
	tm.Run()
	start = time.Now()

	for i := 0; i < 5; i++ {
		ela := tm.Elapsed()
		rm := tm.Remain()
		expectDuration(t, ela, time.Duration(i)*time.Second)
		expectDuration(t, rm, time.Duration(secs-i)*time.Second)
		t.Logf("elapsed %v", ela)
		SleepToNextSecond()
	}

	time.Sleep(10 * time.Millisecond)
	if tm.Running() {
		t.Errorf("timer should be stopped")
		return
	}

	if !gotCb.Load() {
		t.Errorf("callback not invoked!")
		return
	}

	// succ
	t.Logf("callback invoked")
}
