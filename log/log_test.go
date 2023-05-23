package log

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/Andrew-M-C/go.util/log/trace"
	"github.com/smartystreets/goconvey/convey"
)

var (
	cv = convey.Convey
	so = convey.So
	eq = convey.ShouldEqual
)

func TestLog(t *testing.T) {
	testInit(t)
	cv("调试", t, func() { testDebugging(t) })
	cv("测试自动删除", t, func() { testAutoRemove(t) })
	cv("测试 SetSkipCaller", t, func() { testSetSkipCaller(t) })
}

func testInit(t *testing.T) {
	internal.debugf = t.Logf
}

func testDebugging(t *testing.T) {
	SetFileName("./test.log")
	SetLevel(DebugLevel, InfoLevel)

	Debug("Hello,", "debug")
	Warnf("Hello, %s!", "warning")
	Error("Hello", "error")

	l := NewLogger()
	l.Debug("Hello,", "debug", "logger")
	l.Warnf("Hello, %s logger!", "warning")
	l.Error("Hello", "error", "logger")

	ctx := context.Background()
	InfoContext(ctx, "Hello, info context")
	InfoContextf(ctx, "Hello, %s context", "infof")

	ctx = trace.EnsureTraceID(ctx)
	InfoContext(ctx, "Hello, info context")
	InfoContextf(ctx, "Hello, %s trace context", "infof")

	time.Sleep(1 * time.Second)

	// 更新文件大小并快速创建多个日志
	SetLevel(DebugLevel, NoLog)
	SetFileSize(10 * 1000)

	logMany := func() {
		for i := 0; i < 100000; i++ {
			Warn("填充日志, 第", i+1, "条")
		}
	}
	logMany()
	time.Sleep(time.Second)

	logMany()
	time.Sleep(time.Second)

	// 尝试读取日志文件
	file, err := os.ReadFile(*internal.file.name)
	so(err, eq, nil)

	t.Logf("files size: %d", len(file))

	Infof("以下应该没有日志")
	SetLevel(NoLog, NoLog)
	Error("Hello", "no error")
}

func testAutoRemove(t *testing.T) {
	logMany := func() {
		for i := 0; i < 1000000; i++ {
			Warn("再次填充日志, 第", i+1, "条")
		}
	}

	SetLevel(DebugLevel, NoLog)
	start := time.Now()

	for i := 0; i < 15; i++ {
		logMany()
		time.Sleep(500 * time.Millisecond)

		files, _ := os.ReadDir(".")
		cnt := 0
		for _, f := range files {
			if strings.HasSuffix(f.Name(), ".log") {
				cnt++
				t.Log(f.Name())
			}
		}
		t.Logf("%v - 共有 %d 个日志文件", time.Since(start), cnt)
	}
}

func testSetSkipCaller(t *testing.T) {
	ctx := trace.SetTraceID(context.Background(), "testSetSkipCaller")

	SetFileLevel(NoLog)
	SetConsoleLevel(DebugLevel)
	SetSkipCaller(1)

	InfoContext(ctx, "这一个日志应该是包含了 TestLog() 函数的信息")

	time.Sleep(time.Second)
}
