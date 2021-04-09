package time

import (
	"sync"
	"time"
)

// reference: [GO语言提前取消定时器](https://blog.csdn.net/u012265809/article/details/114939168)

type TimeoutCallback func()

// Timer is re-packaged time.Timer object, but have different methods.
type Timer interface {
	// Run starts timer in the background.
	Run() error

	// Stop stops the timer if it is running.
	Stop() error

	// Running check whether timer is running.
	Running() bool

	// Remain returns remaining time in the timer
	Remain() time.Duration

	// Elapsed returns elapsed time in the timer
	Elapsed() time.Duration
}

// tmr is internal implementation of Timer interface
type tmr struct {
	rwlock   sync.RWMutex
	running  bool
	paused   bool
	cb       TimeoutCallback
	duration time.Duration
	end      time.Time
	stop     chan struct{}
}

// NewTimer returns a new timer
func NewTimer(d time.Duration, cb TimeoutCallback) Timer {
	return newTmr(d, cb)
}

func newTmr(d time.Duration, cb TimeoutCallback) *tmr {
	if d < 0 {
		d = 0
	}

	t := &tmr{
		rwlock:   sync.RWMutex{},
		running:  false,
		paused:   false,
		cb:       cb,
		duration: d,
		end:      time.Time{},
		stop:     nil,
	}
	return t
}

func (t *tmr) Run() error {
	t.rwlock.Lock()
	defer t.rwlock.Unlock()

	if t.running {
		return ErrTimerIsAlreadyRunning
	}

	// init parameters
	t.running = true
	t.stop = make(chan struct{})
	t.end = time.Now().Add(t.duration)
	timer := time.NewTimer(t.duration)

	cleanWithLock := func() {
		t.rwlock.Lock()
		{
			close(t.stop)
			timer.Stop()
			t.running = false
			t.stop = nil
		}
		t.rwlock.Unlock()
	}

	// Start timer and listen timeout event
	go func() {
		select {
		case <-timer.C:
			cleanWithLock()

			if t.cb != nil {
				t.cb()
			}
			return

		case <-t.stop:
			cleanWithLock()
			return
		}
	}()

	return nil
}

func (t *tmr) Stop() error {
	t.rwlock.RLock()
	defer t.rwlock.RUnlock()

	if !t.running {
		return ErrTimerIsNotRunning
	}

	t.stop <- struct{}{}
	return nil
}

func (t *tmr) Running() bool {
	t.rwlock.RLock()
	b := t.running
	t.rwlock.RUnlock()
	return b
}

func (t *tmr) Remain() time.Duration {
	now := time.Now()

	t.rwlock.RLock()
	defer t.rwlock.RUnlock()

	if !t.running {
		return t.duration
	}

	d := t.end.Sub(now)
	if d >= 0 {
		return d
	}
	return 0
}

func (t *tmr) Elapsed() time.Duration {
	now := time.Now()

	t.rwlock.RLock()
	defer t.rwlock.RUnlock()

	if !t.running {
		return 0
	}

	d := t.end.Sub(now)
	if d >= 0 {
		return t.duration - d
	}
	return t.duration
}
