package cpu

import (
	"runtime"
	"sync"
	"sync/atomic"

	"github.com/Andrew-M-C/go.util/runtime/cpu/internal/ring"
	"go.uber.org/automaxprocs/maxprocs"
)

var internal = struct {
	cpuNum int

	debugger func(string, ...any)

	ensureCPUUsageRunning  atomic.Bool
	ensureCPUUsageTarget   int32 // 目标 CPU 使用率, 0~100
	ensureCPUUsageInterval int64 // 统计间隔, 值等于 time.Duration 的直接转换

	pid struct {
		lock sync.Mutex
		kp   float64 // 比例系数
		ki   float64 // 积分系数
		kd   float64 // 微分系数
		// 这里的 error 表示误差, 不是错误, 请留意
		previousErrors *ring.Queue[float64]
	}
}{}

const (
	defaultPIDKp = 0.2
	defaultPIDKi = 0.1
	defaultPIDKd = 0.05

	// 表示 PID 默认的求和历史数量
	defaultPIDSumCount = 10
)

func init() {
	undo, _ := maxprocs.Set()
	internal.cpuNum = runtime.GOMAXPROCS(0)
	internal.debugger = func(format string, args ...any) {}
	internal.pid.kp = defaultPIDKp
	internal.pid.ki = defaultPIDKi
	internal.pid.kd = defaultPIDKd
	internal.pid.previousErrors = ring.NewRingQueue[float64](defaultPIDSumCount)
	undo()
}
