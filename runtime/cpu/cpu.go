package cpu

import (
	"errors"
	"math/rand"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/Andrew-M-C/go.util/runtime/cpu/internal/procfs"
)

// CoreNum 返回当前环境的 CPU 数
func CoreNum() int {
	return internal.cpuNum
}

// EnsureCPUUsage 确保 CPU 使用率不低于指定的百分比 (0 <= percentage < 100)。目前只支持
// Linux 系统。一旦启动就无法停止, 除非进程退出
func EnsureCPUUsage(percentage int, statisticInterval time.Duration) error {
	// 环境检查
	if runtime.GOOS != "linux" {
		return errors.New("EnsureCPUUsage only supports Linux")
	}

	// 参数检查
	if percentage <= 0 {
		return nil
	}
	if percentage >= 100 {
		return errors.New("percentage must be less than 100")
	}
	if statisticInterval < time.Second {
		return errors.New("statisticInterval must be greater than 1 second")
	}

	// 更新目标配置
	atomic.StoreInt32(&internal.ensureCPUUsageTarget, int32(percentage))
	atomic.StoreInt64(&internal.ensureCPUUsageInterval, int64(statisticInterval))

	// 开始唯一任务
	if !internal.ensureCPUUsageRunning.CompareAndSwap(false, true) {
		// 已经运行中了, 直接返回即可
		return nil
	}

	// 启动 routine
	go ensureCPUUsageRoutine()
	return nil
}

// SetPIDParameters 设置 PID 参数
func SetPIDParameters(kp, ki, kd float64) {
	internal.pid.lock.Lock()
	defer internal.pid.lock.Unlock()
	internal.pid.kp = kp
	internal.pid.ki = ki
	internal.pid.kd = kd
}

// SetDebugger 设置调试器, 用于输出调试信息
func SetDebugger(f func(string, ...any)) {
	if f != nil {
		internal.debugger = f
	}
}

func catchPanic(callback func()) {
	if e := recover(); e != nil {
		if callback != nil {
			callback()
		}
	}
}

func ensureCPUUsageRoutine() {
	// 再次标记
	internal.ensureCPUUsageRunning.Store(true)
	defer catchPanic(func() { go ensureCPUUsageRoutine() })

	// 测试代码
	prevStat, err := procfs.ReadCPUStat()
	if err != nil {
		internal.debugger("读取 CPU 状态失败: %v", err)
		internal.ensureCPUUsageRunning.Store(false)
		return
	}
	{
		interval := time.Duration(atomic.LoadInt64(&internal.ensureCPUUsageInterval))
		time.Sleep(interval)
	}

	// 表示根据 PID 算法得出的结果, 也就是不需要强行浪费的 CPU 时间比例
	var consumeCPURatio float64
	concurrency := CoreNum() * 4 // 并发数, 拍脑袋 4 个, 强压 CPU
	wg := &sync.WaitGroup{}

	iterate := func() {
		interval := time.Duration(atomic.LoadInt64(&internal.ensureCPUUsageInterval))
		nowStat, err := procfs.ReadCPUStat()
		if err != nil {
			internal.ensureCPUUsageRunning.Store(false)
			return
		}
		defer func() { prevStat = nowStat }()

		cpuUsage := getNormalizedCPUUsage(prevStat, nowStat)
		delta := pidCalculate(cpuUsage)

		actionCPURatio := consumeCPURatio + delta
		if actionCPURatio < 0 {
			actionCPURatio = 0
		} else if actionCPURatio > 1 {
			actionCPURatio = 1
		}

		consumeCPURatio = actionCPURatio
		internal.debugger("consumeCPURatio: %.2f", actionCPURatio)

		wg.Add(concurrency)
		for i := 0; i < concurrency; i++ {
			go func() {
				defer wg.Done()
				runCPUWithPercentage(consumeCPURatio, interval)
			}()
		}
		wg.Wait()

	}

	for {
		iterate()
	}
}

func runCPUWithPercentage(targetUsage float64, interval time.Duration) {
	const count = 10
	interval = interval / count

	runTime := time.Duration(float64(interval) * targetUsage)
	idleTime := interval - runTime

	for i := 0; i < count; i++ {
		runCPU(runTime)
		time.Sleep(idleTime)
	}
}

func runCPU(interval time.Duration) {
	start := time.Now()
	for time.Since(start) <= interval {
		_ = rand.Int() * rand.Int() * rand.Int() * rand.Int()
	}
}

func getNormalizedCPUUsage(prev, now procfs.CPUStat) float64 {
	cpuTotal := now.Total() - prev.Total()
	cpuIdle := now.Idle - prev.Idle
	return float64(cpuTotal-cpuIdle) / float64(cpuTotal)
}

// 计算根据 PID 算法的过滤比例。用在 ensureCPUUsageRoutine 函数中
//
// PID 公式:
//
//	u(k) = Kp·e(k) + Ki·SUM(e(n)) + Kd·(e(k) - e(k-1))
//
// 其中 SUM 实际上只取过去的一段时间, 这里是使用 ring 来实现的
func pidCalculate(cpuUsage float64) float64 {
	tgtUsage := float64(atomic.LoadInt32(&internal.ensureCPUUsageTarget)) / 100

	// PID 各参数
	currentError := tgtUsage - cpuUsage
	var kp, ki, kd float64
	var allPreviousErrors []float64

	internal.pid.lock.Lock()
	{
		kp, ki, kd = internal.pid.kp, internal.pid.ki, internal.pid.kd
		allPreviousErrors = internal.pid.previousErrors.GetAllValues()
		internal.pid.previousErrors.Push(currentError)
	}
	internal.pid.lock.Unlock()

	// P
	p := kp * currentError

	// I
	i := ki * (sum(allPreviousErrors) + currentError)

	// D
	d := 0.0
	if len(allPreviousErrors) > 0 {
		d = kd * (currentError - allPreviousErrors[0])
	}

	ut := p + i + d

	internal.debugger("(CPU %.2f, target %.2f) P: %.2f, I: %.2f, D: %.2f, ut: %.2f",
		cpuUsage, tgtUsage, p, i, d, ut,
	)

	return ut
}

func sum(v []float64) float64 {
	sum := 0.0
	for _, v := range v {
		sum += v
	}
	return sum
}
