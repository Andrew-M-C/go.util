package time

import "time"

// PeriodicSleeper 用在以下场景: 每次 sleep 结束之后, 当指定下一次 sleep 时长 T 时, 不会
// 完全按照 T 来 sleep, 而是参照自身上一次 sleep 的预期时间 tₙ₋₁ + T 来 sleep。这与 Tick 类似,
// 不同的是每次 sleep 的时间需要重新指定。
//
// 需要注意的是: 如果 tₙ₋₁ + T 小于当前时间, 则会触发一次 Reset 动作之后再 + T
type PeriodicSleeper interface {
	Sleep(d time.Duration)
	Reset()
}

// NewPeriodicSleeper 以当前时间新建一个 PeriodicSleeper
func NewPeriodicSleeper() PeriodicSleeper {
	p := &periodicSleeperImpl{}
	p.Reset()
	return p
}

type periodicSleeperImpl struct {
	lastSleepFinTime time.Duration
}

func (p *periodicSleeperImpl) Reset() {
	p.lastSleepFinTime = UpTime()
}

func (p *periodicSleeperImpl) Sleep(d time.Duration) {
	now := UpTime()
	next := p.lastSleepFinTime + d
	if actualDuration := next - now; actualDuration < 0 {
		p.lastSleepFinTime = now + d
		time.Sleep(d)
	} else {
		p.lastSleepFinTime = next
		time.Sleep(actualDuration)
	}
}
