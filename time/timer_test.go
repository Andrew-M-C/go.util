package time

import (
	"testing"
	"time"

	"github.com/Andrew-M-C/go.util/constraints"
	"github.com/smartystreets/goconvey/convey"
	"go.uber.org/atomic"
)

var (
	cv = convey.Convey
	so = convey.So

	eq = convey.ShouldEqual
	le = convey.ShouldBeLessThanOrEqualTo
	ge = convey.ShouldBeGreaterThanOrEqualTo

	isNil  = convey.ShouldBeNil
	notNil = convey.ShouldNotBeNil

	hasSubStr = convey.ShouldContainSubstring
)

func TestTime(t *testing.T) {
	cv("测试 Age", t, func() { testAge(t) })
	cv("测试 UpTime", t, func() { testUpTime(t) })
	cv("测试 SleepToNextSecond", t, func() { testSleepToNextSecond((t)) })
	cv("测试 SleepToNextSecondsN", t, func() { testSleepToNextSecondsN((t)) })
	cv("测试 Wait", t, func() { testWait((t)) })
	cv("测试 Sleep", t, func() { testSleep((t)) })
	cv("测试 Timer", t, func() { testTimer(t) })
	cv("测试 Timer.Stop", t, func() { testTimerStop(t) })
	cv("测试 PeriodicSleeper", t, func() { testPeriodicSleeper(t) })
	cv("测试 Tick", t, func() { testTick(t) })
	cv("测试 UnixFloat", t, func() { testUnixFloat(t) })
	cv("测试 TimeSection", t, func() { testTimeSection(t) })
}

type number interface {
	constraints.Integer | constraints.Float
}

func percentage[V number, P constraints.Float](v V, percent P) V {
	return V(P(v) * percent)
}

func expectElapsed(t *testing.T, start time.Time, ela time.Duration) {
	// d := time.Since(start)
	// expectDuration(t, d, ela)
	// TODO:
}

func expectDuration(_ *testing.T, d, expect time.Duration) {
	tolerance := 10 * time.Millisecond

	abs := func(d time.Duration) time.Duration {
		if d > 0 {
			return d
		}
		return -d
	}

	so(abs(expect-d), le, tolerance)
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

	_ = tm.Run()
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

	// success
	t.Logf("callback invoked")
}

func testTimerStop(t *testing.T) {
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
	err := tm.Run()
	so(err, eq, nil)

	SleepToNextSecond()
	SleepToNextSecond()

	expectDuration(t, tm.Elapsed(), 2*time.Second)

	err = tm.Stop()
	so(err, eq, nil)

	t.Logf("Timer stopped, now let us sleep for %v", duration)
	time.Sleep(duration)
	if gotCb.Load() {
		t.Errorf("callback should NOT be invoked!")
		return
	}
	t.Logf("callback is not invoked, this is expected")

	// try re-start timer
	err = tm.Run()
	so(err, eq, nil)
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

	// success
	t.Logf("callback invoked")
}

func testUnixFloat(*testing.T) {
	tm := time.Unix(1704042061, 123456789)
	f := UnixFloat(tm)
	so(f, eq, float64(1704042061.123456))
}
