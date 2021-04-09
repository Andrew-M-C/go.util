package time

import (
	"testing"
	"time"

	"github.com/Andrew-M-C/go-tools/callstack"
	"go.uber.org/atomic"
)

func expectElapsed(t *testing.T, start time.Time, ela time.Duration) bool {
	d := time.Since(start)
	return expectDuration(t, d, ela)
}

func expectDuration(t *testing.T, d, expect time.Duration) bool {
	tolerance := 5 * time.Millisecond

	abs := func(d time.Duration) time.Duration {
		if d > 0 {
			return d
		}
		return -d
	}

	if abs(expect-d) <= tolerance {
		return true
	}

	f, l, _ := callstack.CallerInfo(1)
	t.Errorf("%s, Line %d: expect duration %v, but got %v", f, l, expect, d)
	return false
}

func sleepToNextSecond(t *testing.T) {
	now := time.Now()
	nano := (now.UnixNano() - now.Unix()*1000000000)

	next := time.Duration(1000000000 - nano)
	// t.Logf("now: %v, milli %v, sleep %v", now, milli, next)
	time.Sleep(next)
}

func TestTimer(t *testing.T) {
	secs := 5

	duration := time.Duration(secs) * time.Second
	gotCb := atomic.NewBool(false)
	var start time.Time

	tm := NewTimer(duration, func() {
		gotCb.Store(true)

		if !expectElapsed(t, start, duration) {
			return
		}
		t.Logf("Timeout received, elapsed %v", time.Since(start))
	})

	sleepToNextSecond(t)
	sleepToNextSecond(t)

	if !expectDuration(t, tm.Elapsed(), 0) {
		return
	}

	tm.Run()
	start = time.Now()

	for i := 0; i < 5; i++ {
		ela := tm.Elapsed()
		rm := tm.Remain()
		if !expectDuration(t, ela, time.Duration(i)*time.Second) {
			return
		}
		if !expectDuration(t, rm, time.Duration(secs-i)*time.Second) {
			return
		}
		t.Logf("elapsed %v", ela)
		sleepToNextSecond(t)
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

func TestTimer_Stop(t *testing.T) {
	secs := 5
	duration := time.Duration(secs) * time.Second
	gotCb := atomic.NewBool(false)
	var start time.Time

	tm := NewTimer(duration, func() {
		gotCb.Store(true)

		if !expectElapsed(t, start, duration) {
			return
		}
		t.Logf("Timeout received, elapsed %v", time.Since(start))
	})

	sleepToNextSecond(t)
	sleepToNextSecond(t)
	tm.Run()

	sleepToNextSecond(t)
	sleepToNextSecond(t)
	if !expectDuration(t, tm.Elapsed(), 2*time.Second) {
		return
	}

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
		if !expectDuration(t, ela, time.Duration(i)*time.Second) {
			return
		}
		if !expectDuration(t, rm, time.Duration(secs-i)*time.Second) {
			return
		}
		t.Logf("elapsed %v", ela)
		sleepToNextSecond(t)
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
