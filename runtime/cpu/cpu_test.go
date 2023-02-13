package cpu_test

import (
	"runtime"
	"testing"

	_ "go.uber.org/automaxprocs"

	"github.com/Andrew-M-C/go.util/runtime/cpu"
	"github.com/smartystreets/goconvey/convey"
)

var (
	cv = convey.Convey
	so = convey.So
	eq = convey.ShouldEqual
)

func TestCPU(t *testing.T) {
	cv("测试 CoreNum 函数", t, func() { testCoreNum(t) })
}

func testCoreNum(t *testing.T) {
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
