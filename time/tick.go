package time

import (
	"errors"
	"sync/atomic"
	"time"
)

// Tick 提供一个长期尽可能周期性的 tick
type Tick interface {
	Run()
	Stop()
	SetCallback(fn TickCallback)
}

type TickCallbackParam struct {
	// TODO: 预留字段
}

type TickCallback func(param TickCallbackParam)

// NewTickBeta 新建一个 tick, 目前暂时实现毫秒级, 再低了不支持。此外, 仅支持精确到毫秒, 更低的不支持
//
// 此外, 目前理论上再高并发时会有竞争问题, 建议不要频繁创建销毁
func NewTickBeta(interval time.Duration, callback TickCallback) (Tick, error) {
	if interval < time.Millisecond {
		return nil, errors.New("intervals lower than 1 millisecond are not supported")
	}
	if intvl := interval.Round(time.Millisecond); intvl != interval {
		return nil, errors.New("only supports accuracy to millisecond")
	}
	if interval > 5*time.Second {
		return nil, errors.New("临时错误: 大于5秒的间隔暂不支持")
	}
	t := &tickForMilliSeconds{
		interval: interval,
		callback: callback,
	}
	return t, nil
}

// ======== tickForMilliSeconds ========
type tickForMilliSeconds struct {
	_ noCopy

	shouldRun bool
	running   atomic.Bool
	interval  time.Duration
	callback  TickCallback
}

func (t *tickForMilliSeconds) SetCallback(fn TickCallback) {
	t.callback = fn
}

func (t *tickForMilliSeconds) Run() {
	t.shouldRun = true

	if swapped := t.running.CompareAndSwap(false, true); !swapped {
		// already running
		return
	}

	go t.doRun()
}

func (t *tickForMilliSeconds) Stop() {
	t.shouldRun = false
}

func (t *tickForMilliSeconds) doRun() {
	next := UpTime()

	for {

		if !t.shouldRun {
			t.running.Store(false)
			return
		}

		callback := t.callback
		if callback != nil {
			go callback(TickCallbackParam{})
		}

		next += t.interval
		for next-UpTime() < 0 {
			next += t.interval
		}
		time.Sleep(next - UpTime())
	}
}
