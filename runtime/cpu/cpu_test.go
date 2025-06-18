package cpu_test

import (
	"runtime"
	"testing"
	"time"

	"github.com/Andrew-M-C/go.util/runtime/cpu"
	"github.com/Andrew-M-C/go.util/runtime/cpu/internal/procfs"
	"github.com/smartystreets/goconvey/convey"
	"go.uber.org/automaxprocs/maxprocs"
)

var (
	cv = convey.Convey
	so = convey.So
	eq = convey.ShouldEqual
	gt = convey.ShouldBeGreaterThan
	lt = convey.ShouldBeLessThan
)

func TestCPU(t *testing.T) {
	cv("测试 CoreNum 函数", t, func() { testCoreNum(t) })
	cv("测试 EnsureCPUUsage 函数", t, func() { testEnsureCPUUsage(t) })
}

func testCoreNum(t *testing.T) {
	_, _ = maxprocs.Set(maxprocs.Logger(t.Logf))

	// 首先设置一个正确的数值
	correctNum := runtime.GOMAXPROCS(0)
	t.Logf("Original GOMAXPROCS=%d", correctNum)

	// 我就是要设置一个不同的数值
	// mySet := correctNum * 4
	const mySet = 31

	runtime.GOMAXPROCS(mySet)
	n := runtime.GOMAXPROCS(0)
	t.Logf("Updating GOMAXPROCS=%d", n)
	so(n, eq, mySet)

	got := cpu.CoreNum()
	t.Logf("Got CPU num %d", got)
	so(got, eq, correctNum)

	// 但是这个数值应该是被 undo 回去了
	so(runtime.GOMAXPROCS(0), eq, mySet)
	t.Logf("Final GOMAXPROCS=%d", n)
}

func testEnsureCPUUsage(t *testing.T) {
	startProc, err := procfs.ReadCPUStat()
	so(err, eq, nil)

	cpu.SetDebugger(t.Logf)
	err = cpu.EnsureCPUUsage(60, time.Second)
	so(err, eq, nil)

	// 统计一段时间确认 CPU 使用率应该在一定范围内
	time.Sleep(time.Minute)

	afterProc, err := procfs.ReadCPUStat()
	so(err, eq, nil)

	idle := float64(afterProc.Idle-startProc.Idle) / float64(afterProc.Total()-startProc.Total())
	usage := 1.0 - idle
	t.Logf("平均 CPU 使用率: %.2f", usage)

	so(usage, gt, 0.50)
	so(usage, lt, 0.70)
}
